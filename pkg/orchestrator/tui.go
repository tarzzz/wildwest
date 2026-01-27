package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tarzzz/wildwest/pkg/session"
)

// TickMsg is sent every 2 seconds to refresh session data
type TickMsg time.Time

// AttachCompleteMsg is sent when returning from tmux attach
type AttachCompleteMsg struct{}

// MouseZone represents a clickable area for a persona
type MouseZone struct {
	SessionID string
	X         int
	Y         int
	Width     int
	Height    int
}

// OrchestratorModel is the Bubble Tea model for the TUI
type OrchestratorModel struct {
	sessions       []*session.Session
	sessionManager *session.SessionManager
	workspace      string
	mouseZones     []MouseZone
	lastUpdate     time.Time
	quitting       bool
	width          int
	height         int
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")). // Blue
			Align(lipgloss.Center).
			Padding(1, 0)

	activeStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("42")). // Green
			Padding(0, 1).
			Width(35)

	idleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("226")). // Yellow
			Padding(0, 1).
			Width(35)

	stoppedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")). // Gray
			Padding(0, 1).
			Width(35)

	completedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("46")). // Bright green
			Padding(0, 1).
			Width(35)

	failedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")). // Red
			Padding(0, 1).
			Width(35)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // Gray
			Padding(1, 0).
			Align(lipgloss.Center)

	connectorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")) // Gray
)

// NewOrchestratorModel creates a new TUI model
func NewOrchestratorModel(sm *session.SessionManager, workspace string) OrchestratorModel {
	return OrchestratorModel{
		sessionManager: sm,
		workspace:      workspace,
		mouseZones:     make([]MouseZone, 0),
		sessions:       make([]*session.Session, 0),
		lastUpdate:     time.Now(),
	}
}

// Init initializes the model
func (m OrchestratorModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
	)
}

// Update handles events
func (m OrchestratorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft {
			// Check if click is within any persona zone
			for _, zone := range m.mouseZones {
				if msg.X >= zone.X && msg.X <= zone.X+zone.Width &&
					msg.Y >= zone.Y && msg.Y <= zone.Y+zone.Height {
					return m, attachToTmux(zone.SessionID)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		// Refresh session data
		m.refreshSessions()
		return m, tickCmd()

	case AttachCompleteMsg:
		// Returned from tmux attach, refresh and continue
		m.refreshSessions()
		return m, tickCmd()
	}

	return m, nil
}

// View renders the UI
func (m OrchestratorModel) View() string {
	if m.quitting {
		return "Orchestrator stopped.\n"
	}

	// Clear mouse zones for this render
	m.mouseZones = make([]MouseZone, 0)

	var b strings.Builder
	currentY := 0

	// Header
	header := headerStyle.Render("🚀 WildWest Team Orchestrator")
	b.WriteString(header)
	b.WriteString("\n\n")
	currentY += 3

	if len(m.sessions) == 0 {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  Waiting for team members..."))
		b.WriteString("\n\n")
		b.WriteString(footerStyle.Render("Press 'q' to quit"))
		return b.String()
	}

	// Manager (level 1)
	manager := m.getSessionByType(session.SessionTypeEngineeringManager)
	if manager != nil {
		box, height := m.renderPersonaBox(manager, currentY)
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(box))
		b.WriteString("\n")
		currentY += height + 1

		// Connector down
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(connectorStyle.Render("        │")))
		b.WriteString("\n")
		currentY++
	}

	// Architect (level 2) + QA (cross-functional)
	architect := m.getSessionByType(session.SessionTypeSolutionsArchitect)
	qaList := m.getSessionsByType(session.SessionTypeQA)

	if architect != nil || len(qaList) > 0 {
		level2, height := m.renderLevel2(architect, qaList, currentY)
		b.WriteString(level2)
		currentY += height
	}

	// Engineers (level 3)
	engineers := m.getSessionsByType(session.SessionTypeSoftwareEngineer)
	if len(engineers) > 0 {
		engsView, height := m.renderEngineers(engineers, currentY)
		b.WriteString(engsView)
		currentY += height
	}

	// Interns (level 4)
	interns := m.getSessionsByType(session.SessionTypeIntern)
	if len(interns) > 0 {
		internsView, height := m.renderInterns(interns, currentY)
		b.WriteString(internsView)
		currentY += height
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("Press 'q' to quit | Click on any persona to attach to their tmux session"))
	b.WriteString("\n")

	return b.String()
}

// refreshSessions updates the session list from the workspace
func (m *OrchestratorModel) refreshSessions() {
	sessions, err := m.sessionManager.GetActiveSessions()
	if err == nil {
		m.sessions = sessions
		m.lastUpdate = time.Now()
	}
}

// getSessionByType returns the first session of a given type
func (m *OrchestratorModel) getSessionByType(sessionType session.SessionType) *session.Session {
	for _, sess := range m.sessions {
		if sess.PersonaType == sessionType {
			return sess
		}
	}
	return nil
}

// getSessionsByType returns all sessions of a given type
func (m *OrchestratorModel) getSessionsByType(sessionType session.SessionType) []*session.Session {
	result := make([]*session.Session, 0)
	for _, sess := range m.sessions {
		if sess.PersonaType == sessionType {
			result = append(result, sess)
		}
	}
	return result
}

