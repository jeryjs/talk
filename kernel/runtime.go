package kernel

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type EventType string

const (
	EventCapabilityLoad   EventType = "capability.load"
	EventCapabilityUnload EventType = "capability.unload"
	EventMessage          EventType = "message"
	EventConfig           EventType = "config"
	EventBehavioral       EventType = "behavioral"
	EventSystem           EventType = "system"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type EventHandler func(*Event) error

type Runtime struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	eventChan chan *Event
	handlers  map[EventType][]EventHandler
	mutex     sync.RWMutex
	context   *ContextManager
}

func NewRuntime() *Runtime {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runtime{
		ctx:       ctx,
		cancel:    cancel,
		eventChan: make(chan *Event, 1000),
		handlers:  make(map[EventType][]EventHandler),
		context:   NewContextManager(10000),
	}
}

func (r *Runtime) Start() error {
	r.wg.Add(1)
	go r.eventLoop()
	return nil
}

func (r *Runtime) Stop() {
	r.cancel()
	close(r.eventChan)
	r.wg.Wait()
	r.context.Close()
}

func (r *Runtime) Emit(event *Event) {
	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	select {
	case r.eventChan <- event:
	case <-r.ctx.Done():
	}
}

func (r *Runtime) Subscribe(eventType EventType, handler EventHandler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.handlers[eventType] = append(r.handlers[eventType], handler)
}

func (r *Runtime) Context() *ContextManager {
	return r.context
}

func (r *Runtime) eventLoop() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			return
		case event := <-r.eventChan:
			if event == nil {
				return
			}
			r.handleEvent(event)
		}
	}
}

func (r *Runtime) handleEvent(event *Event) {
	r.mutex.RLock()
	handlers := r.handlers[event.Type]
	r.mutex.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic but don't crash
				}
			}()
			h(event)
		}(handler)
	}
}
