package network

import (
	"testing"

	"github.com/60-24/p2p-60-24/pkg/data"
)

// Test 1: DAGManager Creation
func TestDAGManagerCreation(t *testing.T) {
	store := data.NewEventStore()
	dag := NewDAGManager("Alice", store)

	if dag.nodeID != "Alice" {
		t.Errorf("Expected nodeID Alice")
	}
	if len(dag.edges) != 0 {
		t.Errorf("Expected empty edges initially")
	}
}

// Test 2: Add Edge (Trust Relationship)
func TestAddEdge(t *testing.T) {
	store := data.NewEventStore()
	dag := NewDAGManager("Alice", store)

	err := dag.AddOrUpdateEdge("Alice", "Bob", "proof123")
	if err != nil {
		t.Errorf("AddOrUpdateEdge failed: %v", err)
	}

	trust := dag.GetTrustScore("Alice", "Bob")
	if trust <= 0 || trust > 1.0 {
		t.Errorf("TrustScore should be between 0-1, got %f", trust)
	}
}

// Test 3: TrustMetric Calculation
func TestTrustMetricCalculation(t *testing.T) {
	store := data.NewEventStore()
	dag := NewDAGManager("Alice", store)

	// Add multiple edges to increase trust
	dag.AddOrUpdateEdge("Alice", "Bob", "proof1")
	dag.AddOrUpdateEdge("Alice", "Bob", "proof2") // Same edge twice
	dag.AddOrUpdateEdge("Alice", "Bob", "proof3")

	trust := dag.GetTrustScore("Alice", "Bob")

	// Should increase with more events
	if trust < 0.5 {
		t.Errorf("TrustScore should increase, got %f", trust)
	}
}

// Test 4: Non-existent Relationship
func TestNonExistentRelationship(t *testing.T) {
	store := data.NewEventStore()
	dag := NewDAGManager("Alice", store)

	trust := dag.GetTrustScore("Alice", "Charlie")
	if trust != 0.0 {
		t.Errorf("Non-existent relationship should have 0 trust")
	}
}

// Test 5: Build Quorum
func TestBuildQuorum(t *testing.T) {
	store := data.NewEventStore()
	dag := NewDAGManager("Alice", store)

	// Add some relationships
	for i := 0; i < 10; i++ {
		dag.AddOrUpdateEdge("Alice", "Peer"+string(rune(i)), "proof")
	}

	quorum := dag.BuildQuorum("Alice")

	// Quorum size: max(3, min(5, floor(N/2)))
	// With 10 peers: floor(10/2) = 5
	if len(quorum) > 5 || len(quorum) < 3 {
		t.Errorf("Quorum size should be 3-5, got %d", len(quorum))
	}
}

// Test 6: Reputation Score (Multiplicative)
func TestReputationScore(t *testing.T) {
	store := data.NewEventStore()
	dag := NewDAGManager("Alice", store)

	// Build reputation network
	dag.AddOrUpdateEdge("Alice", "Bob", "proof1")
	dag.AddOrUpdateEdge("Alice", "Charlie", "proof2")
	dag.AddOrUpdateEdge("Alice", "David", "proof3")

	rep := dag.GetReputationScore("Alice")

	// Reputation should be product of scores
	if rep <= 0 || rep >= 1.0 {
		t.Logf("Reputation score: %f (multiplicative model)", rep)
	}
}

// Test 7: DAG Stats
func TestDAGStats(t *testing.T) {
	store := data.NewEventStore()
	dag := NewDAGManager("Alice", store)

	dag.AddOrUpdateEdge("Alice", "Bob", "proof1")
	dag.AddOrUpdateEdge("Bob", "Charlie", "proof2")
	dag.AddOrUpdateEdge("Charlie", "David", "proof3")

	stats := dag.GetDAGStats()

	if stats["edges"] != 3 {
		t.Errorf("Expected 3 edges, got %v", stats["edges"])
	}

	if stats["nodes"] == 0 {
		t.Logf("DAG Stats: %v", stats)
	}
}

// Test 8: Dunbar Circle Update
func TestDunbarCircleUpdate(t *testing.T) {
	store := data.NewEventStore()
	dag := NewDAGManager("Alice", store)

	// Add relationships up to Dunbar circle 1 (5 people)
	for i := 0; i < 5; i++ {
		dag.AddOrUpdateEdge("Alice", "Circle1_Peer"+string(rune(i)), "proof")
	}

	// Add more to circle 2 (15 people)
	for i := 0; i < 15; i++ {
		dag.AddOrUpdateEdge("Alice", "Circle2_Peer"+string(rune(i)), "proof")
	}

	stats := dag.GetDAGStats()
	expectedEdges := 5 + 15
	if stats["edges"] != expectedEdges {
		t.Errorf("Expected %d edges (Dunbar circles), got %v", expectedEdges, stats["edges"])
	}
}
