package cmd

import (
	"fmt"

	"github.com/plotly/claude-wrapper/pkg/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available environments and configurations",
	RunE:  listEnvironments,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listEnvironments(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Available Environments:")
	fmt.Println("=======================")

	if len(cfg.Environments) == 0 {
		fmt.Println("No environments configured")
		return nil
	}

	for name, env := range cfg.Environments {
		fmt.Printf("\n%s:\n", name)
		fmt.Printf("  Description: %s\n", env.Description)
		if env.ClaudePath != "" {
			fmt.Printf("  Claude Path: %s\n", env.ClaudePath)
		}
		if len(env.EnvVars) > 0 {
			fmt.Printf("  Environment Variables: %v\n", env.EnvVars)
		}
		if len(env.DefaultSpecs) > 0 {
			fmt.Printf("  Default Specs: %v\n", env.DefaultSpecs)
		}
	}

	return nil
}
