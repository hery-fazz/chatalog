package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ---------- JSON-RPC ----------
type JSONRPCRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}
type JSONRPCResponse struct {
	Jsonrpc string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ---------- MCP (subset) ----------
type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}
type ToolsCallParams struct {
	Name string           `json:"name"`
	Args *json.RawMessage `json:"arguments,omitempty"`
}
type ToolsCallResult struct {
	Content []ToolContent `json:"content"`
}
type ToolContent struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// ---------- Domain ----------
type Product struct {
	UserID   string  `json:"user_id"`
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	ImageURL *string `json:"image_url,omitempty"`
}
type FetchProductsInput struct {
	SheetURL string `json:"sheet_url,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	UserID   string `json:"user_id,omitempty"` // optional filter
}
type FetchProductsOutput struct {
	Items      []Product `json:"items"`
	Total      int       `json:"total"`
	NextOffset *int      `json:"next_offset,omitempty"`
}

// ---------- main loop ----------
func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		reqBytes, err := readFramedMessage(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}
		var req JSONRPCRequest
		if err := json.Unmarshal(reqBytes, &req); err != nil {
			writeError(nil, -32700, "Parse error (invalid JSON)", err.Error())
			continue
		}
		switch req.Method {
		case "initialize":
			handleInitialize(req.ID)
		case "tools/list":
			handleToolsList(req.ID)
		case "tools/call":
			handleToolsCall(req.ID, req.Params)
		default:
			writeError(req.ID, -32601, "Method not found", req.Method)
		}
	}
}

func handleInitialize(id any) {
	var res InitializeResult
	res.ProtocolVersion = "2024-11-05"
	res.Capabilities = map[string]any{"tools": map[string]any{}}
	res.ServerInfo.Name = "post-mcp-server-go"
	res.ServerInfo.Version = "0.1.1"
	writeResult(id, res)
}
func handleToolsList(id any) {
	tool := Tool{
		Name:        "fetch_products",
		Description: "Fetch products from a public Google Sheet CSV. Headers: id,name,price,currency (image_url,user_id optional).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"sheet_url": map[string]any{"type": "string", "description": "CSV export URL. If omitted, uses SHEET_URL env."},
				"limit":     map[string]any{"type": "number", "description": "Max items (default 50)."},
				"offset":    map[string]any{"type": "number", "description": "Offset (default 0)."},
				"user_id":   map[string]any{"type": "string", "description": "Optional filter by owner user_id (phone)."},
			},
			"additionalProperties": false,
		},
	}
	writeResult(id, ToolsListResult{Tools: []Tool{tool}})
}
func handleToolsCall(id any, params json.RawMessage) {
	var p ToolsCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		writeError(id, -32602, "Invalid params", err.Error())
		return
	}
	if p.Name != "fetch_products" {
		writeError(id, -32601, "Tool not found", p.Name)
		return
	}
	var in FetchProductsInput
	if p.Args != nil {
		if err := json.Unmarshal(*p.Args, &in); err != nil {
			writeError(id, -32602, "Invalid arguments", err.Error())
			return
		}
	}
	if in.Limit <= 0 {
		in.Limit = 50
	}
	if in.Limit > maxLimit() {
		in.Limit = maxLimit()
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	if in.SheetURL == "" {
		in.SheetURL = os.Getenv("SHEET_URL")
	}
	if in.SheetURL == "" {
		writeError(id, 1001, "CONFIG_ERROR", "sheet_url missing (arg or SHEET_URL env)")
		return
	}

	out, err := fetchProducts(in)
	if err != nil {
		writeError(id, 1002, "FETCH_ERROR", err.Error())
		return
	}
	blob, _ := json.MarshalIndent(out, "", "  ")
	writeResult(id, ToolsCallResult{
		Content: []ToolContent{{Type: "text", Text: string(blob)}},
	})
}

func fetchProducts(in FetchProductsInput) (FetchProductsOutput, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(in.SheetURL)
	if err != nil {
		return FetchProductsOutput{}, fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return FetchProductsOutput{}, fmt.Errorf("http status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return FetchProductsOutput{}, fmt.Errorf("read body: %w", err)
	}

	r := csv.NewReader(bytes.NewReader(all))
	r.TrimLeadingSpace = true
	rows, err := r.ReadAll()
	if err != nil {
		return FetchProductsOutput{}, fmt.Errorf("parse csv: %w", err)
	}
	if len(rows) == 0 {
		return FetchProductsOutput{}, fmt.Errorf("empty csv")
	}

	// header -> index
	header := make([]string, 0, len(rows[0]))
	for _, h := range rows[0] {
		header = append(header, strings.ToLower(strings.TrimSpace(h)))
	}
	idx := func(name string) int {
		for i, h := range header {
			if h == name {
				return i
			}
		}
		return -1
	}

	uidI := idx("user_id")
	idI := idx("id")
	nameI := idx("name")
	priceI := idx("price")
	currI := idx("currency")
	imgI := idx("image_url")

	if idI < 0 || nameI < 0 || priceI < 0 || currI < 0 {
		return FetchProductsOutput{}, fmt.Errorf("missing required headers: need id,name,price,currency")
	}

	seen := map[string]bool{}
	allProducts := make([]Product, 0, len(rows)-1)

	for _, row := range rows[1:] {
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}
		get := func(i int) string {
			if i >= 0 && i < len(row) {
				return strings.TrimSpace(row[i])
			}
			return ""
		}

		uid := ""
		if uidI >= 0 {
			uid = get(uidI)
		}

		id := get(idI)
		if id == "" || seen[id] {
			continue
		}
		name := get(nameI)
		if name == "" {
			continue
		}

		if in.UserID != "" && uid != in.UserID {
			continue
		}

		priceStr := strings.ReplaceAll(get(priceI), ",", "")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			continue
		}

		curr := get(currI)
		if curr == "" {
			curr = "IDR"
		}

		var imgPtr *string
		img := get(imgI)
		if img != "" && strings.HasPrefix(strings.ToLower(img), "http") {
			imgPtr = &img
		}

		allProducts = append(allProducts, Product{
			UserID:   uid,
			ID:       id,
			Name:     name,
			Price:    price,
			Currency: curr,
			ImageURL: imgPtr,
		})
		seen[id] = true
	}

	total := len(allProducts)
	start := in.Offset
	if start > total {
		start = total
	}
	end := start + in.Limit
	if end > total {
		end = total
	}
	items := allProducts[start:end]

	var next *int
	if end < total {
		n := end
		next = &n
	}

	return FetchProductsOutput{
		Items:      items,
		Total:      total,
		NextOffset: next,
	}, nil
}

// ---------- framing helpers ----------
func readFramedMessage(r *bufio.Reader) ([]byte, error) {
	var contentLength int
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			v := strings.TrimSpace(line[len("content-length:"):])
			cl, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %q", v)
			}
			contentLength = cl
		}
		// ignore other headers
	}
	if contentLength <= 0 {
		return nil, fmt.Errorf("missing Content-Length")
	}
	buf := make([]byte, contentLength)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func writeFramedJSON(b []byte) {
	h := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/json\r\n\r\n", len(b))
	os.Stdout.Write([]byte(h))
	os.Stdout.Write(b)
}

func writeResult(id any, result any) {
	resp := JSONRPCResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  result,
	}
	b, _ := json.Marshal(resp)
	writeFramedJSON(b)
}

func writeError(id any, code int, message string, data any) {
	resp := JSONRPCResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	b, _ := json.Marshal(resp)
	writeFramedJSON(b)
}

func maxLimit() int {
	if v := os.Getenv("MAX_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 200
}
