package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ical"
	"github.com/terminal-ical/terminal-ical/util"
)

// TimeGrid renders a vertical time grid with events for one or more days.
type TimeGrid struct {
	theme     *config.Theme
	days      []time.Time
	events    map[string][]*ical.Event // key: date string
	startHour int
	endHour   int
	width     int
	height    int
	scrollY   int
	use24h    bool
	selected  time.Time
}

// NewTimeGrid creates a new time grid.
func NewTimeGrid(theme *config.Theme, days []time.Time, store *ical.Store, use24h bool) *TimeGrid {
	events := make(map[string][]*ical.Event)
	for _, day := range days {
		key := day.Format("2006-01-02")
		events[key] = store.EventsForDay(day)
	}

	return &TimeGrid{
		theme:     theme,
		days:      days,
		events:    events,
		startHour: 7,
		endHour:   20,
		width:     80,
		height:    30,
		scrollY:   0,
		use24h:    use24h,
		selected:  days[0],
	}
}

// SetSize sets the grid dimensions.
func (tg *TimeGrid) SetSize(w, h int) {
	tg.width = w
	tg.height = h
}

// ScrollUp scrolls the time grid up.
func (tg *TimeGrid) ScrollUp() {
	if tg.scrollY > 0 {
		tg.scrollY--
	}
}

// ScrollDown scrolls the time grid down.
func (tg *TimeGrid) ScrollDown() {
	visibleHours := tg.height - 4
	if visibleHours < 1 {
		visibleHours = 1
	}
	maxScroll := (tg.endHour - tg.startHour) - visibleHours
	if maxScroll < 0 {
		maxScroll = 0
	}
	if tg.scrollY < maxScroll {
		tg.scrollY++
	}
}

// View renders the time grid.
func (tg *TimeGrid) View() string {
	numDays := len(tg.days)
	if numDays == 0 {
		return ""
	}

	timeLabelWidth := 6
	colWidth := (tg.width - timeLabelWidth - 1) / numDays
	if colWidth < 10 {
		colWidth = 10
	}

	var lines []string

	// Day headers
	var headerCells []string
	headerCells = append(headerCells, strings.Repeat(" ", timeLabelWidth+1))
	now := time.Now()

	for _, day := range tg.days {
		var label string
		if numDays == 1 {
			label = day.Format("Monday, Jan 2")
		} else {
			label = day.Format("Mon 2")
		}

		style := lipgloss.NewStyle().
			Width(colWidth).
			Align(lipgloss.Center).
			Bold(true)

		if util.SameDay(day, now) {
			style = style.Foreground(tg.theme.Today)
		} else {
			style = style.Foreground(tg.theme.Text)
		}

		headerCells = append(headerCells, style.Render(label))
	}
	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))

	// Separator
	sep := lipgloss.NewStyle().
		Foreground(tg.theme.Border).
		Render(strings.Repeat("─", tg.width))
	lines = append(lines, sep)

	// Time rows
	visibleHours := tg.height - 4 // account for header + separator + padding
	if visibleHours < 1 {
		visibleHours = 1
	}

	for h := tg.startHour + tg.scrollY; h < tg.endHour && h < tg.startHour+tg.scrollY+visibleHours; h++ {
		var timeLabel string
		if tg.use24h {
			timeLabel = fmt.Sprintf("%02d:00", h)
		} else {
			if h == 0 {
				timeLabel = "12 AM"
			} else if h < 12 {
				timeLabel = fmt.Sprintf("%d AM", h)
			} else if h == 12 {
				timeLabel = "12 PM"
			} else {
				timeLabel = fmt.Sprintf("%d PM", h-12)
			}
		}

		timeLabelStyle := lipgloss.NewStyle().
			Width(timeLabelWidth).
			Align(lipgloss.Right).
			Foreground(tg.theme.TextFaint)

		row := timeLabelStyle.Render(timeLabel) + " "

		// For each day column, render events or empty slot
		for _, day := range tg.days {
			key := day.Format("2006-01-02")
			dayEvents := tg.events[key]

			cellContent := tg.renderHourCell(h, day, dayEvents, colWidth)
			row += cellContent
		}

		lines = append(lines, row)

		// Now indicator
		if util.SameDay(now, tg.days[0]) || (numDays > 1 && h == now.Hour()) {
			if h == now.Hour() {
				nowMin := now.Minute()
				indicator := fmt.Sprintf("%s%s%s",
					strings.Repeat(" ", timeLabelWidth+1),
					lipgloss.NewStyle().Foreground(tg.theme.Today).Render(
						fmt.Sprintf("──%02d──", nowMin)+strings.Repeat("─", tg.width-timeLabelWidth-8),
					),
					"",
				)
				lines = append(lines, indicator)
			}
		}
	}

	result := strings.Join(lines, "\n")

	// Pad/clip output to exactly fill the available height
	return lipgloss.NewStyle().
		Width(tg.width).
		Height(tg.height).
		MaxHeight(tg.height).
		Render(result)
}

func (tg *TimeGrid) renderHourCell(hour int, day time.Time, events []*ical.Event, width int) string {
	cellStart := time.Date(day.Year(), day.Month(), day.Day(), hour, 0, 0, 0, day.Location())
	cellEnd := cellStart.Add(time.Hour)

	var eventInCell *ical.Event
	for _, e := range events {
		if e.OverlapsWith(cellStart, cellEnd) {
			eventInCell = e
			break
		}
	}

	borderStyle := lipgloss.NewStyle().
		Foreground(tg.theme.Border).
		Render("│")

	if eventInCell != nil {
		color := lipgloss.Color(eventInCell.Color)
		if eventInCell.Color == "" {
			color = tg.theme.CalendarColor(eventInCell.CalendarID)
		}

		// Check if this is the start hour of the event
		if eventInCell.Start.Hour() == hour {
			title := util.TruncateText(eventInCell.Summary, width-3)
			content := lipgloss.NewStyle().
				Background(color).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Width(width-1).
				Padding(0, 1).
				Render(title)
			return borderStyle + content
		}

		// Continuation of event
		content := lipgloss.NewStyle().
			Background(color).
			Width(width - 1).
			Render(strings.Repeat(" ", width-1))
		return borderStyle + content
	}

	// Empty cell
	dotBorder := lipgloss.NewStyle().
		Foreground(tg.theme.Border).
		Render(strings.Repeat("·", width-1))
	return borderStyle + dotBorder
}
