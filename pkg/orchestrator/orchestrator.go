package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	tuiMode         bool // Silent mode for TUI
	startTime       time.Time
	totalSpawned    int
	completedCount  int
	failedCount     int
	tmuxSession     string   // The tmux session this orchestrator is running in
	spawnedSessions []string // List of all spawned tmux session IDs
}

// OrchestratorState represents the orchestrator's state in JSON
type OrchestratorState struct {
	ID                  string    `json:"id"`
	Status              string    `json:"status"`
	StartTime           time.Time `json:"start_time"`
	CurrentWork         string    `json:"current_work"`
	TotalSessionsSpawned int      `json:"total_sessions_spawned"`
	ActiveSessions      int       `json:"active_sessions"`
	CompletedSessions   int       `json:"completed_sessions"`
	FailedSessions      int       `json:"failed_sessions"`
	TmuxSession         string    `json:"tmux_session,omitempty"`
	SpawnedSessions     []string  `json:"spawned_sessions"` // List of all spawned tmux session IDs
}

// log prints a message unless in TUI mode
func (o *Orchestrator) log(format string, args ...interface{}) {
	if !o.tuiMode {
		fmt.Printf(format, args...)
	}
}

// logln prints a line unless in TUI mode
func (o *Orchestrator) logln(args ...interface{}) {
	if !o.tuiMode {
		fmt.Println(args...)
	}
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

	orch := &Orchestrator{
		sm:              sm,
		personas:        personas,
		activeSessions:  make(map[string]bool),
		workspacePath:   workspacePath,
		pollInterval:    5 * time.Second,
		verbose:         verbose,
		startTime:       time.Now(),
		spawnedSessions: make([]string, 0),
	}

	// Detect tmux session name if running inside tmux
	if tmuxEnv := os.Getenv("TMUX"); tmuxEnv != "" {
		// TMUX env format: /tmp/tmux-501/default,12345,0
		// Extract session name using tmux command
		cmd := exec.Command("tmux", "display-message", "-p", "#S")
		if output, err := cmd.Output(); err == nil && len(output) > 0 {
			orch.tmuxSession = strings.TrimSpace(string(output))
		}
	}

	// Create orchestrator directory and initialize state
	orchestratorDir := filepath.Join(workspacePath, "orchestrator")
	if err := os.MkdirAll(orchestratorDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create orchestrator directory: %w", err)
	}

	// Load existing state if it exists (to restore spawned sessions list)
	orch.loadState()

	// Save initial state
	orch.saveState()

	return orch, nil
}

// Run starts the orchestrator daemon
func (o *Orchestrator) Run() error {
	o.logln("🎯 Project Manager Orchestrator Started")
	o.log("   Workspace: %s\n", o.workspacePath)
	o.log("   Poll Interval: %v\n", o.pollInterval)
	o.logln()

	// Start cost monitor in background
	costMonitor := NewCostMonitor(o.sm)
	go func() {
		costMonitor.Start()
	}()

	ticker := time.NewTicker(o.pollInterval)
	defer ticker.Stop()

	// Initial scan
	if err := o.scanAndProcess(); err != nil {
		o.log("⚠️  Error in initial scan: %v\n", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := o.scanAndProcess(); err != nil {
				o.log("⚠️  Error in scan: %v\n", err)
			}
		}
	}
}

