package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui/components"
	"github.com/tipical/tipical/util"
)

// Sidebar renders the left panel with calendar list and mini-month.
type Sidebar struct {
	styles  *Styles
	miniCal *components.MiniCalendar
	store   *ical.Store
	width   int
	height  int

	// calendarLineIDs maps a rendered content line index (0-based, before padding)
	// to a calendar ID for click hit-testing.
	calendarLineIDs map[int]int
}

// NewSidebar creates a new sidebar component.
func NewSidebar(styles *Styles, cfg *config.Config, store *ical.Store) *Sidebar {
	theme := styles.Theme
	miniCal := components.NewMiniCalendar(theme, currentDate(), cfg.StartMonday())

	return &Sidebar{
		styles:          styles,
		miniCal:         miniCal,
		store:           store,
		width:           22,
		calendarLineIDs: make(map[int]int),
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

// HitTestCalendar returns the calendar ID clicked in the sidebar calendar list.
// tx is terminal column, contentY is 0-indexed from the top of the content area.
func (sb *Sidebar) HitTestCalendar(tx, contentY int) (int, bool) {
	// Sidebar has Padding(1, 1).
	relX := tx - 1
	relY := contentY - 1
	if relX < 0 || relY < 0 {
		return 0, false
	}
	calID, ok := sb.calendarLineIDs[relY]
	return calID, ok
}

// View renders the sidebar.
func (sb *Sidebar) View() string {
	var lines []string
	for k := range sb.calendarLineIDs {
		delete(sb.calendarLineIDs, k)
	}

	// Sidebar style has Padding(1,1); keep each content line within the inner width
	// so terminal auto-wrapping doesn't desync click hit-testing.
	innerWidth := sb.width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Mini calendar
	miniLines := strings.Split(sb.miniCal.View(), "\n")
	lines = append(lines, miniLines...)
	lines = append(lines, "")

	// Calendar list
	calTitle := lipgloss.NewStyle().
		Foreground(sb.styles.Theme.TextMuted).
		Bold(true).
		Render(util.TruncateText(" Calendars", innerWidth))
	lines = append(lines, calTitle)

	cals := sb.store.Calendars
	if len(cals) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(sb.styles.Theme.TextFaint).
			Render(util.TruncateText("  (no calendars)", innerWidth)))
	} else {
		// Group calendars by source account
		lastSource := ""
		headerStyle := lipgloss.NewStyle().
			Foreground(sb.styles.Theme.TextMuted).
			Bold(true)
		for _, cal := range cals {
			if cal.Source != "" && cal.Source != lastSource {
				src := util.TruncateText(cal.Source, innerWidth-2)
				lines = append(lines, headerStyle.Render("  "+src))
				lastSource = cal.Source
			}

			visible := sb.store.IsCalendarVisible(cal.ID)

			color := lipgloss.Color(cal.Color)
			if cal.Color == "" {
				color = sb.styles.Theme.CalendarColor(cal.ID)
			}
			dotStyle := lipgloss.NewStyle().Foreground(color)
			nameStyle := lipgloss.NewStyle().Foreground(sb.styles.Theme.Text)
			if !visible {
				dotStyle = dotStyle.Foreground(sb.styles.Theme.TextFaint)
				nameStyle = nameStyle.Foreground(sb.styles.Theme.TextFaint)
			}

			// Prefix is: "  " + "●" + " "
			prefixWidth := 2 + util.DisplayWidth("●") + 1
			nameWidth := innerWidth - prefixWidth
			if nameWidth < 0 {
				nameWidth = 0
			}
			name := util.TruncateText(cal.Name, nameWidth)

			lineIdx := len(lines)
			sb.calendarLineIDs[lineIdx] = cal.ID
			dot := dotStyle.Render("●")
			lines = append(lines, "  "+dot+" "+nameStyle.Render(name))
		}
	}

	content := strings.Join(lines, "\n")

	return sb.styles.Sidebar.
		Width(sb.width).
		Height(sb.height).
		Render(content)
}
