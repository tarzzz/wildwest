package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tarzzz/wildwest/pkg/session"
)

// TickMsg is sent every 2 seconds to refresh session data
type TickMsg time.Time

// Component represents a node in the org chart
type Component struct {
	ID            string
	Name          string
	Role          string
	Description   string
	Emoji         string
	Status        string // "idle", "active", "unavailable"
	StatusMessage string // Brief statement about what they're doing
	TmuxSpawned   bool   // Whether tmux session is spawned
	TmuxSession   string // Tmux session name
}

// OrgChartModel is the TUI model for a static org chart
type OrgChartModel struct {
	components       []Component
	selectedIndex    int
	showingDetails   bool
	width            int
	height           int
	orchestrator     *Orchestrator
	sessionManager   *session.SessionManager
	workspacePath    string
	activeSessions   []*session.Session
	logs             []string
	maxLogs          int
	tickCount        int  // Track ticks for less frequent updates
	initialized      bool // Track if we've done initial load
	attachToSession  string // Tmux session to attach to on exit
	version          string // Version info for display
	goBack           bool   // Signal to return to session selector
}

// Styles
var (
	listItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("252"))

	selectedListItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Background(lipgloss.Color("235"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39"))

	detailsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(1, 2).
			MarginTop(1).
			MarginLeft(2)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 2)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	activeStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true)

	idleStatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226"))

	unavailableStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Italic(true).
				PaddingLeft(2)

	logsHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			PaddingLeft(2).
			MarginTop(1)

	logLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			PaddingLeft(2)

	logsBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("240")).
			MarginTop(1)

	liveOutputStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252")).
			Padding(1).
			MarginTop(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238"))

	liveOutputHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)
)

// NewOrgChartModel creates a new static org chart TUI
func NewOrgChartModel(orch *Orchestrator, sm *session.SessionManager, workspacePath, version string) OrgChartModel {
	// Start with empty components - will be populated from real sessions
	return OrgChartModel{
		components:     make([]Component, 0),
		selectedIndex:  0,
		showingDetails: false,  // Start with details hidden, press 'd' to show
		orchestrator:   orch,
		sessionManager: sm,
		workspacePath:  workspacePath,
		activeSessions: make([]*session.Session, 0),
		logs:           make([]string, 0),
		version:        version,
		maxLogs:        5,
	}
}

func (m OrgChartModel) Init() tea.Cmd {
	// Don't start orchestrator in TUI mode - just read sessions
	// The orchestrator should be run separately with: wildwest orchestrate --workspace .ww-db
	// This keeps the TUI responsive

	// Fire immediate tick for initialization, then regular ticks
	return tea.Batch(
		func() tea.Msg { return TickMsg(time.Now()) },
		tickCmd(),
	)
}

