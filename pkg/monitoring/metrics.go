// Package monitoring implements the observability layer defined in ADR-006.
// It exposes Prometheus-compatible metrics (counters, gauges, histograms)
// via a pull-based /metrics HTTP endpoint.
package monitoring

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// MetricsCollector accumulates counters, gauges, and latency histograms
// and renders them in Prometheus text exposition format.
type MetricsCollector struct {
	mu         sync.RWMutex
	counters   map[string]float64
	gauges     map[string]float64
	histograms map[string][]float64 // raw samples, in milliseconds
	startTime  time.Time
}

// NewMetricsCollector creates a new collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:   make(map[string]float64),
		gauges:     make(map[string]float64),
		histograms: make(map[string][]float64),
		startTime:  time.Now(),
	}
}

// IncrementCounter increments a named counter by 1.
func (mc *MetricsCollector) IncrementCounter(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters[name]++
}

// AddCounter increments a named counter by an arbitrary amount.
func (mc *MetricsCollector) AddCounter(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters[name] += value
}

// GaugeMemory sets a gauge to an absolute value (e.g. current heap size,
// current event log size). Despite the name it is used for any gauge.
func (mc *MetricsCollector) GaugeMemory(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.gauges[name] = value
}

// SetGauge is an alias for GaugeMemory with a more general name.
func (mc *MetricsCollector) SetGauge(name string, value float64) {
	mc.GaugeMemory(name, value)
}

// RecordLatency records a duration sample (in milliseconds) under a
// histogram name, e.g. "event_broadcast", "voting_round".
func (mc *MetricsCollector) RecordLatency(name string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	ms := float64(duration.Microseconds()) / 1000.0
	mc.histograms[name] = append(mc.histograms[name], ms)
}

// GetCounter returns the current value of a counter.
func (mc *MetricsCollector) GetCounter(name string) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.counters[name]
}

// GetGauge returns the current value of a gauge.
func (mc *MetricsCollector) GetGauge(name string) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.gauges[name]
}

// GetHistogramStats returns p50/p95/p99/avg for a named histogram.
func (mc *MetricsCollector) GetHistogramStats(name string) map[string]float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	samples := mc.histograms[name]
	if len(samples) == 0 {
		return map[string]float64{"count": 0}
	}

	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)

	sum := 0.0
	for _, s := range sorted {
		sum += s
	}

	return map[string]float64{
		"count": float64(len(sorted)),
		"p50":   percentile(sorted, 50),
		"p95":   percentile(sorted, 95),
		"p99":   percentile(sorted, 99),
		"avg":   sum / float64(len(sorted)),
		"min":   sorted[0],
		"max":   sorted[len(sorted)-1],
	}
}

func percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := (len(sorted) * p) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// Render produces Prometheus text exposition format output.
func (mc *MetricsCollector) Render() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# HELP p2p_uptime_seconds Node uptime in seconds\n"))
	sb.WriteString(fmt.Sprintf("# TYPE p2p_uptime_seconds counter\n"))
	sb.WriteString(fmt.Sprintf("p2p_uptime_seconds %.2f\n", time.Since(mc.startTime).Seconds()))

	// Counters
	names := make([]string, 0, len(mc.counters))
	for name := range mc.counters {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		metricName := "p2p_" + sanitizeName(name) + "_total"
		sb.WriteString(fmt.Sprintf("# TYPE %s counter\n", metricName))
		sb.WriteString(fmt.Sprintf("%s %g\n", metricName, mc.counters[name]))
	}

	// Gauges
	gaugeNames := make([]string, 0, len(mc.gauges))
	for name := range mc.gauges {
		gaugeNames = append(gaugeNames, name)
	}
	sort.Strings(gaugeNames)
	for _, name := range gaugeNames {
		metricName := "p2p_" + sanitizeName(name)
		sb.WriteString(fmt.Sprintf("# TYPE %s gauge\n", metricName))
		sb.WriteString(fmt.Sprintf("%s %g\n", metricName, mc.gauges[name]))
	}

	// Histograms (rendered as summary: p50/p95/p99)
	histNames := make([]string, 0, len(mc.histograms))
	for name := range mc.histograms {
		histNames = append(histNames, name)
	}
	sort.Strings(histNames)
	for _, name := range histNames {
		stats := mc.getHistogramStatsLocked(name)
		metricName := "p2p_" + sanitizeName(name) + "_latency_ms"
		sb.WriteString(fmt.Sprintf("# TYPE %s summary\n", metricName))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.5\"} %g\n", metricName, stats["p50"]))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.95\"} %g\n", metricName, stats["p95"]))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.99\"} %g\n", metricName, stats["p99"]))
		sb.WriteString(fmt.Sprintf("%s_count %g\n", metricName, stats["count"]))
	}

	return sb.String()
}

// getHistogramStatsLocked is like GetHistogramStats but assumes caller holds the lock.
func (mc *MetricsCollector) getHistogramStatsLocked(name string) map[string]float64 {
	samples := mc.histograms[name]
	if len(samples) == 0 {
		return map[string]float64{"count": 0, "p50": 0, "p95": 0, "p99": 0}
	}
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)
	return map[string]float64{
		"count": float64(len(sorted)),
		"p50":   percentile(sorted, 50),
		"p95":   percentile(sorted, 95),
		"p99":   percentile(sorted, 99),
	}
}

func sanitizeName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "_")
}

// ServeHTTP implements http.Handler, exposing metrics at /metrics.
func (mc *MetricsCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte(mc.Render()))
}

// StartServer starts an HTTP server exposing /metrics on the given port.
// Returns the server so the caller can gracefully shut it down.
func (mc *MetricsCollector) StartServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", mc)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go server.ListenAndServe()

	return server
}

// Reset clears all collected metrics. Primarily useful for tests.
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters = make(map[string]float64)
	mc.gauges = make(map[string]float64)
	mc.histograms = make(map[string][]float64)
}
