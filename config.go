package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// ButtonConfig represents a clickable button in Discord Rich Presence.
type ButtonConfig struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Config represents the user configuration for Discord Rich Presence display.
type Config struct {
	Show    ShowConfig     `json:"show"`
	Display DisplayConfig  `json:"display"`
	Buttons []ButtonConfig `json:"buttons"`
}

// ShowConfig controls which fields are visible in the Discord presence.
type ShowConfig struct {
	ProjectName   *bool `json:"project_name"`
	GitBranch     *bool `json:"git_branch"`
	ModelName     *bool `json:"model_name"`
	Tokens        *bool `json:"tokens"`
	SplitTokens   *bool `json:"split_tokens"`
	Cost          *bool `json:"cost"`
	CostInDetails *bool `json:"cost_in_details"`
	Duration      *bool `json:"duration"`
}

// DisplayConfig controls formatting and customization of the presence.
type DisplayConfig struct {
	DetailsPrefix string `json:"details_prefix"`
	DetailsFormat string `json:"details_format"`
	StateFormat   string `json:"state_format"`
	Separator     string `json:"separator"`
	CostPrecision *int   `json:"cost_precision"`
	IdleTimeout   *int   `json:"idle_timeout"`
	LargeImage    string `json:"large_image"`
	LargeText     string `json:"large_text"`
	DiscordAppID  string `json:"discord_app_id"`
}

var (
	configFilePath string
	currentConfig  *Config
)

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }

// showField safely dereferences a *bool, defaulting to true if nil.
func showField(ptr *bool) bool {
	if ptr == nil {
		return true
	}
	return *ptr
}

// showFieldDefault safely dereferences a *bool with a custom default for nil.
func showFieldDefault(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// DefaultConfig returns a config with all fields set to their default values.
func DefaultConfig() *Config {
	return &Config{
		Show: ShowConfig{
			ProjectName:   boolPtr(true),
			GitBranch:     boolPtr(true),
			ModelName:     boolPtr(true),
			Tokens:        boolPtr(true),
			SplitTokens:   boolPtr(false),
			Cost:          boolPtr(true),
			CostInDetails: boolPtr(false),
			Duration:      boolPtr(true),
		},
		Display: DisplayConfig{
			DetailsPrefix: "Working on",
			Separator:     " | ",
			CostPrecision: intPtr(4),
			IdleTimeout:   intPtr(0),
			LargeText:     "Clawd Code - Discord Rich Presence for Claude Code",
		},
	}
}

// LoadConfig reads the config file and merges with defaults.
// Returns DefaultConfig() if the file does not exist or cannot be parsed.
func LoadConfig() *Config {
	defaults := DefaultConfig()

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return defaults
	}

	var userConfig Config
	if err := json.Unmarshal(data, &userConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: invalid config file %s: %v (using defaults)\n", configFilePath, err)
		return defaults
	}

	return mergeConfig(defaults, &userConfig)
}

// mergeConfig applies user-provided values on top of defaults.
func mergeConfig(defaults, user *Config) *Config {
	result := *defaults

	// Show fields
	if user.Show.ProjectName != nil {
		result.Show.ProjectName = user.Show.ProjectName
	}
	if user.Show.GitBranch != nil {
		result.Show.GitBranch = user.Show.GitBranch
	}
	if user.Show.ModelName != nil {
		result.Show.ModelName = user.Show.ModelName
	}
	if user.Show.Tokens != nil {
		result.Show.Tokens = user.Show.Tokens
	}
	if user.Show.SplitTokens != nil {
		result.Show.SplitTokens = user.Show.SplitTokens
	}
	if user.Show.Cost != nil {
		result.Show.Cost = user.Show.Cost
	}
	if user.Show.CostInDetails != nil {
		result.Show.CostInDetails = user.Show.CostInDetails
	}
	if user.Show.Duration != nil {
		result.Show.Duration = user.Show.Duration
	}

	// Display fields
	if user.Display.DetailsPrefix != "" {
		result.Display.DetailsPrefix = user.Display.DetailsPrefix
	}
	if user.Display.DetailsFormat != "" {
		result.Display.DetailsFormat = user.Display.DetailsFormat
	}
	if user.Display.StateFormat != "" {
		result.Display.StateFormat = user.Display.StateFormat
	}
	if user.Display.Separator != "" {
		result.Display.Separator = user.Display.Separator
	}
	if user.Display.CostPrecision != nil {
		p := *user.Display.CostPrecision
		if p < 0 {
			p = 0
		}
		if p > 10 {
			p = 10
		}
		result.Display.CostPrecision = intPtr(p)
	}
	if user.Display.IdleTimeout != nil {
		t := *user.Display.IdleTimeout
		if t < 0 {
			t = 0
		}
		if t > 3600 {
			t = 3600
		}
		result.Display.IdleTimeout = intPtr(t)
	}
	if user.Display.LargeImage != "" {
		result.Display.LargeImage = user.Display.LargeImage
	}
	if user.Display.LargeText != "" {
		result.Display.LargeText = user.Display.LargeText
	}
	if user.Display.DiscordAppID != "" {
		result.Display.DiscordAppID = user.Display.DiscordAppID
	}

	// Buttons
	if len(user.Buttons) > 0 {
		result.Buttons = validateButtons(user.Buttons)
	}

	return &result
}

