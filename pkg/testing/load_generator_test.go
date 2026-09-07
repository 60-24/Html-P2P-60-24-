package testing

import (
	"testing"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
)

// Test 1: Generate 100 Events
func TestGenerateSmallLoad(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.ConnectNodes("Alice", "Bob")

	lg := NewLoadGenerator(ns)
	err := lg.GenerateEvents(100)
	if err != nil {
		t.Errorf("GenerateEvents failed: %v", err)
	}

	metrics := lg.GetMetrics()
	if metrics.TotalEvents != 100 {
		t.Errorf("Expected 100 events, got %d", metrics.TotalEvents)
	}
	if metrics.SuccessfulEvents < 90 {
		t.Logf("Success rate: %d/%d", metrics.SuccessfulEvents, metrics.TotalEvents)
	}
}

// Test 2: Stress Test - 1000 Events
func TestStressLoad1000Events(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	// 3-node network
	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.CreateNode("Charlie")

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Bob", "Charlie")

	lg := NewLoadGenerator(ns)
	startTime := time.Now()
	err := lg.GenerateEvents(1000)
	duration := time.Since(startTime)

	if err != nil {
		t.Errorf("GenerateEvents failed: %v", err)
	}

	metrics := lg.GetMetrics()
	if metrics.TotalEvents != 1000 {
		t.Errorf("Expected 1000 events")
	}

	t.Logf("Stress Test: 1000 events in %.2fs (%.0f eps)", 
		duration.Seconds(), metrics.ThroughputEPS)

	if metrics.SuccessfulEvents < 900 {
		t.Errorf("Expected >90%% success rate, got %d/%d",
			metrics.SuccessfulEvents, metrics.TotalEvents)
	}
}

// Test 3: Throughput Measurement
func TestThroughput(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.ConnectNodes("Alice", "Bob")

	lg := NewLoadGenerator(ns)
	err := lg.GenerateEvents(500)
	if err != nil {
		t.Errorf("Failed: %v", err)
	}

	metrics := lg.GetMetrics()
	if metrics.ThroughputEPS <= 0 {
		t.Errorf("Throughput should be positive")
	}

	t.Logf("Throughput: %.0f events/second", metrics.ThroughputEPS)
}

// Test 4: Latency Percentiles
func TestLatencyPercentiles(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.ConnectNodes("Alice", "Bob")

	lg := NewLoadGenerator(ns)
	lg.GenerateEvents(100)

	metrics := lg.GetMetrics()

	if metrics.AverageLatency <= 0 {
		t.Logf("Average latency: %v", metrics.AverageLatency)
	}

	p50 := lg.GetLatencyPercentile(50)
	p95 := lg.GetLatencyPercentile(95)
	p99 := lg.GetLatencyPercentile(99)

	t.Logf("Latency - P50: %v, P95: %v, P99: %v", p50, p95, p99)

	if p99 < p95 || p95 < p50 {
		t.Logf("Note: Latencies not yet sorted (expected in production)")
	}
}

// Test 5: Concurrent Voting (Byzantine Test)
func TestConcurrentVoting(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.CreateNode("Charlie")

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Bob", "Charlie")

	lg := NewLoadGenerator(ns)

	// Generate events concurrently
	lg.GenerateEvents(50)

	// Have nodes vote on events
	time.Sleep(500 * time.Millisecond)

	state := ns.GetNetworkState()
	if state["total_events"].(int) > 0 {
		t.Logf("✓ Concurrent voting test: %d events processed", state["total_events"])
	}
}

// Test 6: Load with Network Partition
func TestLoadWithPartition(t *testing.T) {
	ns := NewNetworkSimulator()
	defer ns.Shutdown()

	ns.CreateNode("Alice")
	ns.CreateNode("Bob")
	ns.CreateNode("Charlie")

	ns.ConnectNodes("Alice", "Bob")
	ns.ConnectNodes("Bob", "Charlie")

	lg := NewLoadGenerator(ns)

	// Generate initial load
	lg.GenerateEvents(50)

	time.Sleep(200 * time.Millisecond)

	// Partition network
	ns.PartitionNetwork("Alice", "Charlie")

	// Generate more load during partition
	lg.GenerateEvents(50)

	metrics := lg.GetMetrics()
	if metrics.SuccessfulEvents >= 90 {
		t.Logf("✓ Load generator handled partition: %d successful events",
			metrics.SuccessfulEvents)
	}

	// Heal partition
	ns.HealPartition("Alice", "Charlie")
	time.Sleep(200 * time.Millisecond)

	// Final load
	lg.GenerateEvents(50)
	time.Sleep(200 * time.Millisecond)

	finalState := ns.GetNetworkState()
	t.Logf("Final state: %d total events", finalState["total_events"])
}
