package capabilities

import (
	"fmt"
	"sync"
	"time"
)

type CapabilityState string

const (
	StateUnloaded CapabilityState = "unloaded"
	StateLoading  CapabilityState = "loading"
	StateActive   CapabilityState = "active"
	StateInactive CapabilityState = "inactive"
	StateError    CapabilityState = "error"
)

type LifecycleEvent struct {
	Capability string
	OldState   CapabilityState
	NewState   CapabilityState
	Timestamp  time.Time
	Error      error
}

type LifecycleListener func(LifecycleEvent)

type LifecycleManager struct {
	states    map[string]CapabilityState
	listeners []LifecycleListener
	mutex     sync.RWMutex
}

func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		states:    make(map[string]CapabilityState),
		listeners: make([]LifecycleListener, 0),
	}
}

func (lm *LifecycleManager) AddListener(listener LifecycleListener) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	lm.listeners = append(lm.listeners, listener)
}

func (lm *LifecycleManager) SetState(capability string, newState CapabilityState, err error) {
	lm.mutex.Lock()
	oldState := lm.states[capability]
	lm.states[capability] = newState
	listeners := make([]LifecycleListener, len(lm.listeners))
	copy(listeners, lm.listeners)
	lm.mutex.Unlock()

	event := LifecycleEvent{
		Capability: capability,
		OldState:   oldState,
		NewState:   newState,
		Timestamp:  time.Now(),
		Error:      err,
	}

	// Notify listeners asynchronously
	for _, listener := range listeners {
		go listener(event)
	}
}

func (lm *LifecycleManager) GetState(capability string) CapabilityState {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	return lm.states[capability]
}

func (lm *LifecycleManager) GetAllStates() map[string]CapabilityState {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	result := make(map[string]CapabilityState)
	for k, v := range lm.states {
		result[k] = v
	}
	return result
}

func (lm *LifecycleManager) StartCapability(capability string, startFunc func() error) {
	lm.SetState(capability, StateLoading, nil)

	go func() {
		if err := startFunc(); err != nil {
			lm.SetState(capability, StateError, err)
		} else {
			lm.SetState(capability, StateActive, nil)
		}
	}()
}

func (lm *LifecycleManager) StopCapability(capability string, stopFunc func() error) {
	currentState := lm.GetState(capability)
	if currentState != StateActive {
		return
	}

	lm.SetState(capability, StateInactive, nil)

	go func() {
		if err := stopFunc(); err != nil {
			lm.SetState(capability, StateError, err)
		} else {
			lm.SetState(capability, StateUnloaded, nil)
		}
	}()
}

func (lm *LifecycleManager) RestartCapability(capability string, stopFunc, startFunc func() error) {
	lm.StopCapability(capability, stopFunc)

	// Wait a bit for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	lm.StartCapability(capability, startFunc)
}

func (lm *LifecycleManager) GetHealthySummary() string {
	states := lm.GetAllStates()

	active := 0
	inactive := 0
	errors := 0

	for _, state := range states {
		switch state {
		case StateActive:
			active++
		case StateInactive, StateUnloaded:
			inactive++
		case StateError:
			errors++
		}
	}

	return fmt.Sprintf("Active: %d, Inactive: %d, Errors: %d", active, inactive, errors)
}
