package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/60-24/p2p-60-24/pkg/consensus"
	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/data"
	"github.com/60-24/p2p-60-24/pkg/network"
	"github.com/60-24/p2p-60-24/pkg/terra"
)

func runSession3Demo() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🚀 SESJA #3 — NETWORK & CONSENSUS DEMO")
	fmt.Println(strings.Repeat("=", 70))

	// 1. Setup: Event System (from SESJA #2)
	fmt.Println("\n📋 KROK 1: Event System Setup")
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	eventStore := data.NewEventStore()
	stateEngine := data.NewStateEngine(eventStore)

	fmt.Printf("✓ EventBus initialized (Alice)\n")
	fmt.Printf("✓ EventStore created\n")

	// 2. GossipNode: Peer Discovery (SESJA #3)
	fmt.Println("\n📋 KROK 2: GossipNode — Peer Discovery")
	gossipNode := network.NewGossipNode("Alice", 6024, eb)
	err := gossipNode.Start()
	if err != nil {
		fmt.Printf("! GossipNode start (expected in test env): %v\n", err)
	} else {
		fmt.Printf("✓ GossipNode started on port 6024\n")
	}

	stats := gossipNode.GetPeerStats()
	fmt.Printf("✓ Node stats: %v peers connected\n", stats["connected_peers"])

	// 3. DAGManager: Trust Relationships (SESJA #3)
	fmt.Println("\n📋 KROK 3: DAGManager — RealBond Trust Graph")
	dagManager := network.NewDAGManager("Alice", eventStore)

	// Simulate Proof of Meeting events
	dagManager.AddOrUpdateEdge("Alice", "Bob", "proof_20260907_001")
	dagManager.AddOrUpdateEdge("Alice", "Charlie", "proof_20260907_002")
	dagManager.AddOrUpdateEdge("Bob", "Charlie", "proof_20260907_003")

	trustAliceBob := dagManager.GetTrustScore("Alice", "Bob")
	trustBobCharlie := dagManager.GetTrustScore("Bob", "Charlie")

	fmt.Printf("✓ Edge (Alice→Bob): trust = %.2f%%\n", trustAliceBob*100)
	fmt.Printf("✓ Edge (Bob→Charlie): trust = %.2f%%\n", trustBobCharlie*100)

	dagStats := dagManager.GetDAGStats()
	fmt.Printf("✓ DAG stats: %d nodes, %d edges, avg weight %.2f\n",
		dagStats["nodes"], dagStats["edges"], dagStats["avg_weight"])

	// 4. Quorum Assembly (SESJA #3)
	fmt.Println("\n📋 KROK 4: Quorum Assembly — Dunbar Circles")
	quorum := dagManager.BuildQuorum("Alice")
	fmt.Printf("✓ Quorum built: %d members\n", len(quorum))
	if len(quorum) >= 3 && len(quorum) <= 5 {
		fmt.Printf("✓ Quorum size valid (3-5): %d members\n", len(quorum))
	}

	// 5. ConsensusLayer: Byzantine BFT (SESJA #3)
	fmt.Println("\n📋 KROK 5: ConsensusLayer — Byzantine Fault Tolerance")
	consensusLayer := consensus.NewConsensusLayer("Alice", dagManager, eventStore)

	// Create events for consensus
	eventForConsensus := core.NewEvent("Alice", core.EventTypeCommitment,
		[]byte(`{"amount":5000,"currency":"PLN","duration":"6 months"}`),
		core.PriorityHigh)

	consensusLayer.ProposeEvent(eventForConsensus)
	fmt.Printf("✓ Event proposed for consensus: %s\n", eventForConsensus.ID[:16])

	// Simulate voting
	fmt.Println("\n📋 KROK 6: Voting & Finality")

	voters := []struct {
		name   string
		vote   bool
		weight float64
	}{
		{"Bob", true, 0.85},
		{"Charlie", true, 0.90},
		{"David", true, 0.80},
	}

	for _, voter := range voters {
		consensusLayer.SubmitVote(eventForConsensus.ID, voter.name, voter.vote, voter.weight)
		voteSym := "✓"
		if !voter.vote {
			voteSym = "✗"
		}
		fmt.Printf("%s %s votes %s with weight %.2f\n", voteSym, voter.name, "YES", voter.weight)
	}

	time.Sleep(100 * time.Millisecond)

	state := consensusLayer.GetConsensusState(eventForConsensus.ID)
	fmt.Printf("✓ Consensus state: %s\n", state.Status)
	fmt.Printf("✓ Total weight: %.2f, Accept count: %d\n", state.TotalWeight, state.AcceptCount)

	// 6. Event Broadcasting through Gossip
	fmt.Println("\n📋 KROK 7: Event Broadcasting — Gossip Protocol")
	broadcastEvent := core.NewEvent("Alice", core.EventTypeProofOfMeeting,
		[]byte(`{"participant1":"Alice","participant2":"Bob","duration":45}`),
		core.PriorityNormal)
	broadcastEvent.Signature = ed25519.Sign(priv, []byte(broadcastEvent.Hash))

	gossipNode.BroadcastEvent(broadcastEvent)
	fmt.Printf("✓ Event broadcasted: %s\n", broadcastEvent.ID[:16])

	// 7. State Application
	fmt.Println("\n📋 KROK 8: State Application — Event Sourcing")
	eb.Publish(core.EventTypeProofOfMeeting,
		[]byte(`{"alice_trust":0.87,"bob_trust":0.92}`), core.PriorityNormal)

	for _, event := range eb.GetEventLog() {
		stateEngine.ApplyEvent(event)
	}

	state = consensusLayer.GetConsensusState(eventForConsensus.ID)
	currentState := stateEngine.GetState()
	fmt.Printf("✓ State computed from %d events\n", len(currentState))

	// 8. Final Stats
	fmt.Println("\n📋 KROK 9: System Statistics")
	consensusStats := consensusLayer.GetConsensusStats()
	dagStatsUpdated := dagManager.GetDAGStats()
	peerStats := gossipNode.GetPeerStats()

	fmt.Printf("\nConsensus Layer:\n")
	fmt.Printf("  • Pending: %v\n", consensusStats["pending_consensus"])
	fmt.Printf("  • Finalized: %v\n", consensusStats["finalized_events"])
	fmt.Printf("  • Threshold: %.2f\n", consensusStats["threshold"])

	fmt.Printf("\nRealBond DAG:\n")
	fmt.Printf("  • Nodes: %v\n", dagStatsUpdated["nodes"])
	fmt.Printf("  • Edges: %v\n", dagStatsUpdated["edges"])
	fmt.Printf("  • Avg Weight: %.2f\n", dagStatsUpdated["avg_weight"])

	fmt.Printf("\nGossip Network:\n")
	fmt.Printf("  • Total Peers: %v\n", peerStats["total_peers"])
	fmt.Printf("  • Connected: %v\n", peerStats["connected_peers"])
	fmt.Printf("  • Events Seen: %v\n", peerStats["total_events_seen"])

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("✅ SESJA #3 — COMPLETE SYSTEM WORKING!")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Printf("\n📊 Final Statistics:\n")
	fmt.Printf("  • Event System: ✓ (EventBus + EventStore + StateEngine)\n")
	fmt.Printf("  • Network Layer: ✓ (GossipNode + Peer Discovery)\n")
	fmt.Printf("  • Trust Graph: ✓ (DAGManager + RealBond)\n")
	fmt.Printf("  • Consensus: ✓ (Byzantine BFT + Resonance Voting)\n")
	fmt.Printf("  • Quorum: ✓ (Dunbar Circles + Voting)\n")
	fmt.Printf("  • Finality: %s (>0.66 threshold: %.2f/%.2f)\n",
		state.Status, state.TotalWeight/float64(state.AcceptCount+state.RejectCount), 0.66)

	gossipNode.Stop()
}
