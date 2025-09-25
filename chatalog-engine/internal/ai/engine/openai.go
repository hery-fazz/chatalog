package engine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/defryfazz/fazztalog/internal/ai"
	"github.com/google/uuid"
	"github.com/openai/openai-go"
)

type OpenAIEngine struct {
	client  openai.Client
	tempDir string
}

func NewOpenAIEngine(client openai.Client, tempDir string) *OpenAIEngine {
	return &OpenAIEngine{
		client:  client,
		tempDir: tempDir,
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

func (e *OpenAIEngine) DetermineIntent(ctx context.Context, message string) (*ai.IntentResponse, error) {
	prompt := `
		You are an assistant that extracts user intent from input.
		There are several intents available:
		- %s: Brochure generation. This intent is used when the user wants to create a brochure for a product or service.
		- %s: Unknown. This intent is used when the user's intent is not listed in available list.

		IMPORTANT:
		- If the user input does not match any of the available intents, you must choose "unknown".
		- You must only choose one from the available intents.
		- If the intent is brochure generation, you must get the product's names from the message. If there is no product, just return an empty list.

		Based on the user input, determine the user's intent from the available list. Remember to only choose one from the available intents. If the user's intent is not listed, choose "unknown".
		Always return with correct JSON format without any \n or \t

		Example input and output:
		- Input: I want to create a brochure for my new product.
		  Output: {"intent": "brochure_generation","products": []}
		- Input: Please help to create a brochure for Fried Chicken and Coke.
		  Output: {"intent": "brochure_generation","products": ["Fried Chicken", "Coke"]}
		- Input: Hello, how are you?
		  Output: {"intent": "unknown","products": []}
	`

	resultIntent, err := e.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(fmt.Sprintf(prompt, ai.IntentBrochureGeneration, ai.IntentUnknown)),
			openai.UserMessage(message),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("================================")
	log.Printf("User Input: %s", message)
	log.Printf("DetermineIntent: %s", resultIntent.Choices[0].Message.Content)
	log.Printf("================================")

	var resultIntentData ai.IntentResponse
	err = json.Unmarshal([]byte(resultIntent.Choices[0].Message.Content), &resultIntentData)
	if err != nil {
		return nil, err
	}

	return &resultIntentData, nil
}

func (e *OpenAIEngine) GenerateBrochure(ctx context.Context, details ai.BrochureDetails) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Design a clean, modern ecommerce brochure for brand “%s”. ", details.MerchantName)
	fmt.Fprintf(&b, "Layout: square canvas, white or light background, soft shadows, neat grid. ")
	fmt.Fprintf(&b, "Top header: brand logo (from provided logo reference) at left, brand name at right in bold sans-serif. ")
	fmt.Fprintf(&b, "Products: show each product photo as the hero within rounded cards. Under each photo, show the product name and a clear price tag. ")
	fmt.Fprintf(&b, "Use consistent spacing, balanced margins, and visual hierarchy. If backgrounds are messy, neatly cut out products. ")
	fmt.Fprintf(&b, "Typography: clean sans-serif; prices visually prominent; include subtle accents.\n\n")

	fmt.Fprintf(&b, "Products to include (name → price):\n")
	for _, p := range details.Products {
		fmt.Fprintf(&b, "• %s → Rp %.0f\n", p.Name, p.Price)
	}

	fmt.Fprintf(&b, "\nDesign constraints:\n")
	fmt.Fprintf(&b, "- Arrange items in 2–3 columns depending on count; keep even gutters.\n")
	fmt.Fprintf(&b, "- Preserve product aspect ratios; avoid warping logos or products.\n")
	fmt.Fprintf(&b, "- Include small footer with “%s” and social placeholders.\n", details.MerchantName)
	fmt.Fprintf(&b, "- Export PNG with transparent background where possible.\n")

	prompt := b.String()

	res, err := e.client.Images.Generate(ctx, openai.ImageGenerateParams{
		Model:  openai.ImageModelGPTImage1,
		Prompt: prompt,
	})
	if err != nil {
		return "", err
	}

	if res.Data[0].URL != "" {
		tmpDir := fmt.Sprintf("%s/openai", e.tempDir)
		if err := os.MkdirAll(tmpDir, 0755); err != nil {
			return "", err
		}
		out := filepath.Join(tmpDir, uuid.New().String()+".png")

		// Use net/http to download the image
		resp, err := http.Get(res.Data[0].URL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		f, err := os.Create(out)
		if err != nil {
			return "", err
		}
		defer f.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			return "", err
		}

		return out, nil
	}
	if res.Data[0].B64JSON != "" {
		data, err := base64.StdEncoding.DecodeString(res.Data[0].B64JSON)
		if err != nil {
			return "", err
		}

		tmpDir := fmt.Sprintf("%s/openai", e.tempDir)
		if err := os.MkdirAll(tmpDir, 0755); err != nil {
			return "", err
		}

		out := filepath.Join(tmpDir, uuid.New().String()+".png")
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return "", err
		}

		return out, nil
	}

	return res.Data[0].URL, nil
}

func (e *OpenAIEngine) MatchProducts(ctx context.Context, productNames []string, products []ai.Product) ([]ai.Product, error) {
	prompt := `
		You are an assistant that finds product data by name. The selected product items may not match the exact names in the product, so you need to find the closest match using the product name.

            These are the available product items:
            "%s"

            Selected product item(s):
            "%s"
            
            For each selected item, check both the product name to find the closest match. If there is no exact match, choose the most similar item by name.
			Always return in JSON array format without any \n or \t.

			If the product item does not exist, return an empty list.

			Example input and output:
			- Input:
				Available product items:
					"Fried Chicken (Rp 20000), Coke (Rp 10000), Fries (Rp 15000)"
				Selected product item(s):
					"Fried Chicken, Coke"
				Output:[{"name": "Fried Chicken", "price": 20000},{"name": "Coke", "price": 10000}]
			- Input:
				Available product items:
					"Fried Chicken (Rp 20000), Coke (Rp 10000), Fries (Rp 15000)"
				Selected product item(s):
					"Pizza, Salad"
				Output:
					[]
	`

	availableProducts := make([]string, 0, len(products))
	for _, p := range products {
		availableProducts = append(availableProducts, fmt.Sprintf("%s (Rp %.0f)", p.Name, p.Price))
	}
	availableProductsStr := strings.Join(availableProducts, ", ")
	productNamesStr := strings.Join(productNames, ", ")

	prompt = fmt.Sprintf(prompt, availableProductsStr, productNamesStr)
	resultIntent, err := e.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("================================")
	log.Printf("Available Product: %s", availableProductsStr)
	log.Printf("Product Names: %s", productNamesStr)
	log.Printf("Match Product Result: %s", resultIntent.Choices[0].Message.Content)
	log.Printf("================================")

	var productResult []ai.Product
	err = json.Unmarshal([]byte(resultIntent.Choices[0].Message.Content), &productResult)
	if err != nil {
		return nil, err
	}

	return productResult, nil
}