// RunTUI starts the orchestrator with interactive TUI
func (o *Orchestrator) RunTUI() error {
	// TODO: Integrate with new static TUI once ready
	// For now, just run the static TUI without orchestrator integration
	return RunStaticTUI()
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

	// 4. Update orchestrator state
	o.saveState()

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
				o.log("⚠️  Failed to handle spawn request %s: %v\n", dirName, err)
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
					o.log("⚠️  Failed to spawn session %s: %v\n", dirName, err)
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

	// Handle request directories (dynamic spawns)
	if strings.HasPrefix(dirName, "solutions-architect-request-") {
		personaType = session.SessionTypeSolutionsArchitect
	} else if strings.HasPrefix(dirName, "software-engineer-request-") {
		personaType = session.SessionTypeSoftwareEngineer
	} else if strings.HasPrefix(dirName, "qa-request-") {
		personaType = session.SessionTypeQA
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
	} else if strings.HasPrefix(dirName, "qa-") {
		// Initial QA spawn
		personaType = session.SessionTypeQA
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

	// Mark request directory as active immediately to prevent duplicate spawns
	// This is critical because for request directories, a new session ID will be generated
	// and we need to track BOTH the request directory name AND the new session ID
	o.activeSessions[dirName] = true

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
				o.log("⚠️  Failed to copy instructions: %v\n", err)
			}
		}

		// Remove request directory
		if err := os.RemoveAll(requestPath); err != nil {
			o.log("⚠️  Failed to remove request directory: %v\n", err)
		}
	}

	o.log("\n🚀 Spawning %s: %s\n", personaType, sess.PersonaName)

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

	// Create wrapper script that keeps Claude alive and monitors for new instructions
	wrapperScript := o.createWrapperScript(sess.ID, absSessionDir)
	wrapperPath := filepath.Join(absSessionDir, "worker.sh")
	if err := os.WriteFile(wrapperPath, []byte(wrapperScript), 0755); err != nil {
		return fmt.Errorf("failed to create wrapper script: %w", err)
	}

	// Create tmux session and run the wrapper script
	cmd := exec.Command("tmux", "new-session", "-d", "-s", tmuxSessionName, "bash", wrapperPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start tmux session: %w (output: %s)", err, string(output))
	}

	// Track this spawned session
	o.spawnedSessions = append(o.spawnedSessions, tmuxSessionName)

	// Update session.json with tmux info
	if err := o.sm.UpdateTmuxSession(sess.ID, tmuxSessionName, true); err != nil {
		o.log("⚠️  Failed to update tmux session info: %v\n", err)
	}

	// Write attach command file to persona directory
	attachCmd := fmt.Sprintf("#!/bin/bash\nclear\ntmux attach -t %s\n", tmuxSessionName)
	attachFile := filepath.Join(absSessionDir, "attach.sh")
	if err := os.WriteFile(attachFile, []byte(attachCmd), 0755); err != nil {
		o.log("⚠️  Failed to write attach command: %v\n", err)
	}

	// Mark session as active
	o.activeSessions[sess.ID] = true
	o.totalSpawned++

	o.log("   ✅ Session: %s (tmux: %s)\n", sess.ID, tmuxSessionName)
	o.log("   📎 Attach with: tmux attach -t %s\n", tmuxSessionName)
	o.log("   📄 Or run: %s/attach.sh\n", absSessionDir)

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
			o.log("\n🎉 All tasks completed for %s (%s)\n", sess.PersonaName, sess.ID)

			// Terminate tmux session if still running
			if o.isTmuxSessionRunning(sess.ID) {
				tmuxSessionName := fmt.Sprintf("claude-%s", sess.ID)
				exec.Command("tmux", "kill-session", "-t", tmuxSessionName).Run()
				delete(o.activeSessions, sess.ID)
			}

			// Mark as completed
			o.sm.UpdateSessionStatus(sess.ID, "completed")
			o.completedCount++

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

	o.log("   📦 Archived to: %s\n", newPath)
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
				o.log("\n⚠️  Session stopped: %s (%s)\n", personaName, sessionID)
			} else {
				o.log("\n⚠️  Session stopped: %s\n", sessionID)
			}

			delete(o.activeSessions, sessionID)
			o.sm.UpdateSessionStatus(sessionID, "stopped")

			// Check if it was manually killed vs completed
			tasks, err := o.sm.ReadTasks(sessionID)
			if err == nil && o.areAllTasksCompleted(tasks) {
				o.log("   📋 All tasks were completed\n")
				o.sm.UpdateSessionStatus(sessionID, "completed")
				o.completedCount++
			} else {
				o.log("   📋 Session did not complete all tasks\n")
				o.failedCount++
			}
		}
	}

	return nil
}

