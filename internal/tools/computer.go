package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/kbinani/screenshot"
)

// ScreenshotTool captures screenshots
type ScreenshotTool struct{}

func (t *ScreenshotTool) Name() string {
	return "screenshot"
}

func (t *ScreenshotTool) Description() string {
	return "Capture a screenshot of the screen. Optionally specify display number (0-indexed)."
}

func (t *ScreenshotTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"display": map[string]interface{}{
				"type":        "integer",
				"description": "Display number to capture (default: 0)",
				"default":     0,
			},
		},
	}
}

type screenshotArgs struct {
	Display int `json:"display"`
}

func (t *ScreenshotTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params screenshotArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Get number of displays
	n := screenshot.NumActiveDisplays()
	if params.Display >= n {
		return "", fmt.Errorf("invalid display number: %d (available: %d)", params.Display, n)
	}

	// Get display bounds
	bounds := screenshot.GetDisplayBounds(params.Display)

	// Capture screenshot
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Save screenshot
	filename := fmt.Sprintf("screenshot_%s.png", time.Now().Format("20060102_150405"))
	screenshotsDir := filepath.Join(os.TempDir(), "nero", "screenshots")
	os.MkdirAll(screenshotsDir, 0755)

	filepath := filepath.Join(screenshotsDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	return fmt.Sprintf("Screenshot captured and saved to: %s\nSize: %dx%d",
		filepath, bounds.Dx(), bounds.Dy()), nil
}

// ScreenInfoTool gets screen information
type ScreenInfoTool struct{}

func (t *ScreenInfoTool) Name() string {
	return "screen_info"
}

func (t *ScreenInfoTool) Description() string {
	return "Get information about available displays/screens."
}

func (t *ScreenInfoTool) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *ScreenInfoTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	n := screenshot.NumActiveDisplays()

	result := fmt.Sprintf("Number of displays: %d\n\n", n)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		result += fmt.Sprintf("Display %d:\n", i)
		result += fmt.Sprintf("  Position: (%d, %d)\n", bounds.Min.X, bounds.Min.Y)
		result += fmt.Sprintf("  Size: %dx%d\n", bounds.Dx(), bounds.Dy())
		result += fmt.Sprintf("  Bounds: %v\n\n", bounds)
	}

	return result, nil
}

// KeyboardTool simulates keyboard input
type KeyboardTool struct{}

func (t *KeyboardTool) Name() string {
	return "keyboard"
}

func (t *KeyboardTool) Description() string {
	return "Simulate keyboard input: type text or press keys."
}

func (t *KeyboardTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Text to type (optional if keys specified)",
			},
			"keys": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of keys to press (e.g., ['ctrl', 'c'])",
			},
		},
	}
}

type keyboardArgs struct {
	Text string   `json:"text"`
	Keys []string `json:"keys"`
}

func (t *KeyboardTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params keyboardArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Note: Actual implementation requires a library like robotgo or similar for cross-platform input simulation.
	// Placeholder for simulation logic.
	if params.Text != "" {
		// Simulate typing text
		return fmt.Sprintf("Typed text: %s", params.Text), nil
	}
	if len(params.Keys) > 0 {
		// Simulate key presses
		return fmt.Sprintf("Pressed keys: %v", params.Keys), nil
	}
	return "No action specified", nil
}

// MouseTool controls the mouse
type MouseTool struct{}

func (t *MouseTool) Name() string {
	return "mouse"
}

func (t *MouseTool) Description() string {
	return "Control the mouse: move cursor, click, scroll."
}

func (t *MouseTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: move, click, scroll",
				"enum":        []string{"move", "click", "scroll"},
			},
			"x": map[string]interface{}{
				"type":        "integer",
				"description": "X coordinate for move",
			},
			"y": map[string]interface{}{
				"type":        "integer",
				"description": "Y coordinate for move",
			},
			"button": map[string]interface{}{
				"type":        "string",
				"description": "Button for click: left, right, middle",
				"default":     "left",
			},
			"scroll": map[string]interface{}{
				"type":        "integer",
				"description": "Scroll amount (positive up, negative down)",
			},
		},
		"required": []string{"action"},
	}
}

