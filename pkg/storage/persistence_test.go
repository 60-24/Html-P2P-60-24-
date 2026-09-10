package storage

import (
	"testing"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/data"
)

// Test 1: Store Creation
func TestNewInMemoryStore(t *testing.T) {
	store, err := NewInMemoryStore("/tmp/test-p2p")
	if err != nil {
		t.Fatalf("NewInMemoryStore failed: %v", err)
	}
	defer store.Close()

	stats := store.GetStats()
	if stats["engine"] != "inmemory" {
		t.Errorf("Expected engine inmemory, got %v", stats["engine"])
	}
	if stats["event_log_size"] != 0 {
		t.Errorf("Expected empty event log initially")
	}
}

// Test 2: Store & Get Event (round trip)
func TestStoreAndGetEvent(t *testing.T) {
	store, _ := NewInMemoryStore("/tmp/test-p2p")
	defer store.Close()

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)

	err := store.StoreEvent(event)
	if err != nil {
		t.Errorf("StoreEvent failed: %v", err)
	}

	retrieved, err := store.GetEvent(event.ID)
	if err != nil {
		t.Errorf("GetEvent failed: %v", err)
	}
	if retrieved.ID != event.ID {
		t.Errorf("Retrieved event ID mismatch")
	}
}

// Test 3: Get Non-Existent Event
func TestGetNonExistentEvent(t *testing.T) {
	store, _ := NewInMemoryStore("/tmp/test-p2p")
	defer store.Close()

	_, err := store.GetEvent("does-not-exist")
	if err == nil {
		t.Errorf("Expected error for non-existent event")
	}
}

// Test 4: Store Full Event Log
func TestStoreEventLog(t *testing.T) {
	store, _ := NewInMemoryStore("/tmp/test-p2p")
	defer store.Close()

	eventStore := data.NewEventStore()
	for i := 0; i < 5; i++ {
		e := core.NewEvent("Alice", core.EventTypeCommitment, []byte("test"), core.PriorityNormal)
		eventStore.Append(e)
	}

	err := store.StoreEventLog(eventStore)
	if err != nil {
		t.Errorf("StoreEventLog failed: %v", err)
	}

	stats := store.GetStats()
	if stats["event_log_size"] != 5 {
		t.Errorf("Expected 5 events persisted, got %v", stats["event_log_size"])
	}
}

// Test 5: Snapshot Create & Restore (round trip, deep copy)
func TestSnapshotRoundTrip(t *testing.T) {
	store, _ := NewInMemoryStore("/tmp/test-p2p")
	defer store.Close()

	state := map[string]interface{}{
		"trust_alice_bob": 0.87,
		"event_count":     42,
	}

	snapshotID, err := store.CreateSnapshot("Alice", state)
	if err != nil {
		t.Errorf("CreateSnapshot failed: %v", err)
	}

	restored, err := store.RestoreSnapshot(snapshotID)
	if err != nil {
		t.Errorf("RestoreSnapshot failed: %v", err)
	}
	if restored["trust_alice_bob"] != 0.87 {
		t.Errorf("Restored state mismatch")
	}

	// Verify deep copy: mutating original should not affect stored snapshot
	state["trust_alice_bob"] = 0.0
	restored2, _ := store.RestoreSnapshot(snapshotID)
	if restored2["trust_alice_bob"] != 0.87 {
		t.Errorf("Snapshot should be a deep copy, got mutated value")
	}
}

// Test 6: Checkpoint Create & Restore
func TestCheckpointRoundTrip(t *testing.T) {
	store, _ := NewInMemoryStore("/tmp/test-p2p")
	defer store.Close()

	checkpointID, err := store.CreateCheckpoint("Alice", 10, "hash123", "event456")
	if err != nil {
		t.Errorf("CreateCheckpoint failed: %v", err)
	}

	checkpoint, err := store.RestoreCheckpoint(checkpointID)
	if err != nil {
		t.Errorf("RestoreCheckpoint failed: %v", err)
	}
	if checkpoint.NodeID != "Alice" || checkpoint.EventLogSize != 10 {
		t.Errorf("Checkpoint data mismatch")
	}

	latest, err := store.GetLatestCheckpoint()
	if err != nil {
		t.Errorf("GetLatestCheckpoint failed: %v", err)
	}
	if latest.ID != checkpointID {
		t.Errorf("Latest checkpoint should match most recently created")
	}
}

// Test 7: Compact Removes Old Snapshots
func TestCompact(t *testing.T) {
	store, _ := NewInMemoryStore("/tmp/test-p2p")
	defer store.Close()

	for i := 0; i < 10; i++ {
		store.CreateSnapshot("Alice", map[string]interface{}{"i": i})
	}

	statsBefore := store.GetStats()
	if statsBefore["snapshot_count"] != 10 {
		t.Errorf("Expected 10 snapshots before compact")
	}

	err := store.Compact(3)
	if err != nil {
		t.Errorf("Compact failed: %v", err)
	}

	statsAfter := store.GetStats()
	if statsAfter["snapshot_count"] != 3 {
		t.Errorf("Expected 3 snapshots after compact, got %v", statsAfter["snapshot_count"])
	}
}

// Test 8: Close Prevents Further Writes (graceful shutdown, ADR-005 lifecycle)
func TestCloseGracefulShutdown(t *testing.T) {
	store, _ := NewInMemoryStore("/tmp/test-p2p")

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	store.StoreEvent(event)

	err := store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Further writes must fail after close
	err = store.StoreEvent(event)
	if err == nil {
		t.Errorf("Expected error when writing to closed store")
	}

	_, err = store.CreateSnapshot("Alice", map[string]interface{}{})
	if err == nil {
		t.Errorf("Expected error when snapshotting closed store")
	}
}

// Bonus Test: Export produces valid JSON-shaped metadata
func TestExport(t *testing.T) {
	store, _ := NewInMemoryStore("/tmp/test-p2p")
	defer store.Close()

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	store.StoreEvent(event)

	data, err := store.Export()
	if err != nil {
		t.Errorf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("Export should return non-empty data")
	}
}
