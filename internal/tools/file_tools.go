package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileReadTool reads file contents
type FileReadTool struct{}

func (t *FileReadTool) Name() string {
	return "read_file"
}

func (t *FileReadTool) Description() string {
	return "Read the contents of a file. Returns the file content as a string."
}

func (t *FileReadTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute or relative path to the file",
			},
		},
		"required": []string{"path"},
	}
}

type FileReadArgs struct {
	Path string `json:"path"`
}

func (t *FileReadTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params FileReadArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	content, err := os.ReadFile(params.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// FileWriteTool writes to a file
type FileWriteTool struct{}

func (t *FileWriteTool) Name() string {
	return "write_file"
}

func (t *FileWriteTool) Description() string {
	return "Write content to a file. Creates the file if it doesn't exist, overwrites if it does."
}

func (t *FileWriteTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute or relative path to the file",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

type FileWriteArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (t *FileWriteTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params FileWriteArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(params.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(params.Path, []byte(params.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(params.Content), params.Path), nil
}

// FileListTool lists directory contents
type FileListTool struct{}

func (t *FileListTool) Name() string {
	return "list_directory"
}

func (t *FileListTool) Description() string {
	return "List the contents of a directory. Returns a list of files and subdirectories."
}

func (t *FileListTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute or relative path to the directory",
			},
		},
		"required": []string{"path"},
	}
}

type FileListArgs struct {
	Path string `json:"path"`
}

func (t *FileListTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params FileListArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	entries, err := os.ReadDir(params.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	result := fmt.Sprintf("Contents of %s:\n", params.Path)
	for _, entry := range entries {
		if entry.IsDir() {
			result += fmt.Sprintf("  📁 %s/\n", entry.Name())
		} else {
			result += fmt.Sprintf("  📄 %s\n", entry.Name())
		}
	}

	return result, nil
}

func init() {
	Register(&FileReadTool{})
	Register(&FileWriteTool{})
	Register(&FileListTool{})
}
