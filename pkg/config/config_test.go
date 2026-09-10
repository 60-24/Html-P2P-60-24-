package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Test 1: Default Config Values
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Network.Port != 6024 {
		t.Errorf("Expected default port 6024, got %d", cfg.Network.Port)
	}
	if cfg.Persistence.Engine != "inmemory" {
		t.Errorf("Expected default engine inmemory, got %s", cfg.Persistence.Engine)
	}
	if cfg.Consensus.ByzantineThreshold != 0.66 {
		t.Errorf("Expected default threshold 0.66, got %f", cfg.Consensus.ByzantineThreshold)
	}
}

// Test 2: Load & Parse YAML File
func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-config.yaml")

	yamlContent := `
network:
  port: 7000
  peers:
    - "peer1:6024"
persistence:
  engine: inmemory
  path: /tmp/data
monitoring:
  prometheus_port: 9091
consensus:
  timeout_sec: 10
  byzantine_threshold: 0.75
`
	os.WriteFile(path, []byte(yamlContent), 0644)

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if cfg.Network.Port != 7000 {
		t.Errorf("Expected port 7000, got %d", cfg.Network.Port)
	}
	if len(cfg.Network.Peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(cfg.Network.Peers))
	}
	if cfg.Consensus.TimeoutSec != 10 {
		t.Errorf("Expected timeout 10, got %d", cfg.Consensus.TimeoutSec)
	}
}

// Test 3: Environment Variable Overrides (12-factor pattern)
func TestEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-config.yaml")
	os.WriteFile(path, []byte("network:\n  port: 6024\npersistence:\n  engine: inmemory\n  path: /tmp\nmonitoring:\n  prometheus_port: 9090\nconsensus:\n  timeout_sec: 5\n  byzantine_threshold: 0.66\n"), 0644)

	os.Setenv("P2P_NETWORK_PORT", "8080")
	os.Setenv("P2P_PERSISTENCE_PATH", "/custom/path")
	defer os.Unsetenv("P2P_NETWORK_PORT")
	defer os.Unsetenv("P2P_PERSISTENCE_PATH")

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if cfg.Network.Port != 8080 {
		t.Errorf("Expected env override port 8080, got %d", cfg.Network.Port)
	}
	if cfg.Persistence.Path != "/custom/path" {
		t.Errorf("Expected env override path /custom/path, got %s", cfg.Persistence.Path)
	}
}

// Test 4: Validation - Required Fields & Range Checks
func TestValidationRejectsInvalidPort(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Network.Port = 80 // below 1024, invalid

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for port 80")
	}
}

func TestValidationRejectsUnknownEngine(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Persistence.Engine = "mongodb" // not "inmemory" or "rocksdb"

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for unknown engine")
	}
}

func TestValidationRejectsRocksDBWithoutPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Persistence.Engine = "rocksdb"
	cfg.Persistence.Path = ""

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for rocksdb without path")
	}
}

// Test 5: Validation - Byzantine Threshold Range
func TestValidationRejectsInvalidThreshold(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Consensus.ByzantineThreshold = 1.5 // out of [0,1]

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for threshold > 1")
	}
}

// Test 6: Port Collision Detection (network vs monitoring)
func TestValidationRejectsPortCollision(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Network.Port = 9090
	cfg.Monitoring.PrometheusPort = 9090 // same as network port

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for port collision")
	}
}

// Bonus: ToYAML round trip (used by `p2p config show`)
func TestToYAML(t *testing.T) {
	cfg := DefaultConfig()

	yamlStr, err := cfg.ToYAML()
	if err != nil {
		t.Errorf("ToYAML failed: %v", err)
	}
	if len(yamlStr) == 0 {
		t.Errorf("Expected non-empty YAML output")
	}
}

// Bonus: Missing file produces clear error
func TestLoadFromFileMissing(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Errorf("Expected error for missing config file")
	}
}
