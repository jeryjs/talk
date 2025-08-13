package kernel

import (
	"sync"
	"time"
)

// Provide advanced context and state management
type Memory struct {
	contexts map[string]*Context
	sessions map[string]*Session
	cache    *Cache
	mu       sync.RWMutex
}

// Represent rich contextual information
type Context struct {
	ID        string
	Type      string
	Data      map[string]interface{}
	Metadata  map[string]string
	Timestamp time.Time
	TTL       time.Duration
}

// Manage conversation and interaction state
type Session struct {
	ID      string
	UserID  string
	Started time.Time
	Updated time.Time
	State   map[string]interface{}
	History []Interaction
}

// Represent a single interaction in a session
type Interaction struct {
	ID        string
	Type      string
	Input     interface{}
	Output    interface{}
	Timestamp time.Time
	Metadata  map[string]string
}

// Provide high-performance caching for frequently accessed data
type Cache struct {
	data map[string]*CacheEntry
	mu   sync.RWMutex
}

// Represent a cached item with TTL
type CacheEntry struct {
	Value   interface{}
	Expires time.Time
}

// Create a new memory management system
func NewMemory() *Memory {
	return &Memory{
		contexts: make(map[string]*Context),
		sessions: make(map[string]*Session),
		cache:    NewCache(),
	}
}

// Create a new cache instance
func NewCache() *Cache {
	cache := &Cache{
		data: make(map[string]*CacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanup()
	return cache
}

// Store contextual information
func (m *Memory) SetContext(ctx *Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.contexts[ctx.ID] = ctx
}

// Retrieve contextual information
func (m *Memory) GetContext(id string) (*Context, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, exists := m.contexts[id]
	if exists && time.Since(ctx.Timestamp) > ctx.TTL {
		delete(m.contexts, id)
		return nil, false
	}

	return ctx, exists
}

// Start a new interaction session
func (m *Memory) CreateSession(userID string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &Session{
		ID:      generateID(),
		UserID:  userID,
		Started: time.Now(),
		Updated: time.Now(),
		State:   make(map[string]interface{}),
		History: make([]Interaction, 0),
	}

	m.sessions[session.ID] = session
	return session
}

// Add an interaction to a session
func (m *Memory) AddInteraction(sessionID string, interaction Interaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.History = append(session.History, interaction)
		session.Updated = time.Now()
	}
}

// Store a value in cache with TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &CacheEntry{
		Value:   value,
		Expires: time.Now().Add(ttl),
	}
}

// Retrieve a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.Expires) {
		return nil, false
	}

	return entry.Value, true
}

// Remove expired cache entries
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.Expires) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// Create a unique identifier
func generateID() string {
	// Simple ID generation - replace with UUID in production
	return time.Now().Format("20060102150405") + "-" + "random"
}
