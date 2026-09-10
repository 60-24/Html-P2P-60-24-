package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Example: `  p2p version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("p2p 60-24 OneClick Evo — %s\n", Version)
			fmt.Println("6/6 TERRA OS engines | Persistence + Observability + Config layers")
			return nil
		},
	}
}
