package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Show fields that default to true
	if !showField(cfg.Show.ProjectName) {
		t.Error("ProjectName should default to true")
	}
	if !showField(cfg.Show.GitBranch) {
		t.Error("GitBranch should default to true")
	}
	if !showField(cfg.Show.ModelName) {
		t.Error("ModelName should default to true")
	}
	if !showField(cfg.Show.Tokens) {
		t.Error("Tokens should default to true")
	}
	if !showField(cfg.Show.Cost) {
		t.Error("Cost should default to true")
	}
	if !showField(cfg.Show.Duration) {
		t.Error("Duration should default to true")
	}

	// Show fields that default to false
	if showField(cfg.Show.CostInDetails) {
		t.Error("CostInDetails should default to false")
	}
	if showFieldDefault(cfg.Show.SplitTokens, false) {
		t.Error("SplitTokens should default to false")
	}
	if showFieldDefault(cfg.Show.SessionFocus, false) {
		t.Error("SessionFocus should default to false")
	}

	// Display defaults
	if cfg.Display.DetailsPrefix != "Working on" {
		t.Errorf("DetailsPrefix = %q, want %q", cfg.Display.DetailsPrefix, "Working on")
	}
	if cfg.Display.Separator != " | " {
		t.Errorf("Separator = %q, want %q", cfg.Display.Separator, " | ")
	}
	if cfg.Display.CostPrecision == nil || *cfg.Display.CostPrecision != 4 {
		t.Error("CostPrecision should default to 4")
	}
	if cfg.Display.IdleTimeout == nil || *cfg.Display.IdleTimeout != 0 {
		t.Error("IdleTimeout should default to 0")
	}
	if cfg.Display.DetailsFormat != "" {
		t.Error("DetailsFormat should default to empty")
	}
	if cfg.Display.StateFormat != "" {
		t.Error("StateFormat should default to empty")
	}
	if cfg.Display.LargeText != "Clawd Code - Discord Rich Presence for Claude Code" {
		t.Errorf("LargeText = %q, want default", cfg.Display.LargeText)
	}
	if cfg.Display.DiscordAppID != "" {
		t.Error("DiscordAppID should default to empty string")
	}

	// Buttons default to nil
	if len(cfg.Buttons) != 0 {
		t.Error("Buttons should default to empty")
	}
}

