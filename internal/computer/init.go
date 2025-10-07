package computer

import (
	"nero/internal/tools"
)

func init() {
	// Register computer use tools
	tools.Register(&ScreenshotTool{})
	tools.Register(&GetScreenInfoTool{})
}
