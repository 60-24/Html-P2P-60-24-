package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/60-24/p2p-60-24/pkg/config"
)

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Health check (exit code 0=healthy, 1=unhealthy)",
		Long: `Validate that the node's configuration is loadable and correct.

Exit code 0 means the config is healthy; exit code 1 means it is not.
Designed for use in scripts and container healthchecks:

  p2p status && echo "Node is healthy" || exit 1`,
		Example: `  p2p status --config config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadFromFile(configPath)
			if err != nil {
				// Return the error rather than calling os.Exit directly:
				// (1) root.Execute() already exits 1 on any command error,
				//     so this achieves the documented exit-code contract
				//     without duplicating exit logic here, and
				// (2) it keeps this command testable in-process — a direct
				//     os.Exit() call here would kill the test binary itself
				//     rather than just failing the test.
				return fmt.Errorf("✗ UNHEALTHY: %w", err)
			}

			fmt.Printf("✓ HEALTHY\n")
			fmt.Printf("  Config:     %s (valid)\n", configPath)
			fmt.Printf("  Network:    port %d\n", cfg.Network.Port)
			fmt.Printf("  Monitoring: port %d\n", cfg.Monitoring.PrometheusPort)
			return nil
		},
	}
}
