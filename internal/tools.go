package internal

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
)

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Success bool
	Raw     string
}

// ToolFunc is the function signature for tool implementations.
type ToolFunc func(args map[string]any) ToolResult

// GetTools returns the list of tool definitions for the OpenAI API.
func GetTools() []any {
	return []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "read_file",
				"description": "Read the contents of a file and return them as text.",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]string{
							"type":        "string",
							"description": "The path to the file to read.",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "write_file",
				"description": "Write content to a file, creating it if it doesn't exist.",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]string{
							"type":        "string",
							"description": "The path to the file to write.",
						},
						"content": map[string]string{
							"type":        "string",
							"description": "The content to write to the file.",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "edit_file",
				"description": "Replace a specific string in a file with new content.",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]string{
							"type":        "string",
							"description": "The path to the file to edit.",
						},
						"old_string": map[string]string{
							"type":        "string",
							"description": "The text to replace.",
						},
						"new_string": map[string]string{
							"type":        "string",
							"description": "The text to replace it with.",
						},
					},
					"required": []string{"path", "old_string", "new_string"},
				},
			},
		},
	}
}

// GetToolFunc returns the function to execute for a given tool name.
func GetToolFunc(name string) ToolFunc {
	switch name {
	case "read_file":
		return readFile
	case "write_file":
		return writeFile
	case "edit_file":
		return editFile
	default:
		return nil
	}
}

func readFile(args map[string]any) ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResult{Success: false, Raw: "missing required argument: path"}
	}

	slog.Debug("reading file", "path", path)

	content, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("file read failed", "path", path, "err", err)
		return ToolResult{Success: false, Raw: err.Error()}
	}

	lines := strings.Count(string(content), "\n") + 1
	slog.Info("file read success", "path", path, "lines", lines)
	return ToolResult{Success: true, Raw: string(content)}
}

func writeFile(args map[string]any) ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResult{Success: false, Raw: "missing required argument: path"}
	}
	content, ok := args["content"].(string)
	if !ok {
		return ToolResult{Success: false, Raw: "missing required argument: content"}
	}

	slog.Debug("writing file", "path", path)

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		slog.Warn("file write failed", "path", path, "err", err)
		return ToolResult{Success: false, Raw: err.Error()}
	}

	slog.Info("file write success", "path", path)
	return ToolResult{Success: true, Raw: "File written successfully."}
}

func editFile(args map[string]any) ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ToolResult{Success: false, Raw: "missing required argument: path"}
	}
	oldStr, ok := args["old_string"].(string)
	if !ok {
		return ToolResult{Success: false, Raw: "missing required argument: old_string"}
	}
	newStr, ok := args["new_string"].(string)
	if !ok {
		return ToolResult{Success: false, Raw: "missing required argument: new_string"}
	}

	slog.Debug("editing file", "path", path)

	content, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("file read failed during edit", "path", path, "err", err)
		return ToolResult{Success: false, Raw: err.Error()}
	}

	text := string(content)
	if !strings.Contains(text, oldStr) {
		slog.Warn("old_string not found in file", "path", path)
		return ToolResult{Success: false, Raw: "old_string not found in file"}
	}

	newContent := strings.Replace(text, oldStr, newStr, 1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		slog.Warn("file write failed during edit", "path", path, "err", err)
		return ToolResult{Success: false, Raw: err.Error()}
	}

	slog.Info("file edit success", "path", path)
	return ToolResult{Success: true, Raw: "File edited successfully."}
}

// ParseToolArgs parses a JSON arguments string into a map.
func ParseToolArgs(jsonStr string) (map[string]any, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &args); err != nil {
		return nil, err
	}
	return args, nil
}
