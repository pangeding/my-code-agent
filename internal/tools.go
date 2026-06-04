package internal

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Success bool
	Raw     string
}

// ConfirmFunc 是用户确认回调函数签名。
// command: 待执行的命令, riskType: 命中的风险类别
// 返回 true 表示用户确认执行，false 表示拒绝。
type ConfirmFunc func(command, riskType string) bool

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
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "bash",
				"description": "Execute a bash command and return its output (stdout and stderr).",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"command": map[string]string{
							"type":        "string",
							"description": "The bash command to execute.",
						},
					},
					"required": []string{"command"},
				},
			},
		},
	}
}

// GetToolFunc returns the function to execute for a given tool name.
// 注意：bash 工具应通过 GetBashToolFunc 获取，以支持安全确认回调。
func GetToolFunc(name string) ToolFunc {
	switch name {
	case "read_file":
		return readFile
	case "write_file":
		return writeFile
	case "edit_file":
		return editFile
	case "bash":
		// 返回一个占位函数，实际执行应使用 GetBashToolFunc
		return func(args map[string]any) ToolResult {
			return ToolResult{Success: false, Raw: "bash tool requires confirmation callback, use GetBashToolFunc"}
		}
	default:
		return nil
	}
}

// GetBashToolFunc returns a ToolFunc backed by a BashTool with the given confirm callback.
func GetBashToolFunc(confirm ConfirmFunc) ToolFunc {
	tool := NewBashTool(confirm)
	return tool.Execute
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

// BashTool 封装 bash 命令执行工具，支持安全确认回调。
type BashTool struct {
	Confirm ConfirmFunc // 用户确认回调，为 nil 时默认拒绝
}

// NewBashTool 创建 BashTool 实例。
func NewBashTool(confirm ConfirmFunc) *BashTool {
	return &BashTool{Confirm: confirm}
}

// Execute 执行 bash 命令。
func (b *BashTool) Execute(args map[string]any) ToolResult {
	command, ok := args["command"].(string)
	if !ok {
		return ToolResult{Success: false, Raw: "missing required argument: command"}
	}

	slog.Debug("executing bash command", "command", command)

	// 危险命令检查
	if dangerous, reason := isDangerousCommand(command); dangerous {
		slog.Warn("dangerous command detected", "command", command)
		// 需要用户确认
		if b.Confirm == nil || !b.Confirm(command, reason) {
			return ToolResult{Success: false, Raw: "用户已拒绝执行该命令"}
		}
		slog.Info("dangerous command confirmed by user", "command", command)
	}

	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("command execution failed", "command", command, "err", err)
		return ToolResult{Success: false, Raw: string(output) + "\nError: " + err.Error()}
	}

	slog.Info("command execution success", "command", command)
	return ToolResult{Success: true, Raw: string(output)}
}

// ParseToolArgs parses a JSON arguments string into a map.
func ParseToolArgs(jsonStr string) (map[string]any, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &args); err != nil {
		return nil, err
	}
	return args, nil
}

// dangerousPatterns 危险命令正则表达式黑名单。
var dangerousPatterns = []*regexp.Regexp{
	// 危险删除操作
	regexp.MustCompile(`(?i)\brm\s+(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*)\s+`),
	regexp.MustCompile(`(?i)\brm\s+-rf\s+/$`),
	regexp.MustCompile(`(?i)\brm\s+-rf\s+/(\s|$)`),
	regexp.MustCompile(`(?i)\brm\s+-rf\s+~`),
	regexp.MustCompile(`(?i)\bshred\b`),

	// 磁盘格式化/写操作
	regexp.MustCompile(`(?i)\bmkfs\.`),
	regexp.MustCompile(`(?i)\bmkfs\b`),
	regexp.MustCompile(`(?i)\bfdisk\b`),
	regexp.MustCompile(`(?i)\bdd\s+if=/dev/zero\b`),
	regexp.MustCompile(`(?i)\bdd\s+if=/dev/null\b`),

	// 危险权限修改
	regexp.MustCompile(`(?i)\bchmod\s+-?777\b`),
	regexp.MustCompile(`(?i)\bchmod\s+-?776\b`),
	regexp.MustCompile(`(?i)\bchown\s+-R\b`),

	// 网络/防火墙操作
	regexp.MustCompile(`(?i)\biptables\s+-F\b`),
	regexp.MustCompile(`(?i)\bufw\s+disable\b`),
	regexp.MustCompile(`(?i)\bsystemctl\s+stop\b`),

	// 危险信号操作
	regexp.MustCompile(`(?i)\bkill\s+-9\b`),

	// 覆盖系统关键文件
	regexp.MustCompile(`(?i)\bmv\s+.+\s+/etc/passwd\b`),
	regexp.MustCompile(`(?i)\bmv\s+.+\s+/etc/shadow\b`),
	regexp.MustCompile(`(?i)\bmv\s+.+\s+/etc/sudoers\b`),

	// fork bomb
	regexp.MustCompile(`:\{\s*:\|:&\s*\};:`),
}

// isDangerousCommand 检查命令是否在黑名单中。
func isDangerousCommand(cmd string) (bool, string) {
	// 去除前后空白
	trimmed := strings.TrimSpace(cmd)

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(trimmed) {
			return true, "command blocked by safety filter: " + pattern.String()
		}
	}

	return false, ""
}
