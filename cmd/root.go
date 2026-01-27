package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "wildwest",
	Short: "A wrapper for Claude Code with custom environments and specs",
	Long: `Claude Wrapper - Multi-Agent Team Management for Claude Code

A hierarchical multi-agent system where Claude personas collaborate to complete projects.
Each persona runs independently in tmux sessions and communicates via file-based messaging.

QUICK START:

  1. Create a team:
     wildwest team start "Build a REST API for todo items"

  2. Start the orchestrator (runs in tmux, returns immediately):
     wildwest orchestrate --workspace .database

  3. View all sessions (including orchestrator):
     tmux ls | grep claude

  4. Attach to orchestrator to monitor:
     tmux attach -t claude-orchestrator-*

  5. Attach to persona sessions:
     wildwest attach                 # Manager (default)
     wildwest attach <session-id>    # Specific persona

  6. Detach from any tmux session:
     Press Ctrl+B then D

  7. Clean up stopped sessions:
     wildwest cleanup

TEAM HIERARCHY:

  Engineering Manager (Level 1)
    └─> Solutions Architect (Level 2)
          └─> Software Engineers (Level 3)
                └─> Interns (Level 4)

HOW IT WORKS:

  - Each persona runs in its own tmux session (claude-{session-id})
  - Personas monitor instructions.md every 5 seconds for new tasks
  - Communication happens via writing to each other's instructions.md
  - Task progress tracked in individual tasks.md files
  - Completed sessions are automatically archived

EXAMPLES:

  # Create team with 2 engineers
  wildwest team start "Build a web scraper" --engineers 2

  # Attach to specific session
  wildwest attach engineering-manager-1234567890

  # Filter sessions by type
  wildwest attach --list --filter engineer

  # View orchestrator status
  tmux ls | grep claude

For more information: https://github.com/tarzzz/wildwest`,
	Version: "0.1.0",
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