// validateButtons filters and validates button configs.
func validateButtons(buttons []ButtonConfig) []ButtonConfig {
	var valid []ButtonConfig
	for _, b := range buttons {
		if b.Label == "" || b.URL == "" {
			continue
		}
		if len(b.Label) > 32 {
			b.Label = b.Label[:32]
		}
		if !strings.HasPrefix(b.URL, "http://") && !strings.HasPrefix(b.URL, "https://") {
			continue
		}
		valid = append(valid, b)
		if len(valid) >= 2 {
			break
		}
	}
	return valid
}

// buildDetailsLine constructs the Details field for Discord presence.
func buildDetailsLine(session *SessionData, cfg *Config) string {
	showProject := showField(cfg.Show.ProjectName)
	showBranch := showField(cfg.Show.GitBranch) && session.GitBranch != ""

	if !showProject && !showBranch {
		return ""
	}

	prefix := cfg.Display.DetailsPrefix
	var base string

	if showProject && showBranch {
		base = fmt.Sprintf("%s: %s (%s)", prefix, session.ProjectName, session.GitBranch)
	} else if showProject {
		base = fmt.Sprintf("%s: %s", prefix, session.ProjectName)
	} else {
		base = fmt.Sprintf("%s: %s", prefix, session.GitBranch)
	}

	if showField(cfg.Show.Cost) && showField(cfg.Show.CostInDetails) {
		precision := 4
		if cfg.Display.CostPrecision != nil {
			precision = *cfg.Display.CostPrecision
		}
		base = fmt.Sprintf("%s | $%.*f", base, precision, session.TotalCost)
	}

	return truncate(base, 128)
}

// buildStateLine constructs the State field for Discord presence.
func buildStateLine(session *SessionData, cfg *Config) string {
	sep := cfg.Display.Separator
	precision := 4
	if cfg.Display.CostPrecision != nil {
		precision = *cfg.Display.CostPrecision
	}

	var parts []string

	if showField(cfg.Show.ModelName) && session.ModelName != "" {
		parts = append(parts, session.ModelName)
	}
	if showField(cfg.Show.Tokens) {
		if showFieldDefault(cfg.Show.SplitTokens, false) {
			parts = append(parts, fmt.Sprintf("%s in | %s out",
				formatNumber(session.InputTokens),
				formatNumber(session.OutputTokens)))
		} else {
			parts = append(parts, fmt.Sprintf("%s tokens", formatNumber(session.TotalTokens)))
		}
	}
	if showField(cfg.Show.Cost) && !showField(cfg.Show.CostInDetails) {
		parts = append(parts, fmt.Sprintf("$%.*f", precision, session.TotalCost))
	}

	return truncate(strings.Join(parts, sep), 128)
}

// formatTemplate replaces template variables with session data values.
// Available variables: {project}, {branch}, {model}, {tokens},
// {in_tokens}, {out_tokens}, {cost}, {duration}, {separator}
func formatTemplate(template string, session *SessionData, cfg *Config) string {
	precision := 4
	if cfg.Display.CostPrecision != nil {
		precision = *cfg.Display.CostPrecision
	}

	duration := time.Since(session.StartTime).Truncate(time.Second).String()

	replacements := map[string]string{
		"{project}":    session.ProjectName,
		"{branch}":     session.GitBranch,
		"{model}":      session.ModelName,
		"{tokens}":     formatNumber(session.TotalTokens),
		"{in_tokens}":  formatNumber(session.InputTokens),
		"{out_tokens}": formatNumber(session.OutputTokens),
		"{cost}":       fmt.Sprintf("%.*f", precision, session.TotalCost),
		"{duration}":   duration,
		"{separator}":  cfg.Display.Separator,
	}

	result := template
	for key, value := range replacements {
		result = strings.Replace(result, key, value, -1)
	}

	return truncate(result, 128)
}

// truncate shortens a string to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
