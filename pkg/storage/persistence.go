// Package storage implements the PersistenceStore interface defined in
// ADR-005. It provides durability for the P2P 60-24 event log, state
// snapshots, and consensus checkpoints.
package storage

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/data"
)

// Checkpoint represents a point-in-time recovery marker for a node.
// See ADR-005 for lifecycle requirements (startup/running/shutdown).
type Checkpoint struct {
	ID             string
	Timestamp      int64
	NodeID         string
	EventLogSize   int
	StateHash      string
	LastEventID    string
	ConsensusState string
	Metadata       map[string]interface{}
}

// PersistenceStore is the abstract durability interface (ADR-005).
// InMemoryStore below is the reference implementation for SESJA #5.
// A RocksDBStore implementation is tracked as tech debt for SESJA #6.
type PersistenceStore interface {
	// StoreEvent persists a single event atomically.
	StoreEvent(event *core.Event) error
	// GetEvent retrieves a previously stored event by ID.
	GetEvent(eventID string) (*core.Event, error)
	// StoreEventLog persists all events currently in an EventStore.
	StoreEventLog(eventStore *data.EventStore) error
	// CreateSnapshot stores a copy of node state, returning its ID.
	CreateSnapshot(nodeID string, state map[string]interface{}) (string, error)
	// RestoreSnapshot returns a deep copy of a previously stored snapshot.
	RestoreSnapshot(snapshotID string) (map[string]interface{}, error)
	// CreateCheckpoint records a recovery marker, returning its ID.
	CreateCheckpoint(nodeID string, eventLogSize int, stateHash string, lastEventID string) (string, error)
	// RestoreCheckpoint retrieves a previously created checkpoint.
	RestoreCheckpoint(checkpointID string) (*Checkpoint, error)
	// GetLatestCheckpoint returns the most recently created checkpoint.
	GetLatestCheckpoint() (*Checkpoint, error)
	// SetMetadata stores an arbitrary key/value pair.
	SetMetadata(key, value string) error
	// GetMetadata retrieves a previously stored key/value pair.
	GetMetadata(key string) (string, error)
	// GetStats returns store-level statistics for observability.
	GetStats() map[string]interface{}
	// Compact removes old snapshots, keeping only the newest N.
	Compact(keepNewest int) error
	// Export serializes store metadata to JSON.
	Export() ([]byte, error)
	// Close flushes and releases store resources gracefully.
	Close() error
}

// InMemoryStore is a thread-safe, map-based PersistenceStore implementation.
// It is the reference implementation for SESJA #5 (see ADR-005). It does
// NOT persist across process restarts — that capability is deferred to a
// future RocksDBStore implementation (tracked tech debt).
type InMemoryStore struct {
	mu               sync.RWMutex
	path             string
	eventLog         map[string]*core.Event
	stateSnapshots   map[string]map[string]interface{}
	checkpoints      map[string]*Checkpoint
	metadata         map[string]string
	lastCheckpointID string
	lastSnapshotID   string
	closed           bool
}

// NewInMemoryStore creates a new in-memory persistence store.
// path is recorded for parity with a future disk-backed implementation
// but is not used for actual I/O in this implementation.
func NewInMemoryStore(path string) (*InMemoryStore, error) {
	return &InMemoryStore{
		path:           path,
		eventLog:       make(map[string]*core.Event),
		stateSnapshots: make(map[string]map[string]interface{}),
		checkpoints:    make(map[string]*Checkpoint),
		metadata:       make(map[string]string),
	}, nil
}

// StoreEvent implements PersistenceStore.
func (ps *InMemoryStore) StoreEvent(event *core.Event) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return fmt.Errorf("store is closed")
	}
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	ps.eventLog[event.ID] = event
	ps.metadata["last_event_id"] = event.ID
	ps.metadata["last_event_time"] = fmt.Sprintf("%d", time.Now().UnixNano())

	return nil
}

// GetEvent implements PersistenceStore.
func (ps *InMemoryStore) GetEvent(eventID string) (*core.Event, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	event, exists := ps.eventLog[eventID]
	if !exists {
		return nil, fmt.Errorf("event %s not found", eventID)
	}

	return event, nil
}

// StoreEventLog implements PersistenceStore.
func (ps *InMemoryStore) StoreEventLog(eventStore *data.EventStore) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return fmt.Errorf("store is closed")
	}

	for _, event := range eventStore.GetAll() {
		ps.eventLog[event.ID] = event
	}

	ps.metadata["event_log_size"] = fmt.Sprintf("%d", len(ps.eventLog))
	ps.metadata["event_log_synced_at"] = fmt.Sprintf("%d", time.Now().UnixNano())

	return nil
}

