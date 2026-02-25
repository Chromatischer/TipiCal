package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui/components"
)

// Sidebar renders the left panel with calendar list and mini-month.
type Sidebar struct {
	styles *Styles
	miniCal *components.MiniCalendar
	store  *ical.Store
	width  int
	height int
}

// NewSidebar creates a new sidebar component.
func NewSidebar(styles *Styles, cfg *config.Config, store *ical.Store) *Sidebar {
	theme := styles.Theme
	miniCal := components.NewMiniCalendar(theme, currentDate(), cfg.StartMonday())

	return &Sidebar{
		styles:  styles,
		miniCal: miniCal,
		store:   store,
		width:   22,
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

// HitTestDate returns the date clicked in the sidebar mini calendar.
// tx is the terminal column, contentY is 0-indexed from the top of the content area
// (i.e., terminal row minus 2 for the header). Returns the date and true if a day was clicked.
func (sb *Sidebar) HitTestDate(tx, contentY int) (time.Time, bool) {
	// Sidebar has Padding(1, 1): 1 col left padding, 1 row top padding.
	// Content (mini calendar) starts at tx=1, contentY=1.
	relX := tx - 1
	relY := contentY - 1
	if relX < 0 || relY < 0 {
		return time.Time{}, false
	}
	return sb.miniCal.HitTestDate(relX, relY)
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

	cals := sb.store.Calendars
	if len(cals) == 0 {
		// Show default calendars when no config and no discovered calendars
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
		// Group calendars by source account
		lastSource := ""
		headerStyle := lipgloss.NewStyle().
			Foreground(sb.styles.Theme.TextMuted).
			Bold(true)
		for _, cal := range cals {
			if cal.Source != "" && cal.Source != lastSource {
				sections = append(sections, headerStyle.Render("  "+cal.Source))
				lastSource = cal.Source
			}
			color := lipgloss.Color(cal.Color)
			if cal.Color == "" {
				color = sb.styles.Theme.CalendarColor(cal.ID)
			}
			dot := lipgloss.NewStyle().
				Foreground(color).
				Render("    ●")
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
