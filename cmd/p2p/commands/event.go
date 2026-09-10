package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/60-24/p2p-60-24/pkg/storage"
)

func newEventCommand() *cobra.Command {
	parent := &cobra.Command{
		Use:   "event",
		Short: "Inspect and export the event log",
	}

	var outputPath string
	export := &cobra.Command{
		Use:   "export",
		Short: "Export the event log to a file",
		Long: `Export the persistence store's metadata (event count, snapshot
count, checkpoint count) as JSON to a file.`,
		Example: `  p2p event export --output events.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storage.NewInMemoryStore("")
			if err != nil {
				return fmt.Errorf("failed to open persistence store: %w", err)
			}
			defer store.Close()

			data, err := store.Export()
			if err != nil {
				return fmt.Errorf("failed to export event log: %w", err)
			}

			if outputPath == "" {
				fmt.Println(string(data))
				return nil
			}

			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write export to %s: %w", outputPath, err)
			}

			fmt.Printf("✓ Exported event log metadata to %s\n", outputPath)
			return nil
		},
	}
	export.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: stdout)")

	parent.AddCommand(export)
	return parent
}
