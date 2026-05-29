package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"my-code-agent/internal"
)

func main() {
	cfg := internal.LoadConfig()

	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: AI_API_KEY is not set")
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	tools := internal.GetTools()
	client := internal.NewLLMClient(cfg.BaseURL, cfg.APIKey, cfg.Model, tools)

	fmt.Println("AI 文件编辑助手已启动。输入 q 或 quit 退出。")
	fmt.Println()

	var messages []openai.ChatCompletionMessageParamUnion

	if cfg.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(cfg.SystemPrompt))
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "q") || strings.EqualFold(input, "quit") {
			fmt.Println("再见！")
			break
		}

		messages = append(messages, openai.UserMessage(input))

		ctx := context.Background()
		if err := client.Chat(ctx, messages); err != nil {
			slog.Error("chat error", "err", err)
		}
	}
}
