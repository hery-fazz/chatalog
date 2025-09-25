package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/defryfazz/fazztalog/internal/ai/engine"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func downloadToFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil { return err }
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return err }

	out, err := os.Create(path)
	if err != nil { return err }
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	_ = godotenv.Load(".env")

	apiKey := strings.TrimSpace(os.Getenv("OPEN_AI_TOKEN"))

	if apiKey == "" {
		os.Exit(1)
	}

	text := strings.Join(os.Args[1:], " ")
	if strings.TrimSpace(text) == "" {
		text = "Buat poster promosi warteg, dengan menu nasi goreng, dengan harga 20.000"
	}

	cli := openai.NewClient(option.WithAPIKey(apiKey))
	eng := engine.NewOpenAIEngine(cli)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	urlOrPath, caption, err := eng.GenerateBrochureImage(ctx, text)
	if err != nil {
		fmt.Println("generate error:", err)
		os.Exit(1)
	}

	savePath := urlOrPath
	if strings.HasPrefix(urlOrPath, "http") {
		savePath = filepath.Join(os.TempDir(), "chatalog_gen.jpg")
		if err := downloadToFile(urlOrPath, savePath); err != nil {
			fmt.Println("download error:", err)
			os.Exit(1)
		}
	}

	out := map[string]string{"caption": caption, "file": savePath}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}
