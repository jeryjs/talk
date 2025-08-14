package capabilities

import (
	"fmt"
	"sync"
)

type DependencyNode struct {
	Name         string
	Dependencies []string
	Dependents   []string
	Loaded       bool
}

type DependencyGraph struct {
	nodes map[string]*DependencyNode
	mutex sync.RWMutex
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*DependencyNode),
	}
}

func (dg *DependencyGraph) AddNode(name string, dependencies []string) {
	dg.mutex.Lock()
	defer dg.mutex.Unlock()

	node := &DependencyNode{
		Name:         name,
		Dependencies: dependencies,
		Dependents:   make([]string, 0),
		Loaded:       false,
	}

	dg.nodes[name] = node

	// Update dependents
	for _, dep := range dependencies {
		if depNode, exists := dg.nodes[dep]; exists {
			depNode.Dependents = append(depNode.Dependents, name)
		} else {
			// Create placeholder node
			dg.nodes[dep] = &DependencyNode{
				Name:         dep,
				Dependencies: make([]string, 0),
				Dependents:   []string{name},
				Loaded:       false,
			}
		}
	}
}

func (dg *DependencyGraph) RemoveNode(name string) error {
	dg.mutex.Lock()
	defer dg.mutex.Unlock()

	node, exists := dg.nodes[name]
	if !exists {
		return fmt.Errorf("node not found: %s", name)
	}

	// Check if any dependents are loaded
	for _, dependent := range node.Dependents {
		if depNode, exists := dg.nodes[dependent]; exists && depNode.Loaded {
			return fmt.Errorf("cannot remove %s: %s depends on it", name, dependent)
		}
	}

	// Remove from dependencies
	for _, dep := range node.Dependencies {
		if depNode, exists := dg.nodes[dep]; exists {
			for i, dependent := range depNode.Dependents {
				if dependent == name {
					depNode.Dependents = append(depNode.Dependents[:i], depNode.Dependents[i+1:]...)
					break
				}
			}
		}
	}

	delete(dg.nodes, name)
	return nil
}

func (dg *DependencyGraph) GetLoadOrder(target string) ([]string, error) {
	dg.mutex.RLock()
	defer dg.mutex.RUnlock()

	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	order := make([]string, 0)

	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("circular dependency detected involving %s", name)
		}

		if visited[name] {
			return nil
		}

		node, exists := dg.nodes[name]
		if !exists {
			return fmt.Errorf("dependency not found: %s", name)
		}

		visiting[name] = true

		for _, dep := range node.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		order = append(order, name)

		return nil
	}

	if err := visit(target); err != nil {
		return nil, err
	}

	return order, nil
}

func (dg *DependencyGraph) GetUnloadOrder(target string) ([]string, error) {
	dg.mutex.RLock()
	defer dg.mutex.RUnlock()

	visited := make(map[string]bool)
	order := make([]string, 0)

	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}

		node, exists := dg.nodes[name]
		if !exists {
			return
		}

		visited[name] = true

		// Visit dependents first
		for _, dependent := range node.Dependents {
			visit(dependent)
		}

		order = append(order, name)
	}

	visit(target)
	return order, nil
}

func (dg *DependencyGraph) MarkLoaded(name string) {
	dg.mutex.Lock()
	defer dg.mutex.Unlock()

	if node, exists := dg.nodes[name]; exists {
		node.Loaded = true
	}
}

func (dg *DependencyGraph) MarkUnloaded(name string) {
	dg.mutex.Lock()
	defer dg.mutex.Unlock()

	if node, exists := dg.nodes[name]; exists {
		node.Loaded = false
	}
}

func (dg *DependencyGraph) IsLoaded(name string) bool {
	dg.mutex.RLock()
	defer dg.mutex.RUnlock()

	if node, exists := dg.nodes[name]; exists {
		return node.Loaded
	}
	return false
}

func (dg *DependencyGraph) CanLoad(name string) (bool, []string) {
	dg.mutex.RLock()
	defer dg.mutex.RUnlock()

	node, exists := dg.nodes[name]
	if !exists {
		return false, []string{"node not found"}
	}

	missing := make([]string, 0)
	for _, dep := range node.Dependencies {
		if depNode, exists := dg.nodes[dep]; !exists || !depNode.Loaded {
			missing = append(missing, dep)
		}
	}

	return len(missing) == 0, missing
}

func (dg *DependencyGraph) GetStats() map[string]interface{} {
	dg.mutex.RLock()
	defer dg.mutex.RUnlock()

	total := len(dg.nodes)
	loaded := 0
	orphans := 0

	for _, node := range dg.nodes {
		if node.Loaded {
			loaded++
		}
		if len(node.Dependencies) == 0 && len(node.Dependents) == 0 {
			orphans++
		}
	}

	return map[string]interface{}{
		"total_nodes":  total,
		"loaded_nodes": loaded,
		"orphan_nodes": orphans,
		"load_ratio":   float64(loaded) / float64(total),
	}
}
