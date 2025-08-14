package capabilities

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

type Manifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Dependencies []string          `json:"dependencies"`
	Capabilities []string          `json:"capabilities"`
	Metadata     map[string]string `json:"metadata"`
}

type Capability interface {
	Initialize() error
	Shutdown() error
	GetName() string
	GetVersion() string
}

type Loader struct {
	capabilities  map[string]Capability
	manifests     map[string]*Manifest
	mutex         sync.RWMutex
	extensionsDir string
}

func NewLoader(extensionsDir string) *Loader {
	return &Loader{
		capabilities:  make(map[string]Capability),
		manifests:     make(map[string]*Manifest),
		extensionsDir: extensionsDir,
	}
}

func (l *Loader) LoadAll() error {
	return filepath.Walk(l.extensionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() != filepath.Base(l.extensionsDir) {
			manifestPath := filepath.Join(path, "manifest.json")
			if _, err := os.Stat(manifestPath); err == nil {
				return l.LoadExtension(path)
			}
		}
		return nil
	})
}

func (l *Loader) LoadExtension(extensionPath string) error {
	manifestPath := filepath.Join(extensionPath, "manifest.json")
	manifest, err := l.loadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Check dependencies
	if err := l.checkDependencies(manifest); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Load capability (if it's a Go plugin)
	pluginPath := filepath.Join(extensionPath, manifest.Name+".so")
	if _, err := os.Stat(pluginPath); err == nil {
		return l.loadGoPlugin(pluginPath, manifest)
	}

	// For built-in capabilities, register them here
	l.mutex.Lock()
	l.manifests[manifest.Name] = manifest
	l.mutex.Unlock()

	return nil
}

func (l *Loader) UnloadExtension(name string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if cap, exists := l.capabilities[name]; exists {
		if err := cap.Shutdown(); err != nil {
			return err
		}
		delete(l.capabilities, name)
	}

	delete(l.manifests, name)
	return nil
}

func (l *Loader) GetCapability(name string) (Capability, bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	cap, exists := l.capabilities[name]
	return cap, exists
}

func (l *Loader) ListCapabilities() []string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	names := make([]string, 0, len(l.capabilities))
	for name := range l.capabilities {
		names = append(names, name)
	}
	return names
}

func (l *Loader) loadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (l *Loader) checkDependencies(manifest *Manifest) error {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for _, dep := range manifest.Dependencies {
		if _, exists := l.capabilities[dep]; !exists {
			return fmt.Errorf("missing dependency: %s", dep)
		}
	}
	return nil
}

func (l *Loader) loadGoPlugin(pluginPath string, manifest *Manifest) error {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return err
	}

	// Look for NewCapability function
	sym, err := p.Lookup("NewCapability")
	if err != nil {
		return err
	}

	newFunc, ok := sym.(func() Capability)
	if !ok {
		return fmt.Errorf("invalid NewCapability function signature")
	}

	cap := newFunc()
	if err := cap.Initialize(); err != nil {
		return err
	}

	l.mutex.Lock()
	l.capabilities[manifest.Name] = cap
	l.manifests[manifest.Name] = manifest
	l.mutex.Unlock()

	return nil
}
