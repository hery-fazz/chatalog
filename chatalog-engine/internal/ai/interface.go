package ai

import (
	"context"
	"io"
)

type Engine interface {
	TranscribeAudio(ctx context.Context, file io.Reader) (string, error)
	DetermineIntent(ctx context.Context, message string) (string, error)
	GenerateBrochureImage(ctx context.Context, userMessage string) (url string, caption string, err error)
}
