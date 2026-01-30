package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tarzzz/wildwest/pkg/orchestrator"
)

var (
	tuiWorkspace string
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the org chart TUI",
	Long:  `Launch an interactive TUI showing the team organization chart from the workspace directory.`,
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().StringVarP(&tuiWorkspace, "workspace", "w", ".ww-db", "workspace directory to monitor")
}

func runTUI(cmd *cobra.Command, args []string) error {
	version := Version
	if GitCommit != "unknown" && GitCommit != "" {
		version = GitCommit[:7] // Show short commit hash
	}
	return orchestrator.RunStaticTUIWithWorkspace(tuiWorkspace, version)
}
