package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plotly/claude-wrapper/pkg/persona"
	"github.com/spf13/cobra"
)

var personaCmd = &cobra.Command{
	Use:   "persona",
	Short: "Manage personas for Claude interactions",
	Long: `Personas define role-based configurations for Claude Code execution.
Each persona has specific instructions, capabilities, and constraints.`,
}

var personaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available personas",
	RunE:  listPersonas,
}

var personaShowCmd = &cobra.Command{
	Use:   "show [persona-name]",
	Short: "Show details of a specific persona",
	Args:  cobra.ExactArgs(1),
	RunE:  showPersona,
}

var personaInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default personas file",
	RunE:  initPersonas,
}

func init() {
	rootCmd.AddCommand(personaCmd)
	personaCmd.AddCommand(personaListCmd)
	personaCmd.AddCommand(personaShowCmd)
	personaCmd.AddCommand(personaInitCmd)
}

func listPersonas(cmd *cobra.Command, args []string) error {
	personas, err := persona.LoadPersonas("")
	if err != nil {
		return fmt.Errorf("failed to load personas: %w", err)
	}

	fmt.Println("Available Personas:")
	fmt.Println("==================")
	fmt.Println()

	for key, p := range personas.Personas {
		fmt.Printf("%s (%s)\n", p.Name, key)
		fmt.Printf("  Description: %s\n", p.Description)
		fmt.Println()
	}

	return nil
}

func showPersona(cmd *cobra.Command, args []string) error {
	personaName := args[0]

	personas, err := persona.LoadPersonas("")
	if err != nil {
		return fmt.Errorf("failed to load personas: %w", err)
	}

	p, err := personas.GetPersona(personaName)
	if err != nil {
		return err
	}

	fmt.Printf("Persona: %s\n", p.Name)
	fmt.Println("=" + string(make([]byte, len(p.Name)+9)))
	fmt.Println()

	fmt.Printf("Description: %s\n\n", p.Description)

	fmt.Println("Instructions:")
	fmt.Println("-------------")
	fmt.Println(p.Instructions)
	fmt.Println()

	if len(p.Capabilities) > 0 {
		fmt.Println("Capabilities:")
		fmt.Println("-------------")
		for _, cap := range p.Capabilities {
			fmt.Printf("  - %s\n", cap)
		}
		fmt.Println()
	}

	if len(p.Constraints) > 0 {
		fmt.Println("Constraints:")
		fmt.Println("------------")
		for _, constraint := range p.Constraints {
			fmt.Printf("  - %s\n", constraint)
		}
		fmt.Println()
	}

	return nil
}

func initPersonas(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	path := filepath.Join(home, ".claude-personas.yaml")

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("personas file already exists at %s", path)
	}

	if err := persona.SaveDefaultPersonas(path); err != nil {
		return err
	}

	fmt.Printf("Default personas initialized at: %s\n", path)
	fmt.Println("\nYou can now customize the personas by editing this file.")

	return nil
}
