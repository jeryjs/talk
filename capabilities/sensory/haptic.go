package sensory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type HapticProcessor struct {
	isActive bool
	devices  map[string]HapticDevice
	patterns map[string]HapticPattern
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

type HapticDevice struct {
	ID           string
	Name         string
	Type         string
	Capabilities []string
	IsConnected  bool
}

type HapticPattern struct {
	Name      string
	Duration  time.Duration
	Intensity float64
	Frequency float64
	Pulses    []HapticPulse
}

type HapticPulse struct {
	Delay     time.Duration
	Duration  time.Duration
	Intensity float64
}

type HapticFeedback struct {
	DeviceID  string
	Pattern   string
	Intensity float64
	Duration  time.Duration
}

func NewHapticProcessor() *HapticProcessor {
	return &HapticProcessor{
		devices:  make(map[string]HapticDevice),
		patterns: make(map[string]HapticPattern),
	}
}

func (hp *HapticProcessor) Start(ctx context.Context) error {
	hp.mutex.Lock()
	defer hp.mutex.Unlock()

	if hp.isActive {
		return fmt.Errorf("haptic processor already active")
	}

	hp.ctx, hp.cancel = context.WithCancel(ctx)
	hp.isActive = true

	// Initialize default patterns
	hp.loadDefaultPatterns()

	// Discover haptic devices
	go hp.deviceDiscovery()

	return nil
}

func (hp *HapticProcessor) Stop() {
	hp.mutex.Lock()
	defer hp.mutex.Unlock()

	if !hp.isActive {
		return
	}

	hp.isActive = false
	hp.cancel()
}

func (hp *HapticProcessor) loadDefaultPatterns() {
	patterns := map[string]HapticPattern{
		"notification": {
			Name:      "notification",
			Duration:  500 * time.Millisecond,
			Intensity: 0.7,
			Frequency: 250,
			Pulses: []HapticPulse{
				{Delay: 0, Duration: 100 * time.Millisecond, Intensity: 0.7},
				{Delay: 200 * time.Millisecond, Duration: 100 * time.Millisecond, Intensity: 0.5},
			},
		},
		"success": {
			Name:      "success",
			Duration:  300 * time.Millisecond,
			Intensity: 0.8,
			Frequency: 300,
			Pulses: []HapticPulse{
				{Delay: 0, Duration: 150 * time.Millisecond, Intensity: 0.8},
			},
		},
		"error": {
			Name:      "error",
			Duration:  1000 * time.Millisecond,
			Intensity: 1.0,
			Frequency: 100,
			Pulses: []HapticPulse{
				{Delay: 0, Duration: 200 * time.Millisecond, Intensity: 1.0},
				{Delay: 300 * time.Millisecond, Duration: 200 * time.Millisecond, Intensity: 0.8},
				{Delay: 600 * time.Millisecond, Duration: 200 * time.Millisecond, Intensity: 0.6},
			},
		},
		"heartbeat": {
			Name:      "heartbeat",
			Duration:  2000 * time.Millisecond,
			Intensity: 0.6,
			Frequency: 60,
			Pulses: []HapticPulse{
				{Delay: 0, Duration: 100 * time.Millisecond, Intensity: 0.8},
				{Delay: 200 * time.Millisecond, Duration: 80 * time.Millisecond, Intensity: 0.4},
				{Delay: 1000 * time.Millisecond, Duration: 100 * time.Millisecond, Intensity: 0.8},
				{Delay: 1200 * time.Millisecond, Duration: 80 * time.Millisecond, Intensity: 0.4},
			},
		},
	}

	for name, pattern := range patterns {
		hp.patterns[name] = pattern
	}
}

func (hp *HapticProcessor) deviceDiscovery() {
	// This would scan for haptic devices
	// For now, simulate a gamepad controller
	hp.mutex.Lock()
	hp.devices["controller"] = HapticDevice{
		ID:           "controller",
		Name:         "Game Controller",
		Type:         "gamepad",
		Capabilities: []string{"vibration", "rumble"},
		IsConnected:  false, // Would check actual device
	}
	hp.mutex.Unlock()
}

func (hp *HapticProcessor) TriggerFeedback(feedback HapticFeedback) error {
	hp.mutex.RLock()
	device, exists := hp.devices[feedback.DeviceID]
	hp.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("device not found: %s", feedback.DeviceID)
	}

	if !device.IsConnected {
		// Simulate haptic feedback in terminal
		hp.simulateHaptic(feedback)
		return nil
	}

	// Would send actual haptic commands to device
	return hp.sendToDevice(device, feedback)
}

func (hp *HapticProcessor) simulateHaptic(feedback HapticFeedback) {
	intensity := int(feedback.Intensity * 10)
	vibration := ""
	for i := 0; i < intensity; i++ {
		vibration += "▓"
	}
	for i := intensity; i < 10; i++ {
		vibration += "░"
	}

	fmt.Printf("🎮 Haptic: [%s] %s (%.1fs)\n", vibration, feedback.Pattern, feedback.Duration.Seconds())
}

func (hp *HapticProcessor) sendToDevice(device HapticDevice, feedback HapticFeedback) error {
	// Platform-specific haptic device communication
	return fmt.Errorf("haptic device communication not implemented")
}

func (hp *HapticProcessor) GetPatterns() map[string]HapticPattern {
	hp.mutex.RLock()
	defer hp.mutex.RUnlock()

	patterns := make(map[string]HapticPattern)
	for k, v := range hp.patterns {
		patterns[k] = v
	}
	return patterns
}

func (hp *HapticProcessor) GetDevices() map[string]HapticDevice {
	hp.mutex.RLock()
	defer hp.mutex.RUnlock()

	devices := make(map[string]HapticDevice)
	for k, v := range hp.devices {
		devices[k] = v
	}
	return devices
}

func (hp *HapticProcessor) GetStatus() map[string]interface{} {
	hp.mutex.RLock()
	defer hp.mutex.RUnlock()

	return map[string]interface{}{
		"active":       hp.isActive,
		"deviceCount":  len(hp.devices),
		"patternCount": len(hp.patterns),
	}
}
