package claude

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tarzzz/wildwest/pkg/config"
)

// GetClaudeBinary returns the path to the claude binary
// It checks CLAUDE_BIN environment variable first, then falls back to "claude"
func GetClaudeBinary() string {
	if claudeBin := os.Getenv("CLAUDE_BIN"); claudeBin != "" {
		return claudeBin
	}
	return "claude"
}

// ExecutorOptions contains options for executing Claude
type ExecutorOptions struct {
	Prompt              string
	Environment         string
	Instructions        string
	PersonaInstructions string
	ExpandPrompt        bool
	CustomSpecs         []string
	Verbose             bool
}

// Executor handles Claude Code execution
type Executor struct {
	config *config.Config
}

// NewExecutor creates a new Executor
func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{
		config: cfg,
	}
}

// Run executes Claude Code with the given options
func (e *Executor) Run(opts ExecutorOptions) error {
	// Check CLAUDE_BIN environment variable first
	claudePath := GetClaudeBinary()

	// Then check config
	if e.config.ClaudePath != "" {
		claudePath = e.config.ClaudePath
	}

	var env *config.Environment

	// Load environment if specified
	if opts.Environment != "" {
		var err error
		env, err = e.config.GetEnvironment(opts.Environment)
		if err != nil {
			return err
		}

		if env.ClaudePath != "" {
			claudePath = env.ClaudePath
		}
	}

	// Build the prompt
	prompt := opts.Prompt
	if opts.ExpandPrompt {
		prompt = e.expandPromptWithClaude(opts)
	}

	// Build command arguments
	args := []string{}

	// Add persona instructions (takes precedence)
	if opts.PersonaInstructions != "" {
		// Create a temporary file for persona instructions
		tmpFile, err := os.CreateTemp("", "claude-persona-*.md")
		if err != nil {
			return fmt.Errorf("failed to create temp file for persona: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(opts.PersonaInstructions); err != nil {
			return fmt.Errorf("failed to write persona instructions: %w", err)
		}
		tmpFile.Close()

		args = append(args, "--instructions", tmpFile.Name())
	} else if opts.Instructions != "" {
		// Add custom instructions if no persona
		args = append(args, "--instructions", opts.Instructions)
	}

	// Add custom specs
	if env != nil && len(env.DefaultSpecs) > 0 {
		for _, spec := range env.DefaultSpecs {
			args = append(args, "--spec", spec)
		}
	}

	for _, spec := range opts.CustomSpecs {
		args = append(args, "--spec", spec)
	}

	// Add the prompt
	args = append(args, prompt)

	if opts.Verbose {
		fmt.Printf("Executing: %s %s\n", claudePath, strings.Join(args, " "))
	}

	// Execute pre-commands if any
	if env != nil && len(env.PreCommands) > 0 {
		if err := e.executeCommands(env.PreCommands, env, opts.Verbose); err != nil {
			return fmt.Errorf("pre-command failed: %w", err)
		}
	}

	// Create and execute command
	cmd := exec.Command(claudePath, args...)

	// Set working directory if specified
	if env != nil && env.WorkingDir != "" {
		cmd.Dir = env.WorkingDir
	}

	// Set environment variables
	cmd.Env = os.Environ()
	if env != nil {
		for key, value := range env.EnvVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude execution failed: %w", err)
	}

	// Execute post-commands if any
	if env != nil && len(env.PostCommands) > 0 {
		if err := e.executeCommands(env.PostCommands, env, opts.Verbose); err != nil {
			return fmt.Errorf("post-command failed: %w", err)
		}
	}

	return nil
}

// Expand expands a minimal prompt into detailed instructions
func (e *Executor) Expand(opts ExecutorOptions) error {
	expandPrompt := fmt.Sprintf(`You are a technical instruction expander. Take the following minimal prompt and expand it into detailed, actionable instructions:

Minimal Prompt: %s

Please provide:
1. Detailed step-by-step instructions
2. Any assumptions or prerequisites
3. Expected outcomes
4. Potential challenges or considerations

Format the response as a clear, structured set of instructions.`, opts.Prompt)

	// Check CLAUDE_BIN environment variable first
	claudePath := GetClaudeBinary()

	// Then check config
	if e.config.ClaudePath != "" {
		claudePath = e.config.ClaudePath
	}

	if opts.Environment != "" {
		env, err := e.config.GetEnvironment(opts.Environment)
		if err != nil {
			return err
		}
		if env.ClaudePath != "" {
			claudePath = env.ClaudePath
		}
	}

	args := []string{expandPrompt}

	if opts.Verbose {
		fmt.Printf("Expanding prompt: %s\n", opts.Prompt)
	}

	cmd := exec.Command(claudePath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// expandPromptWithClaude uses Claude to expand a minimal prompt
func (e *Executor) expandPromptWithClaude(opts ExecutorOptions) string {
	// This is a simplified version - in production you'd want to capture output
	return fmt.Sprintf("Expand and execute: %s", opts.Prompt)
}

// executeCommands executes a list of shell commands
func (e *Executor) executeCommands(commands []string, env *config.Environment, verbose bool) error {
	for _, cmdStr := range commands {
		if verbose {
			fmt.Printf("Executing: %s\n", cmdStr)
		}

		cmd := exec.Command("sh", "-c", cmdStr)

		if env != nil && env.WorkingDir != "" {
			cmd.Dir = env.WorkingDir
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command '%s' failed: %w", cmdStr, err)
		}
	}
	return nil
}
