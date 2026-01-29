package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tarzzz/wildwest/pkg/session"
	"github.com/spf13/cobra"
)

var (
	sessionFilter string
	listOnly      bool
)

var attachCmd = &cobra.Command{
	Use:   "attach [session-id]",
	Short: "Attach to a running Claude instance",
	Long: `Attach to a running Claude Code session to interact with it directly.

If no session-id is provided, attaches to the Engineering Manager by default.
Use --list to see all available sessions.`,
	RunE: attachToSession,
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up stopped or dead sessions",
	Long: `Remove session directories for sessions that are no longer running.
This helps keep the workspace clean by archiving completed or stopped sessions.`,
	RunE: cleanupSessions,
}

func init() {
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(cleanupCmd)
	attachCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", ".ww-db", "workspace directory")
	attachCmd.Flags().BoolVarP(&listOnly, "list", "l", false, "list all running sessions")
	attachCmd.Flags().StringVarP(&sessionFilter, "filter", "f", "", "filter sessions by type (e.g., engineer, intern)")
	cleanupCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", ".ww-db", "workspace directory")
}

func attachToSession(cmd *cobra.Command, args []string) error {
	sm, err := session.NewSessionManager(workspaceDir)
	if err != nil {
		return err
	}

	sessions, err := sm.GetAllSessions()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	// List sessions
	if listOnly || len(args) == 0 {
		return listSessions(sessions, sessionFilter)
	}

	// Attach to specific session
	sessionID := args[0]
	return attachTo(sm, sessionID)
}

func cleanupSessions(cmd *cobra.Command, args []string) error {
	sm, err := session.NewSessionManager(workspaceDir)
	if err != nil {
		return err
	}

	sessions, err := sm.GetAllSessions()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions to clean up")
		return nil
	}

	fmt.Println("🧹 Cleaning up stopped sessions...")
	fmt.Println()

	cleaned := 0
	checked := 0
	for _, sess := range sessions {
		checked++
		// Check if tmux session is running
		isRunning := isTmuxSessionRunning(sess.ID)

		fmt.Printf("📋 Checking %s (%s): tmux=%v, status=%s\n", sess.PersonaName, sess.ID, isRunning, sess.Status)

		// Skip running sessions
		if isRunning {
			fmt.Printf("   → Skipping (still running)\n")
			continue
		}

		// Skip already archived
		if sess.Status == "archived" {
			fmt.Printf("   → Skipping (already archived)\n")
			continue
		}

		fmt.Printf("📦 Archiving: %s (%s)\n", sess.PersonaName, sess.ID)
		fmt.Printf("   Status: %s\n", sess.Status)

		// Update status to stopped
		sm.UpdateSessionStatus(sess.ID, "stopped")

		// Archive the directory
		oldPath := filepath.Join(sm.GetWorkspacePath(), sess.ID)
		newPath := filepath.Join(sm.GetWorkspacePath(), sess.ID+"-archived")

		if err := os.Rename(oldPath, newPath); err != nil {
			fmt.Printf("   ⚠️  Failed to archive: %v\n", err)
		} else {
			fmt.Printf("   ✅ Archived to: %s\n", newPath)
			cleaned++
		}

		fmt.Println()
	}

	if cleaned == 0 {
		fmt.Println("✨ No sessions need cleanup - all are either running or already archived")
	} else {
		fmt.Printf("✅ Cleaned up %d session(s)\n", cleaned)
	}

	return nil
}

func listSessions(sessions []*session.Session, filter string) error {
	fmt.Println("Available Sessions:")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()

	// Update session statuses based on tmux
	sm, _ := session.NewSessionManager(workspaceDir)
	for _, sess := range sessions {
		updateSessionStatus(sm, sess)
	}

	// Group by type
	byType := make(map[session.SessionType][]*session.Session)
	for _, sess := range sessions {
		if filter != "" && !matchesFilter(sess, filter) {
			continue
		}
		byType[sess.PersonaType] = append(byType[sess.PersonaType], sess)
	}

	// Display in hierarchy order
	displaySessionType(byType, session.SessionTypeEngineeringManager, "ENGINEERING MANAGER")
	displaySessionType(byType, session.SessionTypeSolutionsArchitect, "SOLUTIONS ARCHITECT")
	displaySessionType(byType, session.SessionTypeSoftwareEngineer, "SOFTWARE ENGINEERS")
	displaySessionType(byType, session.SessionTypeIntern, "INTERNS")

	fmt.Println("\nTo attach to a session:")
	fmt.Println("  wildwest attach <session-id>")
	fmt.Println("\nTo attach to manager (default):")
	fmt.Println("  wildwest attach")

	return nil
}

