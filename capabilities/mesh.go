package capabilities

import (
	"fmt"
	"sync"
)

type MessageType string

const (
	MessageRequest  MessageType = "request"
	MessageResponse MessageType = "response"
	MessageEvent    MessageType = "event"
	MessageError    MessageType = "error"
)

type Message struct {
	ID       string                 `json:"id"`
	Type     MessageType            `json:"type"`
	From     string                 `json:"from"`
	To       string                 `json:"to"`
	Method   string                 `json:"method,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Response interface{}            `json:"response,omitempty"`
}

type MessageHandler func(*Message) (*Message, error)

type Mesh struct {
	capabilities  map[string]Capability
	handlers      map[string]map[string]MessageHandler // capability -> method -> handler
	subscriptions map[string][]string                  // event -> subscribers
	mutex         sync.RWMutex
}

func NewMesh() *Mesh {
	return &Mesh{
		capabilities:  make(map[string]Capability),
		handlers:      make(map[string]map[string]MessageHandler),
		subscriptions: make(map[string][]string),
	}
}

func (m *Mesh) Register(name string, capability Capability) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.capabilities[name] = capability
	m.handlers[name] = make(map[string]MessageHandler)
}

func (m *Mesh) Unregister(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.capabilities, name)
	delete(m.handlers, name)

	// Remove from subscriptions
	for event, subscribers := range m.subscriptions {
		for i, sub := range subscribers {
			if sub == name {
				m.subscriptions[event] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}
}

func (m *Mesh) RegisterHandler(capability, method string, handler MessageHandler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.handlers[capability] == nil {
		m.handlers[capability] = make(map[string]MessageHandler)
	}
	m.handlers[capability][method] = handler
}

func (m *Mesh) Subscribe(capability, event string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.subscriptions[event] = append(m.subscriptions[event], capability)
}

func (m *Mesh) Send(message *Message) (*Message, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check if target capability exists
	if _, exists := m.capabilities[message.To]; !exists {
		return nil, fmt.Errorf("capability not found: %s", message.To)
	}

	// Find handler for the method
	if handlers, exists := m.handlers[message.To]; exists {
		if handler, exists := handlers[message.Method]; exists {
			return handler(message)
		}
	}

	return nil, fmt.Errorf("method not found: %s.%s", message.To, message.Method)
}

func (m *Mesh) Broadcast(event string, data map[string]interface{}) {
	m.mutex.RLock()
	subscribers := m.subscriptions[event]
	m.mutex.RUnlock()

	message := &Message{
		Type: MessageEvent,
		From: "mesh",
		Data: data,
	}

	for _, subscriber := range subscribers {
		message.To = subscriber
		go m.Send(message) // Send asynchronously
	}
}

func (m *Mesh) ListCapabilities() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.capabilities))
	for name := range m.capabilities {
		names = append(names, name)
	}
	return names
}

func (m *Mesh) GetCapabilityInfo(name string) map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	info := make(map[string]interface{})

	if cap, exists := m.capabilities[name]; exists {
		info["name"] = cap.GetName()
		info["version"] = cap.GetVersion()

		// List available methods
		if methods, exists := m.handlers[name]; exists {
			methodNames := make([]string, 0, len(methods))
			for method := range methods {
				methodNames = append(methodNames, method)
			}
			info["methods"] = methodNames
		}
	}

	return info
}
