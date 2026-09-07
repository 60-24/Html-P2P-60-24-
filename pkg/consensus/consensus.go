package consensus

import (
	"sync"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/data"
	"github.com/60-24/p2p-60-24/pkg/network"
)

// ResonanceVote reprezentuje głos w consensusie
type ResonanceVote struct {
	EventID   string    // Event being voted on
	Voter     string    // NodeID who voted
	Vote      bool      // true = accept, false = reject
	Weight    float64   // TrustMetric of voter (0-1)
	Timestamp int64     // Unix nanos
}

// ConsensusState reprezentuje stan consensusu dla event'a
type ConsensusState struct {
	EventID      string
	Event        *core.Event
	Votes        []*ResonanceVote
	Status       string // "pending", "finalized", "rejected"
	CreatedAt    time.Time
	FinalizedAt  *time.Time
	TotalWeight  float64 // Sum of vote weights
	AcceptCount  int
	RejectCount  int
}

// ConsensusLayer implementuje Byzantine Fault Tolerant consensus via Resonance
type ConsensusLayer struct {
	mu           sync.RWMutex
	nodeID       string
	dagManager   *network.DAGManager
	eventStore   *data.EventStore
	states       map[string]*ConsensusState    // EventID -> ConsensusState
	pendingVotes map[string][]*ResonanceVote   // EventID -> votes
	quorumSize   int
	threshold    float64 // >0.66 for finality
	timeoutSec   int     // Timeout for consensus
	finalized    map[string]bool               // EventID -> finalized?
}

// NewConsensusLayer tworzy nowy ConsensusLayer
func NewConsensusLayer(nodeID string, dagMgr *network.DAGManager, eventStore *data.EventStore) *ConsensusLayer {
	return &ConsensusLayer{
		nodeID:       nodeID,
		dagManager:   dagMgr,
		eventStore:   eventStore,
		states:       make(map[string]*ConsensusState),
		pendingVotes: make(map[string][]*ResonanceVote),
		quorumSize:   3,
		threshold:    0.66, // >2/3 for Byzantine tolerance
		timeoutSec:   5,
		finalized:    make(map[string]bool),
	}
}

// ProposeEvent proponuje event do consensus
func (cl *ConsensusLayer) ProposeEvent(event *core.Event) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Sprawdź czy już jest w consensus
	if _, exists := cl.states[event.ID]; exists {
		return nil // Already proposed
	}

	// Create consensus state
	state := &ConsensusState{
		EventID:   event.ID,
		Event:     event,
		Status:    "pending",
		CreatedAt: time.Now(),
		Votes:     make([]*ResonanceVote, 0),
	}

	cl.states[event.ID] = state
	cl.pendingVotes[event.ID] = make([]*ResonanceVote, 0)

	// Start timeout
	go cl.timeoutConsensus(event.ID)

	return nil
}

// SubmitVote dodaje głos do consensus
func (cl *ConsensusLayer) SubmitVote(eventID string, voter string, vote bool, weight float64) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	state, exists := cl.states[eventID]
	if !exists {
		return nil // Event not in consensus
	}

	// Check if already finalized
	if state.Status == "finalized" || state.Status == "rejected" {
		return nil
	}

	// Create vote
	v := &ResonanceVote{
		EventID:   eventID,
		Voter:     voter,
		Vote:      vote,
		Weight:    weight,
		Timestamp: time.Now().UnixNano(),
	}

	state.Votes = append(state.Votes, v)
	cl.pendingVotes[eventID] = append(cl.pendingVotes[eventID], v)

	// Update running totals
	if vote {
		state.AcceptCount++
		state.TotalWeight += weight
	} else {
		state.RejectCount++
		state.TotalWeight += weight
	}

	// Check if we can finalize
	cl.checkFinality(eventID)

	return nil
}

// checkFinality sprawdza czy consensus jest możliwy
// Resonance: >66% weighted votes = finality
func (cl *ConsensusLayer) checkFinality(eventID string) {
	state, exists := cl.states[eventID]
	if !exists || state.Status != "pending" {
		return
	}

	// Calculate weighted vote ratio
	if state.TotalWeight == 0 {
		return
	}

	acceptWeight := 0.0
	for _, v := range state.Votes {
		if v.Vote {
			acceptWeight += v.Weight
		}
	}

	ratio := acceptWeight / state.TotalWeight

	// Finality threshold: >0.66 (Byzantine tolerance)
	if ratio > cl.threshold {
		cl.finalizeEvent(eventID, true)
	} else if (1.0 - ratio) > cl.threshold {
		// Rejection: >66% reject
		cl.finalizeEvent(eventID, false)
	}
}

// finalizeEvent finalizuje event (irreversible)
func (cl *ConsensusLayer) finalizeEvent(eventID string, accepted bool) {
	state, exists := cl.states[eventID]
	if !exists {
		return
	}

	now := time.Now()
	state.FinalizedAt = &now

	if accepted {
		state.Status = "finalized"
		cl.finalized[eventID] = true

		// Store in EventStore (permanent)
		cl.eventStore.Append(state.Event)
	} else {
		state.Status = "rejected"
		cl.finalized[eventID] = false
	}
}

// GetConsensusState zwraca stan consensus dla event'a
func (cl *ConsensusLayer) GetConsensusState(eventID string) *ConsensusState {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	return cl.states[eventID]
}

// GetFinalizedEvents zwraca finalizowane event'y
func (cl *ConsensusLayer) GetFinalizedEvents() []string {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	var finalized []string
	for eventID, isFinal := range cl.finalized {
		if isFinal {
			finalized = append(finalized, eventID)
		}
	}

	return finalized
}

// timeoutConsensus obsługuje timeout consensus'a
func (cl *ConsensusLayer) timeoutConsensus(eventID string) {
	time.Sleep(time.Duration(cl.timeoutSec) * time.Second)

	cl.mu.Lock()
	defer cl.mu.Unlock()

	state, exists := cl.states[eventID]
	if !exists || state.Status != "pending" {
		return
	}

	// Timeout: use current votes to decide
	cl.checkFinality(eventID)

	// If still pending, reject
	if state.Status == "pending" {
		cl.finalizeEvent(eventID, false)
	}
}

// GetConsensusStats zwraca statystyki consensusu
func (cl *ConsensusLayer) GetConsensusStats() map[string]interface{} {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	pending := 0
	finalized := 0
	rejected := 0

	for _, state := range cl.states {
		switch state.Status {
		case "pending":
			pending++
		case "finalized":
			finalized++
		case "rejected":
			rejected++
		}
	}

	return map[string]interface{}{
		"pending_consensus":   pending,
		"finalized_events":    finalized,
		"rejected_events":     rejected,
		"threshold":           cl.threshold,
		"quorum_size":         cl.quorumSize,
		"timeout_seconds":     cl.timeoutSec,
	}
}

// BuildVotingQuorum assembluje quorum do głosowania
func (cl *ConsensusLayer) BuildVotingQuorum() []string {
	// Use DAGManager to build quorum
	return cl.dagManager.BuildQuorum(cl.nodeID)
}
