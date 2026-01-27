package cmd

import (
	"fmt"
	"strings"

	"github.com/plotly/claude-wrapper/pkg/session"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Track team progress (Project Manager view)",
	Long: `Acts as a Project Manager to monitor and report on all team members' progress.
This is a read-only view that shows:
- What each persona is working on
- Task completion status
- Instructions assigned to each persona
- Overall project progress`,
	RunE: trackTeam,
}

func init() {
	rootCmd.AddCommand(trackCmd)
	trackCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", ".database", "workspace directory")
}

func trackTeam(cmd *cobra.Command, args []string) error {
	sm, err := session.NewSessionManager(workspaceDir)
	if err != nil {
		return err
	}

	sessions, err := sm.GetAllSessions()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No team sessions found in workspace")
		return nil
	}

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("           PROJECT STATUS DASHBOARD")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()

	// Group by persona type
	personaMap := make(map[session.SessionType][]*session.Session)
	for _, sess := range sessions {
		personaMap[sess.PersonaType] = append(personaMap[sess.PersonaType], sess)
	}

	// Display in hierarchy order
	displayPersonaGroup(sm, session.SessionTypeEngineeringManager, "ENGINEERING MANAGER", personaMap)
	displayPersonaGroup(sm, session.SessionTypeSolutionsArchitect, "SOLUTIONS ARCHITECT", personaMap)
	displayPersonaGroup(sm, session.SessionTypeSoftwareEngineer, "SOFTWARE ENGINEERS", personaMap)
	displayPersonaGroup(sm, session.SessionTypeIntern, "INTERNS", personaMap)

	// Overall summary
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("                OVERALL SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════")

	totalTasks := 0
	completedTasks := 0
	inProgressTasks := 0

	for _, sess := range sessions {
		tasks, _ := sm.ReadTasks(sess.ID)
		completed := strings.Count(tasks, "completed")
		inProgress := strings.Count(tasks, "in progress")
		totalTasks += strings.Count(tasks, "## Task:")
		completedTasks += completed
		inProgressTasks += inProgress
	}

	fmt.Printf("\nTotal Team Members: %d\n", len(sessions))
	fmt.Printf("Total Tasks: %d\n", totalTasks)
	fmt.Printf("Completed: %d\n", completedTasks)
	fmt.Printf("In Progress: %d\n", inProgressTasks)
	fmt.Printf("Not Started: %d\n", totalTasks-completedTasks-inProgressTasks)

	if totalTasks > 0 {
		completion := float64(completedTasks) / float64(totalTasks) * 100
		fmt.Printf("\nOverall Completion: %.1f%%\n", completion)
		showProgressBar(int(completion))
	}

	return nil
}

func displayPersonaGroup(sm *session.SessionManager, personaType session.SessionType, title string, personaMap map[session.SessionType][]*session.Session) {
	sessions := personaMap[personaType]
	if len(sessions) == 0 {
		return
	}

	fmt.Printf("\n╔══════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %s\n", title)
	fmt.Printf("╚══════════════════════════════════════════════════╝\n\n")

	for _, sess := range sessions {
		fmt.Printf("📋 %s (%s)\n", sess.PersonaName, sess.ID)
		fmt.Printf("   Status: %s\n", sess.Status)
		fmt.Printf("   Started: %s\n", sess.StartTime.Format("2006-01-02 15:04:05"))

		// Read and display tasks
		tasks, err := sm.ReadTasks(sess.ID)
		if err == nil && tasks != "" {
			fmt.Println("\n   Current Tasks:")
			displayTaskSummary(tasks)
		}

		// Read and display recent instructions
		instructions, err := sm.ReadInstructions(sess.ID)
		if err == nil && instructions != "" {
			fmt.Println("\n   Latest Instructions:")
			displayLatestInstructions(instructions)
		}

		// List output files
		files, err := sm.ListPersonaFiles(sess.ID)
		if err == nil && len(files) > 0 {
			fmt.Println("\n   Output Files:")
			for _, file := range files {
				if file != "tasks.md" && file != "instructions.md" {
					fmt.Printf("      • %s\n", file)
				}
			}
		}

		fmt.Println()
	}
}

func displayTaskSummary(tasks string) {
	lines := strings.Split(tasks, "\n")
	currentTask := ""
	currentStatus := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## Task:") {
			if currentTask != "" {
				displayTask(currentTask, currentStatus)
			}
			currentTask = strings.TrimPrefix(line, "## Task:")
			currentTask = strings.TrimSpace(currentTask)
			currentStatus = ""
		} else if strings.HasPrefix(line, "- **Status**:") {
			currentStatus = strings.TrimPrefix(line, "- **Status**:")
			currentStatus = strings.TrimSpace(currentStatus)
		}
	}

	if currentTask != "" {
		displayTask(currentTask, currentStatus)
	}
}

func displayTask(task string, status string) {
	var icon string
	switch status {
	case "completed":
		icon = "✅"
	case "in progress":
		icon = "🔄"
	case "not started":
		icon = "⏸️"
	default:
		icon = "❓"
	}

	// Truncate long task descriptions
	if len(task) > 60 {
		task = task[:57] + "..."
	}

	fmt.Printf("      %s %s [%s]\n", icon, task, status)
}

func displayLatestInstructions(instructions string) {
	lines := strings.Split(instructions, "\n")
	lastSection := ""
	lineCount := 0

	// Get the last instruction section (up to 5 lines)
	for i := len(lines) - 1; i >= 0 && lineCount < 5; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			lastSection = line + "\n" + lastSection
			lineCount++
		}
		if strings.HasPrefix(line, "## Instructions from") {
			break
		}
	}

	if lastSection != "" {
		// Indent for display
		indented := strings.ReplaceAll(lastSection, "\n", "\n      ")
		fmt.Printf("      %s\n", strings.TrimSpace(indented))
	}
}

func showProgressBar(percentage int) {
	const barWidth = 40
	filled := (percentage * barWidth) / 100
	empty := barWidth - filled

	bar := "["
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	bar += "]"

	fmt.Printf("\n%s %d%%\n", bar, percentage)
}
