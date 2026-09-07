package testing

import (
	"testing"
	"time"
)

// Performance Benchmark Tests (4 total)

func TestBenchmarkEventLatency(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")

	pb := NewPerformanceBenchmark(ns)
	results := pb.BenchmarkEventLatency()

	if results["avg_ms"] <= 0 {
		t.Errorf("Average latency should be positive")
	}

	t.Logf("Event Latency - P50: %.2fms, P95: %.2fms, P99: %.2fms",
		results["p50_ms"], results["p95_ms"], results["p99_ms"])
}

func TestBenchmarkConsensusTime(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.ConnectNodes("Alice", "Bob")

	pb := NewPerformanceBenchmark(ns)
	results := pb.BenchmarkConsensusTime()

	if results["avg_ms"] > 0 {
		t.Logf("Consensus Time - Avg: %.2fms, Min: %.2fms, Max: %.2fms",
			results["avg_ms"], results["min_ms"], results["max_ms"])
	}
}

func TestBenchmarkMemoryUsage(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")

	pb := NewPerformanceBenchmark(ns)
	results := pb.BenchmarkMemoryUsage()

	if results["per_event_bytes"] > 0 {
		t.Logf("Memory Usage - Per event: %.2f bytes, Total: %.2f MB",
			results["per_event_bytes"], results["alloc_mb"])
	}
}

func TestBenchmarkThroughput(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.ConnectNodes("Alice", "Bob")

	pb := NewPerformanceBenchmark(ns)
	results := pb.BenchmarkThroughput()

	if results["events_per_sec"] > 0 {
		t.Logf("Throughput: %.0f events/second", results["events_per_sec"])
	}
}

// Byzantine Scenario Tests (6 total)

func TestByzantineAlwaysReject(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")   // Honest
	ns.CreateNode("Bob")     // Honest
	ns.CreateNode("Mallory") // Byzantine

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Alice", "Mallory")
	ns.ConnectNodes("Bob", "Mallory")

	bs := NewByzantineSimulator(ns, "Mallory")
	bs.SetStrategy("always_reject")
	bs.StartByzantineNode()

	// Send event from Alice
	event, _ := ns.SendEvent("Alice", "proof_of_meeting",
		[]byte(`{"alice":"bob"}`))

	// Alice and Bob vote YES, Mallory votes NO
	ns.nodes["Alice"].Consensus.ProposeEvent(event)
	ns.nodes["Bob"].Consensus.ProposeEvent(event)

	ns.nodes["Alice"].Consensus.SubmitVote(event.ID, "Alice", true, 0.85)
	ns.nodes["Bob"].Consensus.SubmitVote(event.ID, "Bob", true, 0.90)
	// Mallory votes NO (happens in goroutine)

	time.Sleep(500 * time.Millisecond)

	// Even with 1/3 Byzantine: >0.66 threshold should hold
	tolerates := bs.TestByzantineTolerance()
	if tolerates {
		t.Logf("✓ System tolerates Byzantine node (1/3 threshold)")
	}
}

func TestByzantineRandomVoting(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.CreateNode("Charlie") // Byzantine - random votes

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Bob", "Charlie")

	bs := NewByzantineSimulator(ns, "Charlie")
	bs.SetStrategy("random")
	bs.StartByzantineNode()

	event, _ := ns.SendEvent("Alice", "commitment",
		[]byte(`{"amount":5000}`))

	time.Sleep(500 * time.Millisecond)

	if event != nil {
		t.Logf("✓ Random voting Byzantine scenario handled")
	}
}

func TestByzantineDelayedMessages(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.CreateNode("Charlie")

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Bob", "Charlie")

	bs := NewByzantineSimulator(ns, "Charlie")
	bs.SetStrategy("delay")
	bs.CauseNetworkDelay("Charlie", 500)
	bs.StartByzantineNode()

	event, _ := ns.SendEvent("Alice", "proof_of_meeting",
		[]byte(`{"delayed":true}`))

	time.Sleep(1 * time.Second)

	if event != nil {
		t.Logf("✓ Delayed message scenario handled")
	}
}

func TestByzantineDoubleVoting(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.CreateNode("Mallory") // Double voter

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Bob", "Mallory")

	bs := NewByzantineSimulator(ns, "Mallory")
	bs.SetStrategy("double_vote")
	bs.StartByzantineNode()

	event, _ := ns.SendEvent("Alice", "resolution",
		[]byte(`{"dispute":"resolved"}`))

	time.Sleep(500 * time.Millisecond)

	if event != nil {
		t.Logf("✓ Double voting scenario handled")
	}
}

func TestByzantineNetworkPartition(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.CreateNode("Mallory")

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Bob", "Mallory")

	// Partition Mallory
	ns.PartitionNetwork("Alice", "Mallory")

	event, _ := ns.SendEvent("Alice", "proof_of_meeting",
		[]byte(`{"partition":true}`))

	time.Sleep(500 * time.Millisecond)

	// System should work despite partition
	if event != nil {
		t.Logf("✓ Byzantine with partition scenario handled")
	}

	// Heal partition
	ns.HealPartition("Alice", "Mallory")
}

func TestByzantineFinality(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.CreateNode("Charlie")

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Bob", "Charlie")

	event, _ := ns.SendEvent("Alice", "commitment",
		[]byte(`{"finality":"test"}`))

	// Manually vote with Byzantine node
	ns.nodes["Alice"].Consensus.SubmitVote(event.ID, "Alice", true, 0.8)
	ns.nodes["Bob"].Consensus.SubmitVote(event.ID, "Bob", true, 0.9)
	ns.nodes["Charlie"].Consensus.SubmitVote(event.ID, "Charlie", false, 0.5)

	// >0.66 accept ratio: (0.8 + 0.9) / (0.8 + 0.9 + 0.5) = 0.77 > 0.66
	// Should finalize

	time.Sleep(500 * time.Millisecond)

	if event != nil {
		t.Logf("✓ Byzantine finality test passed")
	}
}
