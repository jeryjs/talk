package computer

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"nero/internal/config"

	"github.com/kbinani/screenshot"
)

// ScreenshotTool captures screenshots
type ScreenshotTool struct{}

func (t *ScreenshotTool) Name() string {
	return "take_screenshot"
}

func (t *ScreenshotTool) Description() string {
	return "Take a screenshot of the entire screen or a specific display. Returns the path to the saved screenshot."
}

func (t *ScreenshotTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"display": map[string]interface{}{
				"type":        "integer",
				"description": "Display number to capture (0 for primary, 1+ for additional displays)",
				"default":     0,
			},
			"save": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to save the screenshot to disk",
				"default":     true,
			},
		},
	}
}

type ScreenshotArgs struct {
	Display int  `json:"display"`
	Save    bool `json:"save"`
}

func (t *ScreenshotTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params ScreenshotArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Get number of displays
	n := screenshot.NumActiveDisplays()
	if params.Display >= n {
		return "", fmt.Errorf("display %d not found (only %d displays available)", params.Display, n)
	}

	// Get display bounds
	bounds := screenshot.GetDisplayBounds(params.Display)

	// Capture screenshot
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", fmt.Errorf("failed to capture screenshot: %w", err)
	}

	if !params.Save {
		return fmt.Sprintf("Screenshot captured from display %d (%dx%d)",
			params.Display, bounds.Dx(), bounds.Dy()), nil
	}

	// Save to file
	cfg := config.Get()
	screenshotsDir := filepath.Join(cfg.LogsDir, "screenshots")
	os.MkdirAll(screenshotsDir, 0755)

	filename := filepath.Join(screenshotsDir,
		fmt.Sprintf("screenshot_%s.png", time.Now().Format("20060102_150405")))

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return "", fmt.Errorf("failed to encode screenshot: %w", err)
	}

	return fmt.Sprintf("Screenshot saved to: %s (%dx%d)",
		filename, bounds.Dx(), bounds.Dy()), nil
}

// GetScreenInfoTool gets information about displays
type GetScreenInfoTool struct{}

func (t *GetScreenInfoTool) Name() string {
	return "get_screen_info"
}

func (t *GetScreenInfoTool) Description() string {
	return "Get information about available displays (number of displays, resolutions, etc.)"
}

func (t *GetScreenInfoTool) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *GetScreenInfoTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	n := screenshot.NumActiveDisplays()

	result := fmt.Sprintf("Number of active displays: %d\n\n", n)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		result += fmt.Sprintf("Display %d:\n", i)
		result += fmt.Sprintf("  Resolution: %dx%d\n", bounds.Dx(), bounds.Dy())
		result += fmt.Sprintf("  Position: (%d, %d)\n", bounds.Min.X, bounds.Min.Y)
		result += "\n"
	}

	return result, nil
}

// CaptureScreen captures the screen and returns the image
func CaptureScreen(displayNum int) (image.Image, error) {
	n := screenshot.NumActiveDisplays()
	if displayNum >= n {
		return nil, fmt.Errorf("display %d not found", displayNum)
	}

	bounds := screenshot.GetDisplayBounds(displayNum)
	return screenshot.CaptureRect(bounds)
}

// CaptureAllScreens captures all screens
func CaptureAllScreens() ([]image.Image, error) {
	n := screenshot.NumActiveDisplays()
	images := make([]image.Image, n)

	for i := 0; i < n; i++ {
		img, err := CaptureScreen(i)
		if err != nil {
			return nil, err
		}
		images[i] = img
	}

	return images, nil
}
