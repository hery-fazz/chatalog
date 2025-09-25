// client.go
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ---------- JSON-RPC ----------
type rpcResp struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcErr         `json:"error,omitempty"`
}
type rpcErr struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type toolsCallResult struct {
	Content []toolContent `json:"content"`
}
type toolContent struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// ---------- Domain output dari MCP ----------
type Product struct {
	UserID   string  `json:"user_id"`
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	ImageURL *string `json:"image_url,omitempty"`
}
type FetchProductsOutput struct {
	Items      []Product `json:"items"`
	Total      int       `json:"total"`
	NextOffset *int      `json:"next_offset,omitempty"`
}

func main() {
	// Flags
	serverPath := flag.String("server", "./mcp-sheet", "path ke binary MCP server")
	sheetURL := flag.String("sheet", "", "CSV export URL Google Sheet (opsional, override env di server)")
	userID := flag.String("user", "", "filter user_id (nomor HP), opsional")
	limit := flag.Int("limit", 5, "limit item")
	offset := flag.Int("offset", 0, "offset item")
	flag.Parse()

	// Spawn server MCP
	cmd := exec.Command(*serverPath)
	// Teruskan SHEET_URL via env ke server (boleh kosong jika kirim via args)
	if *sheetURL != "" {
		cmd.Env = append(os.Environ(), "SHEET_URL="+*sheetURL)
	} else {
		cmd.Env = os.Environ()
	}
	stdin, err := cmd.StdinPipe()
	must(err)
	stdout, err := cmd.StdoutPipe()
	must(err)
	cmd.Stderr = os.Stderr
	must(cmd.Start())

	w := bufio.NewWriter(stdin)
	r := bufio.NewReader(stdout)

	// 1) initialize
	sendJSONRPC(w, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})
	_, _ = readFrame(r) // boleh diabaikan

	// 2) tools/call fetch_products
	args := map[string]any{"limit": *limit, "offset": *offset}
	if *userID != "" {
		args["user_id"] = *userID
	}
	// Jika tidak set env SHEET_URL di atas, bisa kirim sheet_url di args:
	// if *sheetURL != "" { args["sheet_url"] = *sheetURL }

	sendJSONRPC(w, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "fetch_products",
			"arguments": args,
		},
	})

	respBody, err := readFrame(r)
	must(err)

	var rpc rpcResp
	must(json.Unmarshal(respBody, &rpc))
	if rpc.Error != nil {
		fmt.Printf("RPC Error %d: %s\n", rpc.Error.Code, rpc.Error.Message)
		if len(rpc.Error.Data) > 0 {
			fmt.Printf("Data: %s\n", string(rpc.Error.Data))
		}
		_ = cmd.Process.Kill()
		os.Exit(1)
	}

	var callRes toolsCallResult
	must(json.Unmarshal(rpc.Result, &callRes))
	if len(callRes.Content) == 0 {
		fmt.Println("No content from server.")
		_ = cmd.Process.Kill()
		return
	}

	var out FetchProductsOutput
	must(json.Unmarshal([]byte(callRes.Content[0].Text), &out))

	// Print hasil
	fmt.Printf("Total items: %d\n", out.Total)
	for i, p := range out.Items {
		img := "-"
		if p.ImageURL != nil {
			img = *p.ImageURL
		}
		fmt.Printf("%2d) user:%s  id:%s  name:%s  price:%.0f %s  img:%s\n",
			i+1, nz(p.UserID, "-"), p.ID, p.Name, p.Price, p.Currency, img)
	}

	// stop server (PoC)
	_ = cmd.Process.Kill()
}

// ---------- framing: write/read LSP-style frames ----------
func sendJSONRPC(w *bufio.Writer, payload any) {
	body, _ := json.Marshal(payload)
	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/json\r\n\r\n", len(body))
	w.WriteString(header)
	w.Write(body)
	w.Flush()
}

func readFrame(r *bufio.Reader) ([]byte, error) {
	var contentLength int
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // end headers
		}
		l := strings.ToLower(line)
		if strings.HasPrefix(l, "content-length:") {
			v := strings.TrimSpace(line[len("content-length:"):])
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %q", v)
			}
			contentLength = n
		}
	}
	if contentLength <= 0 {
		return nil, io.ErrUnexpectedEOF
	}
	buf := make([]byte, contentLength)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func nz(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
