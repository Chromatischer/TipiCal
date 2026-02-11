package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ical"
	"github.com/terminal-ical/terminal-ical/ui/components"
	"github.com/terminal-ical/terminal-ical/util"
)

// MonthView renders the traditional month grid calendar.
type MonthView struct {
	theme       *config.Theme
	store       *ical.Store
	year        int
	month       time.Month
	selectedDay time.Time
	width       int
	height      int
	use24h      bool
	startMonday bool
	cursor      int // selected cell index (0-41)
}

// NewMonthView creates a new month view.
func NewMonthView(theme *config.Theme, store *ical.Store, cfg *config.Config) *MonthView {
	now := time.Now()
	return &MonthView{
		theme:       theme,
		store:       store,
		year:        now.Year(),
		month:       now.Month(),
		selectedDay: now,
		use24h:      cfg.Use24h(),
		startMonday: cfg.StartMonday(),
	}
}

// SetSize sets the view dimensions.
func (mv *MonthView) SetSize(w, h int) {
	mv.width = w
	mv.height = h
}

// SelectedDate returns the currently selected date.
func (mv *MonthView) SelectedDate() time.Time {
	return mv.selectedDay
}

// SetDate navigates to a specific date.
func (mv *MonthView) SetDate(d time.Time) {
	mv.year = d.Year()
	mv.month = d.Month()
	mv.selectedDay = d
}

// NextPeriod advances to the next month.
func (mv *MonthView) NextPeriod() {
	if mv.month == time.December {
		mv.month = time.January
		mv.year++
	} else {
		mv.month++
	}
	mv.selectedDay = time.Date(mv.year, mv.month, 1, 0, 0, 0, 0, time.Local)
}

// PrevPeriod goes to the previous month.
func (mv *MonthView) PrevPeriod() {
	if mv.month == time.January {
		mv.month = time.December
		mv.year--
	} else {
		mv.month--
	}
	mv.selectedDay = time.Date(mv.year, mv.month, 1, 0, 0, 0, 0, time.Local)
}

// MoveUp moves selection up one row.
func (mv *MonthView) MoveUp() {
	mv.selectedDay = mv.selectedDay.AddDate(0, 0, -7)
	mv.updateMonth()
}

// MoveDown moves selection down one row.
func (mv *MonthView) MoveDown() {
	mv.selectedDay = mv.selectedDay.AddDate(0, 0, 7)
	mv.updateMonth()
}

// MoveLeft moves selection left one day.
func (mv *MonthView) MoveLeft() {
	mv.selectedDay = mv.selectedDay.AddDate(0, 0, -1)
	mv.updateMonth()
}

// MoveRight moves selection right one day.
func (mv *MonthView) MoveRight() {
	mv.selectedDay = mv.selectedDay.AddDate(0, 0, 1)
	mv.updateMonth()
}

func (mv *MonthView) updateMonth() {
	mv.year = mv.selectedDay.Year()
	mv.month = mv.selectedDay.Month()
}

// View renders the month grid.
func (mv *MonthView) View() string {
	grid := util.MonthGrid(mv.year, mv.month, mv.startMonday)
	now := time.Now()

	colWidth := (mv.width - 2) / 7
	if colWidth < 12 {
		colWidth = 12
	}

	cellHeight := (mv.height - 4) / 6 // 6 rows, minus header
	if cellHeight < 3 {
		cellHeight = 3
	}

	var lines []string

	// Weekday headers
	var dayNames []string
	if mv.startMonday {
		dayNames = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	} else {
		dayNames = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	}

	var headerCells []string
	for _, name := range dayNames {
		cell := lipgloss.NewStyle().
			Width(colWidth).
			Align(lipgloss.Center).
			Foreground(mv.theme.TextMuted).
			Bold(true).
			Render(name)
		headerCells = append(headerCells, cell)
	}
	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))

	// Separator
	sep := lipgloss.NewStyle().
		Foreground(mv.theme.Border).
		Render(strings.Repeat("─", mv.width-2))
	lines = append(lines, sep)

	// Day cells
	for _, week := range grid {
		var rowCells []string

		for _, day := range week {
			cell := mv.renderDayCell(day, now, colWidth, cellHeight)
			rowCells = append(rowCells, cell)
		}

		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}

	return strings.Join(lines, "\n")
}

func (mv *MonthView) renderDayCell(day, now time.Time, width, height int) string {
	isToday := util.SameDay(day, now)
	isSelected := util.SameDay(day, mv.selectedDay)
	isOtherMonth := day.Month() != mv.month

	// Day number
	numStr := fmt.Sprintf("%d", day.Day())
	numStyle := lipgloss.NewStyle()

	if isToday {
		numStyle = numStyle.Foreground(mv.theme.Today).Bold(true)
	} else if isOtherMonth {
		numStyle = numStyle.Foreground(mv.theme.TextFaint)
	} else {
		numStyle = numStyle.Foreground(mv.theme.Text)
	}

	// Events for this day
	events := mv.store.EventsForDay(day)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Start.Before(events[j].Start)
	})

	var cellLines []string
	cellLines = append(cellLines, numStyle.Render(numStr))

	maxEvents := height - 2 // Reserve lines for day number and potential overflow
	if maxEvents < 1 {
		maxEvents = 1
	}

	eventWidth := width - 2
	if eventWidth < 5 {
		eventWidth = 5
	}

	for i, e := range events {
		if i >= maxEvents {
			remaining := len(events) - maxEvents
			moreStr := fmt.Sprintf("+%d more", remaining)
			cellLines = append(cellLines, lipgloss.NewStyle().
				Foreground(mv.theme.TextFaint).
				Render(moreStr))
			break
		}
		cellLines = append(cellLines, components.EventDot(mv.theme, e, eventWidth, mv.use24h))
	}

	// Pad to height
	for len(cellLines) < height-1 {
		cellLines = append(cellLines, "")
	}

	content := strings.Join(cellLines, "\n")

	// Cell style
	cellStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1)

	if isSelected && isToday {
		cellStyle = cellStyle.
			Background(mv.theme.SurfaceAlt).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mv.theme.Today)
	} else if isSelected {
		cellStyle = cellStyle.
			Background(mv.theme.SurfaceAlt).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mv.theme.Selected)
	} else if isToday {
		cellStyle = cellStyle.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mv.theme.Today)
	} else if isOtherMonth {
		cellStyle = cellStyle.
			BorderStyle(lipgloss.HiddenBorder())
	} else {
		cellStyle = cellStyle.
			BorderStyle(lipgloss.HiddenBorder())
	}

	return cellStyle.Render(content)
}