// tickCmd returns a tick command that fires every 2 seconds
func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m OrgChartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "b":
			// Go back to session selector
			m.goBack = true
			return m, tea.Quit

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.components)-1 {
				m.selectedIndex++
			}

		case "d":
			// Toggle details popup
			m.showingDetails = !m.showingDetails

		case "a":
			// Attach to selected tmux session
			if m.selectedIndex >= 0 && m.selectedIndex < len(m.components) {
				comp := m.components[m.selectedIndex]
				if comp.TmuxSpawned && comp.TmuxSession != "" {
					m.attachToSession = comp.TmuxSession
					return m, tea.Quit
				}
			}

		case "K":
			// Kill session and delete database files
			return m, m.killSession()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		// Do initial load on first tick
		if !m.initialized {
			m.initialized = true
			m.addLog(fmt.Sprintf("Monitoring %s", m.workspacePath))

			if m.sessionManager != nil {
				m.addLog("Loading sessions...")
				sessions, err := m.sessionManager.GetActiveSessions()
				if err != nil {
					m.addLog(fmt.Sprintf("Error loading: %v", err))
				} else {
					m.addLog(fmt.Sprintf("Found %d sessions", len(sessions)))
					m.activeSessions = sessions
					m.updateComponentsFromSessions()
					m.addLog(fmt.Sprintf("Created %d components", len(m.components)))
					if len(sessions) > 0 {
						m.addLog(m.generateStatusSummary())
					} else {
						m.addLog("No active sessions in workspace")
					}
				}
			} else {
				m.addLog("ERROR: SessionManager is nil")
			}
			return m, tickCmd()
		}

		m.tickCount++

		// Only refresh sessions every 3 ticks (6 seconds) to avoid blocking UI
		if m.tickCount%3 == 0 && m.sessionManager != nil {
			sessions, err := m.sessionManager.GetActiveSessions()
			if err == nil {
				oldCount := len(m.activeSessions)
				newCount := len(sessions)

				// Only update if count changed
				if newCount != oldCount {
					m.activeSessions = sessions
					m.updateComponentsFromSessions()

					// Log status summary on changes
					if newCount > 0 {
						if newCount > oldCount {
							m.addLog(m.generateSpawnMessage(sessions, m.activeSessions))
						}
						m.addLog(m.generateStatusSummary())
					} else {
						m.addLog("All sessions ended")
					}

					// Ensure selectedIndex is still valid
					if m.selectedIndex >= len(m.components) {
						m.selectedIndex = len(m.components) - 1
					}
					if m.selectedIndex < 0 && len(m.components) > 0 {
						m.selectedIndex = 0
					}
				}
			}
		}
		// Schedule next tick
		return m, tickCmd()
	}

	return m, nil
}

// addLog adds a log message with timestamp
func (m *OrgChartModel) addLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logLine := fmt.Sprintf("[%s] %s", timestamp, message)

	m.logs = append(m.logs, logLine)

	// Keep only the last N logs
	if len(m.logs) > m.maxLogs {
		m.logs = m.logs[len(m.logs)-m.maxLogs:]
	}
}

// loadOrchestratorState loads orchestrator as top-level component
func (m *OrgChartModel) loadOrchestratorState() {
	orchestratorFile := filepath.Join(m.workspacePath, "orchestrator", "state.json")
	data, err := os.ReadFile(orchestratorFile)
	if err != nil {
		// Orchestrator state not found, skip
		return
	}

	var orch struct {
		ID             string `json:"id"`
		Status         string `json:"status"`
		CurrentWork    string `json:"current_work"`
		ActiveSessions int    `json:"active_sessions"`
		TmuxSession    string `json:"tmux_session,omitempty"`
	}

	if err := json.Unmarshal(data, &orch); err != nil {
		return
	}

	// Check if orchestrator tmux session is actually running
	tmuxSpawned := false
	tmuxSession := orch.TmuxSession
	if tmuxSession != "" {
		// Verify tmux session exists
		cmd := exec.Command("tmux", "has-session", "-t", tmuxSession)
		if cmd.Run() == nil {
			tmuxSpawned = true
		}
	}

	// If no tmux session in state, try to find it by pattern
	if tmuxSession == "" {
		cmd := exec.Command("bash", "-c", "tmux ls 2>/dev/null | grep -E 'wildwest-orchestrator-|claude-orchestrator-' | head -1 | cut -d: -f1")
		if output, err := cmd.Output(); err == nil && len(output) > 0 {
			tmuxSession = string(output[:len(output)-1]) // trim newline
			tmuxSpawned = true
		}
	}

	// Determine emoji based on spawn status
	emoji := "⏸️"  // not spawned
	if tmuxSpawned {
		emoji = "🚀" // spawned
	}

	// Prepend orchestrator as first component
	orchestratorComp := Component{
		ID:            "orchestrator",
		Name:          "Orchestrator",
		Role:          "System",
		Emoji:         emoji,
		Description:   "Manages team spawning, monitoring, and coordination",
		Status:        orch.Status,
		StatusMessage: orch.CurrentWork,
		TmuxSpawned:   tmuxSpawned,
		TmuxSession:   tmuxSession,
	}

	m.components = append([]Component{orchestratorComp}, m.components...)
}