func displaySessionType(byType map[session.SessionType][]*session.Session, sessionType session.SessionType, title string) {
	sessions := byType[sessionType]
	if len(sessions) == 0 {
		return
	}

	fmt.Printf("╔══════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %s\n", title)
	fmt.Printf("╚══════════════════════════════════════════════════╝\n\n")

	for _, sess := range sessions {
		// Check actual tmux status
		isRunning := isTmuxSessionRunning(sess.ID)

		statusIcon := "🔄"
		statusText := sess.Status

		if sess.Status == "completed" {
			statusIcon = "✅"
		} else if sess.Status == "failed" {
			statusIcon = "❌"
		} else if sess.Status == "stopped" || !isRunning {
			statusIcon = "⏸️"
			statusText = "stopped"
		} else if isRunning {
			statusIcon = "🟢"
			statusText = "running"
		}

		fmt.Printf("%s %s\n", statusIcon, sess.PersonaName)
		fmt.Printf("   Session ID: %s\n", sess.ID)
		fmt.Printf("   Status: %s\n", statusText)
		fmt.Printf("   Started: %s\n", sess.StartTime.Format("2006-01-02 15:04:05"))

		if isRunning {
			tmuxSessionName := fmt.Sprintf("claude-%s", sess.ID)
			fmt.Printf("   Tmux: %s\n", tmuxSessionName)
		}

		if sess.PID > 0 {
			fmt.Printf("   PID: %d\n", sess.PID)
		}

		fmt.Println()
	}
}

func matchesFilter(sess *session.Session, filter string) bool {
	return sess.PersonaName == filter ||
		string(sess.PersonaType) == filter ||
		sess.ID == filter
}

func updateSessionStatus(sm *session.SessionManager, sess *session.Session) {
	// Skip if already marked as completed or archived
	if sess.Status == "completed" || sess.Status == "archived" {
		return
	}

	// Check if tmux session is running
	tmuxSessionName := fmt.Sprintf("claude-%s", sess.ID)
	checkCmd := exec.Command("tmux", "has-session", "-t", tmuxSessionName)
	err := checkCmd.Run()

	if err != nil {
		// Tmux session not running
		if sess.Status == "active" || sess.Status == "running" {
			// Was active but now stopped
			sm.UpdateSessionStatus(sess.ID, "stopped")
			sess.Status = "stopped"
		}
	} else {
		// Tmux session is running
		if sess.Status != "active" && sess.Status != "running" {
			// Update to active
			sm.UpdateSessionStatus(sess.ID, "active")
			sess.Status = "active"
		}
	}
}

func isTmuxSessionRunning(sessionID string) bool {
	tmuxSessionName := fmt.Sprintf("claude-%s", sessionID)
	checkCmd := exec.Command("tmux", "has-session", "-t", tmuxSessionName)
	return checkCmd.Run() == nil
}

func attachTo(sm *session.SessionManager, sessionID string) error {
	// If no session ID or "manager", find the manager
	if sessionID == "" || sessionID == "manager" {
		sessions, err := sm.GetAllSessions()
		if err != nil {
			return err
		}

		for _, sess := range sessions {
			if sess.PersonaType == session.SessionTypeEngineeringManager {
				sessionID = sess.ID
				break
			}
		}

		if sessionID == "" || sessionID == "manager" {
			return fmt.Errorf("no manager session found")
		}
	}

	// Get session info
	sessionDir := filepath.Join(sm.GetWorkspacePath(), sessionID)
	sessionFile := filepath.Join(sessionDir, "session.json")

	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Check if tmux session exists
	tmuxSessionName := fmt.Sprintf("claude-%s", sessionID)
	checkCmd := exec.Command("tmux", "has-session", "-t", tmuxSessionName)
	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf("tmux session %s not running. Start the orchestrator first.", tmuxSessionName)
	}

	fmt.Printf("🔗 Attaching to Claude session: %s\n", sessionID)
	fmt.Printf("   Tmux session: %s\n", tmuxSessionName)
	fmt.Printf("   Directory: %s\n\n", sessionDir)
	fmt.Println("Press Ctrl+B then D to detach from this session")
	fmt.Println()

	// Attach to tmux session
	attachCmd := exec.Command("tmux", "attach-session", "-t", tmuxSessionName)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr

	return attachCmd.Run()
}
