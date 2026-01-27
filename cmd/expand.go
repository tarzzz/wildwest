package cmd

import (
	"fmt"

	"github.com/plotly/claude-wrapper/pkg/claude"
	"github.com/plotly/claude-wrapper/pkg/config"
	"github.com/spf13/cobra"
)

var expandCmd = &cobra.Command{
	Use:   "expand [minimal-prompt]",
	Short: "Expand a minimal prompt into detailed instructions",
	Long: `Use Claude Code to expand a minimal prompt into detailed,
actionable instructions without executing them.`,
	Args: cobra.MinimumNArgs(1),
	RunE: expandPrompt,
}

func init() {
	rootCmd.AddCommand(expandCmd)

	expandCmd.Flags().StringVarP(&envName, "env", "e", "", "environment name from config")
	expandCmd.Flags().StringSliceVarP(&customSpecs, "spec", "s", []string{}, "custom specifications")
}

func expandPrompt(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	executor := claude.NewExecutor(cfg)

	opts := claude.ExecutorOptions{
		Prompt:       prompt,
		Environment:  envName,
		CustomSpecs:  customSpecs,
		ExpandPrompt: true,
		Verbose:      verbose,
	}

	return executor.Expand(opts)
}
