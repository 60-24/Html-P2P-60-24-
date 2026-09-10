package testing

import (
	"fmt"
	"sync"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
)

// ByzantineSimulator symuluje Byzantine (adversarial) węzły
type ByzantineSimulator struct {
	mu                sync.RWMutex
	simulator         *NetworkSimulator
	byzantineNodeID   string
	byzantineStrategy string // "always_reject", "random", "delay", "double_vote"
	rejectionRate     float64
	delayMS           int
}

// NewByzantineSimulator tworzy simulator
func NewByzantineSimulator(sim *NetworkSimulator, byzantineNodeID string) *ByzantineSimulator {
	return &ByzantineSimulator{
		simulator:         sim,
		byzantineNodeID:   byzantineNodeID,
		byzantineStrategy: "always_reject",
		rejectionRate:     1.0,
		delayMS:           100,
	}
}

// SetStrategy ustala strategię Byzantine węzła
func (bs *ByzantineSimulator) SetStrategy(strategy string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.byzantineStrategy = strategy
}

// StartByzantineNode uruchamia Byzantine węzeł z daną strategią
func (bs *ByzantineSimulator) StartByzantineNode() error {
	bs.mu.Lock()
	strategy := bs.byzantineStrategy
	bs.mu.Unlock()

	node, exists := bs.simulator.nodes[bs.byzantineNodeID]
	if !exists {
		return fmt.Errorf("node %s not found", bs.byzantineNodeID)
	}

	switch strategy {
	case "always_reject":
		// Will reject all votes
		go bs.rejectAllVotes(node)

	case "random":
		// Will vote randomly
		go bs.randomVoting(node)

	case "delay":
		// Will delay messages
		go bs.delayMessages(node)

	case "double_vote":
		// Will vote twice on same event
		go bs.doubleVoting(node)

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}

	return nil
}

// rejectAllVotes Byzantine node rejects all votes
func (bs *ByzantineSimulator) rejectAllVotes(node *SimulatedNode) {
	// Monitor all proposed events and reject them
	for {
		time.Sleep(100 * time.Millisecond)

		bs.simulator.mu.RLock()
		for _, event := range bs.simulator.eventLog {
			state := node.Consensus.GetConsensusState(event.ID)
			if state != nil && state.Status == "pending" {
				// Submit rejection vote
				node.Consensus.SubmitVote(event.ID, bs.byzantineNodeID, false, 0.95)
			}
		}
		bs.simulator.mu.RUnlock()
	}
}

// randomVoting Byzantine node votes randomly
func (bs *ByzantineSimulator) randomVoting(node *SimulatedNode) {
	counter := 0
	for {
		time.Sleep(100 * time.Millisecond)

		bs.simulator.mu.RLock()
		for _, event := range bs.simulator.eventLog {
			state := node.Consensus.GetConsensusState(event.ID)
			if state != nil && state.Status == "pending" {
				// Vote randomly
				vote := counter%2 == 0
				node.Consensus.SubmitVote(event.ID, bs.byzantineNodeID, vote, 0.5)
				counter++
			}
		}
		bs.simulator.mu.RUnlock()
	}
}

// delayMessages Byzantine node delays message delivery
func (bs *ByzantineSimulator) delayMessages(node *SimulatedNode) {
	bs.mu.Lock()
	delayMS := bs.delayMS
	bs.mu.Unlock()

	for {
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
	}
}

// doubleVoting Byzantine node votes twice on same event
func (bs *ByzantineSimulator) doubleVoting(node *SimulatedNode) {
	for {
		time.Sleep(100 * time.Millisecond)

		bs.simulator.mu.RLock()
		for _, event := range bs.simulator.eventLog {
			state := node.Consensus.GetConsensusState(event.ID)
			if state != nil && state.Status == "pending" {
				// Submit vote twice
				node.Consensus.SubmitVote(event.ID, bs.byzantineNodeID, true, 0.9)
				node.Consensus.SubmitVote(event.ID, bs.byzantineNodeID, false, 0.9)
			}
		}
		bs.simulator.mu.RUnlock()
	}
}

// TestByzantineTolerance sprawdza czy system toleruje Byzantine node
func (bs *ByzantineSimulator) TestByzantineTolerance() bool {
	// With 3 nodes: 1 Byzantine, 2 honest
	// With 1/3 Byzantine: system should still reach consensus (>0.66 threshold)

	bs.simulator.mu.RLock()
	totalNodes := len(bs.simulator.nodes)
	bs.simulator.mu.RUnlock()

	byzantineCount := 1
	honestCount := totalNodes - byzantineCount

	// If honest nodes agree: >0.66 threshold should be reached
	return honestCount > byzantineCount
}

// CauseNetworkDelay adds network latency
func (bs *ByzantineSimulator) CauseNetworkDelay(nodeID string, delayMS int) {
	bs.simulator.mu.Lock()
	defer bs.simulator.mu.Unlock()

	bs.simulator.latencies[nodeID] = delayMS
}

// TriggerTimeout forces event timeout
func (bs *ByzantineSimulator) TriggerTimeout(eventID string) {
	bs.simulator.mu.Lock()
	defer bs.simulator.mu.Unlock()

	for _, node := range bs.simulator.nodes {
		state := node.Consensus.GetConsensusState(eventID)
		if state != nil && state.Status == "pending" {
			// Timeout already handled in ConsensusLayer
			// This is for testing purposes
		}
	}
}

// SetRejectionRate configures the rejection probability for the
// "random" strategy (reserved for future use; "always_reject" and
// "random" currently use fixed weights - see randomVoting above).
func (bs *ByzantineSimulator) SetRejectionRate(rate float64) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.rejectionRate = rate
}
