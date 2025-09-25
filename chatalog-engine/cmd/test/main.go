package main

import (
	"context"
	"fmt"

	"github.com/defryfazz/fazztalog/config"
	"github.com/defryfazz/fazztalog/internal/ai"
	"github.com/defryfazz/fazztalog/internal/ai/engine"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	client := openai.NewClient(option.WithAPIKey(config.OpenAIToken))
	eng := engine.NewOpenAIEngine(client, config.TempFolderPath)

	res, err := eng.GenerateBrochure(context.Background(), ai.BrochureDetails{
		MerchantName: "Fazz Coffee",
		Products: []ai.Product{
			{Name: "Arabica Blend", Price: 49000},
			{Name: "Robusta Blend", Price: 39000},
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
