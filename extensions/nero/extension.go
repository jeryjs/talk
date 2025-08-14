package extensions

import (
	"encoding/json"
	"fmt"
)

type NeroManifest struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Commands    []CommandDef      `json:"commands"`
	Resources   []ResourceDef     `json:"resources"`
	Config      map[string]string `json:"config"`
}

type CommandDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Essential   bool   `json:"essential"`
}

type ResourceDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type ConfigSpin string

const (
	SpinFull ConfigSpin = "full"
	SpinLite ConfigSpin = "lite"
)

type NeroConfig struct {
	Spin         ConfigSpin        `json:"spin"`
	Personality  string            `json:"personality"`
	VoiceEnabled bool              `json:"voice_enabled"`
	AutoSave     bool              `json:"auto_save"`
	Preferences  map[string]string `json:"preferences"`
}

type NeroExtension struct {
	config   *NeroConfig
	commands map[string]func([]string) (string, error)
}

func NewNeroExtension() *NeroExtension {
	return &NeroExtension{
		config: &NeroConfig{
			Spin:         SpinFull,
			Personality:  "tsundere",
			VoiceEnabled: true,
			AutoSave:     true,
			Preferences:  make(map[string]string),
		},
		commands: make(map[string]func([]string) (string, error)),
	}
}

func (ne *NeroExtension) Initialize() error {
	// Register essential commands with /prefix
	ne.commands["/config"] = ne.handleConfig
	ne.commands["/spin"] = ne.handleSpin
	ne.commands["/reset"] = ne.handleReset
	ne.commands["/status"] = ne.handleStatus
	ne.commands["/help"] = ne.handleHelp

	return nil
}

func (ne *NeroExtension) Shutdown() error {
	return nil
}

func (ne *NeroExtension) GetName() string {
	return "nero"
}

func (ne *NeroExtension) GetVersion() string {
	return "1.0.0"
}

func (ne *NeroExtension) GetManifest() NeroManifest {
	return NeroManifest{
		Name:        "nero",
		Version:     "1.0.0",
		Description: "Nero self-configuration and essential commands",
		Commands: []CommandDef{
			{Name: "config", Description: "Configure Nero settings", Usage: "@nero config [key] [value]", Essential: true},
			{Name: "spin", Description: "Switch between full/lite mode", Usage: "@nero spin [full|lite]", Essential: true},
			{Name: "reset", Description: "Reset behavioral state", Usage: "@nero reset", Essential: false},
			{Name: "status", Description: "Show Nero status", Usage: "@nero status", Essential: true},
			{Name: "help", Description: "Show available commands", Usage: "@nero help", Essential: true},
		},
		Resources: []ResourceDef{
			{Name: "config", Type: "settings", Description: "Configuration management", Color: "cyan"},
			{Name: "memory", Type: "storage", Description: "Context and behavioral memory", Color: "yellow"},
		},
		Config: map[string]string{
			"spin":          "full",
			"personality":   "tsundere",
			"voice_enabled": "true",
			"auto_save":     "true",
		},
	}
}

func (ne *NeroExtension) ExecuteCommand(command string, args []string) (string, error) {
	if handler, exists := ne.commands[command]; exists {
		return handler(args)
	}
	return "", fmt.Errorf("command not found: %s", command)
}

func (ne *NeroExtension) GetConfig() *NeroConfig {
	return ne.config
}

func (ne *NeroExtension) handleConfig(args []string) (string, error) {
	if len(args) == 0 {
		// Show current config
		data, err := json.MarshalIndent(ne.config, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	if len(args) == 1 {
		// Show specific config value
		key := args[0]
		switch key {
		case "spin":
			return string(ne.config.Spin), nil
		case "personality":
			return ne.config.Personality, nil
		case "voice_enabled":
			return fmt.Sprintf("%t", ne.config.VoiceEnabled), nil
		case "auto_save":
			return fmt.Sprintf("%t", ne.config.AutoSave), nil
		default:
			if val, exists := ne.config.Preferences[key]; exists {
				return val, nil
			}
			return "", fmt.Errorf("config key not found: %s", key)
		}
	}

	if len(args) == 2 {
		// Set config value
		key, value := args[0], args[1]
		switch key {
		case "spin":
			if value == "full" || value == "lite" {
				ne.config.Spin = ConfigSpin(value)
				return fmt.Sprintf("Spin mode set to: %s", value), nil
			}
			return "", fmt.Errorf("invalid spin mode: %s (use full or lite)", value)
		case "personality":
			ne.config.Personality = value
			return fmt.Sprintf("Personality set to: %s", value), nil
		case "voice_enabled":
			if value == "true" || value == "false" {
				ne.config.VoiceEnabled = value == "true"
				return fmt.Sprintf("Voice enabled: %s", value), nil
			}
			return "", fmt.Errorf("invalid boolean value: %s", value)
		case "auto_save":
			if value == "true" || value == "false" {
				ne.config.AutoSave = value == "true"
				return fmt.Sprintf("Auto save: %s", value), nil
			}
			return "", fmt.Errorf("invalid boolean value: %s", value)
		default:
			ne.config.Preferences[key] = value
			return fmt.Sprintf("Preference %s set to: %s", key, value), nil
		}
	}

	return "", fmt.Errorf("invalid config command format")
}

func (ne *NeroExtension) handleSpin(args []string) (string, error) {
	if len(args) == 0 {
		return string(ne.config.Spin), nil
	}

	spin := args[0]
	if spin != "full" && spin != "lite" {
		return "", fmt.Errorf("invalid spin mode: %s (use full or lite)", spin)
	}

	ne.config.Spin = ConfigSpin(spin)

	switch spin {
	case "full":
		return "🚀 Full mode activated - all capabilities enabled", nil
	case "lite":
		return "⚡ Lite mode activated - minimal resource usage", nil
	}

	return "", nil
}

func (ne *NeroExtension) handleReset(args []string) (string, error) {
	// Reset to defaults but keep user preferences
	preferences := ne.config.Preferences
	ne.config = &NeroConfig{
		Spin:         SpinFull,
		Personality:  "tsundere",
		VoiceEnabled: true,
		AutoSave:     true,
		Preferences:  preferences,
	}

	return "🔄 Behavioral state reset to defaults", nil
}

func (ne *NeroExtension) handleStatus(args []string) (string, error) {
	status := fmt.Sprintf(`Nero Status:
  Spin: %s
  Personality: %s  
  Voice: %t
  Auto-save: %t
  Preferences: %d custom settings`,
		ne.config.Spin,
		ne.config.Personality,
		ne.config.VoiceEnabled,
		ne.config.AutoSave,
		len(ne.config.Preferences))

	return status, nil
}

func (ne *NeroExtension) handleHelp(args []string) (string, error) {
	help := `@nero commands:
  config [key] [value] - Configure settings
  spin [full|lite]     - Switch modes  
  reset                - Reset to defaults
  status               - Show current status
  help                 - Show this help

Essential commands are always available.
Use @nero to reveal extension-specific commands.`

	return help, nil
}
