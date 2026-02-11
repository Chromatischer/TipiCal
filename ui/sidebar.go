package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ui/components"
)

// Sidebar renders the left panel with calendar list and mini-month.
type Sidebar struct {
	styles    *Styles
	miniCal   *components.MiniCalendar
	calendars []config.CalendarConfig
	width     int
	height    int
}

// NewSidebar creates a new sidebar component.
func NewSidebar(styles *Styles, cfg *config.Config) *Sidebar {
	theme := styles.Theme
	miniCal := components.NewMiniCalendar(theme, currentDate(), cfg.StartMonday())

	return &Sidebar{
		styles:    styles,
		miniCal:   miniCal,
		calendars: cfg.Calendars,
		width:     22,
	}
}

// SetSize sets the sidebar dimensions.
func (sb *Sidebar) SetSize(w, h int) {
	sb.width = w
	sb.height = h
}

// UpdateDate updates the mini calendar date.
func (sb *Sidebar) UpdateDate(d interface {
	Year() int
	Month() interface{ String() string }
}) {
	// This is simplified; actual implementation takes time.Time
}

// View renders the sidebar.
func (sb *Sidebar) View() string {
	var sections []string

	// Mini calendar
	sections = append(sections, sb.miniCal.View())
	sections = append(sections, "")

	// Calendar list
	calTitle := lipgloss.NewStyle().
		Foreground(sb.styles.Theme.TextMuted).
		Bold(true).
		Render(" Calendars")
	sections = append(sections, calTitle)

	if len(sb.calendars) == 0 {
		// Show default calendars when no config
		defaults := []struct {
			name  string
			color lipgloss.Color
		}{
			{"Work", sb.styles.Theme.CalendarColor(0)},
			{"Personal", sb.styles.Theme.CalendarColor(1)},
		}
		for _, cal := range defaults {
			dot := lipgloss.NewStyle().
				Foreground(cal.color).
				Render("  ●")
			name := lipgloss.NewStyle().
				Foreground(sb.styles.Theme.Text).
				Render(" " + cal.name)
			sections = append(sections, dot+name)
		}
	} else {
		for i, cal := range sb.calendars {
			color := sb.styles.Theme.CalendarColor(i)
			if cal.Color != "" {
				color = lipgloss.Color(cal.Color)
			}
			dot := lipgloss.NewStyle().
				Foreground(color).
				Render("  ●")
			name := lipgloss.NewStyle().
				Foreground(sb.styles.Theme.Text).
				Render(" " + cal.Name)
			sections = append(sections, dot+name)
		}
	}

	content := strings.Join(sections, "\n")

	return sb.styles.Sidebar.
		Width(sb.width).
		Height(sb.height).
		Render(content)
}
