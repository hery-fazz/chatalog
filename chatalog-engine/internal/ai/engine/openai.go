package engine

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func (e *OpenAIEngine) GenerateBrochureImage(ctx context.Context, userMessage string) (string, string, error) {
	sys := `You are a graphic designer generating a clean Indonesian promo poster.
			- Keep layout simple, high contrast, readable big title and price.
			- Fit for WhatsApp sharing. No watermark. No NSFW.`

	userPrompt := fmt.Sprintf(
			"Buat poster promosi berdasarkan instruksi ini (Bahasa Indonesia): %s\n"+
			"Fokuskan pada judul promo, harga, 2–3 bullet benefit, dan tone warna berkontras tinggi.",
		userMessage,
	)

	captionResp, err := e.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Tuliskan caption WA yang singkat dan ramah untuk menyertai poster, maksimal 140 karakter."),
			openai.UserMessage(userMessage),
		},
	})
	if err != nil {
		return "", "", err
	}
	caption := strings.TrimSpace(captionResp.Choices[0].Message.Content)
	if caption == "" {
		caption = "Siap! Berikut poster promonya 🚀"
	}

	img, err := e.client.Images.Generate(ctx, openai.ImageGenerateParams{
		Model:  openai.ImageModelGPTImage1,      
		Prompt: sys + "\n\n" + userPrompt,        
		Size:   "1024x1024",                     
	})
	if err != nil {
		return "", "", err
	}
	if len(img.Data) == 0 {
		return "", "", fmt.Errorf("no image returned")
	}

	if img.Data[0].URL != "" {
		return img.Data[0].URL, caption, nil
	}
	if img.Data[0].B64JSON != "" {
		data, err := base64.StdEncoding.DecodeString(img.Data[0].B64JSON)
		if err != nil {
			return "", "", err
		}
		tmpDir := os.TempDir()
		out := filepath.Join(tmpDir, "chatalog_poster.jpg")
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return "", "", err
		}
		return out, caption, nil
	}

	return "", "", fmt.Errorf("image has neither URL nor base64 data")
}