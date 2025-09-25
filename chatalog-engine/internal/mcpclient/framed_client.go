package mcpclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	cmd *exec.Cmd
	inw *bufio.Writer
	out *bufio.Reader
	mu  sync.Mutex
	id  int64
}

type rpcReq struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}
type rpcErr struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
type rpcResp struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcErr         `json:"error,omitempty"`
}

// MCP framed write (Content-Length + body)
func writeFramedJSON(w *bufio.Writer, b []byte) error {
	h := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/json\r\n\r\n", len(b))
	if _, err := w.WriteString(h); err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	return w.Flush()
}

// MCP framed read: parse headers → read content-length → read body bytes
func readFramedJSON(r *bufio.Reader) ([]byte, error) {
	// Read headers
	var headers []string
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" { // blank line separates headers/body
			break
		}
		headers = append(headers, line)
	}
	var contentLength int
	for _, h := range headers {
		if strings.HasPrefix(strings.ToLower(h), "content-length:") {
			v := strings.TrimSpace(strings.TrimPrefix(h, "Content-Length:"))
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("bad Content-Length: %v", err)
			}
			contentLength = n
			break
		}
	}
	if contentLength <= 0 {
		return nil, errors.New("missing Content-Length")
	}

	buf := make([]byte, contentLength)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func Start(ctx context.Context, command string, args ...string) (*Client, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	c := &Client{
		cmd: cmd,
		inw: bufio.NewWriter(stdin),
		out: bufio.NewReader(stdout),
		id:  1,
	}
	// initialize handshake
	_, _ = c.callRaw(ctx, "initialize", nil)
	return c, nil
}

func (c *Client) Close() error {
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	return nil
}

func (c *Client) CallToolTextJSON(ctx context.Context, name string, args map[string]any, out interface{}) error {
	// Call tools/call → server returns { result: { content: [{type:"text", text:"<json>"}] } }
	raw, err := c.callRaw(ctx, "tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})
	if err != nil {
		return err
	}

	// Envelope
	var env struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode tools/call envelope: %w", err)
	}
	if len(env.Content) == 0 || env.Content[0].Type != "text" {
		return errors.New("unexpected tools/call content shape")
	}
	// The text holds JSON string; parse into out if provided
	if out == nil {
		return nil
	}
	dec := json.NewDecoder(bytes.NewBufferString(env.Content[0].Text))
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func (c *Client) callRaw(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	id := c.id
	c.id++
	c.mu.Unlock()

	reqb, _ := json.Marshal(rpcReq{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	})
	if err := writeFramedJSON(c.inw, reqb); err != nil {
		return nil, err
	}

	// Read framed response
	body, err := readFramedJSON(c.out)
	if err != nil {
		return nil, err
	}

	var r rpcResp
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshal rpcResp: %w", err)
	}
	if r.ID != id {
		return nil, fmt.Errorf("mismatched id: want %d got %d", id, r.ID)
	}
	if r.Error != nil {
		return nil, fmt.Errorf("mcp error %d: %s", r.Error.Code, r.Error.Message)
	}
	return r.Result, nil
}
