package internal

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// LLMClient wraps the OpenAI client for chat completions with tool support.
type LLMClient struct {
	client openai.Client
	model  string
	tools  []openai.ChatCompletionToolParam
}

// NewLLMClient creates a new LLM client.
func NewLLMClient(baseURL, apiKey, model string, tools []any) *LLMClient {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	}
	client := openai.NewClient(opts...)

	toolParams := convertTools(tools)

	return &LLMClient{
		client: client,
		model:  model,
		tools:  toolParams,
	}
}

// Chat runs the chat loop: sends messages to the LLM, handles tool calls,
// and streams responses to stdout.
func (c *LLMClient) Chat(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) error {
	for {
		slog.Info("calling LLM", "model", c.model, "messages", len(messages))

		stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    c.model,
			Messages: messages,
			Tools:    c.tools,
		})

		// Accumulators for tool calls
		toolCallIndices := make(map[int]int) // api_index -> accumulated slice index
		var toolCalls []openai.ChatCompletionMessageToolCall
		var textBuf strings.Builder

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}
			delta := chunk.Choices[0].Delta

			// Collect tool calls from delta
			for _, tc := range delta.ToolCalls {
				idx := int(tc.Index)
				if _, exists := toolCallIndices[idx]; !exists {
					toolCallIndices[idx] = len(toolCalls)
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCall{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageToolCallFunction{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					})
				} else {
					realIdx := toolCallIndices[idx]
					toolCalls[realIdx].Function.Arguments += tc.Function.Arguments
				}
			}

			// Stream text to stdout
			if delta.Content != "" {
				textBuf.WriteString(delta.Content)
				fmt.Print(delta.Content)
				os.Stdout.Sync()
			}
		}

		if err := stream.Err(); err != nil {
			slog.Error("stream error", "err", err)
			return err
		}

		// If no tool calls, we're done with this turn
		if len(toolCalls) == 0 {
			if textBuf.Len() > 0 {
				fmt.Println()
				messages = append(messages, openai.AssistantMessage(textBuf.String()))
			}
			return nil
		}

		// Add assistant message with tool calls to history
		messages = append(messages, buildAssistantMessageWithToolCalls(textBuf.String(), toolCalls))

		// Execute each tool call
		for _, tc := range toolCalls {
			slog.Info("tool called", "tool", tc.Function.Name, "args", tc.Function.Arguments)

			args, err := ParseToolArgs(tc.Function.Arguments)
			if err != nil {
				slog.Warn("failed to parse tool arguments", "tool", tc.Function.Name, "err", err)
				result := "failed to parse arguments: " + err.Error()
				slog.Info("tool result", "tool", tc.Function.Name, "success", false)
				messages = append(messages, openai.ToolMessage(tc.ID, result))
				continue
			}

			toolFunc := GetToolFunc(tc.Function.Name)
			if toolFunc == nil {
				slog.Warn("unknown tool", "tool", tc.Function.Name)
				result := "unknown tool: " + tc.Function.Name
				slog.Info("tool result", "tool", tc.Function.Name, "success", false)
				messages = append(messages, openai.ToolMessage(tc.ID, result))
				continue
			}

			result := toolFunc(args)
			slog.Info("tool result", "tool", tc.Function.Name, "success", result.Success)
			messages = append(messages, openai.ToolMessage(tc.ID, result.Raw))
		}

		// Loop back to call LLM again with tool results
	}
}

// convertTools converts []any tool definitions to []openai.ChatCompletionToolParam.
func convertTools(tools []any) []openai.ChatCompletionToolParam {
	var result []openai.ChatCompletionToolParam
	for _, t := range tools {
		m, ok := t.(map[string]any)
		if !ok {
			continue
		}
		fn, ok := m["function"].(map[string]any)
		if !ok {
			continue
		}

		name, _ := fn["name"].(string)
		desc, _ := fn["description"].(string)
		params, _ := fn["parameters"].(map[string]any)

		result = append(result, openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        name,
				Description: openai.String(desc),
				Parameters:  openai.FunctionParameters(params),
			},
		})
	}
	return result
}

// buildAssistantMessageWithToolCalls creates an assistant message with tool calls.
func buildAssistantMessageWithToolCalls(content string, toolCalls []openai.ChatCompletionMessageToolCall) openai.ChatCompletionMessageParamUnion {
	// Convert tool calls to param format
	var tcParams []openai.ChatCompletionMessageToolCallParam
	for _, tc := range toolCalls {
		tcParams = append(tcParams, openai.ChatCompletionMessageToolCallParam{
			ID: tc.ID,
			Function: openai.ChatCompletionMessageToolCallFunctionParam{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}

	// Build the assistant message
	msg := openai.ChatCompletionAssistantMessageParam{
		ToolCalls: tcParams,
	}

	if content != "" {
		msg.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
			OfString: openai.String(content),
		}
	}

	return openai.ChatCompletionMessageParamUnion{
		OfAssistant: &msg,
	}
}
