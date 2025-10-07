package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ShellExecuteTool executes shell commands
type ShellExecuteTool struct{}

func (t *ShellExecuteTool) Name() string {
	return "execute_shell"
}

func (t *ShellExecuteTool) Description() string {
	return "Execute a shell command and return the output. Use with caution."
}

func (t *ShellExecuteTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for the command (optional)",
			},
		},
		"required": []string{"command"},
	}
}

type ShellExecuteArgs struct {
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir"`
}

func (t *ShellExecuteTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params ShellExecuteArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	var cmd *exec.Cmd

	// Choose shell based on OS
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "powershell", "-Command", params.Command)
	default:
		cmd = exec.CommandContext(ctx, "sh", "-c", params.Command)
	}

	if params.WorkingDir != "" {
		cmd.Dir = params.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// GetTerminalContextTool gets current terminal context
type GetTerminalContextTool struct{}

func (t *GetTerminalContextTool) Name() string {
	return "get_terminal_context"
}

func (t *GetTerminalContextTool) Description() string {
	return "Get information about the current terminal session (working directory, recent commands, etc.)"
}

func (t *GetTerminalContextTool) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *GetTerminalContextTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var info strings.Builder

	// Get current directory
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", "Get-Location")
	} else {
		cmd = exec.Command("pwd")
	}

	output, err := cmd.Output()
	if err == nil {
		info.WriteString(fmt.Sprintf("Current Directory: %s\n", strings.TrimSpace(string(output))))
	}

	// Get OS info
	info.WriteString(fmt.Sprintf("Operating System: %s\n", runtime.GOOS))
	info.WriteString(fmt.Sprintf("Architecture: %s\n", runtime.GOARCH))

	// Get environment variables (selected ones)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", "$env:USERNAME")
	} else {
		cmd = exec.Command("whoami")
	}

	if output, err := cmd.Output(); err == nil {
		info.WriteString(fmt.Sprintf("User: %s\n", strings.TrimSpace(string(output))))
	}

	return info.String(), nil
}

func init() {
	Register(&ShellExecuteTool{})
	Register(&GetTerminalContextTool{})
}
