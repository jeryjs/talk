package kernel

import (
	"context"
	"sync"
)

// Manage capability registration and discovery
type Registry struct {
	capabilities map[string]CapabilityInfo
	dependencies map[string][]string
	mu           sync.RWMutex
}

// Define the interface all extensions must implement
type Capability interface {
	Initialize(runtime *Runtime) error
	Name() string
	Version() string
	Dependencies() []string
	Shutdown(ctx context.Context) error
}

// Contain metadata about a registered capability
type CapabilityInfo struct {
	Name         string
	Version      string
	Description  string
	Author       string
	Dependencies []string
	Instance     Capability
	Status       CapabilityStatus
}

// Represent the current state of a capability
type CapabilityStatus int

const (
	StatusUnloaded CapabilityStatus = iota
	StatusLoading
	StatusLoaded
	StatusError
)

// Create a new capability registry
func NewRegistry() *Registry {
	return &Registry{
		capabilities: make(map[string]CapabilityInfo),
		dependencies: make(map[string][]string),
	}
}

// Add a capability to the registry
func (r *Registry) Register(info CapabilityInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for circular dependencies
	if r.hasCircularDependency(info.Name, info.Dependencies) {
		return ErrCircularDependency
	}

	r.capabilities[info.Name] = info
	r.dependencies[info.Name] = info.Dependencies

	return nil
}

// Retrieve capability information
func (r *Registry) Get(name string) (CapabilityInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.capabilities[name]
	return info, exists
}

// Return all registered capabilities
func (r *Registry) List() []CapabilityInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]CapabilityInfo, 0, len(r.capabilities))
	for _, info := range r.capabilities {
		result = append(result, info)
	}

	return result
}

// Return capabilities in dependency order
func (r *Registry) ResolveDependencies() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Topological sort implementation
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	result := make([]string, 0)

	var visit func(string) error
	visit = func(name string) error {
		if temp[name] {
			return ErrCircularDependency
		}
		if visited[name] {
			return nil
		}

		temp[name] = true

		for _, dep := range r.dependencies[name] {
			if err := visit(dep); err != nil {
				return err
			}
		}

		temp[name] = false
		visited[name] = true
		result = append(result, name)

		return nil
	}

	for name := range r.capabilities {
		if !visited[name] {
			if err := visit(name); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// Check for circular dependencies
func (r *Registry) hasCircularDependency(name string, deps []string) bool {
	visited := make(map[string]bool)

	var check func(string) bool
	check = func(current string) bool {
		if current == name {
			return true
		}
		if visited[current] {
			return false
		}

		visited[current] = true

		for _, dep := range r.dependencies[current] {
			if check(dep) {
				return true
			}
		}

		return false
	}

	for _, dep := range deps {
		if check(dep) {
			return true
		}
	}

	return false
}

// Define custom errors
var (
	ErrCircularDependency = &RegistryError{"circular dependency detected"}
	ErrCapabilityNotFound = &RegistryError{"capability not found"}
)

type RegistryError struct {
	message string
}

func (e *RegistryError) Error() string {
	return e.message
}
