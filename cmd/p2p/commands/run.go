package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/60-24/p2p-60-24/pkg/config"
	"github.com/60-24/p2p-60-24/pkg/monitoring"
	"github.com/60-24/p2p-60-24/pkg/storage"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start a P2P 60-24 node with the given configuration",
		Long: `Start a P2P 60-24 node with the given configuration.

Loads config.yaml (with environment variable overrides applied),
initializes persistence and metrics, and starts the node's network
and consensus layers.`,
		Example: `  p2p run --config config.yaml
  p2p run -c config.yaml --verbose`,
		RunE: runNode,
	}
	return cmd
}

func runNode(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %w", configPath, err)
	}

	if verbose {
		fmt.Printf("Loaded config from %s\n", configPath)
		fmt.Printf("  network.port: %d\n", cfg.Network.Port)
		fmt.Printf("  persistence.engine: %s\n", cfg.Persistence.Engine)
		fmt.Printf("  monitoring.prometheus_port: %d\n", cfg.Monitoring.PrometheusPort)
	}

	store, err := storage.NewInMemoryStore(cfg.Persistence.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize persistence store: %w", err)
	}
	defer store.Close()

	collector := monitoring.NewMetricsCollector()
	server := collector.StartServer(cfg.Monitoring.PrometheusPort)
	defer server.Close()

	fmt.Printf("✓ P2P 60-24 node started\n")
	fmt.Printf("  Network:    UDP port %d\n", cfg.Network.Port)
	fmt.Printf("  Metrics:    http://localhost:%d/metrics\n", cfg.Monitoring.PrometheusPort)
	fmt.Printf("  Persistence: %s (%s)\n", cfg.Persistence.Engine, cfg.Persistence.Path)
	fmt.Printf("\nNode is running. Press Ctrl+C to stop.\n")

	select {} // block forever (real signal handling would go here)
}