// createWrapperScript creates a shell script that runs Claude interactively with background monitoring
func (o *Orchestrator) createWrapperScript(sessionID, sessionDir string) string {
	// Get absolute path
	absSessionDir, _ := filepath.Abs(sessionDir)
	script := fmt.Sprintf(`#!/bin/bash
set -e

SESSION_DIR="%s"
cd "$SESSION_DIR"

echo "🤖 Starting Claude worker for session: %s"
echo "📂 Working directory: $SESSION_DIR"
echo ""

# Function to get file size (cross-platform)
get_file_size() {
    if [ -f "$1" ]; then
        wc -c < "$1" | tr -d ' '
    else
        echo "0"
    fi
}

# Start background monitoring script
(
    LAST_INSTRUCTIONS_SIZE=$(get_file_size "instructions.md")
    LAST_TASKS_SIZE=$(get_file_size "tasks.md")

    while true; do
        sleep 5

        # Check for manual ping file
        if [ -f ".ping" ]; then
            rm .ping
            echo ""
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "🔔 PING! Manual check requested."
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo ""
        fi

        # Check if instructions.md has new content
        if [ -f "instructions.md" ]; then
            CURRENT_SIZE=$(get_file_size "instructions.md")
            if [ "$CURRENT_SIZE" -gt "$LAST_INSTRUCTIONS_SIZE" ]; then
                echo ""
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo "📨 NEW INSTRUCTIONS DETECTED!"
                echo "   Previous size: $LAST_INSTRUCTIONS_SIZE bytes"
                echo "   Current size:  $CURRENT_SIZE bytes"
                echo ""
                echo "   👉 Ask me to check instructions.md for new tasks!"
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo ""
                LAST_INSTRUCTIONS_SIZE=$CURRENT_SIZE
            fi
        fi

        # Check if tasks.md was updated
        if [ -f "tasks.md" ]; then
            CURRENT_TASKS_SIZE=$(get_file_size "tasks.md")
            if [ "$CURRENT_TASKS_SIZE" -gt "$LAST_TASKS_SIZE" ]; then
                echo ""
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo "📋 TASKS FILE UPDATED!"
                echo "   👉 Check tasks.md for updates!"
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo ""
                LAST_TASKS_SIZE=$CURRENT_TASKS_SIZE
            fi
        fi
    done
) &

# Save background process PID
MONITOR_PID=$!
echo "📡 Background monitor started (PID: $MONITOR_PID)"
echo "   Checking for new instructions every 5 seconds"
echo ""

# Cleanup function to kill background monitor on exit
cleanup() {
    kill $MONITOR_PID 2>/dev/null || true
}
trap cleanup EXIT

# Start Claude in interactive mode with initial instructions
claude --dangerously-skip-permissions \
    --append-system-prompt "$(cat persona-instructions.md)" \
    "Read your tasks.md file and start working. You're running in INTERACTIVE mode - a background script monitors for new instructions and will alert you. When you see a notification about new instructions, read instructions.md and act on them."
`, absSessionDir, sessionID)
	return script
}

