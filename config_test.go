package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// All Show fields should default to true
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

	// Display defaults
	if cfg.Display.DetailsPrefix != "Working on" {
		t.Errorf("DetailsPrefix = %q, want %q", cfg.Display.DetailsPrefix, "Working on")
	}
	if cfg.Display.Separator != " | " {
		t.Errorf("Separator = %q, want %q", cfg.Display.Separator, " | ")
	}
	if cfg.Display.CostPrecision == nil || *cfg.Display.CostPrecision != 4 {
		t.Errorf("CostPrecision should default to 4")
	}
	if cfg.Display.LargeText != "Clawd Code - Discord Rich Presence for Claude Code" {
		t.Errorf("LargeText = %q, want default", cfg.Display.LargeText)
	}
	if cfg.Display.DiscordAppID != "" {
		t.Errorf("DiscordAppID should default to empty string")
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
	if !showField(cfg.Show.ModelName) {
		t.Error("model_name should remain true (default)")
	}
	if !showField(cfg.Show.Tokens) {
		t.Error("tokens should remain true (default)")
	}
	if !showField(cfg.Show.Duration) {
		t.Error("duration should remain true (default)")
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
			"cost": false,
			"duration": true
		},
		"display": {
			"details_prefix": "Coding",
			"separator": " - ",
			"cost_precision": 2,
			"large_text": "My Custom Text",
			"discord_app_id": "123456789"
		}
	}`), 0644)

	cfg := LoadConfig()

	if !showField(cfg.Show.ProjectName) {
		t.Error("project_name should be true")
	}
	if showField(cfg.Show.GitBranch) {
		t.Error("git_branch should be false")
	}
	if !showField(cfg.Show.ModelName) {
		t.Error("model_name should be true")
	}
	if showField(cfg.Show.Tokens) {
		t.Error("tokens should be false")
	}
	if showField(cfg.Show.Cost) {
		t.Error("cost should be false")
	}
	if !showField(cfg.Show.Duration) {
		t.Error("duration should be true")
	}
	if cfg.Display.DetailsPrefix != "Coding" {
		t.Errorf("DetailsPrefix = %q, want %q", cfg.Display.DetailsPrefix, "Coding")
	}
	if cfg.Display.Separator != " - " {
		t.Errorf("Separator = %q, want %q", cfg.Display.Separator, " - ")
	}
	if cfg.Display.CostPrecision == nil || *cfg.Display.CostPrecision != 2 {
		t.Error("CostPrecision should be 2")
	}
	if cfg.Display.LargeText != "My Custom Text" {
		t.Errorf("LargeText = %q, want %q", cfg.Display.LargeText, "My Custom Text")
	}
	if cfg.Display.DiscordAppID != "123456789" {
		t.Errorf("DiscordAppID = %q, want %q", cfg.Display.DiscordAppID, "123456789")
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
		name    string
		show    ShowConfig
		prefix  string
		want    string
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
		noBranchSession := &SessionData{
			ProjectName: "my-project",
			GitBranch:   "",
		}
		cfg := DefaultConfig()
		got := buildDetailsLine(noBranchSession, cfg)
		if got != "Working on: my-project" {
			t.Errorf("empty branch should not show parens, got %q", got)
		}
	})

	t.Run("project hidden branch empty", func(t *testing.T) {
		noBranchSession := &SessionData{
			ProjectName: "my-project",
			GitBranch:   "",
		}
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
		ProjectName: "my-project",
		ModelName:   "Sonnet 4",
		TotalTokens: 10500,
		TotalCost:   0.1234,
		StartTime:   time.Now(),
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

	t.Run("empty model name with model shown", func(t *testing.T) {
		noModelSession := &SessionData{
			TotalTokens: 5000,
			TotalCost:   0.05,
		}
		cfg := DefaultConfig()
		got := buildStateLine(noModelSession, cfg)
		if got != "5.0K tokens | $0.0500" {
			t.Errorf("empty model should be omitted, got %q", got)
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
