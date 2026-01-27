package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	ClaudePath   string                 `yaml:"claude_path"`
	Environments map[string]Environment `yaml:"environments"`
	Templates    map[string]string      `yaml:"templates"`
}

// Environment represents a custom environment configuration
type Environment struct {
	Description  string            `yaml:"description"`
	ClaudePath   string            `yaml:"claude_path,omitempty"`
	WorkingDir   string            `yaml:"working_dir,omitempty"`
	EnvVars      map[string]string `yaml:"env_vars,omitempty"`
	DefaultSpecs []string          `yaml:"default_specs,omitempty"`
	PreCommands  []string          `yaml:"pre_commands,omitempty"`
	PostCommands []string          `yaml:"post_commands,omitempty"`
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		// Try default locations
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		possiblePaths := []string{
			filepath.Join(home, ".claude-wrapper.yaml"),
			filepath.Join(home, ".claude-wrapper.yml"),
			".claude-wrapper.yaml",
			".claude-wrapper.yml",
		}

		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}

	// If no config file found, return default config
	if path == "" {
		return defaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set default claude path if not specified
	if cfg.ClaudePath == "" {
		cfg.ClaudePath = "claude"
	}

	return &cfg, nil
}

// defaultConfig returns a default configuration
func defaultConfig() *Config {
	return &Config{
		ClaudePath:   "claude",
		Environments: make(map[string]Environment),
		Templates:    make(map[string]string),
	}
}

// GetEnvironment retrieves an environment by name
func (c *Config) GetEnvironment(name string) (*Environment, error) {
	if name == "" {
		return nil, nil
	}

	env, exists := c.Environments[name]
	if !exists {
		return nil, fmt.Errorf("environment '%s' not found", name)
	}

	return &env, nil
}
