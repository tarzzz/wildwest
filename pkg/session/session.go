package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tarzzz/wildwest/pkg/names"
)

// SessionType represents the type of persona session
type SessionType string

const (
	SessionTypeProjectManager     SessionType = "project-manager"
	SessionTypeEngineeringManager SessionType = "engineering-manager"
	SessionTypeSoftwareEngineer   SessionType = "software-engineer"
	SessionTypeIntern             SessionType = "intern"
	SessionTypeSolutionsArchitect SessionType = "solutions-architect"
	SessionTypeQA                 SessionType = "qa"
	SessionTypeDevOps             SessionType = "devops"
)

// Session represents a persona's active session
type Session struct {
	ID              string      `json:"id"`
	PersonaType     SessionType `json:"persona_type"`
	PersonaName     string      `json:"persona_name"`
	StartTime       time.Time   `json:"start_time"`
	Status          string      `json:"status"` // active, completed, failed
	WorkspaceID     string      `json:"workspace_id"`
	PID             int         `json:"pid,omitempty"`
	CurrentWork     string      `json:"current_work,omitempty"`     // One-liner status updated by worker
	TmuxSession     string      `json:"tmux_session,omitempty"`     // Tmux session name
	TmuxSpawned     bool        `json:"tmux_spawned"`               // Whether tmux session is spawned
	TmuxAttachCmd   string      `json:"tmux_attach_cmd,omitempty"`  // Command to attach to tmux session
	// Token usage tracking
	InputTokens     int64       `json:"input_tokens,omitempty"`     // Total input tokens used
	OutputTokens    int64       `json:"output_tokens,omitempty"`    // Total output tokens used
	TotalTokens     int64       `json:"total_tokens,omitempty"`     // Total tokens (input + output)
	EstimatedCost   float64     `json:"estimated_cost,omitempty"`   // Estimated cost in USD
	Model           string      `json:"model,omitempty"`            // Model used (sonnet, opus, haiku)
}

// Workspace manages the shared database directory
type Workspace struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	CreatedAt    time.Time `json:"created_at"`
	Description  string    `json:"description"`
	ActiveTasks  []string  `json:"active_tasks"`
}

