package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tarzzz/wildwest/pkg/orchestrator"
	"github.com/tarzzz/wildwest/pkg/persona"
	"github.com/tarzzz/wildwest/pkg/session"
	"github.com/spf13/cobra"
)

var (
	workspaceDir     string
	numEngineers     int
	numInterns       int
	teamTask         string
	autoRun          bool
	useTUITeam       bool
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage multi-persona team sessions",
	Long: `Coordinate multiple persona sessions working together on a task.
Sessions communicate through a shared workspace database.`,
}

var teamStartCmd = &cobra.Command{
	Use:   "start [task]",
	Short: "Start a team session with multiple personas",
	Args:  cobra.MinimumNArgs(1),
	RunE:  startTeam,
}

var teamStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of active team sessions",
	RunE:  teamStatus,
}

var teamStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all active team sessions",
	RunE:  stopTeam,
}

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(teamStartCmd)
	teamCmd.AddCommand(teamStatusCmd)
	teamCmd.AddCommand(teamStopCmd)

	teamStartCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", ".ww-db", "workspace directory for team collaboration")
	teamStartCmd.Flags().IntVar(&numEngineers, "engineers", 1, "number of software engineer sessions")
	teamStartCmd.Flags().IntVar(&numInterns, "interns", 0, "number of intern sessions")
	teamStartCmd.Flags().BoolVar(&autoRun, "run", false, "automatically start orchestration daemon after team creation")
	teamStartCmd.Flags().BoolVar(&useTUITeam, "tui", false, "use interactive TUI for orchestrator (requires --run)")
}

