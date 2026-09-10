package monitoring

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Test 1: Counter Increment
func TestCounterIncrement(t *testing.T) {
	mc := NewMetricsCollector()

	mc.IncrementCounter("consensus_rounds")
	mc.IncrementCounter("consensus_rounds")
	mc.IncrementCounter("consensus_rounds")

	value := mc.GetCounter("consensus_rounds")
	if value != 3 {
		t.Errorf("Expected counter=3, got %f", value)
	}
}

// Test 2: Counter Add (arbitrary amount)
func TestCounterAdd(t *testing.T) {
	mc := NewMetricsCollector()

	mc.AddCounter("events_processed", 42)
	mc.AddCounter("events_processed", 8)

	value := mc.GetCounter("events_processed")
	if value != 50 {
		t.Errorf("Expected counter=50, got %f", value)
	}
}

// Test 3: Gauge Set
func TestGaugeSet(t *testing.T) {
	mc := NewMetricsCollector()

	mc.GaugeMemory("event_log_size", 128)
	mc.GaugeMemory("event_log_size", 256) // overwrite, not accumulate

	value := mc.GetGauge("event_log_size")
	if value != 256 {
		t.Errorf("Expected gauge=256 (latest value), got %f", value)
	}
}

// Test 4: Latency Histogram Percentiles
func TestLatencyHistogram(t *testing.T) {
	mc := NewMetricsCollector()

	// Record 100 samples: 1ms to 100ms
	for i := 1; i <= 100; i++ {
		mc.RecordLatency("event_broadcast", time.Duration(i)*time.Millisecond)
	}

	stats := mc.GetHistogramStats("event_broadcast")

	if stats["count"] != 100 {
		t.Errorf("Expected 100 samples, got %f", stats["count"])
	}
	if stats["p50"] < 40 || stats["p50"] > 60 {
		t.Errorf("Expected p50 around 50ms, got %f", stats["p50"])
	}
	if stats["p99"] < stats["p50"] {
		t.Errorf("p99 should be >= p50")
	}
}

// Test 5: Prometheus Format Rendering
func TestPrometheusRender(t *testing.T) {
	mc := NewMetricsCollector()

	mc.IncrementCounter("consensus_rounds")
	mc.GaugeMemory("peer_count", 5)
	mc.RecordLatency("voting_round", 50*time.Millisecond)

	output := mc.Render()

	if !strings.Contains(output, "p2p_consensus_rounds_total") {
		t.Errorf("Expected counter in output, got: %s", output)
	}
	if !strings.Contains(output, "p2p_peer_count") {
		t.Errorf("Expected gauge in output")
	}
	if !strings.Contains(output, "p2p_voting_round_latency_ms") {
		t.Errorf("Expected histogram in output")
	}
	if !strings.Contains(output, "p2p_uptime_seconds") {
		t.Errorf("Expected uptime metric in output")
	}
}

// Test 6: HTTP /metrics Endpoint (integration - verifies Prometheus can scrape it)
func TestMetricsHTTPEndpoint(t *testing.T) {
	mc := NewMetricsCollector()
	mc.IncrementCounter("finalized_events")
	mc.GaugeMemory("peer_count", 3)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	mc.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Expected text/plain content type, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "p2p_finalized_events_total") {
		t.Errorf("Expected finalized_events metric in response body")
	}
}
