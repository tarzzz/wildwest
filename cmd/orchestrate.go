package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/tarzzz/wildwest/pkg/orchestrator"
	"github.com/spf13/cobra"
)

var (
	useTUI bool
)

var orchestrateCmd = &cobra.Command{
	Use:   "orchestrate",
	Short: "Run the Project Manager orchestrator daemon",
	Long: `Starts the Project Manager orchestrator in a tmux session.

The orchestrator:
- Watches for spawn requests (*-request-* directories)
- Spawns Claude Code instances for requested personas
- Monitors running sessions
- Terminates completed sessions
- Archives finished work

The orchestrator runs in its own tmux session in the background.
You can attach to it at any time to monitor progress.

Example:
  wildwest orchestrate --workspace .ww-db
  wildwest orchestrate --workspace .ww-db --tui  # Interactive TUI

  # Then attach to monitor:
  tmux attach -t claude-orchestrator-*`,
	RunE: runOrchestrator,
}

func init() {
	rootCmd.AddCommand(orchestrateCmd)
	orchestrateCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", ".ww-db", "workspace directory")
	orchestrateCmd.Flags().BoolVar(&useTUI, "tui", true, "run orchestrator with interactive TUI (default)")
}

func runOrchestrator(cmd *cobra.Command, args []string) error {
	// Check if we're already inside a tmux session FIRST
	if os.Getenv("TMUX") != "" {
		// Already in tmux - run orchestrator with appropriate mode
		orch, err := orchestrator.NewOrchestrator(workspaceDir, verbose)
		if err != nil {
			return fmt.Errorf("failed to create orchestrator: %w", err)
		}

		// If TUI requested, run with TUI, otherwise run normal loop
		if useTUI {
			return orch.RunTUI()
		}
		return orch.Run()
	}

	// Not in tmux - if TUI mode requested, run directly without tmux
	if useTUI {
		// Minimal output - just start TUI
		orch, err := orchestrator.NewOrchestrator(workspaceDir, verbose)
		if err != nil {
			return fmt.Errorf("failed to create orchestrator: %w", err)
		}

		return orch.RunTUI()
	}

	// Not in tmux and not TUI, spawn orchestrator in a new tmux session
	return spawnOrchestratorInTmux()
}

func spawnOrchestratorInTmux() error {
	// Get absolute path to workspace
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute workspace path: %w", err)
	}

	// Create unique tmux session name with timestamp
	timestamp := time.Now().UnixMilli()
	tmuxSessionName := fmt.Sprintf("claude-orchestrator-%d", timestamp)

	// Get the path to the current executable
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build the command to run inside tmux
	orchestratorCmd := fmt.Sprintf("%s orchestrate --workspace %s", executable, absWorkspace)
	if verbose {
		orchestratorCmd += " --verbose"
	}
	if useTUI {
		orchestratorCmd += " --tui"
	}

	// Create tmux session
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", tmuxSessionName, orchestratorCmd)
	if err := tmuxCmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	// Brief success message
	fmt.Printf("✅ Orchestrator started: tmux attach -t %s\n", tmuxSessionName)

	return nil
}
