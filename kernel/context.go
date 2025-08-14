package kernel

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// ContextType defines different types of context data
type ContextType string

const (
	ContextMemory     ContextType = "memory"
	ContextSession    ContextType = "session"
	ContextBehavioral ContextType = "behavioral"
	ContextCapability ContextType = "capability"
)

// ContextEntry represents a single piece of contextual information
type ContextEntry struct {
	ID        string                 `json:"id"`
	Type      ContextType            `json:"type"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	Relevance float64                `json:"relevance"` // 0.0 - 1.0
	TTL       *time.Duration         `json:"ttl,omitempty"`
}

// ContextManager handles all context and memory operations
type ContextManager struct {
	entries     map[string]*ContextEntry
	byType      map[ContextType][]*ContextEntry
	mutex       sync.RWMutex
	maxEntries  int
	cleanupTick *time.Ticker
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewContextManager creates a new context manager
func NewContextManager(maxEntries int) *ContextManager {
	ctx, cancel := context.WithCancel(context.Background())

	cm := &ContextManager{
		entries:    make(map[string]*ContextEntry),
		byType:     make(map[ContextType][]*ContextEntry),
		maxEntries: maxEntries,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start cleanup routine
	cm.cleanupTick = time.NewTicker(5 * time.Minute)
	go cm.cleanupRoutine()

	return cm
}

// Add inserts or updates a context entry
func (cm *ContextManager) Add(entry *ContextEntry) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Remove existing entry if updating
	if existing, exists := cm.entries[entry.ID]; exists {
		cm.removeFromType(existing)
	}

	// Set timestamp
	entry.Timestamp = time.Now()

	// Add to maps
	cm.entries[entry.ID] = entry
	cm.byType[entry.Type] = append(cm.byType[entry.Type], entry)

	// Enforce max entries per type
	cm.enforceMaxEntries(entry.Type)
}

// Get retrieves a context entry by ID
func (cm *ContextManager) Get(id string) (*ContextEntry, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	entry, exists := cm.entries[id]
	return entry, exists
}

// GetByType retrieves all entries of a specific type
func (cm *ContextManager) GetByType(contextType ContextType) []*ContextEntry {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	entries := cm.byType[contextType]
	result := make([]*ContextEntry, len(entries))
	copy(result, entries)
	return result
}

// GetRelevant retrieves entries above a relevance threshold
func (cm *ContextManager) GetRelevant(threshold float64, limit int) []*ContextEntry {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var relevant []*ContextEntry
	for _, entry := range cm.entries {
		if entry.Relevance >= threshold {
			relevant = append(relevant, entry)
		}
	}

	// Sort by relevance (highest first)
	for i := 0; i < len(relevant)-1; i++ {
		for j := i + 1; j < len(relevant); j++ {
			if relevant[i].Relevance < relevant[j].Relevance {
				relevant[i], relevant[j] = relevant[j], relevant[i]
			}
		}
	}

	if limit > 0 && len(relevant) > limit {
		relevant = relevant[:limit]
	}

	return relevant
}

// UpdateRelevance adjusts the relevance score of an entry
func (cm *ContextManager) UpdateRelevance(id string, relevance float64) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if entry, exists := cm.entries[id]; exists {
		entry.Relevance = relevance
	}
}

// Remove deletes a context entry
func (cm *ContextManager) Remove(id string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if entry, exists := cm.entries[id]; exists {
		delete(cm.entries, id)
		cm.removeFromType(entry)
	}
}

// Clear removes all entries of a specific type
func (cm *ContextManager) Clear(contextType ContextType) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Remove from main map
	for _, entry := range cm.byType[contextType] {
		delete(cm.entries, entry.ID)
	}

	// Clear type map
	cm.byType[contextType] = nil
}

// Export returns all context data as JSON
func (cm *ContextManager) Export() ([]byte, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var entries []*ContextEntry
	for _, entry := range cm.entries {
		entries = append(entries, entry)
	}

	return json.Marshal(entries)
}

// Import loads context data from JSON
func (cm *ContextManager) Import(data []byte) error {
	var entries []*ContextEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Clear existing data
	cm.entries = make(map[string]*ContextEntry)
	cm.byType = make(map[ContextType][]*ContextEntry)

	// Add imported entries
	for _, entry := range entries {
		cm.entries[entry.ID] = entry
		cm.byType[entry.Type] = append(cm.byType[entry.Type], entry)
	}

	return nil
}

// Close shuts down the context manager
func (cm *ContextManager) Close() {
	cm.cancel()
	if cm.cleanupTick != nil {
		cm.cleanupTick.Stop()
	}
}

// cleanupRoutine removes expired entries
func (cm *ContextManager) cleanupRoutine() {
	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-cm.cleanupTick.C:
			cm.cleanup()
		}
	}
}

// cleanup removes expired and low-relevance entries
func (cm *ContextManager) cleanup() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for id, entry := range cm.entries {
		// Check TTL expiration
		if entry.TTL != nil && now.Sub(entry.Timestamp) > *entry.TTL {
			toRemove = append(toRemove, id)
			continue
		}

		// Remove very low relevance entries older than 1 hour
		if entry.Relevance < 0.1 && now.Sub(entry.Timestamp) > time.Hour {
			toRemove = append(toRemove, id)
		}
	}

	// Remove expired entries
	for _, id := range toRemove {
		if entry, exists := cm.entries[id]; exists {
			delete(cm.entries, id)
			cm.removeFromType(entry)
		}
	}
}

// removeFromType removes entry from type-specific slice
func (cm *ContextManager) removeFromType(entry *ContextEntry) {
	entries := cm.byType[entry.Type]
	for i, e := range entries {
		if e.ID == entry.ID {
			cm.byType[entry.Type] = append(entries[:i], entries[i+1:]...)
			break
		}
	}
}

// enforceMaxEntries removes oldest entries if limit exceeded
func (cm *ContextManager) enforceMaxEntries(contextType ContextType) {
	entries := cm.byType[contextType]
	if len(entries) <= cm.maxEntries {
		return
	}

	// Sort by timestamp (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].Timestamp.After(entries[j].Timestamp) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest entries
	excess := len(entries) - cm.maxEntries
	for i := 0; i < excess; i++ {
		delete(cm.entries, entries[i].ID)
	}

	cm.byType[contextType] = entries[excess:]
}
