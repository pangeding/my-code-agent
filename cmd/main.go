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
	scanner := bufio.NewScanner(os.Stdin)

	// 确认回调函数：向用户展示危险命令并等待确认
	confirmFunc := func(command, riskType string) bool {
		fmt.Println()
		fmt.Println("⚠  安全确认: 该命令命中安全规则")
		fmt.Printf("  风险类型: %s\n", riskType)
		fmt.Printf("  命令: %s\n", command)
		fmt.Print("  是否执行？(y/n): ")

		if !scanner.Scan() {
			return false
		}
		resp := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return resp == "y" || resp == "yes"
	}

	client := internal.NewLLMClient(cfg.BaseURL, cfg.APIKey, cfg.Model, tools, confirmFunc)

	fmt.Println("AI 文件编辑助手已启动。输入 q 或 quit 退出。")
	fmt.Println()

	var messages []openai.ChatCompletionMessageParamUnion

	if cfg.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(cfg.SystemPrompt))
	}

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
