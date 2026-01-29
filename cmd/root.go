package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

// runDefaultCommand handles the case where wildwest is called with just a task string
// Example: wildwest "Build a REST API"
// This is equivalent to: wildwest team start "Build a REST API" --run --tui
func runDefaultCommand(cmd *cobra.Command, args []string) error {
	// If no args provided, show help
	if len(args) == 0 {
		return cmd.Help()
	}

	// Join args as task description
	task := strings.Join(args, " ")

	// Set up team start parameters
	workspaceDir = ".database"
	autoRun = true
	useTUITeam = true

	// Call team start logic directly with the task
	return startTeam(cmd, []string{task})
}

var rootCmd = &cobra.Command{
	Use:   "wildwest",
	Short: "A wrapper for Claude Code with custom environments and specs",
	Long: `Claude Wrapper - Multi-Agent Team Management for Claude Code

A hierarchical multi-agent system where Claude personas collaborate to complete projects.
Each persona runs independently in tmux sessions and communicates via file-based messaging.

QUICK START:

  1. Create a team and start orchestrator in one command:
     wildwest "Build a REST API for todo items"

  2. Or use the full command:
     wildwest team start "Build a REST API" --run --tui

  3. View all sessions:
     tmux ls | grep claude

  4. Attach to persona sessions:
     Press 'a' in TUI or use: wildwest attach <session-id>

  5. Detach from any tmux session:
     Press Ctrl+B then D

TEAM HIERARCHY:

  Leader Agent (Level 1)
    └─> Architecture Agent (Level 2)
          └─> Coding Agents (Level 3)
                └─> Support Agents (Level 4)

HOW IT WORKS:

  - Each persona runs in its own tmux session (claude-{session-id})
  - Personas monitor instructions.md every 5 seconds for new tasks
  - Communication happens via writing to each other's instructions.md
  - Task progress tracked in individual tasks.md files
  - Completed sessions are automatically archived

EXAMPLES:

  # Quick start (recommended)
  wildwest "Build a web scraper"

  # Full command with options
  wildwest team start "Build a web scraper" --run --tui

  # Start orchestrator separately
  wildwest orchestrate --workspace .database

For more information: https://github.com/tarzzz/wildwest`,
	Version: "0.1.0",
	Args:  cobra.ArbitraryArgs,
	RunE:  runDefaultCommand,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wildwest.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".wildwest")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
