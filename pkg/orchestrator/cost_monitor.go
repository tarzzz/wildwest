package orchestrator

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tarzzz/wildwest/pkg/session"
)

// CostMonitor handles periodic token usage polling and cost tracking
type CostMonitor struct {
	sm            *session.SessionManager
	pollInterval  time.Duration
	activeSessions map[string]bool
}

// NewCostMonitor creates a new cost monitor
func NewCostMonitor(sm *session.SessionManager) *CostMonitor {
	return &CostMonitor{
		sm:             sm,
		pollInterval:   60 * time.Second, // Poll every minute
		activeSessions: make(map[string]bool),
	}
}

// Start begins the cost monitoring loop
func (cm *CostMonitor) Start() {
	ticker := time.NewTicker(cm.pollInterval)
	defer ticker.Stop()

	fmt.Println("💰 Cost Monitor Started")
	fmt.Printf("   Polling interval: %v\n\n", cm.pollInterval)

	// Initial scan
	cm.pollAllSessions()

	for {
		select {
		case <-ticker.C:
			cm.pollAllSessions()
		}
	}
}

// pollAllSessions polls all active tmux sessions for token usage
func (cm *CostMonitor) pollAllSessions() {
	sessions, err := cm.sm.GetAllSessions()
	if err != nil {
		return
	}

	for _, sess := range sessions {
		// Only poll active sessions
		if sess.Status != "active" {
			continue
		}

		// Check if tmux session exists
		tmuxSessionName := fmt.Sprintf("claude-%s", sess.ID)
		if !cm.isTmuxSessionRunning(tmuxSessionName) {
			continue
		}

		// Capture tmux pane content
		output, err := cm.captureTmuxPane(tmuxSessionName)
		if err != nil {
			continue
		}

		// Parse token usage from output
		inputTokens, outputTokens, found := session.ParseTokensFromTmux(output)
		if found {
			// Update token usage
			if err := cm.sm.UpdateTokenUsage(sess.ID, inputTokens, outputTokens); err != nil {
				fmt.Printf("⚠️  Failed to update token usage for %s: %v\n", sess.ID, err)
			}
		}
	}
}

// isTmuxSessionRunning checks if a tmux session exists
func (cm *CostMonitor) isTmuxSessionRunning(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	err := cmd.Run()
	return err == nil
}

// captureTmuxPane captures the last 500 lines from a tmux pane
func (cm *CostMonitor) captureTmuxPane(sessionName string) (string, error) {
	// Capture last 500 lines from the pane
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p", "-S", "-500")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// GetCurrentCostSummary returns a formatted summary of current costs
func (cm *CostMonitor) GetCurrentCostSummary() (string, error) {
	totalCost, usageMap, err := cm.sm.GetTotalTeamCost()
	if err != nil {
		return "", err
	}

	sessions, err := cm.sm.GetAllSessions()
	if err != nil {
		return "", err
	}

	var summary strings.Builder
	summary.WriteString("💰 Team Cost Summary\n")
	summary.WriteString("====================\n\n")

	if len(usageMap) == 0 {
		summary.WriteString("No token usage data available yet.\n")
		summary.WriteString("The cost monitor will update every minute.\n")
		return summary.String(), nil
	}

	// Show per-session costs
	for _, sess := range sessions {
		usage, ok := usageMap[sess.ID]
		if !ok {
			continue
		}

		summary.WriteString(fmt.Sprintf("📊 %s (%s)\n", sess.PersonaName, sess.PersonaType))
		summary.WriteString(fmt.Sprintf("   Session: %s\n", sess.ID))
		summary.WriteString(fmt.Sprintf("   Model: %s\n", usage.Model))
		summary.WriteString(fmt.Sprintf("   Input Tokens: %s\n", session.FormatTokens(usage.InputTokens)))
		summary.WriteString(fmt.Sprintf("   Output Tokens: %s\n", session.FormatTokens(usage.OutputTokens)))
		summary.WriteString(fmt.Sprintf("   Total Tokens: %s\n", session.FormatTokens(usage.TotalTokens)))
		summary.WriteString(fmt.Sprintf("   Cost: %s\n", session.FormatCost(usage.EstimatedCost)))
		summary.WriteString(fmt.Sprintf("   Last Updated: %s\n", usage.LastUpdated.Format("2006-01-02 15:04:05")))
		summary.WriteString("\n")
	}

	summary.WriteString("====================\n")
	summary.WriteString(fmt.Sprintf("💵 Total Team Cost: %s\n", session.FormatCost(totalCost)))

	return summary.String(), nil
}
