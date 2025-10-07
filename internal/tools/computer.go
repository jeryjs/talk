package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
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

// AnalyzeScreenTool captures and analyzes screen content
type AnalyzeScreenTool struct{}

func (t *AnalyzeScreenTool) Name() string {
	return "analyze_screen"
}

func (t *AnalyzeScreenTool) Description() string {
	return "Capture a screenshot and return basic analysis of the screen content."
}

func (t *AnalyzeScreenTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"display": map[string]interface{}{
				"type":        "integer",
				"description": "Display number to analyze (default: 0)",
				"default":     0,
			},
		},
	}
}

func (t *AnalyzeScreenTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
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

	// Basic analysis
	analysis := analyzeImage(img)

	return fmt.Sprintf("Screen Analysis for Display %d:\n%s", params.Display, analysis), nil
}

func analyzeImage(img *image.RGBA) string {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Sample colors from different regions
	centerColor := img.At(width/2, height/2)
	topLeftColor := img.At(width/4, height/4)
	bottomRightColor := img.At(3*width/4, 3*height/4)

	result := fmt.Sprintf("Resolution: %dx%d\n", width, height)
	result += fmt.Sprintf("Center color: %v\n", centerColor)
	result += fmt.Sprintf("Top-left color: %v\n", topLeftColor)
	result += fmt.Sprintf("Bottom-right color: %v\n", bottomRightColor)
	result += "\nNote: For detailed content analysis, use vision-capable models with the screenshot."

	return result
}

func init() {
	Register(&ScreenshotTool{})
	Register(&ScreenInfoTool{})
	Register(&AnalyzeScreenTool{})
}
