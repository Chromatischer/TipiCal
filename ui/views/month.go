package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui/components"
	"github.com/tipical/tipical/util"
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

	colWidth := mv.width/7 - 2
	if colWidth < 10 {
		colWidth = 10
	}

	cellHeight := (mv.height-4)/6 - 2 // 6 rows, minus header; subtract 2 for borders
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
			Width(colWidth + 2). // match cell width including borders
			Align(lipgloss.Center).
			Foreground(mv.theme.TextMuted).
			Bold(true).
			Render(name)
		headerCells = append(headerCells, cell)
	}
	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))

	// Separator
	totalWidth := 7 * (colWidth + 2) // each cell adds 2 for borders
	sep := lipgloss.NewStyle().
		Foreground(mv.theme.Border).
		Render(strings.Repeat("─", totalWidth))
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

	// Determine cell background color for selected cells
	var cellBg lipgloss.Color
	if isSelected {
		cellBg = mv.theme.SurfaceAlt
	}

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

	if cellBg != "" {
		numStyle = numStyle.Background(cellBg)
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
			moreStyle := lipgloss.NewStyle().
				Foreground(mv.theme.TextFaint)
			if cellBg != "" {
				moreStyle = moreStyle.Background(cellBg)
			}
			cellLines = append(cellLines, moreStyle.Render(moreStr))
			break
		}
		cellLines = append(cellLines, components.EventDot(mv.theme, e, eventWidth, mv.use24h, cellBg))
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
