package app

import (
	"github.com/defryfazz/fazztalog/internal/ai"
	"github.com/defryfazz/fazztalog/internal/ai/engine"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type AppContainer struct {
	AIEngine ai.Engine
}

type SetupAppParams struct {
	OpenAIToken string
}

func SetupApp(params SetupAppParams) AppContainer {
	client := openai.NewClient(
		option.WithAPIKey(params.OpenAIToken),
	)
	aiEngine := engine.NewOpenAIEngine(client)

	return AppContainer{
		AIEngine: aiEngine,
	}
}
