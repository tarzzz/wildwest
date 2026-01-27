package cmd

import (
	"fmt"

	"github.com/tarzzz/wildwest/pkg/names"
	"github.com/spf13/cobra"
)

var namesCmd = &cobra.Command{
	Use:   "names",
	Short: "List available persona names",
	Long: `Shows the list of interesting names used for personas.
Names are drawn from scientists, artists, musicians, writers, philosophers, inventors, and explorers.`,
	RunE: listNames,
}

func init() {
	rootCmd.AddCommand(namesCmd)
}

func listNames(cmd *cobra.Command, args []string) error {
	nameList := names.GetNameList()

	fmt.Println("Available Persona Names")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Printf("\nTotal: %d names across 7 categories\n\n", names.CountTotal())

	// Display by category
	categories := []string{"Scientists", "Artists", "Musicians", "Writers", "Philosophers", "Inventors", "Explorers"}

	for _, category := range categories {
		namesList := nameList[category]
		fmt.Printf("╔══════════════════════════════════════════════════╗\n")
		fmt.Printf("║  %s (%d)\n", category, len(namesList))
		fmt.Printf("╚══════════════════════════════════════════════════╝\n")

		// Display names in columns
		for i := 0; i < len(namesList); i += 4 {
			line := "  "
			for j := 0; j < 4 && i+j < len(namesList); j++ {
				line += fmt.Sprintf("%-15s", namesList[i+j])
			}
			fmt.Println(line)
		}
		fmt.Println()
	}

	fmt.Println("Names are assigned automatically based on persona type:")
	fmt.Println("  • Managers: Philosophers")
	fmt.Println("  • Architects: Artists or Inventors")
	fmt.Println("  • Engineers: Scientists or Inventors")
	fmt.Println("  • Interns: Writers or Explorers")

	return nil
}
