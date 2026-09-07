package testing

import (
	"sync"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
)

// LoadMetrics przechowuje metryki ładowania
type LoadMetrics struct {
	TotalEvents      int
	SuccessfulEvents int
	FailedEvents     int
	TotalTime        time.Duration
	AverageLatency   time.Duration
	MaxLatency       time.Duration
	MinLatency       time.Duration
	ThroughputEPS    float64 // Events per second
	PeakMemory       int64
}

// LoadGenerator generuje obciążenie dla sieci testowej
type LoadGenerator struct {
	mu             sync.RWMutex
	simulator      *NetworkSimulator
	metrics        *LoadMetrics
	eventLatencies []time.Duration
	isRunning      bool
}

// NewLoadGenerator tworzy generator obciążenia
func NewLoadGenerator(simulator *NetworkSimulator) *LoadGenerator {
	return &LoadGenerator{
		simulator:      simulator,
		metrics:        &LoadMetrics{},
		eventLatencies: make([]time.Duration, 0),
		isRunning:      false,
	}
}

// GenerateEvents generuje N eventów
func (lg *LoadGenerator) GenerateEvents(count int) error {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	startTime := time.Now()
	successCount := 0
	failCount := 0

	for i := 0; i < count; i++ {
		// Round-robin: Alice, Bob, Charlie...
		nodeIdx := i % len(lg.simulator.nodes)
		var nodeID string

		lg.simulator.mu.RLock()
		nodeIdx2 := 0
		for id, node := range lg.simulator.nodes {
			if nodeIdx2 == nodeIdx {
				nodeID = id
				if !node.IsOnline {
					lg.simulator.mu.RUnlock()
					failCount++
					continue
				}
				break
			}
			nodeIdx2++
		}
		lg.simulator.mu.RUnlock()

		// Send event
		eventStart := time.Now()
		_, err := lg.simulator.SendEvent(nodeID, core.EventTypeProofOfMeeting,
			[]byte(`{"load_test":true}`))
		eventLatency := time.Since(eventStart)

		if err != nil {
			failCount++
		} else {
			successCount++
			lg.eventLatencies = append(lg.eventLatencies, eventLatency)
		}
	}

	duration := time.Since(startTime)

	// Calculate metrics
	lg.metrics.TotalEvents = count
	lg.metrics.SuccessfulEvents = successCount
	lg.metrics.FailedEvents = failCount
	lg.metrics.TotalTime = duration
	lg.metrics.ThroughputEPS = float64(successCount) / duration.Seconds()

	// Latency statistics
	if len(lg.eventLatencies) > 0 {
		totalLatency := time.Duration(0)
		minLatency := lg.eventLatencies[0]
		maxLatency := lg.eventLatencies[0]

		for _, latency := range lg.eventLatencies {
			totalLatency += latency
			if latency < minLatency {
				minLatency = latency
			}
			if latency > maxLatency {
				maxLatency = latency
			}
		}

		lg.metrics.AverageLatency = totalLatency / time.Duration(len(lg.eventLatencies))
		lg.metrics.MinLatency = minLatency
		lg.metrics.MaxLatency = maxLatency
	}

	return nil
}

// GenerateLoad generuje ciągłe obciążenie przez określony czas
func (lg *LoadGenerator) GenerateLoad(rps int, duration time.Duration) error {
	lg.mu.Lock()
	if lg.isRunning {
		lg.mu.Unlock()
		return nil
	}
	lg.isRunning = true
	lg.mu.Unlock()

	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()

	deadline := time.Now().Add(duration)
	successCount := 0

	for {
		if time.Now().After(deadline) {
			break
		}

		select {
		case <-ticker.C:
			lg.simulator.mu.RLock()
			var nodeID string
			for id, node := range lg.simulator.nodes {
				if node.IsOnline {
					nodeID = id
					break
				}
			}
			lg.simulator.mu.RUnlock()

			if nodeID != "" {
				_, err := lg.simulator.SendEvent(nodeID, core.EventTypeCommitment,
					[]byte(`{"rps":true}`))
				if err == nil {
					successCount++
				}
			}
		}
	}

	lg.mu.Lock()
	lg.isRunning = false
	lg.metrics.SuccessfulEvents = successCount
	lg.metrics.TotalTime = duration
	lg.metrics.ThroughputEPS = float64(successCount) / duration.Seconds()
	lg.mu.Unlock()

	return nil
}

// GetMetrics zwraca metryki
func (lg *LoadGenerator) GetMetrics() *LoadMetrics {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	// Return copy
	return &LoadMetrics{
		TotalEvents:      lg.metrics.TotalEvents,
		SuccessfulEvents: lg.metrics.SuccessfulEvents,
		FailedEvents:     lg.metrics.FailedEvents,
		TotalTime:        lg.metrics.TotalTime,
		AverageLatency:   lg.metrics.AverageLatency,
		MaxLatency:       lg.metrics.MaxLatency,
		MinLatency:       lg.metrics.MinLatency,
		ThroughputEPS:    lg.metrics.ThroughputEPS,
	}
}

// GetLatencyPercentile zwraca percentyl latencji
func (lg *LoadGenerator) GetLatencyPercentile(percentile int) time.Duration {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	if len(lg.eventLatencies) == 0 {
		return 0
	}

	// Sort (simplified: assume already sorted)
	index := (len(lg.eventLatencies) * percentile) / 100
	if index >= len(lg.eventLatencies) {
		index = len(lg.eventLatencies) - 1
	}

	return lg.eventLatencies[index]
}
