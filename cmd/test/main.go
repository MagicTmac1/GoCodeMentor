package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func main() {
	// 方法1：最简单的方式（推荐）
	apiKey := getEnv("OPENAI_API_KEY", "sk-tagnhlgrouaooywifnvudysoashkfqfwbzorfhkboxwgzfwd")
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.siliconflow.cn/v1"
	client := openai.NewClientWithConfig(config)

	// 2. 准备对话历史（记住上下文）
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "你是一位热情的小学数学老师，擅长用生活中的例子解释数学概念。",
		},
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🎓 数学老师已上线！输入 'exit' 退出")
	fmt.Println("----------------------------------------")

	for {
		// 输入问题
		fmt.Print("\n👨‍🎓 学生问：")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		// 添加到历史
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})

		// 调用 Kimi API
		fmt.Print("🤖 老师答：")
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    "Qwen/Qwen2.5-7B-Instruct", // 使用8k模型，便宜够用
				Messages: messages,
				Stream:   false, // 先不用流式，简单点
			},
		)

		if err != nil {
			fmt.Printf("出错了：%v\n", err)
			continue
		}

		answer := resp.Choices[0].Message.Content
		fmt.Println(answer)

		// 把回答也加入历史，这样AI能记住上下文
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: answer,
		})
	}

	fmt.Println("再见！")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
