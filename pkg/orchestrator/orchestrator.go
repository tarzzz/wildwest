package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tarzzz/wildwest/pkg/persona"
	"github.com/tarzzz/wildwest/pkg/session"
)

// Orchestrator manages the lifecycle of Claude instances
type Orchestrator struct {
	sm              *session.SessionManager
	personas        *persona.PersonaConfig
	activeSessions  map[string]bool // sessionID -> active status
	workspacePath   string
	pollInterval    time.Duration
	verbose         bool
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(workspacePath string, verbose bool) (*Orchestrator, error) {
	sm, err := session.NewSessionManager(workspacePath)
	if err != nil {
		return nil, err
	}

	personas, err := persona.LoadPersonas("")
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		sm:             sm,
		personas:       personas,
		activeSessions: make(map[string]bool),
		workspacePath:  workspacePath,
		pollInterval:   5 * time.Second,
		verbose:        verbose,
	}, nil
}

// Run starts the orchestrator daemon
func (o *Orchestrator) Run() error {
	fmt.Println("🎯 Project Manager Orchestrator Started")
	fmt.Printf("   Workspace: %s\n", o.workspacePath)
	fmt.Printf("   Poll Interval: %v\n", o.pollInterval)
	fmt.Println()

	// Start cost monitor in background
	costMonitor := NewCostMonitor(o.sm)
	go func() {
		costMonitor.Start()
	}()

	ticker := time.NewTicker(o.pollInterval)
	defer ticker.Stop()

	// Initial scan
	if err := o.scanAndProcess(); err != nil {
		fmt.Printf("⚠️  Error in initial scan: %v\n", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := o.scanAndProcess(); err != nil {
				fmt.Printf("⚠️  Error in scan: %v\n", err)
			}
		}
	}
}

