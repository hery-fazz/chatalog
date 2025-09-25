package app

import (
	"database/sql"

	"github.com/defryfazz/fazztalog/internal/ai"
	"github.com/defryfazz/fazztalog/internal/ai/engine"
	"github.com/defryfazz/fazztalog/internal/merchant"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type AppContainer struct {
	AIEngine        ai.Engine
	MerchantService merchant.Service
}

type SetupAppParams struct {
	DB            *sql.DB
	OpenAIToken   string
	TempDirectory string
}

func SetupApp(params SetupAppParams) AppContainer {
	repositories := setupRepositories(params.DB)

	client := openai.NewClient(
		option.WithAPIKey(params.OpenAIToken),
	)
	aiEngine := engine.NewOpenAIEngine(client, params.TempDirectory)
	merchantService := merchant.NewService(repositories.Merchant, aiEngine)

	return AppContainer{
		AIEngine:        aiEngine,
		MerchantService: merchantService,
	}
}
