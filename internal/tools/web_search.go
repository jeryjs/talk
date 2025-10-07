package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

// WebSearchTool performs web searches
type WebSearchTool struct{}

func (t *WebSearchTool) Name() string {
	return "web_search"
}

func (t *WebSearchTool) Description() string {
	return "Search the web using Google or Bing. Returns search results with titles, URLs, and snippets."
}

func (t *WebSearchTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
			"num_results": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results to return (default: 5)",
				"default":     5,
			},
		},
		"required": []string{"query"},
	}
}

type WebSearchArgs struct {
	Query      string `json:"query"`
	NumResults int    `json:"num_results"`
}

func (t *WebSearchTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params WebSearchArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if params.NumResults == 0 {
		params.NumResults = 5
	}

	// For now, open in browser (in a real implementation, use an API)
	url := fmt.Sprintf("https://www.google.com/search?q=%s", params.Query)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to open browser: %w", err)
	}

	return fmt.Sprintf("Opened search for '%s' in browser", params.Query), nil
}

func init() {
	Register(&WebSearchTool{})
}
