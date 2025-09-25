package ai

import (
	"context"
	"io"
)

type IntentResponse struct {
	Intent   string   `json:"intent"`
	Products []string `json:"products"`
}

type Engine interface {
	TranscribeAudio(ctx context.Context, file io.Reader) (string, error)
	DetermineIntent(ctx context.Context, message string) (*IntentResponse, error)
	GenerateBrochure(ctx context.Context, details BrochureDetails) (string, error)
	MatchProducts(ctx context.Context, productNames []string, products []Product) ([]Product, error)
}
