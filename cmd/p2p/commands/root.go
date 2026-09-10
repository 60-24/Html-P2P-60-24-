// Package commands implements the P2P 60-24 CLI per ADR-008.
// It follows the Cobra subcommand pattern (kubectl/docker/gh style),
// with human-friendly error and help text (not unix minimalism).
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags, defaults to "dev".
var Version = "dev"

// configPath is the shared --config flag value across subcommands.
var configPath string

// verbose is the shared --verbose flag value.
var verbose bool

// NewRootCommand builds the root `p2p` command and wires in all subcommands.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "p2p",
		Short: "P2P 60-24 OneClick Evo — decentralized trust network node",
		Long: `p2p is the command-line interface for running and operating a
P2P 60-24 OneClick Evo node.

It manages the full node lifecycle: configuration loading, persistence,
network gossip, Byzantine consensus, and Prometheus-compatible metrics.

See 'p2p help <command>' for details on a specific command.`,
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVarP(&configPath, "config", "c", "config/default.yaml", "Path to config.yaml")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose logging")

	root.AddCommand(
		newRunCommand(),
		newConfigCommand(),
		newMetricsCommand(),
		newPeerCommand(),
		newEventCommand(),
		newStatusCommand(),
		newVersionCommand(),
	)

	return root
}

// Execute runs the root command and handles top-level errors with a
// human-friendly message (per ADR-008: no unix minimalism).
func Execute() {
	root := NewRootCommand()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\nRun 'p2p --help' for usage.\n", err)
		os.Exit(1)
	}
}
