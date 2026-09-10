package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/60-24/p2p-60-24/pkg/monitoring"
)

func newMetricsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "metrics",
		Short: "Print Prometheus metrics (local snapshot)",
		Long: `Print Prometheus-format metrics for a local, ad-hoc snapshot.

For continuous scraping, point Prometheus at the /metrics HTTP endpoint
exposed by 'p2p run' instead (see config: monitoring.prometheus_port).`,
		Example: `  p2p metrics`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// A real deployment would attach to a running node's collector
			// via IPC; for the CLI-only snapshot we render an empty
			// collector to demonstrate the format (fresh counters).
			collector := monitoring.NewMetricsCollector()
			fmt.Print(collector.Render())
			return nil
		},
	}
}