// RunTUI starts the orchestrator with interactive TUI
func (o *Orchestrator) RunTUI() error {
	// Start cost monitor in background
	costMonitor := NewCostMonitor(o.sm)
	go func() {
		costMonitor.Start()
	}()

	// Start background goroutine to handle orchestrator logic
	go func() {
		ticker := time.NewTicker(o.pollInterval)
		defer ticker.Stop()

		// Initial scan
		if err := o.scanAndProcess(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Error in initial scan: %v\n", err)
		}

		for {
			select {
			case <-ticker.C:
				if err := o.scanAndProcess(); err != nil {
					fmt.Fprintf(os.Stderr, "⚠️  Error in scan: %v\n", err)
				}
			}
		}
	}()

	// Start TUI
	model := NewOrchestratorModel(o.sm, o.workspacePath)
	model.refreshSessions()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

// scanAndProcess scans for requests and manages sessions
func (o *Orchestrator) scanAndProcess() error {
	// 1. Check for new spawn requests
	if err := o.processSpawnRequests(); err != nil {
		return err
	}

	// 2. Check for completed sessions
	if err := o.processCompletedSessions(); err != nil {
		return err
	}

	// 3. Monitor running sessions
	if err := o.monitorRunningSessions(); err != nil {
		return err
	}

	return nil
}

// processSpawnRequests looks for *-request-* directories and spawns Claude instances
func (o *Orchestrator) processSpawnRequests() error {
	entries, err := os.ReadDir(o.workspacePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "shared" {
			continue
		}

		dirName := entry.Name()

		// Check if it's a request directory
		if strings.Contains(dirName, "-request-") {
			if err := o.handleSpawnRequest(dirName); err != nil {
				fmt.Printf("⚠️  Failed to handle spawn request %s: %v\n", dirName, err)
			}
			continue
		}

		// Skip archived or completed directories
		if strings.HasSuffix(dirName, "-archived") || strings.HasSuffix(dirName, "-completed") {
			continue
		}

		// Check for initial sessions that need spawning (not yet running)
		if strings.HasPrefix(dirName, "engineering-manager-") ||
			strings.HasPrefix(dirName, "solutions-architect-") ||
			strings.HasPrefix(dirName, "software-engineer-") ||
			strings.HasPrefix(dirName, "intern-") {

			// Skip if already running
			if o.activeSessions[dirName] {
				continue
			}

			// Check if session exists and is active
			sessionFile := filepath.Join(o.workspacePath, dirName, "session.json")
			if _, err := os.Stat(sessionFile); err == nil {
				// Session exists, spawn it
				if err := o.handleSpawnRequest(dirName); err != nil {
					fmt.Printf("⚠️  Failed to spawn session %s: %v\n", dirName, err)
				}
			}
		}
	}

	return nil
}

// handleSpawnRequest processes a spawn request
func (o *Orchestrator) handleSpawnRequest(dirName string) error {
	requestPath := filepath.Join(o.workspacePath, dirName)

	// Determine persona type
	var personaType session.SessionType
	var isInitialSpawn bool

	if strings.HasPrefix(dirName, "software-engineer-request-") {
		personaType = session.SessionTypeSoftwareEngineer
	} else if strings.HasPrefix(dirName, "intern-request-") {
		personaType = session.SessionTypeIntern
	} else if strings.HasPrefix(dirName, "engineering-manager-") {
		// Initial manager spawn
		personaType = session.SessionTypeEngineeringManager
		isInitialSpawn = true
	} else if strings.HasPrefix(dirName, "solutions-architect-") {
		// Initial architect spawn
		personaType = session.SessionTypeSolutionsArchitect
		isInitialSpawn = true
	} else if strings.HasPrefix(dirName, "software-engineer-") {
		// Initial engineer spawn
		personaType = session.SessionTypeSoftwareEngineer
		isInitialSpawn = true
	} else if strings.HasPrefix(dirName, "intern-") {
		// Initial intern spawn
		personaType = session.SessionTypeIntern
		isInitialSpawn = true
	} else {
		return fmt.Errorf("unknown request type: %s", dirName)
	}

	// Skip if already spawned
	if o.activeSessions[dirName] {
		return nil
	}

	var sess *session.Session
	var err error

	if isInitialSpawn {
		// For initial spawns, the session already exists
		sessions, _ := o.sm.GetAllSessions()
		for _, s := range sessions {
			if s.ID == dirName {
				sess = s
				break
			}
		}
		if sess == nil {
			return fmt.Errorf("session not found: %s", dirName)
		}
	} else {
		// Create new session for request (name will be auto-generated)
		sess, err = o.sm.CreateSession(personaType, "", "main")
		if err != nil {
			return err
		}

		// Move/copy instructions from request directory to new session
		requestInstructions := filepath.Join(requestPath, "instructions.md")
		if data, err := os.ReadFile(requestInstructions); err == nil {
			sessionInstructions := filepath.Join(o.workspacePath, sess.ID, "instructions.md")
			if err := os.WriteFile(sessionInstructions, data, 0644); err != nil {
				fmt.Printf("⚠️  Failed to copy instructions: %v\n", err)
			}
		}

		// Remove request directory
		if err := os.RemoveAll(requestPath); err != nil {
			fmt.Printf("⚠️  Failed to remove request directory: %v\n", err)
		}
	}

	fmt.Printf("\n🚀 Spawning %s: %s\n", personaType, sess.PersonaName)

	// Get persona definition
	p, err := o.personas.GetPersona(string(personaType))
	if err != nil {
		return err
	}

	// Create enhanced instructions
	instructions := o.generateInstructions(p, sess)

	// Write instructions to a temporary file for Claude to read
	instructionsFile := filepath.Join(o.workspacePath, sess.ID, "persona-instructions.md")
	if err := os.WriteFile(instructionsFile, []byte(instructions), 0644); err != nil {
		return fmt.Errorf("failed to write instructions: %w", err)
	}

	// Create tmux session name (sanitized)
	tmuxSessionName := fmt.Sprintf("claude-%s", sess.ID)

	// Get absolute paths for persona files
	absWorkspace, _ := filepath.Abs(o.workspacePath)
	absSessionDir := filepath.Join(absWorkspace, sess.ID)

	// Get claude binary path (respects CLAUDE_BIN env var)
	claudeBin := os.Getenv("CLAUDE_BIN")
	if claudeBin == "" {
		claudeBin = "claude"
	}

	// Create the initial Claude command with instructions to monitor files
	// Run from project root, reference persona files with full paths
	claudeCmd := fmt.Sprintf("%s --dangerously-skip-permissions --append-system-prompt \"$(cat %s/persona-instructions.md)\" 'IMPORTANT: You are working from the project root directory. Your persona files are located in: %s/\n\nCreate a background Bash task that monitors %s/instructions.md every 5 seconds. When new content is detected (file size increases), read and act on the new instructions immediately. Start this monitoring task now using:\n\nBash(PERSONA_DIR=%s; LAST_SIZE=0; while true; do if [ -f \"$PERSONA_DIR/instructions.md\" ]; then NEW_SIZE=$(wc -c < \"$PERSONA_DIR/instructions.md\" | tr -d \" \"); if [ \"$NEW_SIZE\" -gt \"${LAST_SIZE:-0}\" 2>/dev/null ]; then echo \"New instructions detected in $PERSONA_DIR/instructions.md\"; fi; LAST_SIZE=$NEW_SIZE; fi; sleep 5; done, run_in_background=true)\n\nThen begin working on your tasks from %s/tasks.md. All your work should be done in the current directory (project root), but reference your persona directory for instructions and tasks.'",
		claudeBin, absSessionDir, absSessionDir, absSessionDir, absSessionDir, absSessionDir)

	// Create tmux session and run Claude from current directory
	cmd := exec.Command("tmux", "new-session", "-d", "-s", tmuxSessionName, "bash", "-c", claudeCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start tmux session: %w (output: %s)", err, string(output))
	}

	// Mark session as active
	o.activeSessions[sess.ID] = true

	fmt.Printf("   ✅ Session: %s (tmux: %s)\n", sess.ID, tmuxSessionName)
	fmt.Printf("   📎 Attach with: tmux attach -t %s\n", tmuxSessionName)

	return nil
}

// isTmuxSessionRunning checks if a tmux session exists
func (o *Orchestrator) isTmuxSessionRunning(sessionID string) bool {
	tmuxSessionName := fmt.Sprintf("claude-%s", sessionID)
	cmd := exec.Command("tmux", "has-session", "-t", tmuxSessionName)
	err := cmd.Run()
	return err == nil
}

// processCompletedSessions checks for completed sessions and cleans up
func (o *Orchestrator) processCompletedSessions() error {
	sessions, err := o.sm.GetAllSessions()
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		// Skip if still running
		if o.activeSessions[sess.ID] {
			continue
		}

		// Skip if already marked completed
		if sess.Status == "completed" || sess.Status == "archived" {
			continue
		}

		// Check if all tasks are completed
		tasks, err := o.sm.ReadTasks(sess.ID)
		if err != nil {
			continue
		}

		if o.areAllTasksCompleted(tasks) {
			fmt.Printf("\n🎉 All tasks completed for %s (%s)\n", sess.PersonaName, sess.ID)

			// Terminate tmux session if still running
			if o.isTmuxSessionRunning(sess.ID) {
				tmuxSessionName := fmt.Sprintf("claude-%s", sess.ID)
				exec.Command("tmux", "kill-session", "-t", tmuxSessionName).Run()
				delete(o.activeSessions, sess.ID)
			}

			// Mark as completed
			o.sm.UpdateSessionStatus(sess.ID, "completed")

			// Archive the directory
			o.archiveSession(sess.ID)
		}
	}

	return nil
}

