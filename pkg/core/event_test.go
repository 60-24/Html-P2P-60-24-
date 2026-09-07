package core

import (
	"testing"
	"time"
)

// Test 1: Event Creation
func TestEventCreation(t *testing.T) {
	e := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), PriorityNormal)
	
	if e.Sender != "Alice" {
		t.Errorf("Expected sender Alice, got %s", e.Sender)
	}
	if e.EventType != EventTypeProofOfMeeting {
		t.Errorf("Expected type ProofOfMeeting, got %s", e.EventType)
	}
	if e.Priority != PriorityNormal {
		t.Errorf("Expected priority Normal, got %d", e.Priority)
	}
}

// Test 2: Event Hash Computation
func TestEventHashComputation(t *testing.T) {
	e := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), PriorityNormal)
	hash1 := e.ComputeHash()
	hash2 := e.ComputeHash()
	
	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic")
	}
	if hash1 == "" {
		t.Errorf("Hash should not be empty")
	}
}

// Test 3: Event Verification
func TestEventVerification(t *testing.T) {
	e := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), PriorityNormal)
	
	if !e.Verify() {
		t.Errorf("Event should verify correctly")
	}
}

// Test 4: Event ID Uniqueness
func TestEventIDUniqueness(t *testing.T) {
	e1 := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test1"), PriorityNormal)
	time.Sleep(1 * time.Nanosecond)
	e2 := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test2"), PriorityNormal)
	
	if e1.ID == e2.ID {
		t.Errorf("Events should have unique IDs")
	}
}

// Test 5: Event Priority Levels
func TestEventPriorityLevels(t *testing.T) {
	priorityTests := []struct {
		priority int
		name     string
	}{
		{PriorityLow, "Low"},
		{PriorityNormal, "Normal"},
		{PriorityHigh, "High"},
		{PriorityCritical, "Critical"},
	}
	
	for _, tt := range priorityTests {
		e := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), tt.priority)
		if e.Priority != tt.priority {
			t.Errorf("Expected priority %d (%s), got %d", tt.priority, tt.name, e.Priority)
		}
	}
}

// Test 6: Event JSON Marshaling
func TestEventJSONMarshaling(t *testing.T) {
	e := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), PriorityNormal)
	jsonData, err := e.ToJSON()
	
	if err != nil {
		t.Errorf("Failed to marshal event: %v", err)
	}
	if len(jsonData) == 0 {
		t.Errorf("JSON data should not be empty")
	}
}

// Test 7: Event JSON Unmarshaling
func TestEventJSONUnmarshaling(t *testing.T) {
	e1 := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), PriorityNormal)
	jsonData, _ := e1.ToJSON()
	
	e2, err := FromJSON(jsonData)
	if err != nil {
		t.Errorf("Failed to unmarshal event: %v", err)
	}
	if e2.ID != e1.ID {
		t.Errorf("Event ID should match after marshaling")
	}
}

// Test 8: Event Types Constants
func TestEventTypes(t *testing.T) {
	types := []string{
		EventTypeProofOfMeeting,
		EventTypeCommitment,
		EventTypeResolution,
		EventTypeReputationUpdate,
		EventTypeNodeJoined,
		EventTypeNodeLeft,
	}
	
	for _, eventType := range types {
		e := NewEvent("Alice", eventType, []byte("test"), PriorityNormal)
		if e.EventType != eventType {
			t.Errorf("Expected type %s, got %s", eventType, e.EventType)
		}
	}
}

// Test 9: Event Payload Handling
func TestEventPayloadHandling(t *testing.T) {
	payload := []byte("important data")
	e := NewEvent("Alice", EventTypeProofOfMeeting, payload, PriorityNormal)
	
	if string(e.Payload) != string(payload) {
		t.Errorf("Payload mismatch")
	}
}

// Test 10: Event TTL Default
func TestEventTTLDefault(t *testing.T) {
	e := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), PriorityNormal)
	
	if e.TTL != 0 {
		t.Errorf("Default TTL should be 0 (forever), got %d", e.TTL)
	}
}

// Test 11: Event Timestamp Generation
func TestEventTimestampGeneration(t *testing.T) {
	before := time.Now().UnixNano()
	e := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), PriorityNormal)
	after := time.Now().UnixNano()
	
	if e.Timestamp < before || e.Timestamp > after {
		t.Errorf("Timestamp should be current: %d (between %d and %d)", e.Timestamp, before, after)
	}
}

// Test 12: Event String Representation
func TestEventString(t *testing.T) {
	e := NewEvent("Alice", EventTypeProofOfMeeting, []byte("test"), PriorityNormal)
	str := e.String()
	
	if len(str) == 0 {
		t.Errorf("String representation should not be empty")
	}
	if !contains(str, "Alice") {
		t.Errorf("String should contain sender name")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