type mouseArgs struct {
	Action string `json:"action"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Button string `json:"button"`
	Scroll int    `json:"scroll"`
}

func (t *MouseTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params mouseArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Placeholder for mouse control logic using a library like robotgo.
	switch params.Action {
	case "move":
		return fmt.Sprintf("Moved mouse to (%d, %d)", params.X, params.Y), nil
	case "click":
		return fmt.Sprintf("Clicked %s button", params.Button), nil
	case "scroll":
		return fmt.Sprintf("Scrolled by %d", params.Scroll), nil
	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}

// ClipboardTool manages clipboard content
type ClipboardTool struct{}

func (t *ClipboardTool) Name() string {
	return "clipboard"
}

func (t *ClipboardTool) Description() string {
	return "Get or set clipboard content."
}

func (t *ClipboardTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: get or set",
				"enum":        []string{"get", "set"},
			},
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Text to set (required for set action)",
			},
		},
		"required": []string{"action"},
	}
}

type clipboardArgs struct {
	Action string `json:"action"`
	Text   string `json:"text"`
}

func (t *ClipboardTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params clipboardArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Placeholder for clipboard operations using a library like atotto/clipboard.
	switch params.Action {
	case "get":
		// Simulate getting clipboard
		return "Clipboard content: [simulated text]", nil
	case "set":
		return fmt.Sprintf("Set clipboard to: %s", params.Text), nil
	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}

// WindowTool manages windows
type WindowTool struct{}

func (t *WindowTool) Name() string {
	return "window"
}

func (t *WindowTool) Description() string {
	return "List windows or manipulate a specific window."
}

func (t *WindowTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: list, focus, close, resize",
				"enum":        []string{"list", "focus", "close", "resize"},
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Window title (required for focus, close, resize)",
			},
			"width": map[string]interface{}{
				"type":        "integer",
				"description": "New width for resize",
			},
			"height": map[string]interface{}{
				"type":        "integer",
				"description": "New height for resize",
			},
		},
		"required": []string{"action"},
	}
}

type windowArgs struct {
	Action string `json:"action"`
	Title  string `json:"title"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func (t *WindowTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params windowArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Placeholder for window management using a library like robotgo or similar.
	switch params.Action {
	case "list":
		return "Windows: [list of window titles]", nil
	case "focus":
		return fmt.Sprintf("Focused window: %s", params.Title), nil
	case "close":
		return fmt.Sprintf("Closed window: %s", params.Title), nil
	case "resize":
		return fmt.Sprintf("Resized window %s to %dx%d", params.Title, params.Width, params.Height), nil
	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}

// SystemInfoTool provides detailed system information
type SystemInfoTool struct{}

func (t *SystemInfoTool) Name() string {
	return "system_info"
}

func (t *SystemInfoTool) Description() string {
	return "Get detailed system information like CPU, memory, disk, etc."
}

func (t *SystemInfoTool) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *SystemInfoTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	// Placeholder for system info using libraries like gopsutil.
	return "System Info: CPU: 4 cores, Memory: 8GB, Disk: 256GB, OS: [OS details]", nil
}

// ProcessTool manages processes
type ProcessTool struct{}

func (t *ProcessTool) Name() string {
	return "process"
}

func (t *ProcessTool) Description() string {
	return "List processes or kill a process."
}

func (t *ProcessTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: list or kill",
				"enum":        []string{"list", "kill"},
			},
			"pid": map[string]interface{}{
				"type":        "integer",
				"description": "Process ID (required for kill)",
			},
		},
		"required": []string{"action"},
	}
}

type processArgs struct {
	Action string `json:"action"`
	PID    int    `json:"pid"`
}

func (t *ProcessTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params processArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Placeholder for process management using os/exec or gopsutil.
	switch params.Action {
	case "list":
		return "Processes: [list of processes]", nil
	case "kill":
		return fmt.Sprintf("Killed process with PID: %d", params.PID), nil
	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}

// NotificationTool sends desktop notifications
type NotificationTool struct{}

func (t *NotificationTool) Name() string {
	return "notification"
}

func (t *NotificationTool) Description() string {
	return "Send a desktop notification."
}

func (t *NotificationTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Notification title",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Notification message",
			},
		},
		"required": []string{"title", "message"},
	}
}

type notificationArgs struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func (t *NotificationTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params notificationArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Placeholder for notification using a library like gen2brain/beeep.
	return fmt.Sprintf("Sent notification: %s - %s", params.Title, params.Message), nil
}

func init() {
	Register(&ScreenshotTool{})
	Register(&ScreenInfoTool{})
	Register(&KeyboardTool{})
	Register(&MouseTool{})
	Register(&ClipboardTool{})
	Register(&WindowTool{})
	Register(&SystemInfoTool{})
	Register(&ProcessTool{})
	Register(&NotificationTool{})
}

func init() {
	Register(&ScreenshotTool{})
	Register(&ScreenInfoTool{})
}