// areAllTasksCompleted checks if all tasks in tasks.md are completed
func (o *Orchestrator) areAllTasksCompleted(tasks string) bool {
	if !strings.Contains(tasks, "## Task:") {
		return false // No tasks defined
	}

	lines := strings.Split(tasks, "\n")
	hasIncompleteTasks := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "**Status**:") {
			status := strings.TrimSpace(strings.Split(line, ":")[1])
			if status != "completed" {
				hasIncompleteTasks = true
				break
			}
		}
	}

	return !hasIncompleteTasks
}

// archiveSession archives a completed session
func (o *Orchestrator) archiveSession(sessionID string) error {
	oldPath := filepath.Join(o.workspacePath, sessionID)
	newPath := filepath.Join(o.workspacePath, sessionID+"-completed")

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	fmt.Printf("   📦 Archived to: %s\n", newPath)
	return nil
}

// monitorRunningSessions checks health of running sessions
func (o *Orchestrator) monitorRunningSessions() error {
	// Check if tmux sessions are still alive
	for sessionID := range o.activeSessions {
		if !o.isTmuxSessionRunning(sessionID) {
			// Get session info to show which one stopped
			sessions, _ := o.sm.GetAllSessions()
			var personaName string
			for _, s := range sessions {
				if s.ID == sessionID {
					personaName = s.PersonaName
					break
				}
			}

			if personaName != "" {
				fmt.Printf("\n⚠️  Session stopped: %s (%s)\n", personaName, sessionID)
			} else {
				fmt.Printf("\n⚠️  Session stopped: %s\n", sessionID)
			}

			delete(o.activeSessions, sessionID)
			o.sm.UpdateSessionStatus(sessionID, "stopped")

			// Check if it was manually killed vs completed
			tasks, err := o.sm.ReadTasks(sessionID)
			if err == nil && o.areAllTasksCompleted(tasks) {
				fmt.Printf("   📋 All tasks were completed\n")
				o.sm.UpdateSessionStatus(sessionID, "completed")
			} else {
				fmt.Printf("   📋 Session did not complete all tasks\n")
			}
		}
	}

	return nil
}

