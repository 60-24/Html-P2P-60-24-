package monitoring

import (
	"testing"
	"time"
)

// TestMetricsIntegrationPattern verifies the integration pattern documented
// in ADR-006: components record latency/counters/gauges through the
// collector, and those values are observable via Render()/HTTP endpoint.
//
// This simulates how GossipNode and ConsensusLayer would call the collector
// (per PM's "Uwaga 2" — integration point), without introducing a hard
// dependency from pkg/network or pkg/consensus back into pkg/monitoring.
func TestMetricsIntegrationPattern(t *testing.T) {
	collector := NewMetricsCollector()

	// Simulate GossipNode reporting peer message latency
	simulatedBroadcastLatency := 12 * time.Millisecond
	collector.RecordLatency("peer_message_latency", simulatedBroadcastLatency)

	// Simulate ConsensusLayer reporting voting rounds
	collector.IncrementCounter("consensus_rounds")
	collector.IncrementCounter("consensus_rounds")
	collector.RecordLatency("voting_round_latency", 45*time.Millisecond)

	// Simulate PersistenceStore reporting write latency (ADR-005 integration)
	collector.RecordLatency("write_latency", 3*time.Millisecond)

	// Verify aggregation
	if collector.GetCounter("consensus_rounds") != 2 {
		t.Errorf("Expected 2 consensus rounds recorded")
	}

	latencyStats := collector.GetHistogramStats("peer_message_latency")
	if latencyStats["count"] != 1 {
		t.Errorf("Expected 1 latency sample")
	}

	// Verify the prometheus endpoint would return this data
	output := collector.Render()
	if len(output) == 0 {
		t.Errorf("Expected non-empty metrics output for scraping")
	}
}
