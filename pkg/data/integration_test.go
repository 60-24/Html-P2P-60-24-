package data

import (
	"encoding/json"
	"testing"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/sec"
)

// Test EventStore Append & Retrieve
func TestEventStoreAppendAndRetrieve(t *testing.T) {
	store := NewEventStore()
	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	
	store.Append(event)
	retrieved, exists := store.GetByID(event.ID)
	
	if !exists {
		t.Errorf("Event should be retrievable")
	}
	if retrieved.ID != event.ID {
		t.Errorf("Retrieved event ID mismatch")
	}
}

// Test EventStore Size
func TestEventStoreSize(t *testing.T) {
	store := NewEventStore()
	
	if store.Size() != 0 {
		t.Errorf("Initial size should be 0")
	}
	
	store.Append(core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal))
	store.Append(core.NewEvent("Bob", core.EventTypeCommitment, []byte("test"), core.PriorityHigh))
	
	if store.Size() != 2 {
		t.Errorf("Expected size 2, got %d", store.Size())
	}
}

// Test EventStore GetByType
func TestEventStoreGetByType(t *testing.T) {
	store := NewEventStore()
	
	store.Append(core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal))
	store.Append(core.NewEvent("Bob", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal))
	store.Append(core.NewEvent("Charlie", core.EventTypeCommitment, []byte("test"), core.PriorityHigh))
	
	meetings := store.GetByType(core.EventTypeProofOfMeeting)
	if len(meetings) != 2 {
		t.Errorf("Expected 2 meetings, got %d", len(meetings))
	}
}

// Test StateEngine ApplyEvent
func TestStateEngineApplyEvent(t *testing.T) {
	store := NewEventStore()
	engine := NewStateEngine(store)
	
	payload := map[string]interface{}{
		"alice_trust": 0.87,
		"bob_trust":   0.92,
	}
	payloadBytes, _ := json.Marshal(payload)
	
	event := core.NewEvent("Alice", core.EventTypeReputationUpdate, payloadBytes, core.PriorityNormal)
	engine.ApplyEvent(event)
	
	state := engine.GetState()
	if state["alice_trust"] != 0.87 {
		t.Errorf("State application failed")
	}
}

// Test StateEngine Replay (Event Log Idempotence)
func TestStateEngineReplay(t *testing.T) {
	store := NewEventStore()
	engine := NewStateEngine(store)
	
	// Create 3 events
	payload1, _ := json.Marshal(map[string]interface{}{"counter": 1})
	payload2, _ := json.Marshal(map[string]interface{}{"counter": 2})
	payload3, _ := json.Marshal(map[string]interface{}{"counter": 3})
	
	e1 := core.NewEvent("Alice", core.EventTypeReputationUpdate, payload1, core.PriorityNormal)
	e2 := core.NewEvent("Alice", core.EventTypeReputationUpdate, payload2, core.PriorityNormal)
	e3 := core.NewEvent("Alice", core.EventTypeReputationUpdate, payload3, core.PriorityNormal)
	
	// Apply all
	events := []*core.Event{e1, e2, e3}
	engine.ApplyEventLog(events)
	
	state := engine.GetState()
	if state["counter"] != 3 {
		t.Errorf("Final counter should be 3, got %v", state["counter"])
	}
}

// Test KeyManager Sign & Verify
func TestKeyManagerSignAndVerify(t *testing.T) {
	km, _ := sec.NewKeyManager("Alice")
	
	data := []byte("important message")
	sig := km.Sign(data)
	
	if !km.Verify(data, sig) {
		t.Errorf("Signature verification failed")
	}
	
	// Tampering should fail
	tamperedData := []byte("tampered message")
	if km.Verify(tamperedData, sig) {
		t.Errorf("Tampered data should fail verification")
	}
}