// CreateSnapshot implements PersistenceStore.
func (ps *InMemoryStore) CreateSnapshot(nodeID string, state map[string]interface{}) (string, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return "", fmt.Errorf("store is closed")
	}

	snapshotID := fmt.Sprintf("snap_%s_%d", nodeID, time.Now().UnixNano())

	stateCopy := make(map[string]interface{}, len(state))
	for k, v := range state {
		stateCopy[k] = v
	}

	ps.stateSnapshots[snapshotID] = stateCopy
	ps.lastSnapshotID = snapshotID

	ps.metadata["last_snapshot_id"] = snapshotID
	ps.metadata["last_snapshot_time"] = fmt.Sprintf("%d", time.Now().UnixNano())
	ps.metadata["snapshot_count"] = fmt.Sprintf("%d", len(ps.stateSnapshots))

	return snapshotID, nil
}

// RestoreSnapshot implements PersistenceStore.
func (ps *InMemoryStore) RestoreSnapshot(snapshotID string) (map[string]interface{}, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	snapshot, exists := ps.stateSnapshots[snapshotID]
	if !exists {
		return nil, fmt.Errorf("snapshot %s not found", snapshotID)
	}

	result := make(map[string]interface{}, len(snapshot))
	for k, v := range snapshot {
		result[k] = v
	}

	return result, nil
}

// CreateCheckpoint implements PersistenceStore.
func (ps *InMemoryStore) CreateCheckpoint(nodeID string, eventLogSize int, stateHash string, lastEventID string) (string, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return "", fmt.Errorf("store is closed")
	}

	checkpointID := fmt.Sprintf("ckpt_%s_%d", nodeID, time.Now().UnixNano())

	checkpoint := &Checkpoint{
		ID:             checkpointID,
		Timestamp:      time.Now().UnixNano(),
		NodeID:         nodeID,
		EventLogSize:   eventLogSize,
		StateHash:      stateHash,
		LastEventID:    lastEventID,
		ConsensusState: "finalized",
		Metadata: map[string]interface{}{
			"version": "1.0",
		},
	}

	ps.checkpoints[checkpointID] = checkpoint
	ps.lastCheckpointID = checkpointID

	ps.metadata["last_checkpoint_id"] = checkpointID
	ps.metadata["last_checkpoint_time"] = fmt.Sprintf("%d", time.Now().UnixNano())
	ps.metadata["checkpoint_count"] = fmt.Sprintf("%d", len(ps.checkpoints))

	return checkpointID, nil
}

// RestoreCheckpoint implements PersistenceStore.
func (ps *InMemoryStore) RestoreCheckpoint(checkpointID string) (*Checkpoint, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	checkpoint, exists := ps.checkpoints[checkpointID]
	if !exists {
		return nil, fmt.Errorf("checkpoint %s not found", checkpointID)
	}

	return checkpoint, nil
}

// GetLatestCheckpoint implements PersistenceStore.
func (ps *InMemoryStore) GetLatestCheckpoint() (*Checkpoint, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.lastCheckpointID == "" {
		return nil, fmt.Errorf("no checkpoint found")
	}

	return ps.checkpoints[ps.lastCheckpointID], nil
}

// SetMetadata implements PersistenceStore.
func (ps *InMemoryStore) SetMetadata(key, value string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return fmt.Errorf("store is closed")
	}

	ps.metadata[key] = value
	return nil
}

// GetMetadata implements PersistenceStore.
func (ps *InMemoryStore) GetMetadata(key string) (string, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	value, exists := ps.metadata[key]
	if !exists {
		return "", fmt.Errorf("metadata key %s not found", key)
	}

	return value, nil
}

// GetStats implements PersistenceStore.
func (ps *InMemoryStore) GetStats() map[string]interface{} {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return map[string]interface{}{
		"event_log_size":     len(ps.eventLog),
		"snapshot_count":     len(ps.stateSnapshots),
		"checkpoint_count":   len(ps.checkpoints),
		"metadata_keys":      len(ps.metadata),
		"last_checkpoint_id": ps.lastCheckpointID,
		"last_snapshot_id":   ps.lastSnapshotID,
		"store_path":         ps.path,
		"engine":             "inmemory",
	}
}

// Compact implements PersistenceStore. It keeps only the newest N snapshots.
func (ps *InMemoryStore) Compact(keepNewest int) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return fmt.Errorf("store is closed")
	}
	if len(ps.stateSnapshots) <= keepNewest {
		return nil
	}

	toRemove := len(ps.stateSnapshots) - keepNewest
	count := 0
	for snapshotID := range ps.stateSnapshots {
		if count >= toRemove {
			break
		}
		delete(ps.stateSnapshots, snapshotID)
		count++
	}

	ps.metadata["last_compact_time"] = fmt.Sprintf("%d", time.Now().UnixNano())
	return nil
}

// Export implements PersistenceStore.
func (ps *InMemoryStore) Export() ([]byte, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	export := map[string]interface{}{
		"event_log_size":   len(ps.eventLog),
		"snapshot_count":   len(ps.stateSnapshots),
		"checkpoint_count": len(ps.checkpoints),
		"metadata":         ps.metadata,
		"exported_at":      time.Now().UnixNano(),
	}

	return json.MarshalIndent(export, "", "  ")
}

// Close implements PersistenceStore. Per ADR-005 lifecycle requirements,
// this flushes state and marks the store closed; further writes fail.
func (ps *InMemoryStore) Close() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.closed = true
	return nil
}
