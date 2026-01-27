package cmd

import (
	"fmt"

	"github.com/tarzzz/wildwest/pkg/claude"
	"github.com/tarzzz/wildwest/pkg/config"
	"github.com/tarzzz/wildwest/pkg/persona"
	"github.com/spf13/cobra"
)

var (
	envName       string
	personaName   string
	instructions  string
	shouldExpand  bool
	customSpecs   []string
)

var runCmd = &cobra.Command{
	Use:   "run [prompt]",
	Short: "Run Claude Code with custom environment and specs",
	Long: `Run Claude Code in a specified environment with custom specifications.
The prompt can be minimal and will be expanded if --expand is used.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runClaude,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&envName, "env", "e", "", "environment name from config")
	runCmd.Flags().StringVarP(&personaName, "persona", "p", "", "persona to use (engineering-manager, software-engineer, intern, solutions-architect)")
	runCmd.Flags().StringVarP(&instructions, "instructions", "i", "", "custom instructions file path")
	runCmd.Flags().BoolVar(&shouldExpand, "expand", false, "expand minimal prompt to detailed instructions")
	runCmd.Flags().StringSliceVarP(&customSpecs, "spec", "s", []string{}, "custom specifications (can be used multiple times)")
}

func runClaude(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load persona if specified
	var personaInstructions string
	if personaName != "" {
		personas, err := persona.LoadPersonas("")
		if err != nil {
			return fmt.Errorf("failed to load personas: %w", err)
		}

		p, err := personas.GetPersona(personaName)
		if err != nil {
			return err
		}

		personaInstructions = p.FormatInstructions(prompt)
		if verbose {
			fmt.Printf("Using persona: %s\n", p.Name)
		}
	}

	// Create executor options
	opts := claude.ExecutorOptions{
		Prompt:              prompt,
		Environment:         envName,
		Instructions:        instructions,
		PersonaInstructions: personaInstructions,
		ExpandPrompt:        shouldExpand,
		CustomSpecs:         customSpecs,
		Verbose:             verbose,
	}

	// Create and run executor
	executor := claude.NewExecutor(cfg)
	return executor.Run(opts)
}