// updateComponentsFromSessions converts active sessions to components
func (m *OrgChartModel) updateComponentsFromSessions() {
	// Clear components
	m.components = make([]Component, 0)

	if len(m.activeSessions) == 0 {
		// No sessions yet, show empty state
		return
	}

	for _, sess := range m.activeSessions {
		comp := Component{
			ID:          sess.ID,
			Name:        sess.PersonaName,
			Role:        m.getRoleDescription(sess.PersonaType),
			Emoji:       m.getPersonaEmoji(sess.PersonaType),
			Description: m.getPersonaDescription(sess.PersonaType),
			Status:      m.mapSessionStatus(sess.Status),
			TmuxSpawned: sess.TmuxSpawned,
			TmuxSession: sess.TmuxSession,
		}

		// Use current_work from session.json if available
		if sess.CurrentWork != "" {
			comp.StatusMessage = sess.CurrentWork
		} else {
			// Fallback if worker hasn't updated current_work yet
			switch sess.Status {
			case "active":
				comp.StatusMessage = "Working on assigned tasks"
			case "idle":
				comp.StatusMessage = "Available for work"
			default:
				comp.StatusMessage = "Session inactive"
			}
		}

		m.components = append(m.components, comp)
	}

	// Sort by persona type hierarchy (Manager, Architect, QA, Engineers, Interns)
	m.sortComponentsByHierarchy()
}

// sortComponentsByHierarchy sorts components by persona hierarchy
func (m *OrgChartModel) sortComponentsByHierarchy() {
	// Define order priority
	getOrder := func(role string) int {
		switch role {
		case "Leader":
			return 0
		case "Architect":
			return 1
		case "QA":
			return 2
		case "Coder":
			return 3
		case "Support":
			return 4
		default:
			return 5
		}
	}

	// Simple bubble sort by hierarchy
	for i := 0; i < len(m.components); i++ {
		for j := i + 1; j < len(m.components); j++ {
			if getOrder(m.components[i].Role) > getOrder(m.components[j].Role) {
				m.components[i], m.components[j] = m.components[j], m.components[i]
			}
		}
	}
}

// mapSessionStatus maps session status to UI status
func (m *OrgChartModel) mapSessionStatus(sessionStatus string) string {
	switch sessionStatus {
	case "active":
		return "active"
	case "idle":
		return "idle"
	case "stopped", "failed":
		return "unavailable"
	default:
		return "idle"
	}
}

// getRoleDescription returns role description based on persona type
func (m *OrgChartModel) getRoleDescription(personaType session.SessionType) string {
	switch personaType {
	case session.SessionTypeEngineeringManager:
		return "Leader"
	case session.SessionTypeSolutionsArchitect:
		return "Architect"
	case session.SessionTypeQA:
		return "QA"
	case session.SessionTypeSoftwareEngineer:
		return "Coder"
	case session.SessionTypeIntern:
		return "Support"
	default:
		return "Unknown"
	}
}

// getPersonaEmoji returns emoji for persona type
func (m *OrgChartModel) getPersonaEmoji(personaType session.SessionType) string {
	switch personaType {
	case session.SessionTypeEngineeringManager:
		return "🎯"
	case session.SessionTypeSolutionsArchitect:
		return "🏗️"
	case session.SessionTypeSoftwareEngineer:
		return "👷"
	case session.SessionTypeIntern:
		return "📝"
	case session.SessionTypeQA:
		return "🧪"
	default:
		return "👤"
	}
}

// getPersonaDescription returns description for persona type
func (m *OrgChartModel) getPersonaDescription(personaType session.SessionType) string {
	switch personaType {
	case session.SessionTypeEngineeringManager:
		return "Oversees team, delegates tasks, reviews progress"
	case session.SessionTypeSolutionsArchitect:
		return "Designs system architecture, technical decisions"
	case session.SessionTypeQA:
		return "Tests features, ensures quality standards"
	case session.SessionTypeSoftwareEngineer:
		return "Implements features and functionality"
	case session.SessionTypeIntern:
		return "Assists with testing and documentation"
	default:
		return "Team member"
	}
}

