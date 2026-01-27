package cmd

import (
	"fmt"
	"time"

	"github.com/tarzzz/wildwest/pkg/orchestrator"
	"github.com/tarzzz/wildwest/pkg/session"
	"github.com/spf13/cobra"
)

var (
	costWatch bool
)

var teamCostCmd = &cobra.Command{
	Use:   "cost",
	Short: "Show token usage and estimated costs for the team",
	Long: `Display current token usage and estimated costs across all active personas.
Token usage is polled every minute from running Claude sessions.

Examples:
  # Show current cost snapshot
  wildwest team cost

  # Watch costs update in real-time
  wildwest team cost --watch`,
	RunE: teamCost,
}

func init() {
	teamCmd.AddCommand(teamCostCmd)
	teamCostCmd.Flags().BoolVarP(&costWatch, "watch", "w", false, "continuously watch and update costs every minute")
}

func teamCost(cmd *cobra.Command, args []string) error {
	sm, err := session.NewSessionManager(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	monitor := orchestrator.NewCostMonitor(sm)

	if costWatch {
		// Watch mode - update every minute
		fmt.Println("Starting cost monitor in watch mode...")
		fmt.Println("Press Ctrl+C to exit\n")

		// Show initial summary
		summary, err := monitor.GetCurrentCostSummary()
		if err != nil {
			return err
		}
		fmt.Println(summary)

		// Poll and update every minute
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Clear screen (works on Unix-like systems)
				fmt.Print("\033[H\033[2J")

				// Update and show new summary
				summary, err := monitor.GetCurrentCostSummary()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}
				fmt.Println(summary)
				fmt.Printf("Updated at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			}
		}
	} else {
		// One-time display
		summary, err := monitor.GetCurrentCostSummary()
		if err != nil {
			return err
		}

		fmt.Println(summary)

		// Show pricing reference
		fmt.Println("\n💡 Pricing Reference (per 1M tokens)")
		fmt.Println("=====================================")
		fmt.Println("Claude Sonnet: $3.00 input / $15.00 output")
		fmt.Println("Claude Opus:   $15.00 input / $75.00 output")
		fmt.Println("Claude Haiku:  $0.25 input / $1.25 output")
		fmt.Println("\nNote: Token usage is updated every minute by the orchestrator")
	}

	return nil
}
