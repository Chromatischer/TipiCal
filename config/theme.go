package config

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme holds the resolved color palette for the current terminal.
type Theme struct {
	IsDark bool

	// Base colors
	Background lipgloss.Color
	Surface    lipgloss.Color
	SurfaceAlt lipgloss.Color
	Overlay    lipgloss.Color

	// Text colors
	Text      lipgloss.Color
	TextMuted lipgloss.Color
	TextFaint lipgloss.Color
	Link      lipgloss.Color

	// Semantic colors
	Accent   lipgloss.Color
	Today    lipgloss.Color
	Selected lipgloss.Color
	Border   lipgloss.Color
	Error    lipgloss.Color
	Success  lipgloss.Color
	Warning  lipgloss.Color

	// Calendar colors (for distinguishing calendars)
	CalendarColors []lipgloss.Color
}

// DefaultCalendarColors returns a set of distinct colors for calendars.
func defaultCalendarColorsDark() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("#7C3AED"), // violet
		lipgloss.Color("#3B82F6"), // blue
		lipgloss.Color("#10B981"), // emerald
		lipgloss.Color("#F59E0B"), // amber
		lipgloss.Color("#EF4444"), // red
		lipgloss.Color("#EC4899"), // pink
		lipgloss.Color("#06B6D4"), // cyan
		lipgloss.Color("#8B5CF6"), // purple
	}
}

func defaultCalendarColorsLight() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("#6D28D9"), // violet
		lipgloss.Color("#2563EB"), // blue
		lipgloss.Color("#059669"), // emerald
		lipgloss.Color("#D97706"), // amber
		lipgloss.Color("#DC2626"), // red
		lipgloss.Color("#DB2777"), // pink
		lipgloss.Color("#0891B2"), // cyan
		lipgloss.Color("#7C3AED"), // purple
	}
}

// NewTheme creates a theme based on terminal background detection and config.
func NewTheme(cfg *Config) *Theme {
	isDark := lipgloss.HasDarkBackground()

	if cfg.Theme.Mode == "light" {
		isDark = false
	} else if cfg.Theme.Mode == "dark" {
		isDark = true
	}

	accent := lipgloss.Color(cfg.Theme.Accent)

	if isDark {
		return &Theme{
			IsDark:         true,
			Background:     lipgloss.Color("#1E1E2E"),
			Surface:        lipgloss.Color("#313244"),
			SurfaceAlt:     lipgloss.Color("#45475A"),
			Overlay:        lipgloss.Color("#585B70"),
			Text:           lipgloss.Color("#CDD6F4"),
			TextMuted:      lipgloss.Color("#A6ADC8"),
			TextFaint:      lipgloss.Color("#6C7086"),
			Link:           lipgloss.Color("#74C7EC"),
			Accent:         accent,
			Today:          lipgloss.Color("#F38BA8"),
			Selected:       lipgloss.Color("#89B4FA"),
			Border:         lipgloss.Color("#585B70"),
			Error:          lipgloss.Color("#F38BA8"),
			Success:        lipgloss.Color("#A6E3A1"),
			Warning:        lipgloss.Color("#F9E2AF"),
			CalendarColors: defaultCalendarColorsDark(),
		}
	}

	return &Theme{
		IsDark:         false,
		Background:     lipgloss.Color("#EFF1F5"),
		Surface:        lipgloss.Color("#E6E9EF"),
		SurfaceAlt:     lipgloss.Color("#DCE0E8"),
		Overlay:        lipgloss.Color("#BCC0CC"),
		Text:           lipgloss.Color("#4C4F69"),
		TextMuted:      lipgloss.Color("#6C6F85"),
		TextFaint:      lipgloss.Color("#9CA0B0"),
		Link:           lipgloss.Color("#1E66F5"),
		Accent:         accent,
		Today:          lipgloss.Color("#D20F39"),
		Selected:       lipgloss.Color("#1E66F5"),
		Border:         lipgloss.Color("#BCC0CC"),
		Error:          lipgloss.Color("#D20F39"),
		Success:        lipgloss.Color("#40A02B"),
		Warning:        lipgloss.Color("#DF8E1D"),
		CalendarColors: defaultCalendarColorsLight(),
	}
}

// CalendarColor returns the color for calendar at index i.
func (t *Theme) CalendarColor(i int) lipgloss.Color {
	if len(t.CalendarColors) == 0 {
		return t.Accent
	}
	return t.CalendarColors[i%len(t.CalendarColors)]
}
