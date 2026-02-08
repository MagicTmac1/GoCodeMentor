package siliconflow

import (
	"context"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	cli   *openai.Client
	model string
}

// 添加这个结构体定义（如果之前没有）
type Message struct {
	Role    string
	Content string
}

// NewClient 创建硅基流动客户端
func NewClient() *Client {
	// 从环境变量读取，如果没有则使用默认值（实际使用时请替换）
	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		// 临时硬编码，实际生产环境请使用环境变量
		apiKey = "sk-tagnhlgrouaooywifnvudysoashkfqfwbzorfhkboxwgzfwd" // 你的 Key
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.siliconflow.cn/v1"

	return &Client{
		cli:   openai.NewClientWithConfig(config),
		model: "Qwen/Qwen2.5-7B-Instruct", // 硅基支持的模型，免费额度可用
	}
}

// Chat 非流式对话
func (c *Client) Chat(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	return c.ChatCompletion(ctx, userMessage, nil)
}

// ChatCompletion AI对话补全（兼容旧接口）
func (c *Client) ChatCompletion(ctx context.Context, prompt string, messages []Message) (string, error) {
	// 构建消息列表
	var openaiMessages []openai.ChatCompletionMessage

	// 如果有历史消息，添加它们
	if messages != nil && len(messages) > 0 {
		for _, msg := range messages {
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 添加用户当前提示
	openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	resp, err := c.cli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMessages,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

// ChatStream 流式对话（打字机效果）
func (c *Client) ChatStream(ctx context.Context, systemPrompt, userMessage string) (chan string, error) {
	stream, err := c.cli.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		},
		Stream: true,
	})
	if err != nil {
		return nil, err
	}

	resultChan := make(chan string)
	go func() {
		defer close(resultChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					resultChan <- content
				}
			}
		}
	}()

	return resultChan, nil
}

// ChatWithHistory 带历史记录的对话
func (c *Client) ChatWithHistory(ctx context.Context, messages []Message) (string, error) {
	// 转换消息格式
	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	resp, err := c.cli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMessages,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