// createWrapperScript creates a shell script that polls instructions.md
func (o *Orchestrator) createWrapperScript(sessionID, sessionDir string) string {
	// Get absolute path
	absSessionDir, _ := filepath.Abs(sessionDir)
	script := fmt.Sprintf(`#!/bin/bash
set -e

SESSION_DIR="%s"
cd "$SESSION_DIR"

echo "🤖 Starting Claude worker for session: %s"
echo "📂 Working directory: $SESSION_DIR"
echo "⏰ Polling instructions.md every 5 seconds"
echo ""

# Function to get file size (cross-platform)
get_file_size() {
    if [ -f "$1" ]; then
        wc -c < "$1" | tr -d ' '
    else
        echo "0"
    fi
}

LAST_INSTRUCTIONS_SIZE=0
LAST_TASKS_SIZE=0
ITERATION=0

# Initial run
echo "🎬 Initial run - reading tasks"
claude --print --dangerously-skip-permissions \
    --append-system-prompt "$(cat persona-instructions.md)" \
    "Read your tasks.md file. If you have tasks, start working on them. If waiting for instructions, check instructions.md file."

while true; do
    ITERATION=$((ITERATION + 1))

    # Check if instructions.md has new content
    if [ -f "instructions.md" ]; then
        CURRENT_SIZE=$(get_file_size "instructions.md")
        if [ "$CURRENT_SIZE" -gt "$LAST_INSTRUCTIONS_SIZE" ]; then
            echo ""
            echo "📨 [Iteration $ITERATION] New instructions detected! Size: $LAST_INSTRUCTIONS_SIZE → $CURRENT_SIZE bytes"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

            # Run Claude to process new instructions
            claude --print --dangerously-skip-permissions \
                --append-system-prompt "$(cat persona-instructions.md)" \
                "NEW INSTRUCTIONS RECEIVED! Read instructions.md from byte position $LAST_INSTRUCTIONS_SIZE onwards. Act on the new instructions immediately. Update your tasks.md file accordingly."

            LAST_INSTRUCTIONS_SIZE=$CURRENT_SIZE
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        fi
    fi

    # Check if tasks.md was updated (by self or others)
    if [ -f "tasks.md" ]; then
        CURRENT_TASKS_SIZE=$(get_file_size "tasks.md")
        if [ "$CURRENT_TASKS_SIZE" -gt "$LAST_TASKS_SIZE" ]; then
            echo ""
            echo "📋 [Iteration $ITERATION] Tasks file updated"
            LAST_TASKS_SIZE=$CURRENT_TASKS_SIZE

            # Run Claude to check task progress
            claude --print --dangerously-skip-permissions \
                --append-system-prompt "$(cat persona-instructions.md)" \
                "Check your tasks.md file for current task status. Continue working on in-progress tasks or start the next not-started task."
        fi
    fi

    # Brief status check every iteration
    if [ $((ITERATION %% 12)) -eq 0 ]; then
        echo ""
        echo "💭 [Iteration $ITERATION] Periodic check-in (1 minute elapsed)"
        claude --print --dangerously-skip-permissions \
            --append-system-prompt "$(cat persona-instructions.md)" \
            "Quick status check: Review your current tasks and progress. If blocked or waiting, state what you need. If working, provide brief update."
    fi

    sleep 5
done
`, absSessionDir, sessionID)
	return script
}

