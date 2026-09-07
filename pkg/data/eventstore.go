package data

import (
	"encoding/json"
	"sync"

	"github.com/60-24/p2p-60-24/pkg/core"
)

// EventStore przechowuje immutable log wszystkich zdarzeń
type EventStore struct {
	mu     sync.RWMutex
	events []*core.Event
	index  map[string]int // ID -> position
}

// NewEventStore tworzy nowy EventStore
func NewEventStore() *EventStore {
	return &EventStore{
		events: make([]*core.Event, 0),
		index:  make(map[string]int),
	}
}

// Append dodaje event do logu (IMMUTABLE - nie można usunąć)
func (es *EventStore) Append(event *core.Event) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.index[event.ID] = len(es.events)
	es.events = append(es.events, event)
	return nil
}

// GetByID pobiera event po ID
func (es *EventStore) GetByID(id string) (*core.Event, bool) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	idx, exists := es.index[id]
	if !exists {
		return nil, false
	}
	return es.events[idx], true
}

// GetAll zwraca wszystkie eventy
func (es *EventStore) GetAll() []*core.Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	all := make([]*core.Event, len(es.events))
	copy(all, es.events)
	return all
}

// GetByType zwraca eventy określonego typu
func (es *EventStore) GetByType(eventType string) []*core.Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var result []*core.Event
	for _, event := range es.events {
		if event.EventType == eventType {
			result = append(result, event)
		}
	}
	return result
}

// GetBySender zwraca eventy od określonego nadawcy
func (es *EventStore) GetBySender(sender string) []*core.Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var result []*core.Event
	for _, event := range es.events {
		if event.Sender == sender {
			result = append(result, event)
		}
	}
	return result
}

// GetRange zwraca eventy w danym zakresie
func (es *EventStore) GetRange(start, end int) []*core.Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	if start < 0 || end > len(es.events) || start > end {
		return []*core.Event{}
	}

	result := make([]*core.Event, end-start)
	copy(result, es.events[start:end])
	return result
}

// Size zwraca ilość zdarzeń
func (es *EventStore) Size() int {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return len(es.events)
}

// Snapshot tworzy snapshot stanu (do checkpointingu)
func (es *EventStore) Snapshot() []byte {
	es.mu.RLock()
	defer es.mu.RUnlock()

	data, _ := json.Marshal(es.events)
	return data
}

// LoadSnapshot wczytuje snapshot
func (es *EventStore) LoadSnapshot(data []byte) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	var events []*core.Event
	if err := json.Unmarshal(data, &events); err != nil {
		return err
	}

	es.events = events
	es.index = make(map[string]int)
	for i, event := range es.events {
		es.index[event.ID] = i
	}

	return nil
}
