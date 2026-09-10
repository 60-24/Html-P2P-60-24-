package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/60-24/p2p-60-24/pkg/config"
)

func newConfigCommand() *cobra.Command {
	parent := &cobra.Command{
		Use:   "config",
		Short: "Inspect node configuration",
	}

	show := &cobra.Command{
		Use:   "show",
		Short: "Display the current (validated) configuration",
		Example: `  p2p config show --config config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadFromFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config from %s: %w", configPath, err)
			}

			yamlStr, err := cfg.ToYAML()
			if err != nil {
				return fmt.Errorf("failed to render config as YAML: %w", err)
			}

			fmt.Print(yamlStr)
			return nil
		},
	}

	parent.AddCommand(show)
	return parent
}
