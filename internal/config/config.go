package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// API Keys
	OpenAIKey     string `mapstructure:"openai_api_key"`
	AnthropicKey  string `mapstructure:"anthropic_api_key"`
	GoogleKey     string `mapstructure:"google_api_key"`
	OpenRouterKey string `mapstructure:"openrouter_api_key"`
	GroqKey       string `mapstructure:"groq_api_key"`

	// Default Settings
	Model       string  `mapstructure:"default_model"`
	Mood        string  `mapstructure:"default_mood"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`

	// Features
	EnableVoice       bool `mapstructure:"enable_voice"`
	EnableComputerUse bool `mapstructure:"enable_computer_use"`
	EnableMCP         bool `mapstructure:"enable_mcp"`
	EnableWebSearch   bool `mapstructure:"enable_web_search"`

	// Directories
	PluginsDir string `mapstructure:"plugins_dir"`
	ChatsDir   string `mapstructure:"chats_dir"`
	LogsDir    string `mapstructure:"logs_dir"`

	// MCP Servers
	MCPServers []MCPServer `mapstructure:"mcp_servers"`

	// Plugin Settings
	AutoReloadPlugins bool `mapstructure:"auto_reload_plugins"`
}

// MCPServer represents an MCP server configuration
type MCPServer struct {
	Name    string            `mapstructure:"name"`
	Command string            `mapstructure:"command"`
	Args    []string          `mapstructure:"args"`
	Env     map[string]string `mapstructure:"env"`
	Enabled bool              `mapstructure:"enabled"`
}

var current *Config

// Load loads the configuration from file
func Load(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Search config in .nero directory
		configDir := filepath.Join(home, ".nero")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default config
			if err := createDefaultConfig(); err != nil {
				return fmt.Errorf("failed to create default config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Unmarshal into struct
	current = &Config{}
	if err := viper.Unmarshal(current); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand environment variables in paths
	expandPaths(current)

	return nil
}

func setDefaults() {
	// Default model settings - Use Ollama by default (no API key needed)
	viper.SetDefault("default_model", "llama3.2")
	viper.SetDefault("default_mood", "helpful")
	viper.SetDefault("max_tokens", 4096)
	viper.SetDefault("temperature", 0.7)

	// Features
	viper.SetDefault("enable_voice", true)
	viper.SetDefault("enable_computer_use", true)
	viper.SetDefault("enable_mcp", true)
	viper.SetDefault("enable_web_search", true)

	// Directories
	home, _ := os.UserHomeDir()
	viper.SetDefault("plugins_dir", filepath.Join(home, ".nero", "plugins"))
	viper.SetDefault("chats_dir", filepath.Join(home, ".nero", "chats"))
	viper.SetDefault("logs_dir", filepath.Join(home, ".nero", "logs"))

	// Plugin settings
	viper.SetDefault("auto_reload_plugins", true)
}

func createDefaultConfig() error {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".nero", "config.yaml")
	}

	defaultConfig := `# Nero Configuration

# API Keys (optional - Ollama is used by default and needs no API key)
# openai_api_key: "sk-..."
# anthropic_api_key: "sk-ant-..."
# google_api_key: "..."
# groq_api_key: "gsk_..."  # Ultra-fast inference!

# Default Settings
default_model: "llama3.2"  # Local Ollama model (no API key needed)
# Or use cloud models if you have API keys:
# default_model: "op/gpt-4o"           # OpenAI
# default_model: "gq/llama-3.3-70b"   # Groq (ultra-fast!)
# default_model: "g/gemini-2.5-pro"   # Google
# default_model: "c/claude-sonnet-4.5"# Anthropic

# Shortcuts: op/ (openai), g/ (gemini), c/ (claude), gq/ (groq), ol/ (ollama), or/ (openrouter)

default_mood: "helpful"
max_tokens: 4096
temperature: 0.7

# Features
enable_voice: true
enable_computer_use: true
enable_mcp: true
enable_web_search: true

# Plugins
auto_reload_plugins: true

# MCP Servers
mcp_servers:
  # - name: "filesystem"
  #   command: "npx"
  #   args: ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"]
  #   enabled: true
  
  # - name: "github"
  #   command: "npx"
  #   args: ["-y", "@modelcontextprotocol/server-github"]
  #   env:
  #     GITHUB_TOKEN: "ghp_..."
  #   enabled: true
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0600)
}

func expandPaths(cfg *Config) {
	home, _ := os.UserHomeDir()

	cfg.PluginsDir = expandPath(cfg.PluginsDir, home)
	cfg.ChatsDir = expandPath(cfg.ChatsDir, home)
	cfg.LogsDir = expandPath(cfg.LogsDir, home)

	// Create directories if they don't exist
	os.MkdirAll(cfg.PluginsDir, 0755)
	os.MkdirAll(cfg.ChatsDir, 0755)
	os.MkdirAll(cfg.LogsDir, 0755)
}

func expandPath(path, home string) string {
	if len(path) > 0 && path[0] == '~' {
		return filepath.Join(home, path[1:])
	}
	return path
}

// Get returns the current configuration
func Get() *Config {
	if current == nil {
		// Load with defaults if not loaded
		Load("")
	}
	return current
}

// Set updates a configuration value
func Set(key string, value interface{}) error {
	viper.Set(key, value)

	// Save to file
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Reload config
	return Load("")
}
