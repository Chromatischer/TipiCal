package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ViewType represents a calendar view mode.
type ViewType string

const (
	ViewMonth    ViewType = "month"
	ViewWeek     ViewType = "week"
	ViewThreeDay ViewType = "threeday"
	ViewDay      ViewType = "day"
	ViewAgenda   ViewType = "agenda"
	ViewStacked  ViewType = "stacked"
)

// Config is the top-level configuration.
type Config struct {
	General   GeneralConfig    `toml:"general"`
	Theme     ThemeConfig      `toml:"theme"`
	Calendars []CalendarConfig `toml:"calendar"`
	Sync      SyncConfig       `toml:"sync"`
}

// GeneralConfig holds general app settings.
type GeneralConfig struct {
	DefaultView ViewType `toml:"default_view"`
	WeekStart   string   `toml:"week_start"`
	TimeFormat  string   `toml:"time_format"`
	DateFormat  string   `toml:"date_format"`
	AgendaDays  int      `toml:"agenda_days"`
}

// ThemeConfig holds theme settings.
type ThemeConfig struct {
	Mode   string `toml:"mode"`
	Accent string `toml:"accent"`
}

// CalendarConfig holds per-calendar settings.
type CalendarConfig struct {
	Name        string `toml:"name"`
	URL         string `toml:"url"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
	PasswordCmd string `toml:"password_cmd"`
	Color       string `toml:"color"`
	ReadOnly    bool   `toml:"readonly"`
}

// SyncConfig holds sync settings.
type SyncConfig struct {
	IntervalMinutes int    `toml:"interval_minutes"`
	CacheDir        string `toml:"cache_dir"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	cacheDir, _ := os.UserCacheDir()
	return &Config{
		General: GeneralConfig{
			DefaultView: ViewMonth,
			WeekStart:   "monday",
			TimeFormat:  "24h",
			DateFormat:  "2006-01-02",
			AgendaDays:  14,
		},
		Theme: ThemeConfig{
			Mode:   "auto",
			Accent: "#7C3AED",
		},
		Sync: SyncConfig{
			IntervalMinutes: 5,
			CacheDir:        filepath.Join(cacheDir, "tipical"),
		},
	}
}

// Load loads the config from ~/.config/tipical/config.toml.
// If the file doesn't exist, it returns defaults.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return cfg, nil
	}

	configPath := filepath.Join(configDir, "tipical", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, err
	}

	// Apply defaults for zero values
	if cfg.General.DefaultView == "" {
		cfg.General.DefaultView = ViewMonth
	}
	if cfg.General.WeekStart == "" {
		cfg.General.WeekStart = "monday"
	}
	if cfg.General.TimeFormat == "" {
		cfg.General.TimeFormat = "24h"
	}
	if cfg.General.AgendaDays == 0 {
		cfg.General.AgendaDays = 14
	}
	if cfg.Theme.Mode == "" {
		cfg.Theme.Mode = "auto"
	}
	if cfg.Theme.Accent == "" {
		cfg.Theme.Accent = "#7C3AED"
	}

	return cfg, nil
}

// Use24h returns true if 24-hour time format is configured.
func (c *Config) Use24h() bool {
	return c.General.TimeFormat == "24h"
}

// StartMonday returns true if week starts on Monday.
func (c *Config) StartMonday() bool {
	return c.General.WeekStart == "monday"
}