// generateInstructions creates comprehensive instructions for a persona
func (o *Orchestrator) generateInstructions(p *persona.Persona, sess *session.Session) string {
	// Get absolute path for persona directory
	absWorkspace, _ := filepath.Abs(o.workspacePath)
	absPersonaDir := filepath.Join(absWorkspace, sess.ID)

	instructions := fmt.Sprintf(`%s

## Your Session Information
Session ID: %s
Your Persona Directory: %s/
Your Role: %s
Working Directory: PROJECT ROOT (current directory)

## Important: Working Directory
- You are running from the PROJECT ROOT directory (where the project was initialized)
- All your work (code, files, etc.) should be created in the current directory or its subdirectories
- Your persona-specific files are in: %s/
- Reference your persona files using the full path above

## Files in Your Persona Directory
- %s/tasks.md: YOUR task list (you update this)
- %s/instructions.md: Instructions from others (read regularly)
- %s/tracker.json: Reading state tracker (automatic)
- %s/persona-instructions.md: Your role and capabilities

## Important Guidelines

### Automatic Instruction Monitoring
- The system checks %s/instructions.md every 5 seconds automatically
- When new instructions arrive, you'll be prompted to read them
- New instructions are appended with timestamps

### Update Your Tasks
- Update %s/tasks.md with your progress after completing work
- Use statuses: "not started", "in progress", "completed"
- When ALL tasks are completed, you will be automatically terminated
- The system will periodically check your progress

### Communication
- DO NOT modify other personas' files
- To assign work: Write to their instructions.md file (see below)
- For spawning new team members: Create request directories (see below)
- Write your deliverables to the current directory (project root)
- Your persona directory (%s/) is only for instructions/tasks tracking

`, p.Instructions, sess.ID, absPersonaDir, sess.PersonaName, absPersonaDir,
	absPersonaDir, absPersonaDir, absPersonaDir, absPersonaDir,
	absPersonaDir, absPersonaDir, absPersonaDir)

	// Add communication instructions
	instructions += fmt.Sprintf(`
## Communicating with Other Personas

To give instructions to another persona:
1. List their directories: ls %s/solutions-architect-*/
2. Append to their instructions.md file with a timestamp header
3. They will be automatically notified within 5 seconds

Example:
cat >> %s/solutions-architect-*/instructions.md <<EOF

## Instructions from %s ($(date '+%%Y-%%m-%%d %%H:%%M:%%S'))

Please design the database schema for the user management system.
Include the following tables: users, roles, permissions.
Provide an ERD diagram and SQL DDL statements.

EOF

`, absWorkspace, absWorkspace, sess.PersonaName)

	// Add request instructions based on persona type
	if sess.PersonaType == session.SessionTypeSoftwareEngineer {
		instructions += fmt.Sprintf(`
## Requesting Interns

If you need help with tests, linting, or documentation:

1. Create directory: %s/intern-request-{descriptive-name}/
2. Create: intern-request-{name}/instructions.md with their tasks
3. Project Manager will spawn the intern automatically
4. Wait for intern-{timestamp}/ directory to appear

Example:
mkdir %s/intern-request-test-helper
cat > %s/intern-request-test-helper/instructions.md <<EOF
Write unit tests for auth.go with >80%% coverage
EOF

`, absWorkspace, absWorkspace, absWorkspace)
	} else if sess.PersonaType == session.SessionTypeEngineeringManager || sess.PersonaType == session.SessionTypeSolutionsArchitect {
		instructions += fmt.Sprintf(`
## Requesting Software Engineers

If you need additional engineers:

1. Create directory: %s/software-engineer-request-{descriptive-name}/
2. Create: software-engineer-request-{name}/instructions.md with their tasks
3. Project Manager will spawn the engineer automatically
4. Wait for software-engineer-{timestamp}/ directory to appear

Example:
mkdir %s/software-engineer-request-api-developer
cat > %s/software-engineer-request-api-developer/instructions.md <<EOF
Implement REST API endpoints according to the spec in shared/api-spec.md
EOF

`, absWorkspace, absWorkspace, absWorkspace)
	}

	instructions += `
## Completion
When all your tasks are marked "completed", you will be automatically terminated and your work will be archived.
`

	return instructions
}

// GetStatus returns current orchestrator status
func (o *Orchestrator) GetStatus() (string, error) {
	sessions, err := o.sm.GetAllSessions()
	if err != nil {
		return "", err
	}

	status := fmt.Sprintf("Orchestrator Status\n")
	status += fmt.Sprintf("===================\n\n")
	status += fmt.Sprintf("Running Sessions: %d\n", len(o.activeSessions))
	status += fmt.Sprintf("Total Sessions: %d\n\n", len(sessions))

	for sessionID := range o.activeSessions {
		tmuxSessionName := fmt.Sprintf("claude-%s", sessionID)
		status += fmt.Sprintf("  %s (tmux: %s)\n", sessionID, tmuxSessionName)
	}

	return status, nil
}
