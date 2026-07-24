package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
}

// New initializes an OpenAI-compatible client.
func New(apiKey, model, baseURL string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("api_key cannot be empty")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("model cannot be empty")
	}
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("base_url cannot be empty")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return &Client{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}, nil
}

// SendMessage sends a system prompt and message to the LLM model and returns the response.
func (c *Client) SendMessage(ctx context.Context, message, systemPrompt string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", fmt.Errorf("message cannot be empty")
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: message,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("AI completion error: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("AI response is empty")
	}

	return resp.Choices[0].Message.Content, nil
}