func (m OrgChartModel) View() string {
	var b strings.Builder

	// Header with version
	title := "🚀 WildWest Team"
	if m.version != "" {
		title += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("("+m.version+")")
	}
	header := headerStyle.Render(title)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Render team list
	b.WriteString(m.renderList())

	// Show details if selected
	if m.showingDetails {
		b.WriteString("\n")
		b.WriteString(m.renderDetails())
	}

	// Render cost estimate section
	b.WriteString(m.renderCostEstimate())

	// Footer
	b.WriteString("\n")
	instructions := "↑↓/jk: navigate | d: details | a: attach | K: kill session | esc/b: back | q: quit"
	b.WriteString(footerStyle.Render(instructions))

	return b.String()
}

func (m OrgChartModel) renderList() string {
	var b strings.Builder

	if len(m.components) == 0 {
		emptyMsg := listItemStyle.Render("  No active sessions yet...")
		b.WriteString(emptyMsg)
		b.WriteString("\n")
		emptyMsg2 := listItemStyle.Render("  Waiting for orchestrator to spawn team members...")
		b.WriteString(emptyMsg2)
		b.WriteString("\n")
		return b.String()
	}

	for i, comp := range m.components {
		statusMarker := m.getStatusMarker(comp.Status)

		// Tree structure prefix
		var prefix, continuation string
		if i == len(m.components)-1 {
			prefix = "└─"
			continuation = "  "
		} else {
			prefix = "├─"
			continuation = "│ "
		}

		// Main item line with tmux spawn indicator
		var line string
		tmuxIndicator := ""
		if comp.TmuxSpawned {
			tmuxIndicator = " 🖥️"  // Spawned
		} else {
			tmuxIndicator = " ⏳"  // Not spawned yet
		}

		if i == m.selectedIndex {
			line = fmt.Sprintf("%s %s  %s (%s)%s", prefix, statusMarker, comp.Name, comp.Role, tmuxIndicator)
			b.WriteString(selectedListItemStyle.Render(line))
		} else {
			line = fmt.Sprintf("%s %s  %s (%s)%s", prefix, statusMarker, comp.Name, comp.Role, tmuxIndicator)
			b.WriteString(listItemStyle.Render(line))
		}
		b.WriteString("\n")

		// Show status message (multi-line support)
		if comp.StatusMessage != "" {
			// Split by newlines for multi-line status
			statusLines := strings.Split(comp.StatusMessage, "\n")
			for j, statusLine := range statusLines {
				if strings.TrimSpace(statusLine) == "" {
					continue
				}

				var statusPrefix string
				if j == 0 {
					statusPrefix = fmt.Sprintf("%s └─ ", continuation)
				} else {
					statusPrefix = fmt.Sprintf("%s    ", continuation)
				}

				if i == m.selectedIndex {
					b.WriteString(statusMessageStyle.Render(statusPrefix + statusLine))
				} else {
					b.WriteString(dividerStyle.Render(statusPrefix + statusLine))
				}
				b.WriteString("\n")
			}
		}

		// Vertical separator between items (except last)
		if i < len(m.components)-1 {
			b.WriteString(dividerStyle.Render("│"))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m OrgChartModel) getStatusMarker(status string) string {
	switch status {
	case "active":
		return "🔄"  // Active/working
	case "idle":
		return "✅"  // Available/ready
	case "unavailable":
		return "⏸️"  // Paused/unavailable
	default:
		return "✅"
	}
}


// killSession kills all spawned tmux sessions and deletes the session directory
func (m OrgChartModel) killSession() tea.Cmd {
	return func() tea.Msg {
		// Read orchestrator state to get list of spawned sessions
		stateFile := filepath.Join(m.workspacePath, "orchestrator", "state.json")
		data, err := os.ReadFile(stateFile)
		if err != nil {
			return tea.Quit()
		}

		var state struct {
			SpawnedSessions []string `json:"spawned_sessions"`
			TmuxSession     string   `json:"tmux_session"`
		}
		if err := json.Unmarshal(data, &state); err != nil {
			return tea.Quit()
		}

		killed := 0
		// Kill all spawned agent sessions
		for _, tmuxSession := range state.SpawnedSessions {
			cmd := exec.Command("tmux", "kill-session", "-t", tmuxSession)
			if cmd.Run() == nil {
				killed++
			}
		}

		// Kill orchestrator session
		if state.TmuxSession != "" {
			cmd := exec.Command("tmux", "kill-session", "-t", state.TmuxSession)
			cmd.Run()
		}

		// Delete the entire session directory
		if m.workspacePath != "" && m.workspacePath != "." && m.workspacePath != "/" {
			os.RemoveAll(m.workspacePath)
		}

		return tea.Quit()
	}
}

func (m OrgChartModel) renderCostEstimate() string {
	var b strings.Builder

	b.WriteString(logsBorderStyle.Render(""))
	b.WriteString("\n")

	// Get total cost from session manager
	totalCost, usageMap, err := m.sessionManager.GetTotalTeamCost()
	if err != nil || len(usageMap) == 0 {
		b.WriteString(logsHeaderStyle.Render("💰 Cost Estimate: $0.00 (no usage data yet)"))
		b.WriteString("\n")
		return b.String()
	}

	// Calculate total tokens
	var totalInputTokens, totalOutputTokens int64
	for _, usage := range usageMap {
		totalInputTokens += usage.InputTokens
		totalOutputTokens += usage.OutputTokens
	}
	totalTokens := totalInputTokens + totalOutputTokens

	// Format the cost summary
	costLine := fmt.Sprintf("💰 Cost Estimate: %s | Tokens: %s in, %s out (%s total)",
		session.FormatCost(totalCost),
		session.FormatTokens(totalInputTokens),
		session.FormatTokens(totalOutputTokens),
		session.FormatTokens(totalTokens))

	b.WriteString(logsHeaderStyle.Render(costLine))
	b.WriteString("\n")

	return b.String()
}

// generateStatusSummary creates a one-line status summary
func (m *OrgChartModel) generateStatusSummary() string {
	if len(m.components) == 0 {
		return "No active sessions"
	}

	// Group by status
	active := []string{}
	idle := []string{}
	unavailable := []string{}

	for _, comp := range m.components {
		name := comp.Name
		switch comp.Status {
		case "active":
			active = append(active, name)
		case "idle":
			idle = append(idle, name)
		case "unavailable":
			unavailable = append(unavailable, name)
		}
	}

	parts := []string{}
	if len(active) > 0 {
		parts = append(parts, strings.Join(active, ", ")+" working")
	}
	if len(idle) > 0 {
		parts = append(parts, strings.Join(idle, ", ")+" available")
	}
	if len(unavailable) > 0 {
		parts = append(parts, strings.Join(unavailable, ", ")+" unavailable")
	}

	return strings.Join(parts, "; ")
}

// generateSpawnMessage creates a message for newly spawned sessions
func (m *OrgChartModel) generateSpawnMessage(newSessions, oldSessions []*session.Session) string {
	// Find new session
	oldIDs := make(map[string]bool)
	for _, s := range oldSessions {
		oldIDs[s.ID] = true
	}

	newNames := []string{}
	for _, s := range newSessions {
		if !oldIDs[s.ID] {
			newNames = append(newNames, s.PersonaName)
		}
	}

	if len(newNames) > 0 {
		if len(newNames) == 1 {
			return fmt.Sprintf("Spawning %s", newNames[0])
		}
		return fmt.Sprintf("Spawning %s", strings.Join(newNames, ", "))
	}

	return ""
}

// captureTmuxOutput captures the last N lines from a tmux session
func (m OrgChartModel) captureTmuxOutput(tmuxSession string, lines int) string {
	cmd := exec.Command("tmux", "capture-pane",
		"-t", tmuxSession,
		"-p",
		"-S", fmt.Sprintf("-%d", lines))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Session might not exist or no output yet
		return ""
	}

	// Trim and return the output
	result := strings.TrimSpace(string(output))
	if result == "" {
		return ""
	}

	// Limit the output to prevent huge popups
	lines_slice := strings.Split(result, "\n")
	if len(lines_slice) > lines {
		lines_slice = lines_slice[len(lines_slice)-lines:]
	}

	return strings.Join(lines_slice, "\n")
}

// attachToTmux creates a command to attach to a tmux session
// It clears the screen and replaces the current process with tmux attach
func (m OrgChartModel) renderDetails() string {
	if m.selectedIndex >= len(m.components) {
		return ""
	}

	comp := m.components[m.selectedIndex]
	statusMarker := m.getStatusMarker(comp.Status)

	// Get status label
	statusLabel := ""
	switch comp.Status {
	case "active":
		statusLabel = "Working"
	case "idle":
		statusLabel = "Available"
	case "unavailable":
		statusLabel = "Unavailable"
	}

	// Build simplified details
	var detailsBuilder strings.Builder
	detailsBuilder.WriteString(fmt.Sprintf("%s\n\n", comp.Name))
	detailsBuilder.WriteString(fmt.Sprintf("Role:   %s\n", comp.Role))
	detailsBuilder.WriteString(fmt.Sprintf("Status: %s %s\n", statusMarker, statusLabel))

	detailsBuilder.WriteString(fmt.Sprintf("\nCurrent Activity:\n%s\n", comp.StatusMessage))
	detailsBuilder.WriteString(fmt.Sprintf("\nDescription:\n%s", comp.Description))

	// Add live output from tmux session (last 10 lines)
	if comp.TmuxSpawned && comp.TmuxSession != "" {
		liveOutput := m.captureTmuxOutput(comp.TmuxSession, 10)
		if liveOutput != "" {
			var liveBuilder strings.Builder
			liveBuilder.WriteString(liveOutputHeaderStyle.Render("Live Output (last 10 lines)"))
			liveBuilder.WriteString("\n\n")
			liveBuilder.WriteString(liveOutput)
			detailsBuilder.WriteString("\n\n")
			detailsBuilder.WriteString(liveOutputStyle.Render(liveBuilder.String()))
		}
	}

	return detailsStyle.Render(detailsBuilder.String())
}

// RunStaticTUI starts the static org chart TUI with orchestrator
func RunStaticTUI() error {
	return RunStaticTUIWithWorkspace(".ww-db", "")
}

// RunStaticTUIWithWorkspace starts the TUI with a specific workspace
func RunStaticTUIWithWorkspace(workspacePath, version string) error {
	// Create session manager directly (no orchestrator needed for read-only TUI)
	sm, err := session.NewSessionManager(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	// Loop to allow returning to TUI after detaching from tmux
	for {
		// Load sessions BEFORE starting TUI so they're ready immediately
		sessions, err := sm.GetActiveSessions()
		if err != nil {
			return fmt.Errorf("failed to load sessions: %w", err)
		}

		model := NewOrgChartModel(nil, sm, workspacePath, version)
		// Pre-populate with loaded sessions
		model.activeSessions = sessions
		model.updateComponentsFromSessions()
		model.loadOrchestratorState() // Add orchestrator at top level
		model.initialized = true
		model.addLog(fmt.Sprintf("Loaded %d sessions from %s", len(sessions), workspacePath))
		if len(sessions) > 0 {
			model.addLog(model.generateStatusSummary())
		}

		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
		)
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		// Check if we need to attach to a tmux session
		if m, ok := finalModel.(OrgChartModel); ok && m.attachToSession != "" {
			// Clear screen and exec into tmux
			cmd := exec.Command("bash", "-c", fmt.Sprintf("clear && tmux attach -t %s", m.attachToSession))
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				// If tmux command fails, show error and return to TUI
				fmt.Printf("Error attaching to tmux: %v\nPress Enter to return to TUI...", err)
				fmt.Scanln()
			}
			// After detaching from tmux, loop back to TUI
			continue
		}

		// User pressed 'q' to quit - exit the loop
		break
	}

	return nil
}