func startTeam(cmd *cobra.Command, args []string) error {
	task := strings.Join(args, " ")

	// Generate session ID and create session directory
	sessionID := session.GenerateSessionID()
	sessionPath := filepath.Join(workspaceDir, sessionID)

	// Create session directory
	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	// Save task description
	if err := session.SaveSessionDescription(sessionPath, task); err != nil {
		return fmt.Errorf("failed to save session description: %w", err)
	}

	// Save session metadata
	sessionMeta := session.SessionMetadata{
		ID:            sessionID,
		Description:   task,
		CreatedAt:     time.Now(),
		WorkspacePath: sessionPath,
	}
	metaData, _ := json.MarshalIndent(sessionMeta, "", "  ")
	if err := os.WriteFile(filepath.Join(sessionPath, "session.json"), metaData, 0644); err != nil {
		return fmt.Errorf("failed to save session metadata: %w", err)
	}

	// Create session manager with session-specific path
	sm, err := session.NewSessionManager(sessionPath)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	// Create workspace
	workspace, err := sm.CreateWorkspace(task)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	fmt.Printf("Created session: %s\n", sessionID)
	fmt.Printf("Session path: %s\n", sessionPath)
	fmt.Printf("Workspace ID: %s\n\n", workspace.ID)

	// Create initial team structure (Manager only)
	// All other resources will be requested dynamically by the manager

	// Create Engineering Manager directory
	fmt.Println("Creating Engineering Manager...")
	managerSession, err := sm.CreateSession(session.SessionTypeEngineeringManager, "", workspace.ID)
	if err != nil {
		return err
	}
	// Add initial task
	if err := sm.AddTask(managerSession.ID, task, "system"); err != nil {
		fmt.Printf("Warning: failed to add initial task: %v\n", err)
	}
	fmt.Printf("  Name: %s\n", managerSession.PersonaName)
	fmt.Printf("  Directory: %s\n\n", managerSession.ID)

	// Create orchestrator directory with initial state
	orchestratorDir := filepath.Join(sessionPath, "orchestrator")
	if err := os.MkdirAll(orchestratorDir, 0755); err != nil {
		return fmt.Errorf("failed to create orchestrator directory: %w", err)
	}

	// Create initial orchestrator state
	initialState := map[string]interface{}{
		"id":                    "orchestrator",
		"session_id":            sessionID,
		"session_path":          sessionPath,
		"status":                "initializing",
		"start_time":            time.Now(),
		"current_work":          "Waiting to start monitoring",
		"total_sessions_spawned": 0,
		"active_sessions":       0,
		"completed_sessions":    0,
		"failed_sessions":       0,
	}
	stateData, _ := json.MarshalIndent(initialState, "", "  ")
	stateFile := filepath.Join(orchestratorDir, "state.json")
	if err := os.WriteFile(stateFile, stateData, 0644); err != nil {
		return fmt.Errorf("failed to create orchestrator state: %w", err)
	}

	fmt.Println("✅ Engineering Manager created successfully!")
	fmt.Printf("📁 Workspace: %s\n\n", sm.GetWorkspacePath())
	fmt.Println("ℹ️  The Engineering Manager will assess the task and request needed resources")
	fmt.Println("   (Solutions Architect, Software Engineers, QA, Interns) dynamically.\n")

	if autoRun {
		// Spawn orchestrator in tmux session
		fmt.Println("🚀 Starting orchestration daemon...")

		// Create tmux session for orchestrator
		sessionName := fmt.Sprintf("wildwest-orchestrator-%d", time.Now().Unix())

		// Build command: wildwest orchestrate --workspace <workspace> --no-tui
		// (runs orchestrator loop, not TUI)
		orchestrateCmd := fmt.Sprintf("wildwest orchestrate --workspace %s --tui=false", sessionPath)

		// Start tmux session with orchestrator
		tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, orchestrateCmd)
		if err := tmuxCmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to start orchestrator automatically: %v\n", err)
			fmt.Printf("Please run manually: wildwest orchestrate --workspace %s\n", workspaceDir)
			return nil
		}

		fmt.Printf("✅ Orchestrator started: tmux attach -t %s\n", sessionName)

		// Give orchestrator time to initialize and save state
		time.Sleep(1 * time.Second)

		// Open TUI to monitor progress
		version := Version
		if GitCommit != "unknown" && GitCommit != "" {
			version = GitCommit[:7]
		}
		return orchestrator.RunStaticTUIWithWorkspace(sessionPath, version)
	} else {
		fmt.Println("⚠️  IMPORTANT: Start the orchestrator to spawn Claude instances:")
		fmt.Printf("   wildwest orchestrate --workspace %s\n\n", sessionPath)

		fmt.Println("The orchestrator will:")
		fmt.Println("  1. Spawn the Engineering Manager")
		fmt.Println("  2. Manager will assess and request needed resources")
		fmt.Println("  3. Orchestrator spawns requested team members")
		fmt.Println("  4. Manage team lifecycle and terminate completed sessions")
	}

	return nil
}

