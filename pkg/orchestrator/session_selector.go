package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tarzzz/wildwest/pkg/session"
)

// SessionSelectorModel is the TUI for selecting a session
type SessionSelectorModel struct {
	sessions      []session.SessionMetadata
	cursor        int
	baseWorkspace string
	version       string
	err           error
	selected      bool
}

func initialSessionSelector(baseWorkspace, version string) SessionSelectorModel {
	sessions, err := session.ListSessions(baseWorkspace)
	return SessionSelectorModel{
		sessions:      sessions,
		cursor:        0,
		baseWorkspace: baseWorkspace,
		version:       version,
		err:           err,
	}
}

func (m SessionSelectorModel) Init() tea.Cmd {
	return nil
}

func (m SessionSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SessionSelectorModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error loading sessions: %v\n", m.err)
	}

	var b strings.Builder

	// Header
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	title := "🚀 WildWest Sessions"
	if m.version != "" {
		title += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("("+m.version+")")
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Instructions
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)
	b.WriteString(helpStyle.Render("Select a session to monitor (↑/↓ or j/k to navigate, enter to select, q to quit)"))
	b.WriteString("\n\n")

	// Session list
	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Padding(0, 1)

	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	for i, sess := range m.sessions {
		var line strings.Builder

		// Cursor
		if i == m.cursor {
			line.WriteString("▶ ")
		} else {
			line.WriteString("  ")
		}

		// Session info
		timeStr := sess.CreatedAt.Format("2006-01-02 15:04")
		sessionInfo := fmt.Sprintf("[%s] %s %s",
			sess.ID,
			sess.Description,
			dateStyle.Render("("+timeStr+")"),
		)

		// Apply style
		if i == m.cursor {
			line.WriteString(selectedStyle.Render(sessionInfo))
		} else {
			line.WriteString(normalStyle.Render(sessionInfo))
		}

		b.WriteString(line.String())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("Found %d session(s)", len(m.sessions))))

	return b.String()
}

// RunSessionSelector shows a TUI for selecting a session
func RunSessionSelector(baseWorkspace, version string) error {
	for {
		m := initialSessionSelector(baseWorkspace, version)

		p := tea.NewProgram(m, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		// If user selected a session, launch the org chart TUI
		if finalM, ok := finalModel.(SessionSelectorModel); ok {
			if finalM.selected && len(finalM.sessions) > 0 {
				selectedSession := finalM.sessions[finalM.cursor]
				shouldGoBack, err := runSessionWithBackNavigation(selectedSession.WorkspacePath, version)
				if err != nil {
					return err
				}
				// If user pressed back, loop to show selector again
				if shouldGoBack {
					continue
				}
			}
		}

		// User quit without selecting
		return nil
	}
}

// runSessionWithBackNavigation runs the org chart TUI and returns true if user wants to go back
func runSessionWithBackNavigation(workspacePath, version string) (bool, error) {
	// Create session manager
	sm, err := session.NewSessionManager(workspacePath)
	if err != nil {
		return false, fmt.Errorf("failed to create session manager: %w", err)
	}

	// Loop to allow returning to TUI after detaching from tmux
	for {
		// Load sessions
		sessions, err := sm.GetActiveSessions()
		if err != nil {
			return false, fmt.Errorf("failed to load sessions: %w", err)
		}

		model := NewOrgChartModel(nil, sm, workspacePath, version)
		model.activeSessions = sessions
		model.updateComponentsFromSessions()
		model.loadOrchestratorState()
		model.initialized = true
		model.addLog(fmt.Sprintf("Loaded %d sessions from %s", len(sessions), workspacePath))
		if len(sessions) > 0 {
			model.addLog(model.generateStatusSummary())
		}

		p := tea.NewProgram(model, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return false, err
		}

		if m, ok := finalModel.(OrgChartModel); ok {
			// Check if user wants to go back to session selector
			if m.goBack {
				return true, nil
			}

			// Check if we need to attach to a tmux session
			if m.attachToSession != "" {
				cmd := exec.Command("bash", "-c", fmt.Sprintf("clear && tmux attach -t %s", m.attachToSession))
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					fmt.Printf("Error attaching to tmux: %v\nPress Enter to return to TUI...", err)
					fmt.Scanln()
				}
				// After detaching from tmux, loop back to TUI
				continue
			}
		}

		// User quit normally
		return false, nil
	}
}