// Message represents communication between personas
type Message struct {
	ID           string      `json:"id"`
	From         string      `json:"from"`         // Session ID
	FromPersona  SessionType `json:"from_persona"`
	To           string      `json:"to,omitempty"` // Empty means broadcast
	ToPersona    SessionType `json:"to_persona,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
	Type         string      `json:"type"` // task, question, response, notification
	Subject      string      `json:"subject"`
	Content      string      `json:"content"`
	Attachments  []string    `json:"attachments,omitempty"`
	ParentID     string      `json:"parent_id,omitempty"` // For threading
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusNotStarted TaskStatus = "not started"
	TaskStatusInProgress TaskStatus = "in progress"
	TaskStatusCompleted  TaskStatus = "completed"
)

// Task represents a task in a persona's task list
type Task struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	AssignedBy  string     `json:"assigned_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ReadTracker tracks what has been read from instructions.md and tasks.md
type ReadTracker struct {
	SessionID                string    `json:"session_id"`
	InstructionsLastRead     time.Time `json:"instructions_last_read"`
	InstructionsLastPosition int64     `json:"instructions_last_position"` // byte position in file
	TasksLastRead            time.Time `json:"tasks_last_read"`
	TasksLastPosition        int64     `json:"tasks_last_position"` // byte position in file
	LastCheckTime            time.Time `json:"last_check_time"`
}

// SessionManager manages persona sessions and workspace
type SessionManager struct {
	workspacePath string
	nameGen       *names.NameGenerator
}

// NewSessionManager creates a new session manager
func NewSessionManager(workspacePath string) (*SessionManager, error) {
	if workspacePath == "" {
		workspacePath = ".ww-db"
	}

	// Create workspace directory structure
	dirs := []string{
		workspacePath,
		filepath.Join(workspacePath, "shared"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	sm := &SessionManager{
		workspacePath: workspacePath,
		nameGen:       names.NewNameGenerator(),
	}

	// Load existing sessions and mark names as used
	if err := sm.loadExistingNames(); err != nil {
		// Non-fatal, just log
		fmt.Printf("Warning: failed to load existing names: %v\n", err)
	}

	return sm, nil
}

// loadExistingNames loads existing session names to avoid duplicates
func (sm *SessionManager) loadExistingNames() error {
	sessions, err := sm.GetAllSessions()
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		sm.nameGen.MarkUsed(sess.PersonaName)
	}

	return nil
}

// CreateSession creates a new session for a persona
func (sm *SessionManager) CreateSession(personaType SessionType, personaName string, workspaceID string) (*Session, error) {
	// Check singleton constraint
	if personaType == SessionTypeProjectManager || personaType == SessionTypeEngineeringManager || personaType == SessionTypeSolutionsArchitect {
		active, err := sm.GetActiveSessions()
		if err != nil {
			return nil, err
		}

		for _, s := range active {
			if s.PersonaType == personaType {
				return nil, fmt.Errorf("only one %s session is allowed at a time", personaType)
			}
		}
	}

	// Generate interesting name if not provided or if it's generic
	if personaName == "" || personaName == "manager" || personaName == "architect" ||
		strings.HasPrefix(personaName, "engineer-") || strings.HasPrefix(personaName, "intern-") {
		personaName = sm.nameGen.GetNameForPersona(string(personaType))
	}

	session := &Session{
		ID:          fmt.Sprintf("%s-%d", personaType, time.Now().UnixNano()/1000000), // Use milliseconds for uniqueness
		PersonaType: personaType,
		PersonaName: personaName,
		StartTime:   time.Now(),
		Status:      "active",
		WorkspaceID: workspaceID,
		PID:         os.Getpid(),
	}

	// Create persona's database directory
	personaDir := sm.getPersonaDir(session.ID)
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create persona directory: %w", err)
	}

	// Initialize tasks.md
	tasksPath := filepath.Join(personaDir, "tasks.md")
	initialTasks := "# Tasks\n\nNo tasks assigned yet.\n"
	if err := os.WriteFile(tasksPath, []byte(initialTasks), 0644); err != nil {
		return nil, fmt.Errorf("failed to create tasks.md: %w", err)
	}

	// Initialize tracker.json
	tracker := &ReadTracker{
		SessionID:                session.ID,
		InstructionsLastRead:     time.Time{},
		InstructionsLastPosition: 0,
		TasksLastRead:            time.Time{},
		TasksLastPosition:        0,
		LastCheckTime:            time.Now(),
	}
	if err := sm.saveTracker(session.ID, tracker); err != nil {
		return nil, fmt.Errorf("failed to create tracker: %w", err)
	}

	if err := sm.saveSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

// getPersonaDir returns the database directory path for a persona
func (sm *SessionManager) getPersonaDir(sessionID string) string {
	return filepath.Join(sm.workspacePath, sessionID)
}

// GetActiveSessions returns all active sessions
func (sm *SessionManager) GetActiveSessions() ([]*Session, error) {
	entries, err := os.ReadDir(sm.workspacePath)
	if err != nil {
		return nil, err
	}

	var sessions []*Session
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "shared" {
			continue
		}

		sessionFile := filepath.Join(sm.workspacePath, entry.Name(), "session.json")
		data, err := os.ReadFile(sessionFile)
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		if session.Status == "active" {
			sessions = append(sessions, &session)
		}
	}

	return sessions, nil
}

// GetAllSessions returns all sessions (active or not)
func (sm *SessionManager) GetAllSessions() ([]*Session, error) {
	entries, err := os.ReadDir(sm.workspacePath)
	if err != nil {
		return nil, err
	}

	var sessions []*Session
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "shared" {
			continue
		}

		// Skip archived directories
		if strings.HasSuffix(entry.Name(), "-archived") || strings.HasSuffix(entry.Name(), "-completed") {
			continue
		}

		sessionFile := filepath.Join(sm.workspacePath, entry.Name(), "session.json")
		data, err := os.ReadFile(sessionFile)
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// UpdateSessionStatus updates the status of a session
func (sm *SessionManager) UpdateSessionStatus(sessionID string, status string) error {
	sessionPath := filepath.Join(sm.workspacePath, sessionID, "session.json")

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}

	session.Status = status
	return sm.saveSession(&session)
}

// UpdateCurrentWork updates the current work status for a session
func (sm *SessionManager) UpdateCurrentWork(sessionID string, currentWork string) error {
	sessionPath := filepath.Join(sm.workspacePath, sessionID, "session.json")

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}

	session.CurrentWork = currentWork
	return sm.saveSession(&session)
}