func TestShowField(t *testing.T) {
	tests := []struct {
		name string
		ptr  *bool
		want bool
	}{
		{"nil returns true", nil, true},
		{"true returns true", boolPtr(true), true},
		{"false returns false", boolPtr(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := showField(tt.ptr); got != tt.want {
				t.Errorf("showField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShowFieldDefault(t *testing.T) {
	tests := []struct {
		name       string
		ptr        *bool
		defaultVal bool
		want       bool
	}{
		{"nil with default true", nil, true, true},
		{"nil with default false", nil, false, false},
		{"true overrides default false", boolPtr(true), false, true},
		{"false overrides default true", boolPtr(false), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := showFieldDefault(tt.ptr, tt.defaultVal); got != tt.want {
				t.Errorf("showFieldDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	configFilePath = filepath.Join(t.TempDir(), "nonexistent.json")
	cfg := LoadConfig()

	if !showField(cfg.Show.ProjectName) {
		t.Error("should default to true when no file")
	}
	if cfg.Display.DetailsPrefix != "Working on" {
		t.Error("should have default prefix when no file")
	}
}

func TestLoadConfig_EmptyJSON(t *testing.T) {
	dir := t.TempDir()
	configFilePath = filepath.Join(dir, "discord-presence.json")
	os.WriteFile(configFilePath, []byte(`{}`), 0644)

	cfg := LoadConfig()

	if !showField(cfg.Show.ProjectName) {
		t.Error("empty JSON should default all to true")
	}
	if !showField(cfg.Show.GitBranch) {
		t.Error("empty JSON should default all to true")
	}
	if cfg.Display.Separator != " | " {
		t.Errorf("empty JSON should use default separator, got %q", cfg.Display.Separator)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	configFilePath = filepath.Join(dir, "discord-presence.json")
	os.WriteFile(configFilePath, []byte(`{invalid json`), 0644)

	cfg := LoadConfig()

	if !showField(cfg.Show.ProjectName) {
		t.Error("invalid JSON should fall back to defaults")
	}
	if cfg.Display.DetailsPrefix != "Working on" {
		t.Error("invalid JSON should fall back to default prefix")
	}
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	configFilePath = filepath.Join(dir, "discord-presence.json")
	os.WriteFile(configFilePath, []byte(`{"show":{"cost":false,"git_branch":false}}`), 0644)

	cfg := LoadConfig()

	if showField(cfg.Show.Cost) {
		t.Error("cost should be false")
	}
	if showField(cfg.Show.GitBranch) {
		t.Error("git_branch should be false")
	}
	if !showField(cfg.Show.ProjectName) {
		t.Error("project_name should remain true (default)")
	}
}

func TestLoadConfig_FullConfig(t *testing.T) {
	dir := t.TempDir()
	configFilePath = filepath.Join(dir, "discord-presence.json")
	os.WriteFile(configFilePath, []byte(`{
		"show": {
			"project_name": true,
			"git_branch": false,
			"model_name": true,
			"tokens": false,
			"split_tokens": true,
			"cost": false,
			"duration": true,
			"session_focus": true
		},
		"display": {
			"details_prefix": "Coding",
			"details_format": "{project}",
			"state_format": "{model}",
			"separator": " - ",
			"cost_precision": 2,
			"idle_timeout": 60,
			"large_text": "My Custom Text",
			"discord_app_id": "123456789"
		},
		"buttons": [
			{"label": "GitHub", "url": "https://github.com/test"}
		]
	}`), 0644)

	cfg := LoadConfig()

	if showField(cfg.Show.GitBranch) {
		t.Error("git_branch should be false")
	}
	if !showFieldDefault(cfg.Show.SplitTokens, false) {
		t.Error("split_tokens should be true")
	}
	if !showFieldDefault(cfg.Show.SessionFocus, false) {
		t.Error("session_focus should be true")
	}
	if cfg.Display.DetailsFormat != "{project}" {
		t.Errorf("DetailsFormat = %q, want {project}", cfg.Display.DetailsFormat)
	}
	if cfg.Display.StateFormat != "{model}" {
		t.Errorf("StateFormat = %q, want {model}", cfg.Display.StateFormat)
	}
	if *cfg.Display.IdleTimeout != 60 {
		t.Errorf("IdleTimeout = %d, want 60", *cfg.Display.IdleTimeout)
	}
	if len(cfg.Buttons) != 1 || cfg.Buttons[0].Label != "GitHub" {
		t.Error("buttons should have 1 entry with label GitHub")
	}
}

func TestLoadConfig_UnknownFieldsIgnored(t *testing.T) {
	dir := t.TempDir()
	configFilePath = filepath.Join(dir, "discord-presence.json")
	os.WriteFile(configFilePath, []byte(`{"show":{"cost":false},"unknown_field":"value"}`), 0644)

	cfg := LoadConfig()

	if showField(cfg.Show.Cost) {
		t.Error("cost should be false")
	}
	if !showField(cfg.Show.ProjectName) {
		t.Error("project_name should remain default true")
	}
}

func TestMergeConfig(t *testing.T) {
	defaults := DefaultConfig()

	user := &Config{
		Show: ShowConfig{
			Cost:      boolPtr(false),
			GitBranch: boolPtr(false),
		},
		Display: DisplayConfig{
			Separator: " - ",
		},
	}

	result := mergeConfig(defaults, user)

	if showField(result.Show.Cost) {
		t.Error("cost should be overridden to false")
	}
	if showField(result.Show.GitBranch) {
		t.Error("git_branch should be overridden to false")
	}
	if !showField(result.Show.ProjectName) {
		t.Error("project_name should remain default true")
	}
	if result.Display.Separator != " - " {
		t.Errorf("separator should be overridden, got %q", result.Display.Separator)
	}
	if result.Display.DetailsPrefix != "Working on" {
		t.Error("details_prefix should remain default")
	}
}

func TestCostPrecisionClamp(t *testing.T) {
	defaults := DefaultConfig()

	t.Run("negative clamped to 0", func(t *testing.T) {
		user := &Config{Display: DisplayConfig{CostPrecision: intPtr(-5)}}
		result := mergeConfig(defaults, user)
		if *result.Display.CostPrecision != 0 {
			t.Errorf("expected 0, got %d", *result.Display.CostPrecision)
		}
	})

	t.Run("over 10 clamped to 10", func(t *testing.T) {
		user := &Config{Display: DisplayConfig{CostPrecision: intPtr(99)}}
		result := mergeConfig(defaults, user)
		if *result.Display.CostPrecision != 10 {
			t.Errorf("expected 10, got %d", *result.Display.CostPrecision)
		}
	})

	t.Run("valid value kept", func(t *testing.T) {
		user := &Config{Display: DisplayConfig{CostPrecision: intPtr(2)}}
		result := mergeConfig(defaults, user)
		if *result.Display.CostPrecision != 2 {
			t.Errorf("expected 2, got %d", *result.Display.CostPrecision)
		}
	})
}

func TestIdleTimeoutClamp(t *testing.T) {
	defaults := DefaultConfig()

	t.Run("negative clamped to 0", func(t *testing.T) {
		user := &Config{Display: DisplayConfig{IdleTimeout: intPtr(-10)}}
		result := mergeConfig(defaults, user)
		if *result.Display.IdleTimeout != 0 {
			t.Errorf("expected 0, got %d", *result.Display.IdleTimeout)
		}
	})

	t.Run("over 3600 clamped to 3600", func(t *testing.T) {
		user := &Config{Display: DisplayConfig{IdleTimeout: intPtr(9999)}}
		result := mergeConfig(defaults, user)
		if *result.Display.IdleTimeout != 3600 {
			t.Errorf("expected 3600, got %d", *result.Display.IdleTimeout)
		}
	})

	t.Run("valid value kept", func(t *testing.T) {
		user := &Config{Display: DisplayConfig{IdleTimeout: intPtr(120)}}
		result := mergeConfig(defaults, user)
		if *result.Display.IdleTimeout != 120 {
			t.Errorf("expected 120, got %d", *result.Display.IdleTimeout)
		}
	})
}

func TestValidateButtons(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		result := validateButtons(nil)
		if len(result) != 0 {
			t.Error("empty input should return empty")
		}
	})

	t.Run("valid button", func(t *testing.T) {
		result := validateButtons([]ButtonConfig{{Label: "GitHub", URL: "https://github.com"}})
		if len(result) != 1 || result[0].Label != "GitHub" {
			t.Error("valid button should pass through")
		}
	})

	t.Run("max 2 buttons", func(t *testing.T) {
		result := validateButtons([]ButtonConfig{
			{Label: "A", URL: "https://a.com"},
			{Label: "B", URL: "https://b.com"},
			{Label: "C", URL: "https://c.com"},
		})
		if len(result) != 2 {
			t.Errorf("expected 2, got %d", len(result))
		}
	})

	t.Run("empty label dropped", func(t *testing.T) {
		result := validateButtons([]ButtonConfig{{Label: "", URL: "https://a.com"}})
		if len(result) != 0 {
			t.Error("empty label should be dropped")
		}
	})

	t.Run("empty URL dropped", func(t *testing.T) {
		result := validateButtons([]ButtonConfig{{Label: "A", URL: ""}})
		if len(result) != 0 {
			t.Error("empty URL should be dropped")
		}
	})

	t.Run("non-http URL dropped", func(t *testing.T) {
		result := validateButtons([]ButtonConfig{{Label: "A", URL: "ftp://a.com"}})
		if len(result) != 0 {
			t.Error("non-http URL should be dropped")
		}
	})

	t.Run("label truncated to 32 chars", func(t *testing.T) {
		result := validateButtons([]ButtonConfig{{Label: strings.Repeat("a", 50), URL: "https://a.com"}})
		if len(result) != 1 || len(result[0].Label) != 32 {
			t.Errorf("label should be truncated to 32, got %d", len(result[0].Label))
		}
	})

	t.Run("mixed valid and invalid", func(t *testing.T) {
		result := validateButtons([]ButtonConfig{
			{Label: "", URL: "https://a.com"},
			{Label: "Valid", URL: "https://b.com"},
			{Label: "No URL", URL: ""},
			{Label: "Also Valid", URL: "http://c.com"},
		})
		if len(result) != 2 || result[0].Label != "Valid" || result[1].Label != "Also Valid" {
			t.Errorf("expected Valid + Also Valid, got %v", result)
		}
	})
}

func TestBuildDetailsLine(t *testing.T) {
	session := &SessionData{
		ProjectName: "my-project",
		GitBranch:   "main",
		ModelName:   "Sonnet 4",
		TotalTokens: 10500,
		TotalCost:   0.1234,
		StartTime:   time.Now(),
	}

	tests := []struct {
		name   string
		show   ShowConfig
		prefix string
		want   string
	}{
		{
			name:   "all visible with branch",
			show:   ShowConfig{ProjectName: boolPtr(true), GitBranch: boolPtr(true)},
			prefix: "Working on",
			want:   "Working on: my-project (main)",
		},
		{
			name:   "project only",
			show:   ShowConfig{ProjectName: boolPtr(true), GitBranch: boolPtr(false)},
			prefix: "Working on",
			want:   "Working on: my-project",
		},
		{
			name:   "branch only",
			show:   ShowConfig{ProjectName: boolPtr(false), GitBranch: boolPtr(true)},
			prefix: "Working on",
			want:   "Working on: main",
		},
		{
			name:   "all hidden",
			show:   ShowConfig{ProjectName: boolPtr(false), GitBranch: boolPtr(false)},
			prefix: "Working on",
			want:   "",
		},
		{
			name:   "custom prefix",
			show:   ShowConfig{ProjectName: boolPtr(true), GitBranch: boolPtr(true)},
			prefix: "Coding",
			want:   "Coding: my-project (main)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Show.ProjectName = tt.show.ProjectName
			cfg.Show.GitBranch = tt.show.GitBranch
			cfg.Display.DetailsPrefix = tt.prefix

			got := buildDetailsLine(session, cfg)
			if got != tt.want {
				t.Errorf("buildDetailsLine() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("branch visible but empty", func(t *testing.T) {
		noBranchSession := &SessionData{ProjectName: "my-project", GitBranch: ""}
		cfg := DefaultConfig()
		got := buildDetailsLine(noBranchSession, cfg)
		if got != "Working on: my-project" {
			t.Errorf("empty branch should not show parens, got %q", got)
		}
	})

	t.Run("cost in details", func(t *testing.T) {
		costSession := &SessionData{ProjectName: "my-project", GitBranch: "main", TotalCost: 21.89}
		cfg := DefaultConfig()
		cfg.Show.CostInDetails = boolPtr(true)
		cfg.Display.CostPrecision = intPtr(2)
		got := buildDetailsLine(costSession, cfg)
		if got != "Working on: my-project (main) | $21.89" {
			t.Errorf("cost should be in details, got %q", got)
		}
	})

	t.Run("project hidden branch empty", func(t *testing.T) {
		noBranchSession := &SessionData{ProjectName: "my-project", GitBranch: ""}
		cfg := DefaultConfig()
		cfg.Show.ProjectName = boolPtr(false)
		got := buildDetailsLine(noBranchSession, cfg)
		if got != "" {
			t.Errorf("project hidden + empty branch should be empty, got %q", got)
		}
	})
}

func TestBuildStateLine(t *testing.T) {
	session := &SessionData{
		ProjectName:  "my-project",
		ModelName:    "Sonnet 4",
		TotalTokens:  10500,
		InputTokens:  7000,
		OutputTokens: 3500,
		TotalCost:    0.1234,
		StartTime:    time.Now(),
	}

	tests := []struct {
		name      string
		show      ShowConfig
		separator string
		precision *int
		want      string
	}{
		{
			name:      "all visible",
			show:      ShowConfig{ModelName: boolPtr(true), Tokens: boolPtr(true), Cost: boolPtr(true)},
			separator: " | ",
			precision: intPtr(4),
			want:      "Sonnet 4 | 10.5K tokens | $0.1234",
		},
		{
			name:      "model only",
			show:      ShowConfig{ModelName: boolPtr(true), Tokens: boolPtr(false), Cost: boolPtr(false)},
			separator: " | ",
			precision: intPtr(4),
			want:      "Sonnet 4",
		},
		{
			name:      "tokens only",
			show:      ShowConfig{ModelName: boolPtr(false), Tokens: boolPtr(true), Cost: boolPtr(false)},
			separator: " | ",
			precision: intPtr(4),
			want:      "10.5K tokens",
		},
		{
			name:      "cost only",
			show:      ShowConfig{ModelName: boolPtr(false), Tokens: boolPtr(false), Cost: boolPtr(true)},
			separator: " | ",
			precision: intPtr(4),
			want:      "$0.1234",
		},
		{
			name:      "all hidden",
			show:      ShowConfig{ModelName: boolPtr(false), Tokens: boolPtr(false), Cost: boolPtr(false)},
			separator: " | ",
			precision: intPtr(4),
			want:      "",
		},
		{
			name:      "custom separator",
			show:      ShowConfig{ModelName: boolPtr(true), Tokens: boolPtr(true), Cost: boolPtr(true)},
			separator: " - ",
			precision: intPtr(4),
			want:      "Sonnet 4 - 10.5K tokens - $0.1234",
		},
		{
			name:      "custom precision 2",
			show:      ShowConfig{ModelName: boolPtr(false), Tokens: boolPtr(false), Cost: boolPtr(true)},
			separator: " | ",
			precision: intPtr(2),
			want:      "$0.12",
		},
		{
			name:      "precision 0",
			show:      ShowConfig{ModelName: boolPtr(false), Tokens: boolPtr(false), Cost: boolPtr(true)},
			separator: " | ",
			precision: intPtr(0),
			want:      "$0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Show.ModelName = tt.show.ModelName
			cfg.Show.Tokens = tt.show.Tokens
			cfg.Show.Cost = tt.show.Cost
			cfg.Display.Separator = tt.separator
			cfg.Display.CostPrecision = tt.precision

			got := buildStateLine(session, cfg)
			if got != tt.want {
				t.Errorf("buildStateLine() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("cost in details excludes cost from state", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Show.CostInDetails = boolPtr(true)
		got := buildStateLine(session, cfg)
		if got != "Sonnet 4 | 10.5K tokens" {
			t.Errorf("cost_in_details should exclude cost from state, got %q", got)
		}
	})

	t.Run("split tokens enabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Show.SplitTokens = boolPtr(true)
		cfg.Show.Cost = boolPtr(false)
		got := buildStateLine(session, cfg)
		if got != "Sonnet 4 | 7.0K in | 3.5K out" {
			t.Errorf("split tokens should show in/out, got %q", got)
		}
	})

	t.Run("split tokens disabled shows total", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Show.Cost = boolPtr(false)
		got := buildStateLine(session, cfg)
		if got != "Sonnet 4 | 10.5K tokens" {
			t.Errorf("non-split should show total, got %q", got)
		}
	})

	t.Run("empty model name with model shown", func(t *testing.T) {
		noModelSession := &SessionData{TotalTokens: 5000, TotalCost: 0.05}
		cfg := DefaultConfig()
		got := buildStateLine(noModelSession, cfg)
		if got != "5.0K tokens | $0.0500" {
			t.Errorf("empty model should be omitted, got %q", got)
		}
	})
}

func TestFormatTemplate(t *testing.T) {
	session := &SessionData{
		ProjectName:  "my-project",
		GitBranch:    "main",
		ModelName:    "Sonnet 4",
		TotalTokens:  200000,
		InputTokens:  150000,
		OutputTokens: 50000,
		TotalCost:    0.1234,
		StartTime:    time.Now(),
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"project only", "{project}", "my-project"},
		{"project and branch", "{project} ({branch})", "my-project (main)"},
		{"model and tokens", "{model} | {tokens}", "Sonnet 4 | 200.0K"},
		{"split tokens", "{in_tokens} in | {out_tokens} out", "150.0K in | 50.0K out"},
		{"cost with dollar", "${cost}", "$0.1234"},
		{"separator variable", "{model}{separator}{tokens}", "Sonnet 4 | 200.0K"},
		{"plain text", "Just coding", "Just coding"},
		{"unknown variable", "{unknown} var", "{unknown} var"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			got := formatTemplate(tt.template, session, cfg)
			if got != tt.want {
				t.Errorf("formatTemplate(%q) = %q, want %q", tt.template, got, tt.want)
			}
		})
	}

	t.Run("custom cost precision", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Display.CostPrecision = intPtr(2)
		got := formatTemplate("${cost}", session, cfg)
		if got != "$0.12" {
			t.Errorf("expected $0.12, got %q", got)
		}
	})

	t.Run("truncation at 128 chars", func(t *testing.T) {
		cfg := DefaultConfig()
		got := formatTemplate(strings.Repeat("a", 200), session, cfg)
		if len(got) != 128 {
			t.Errorf("expected 128 chars, got %d", len(got))
		}
	})
}

func TestSessionDataChanged(t *testing.T) {
	base := &SessionData{
		ProjectName: "proj",
		ModelName:   "Sonnet 4",
		GitBranch:   "main",
		TotalTokens: 100,
		TotalCost:   0.5,
	}

	t.Run("nil old is changed", func(t *testing.T) {
		if !sessionDataChanged(nil, base) {
			t.Error("nil old should be changed")
		}
	})

	t.Run("same data not changed", func(t *testing.T) {
		same := *base
		if sessionDataChanged(base, &same) {
			t.Error("same data should not be changed")
		}
	})

	t.Run("different tokens is changed", func(t *testing.T) {
		diff := *base
		diff.TotalTokens = 200
		if !sessionDataChanged(base, &diff) {
			t.Error("different tokens should be changed")
		}
	})

	t.Run("different cost is changed", func(t *testing.T) {
		diff := *base
		diff.TotalCost = 1.0
		if !sessionDataChanged(base, &diff) {
			t.Error("different cost should be changed")
		}
	})

	t.Run("different project is changed", func(t *testing.T) {
		diff := *base
		diff.ProjectName = "other"
		if !sessionDataChanged(base, &diff) {
			t.Error("different project should be changed")
		}
	})
}

func TestCheckIdle(t *testing.T) {
	base := &SessionData{TotalTokens: 100, TotalCost: 0.5, ProjectName: "p", ModelName: "m"}

	t.Run("timeout 0 never idle", func(t *testing.T) {
		idle, _ := checkIdle(base, base, time.Now().Add(-time.Hour), 0)
		if idle {
			t.Error("timeout 0 should never be idle")
		}
	})

	t.Run("data changed resets idle", func(t *testing.T) {
		changed := *base
		changed.TotalTokens = 200
		idle, newChange := checkIdle(base, &changed, time.Now().Add(-time.Hour), 10)
		if idle {
			t.Error("changed data should not be idle")
		}
		if time.Since(newChange) > time.Second {
			t.Error("lastChange should be reset to now")
		}
	})

	t.Run("no change within timeout not idle", func(t *testing.T) {
		idle, _ := checkIdle(base, base, time.Now(), 60)
		if idle {
			t.Error("within timeout should not be idle")
		}
	})

	t.Run("no change past timeout is idle", func(t *testing.T) {
		idle, _ := checkIdle(base, base, time.Now().Add(-2*time.Minute), 60)
		if !idle {
			t.Error("past timeout should be idle")
		}
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 128, "hello"},
		{"exact length", "abc", 3, "abc"},
		{"needs truncation", "hello world this is a very long string", 10, "hello w..."},
		{"maxLen 3", "hello", 3, "hel"},
		{"maxLen 2", "hello", 2, "he"},
		{"maxLen 1", "hello", 1, "h"},
		{"empty string", "", 128, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
