package data

import (
	"encoding/json"
	"sync"

	"github.com/60-24/p2p-60-24/pkg/core"
)

// StateEngine aplikuje eventy do stanu
type StateEngine struct {
	mu    sync.RWMutex
	state map[string]interface{} // Current state (CRDT-compatible)
	log   *EventStore             // Event log (source of truth)
}

// NewStateEngine tworzy nowy StateEngine
func NewStateEngine(log *EventStore) *StateEngine {
	return &StateEngine{
		state: make(map[string]interface{}),
		log:   log,
	}
}

// ApplyEvent aplikuje event do stanu
func (se *StateEngine) ApplyEvent(event *core.Event) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	// Deserializuj payload
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	// Aplikuj do stanu (simplified CRDT)
	for k, v := range payload {
		se.state[k] = v
	}

	return nil
}

// ApplyEventLog aplikuje cały event log (replay)
func (se *StateEngine) ApplyEventLog(events []*core.Event) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	for _, event := range events {
		var payload map[string]interface{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		for k, v := range payload {
			se.state[k] = v
		}
	}

	return nil
}

// GetState zwraca aktualny stan
func (se *StateEngine) GetState() map[string]interface{} {
	se.mu.RLock()
	defer se.mu.RUnlock()

	state := make(map[string]interface{})
	for k, v := range se.state {
		state[k] = v
	}
	return state
}

// GetValue pobiera wartość ze stanu
func (se *StateEngine) GetValue(key string) (interface{}, bool) {
	se.mu.RLock()
	defer se.mu.RUnlock()

	val, exists := se.state[key]
	return val, exists
}

// Reset resetuje stan (dla testów)
func (se *StateEngine) Reset() {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.state = make(map[string]interface{})
}

// Snapshot tworzy snapshot stanu
func (se *StateEngine) Snapshot() []byte {
	se.mu.RLock()
	defer se.mu.RUnlock()

	data, _ := json.Marshal(se.state)
	return data
}
