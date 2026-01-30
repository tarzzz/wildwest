package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tarzzz/wildwest/pkg/orchestrator"
	"github.com/tarzzz/wildwest/pkg/session"
)

var (
	tuiWorkspace string
	baseWorkspace string
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the org chart TUI",
	Long:  `Launch an interactive TUI showing the team organization chart from the workspace directory.`,
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().StringVarP(&tuiWorkspace, "workspace", "w", "", "specific workspace/session directory to monitor")
	tuiCmd.Flags().StringVarP(&baseWorkspace, "base", "b", ".ww-db", "base workspace directory containing sessions")
}

func runTUI(cmd *cobra.Command, args []string) error {
	version := Version
	if GitCommit != "unknown" && GitCommit != "" {
		version = GitCommit[:7] // Show short commit hash
	}

	// If specific workspace provided, use it directly
	if tuiWorkspace != "" {
		return orchestrator.RunStaticTUIWithWorkspace(tuiWorkspace, version)
	}

	// Otherwise, list sessions and let user select
	sessions, err := session.ListSessions(baseWorkspace)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// No sessions found
	if len(sessions) == 0 {
		return fmt.Errorf("no sessions found in %s\nRun 'wildwest team start <task>' to create a new session", baseWorkspace)
	}

	// If only one session, load it directly
	if len(sessions) == 1 {
		fmt.Printf("Loading session: %s\n", sessions[0].Description)
		return orchestrator.RunStaticTUIWithWorkspace(sessions[0].WorkspacePath, version)
	}

	// Multiple sessions - show selector
	return orchestrator.RunSessionSelector(baseWorkspace, version)
}
