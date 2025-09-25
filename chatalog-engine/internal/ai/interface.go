package ai

import (
	"context"
	"io"
)

type Engine interface {
	TranscribeAudio(ctx context.Context, file io.Reader) (string, error)
	DetermineIntent(ctx context.Context, message string) (string, error)
	GenerateBrochure(ctx context.Context, details BrochureDetails) (string, error)
}
