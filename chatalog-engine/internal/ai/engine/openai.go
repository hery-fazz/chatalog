package engine

import (
	"context"
	"fmt"
	"io"

	"github.com/defryfazz/fazztalog/internal/ai"
	"github.com/openai/openai-go"
)

type OpenAIEngine struct {
	client openai.Client
}

func NewOpenAIEngine(client openai.Client) *OpenAIEngine {
	return &OpenAIEngine{
		client: client,
	}
}

func (e *OpenAIEngine) TranscribeAudio(ctx context.Context, file io.Reader) (string, error) {
	res, err := e.client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		Model: openai.AudioModelWhisper1,
		File:  file,
	})
	if err != nil {
		return "", err
	}

	return res.Text, nil
}

func (e *OpenAIEngine) DetermineIntent(ctx context.Context, message string) (string, error) {
	prompt := `
		You are an assistant that extracts user intent from input.
		There are several intents available:
		- %s: Brochure generation. This intent is used when the user wants to create a brochure for a product or service.
		- %s: Unknown. This intent is used when the user's intent is not listed in available list.

		Based on the user input, determine the user's intent from the available list. Remember to only choose one from the available intents. If the user's intent is not listed, choose "unknown".

		Example input and output:
		Input: I want to create a brochure for my new product.
		Output: brochure_generation
	`

	resultIntent, err := e.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(fmt.Sprintf(prompt, ai.IntentBrochureGeneration, ai.IntentUnknown)),
			openai.UserMessage(message),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return "", err
	}

	return resultIntent.Choices[0].Message.Content, nil
}