// UpdateTmuxSession updates the tmux session information for a session
func (sm *SessionManager) UpdateTmuxSession(sessionID string, tmuxSession string, spawned bool) error {
	sessionPath := filepath.Join(sm.workspacePath, sessionID, "session.json")

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}

	session.TmuxSession = tmuxSession
	session.TmuxSpawned = spawned
	session.TmuxAttachCmd = fmt.Sprintf("tmux attach -t %s", tmuxSession)
	return sm.saveSession(&session)
}

// WriteInstructions writes instructions for a target persona
func (sm *SessionManager) WriteInstructions(fromSessionID, toSessionID, instructions string) error {
	targetDir := sm.getPersonaDir(toSessionID)
	instructionsPath := filepath.Join(targetDir, "instructions.md")

	// Read existing instructions if any
	var existingInstructions string
	if data, err := os.ReadFile(instructionsPath); err == nil {
		existingInstructions = string(data)
	}

	// Append new instructions with timestamp and source
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newInstructions := fmt.Sprintf("\n\n---\n## Instructions from %s (%s)\n\n%s\n",
		fromSessionID, timestamp, instructions)

	content := existingInstructions + newInstructions

	return os.WriteFile(instructionsPath, []byte(content), 0644)
}

// ReadTasks reads the tasks.md file for a persona
func (sm *SessionManager) ReadTasks(sessionID string) (string, error) {
	tasksPath := filepath.Join(sm.getPersonaDir(sessionID), "tasks.md")
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpdateTasks updates the tasks.md file for a persona
func (sm *SessionManager) UpdateTasks(sessionID string, tasks string) error {
	tasksPath := filepath.Join(sm.getPersonaDir(sessionID), "tasks.md")
	return os.WriteFile(tasksPath, []byte(tasks), 0644)
}

// AddTask adds a new task to a persona's task list
func (sm *SessionManager) AddTask(sessionID string, description string, assignedBy string) error {
	tasksPath := filepath.Join(sm.getPersonaDir(sessionID), "tasks.md")

	// Read existing tasks
	var existingTasks string
	if data, err := os.ReadFile(tasksPath); err == nil {
		existingTasks = string(data)
	}

	// Add new task
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newTask := fmt.Sprintf("\n## Task: %s\n- **Status**: not started\n- **Assigned by**: %s\n- **Created**: %s\n",
		description, assignedBy, timestamp)

	content := existingTasks + newTask

	return os.WriteFile(tasksPath, []byte(content), 0644)
}

// ReadInstructions reads instructions for a persona
func (sm *SessionManager) ReadInstructions(sessionID string) (string, error) {
	instructionsPath := filepath.Join(sm.getPersonaDir(sessionID), "instructions.md")
	data, err := os.ReadFile(instructionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No instructions yet
		}
		return "", err
	}
	return string(data), nil
}

// WriteOutput writes output for a session
func (sm *SessionManager) WriteOutput(sessionID string, filename string, content string) error {
	outputPath := filepath.Join(sm.getPersonaDir(sessionID), filename)
	return os.WriteFile(outputPath, []byte(content), 0644)
}

