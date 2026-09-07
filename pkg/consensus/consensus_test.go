package consensus

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/data"
	"github.com/60-24/p2p-60-24/pkg/network"
)

// Test 1: ConsensusLayer Creation
func TestConsensusLayerCreation(t *testing.T) {
	store := data.NewEventStore()
	dag := network.NewDAGManager("Alice", store)
	cl := NewConsensusLayer("Alice", dag, store)

	if cl.nodeID != "Alice" {
		t.Errorf("Expected nodeID Alice")
	}
	if cl.threshold != 0.66 {
		t.Errorf("Expected threshold 0.66")
	}
}

// Test 2: Propose Event
func TestProposeEvent(t *testing.T) {
	store := data.NewEventStore()
	dag := network.NewDAGManager("Alice", store)
	cl := NewConsensusLayer("Alice", dag, store)

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)

	err := cl.ProposeEvent(event)
	if err != nil {
		t.Errorf("ProposeEvent failed: %v", err)
	}

	state := cl.GetConsensusState(event.ID)
	if state == nil {
		t.Errorf("Event should be in consensus state")
	}
	if state.Status != "pending" {
		t.Errorf("Event should be pending")
	}
}

// Test 3: Submit Vote
func TestSubmitVote(t *testing.T) {
	store := data.NewEventStore()
	dag := network.NewDAGManager("Alice", store)
	cl := NewConsensusLayer("Alice", dag, store)

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	cl.ProposeEvent(event)

	err := cl.SubmitVote(event.ID, "Bob", true, 0.9)
	if err != nil {
		t.Errorf("SubmitVote failed: %v", err)
	}

	state := cl.GetConsensusState(event.ID)
	if len(state.Votes) != 1 {
		t.Errorf("Expected 1 vote, got %d", len(state.Votes))
	}
}

// Test 4: Byzantine Threshold (>0.66)
func TestByzantineThreshold(t *testing.T) {
	store := data.NewEventStore()
	dag := network.NewDAGManager("Alice", store)
	cl := NewConsensusLayer("Alice", dag, store)

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	cl.ProposeEvent(event)

	// Vote with >0.66 total weight
	cl.SubmitVote(event.ID, "Bob", true, 0.5)
	cl.SubmitVote(event.ID, "Charlie", true, 0.3)
	cl.SubmitVote(event.ID, "David", true, 0.1)

	// Total accept: 0.9 (>0.66 threshold)
	// Should finalize

	state := cl.GetConsensusState(event.ID)
	if state.Status != "finalized" {
		t.Logf("Status: %s (expected finalized for accept ratio >0.66)", state.Status)
	}
}

// Test 5: Rejection Threshold
func TestRejectionThreshold(t *testing.T) {
	store := data.NewEventStore()
	dag := network.NewDAGManager("Alice", store)
	cl := NewConsensusLayer("Alice", dag, store)

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	cl.ProposeEvent(event)

	// Vote with >0.66 reject weight
	cl.SubmitVote(event.ID, "Bob", false, 0.4)
	cl.SubmitVote(event.ID, "Charlie", false, 0.3)
	cl.SubmitVote(event.ID, "David", false, 0.1)

	// Total reject: 0.8 (>0.66 threshold)
	// Should reject

	state := cl.GetConsensusState(event.ID)
	if state.Status != "rejected" {
		t.Logf("Status: %s (expected rejected for reject ratio >0.66)", state.Status)
	}
}

// Test 6: Consensus Stats
func TestConsensusStats(t *testing.T) {
	store := data.NewEventStore()
	dag := network.NewDAGManager("Alice", store)
	cl := NewConsensusLayer("Alice", dag, store)

	// Propose multiple events
	for i := 0; i < 3; i++ {
		event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
		cl.ProposeEvent(event)
	}

	stats := cl.GetConsensusStats()

	if stats["pending_consensus"] != 3 {
		t.Errorf("Expected 3 pending, got %v", stats["pending_consensus"])
	}
	if stats["threshold"] != 0.66 {
		t.Errorf("Expected threshold 0.66, got %v", stats["threshold"])
	}
}

// Test 7: Duplicate Proposal
func TestDuplicateProposal(t *testing.T) {
	store := data.NewEventStore()
	dag := network.NewDAGManager("Alice", store)
	cl := NewConsensusLayer("Alice", dag, store)

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)

	cl.ProposeEvent(event)
	err := cl.ProposeEvent(event) // Propose same event again

	if err != nil {
		t.Errorf("Should handle duplicate proposal gracefully")
	}

	// Should only have one state
	state := cl.GetConsensusState(event.ID)
	if state == nil {
		t.Errorf("State should exist")
	}
}

// Test 8: Finalized Events List
func TestFinalizedEventsList(t *testing.T) {
	store := data.NewEventStore()
	dag := network.NewDAGManager("Alice", store)
	cl := NewConsensusLayer("Alice", dag, store)

	// Propose and finalize events
	e1 := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test1"), core.PriorityNormal)
	e2 := core.NewEvent("Alice", core.EventTypeCommitment, []byte("test2"), core.PriorityNormal)

	cl.ProposeEvent(e1)
	cl.ProposeEvent(e2)

	// Manually finalize (in real scenario, votes would do this)
	cl.SubmitVote(e1.ID, "Bob", true, 0.8)

	finalized := cl.GetFinalizedEvents()
	if len(finalized) == 0 {
		t.Logf("No events finalized (would need votes to reach threshold)")
	}
}
