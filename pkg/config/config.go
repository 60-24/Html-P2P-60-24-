// Package config implements the configuration management strategy defined
// in ADR-007: YAML files with environment-variable overrides (12-factor
// app pattern).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// NetworkConfig holds network-layer settings.
type NetworkConfig struct {
	Port  int      `yaml:"port"`
	Peers []string `yaml:"peers"`
}

// PersistenceConfig holds storage engine settings (see ADR-005).
type PersistenceConfig struct {
	Engine string `yaml:"engine"` // "inmemory" | "rocksdb"
	Path   string `yaml:"path"`
}

// MonitoringConfig holds observability settings (see ADR-006).
type MonitoringConfig struct {
	PrometheusPort int `yaml:"prometheus_port"`
}

// ConsensusConfig holds Byzantine consensus tuning (see SESJA #3).
type ConsensusConfig struct {
	TimeoutSec         int     `yaml:"timeout_sec"`
	ByzantineThreshold float64 `yaml:"byzantine_threshold"`
}

// Config is the root configuration structure, matching the YAML schema
// documented in ADR-007.
type Config struct {
	Network     NetworkConfig     `yaml:"network"`
	Persistence PersistenceConfig `yaml:"persistence"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	Consensus   ConsensusConfig   `yaml:"consensus"`
}

// DefaultConfig returns a Config populated with sane defaults, matching
// config/default.yaml.
func DefaultConfig() *Config {
	return &Config{
		Network: NetworkConfig{
			Port:  6024,
			Peers: []string{},
		},
		Persistence: PersistenceConfig{
			Engine: "inmemory",
			Path:   "/var/p2p/data",
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: 9090,
		},
		Consensus: ConsensusConfig{
			TimeoutSec:         5,
			ByzantineThreshold: 0.66,
		},
	}
}

// LoadFromFile reads and parses a YAML config file, then applies
// environment variable overrides and validates the result.
func LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config %s: %w", path, err)
	}

	cfg.ApplyEnvOverrides()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// ApplyEnvOverrides applies P2P_* environment variables on top of the
// loaded config, per ADR-007's 12-factor override pattern.
//
// Supported overrides:
//   P2P_NETWORK_PORT
//   P2P_PERSISTENCE_ENGINE
//   P2P_PERSISTENCE_PATH
//   P2P_MONITORING_PROMETHEUS_PORT
//   P2P_CONSENSUS_TIMEOUT_SEC
//   P2P_CONSENSUS_BYZANTINE_THRESHOLD
func (c *Config) ApplyEnvOverrides() {
	if v := os.Getenv("P2P_NETWORK_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Network.Port = port
		}
	}
	if v := os.Getenv("P2P_PERSISTENCE_ENGINE"); v != "" {
		c.Persistence.Engine = v
	}
	if v := os.Getenv("P2P_PERSISTENCE_PATH"); v != "" {
		c.Persistence.Path = v
	}
	if v := os.Getenv("P2P_MONITORING_PROMETHEUS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Monitoring.PrometheusPort = port
		}
	}
	if v := os.Getenv("P2P_CONSENSUS_TIMEOUT_SEC"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil {
			c.Consensus.TimeoutSec = sec
		}
	}
	if v := os.Getenv("P2P_CONSENSUS_BYZANTINE_THRESHOLD"); v != "" {
		if threshold, err := strconv.ParseFloat(v, 64); err == nil {
			c.Consensus.ByzantineThreshold = threshold
		}
	}
}

// Validate checks the config against the rules documented in ADR-007:
// required fields, type ranges, and cross-field consistency.
func (c *Config) Validate() error {
	var errs []string

	// network.port: required, range [1024, 65535]
	if c.Network.Port < 1024 || c.Network.Port > 65535 {
		errs = append(errs, fmt.Sprintf("network.port must be in [1024, 65535], got %d", c.Network.Port))
	}

	// persistence.engine: required, one of known engines
	if c.Persistence.Engine != "inmemory" && c.Persistence.Engine != "rocksdb" {
		errs = append(errs, fmt.Sprintf(`persistence.engine must be "inmemory" or "rocksdb", got %q`, c.Persistence.Engine))
	}

	// persistence.path: required if engine=rocksdb
	if c.Persistence.Engine == "rocksdb" && c.Persistence.Path == "" {
		errs = append(errs, "persistence.path is required when persistence.engine=rocksdb")
	}

	// monitoring.prometheus_port: range check
	if c.Monitoring.PrometheusPort < 1024 || c.Monitoring.PrometheusPort > 65535 {
		errs = append(errs, fmt.Sprintf("monitoring.prometheus_port must be in [1024, 65535], got %d", c.Monitoring.PrometheusPort))
	}

	// network.port and monitoring.prometheus_port must not collide
	if c.Network.Port == c.Monitoring.PrometheusPort {
		errs = append(errs, "network.port and monitoring.prometheus_port must differ")
	}

	// consensus.timeout_sec: must be positive
	if c.Consensus.TimeoutSec <= 0 {
		errs = append(errs, fmt.Sprintf("consensus.timeout_sec must be positive, got %d", c.Consensus.TimeoutSec))
	}

	// consensus.byzantine_threshold: range [0, 1]
	if c.Consensus.ByzantineThreshold < 0 || c.Consensus.ByzantineThreshold > 1 {
		errs = append(errs, fmt.Sprintf("consensus.byzantine_threshold must be in [0, 1], got %f", c.Consensus.ByzantineThreshold))
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// ToYAML serializes the config back to YAML (used by `p2p config show`).
func (c *Config) ToYAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
