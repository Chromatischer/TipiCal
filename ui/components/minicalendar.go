package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/util"
)

// MiniCalendar renders a small month calendar for the sidebar.
type MiniCalendar struct {
	theme       *config.Theme
	year        int
	month       time.Month
	selectedDay time.Time
	startMonday bool
}

// NewMiniCalendar creates a mini calendar.
func NewMiniCalendar(theme *config.Theme, selected time.Time, startMonday bool) *MiniCalendar {
	return &MiniCalendar{
		theme:       theme,
		year:        selected.Year(),
		month:       selected.Month(),
		selectedDay: selected,
		startMonday: startMonday,
	}
}

// SetDate updates the displayed month.
func (m *MiniCalendar) SetDate(d time.Time) {
	m.year = d.Year()
	m.month = d.Month()
	m.selectedDay = d
}

// HitTestDate returns the date at the given position relative to the mini calendar's
// top-left corner (0-indexed). Returns the date and true if a day cell was clicked.
func (m *MiniCalendar) HitTestDate(relX, relY int) (time.Time, bool) {
	// Row 0: month/year header (not a day)
	// Row 1: weekday headers (not a day)
	// Rows 2-7: day grid; each day cell is Width(3) = 3 cols
	if relY < 2 || relX < 0 {
		return time.Time{}, false
	}
	weekRow := relY - 2
	colIdx := relX / 3
	if colIdx > 6 {
		return time.Time{}, false
	}
	grid := util.MonthGrid(m.year, m.month, m.startMonday)
	if weekRow >= len(grid) {
		return time.Time{}, false
	}
	if colIdx >= len(grid[weekRow]) {
		return time.Time{}, false
	}
	return grid[weekRow][colIdx], true
}

// View renders the mini calendar (20 chars wide).
func (m *MiniCalendar) View() string {
	var lines []string

	// Month/year header
	header := fmt.Sprintf(" %s %d", m.month.String()[:3], m.year)
	headerStyle := lipgloss.NewStyle().
		Foreground(m.theme.Accent).
		Bold(true)
	lines = append(lines, headerStyle.Render(header))

	// Weekday headers
	var dayHeaders []string
	days := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	if !m.startMonday {
		days = []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	}
	for _, d := range days {
		dayHeaders = append(dayHeaders, lipgloss.NewStyle().
			Foreground(m.theme.TextFaint).
			Width(3).
			Align(lipgloss.Center).
			Render(d))
	}
	lines = append(lines, strings.Join(dayHeaders, ""))

	// Day grid
	grid := util.MonthGrid(m.year, m.month, m.startMonday)
	now := time.Now()

	for _, week := range grid {
		var dayCells []string
		for _, day := range week {
			num := fmt.Sprintf("%2d", day.Day())
			style := lipgloss.NewStyle().Width(3).Align(lipgloss.Center)

			if day.Month() != m.month {
				style = style.Foreground(m.theme.TextFaint)
			} else if util.SameDay(day, now) {
				style = style.Foreground(m.theme.Today).Bold(true)
			} else if util.SameDay(day, m.selectedDay) {
				style = style.Foreground(m.theme.Selected).Bold(true)
			} else {
				style = style.Foreground(m.theme.Text)
			}

			dayCells = append(dayCells, style.Render(num))
		}
		lines = append(lines, strings.Join(dayCells, ""))
	}

	return strings.Join(lines, "\n")
}