// ReadOutput reads an output file from a session
func (sm *SessionManager) ReadOutput(sessionID string, filename string) (string, error) {
	outputPath := filepath.Join(sm.getPersonaDir(sessionID), filename)
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListPersonaFiles lists all files in a persona's directory
func (sm *SessionManager) ListPersonaFiles(sessionID string) ([]string, error) {
	personaDir := sm.getPersonaDir(sessionID)
	entries, err := os.ReadDir(personaDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "session.json" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ReadSharedFile reads a file from the shared directory
func (sm *SessionManager) ReadSharedFile(filename string) (string, error) {
	path := filepath.Join(sm.workspacePath, "shared", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteSharedFile writes a file to the shared directory
func (sm *SessionManager) WriteSharedFile(filename string, content string) error {
	path := filepath.Join(sm.workspacePath, "shared", filename)
	return os.WriteFile(path, []byte(content), 0644)
}

// saveSession saves a session to disk
func (sm *SessionManager) saveSession(session *Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	sessionPath := filepath.Join(sm.workspacePath, session.ID, "session.json")
	return os.WriteFile(sessionPath, data, 0644)
}

// GetWorkspacePath returns the workspace path
func (sm *SessionManager) GetWorkspacePath() string {
	return sm.workspacePath
}

// GetTracker retrieves the read tracker for a session
func (sm *SessionManager) GetTracker(sessionID string) (*ReadTracker, error) {
	trackerPath := filepath.Join(sm.getPersonaDir(sessionID), "tracker.json")

	data, err := os.ReadFile(trackerPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new tracker if doesn't exist
			tracker := &ReadTracker{
				SessionID:     sessionID,
				LastCheckTime: time.Now(),
			}
			return tracker, nil
		}
		return nil, err
	}

	var tracker ReadTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return nil, err
	}

	return &tracker, nil
}

// saveTracker saves the read tracker
func (sm *SessionManager) saveTracker(sessionID string, tracker *ReadTracker) error {
	data, err := json.MarshalIndent(tracker, "", "  ")
	if err != nil {
		return err
	}

	trackerPath := filepath.Join(sm.getPersonaDir(sessionID), "tracker.json")
	return os.WriteFile(trackerPath, data, 0644)
}

// GetNewInstructions returns only new instructions since last read
func (sm *SessionManager) GetNewInstructions(sessionID string) (string, error) {
	tracker, err := sm.GetTracker(sessionID)
	if err != nil {
		return "", err
	}

	instructionsPath := filepath.Join(sm.getPersonaDir(sessionID), "instructions.md")

	// Check if file exists
	fileInfo, err := os.Stat(instructionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No instructions yet
		}
		return "", err
	}

	// If file hasn't changed, no new instructions
	if !fileInfo.ModTime().After(tracker.InstructionsLastRead) {
		return "", nil
	}

	// Read file from last position
	file, err := os.Open(instructionsPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Seek to last read position
	if _, err := file.Seek(tracker.InstructionsLastPosition, 0); err != nil {
		return "", err
	}

	// Read new content
	var newContent strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			newContent.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	// Update tracker
	newPosition, _ := file.Seek(0, 2) // Seek to end to get new position
	tracker.InstructionsLastRead = time.Now()
	tracker.InstructionsLastPosition = newPosition
	tracker.LastCheckTime = time.Now()

	if err := sm.saveTracker(sessionID, tracker); err != nil {
		return "", err
	}

	return newContent.String(), nil
}

// CheckForUpdates checks if there are new instructions or task updates
func (sm *SessionManager) CheckForUpdates(sessionID string) (bool, string, error) {
	tracker, err := sm.GetTracker(sessionID)
	if err != nil {
		return false, "", err
	}

	var updates []string
	hasUpdates := false

	// Check instructions.md
	instructionsPath := filepath.Join(sm.getPersonaDir(sessionID), "instructions.md")
	if fileInfo, err := os.Stat(instructionsPath); err == nil {
		if fileInfo.ModTime().After(tracker.InstructionsLastRead) {
			hasUpdates = true
			updates = append(updates, "New instructions received")
		}
	}

	// Check tasks.md
	tasksPath := filepath.Join(sm.getPersonaDir(sessionID), "tasks.md")
	if fileInfo, err := os.Stat(tasksPath); err == nil {
		if fileInfo.ModTime().After(tracker.TasksLastRead) {
			hasUpdates = true
			updates = append(updates, "Tasks have been updated")
		}
	}

	// Update last check time
	tracker.LastCheckTime = time.Now()
	sm.saveTracker(sessionID, tracker)

	return hasUpdates, strings.Join(updates, ", "), nil
}

// CreateWorkspace creates a new workspace
func (sm *SessionManager) CreateWorkspace(description string) (*Workspace, error) {
	workspace := &Workspace{
		ID:          fmt.Sprintf("ws-%d", time.Now().Unix()),
		Path:        sm.workspacePath,
		CreatedAt:   time.Now(),
		Description: description,
		ActiveTasks: []string{},
	}

	data, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return nil, err
	}

	workspacePath := filepath.Join(sm.workspacePath, "workspace.json")
	if err := os.WriteFile(workspacePath, data, 0644); err != nil {
		return nil, err
	}

	return workspace, nil
}

// GetCurrentWork generates an intelligent summary of what the team member is working on
func (sm *SessionManager) GetCurrentWork(sessionID string) string {
	personaDir := sm.getPersonaDir(sessionID)

	// Check if directory exists
	if _, err := os.Stat(personaDir); os.IsNotExist(err) {
		return "Directory not found"
	}

	// Get claude binary path (respects CLAUDE_BIN env var)
	claudeBin := os.Getenv("CLAUDE_BIN")
	if claudeBin == "" {
		claudeBin = "claude"
	}

	// Use claude -p to generate a concise summary
	prompt := `Analyze this persona's workspace and provide a ONE-LINE summary (max 100 chars) of what they are currently working on.

Look at:
- tasks.md for assigned tasks and their status
- Any recent files they've created or modified
- instructions.md for context

Output ONLY the one-line summary, nothing else. Use present tense.
Examples:
- "Implementing user authentication endpoints"
- "Designing database schema for orders"
- "Writing unit tests for payment service"
- "Awaiting task assignment"
- "All tasks completed"

If tasks.md shows "in progress", focus on that task. If only "not started", say "Awaiting: [task]".`

	// Set timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmdWithCtx := exec.CommandContext(ctx, claudeBin, "-p", prompt)
	cmdWithCtx.Dir = personaDir

	output, err := cmdWithCtx.CombinedOutput()
	if err != nil {
		// Fallback to simple parsing if claude fails
		return sm.getSimpleCurrentWork(sessionID)
	}

	summary := strings.TrimSpace(string(output))

	// Clean up the output - remove any markdown, quotes, or extra formatting
	summary = strings.Trim(summary, "`\"'")
	summary = strings.TrimPrefix(summary, "Summary: ")
	summary = strings.TrimPrefix(summary, "Currently: ")

	// Ensure it's not too long
	if len(summary) > 100 {
		summary = summary[:97] + "..."
	}

	// If empty or too short, fallback
	if len(summary) < 5 {
		return sm.getSimpleCurrentWork(sessionID)
	}

	return summary
}

// getSimpleCurrentWork is a fallback that parses tasks.md directly
func (sm *SessionManager) getSimpleCurrentWork(sessionID string) string {
	tasksContent, err := sm.ReadTasks(sessionID)
	if err != nil {
		return "No tasks found"
	}

	// Parse tasks.md to find current work
	lines := strings.Split(tasksContent, "\n")
	var currentTask string
	var currentStatus string
	inTask := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for task header
		if strings.HasPrefix(trimmed, "## Task:") {
			currentTask = strings.TrimPrefix(trimmed, "## Task:")
			currentTask = strings.TrimSpace(currentTask)
			inTask = true
			currentStatus = ""
			continue
		}

		// Check for status line
		if inTask && strings.HasPrefix(trimmed, "- **Status**:") {
			currentStatus = strings.TrimPrefix(trimmed, "- **Status**:")
			currentStatus = strings.TrimSpace(currentStatus)

			// If status is "in progress", return this task immediately
			if currentStatus == "in progress" {
				if len(currentTask) > 80 {
					return currentTask[:77] + "..."
				}
				return currentTask
			}
		}
	}

	// If no "in progress" task found, look for first "not started" task
	inTask = false
	currentTask = ""
	currentStatus = ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## Task:") {
			currentTask = strings.TrimPrefix(trimmed, "## Task:")
			currentTask = strings.TrimSpace(currentTask)
			inTask = true
			currentStatus = ""
			continue
		}

		if inTask && strings.HasPrefix(trimmed, "- **Status**:") {
			currentStatus = strings.TrimPrefix(trimmed, "- **Status**:")
			currentStatus = strings.TrimSpace(currentStatus)

			if currentStatus == "not started" {
				if len(currentTask) > 80 {
					return "Awaiting: " + currentTask[:72] + "..."
				}
				return "Awaiting: " + currentTask
			}
		}
	}

	// If all tasks are completed
	if strings.Contains(tasksContent, "completed") {
		return "All tasks completed"
	}

	return "No active tasks"
}
