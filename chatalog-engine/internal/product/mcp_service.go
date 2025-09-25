package product

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/defryfazz/fazztalog/internal/mcpclient"
)

type MCPService struct {
	Client       *mcpclient.Client
	ToolName     string // "fetch_products"
	DefaultLimit int
	SheetURL     string // optional override; kalau kosong, server pakai ENV SHEET_URL miliknya
}

func NewMCPService(ctx context.Context) (*MCPService, error) {
	cmd := getenv("MCP_CMD", "")
	args := getenv("MCP_ARGS", "")
	tool := getenv("MCP_PRODUCT_TOOL", "fetch_products")
	limit := atoi(getenv("MCP_PRODUCT_LIMIT", "50"))
	sheet := getenv("MCP_SHEET_URL", "") // opsional

	if cmd == "" {
		return nil, fmt.Errorf("MCP_CMD empty")
	}
	cl, err := mcpclient.Start(ctx, cmd, args)
	if err != nil {
		return nil, err
	}

	return &MCPService{
		Client:       cl,
		ToolName:     tool,
		DefaultLimit: limit,
		SheetURL:     sheet,
	}, nil
}

func (s *MCPService) Close() error { return s.Client.Close() }

func (s *MCPService) ListByPhone(ctx context.Context, phone string, limit, offset int) (ListOutput, error) {
	if limit <= 0 {
		limit = s.DefaultLimit
	}
	args := map[string]any{
		"user_id": phone,
		"limit":   limit,
		"offset":  offset,
	}
	if s.SheetURL != "" {
		args["sheet_url"] = s.SheetURL
	}
	var out ListOutput
	err := s.Client.CallToolTextJSON(ctx, s.ToolName, args, &out)
	return out, err
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
func atoi(s string) int { n, _ := strconv.Atoi(s); return n }
