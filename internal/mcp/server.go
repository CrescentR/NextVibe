package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nextvibe/nextvibe/internal/brand"
	"github.com/nextvibe/nextvibe/internal/checker"
	"github.com/nextvibe/nextvibe/internal/detector"
	"github.com/nextvibe/nextvibe/internal/planner"
	"github.com/nextvibe/nextvibe/internal/protocol"
	"github.com/nextvibe/nextvibe/internal/scanner"
	"github.com/nextvibe/nextvibe/internal/state"
	"github.com/nextvibe/nextvibe/internal/taskgen"
)

const protocolVersion = "2026-07-02"

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type response struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  any              `json:"result,omitempty"`
	Error   *rpcError        `json:"error,omitempty"`
}

type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type toolCallResult struct {
	Content           []contentBlock `json:"content"`
	StructuredContent any            `json:"structuredContent,omitempty"`
	IsError           bool           `json:"isError,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Serve runs a newline-delimited JSON-RPC MCP server over the provided streams.
func Serve(root string, input io.Reader, output io.Writer, errOutput io.Writer) error {
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = wd
	}

	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		resp, writeResponse := handleMessage(root, line)
		if !writeResponse {
			continue
		}
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		if errOutput != nil {
			fmt.Fprintf(errOutput, "mcp read error: %v\n", err)
		}
		return err
	}
	return nil
}

func handleMessage(root string, line []byte) (response, bool) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		return errorResponse(nil, -32700, "parse error"), true
	}

	id, hasID := raw["id"]
	idPtr := (*json.RawMessage)(nil)
	if hasID {
		idCopy := append(json.RawMessage(nil), id...)
		idPtr = &idCopy
	}

	var method string
	if err := json.Unmarshal(raw["method"], &method); err != nil || method == "" {
		if !hasID {
			return response{}, false
		}
		return errorResponse(idPtr, -32600, "invalid request"), true
	}

	result, rpcErr := dispatch(root, method, raw["params"])
	if !hasID {
		return response{}, false
	}
	if rpcErr != nil {
		return response{JSONRPC: "2.0", ID: idPtr, Error: rpcErr}, true
	}
	return response{JSONRPC: "2.0", ID: idPtr, Result: result}, true
}

func dispatch(root, method string, params json.RawMessage) (any, *rpcError) {
	switch method {
	case "initialize":
		return map[string]any{
			"protocolVersion": protocolVersion,
			"serverInfo": map[string]any{
				"name":    brand.CommandName,
				"version": protocol.Version,
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		}, nil
	case "notifications/initialized":
		return map[string]any{}, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": tools()}, nil
	case "tools/call":
		result, err := callTool(root, params)
		if err != nil {
			return nil, &rpcError{Code: -32602, Message: err.Error()}
		}
		return result, nil
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found"}
	}
}

func tools() []tool {
	schema := map[string]any{
		"type":                 "object",
		"properties":           map[string]any{},
		"additionalProperties": false,
	}
	return []tool{
		{
			Name:        "nextvibe_scan",
			Description: "Scan the current project and persist the scan result to .nextvibe/state.json.",
			InputSchema: schema,
		},
		{
			Name:        "nextvibe_suggest",
			Description: "Suggest the next bounded task from the current project state.",
			InputSchema: schema,
		},
		{
			Name:        "nextvibe_task",
			Description: "Create or read the current NextVibe task.",
			InputSchema: schema,
		},
		{
			Name:        "nextvibe_check",
			Description: "Check whether the active task has passed its completion evidence.",
			InputSchema: schema,
		},
	}
}

func callTool(root string, params json.RawMessage) (toolCallResult, error) {
	var call struct {
		Name string `json:"name"`
	}
	if len(params) == 0 {
		return toolCallResult{}, fmt.Errorf("tools/call params are required")
	}
	if err := json.Unmarshal(params, &call); err != nil {
		return toolCallResult{}, err
	}

	var result any
	var err error
	switch call.Name {
	case "nextvibe_scan":
		result, err = scan(root)
	case "nextvibe_suggest":
		result, err = suggest(root)
	case "nextvibe_task":
		result, err = task(root)
	case "nextvibe_check":
		result, err = check(root)
	default:
		return toolCallResult{}, fmt.Errorf("unknown tool %q", call.Name)
	}
	if err != nil {
		return toolResult(map[string]string{"error": err.Error()}, true), nil
	}
	return toolResult(result, false), nil
}

func scan(root string) (detector.Result, error) {
	inventory, err := scanner.Scan(root)
	if err != nil {
		return detector.Result{}, err
	}
	result := detector.Detect(inventory)
	if err := state.SaveScan(root, result); err != nil {
		return detector.Result{}, err
	}
	return result, nil
}

func suggest(root string) (planner.Suggestion, error) {
	scanResult, err := scan(root)
	if err != nil {
		return planner.Suggestion{}, err
	}
	result := planner.Suggest(scanResult)
	if err := state.SaveSuggestion(root, result); err != nil {
		return planner.Suggestion{}, err
	}
	return result, nil
}

func task(root string) (taskgen.Task, error) {
	suggestion, err := suggest(root)
	if err != nil {
		return taskgen.Task{}, err
	}
	result, err := taskgen.CurrentOrCreate(root, suggestion)
	if err != nil {
		return taskgen.Task{}, err
	}
	if err := state.SaveTask(root, result.Task); err != nil {
		return taskgen.Task{}, err
	}
	return result.Task, nil
}

func check(root string) (checker.Result, error) {
	result := checker.CheckCurrent(root)
	if err := state.SaveCheck(root, result); err != nil {
		return checker.Result{}, err
	}
	return result, nil
}

func toolResult(result any, isError bool) toolCallResult {
	return toolCallResult{
		Content: []contentBlock{{
			Type: "text",
			Text: jsonText(result),
		}},
		StructuredContent: result,
		IsError:           isError,
	}
}

func jsonText(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func errorResponse(id *json.RawMessage, code int, message string) response {
	return response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: message},
	}
}

func RootFromArgs(args []string, defaultRoot string) (string, error) {
	if len(args) == 0 {
		return defaultRoot, nil
	}
	if len(args) == 2 && args[0] == "--root" {
		return filepath.Abs(args[1])
	}
	return "", fmt.Errorf("mcp accepts no arguments except --root <path>")
}
