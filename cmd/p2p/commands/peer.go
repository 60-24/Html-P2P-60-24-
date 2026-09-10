package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/60-24/p2p-60-24/pkg/config"
)

func newPeerCommand() *cobra.Command {
	parent := &cobra.Command{
		Use:   "peer",
		Short: "Manage and inspect network peers",
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "Show configured peers",
		Example: `  p2p peer list --config config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadFromFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config from %s: %w", configPath, err)
			}

			if len(cfg.Network.Peers) == 0 {
				fmt.Println("No peers configured. Add peers under network.peers in your config file.")
				return nil
			}

			fmt.Println("Configured peers:")
			for _, peer := range cfg.Network.Peers {
				fmt.Printf("  - %s\n", peer)
			}
			return nil
		},
	}

	parent.AddCommand(list)
	return parent
}