// generateInstructions creates comprehensive instructions for a persona
func (o *Orchestrator) generateInstructions(p *persona.Persona, sess *session.Session) string {
	// Get absolute path for persona directory
	absWorkspace, _ := filepath.Abs(o.workspacePath)
	absPersonaDir := filepath.Join(absWorkspace, sess.ID)

	// Read CLAUDE.md if it exists for project-specific instructions
	claudeMdPath := filepath.Join(o.workspacePath, "..", "CLAUDE.md")
	claudeMdContent := ""
	if data, err := os.ReadFile(claudeMdPath); err == nil {
		claudeMdContent = fmt.Sprintf(`
## Project Guidelines (from CLAUDE.md)
%s

`, string(data))
	}

	instructions := p.Instructions + "\n\n" + claudeMdContent + fmt.Sprintf(`
## Your Session Information
Session ID: %s
Your Persona Directory: %s/
Your Role: %s
Working Directory: PROJECT ROOT (current directory)

## IMPORTANT: Read Shell Configuration First
Before starting work, read ~/.zshrc to discover available commands, aliases, and functions:
- Custom functions defined by the user
- Useful aliases and shortcuts
- Environment-specific tools and utilities

Read ~/.zshrc NOW to understand your environment.

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
- A background task monitors your instructions.md every 5 seconds automatically
- When new instructions arrive, you'll be notified
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

`, sess.ID, absPersonaDir, sess.PersonaName, absPersonaDir,
	absPersonaDir, absPersonaDir, absPersonaDir, absPersonaDir,
	absPersonaDir, absPersonaDir)

	// Add communication instructions
	instructions += fmt.Sprintf(`
## Communicating with Other Agents

You can communicate with ANY agent - there are NO hierarchy restrictions.
Write to any agent's instructions.md file to give them tasks, ask questions, or provide feedback.

To send instructions to another agent:
1. List available agents: ls %s/*/instructions.md
2. Append to their instructions.md file with a timestamp header
3. They will be automatically notified within 5 seconds

Examples:

# Send instructions to Leader Agent
cat >> %s/engineering-manager-*/instructions.md <<EOF
## Instructions from %s ($(date '+%%Y-%%m-%%d %%H:%%M:%%S'))
We need to pivot the project direction. Please review and approve.
EOF

# Send instructions to Architect
cat >> %s/solutions-architect-*/instructions.md <<EOF
## Instructions from %s ($(date '+%%Y-%%m-%%d %%H:%%M:%%S'))
Please design the database schema for the user management system.
EOF

# Send instructions to any Coder
cat >> %s/software-engineer-*/instructions.md <<EOF
## Instructions from %s ($(date '+%%Y-%%m-%%d %%H:%%M:%%S'))
Implement the API endpoints according to the spec.
EOF

`, absWorkspace, absWorkspace, sess.PersonaName, absWorkspace, sess.PersonaName, absWorkspace, sess.PersonaName)

	// Add resource request instructions - ANY agent can request ANY resource
	instructions += fmt.Sprintf(`
## Requesting Additional Resources

ANY agent can request ANY type of resource - there are NO restrictions.
Need an architect? Request one. Need the leader's input? Request a conversation.

To request a new agent:
1. Create directory: %s/{agent-type}-request-{descriptive-name}/
2. Create: instructions.md in that directory with their initial task
3. Orchestrator will spawn the agent automatically
4. Directory will be renamed to {agent-type}-{timestamp}/

Available agent types:
- solutions-architect-request-* → Architecture Agent
- software-engineer-request-* → Coding Agent
- qa-request-* → QA Agent
- intern-request-* → Support Agent

Examples:

# Request an Architect
mkdir %s/solutions-architect-request-api-designer
cat > %s/solutions-architect-request-api-designer/instructions.md <<EOF
Design the REST API architecture for our user management system.
EOF

# Request a Coder
mkdir %s/software-engineer-request-backend
cat > %s/software-engineer-request-backend/instructions.md <<EOF
Implement the backend API endpoints according to the architecture spec.
EOF

# Request QA
mkdir %s/qa-request-api-tester
cat > %s/qa-request-api-tester/instructions.md <<EOF
Write integration tests for the user management API.
EOF

# Request Support
mkdir %s/intern-request-documentation
cat > %s/intern-request-documentation/instructions.md <<EOF
Write API documentation for all endpoints in OpenAPI format.
EOF

`, absWorkspace, absWorkspace, absWorkspace, absWorkspace, absWorkspace, absWorkspace, absWorkspace, absWorkspace, absWorkspace)

	instructions += `
## Completion
When all your tasks are marked "completed", you will be automatically terminated and your work will be archived.
`

	// Add background task instructions
	instructions += fmt.Sprintf(`
## IMPORTANT: Background Tasks

Start these two background tasks IMMEDIATELY when you begin:

### Task 1: Instruction Monitoring
Monitor your instructions.md file every 5 seconds. When new instructions arrive (file size increases), READ AND ACT ON THEM IMMEDIATELY.

Bash(PERSONA_DIR=%s; LAST_SIZE=0; while true; do if [ -f "$PERSONA_DIR/instructions.md" ]; then NEW_SIZE=$(wc -c < "$PERSONA_DIR/instructions.md" | tr -d " "); if [ "$NEW_SIZE" -gt "${LAST_SIZE:-0}" 2>/dev/null ]; then echo "🔔 NEW INSTRUCTIONS DETECTED! File grew from $LAST_SIZE to $NEW_SIZE bytes. READ instructions.md NOW and act on new tasks!"; fi; LAST_SIZE=$NEW_SIZE; fi; sleep 5; done, run_in_background=true)

### Task 2: Status Updates
Update your session.json with current_work every 10 seconds. Extract just the task title from tasks.md (details shown in popup).

Bash(PERSONA_DIR=%s; while true; do CURRENT=$(grep '^## Task:' $PERSONA_DIR/tasks.md 2>/dev/null | head -1 | sed 's/^## Task: //' || echo "No tasks assigned"); jq --arg status "$CURRENT" '.current_work = $status' $PERSONA_DIR/session.json > $PERSONA_DIR/session.tmp && mv $PERSONA_DIR/session.tmp $PERSONA_DIR/session.json; sleep 10; done, run_in_background=true)

## CRITICAL: After Completing Tasks

When you complete all your current tasks:
1. Check instructions.md for new assignments
2. If new instructions found, act on them immediately
3. If no new instructions, check again every 30 seconds
4. Update tasks.md with "Waiting for instructions" status

## Startup Sequence
1. Read ~/.zshrc to discover available commands and functions
2. Start both background tasks above
3. Begin working on your tasks from %s/tasks.md
`, absPersonaDir, absPersonaDir, absPersonaDir)

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

// saveState saves the orchestrator's current state to JSON
// loadState loads existing orchestrator state from disk
func (o *Orchestrator) loadState() error {
	stateFile := filepath.Join(o.workspacePath, "orchestrator", "state.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		// State file doesn't exist yet, that's ok
		return nil
	}

	var state OrchestratorState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	// Restore spawned sessions list
	if state.SpawnedSessions != nil {
		o.spawnedSessions = state.SpawnedSessions
	}

	return nil
}

func (o *Orchestrator) saveState() error {
	state := OrchestratorState{
		ID:                  "orchestrator",
		Status:              "active",
		StartTime:           o.startTime,
		CurrentWork:         o.generateCurrentWork(),
		TotalSessionsSpawned: o.totalSpawned,
		ActiveSessions:      len(o.activeSessions),
		CompletedSessions:   o.completedCount,
		FailedSessions:      o.failedCount,
		TmuxSession:         o.tmuxSession,
		SpawnedSessions:     o.spawnedSessions,
	}

	stateFile := filepath.Join(o.workspacePath, "orchestrator", "state.json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

// generateCurrentWork creates a concise status message
func (o *Orchestrator) generateCurrentWork() string {
	activeCount := len(o.activeSessions)
	if activeCount == 0 {
		return "Waiting for sessions to spawn"
	}
	if activeCount == 1 {
		return "Monitoring 1 session"
	}
	return fmt.Sprintf("Monitoring %d sessions", activeCount)
}

// KillAllSessions kills all spawned tmux sessions including the orchestrator
func (o *Orchestrator) KillAllSessions() error {
	killed := 0
	failed := 0

	// Kill all spawned agent sessions
	for _, tmuxSession := range o.spawnedSessions {
		cmd := exec.Command("tmux", "kill-session", "-t", tmuxSession)
		if err := cmd.Run(); err != nil {
			// Session might already be dead, that's ok
			failed++
		} else {
			killed++
		}
	}

	// Kill the orchestrator's own tmux session if it exists
	if o.tmuxSession != "" {
		cmd := exec.Command("tmux", "kill-session", "-t", o.tmuxSession)
		if err := cmd.Run(); err != nil {
			failed++
		} else {
			killed++
		}
	}

	o.log("💀 Killed %d sessions (%d already dead)\n", killed, failed)
	return nil
}