func startPersonaSession(sm *session.SessionManager, personas *persona.PersonaConfig, personaType session.SessionType, name string, workspaceID string, task string) (*session.Session, error) {
	// Create session record
	sess, err := sm.CreateSession(personaType, name, workspaceID)
	if err != nil {
		return nil, err
	}

	// Get persona
	p, err := personas.GetPersona(string(personaType))
	if err != nil {
		return nil, err
	}

	// Create instructions with workspace context
	instructions := fmt.Sprintf(`%s

## Workspace Information
Session ID: %s
Workspace Root: %s
Your Directory: %s/%s/
Your Role: %s

## Directory Structure
Each persona has their own directory:
- %s/%s/               (your directory)
  - tasks.md            (your task list - YOU maintain this)
  - instructions.md     (instructions from other personas)
  - tracker.json        (tracks what you've read - DO NOT modify manually)
  - *.md, *.go, etc.    (your output files)

- %s/shared/           (shared resources accessible to all)

## Collaboration Guidelines

### Your Tasks (tasks.md)
- Regularly check your tasks.md file for assigned tasks
- Update task statuses as you work:
  - "not started" - Task assigned but not started
  - "in progress" - Currently working on it
  - "completed" - Task is done
- Only YOU can modify YOUR tasks.md

### Reading Instructions (instructions.md)
- Check your instructions.md REGULARLY (every few minutes)
- This file contains timestamped instructions from other team members
- The tracker.json file helps you resume if you go offline
- Each instruction section is timestamped - look for new sections

### Assigning Work to Others
- To assign work to another persona, write to their instructions.md
- Example: Write to .ww-db/<other-session-id>/instructions.md
- Always include a timestamp header like: "## Instructions from %s (YYYY-MM-DD HH:MM:SS)"
- Be clear and specific in your instructions

### Reading Others' Work
- You can read any other persona's directory to see their progress
- Check their tasks.md to see what they're working on
- Read their output files to review their work

### Shared Resources
- Use .ww-db/shared/ for files that everyone needs access to
- Examples: architecture docs, shared configs, common utilities

## Communication Protocol
1. Check your instructions.md regularly for new assignments
2. Update your tasks.md status as you progress
3. Write your deliverables as files in your directory
4. Assign work to others by writing to their instructions.md with timestamps

## Recovery and State Management
- tracker.json maintains your reading state
- If you restart or reconnect, check tracker.json to see what you've already read
- Always read from your last known position to avoid missing updates

## Your Current Task
%s

`, p.Instructions, sess.ID, sm.GetWorkspacePath(), sm.GetWorkspacePath(), sess.ID, name,
	sm.GetWorkspacePath(), sess.ID, sm.GetWorkspacePath(), sess.ID, task)

	// Create a temporary instructions file
	instructionsFile := fmt.Sprintf("%s/workspace-instructions-%s.md", sm.GetWorkspacePath(), sess.ID)
	if err := os.WriteFile(instructionsFile, []byte(instructions), 0644); err != nil {
		return nil, err
	}

	// Add initial task to tasks.md
	if err := sm.AddTask(sess.ID, task, "system"); err != nil {
		fmt.Printf("Warning: failed to add initial task: %v\n", err)
	}

	// Start Claude in background with persona instructions
	go func() {
		// Get claude binary path (respects CLAUDE_BIN env var)
		claudeBin := os.Getenv("CLAUDE_BIN")
		if claudeBin == "" {
			claudeBin = "claude"
		}

		cmd := exec.Command(claudeBin, "--instructions", instructionsFile, task)
		cmd.Dir = sm.GetWorkspacePath()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			fmt.Printf("Session %s failed: %v\n", sess.ID, err)
			sm.UpdateSessionStatus(sess.ID, "failed")
		} else {
			sm.UpdateSessionStatus(sess.ID, "completed")
		}
	}()

	return sess, nil
}

func teamStatus(cmd *cobra.Command, args []string) error {
	sm, err := session.NewSessionManager(workspaceDir)
	if err != nil {
		return err
	}

	sessions, err := sm.GetActiveSessions()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No active team sessions")
		return nil
	}

	fmt.Println("Active Team Sessions:")
	fmt.Println("====================")
	for _, sess := range sessions {
		// Get current work
		currentWork := sm.GetCurrentWork(sess.ID)

		fmt.Printf("\n%s (%s)\n", sess.PersonaName, sess.PersonaType)
		fmt.Printf("  Session ID: %s\n", sess.ID)
		fmt.Printf("  Status: %s\n", sess.Status)
		fmt.Printf("  Started: %s\n", sess.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Current Work: %s\n", currentWork)
	}

	return nil
}

func stopTeam(cmd *cobra.Command, args []string) error {
	sm, err := session.NewSessionManager(workspaceDir)
	if err != nil {
		return err
	}

	sessions, err := sm.GetActiveSessions()
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		sm.UpdateSessionStatus(sess.ID, "stopped")
		fmt.Printf("Stopped session: %s (%s)\n", sess.PersonaName, sess.ID)
	}

	return nil
}
