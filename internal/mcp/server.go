// Package mcp implements a lightweight MCP JSON-RPC stdio server.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/queue"
	"github.com/coffeegraph/coffeegraph/internal/runner"
)

type req struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type resp struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcErr     `json:"error,omitempty"`
}

type rpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Run serves MCP over stdio.
func Run(ctx context.Context, root string, engine *runner.Engine, in io.Reader, out io.Writer) error {
	sc := bufio.NewScanner(in)
	enc := json.NewEncoder(out)
	tools := discoverSkills(root)
	for sc.Scan() {
		var r req
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			_ = enc.Encode(resp{JSONRPC: "2.0", Error: &rpcErr{Code: -32700, Message: "parse error"}})
			continue
		}
		switch r.Method {
		case "initialize":
			_ = enc.Encode(resp{JSONRPC: "2.0", ID: r.ID, Result: map[string]interface{}{
				"serverInfo": map[string]string{"name": "coffeegraph", "version": "0.1.0"},
				"capabilities": map[string]interface{}{
					"tools": map[string]bool{"listChanged": false},
				},
			}})
		case "tools/list":
			list := make([]map[string]string, 0, len(tools))
			for _, t := range tools {
				list = append(list, map[string]string{
					"name":        t,
					"description": fmt.Sprintf("Run CoffeeGraph skill %s with input.task", t),
				})
			}
			_ = enc.Encode(resp{JSONRPC: "2.0", ID: r.ID, Result: map[string]interface{}{"tools": list}})
		case "tools/call":
			var p struct {
				Name      string `json:"name"`
				Arguments struct {
					Task string `json:"task"`
				} `json:"arguments"`
			}
			if err := json.Unmarshal(r.Params, &p); err != nil {
				_ = enc.Encode(resp{JSONRPC: "2.0", ID: r.ID, Error: &rpcErr{Code: -32602, Message: "invalid params"}})
				continue
			}
			res, err := engine.ExecuteTask(ctx, queue.Item{Skill: p.Name, Task: p.Arguments.Task, Priority: 3})
			if err != nil {
				_ = enc.Encode(resp{JSONRPC: "2.0", ID: r.ID, Error: &rpcErr{Code: -32000, Message: err.Error()}})
				continue
			}
			_ = enc.Encode(resp{JSONRPC: "2.0", ID: r.ID, Result: map[string]interface{}{
				"content": []map[string]string{{"type": "text", "text": res.Text}},
			}})
		default:
			_ = enc.Encode(resp{JSONRPC: "2.0", ID: r.ID, Error: &rpcErr{Code: -32601, Message: "method not found"}})
		}
	}
	return sc.Err()
}

func discoverSkills(root string) []string {
	ents, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(ents))
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		out = append(out, e.Name())
	}
	return out
}
