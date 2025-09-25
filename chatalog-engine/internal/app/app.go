package app

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type AppContainer struct {
	OpenAIClient openai.Client
}

type SetupAppParams struct {
	OpenAIToken string
}

func SetupApp(params SetupAppParams) AppContainer {
	client := openai.NewClient(
		option.WithAPIKey(params.OpenAIToken),
	)

	return AppContainer{
		OpenAIClient: client,
	}
}
