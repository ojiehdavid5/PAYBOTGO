package ai

import (
	"context"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func analyzeIntent(message string) (string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4, // or GPT3.5
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: `You are an intent classifier. For the user's input, return one of the following intents: airtime, balance, transfer, register, login, unknown.`,
			},
			{
				Role:    "user",
				Content: message,
			},
		},
	})
	if err != nil {
		
		return "", err
	}

	return strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content)), nil
}
