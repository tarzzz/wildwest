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
  claude-wrapper orchestrate --workspace .database

  # Then attach to monitor:
  tmux attach -t claude-orchestrator-*`,
	RunE: runOrchestrator,
}

func init() {
	rootCmd.AddCommand(orchestrateCmd)
	orchestrateCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", ".database", "workspace directory")
}

func runOrchestrator(cmd *cobra.Command, args []string) error {
	// Check if we're already inside a tmux session
	if os.Getenv("TMUX") != "" {
		// Already in tmux, run orchestrator directly
		fmt.Println("🎯 Starting Project Manager Orchestrator...")
		fmt.Println()

		orch, err := orchestrator.NewOrchestrator(workspaceDir, verbose)
		if err != nil {
			return fmt.Errorf("failed to create orchestrator: %w", err)
		}

		// Run orchestrator (blocks)
		return orch.Run()
	}

	// Not in tmux, spawn orchestrator in a new tmux session
	return spawnOrchestratorInTmux()
}

func spawnOrchestratorInTmux() error {
	// Pre-flight checks
	if err := checkClaudeAvailability(); err != nil {
		return err
	}

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

	// Create tmux session
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", tmuxSessionName, orchestratorCmd)
	if err := tmuxCmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	// Print success message with attach instructions
	fmt.Println("✅ Project Manager Orchestrator started in tmux")
	fmt.Println()
	fmt.Printf("📋 Session Name: %s\n", tmuxSessionName)
	fmt.Printf("📁 Workspace: %s\n", absWorkspace)
	fmt.Println()
	fmt.Println("To attach to the orchestrator:")
	fmt.Printf("  tmux attach -t %s\n", tmuxSessionName)
	fmt.Println()
	fmt.Println("To detach from the orchestrator:")
	fmt.Println("  Press: Ctrl+B, then D")
	fmt.Println()
	fmt.Println("To view all Claude sessions (including orchestrator):")
	fmt.Println("  tmux ls | grep claude")
	fmt.Println()
	fmt.Println("The orchestrator is now running in the background and will:")
	fmt.Println("  - Spawn Claude instances for each persona")
	fmt.Println("  - Monitor session health")
	fmt.Println("  - Handle spawn requests")
	fmt.Println("  - Archive completed sessions")

	return nil
}

// checkClaudeAvailability verifies Claude is installed and user is logged in
func checkClaudeAvailability() error {
	fmt.Println("🔍 Checking Claude availability...")

	// Get claude binary path (respects CLAUDE_BIN env var)
	claudeBin := os.Getenv("CLAUDE_BIN")
	if claudeBin == "" {
		claudeBin = "claude"
	}

	// Check if claude binary exists
	_, err := exec.LookPath(claudeBin)
	if err != nil {
		return fmt.Errorf("❌ Claude binary not found: %s\n\nPlease ensure Claude Code is installed and in your PATH.\nAlternatively, set CLAUDE_BIN environment variable to point to the claude binary.\n\nInstall Claude Code: https://claude.ai/code", claudeBin)
	}

	fmt.Printf("   ✓ Claude binary found: %s\n", claudeBin)

	// Check if user is logged in by running a simple prompt
	fmt.Println("   Checking authentication...")
	cmd := exec.Command(claudeBin, "-p", "return the word 'authenticated' and nothing else")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("❌ Failed to execute Claude command: %w\n\nPlease ensure you are logged in to Claude Code.\nRun: %s login", err, claudeBin)
	}

	// Check if output contains authentication error messages
	outputStr := string(output)
	if len(outputStr) == 0 {
		return fmt.Errorf("❌ Claude returned empty output. Please ensure you are logged in.\nRun: %s login", claudeBin)
	}

	fmt.Println("   ✓ Authentication verified")
	fmt.Println()

	return nil
}
