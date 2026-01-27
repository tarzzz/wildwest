package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// TokenUsage tracks token consumption for a session
type TokenUsage struct {
	SessionID      string    `json:"session_id"`
	Model          string    `json:"model"`           // sonnet, opus, haiku
	InputTokens    int64     `json:"input_tokens"`
	OutputTokens   int64     `json:"output_tokens"`
	TotalTokens    int64     `json:"total_tokens"`
	LastUpdated    time.Time `json:"last_updated"`
	EstimatedCost  float64   `json:"estimated_cost"`  // in USD
}

// ModelPricing defines the cost per million tokens for each model
type ModelPricing struct {
	InputPer1M  float64
	OutputPer1M float64
}

// Pricing for Claude models (per million tokens)
var modelPricing = map[string]ModelPricing{
	"sonnet": {InputPer1M: 3.0, OutputPer1M: 15.0},
	"opus":   {InputPer1M: 15.0, OutputPer1M: 75.0},
	"haiku":  {InputPer1M: 0.25, OutputPer1M: 1.25},
}

// GetTokenUsage reads token usage from a session's tokens.json file
func (sm *SessionManager) GetTokenUsage(sessionID string) (*TokenUsage, error) {
	tokensPath := filepath.Join(sm.getPersonaDir(sessionID), "tokens.json")

	data, err := os.ReadFile(tokensPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new token usage if doesn't exist
			usage := &TokenUsage{
				SessionID:     sessionID,
				Model:         "sonnet", // default
				InputTokens:   0,
				OutputTokens:  0,
				TotalTokens:   0,
				LastUpdated:   time.Now(),
				EstimatedCost: 0.0,
			}
			return usage, nil
		}
		return nil, err
	}

	var usage TokenUsage
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil, err
	}

	return &usage, nil
}

// SaveTokenUsage saves token usage to disk
func (sm *SessionManager) SaveTokenUsage(usage *TokenUsage) error {
	data, err := json.MarshalIndent(usage, "", "  ")
	if err != nil {
		return err
	}

	tokensPath := filepath.Join(sm.getPersonaDir(usage.SessionID), "tokens.json")
	return os.WriteFile(tokensPath, data, 0644)
}

// UpdateTokenUsage updates token counts and recalculates cost
func (sm *SessionManager) UpdateTokenUsage(sessionID string, inputTokens, outputTokens int64) error {
	usage, err := sm.GetTokenUsage(sessionID)
	if err != nil {
		return err
	}

	usage.InputTokens = inputTokens
	usage.OutputTokens = outputTokens
	usage.TotalTokens = inputTokens + outputTokens
	usage.LastUpdated = time.Now()

	// Calculate estimated cost
	pricing, ok := modelPricing[usage.Model]
	if !ok {
		pricing = modelPricing["sonnet"] // default
	}

	inputCost := (float64(inputTokens) / 1_000_000.0) * pricing.InputPer1M
	outputCost := (float64(outputTokens) / 1_000_000.0) * pricing.OutputPer1M
	usage.EstimatedCost = inputCost + outputCost

	return sm.SaveTokenUsage(usage)
}

// ParseTokensFromTmux extracts token usage from tmux pane output
func ParseTokensFromTmux(tmuxOutput string) (inputTokens, outputTokens int64, found bool) {
	// Look for patterns like:
	// "Token usage: 12345/200000"
	// "Tokens used: 12345 input, 6789 output"
	// "<system>Token usage: 12345/200000; 187655 remaining</system>"

	// Pattern 1: Token usage: X/Y; Z remaining
	re1 := regexp.MustCompile(`Token usage:\s*(\d+)/\d+;\s*(\d+)\s*remaining`)
	if matches := re1.FindStringSubmatch(tmuxOutput); len(matches) >= 3 {
		var used, remaining int64
		fmt.Sscanf(matches[1], "%d", &used)
		fmt.Sscanf(matches[2], "%d", &remaining)

		// Approximate: assume 75% input, 25% output split
		inputTokens = int64(float64(used) * 0.75)
		outputTokens = int64(float64(used) * 0.25)
		return inputTokens, outputTokens, true
	}

	// Pattern 2: explicit input/output counts
	re2 := regexp.MustCompile(`(\d+)\s*input.*?(\d+)\s*output`)
	if matches := re2.FindStringSubmatch(strings.ToLower(tmuxOutput)); len(matches) >= 3 {
		fmt.Sscanf(matches[1], "%d", &inputTokens)
		fmt.Sscanf(matches[2], "%d", &outputTokens)
		return inputTokens, outputTokens, true
	}

	return 0, 0, false
}

// GetTotalTeamCost calculates the total cost across all active sessions
func (sm *SessionManager) GetTotalTeamCost() (float64, map[string]*TokenUsage, error) {
	sessions, err := sm.GetAllSessions()
	if err != nil {
		return 0, nil, err
	}

	totalCost := 0.0
	usageMap := make(map[string]*TokenUsage)

	for _, sess := range sessions {
		usage, err := sm.GetTokenUsage(sess.ID)
		if err != nil {
			continue // Skip sessions with no token data
		}

		totalCost += usage.EstimatedCost
		usageMap[sess.ID] = usage
	}

	return totalCost, usageMap, nil
}

// FormatCost formats a cost value as a currency string
func FormatCost(cost float64) string {
	return fmt.Sprintf("$%.4f", cost)
}

// FormatTokens formats token count with thousands separators
func FormatTokens(tokens int64) string {
	str := fmt.Sprintf("%d", tokens)
	if len(str) <= 3 {
		return str
	}

	// Add commas
	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}

	return result.String()
}
