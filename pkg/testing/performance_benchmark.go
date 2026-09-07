package testing

import (
	"runtime"
	"sync"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
)

// PerformanceBenchmark mierzy wydajność systemu
type PerformanceBenchmark struct {
	mu          sync.RWMutex
	simulator   *NetworkSimulator
	results     map[string]float64
	latencies   []time.Duration
	startMem    runtime.MemStats
	endMem      runtime.MemStats
}

// NewPerformanceBenchmark tworzy benchmark
func NewPerformanceBenchmark(simulator *NetworkSimulator) *PerformanceBenchmark {
	return &PerformanceBenchmark{
		simulator:   simulator,
		results:     make(map[string]float64),
		latencies:   make([]time.Duration, 0),
	}
}

// BenchmarkEventLatency mierzy latencję eventów
func (pb *PerformanceBenchmark) BenchmarkEventLatency() map[string]float64 {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	latencies := make([]time.Duration, 0)

	for i := 0; i < 100; i++ {
		start := time.Now()
		pb.simulator.SendEvent("Alice", core.EventTypeProofOfMeeting,
			[]byte(`{"benchmark":true}`))
		latency := time.Since(start)
		latencies = append(latencies, latency)
	}

	// Calculate percentiles
	results := make(map[string]float64)

	if len(latencies) > 0 {
		// P50
		results["p50_ms"] = float64(latencies[len(latencies)/2].Microseconds()) / 1000.0

		// P95
		idx95 := (len(latencies) * 95) / 100
		if idx95 < len(latencies) {
			results["p95_ms"] = float64(latencies[idx95].Microseconds()) / 1000.0
		}

		// P99
		idx99 := (len(latencies) * 99) / 100
		if idx99 < len(latencies) {
			results["p99_ms"] = float64(latencies[idx99].Microseconds()) / 1000.0
		}

		// Average
		totalLatency := time.Duration(0)
		for _, lat := range latencies {
			totalLatency += lat
		}
		results["avg_ms"] = float64(totalLatency.Microseconds()) / 1000.0 / float64(len(latencies))
	}

	pb.results["event_latency"] = results["avg_ms"]
	return results
}

// BenchmarkConsensusTime mierzy czas osiągnięcia consensus
func (pb *PerformanceBenchmark) BenchmarkConsensusTime() map[string]float64 {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.simulator.mu.Lock()
	if len(pb.simulator.nodes) < 2 {
		pb.simulator.mu.Unlock()
		return map[string]float64{"error": 1.0}
	}
	pb.simulator.mu.Unlock()

	consensusTimes := make([]time.Duration, 0)

	for i := 0; i < 50; i++ {
		event, _ := pb.simulator.SendEvent("Alice", core.EventTypeCommitment,
			[]byte(`{"consensus":"test"}`))

		if event != nil {
			start := time.Now()

			// Wait for consensus
			pb.simulator.AwaitConsensus(event.ID, 5*time.Second)

			consensusTimes = append(consensusTimes, time.Since(start))
		}
	}

	results := make(map[string]float64)

	if len(consensusTimes) > 0 {
		totalTime := time.Duration(0)
		for _, ct := range consensusTimes {
			totalTime += ct
		}

		avgTime := totalTime / time.Duration(len(consensusTimes))
		results["avg_ms"] = float64(avgTime.Milliseconds())
		results["min_ms"] = float64(consensusTimes[0].Milliseconds())
		results["max_ms"] = float64(consensusTimes[len(consensusTimes)-1].Milliseconds())
	}

	pb.results["consensus_time"] = results["avg_ms"]
	return results
}

// BenchmarkMemoryUsage mierzy użycie pamięci
func (pb *PerformanceBenchmark) BenchmarkMemoryUsage() map[string]float64 {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	runtime.ReadMemStats(&pb.startMem)

	// Generate events
	for i := 0; i < 1000; i++ {
		pb.simulator.SendEvent("Alice", core.EventTypeProofOfMeeting,
			[]byte(`{"memory":"test"}`))
	}

	runtime.ReadMemStats(&pb.endMem)

	memUsedMB := float64(pb.endMem.Alloc-pb.startMem.Alloc) / 1024.0 / 1024.0

	results := map[string]float64{
		"alloc_mb":        float64(pb.endMem.Alloc) / 1024.0 / 1024.0,
		"used_increase_mb": memUsedMB,
		"per_event_bytes": (memUsedMB * 1024.0 * 1024.0) / 1000.0,
	}

	pb.results["memory_usage"] = memUsedMB
	return results
}

// BenchmarkThroughput mierzy throughput
func (pb *PerformanceBenchmark) BenchmarkThroughput() map[string]float64 {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	start := time.Now()
	eventCount := 0

	for i := 0; i < 500; i++ {
		_, err := pb.simulator.SendEvent("Alice", core.EventTypeProofOfMeeting,
			[]byte(`{"throughput":true}`))
		if err == nil {
			eventCount++
		}
	}

	duration := time.Since(start)
	eps := float64(eventCount) / duration.Seconds()

	results := map[string]float64{
		"total_events":    float64(eventCount),
		"duration_sec":    duration.Seconds(),
		"events_per_sec":  eps,
	}

	pb.results["throughput"] = eps
	return results
}

// GetResults zwraca wszystkie wyniki
func (pb *PerformanceBenchmark) GetResults() map[string]float64 {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	return pb.results
}
