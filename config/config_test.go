package config

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.General.DefaultView != ViewMonth {
		t.Errorf("DefaultView = %q, want %q", cfg.General.DefaultView, ViewMonth)
	}
	if cfg.General.WeekStart != "monday" {
		t.Errorf("WeekStart = %q, want %q", cfg.General.WeekStart, "monday")
	}
	if cfg.General.TimeFormat != "24h" {
		t.Errorf("TimeFormat = %q, want %q", cfg.General.TimeFormat, "24h")
	}
	if cfg.General.DateFormat != "2006-01-02" {
		t.Errorf("DateFormat = %q, want %q", cfg.General.DateFormat, "2006-01-02")
	}
	if cfg.General.AgendaDays != 14 {
		t.Errorf("AgendaDays = %d, want 14", cfg.General.AgendaDays)
	}
	if cfg.Theme.Mode != "auto" {
		t.Errorf("Theme.Mode = %q, want %q", cfg.Theme.Mode, "auto")
	}
	if cfg.Theme.Accent != "#7C3AED" {
		t.Errorf("Theme.Accent = %q, want %q", cfg.Theme.Accent, "#7C3AED")
	}
	if cfg.Sync.IntervalMinutes != 5 {
		t.Errorf("Sync.IntervalMinutes = %d, want 5", cfg.Sync.IntervalMinutes)
	}
	if !strings.Contains(cfg.Sync.CacheDir, "tipical") {
		t.Errorf("Sync.CacheDir = %q, want path containing \"tipical\"", cfg.Sync.CacheDir)
	}
}

func TestDefaultConfigHelpers(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Use24h() {
		t.Error("Use24h() = false, want true for default 24h config")
	}
	if !cfg.StartMonday() {
		t.Error("StartMonday() = false, want true for default monday config")
	}
}

func TestCalendarColor(t *testing.T) {
	colors := []lipgloss.Color{"#ff0000", "#00ff00", "#0000ff"}
	theme := &Theme{
		Accent:         lipgloss.Color("#7C3AED"),
		CalendarColors: colors,
	}

	t.Run("index 0 returns first color", func(t *testing.T) {
		got := theme.CalendarColor(0)
		if got != colors[0] {
			t.Errorf("CalendarColor(0) = %q, want %q", got, colors[0])
		}
	})

	t.Run("index within bounds returns correct color", func(t *testing.T) {
		got := theme.CalendarColor(2)
		if got != colors[2] {
			t.Errorf("CalendarColor(2) = %q, want %q", got, colors[2])
		}
	})

	t.Run("index equal to len wraps to first", func(t *testing.T) {
		got := theme.CalendarColor(3) // len=3, 3%3=0
		if got != colors[0] {
			t.Errorf("CalendarColor(3) = %q, want %q (first color)", got, colors[0])
		}
	})

	t.Run("index larger than len wraps without panic", func(t *testing.T) {
		got := theme.CalendarColor(100)
		want := colors[100%len(colors)]
		if got != want {
			t.Errorf("CalendarColor(100) = %q, want %q", got, want)
		}
	})

	t.Run("empty palette returns Accent color", func(t *testing.T) {
		emptyTheme := &Theme{
			Accent:         lipgloss.Color("#7C3AED"),
			CalendarColors: nil,
		}
		got := emptyTheme.CalendarColor(0)
		if got != emptyTheme.Accent {
			t.Errorf("CalendarColor(0) with empty palette = %q, want Accent %q", got, emptyTheme.Accent)
		}
	})
}
