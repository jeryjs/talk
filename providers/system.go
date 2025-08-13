package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Handle system interactions
type SystemProvider struct{}

// Create a new system provider
func NewSystemProvider() *SystemProvider {
	return &SystemProvider{}
}

// Execute a system command
func (s *SystemProvider) RunCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Open an application
func (s *SystemProvider) OpenApp(appName string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", appName)
	case "darwin":
		cmd = exec.Command("open", "-a", appName)
	case "linux":
		cmd = exec.Command("xdg-open", appName)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// Read a file and return its content
func (s *SystemProvider) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Write content to a file
func (s *SystemProvider) WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// List files in a directory
func (s *SystemProvider) ListDirectory(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files, nil
}

// Return current working directory
func (s *SystemProvider) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

// Change the working directory
func (s *SystemProvider) ChangeDirectory(path string) error {
	return os.Chdir(path)
}

// Return system information
func (s *SystemProvider) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cpu":  runtime.NumCPU(),
	}
}

// Handle conversation memory
type MemoryProvider struct {
	memoryFile string
	memories   []Memory
}

// Represent a stored memory
type Memory struct {
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp"`
	Type      string                 `json:"type"`
	Content   string                 `json:"content"`
	Emotions  string                 `json:"emotions"`
	Context   map[string]interface{} `json:"context"`
	Tags      []string               `json:"tags"`
}

// Create a new memory provider
func NewMemoryProvider() *MemoryProvider {
	homeDir, _ := os.UserHomeDir()
	memoryFile := filepath.Join(homeDir, ".nero", "memories.json")

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(memoryFile), 0755)

	mp := &MemoryProvider{
		memoryFile: memoryFile,
		memories:   make([]Memory, 0),
	}

	mp.loadMemories()
	return mp
}

// Save a new memory
func (m *MemoryProvider) StoreMemory(memory Memory) error {
	m.memories = append(m.memories, memory)
	return m.saveMemories()
}

// Retrieve memories based on criteria
func (m *MemoryProvider) GetMemories(limit int, tags ...string) []Memory {
	var filtered []Memory

	if len(tags) == 0 {
		// Return recent memories
		start := len(m.memories) - limit
		if start < 0 {
			start = 0
		}
		return m.memories[start:]
	}

	// Filter by tags
	for _, memory := range m.memories {
		for _, tag := range tags {
			for _, memTag := range memory.Tags {
				if memTag == tag {
					filtered = append(filtered, memory)
					break
				}
			}
		}
		if len(filtered) >= limit {
			break
		}
	}

	return filtered
}

// Find memories containing specific content
func (m *MemoryProvider) SearchMemories(query string, limit int) []Memory {
	var results []Memory
	query = strings.ToLower(query)

	for _, memory := range m.memories {
		if strings.Contains(strings.ToLower(memory.Content), query) {
			results = append(results, memory)
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

// Load memories from file
func (m *MemoryProvider) loadMemories() error {
	data, err := os.ReadFile(m.memoryFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No memories yet
		}
		return err
	}

	return json.Unmarshal(data, &m.memories)
}

// Save memories to file
func (m *MemoryProvider) saveMemories() error {
	data, err := json.MarshalIndent(m.memories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.memoryFile, data, 0644)
}

// Generate expressions using AI
type KaomojiProvider struct {
	aiProvider AIProvider
}

// Create a new kaomoji provider
func NewKaomojiProvider(ai AIProvider) *KaomojiProvider {
	return &KaomojiProvider{
		aiProvider: ai,
	}
}

// Create a kaomoji based on response content
func (k *KaomojiProvider) GenerateKaomoji(ctx context.Context, response string, mood string) (string, error) {
	if len(strings.Fields(response)) > 20 {
		return "", nil // Don't generate for long responses
	}

	prompt := fmt.Sprintf(`Generate a single appropriate kaomoji (Japanese emoticon) for this response. Consider the mood: %s
Response: "%s"

Rules:
- Return ONLY the kaomoji, nothing else
- Make it match the emotion and tone
- Keep it simple and expressive
- Examples: (´∀｀) (╯°□°）╯ (˘▾˘~) ♪(´▽｀) (￣ω￣) 

Kaomoji:`, mood, response)

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	options := &ChatOptions{
		Temperature: 0.8,
		MaxTokens:   20,
	}

	resp, err := k.aiProvider.Chat(ctx, messages, options)
	if err != nil {
		return "", err
	}

	// Clean up the response
	kaomoji := strings.TrimSpace(resp.Content)
	kaomoji = strings.Trim(kaomoji, "\"'`")

	return kaomoji, nil
}