// renderPersonaBox renders a single persona box
func (m *OrchestratorModel) renderPersonaBox(sess *session.Session, currentY int) (string, int) {
	// Get emoji based on persona type
	emoji := m.getPersonaEmoji(sess.PersonaType)

	// Get status indicator
	statusEmoji := m.getStatusEmoji(sess.Status)

	// Get current work
	currentWork := m.sessionManager.GetCurrentWork(sess.ID)
	if currentWork == "" {
		currentWork = "Idle"
	}
	if len(currentWork) > 30 {
		currentWork = currentWork[:27] + "..."
	}

	// Build content
	content := fmt.Sprintf("%s %s (%s)\n", emoji, sess.PersonaType, sess.PersonaName)
	content += fmt.Sprintf("Status: %s %s\n", statusEmoji, sess.Status)
	content += fmt.Sprintf("Work: %s", currentWork)

	// Choose style based on status
	var style lipgloss.Style
	switch sess.Status {
	case "active":
		style = activeStyle
	case "completed":
		style = completedStyle
	case "failed":
		style = failedStyle
	case "stopped":
		style = stoppedStyle
	default:
		style = idleStyle
	}

	rendered := style.Render(content)

	// Track mouse zone (approximate - centered)
	centerX := (m.width - 37) / 2
	m.mouseZones = append(m.mouseZones, MouseZone{
		SessionID: sess.ID,
		X:         centerX,
		Y:         currentY,
		Width:     37,
		Height:    5,
	})

	return rendered, 5
}

// renderLevel2 renders architect and QA side by side
func (m *OrchestratorModel) renderLevel2(architect *session.Session, qaList []*session.Session, currentY int) (string, int) {
	var b strings.Builder

	if architect == nil && len(qaList) == 0 {
		return "", 0
	}

	// Render architect
	var architectBox string
	if architect != nil {
		box, _ := m.renderPersonaBox(architect, currentY)
		architectBox = box
	} else {
		architectBox = lipgloss.NewStyle().Width(37).Height(5).Render("")
	}

	// Render QA (show first one if multiple)
	var qaBox string
	if len(qaList) > 0 {
		box, _ := m.renderPersonaBox(qaList[0], currentY)
		qaBox = box
	} else {
		qaBox = lipgloss.NewStyle().Width(37).Height(5).Render("")
	}

	// Join side by side
	combined := lipgloss.JoinHorizontal(lipgloss.Top, architectBox, "  ", qaBox)
	b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(combined))
	b.WriteString("\n")

	// Connector down from architect
	if architect != nil {
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(connectorStyle.Render("        │")))
		b.WriteString("\n")
	}

	return b.String(), 7
}

// renderEngineers renders multiple engineers horizontally
func (m *OrchestratorModel) renderEngineers(engineers []*session.Session, currentY int) (string, int) {
	if len(engineers) == 0 {
		return "", 0
	}

	var b strings.Builder
	startY := currentY

	// Render up to 3 engineers per row
	for i := 0; i < len(engineers); i += 3 {
		end := i + 3
		if end > len(engineers) {
			end = len(engineers)
		}

		rowEngineers := engineers[i:end]
		boxes := make([]string, len(rowEngineers))

		for j, eng := range rowEngineers {
			box, _ := m.renderPersonaBox(eng, currentY)
			boxes[j] = box
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, boxes...)
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(row))
		b.WriteString("\n")
		currentY += 6
	}

	// Connector down to interns
	b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(connectorStyle.Render("        │")))
	b.WriteString("\n")
	currentY++

	return b.String(), currentY - startY
}

// renderInterns renders interns
func (m *OrchestratorModel) renderInterns(interns []*session.Session, currentY int) (string, int) {
	if len(interns) == 0 {
		return "", 0
	}

	var b strings.Builder

	// Render interns (up to 3 per row)
	for i := 0; i < len(interns); i += 3 {
		end := i + 3
		if end > len(interns) {
			end = len(interns)
		}

		rowInterns := interns[i:end]
		boxes := make([]string, len(rowInterns))

		for j, intern := range rowInterns {
			box, _ := m.renderPersonaBox(intern, currentY)
			boxes[j] = box
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, boxes...)
		b.WriteString(lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(row))
		b.WriteString("\n")
		currentY += 6
	}

	return b.String(), len(interns)/3*6
}

// getPersonaEmoji returns emoji for persona type
func (m *OrchestratorModel) getPersonaEmoji(personaType session.SessionType) string {
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

// getStatusEmoji returns emoji for status
func (m *OrchestratorModel) getStatusEmoji(status string) string {
	switch status {
	case "active":
		return "🟢"
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "stopped":
		return "⏸️"
	default:
		return "🟡"
	}
}

// tickCmd returns a tick command that fires every 2 seconds
func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// attachToTmux attaches to a tmux session
func attachToTmux(sessionID string) tea.Cmd {
	return func() tea.Msg {
		// Run tmux attach in a blocking way
		// The Bubble Tea program will be suspended automatically
		cmd := exec.Command("tmux", "attach", "-t", fmt.Sprintf("claude-%s", sessionID))
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		_ = cmd.Run()

		// Return message to refresh after returning from tmux
		return AttachCompleteMsg{}
	}
}
