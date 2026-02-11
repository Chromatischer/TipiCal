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

	// Event selection: index into the selected day's events (-1 = no selection)
	selectedEventIdx int
}

// NewTimeGrid creates a new time grid.
func NewTimeGrid(theme *config.Theme, days []time.Time, store *ical.Store, use24h bool) *TimeGrid {
	events := make(map[string][]*ical.Event)
	for _, day := range days {
		key := day.Format("2006-01-02")
		events[key] = store.EventsForDay(day)
	}

	return &TimeGrid{
		theme:            theme,
		days:             days,
		events:           events,
		startHour:        7,
		endHour:          20,
		width:            80,
		height:           30,
		scrollY:          0,
		use24h:           use24h,
		selected:         days[0],
		selectedEventIdx: -1,
	}
}

// SetSelected sets which day column is the selected day.
func (tg *TimeGrid) SetSelected(day time.Time) {
	tg.selected = day
}

// SelectNextEvent moves event selection to the next event on the selected day.
func (tg *TimeGrid) SelectNextEvent() {
	key := tg.selected.Format("2006-01-02")
	dayEvents := tg.events[key]
	if len(dayEvents) == 0 {
		tg.selectedEventIdx = -1
		return
	}
	if tg.selectedEventIdx < len(dayEvents)-1 {
		tg.selectedEventIdx++
	}
	tg.scrollToSelectedEvent()
}

// SelectPrevEvent moves event selection to the previous event on the selected day.
func (tg *TimeGrid) SelectPrevEvent() {
	key := tg.selected.Format("2006-01-02")
	dayEvents := tg.events[key]
	if len(dayEvents) == 0 {
		tg.selectedEventIdx = -1
		return
	}
	if tg.selectedEventIdx > 0 {
		tg.selectedEventIdx--
	}
	tg.scrollToSelectedEvent()
}

// ResetEventSelection resets event selection to the first event (or -1 if none).
func (tg *TimeGrid) ResetEventSelection() {
	key := tg.selected.Format("2006-01-02")
	dayEvents := tg.events[key]
	if len(dayEvents) > 0 {
		tg.selectedEventIdx = 0
	} else {
		tg.selectedEventIdx = -1
	}
}

// SelectedEvent returns the currently selected event, or nil if none.
func (tg *TimeGrid) SelectedEvent() *ical.Event {
	key := tg.selected.Format("2006-01-02")
	dayEvents := tg.events[key]
	if tg.selectedEventIdx >= 0 && tg.selectedEventIdx < len(dayEvents) {
		return dayEvents[tg.selectedEventIdx]
	}
	return nil
}

// scrollToSelectedEvent adjusts scrollY so the selected event is visible.
func (tg *TimeGrid) scrollToSelectedEvent() {
	ev := tg.SelectedEvent()
	if ev == nil {
		return
	}

	rph := tg.rowsPerHour()
	headerLines := 2
	availableRows := tg.height - headerLines
	if availableRows < 1 {
		availableRows = 1
	}
	visibleHours := availableRows / rph
	if visibleHours < 1 {
		visibleHours = 1
	}

	eventHour := ev.Start.Hour()
	viewStart := tg.startHour + tg.scrollY
	viewEnd := viewStart + visibleHours

	if eventHour < viewStart {
		tg.scrollY = eventHour - tg.startHour
	} else if eventHour >= viewEnd {
		tg.scrollY = eventHour - tg.startHour - visibleHours + 1
	}

	// Clamp scrollY
	maxScroll := (tg.endHour - tg.startHour) - visibleHours
	if maxScroll < 0 {
		maxScroll = 0
	}
	if tg.scrollY < 0 {
		tg.scrollY = 0
	}
	if tg.scrollY > maxScroll {
		tg.scrollY = maxScroll
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
	rph := tg.rowsPerHour()
	headerLines := 2
	availableRows := tg.height - headerLines
	if availableRows < 1 {
		availableRows = 1
	}
	visibleHours := availableRows / rph
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

// rowsPerHour calculates how many terminal rows each hour slot should occupy
// to fill the available vertical space.
func (tg *TimeGrid) rowsPerHour() int {
	headerLines := 2 // day header + separator
	availableRows := tg.height - headerLines
	if availableRows < 1 {
		availableRows = 1
	}
	totalHours := tg.endHour - tg.startHour
	if totalHours < 1 {
		totalHours = 1
	}
	rph := availableRows / totalHours
	if rph < 1 {
		rph = 1
	}
	return rph
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

		isSelected := util.SameDay(day, tg.selected) && numDays > 1
		isToday := util.SameDay(day, now)

		style := lipgloss.NewStyle().
			Width(colWidth).
			Align(lipgloss.Center).
			Bold(true)

		if isSelected && isToday {
			style = style.Foreground(tg.theme.Today).
				Underline(true)
		} else if isSelected {
			style = style.Foreground(tg.theme.Selected).
				Underline(true)
		} else if isToday {
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
	rph := tg.rowsPerHour()
	headerLines := 2
	availableRows := tg.height - headerLines
	if availableRows < 1 {
		availableRows = 1
	}

	// Calculate how many hours we can display based on scroll and available rows
	visibleHours := availableRows / rph
	if visibleHours < 1 {
		visibleHours = 1
	}

	rowsRendered := 0
	for h := tg.startHour + tg.scrollY; h < tg.endHour && h < tg.startHour+tg.scrollY+visibleHours; h++ {
		// Determine how many rows this hour gets
		// Give extra rows to earlier hours to distribute remainder
		hourIndex := h - tg.startHour - tg.scrollY
		currentRPH := rph
		totalHours := tg.endHour - tg.startHour - tg.scrollY
		if totalHours > visibleHours {
			totalHours = visibleHours
		}
		remainder := availableRows - (rph * totalHours)
		if hourIndex < remainder {
			currentRPH = rph + 1
		}

		for r := 0; r < currentRPH && rowsRendered < availableRows; r++ {
			if r == 0 {
				// First row: show time label
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
			} else {
				// Continuation rows: empty time label, continuation cells
				emptyLabel := lipgloss.NewStyle().
					Width(timeLabelWidth).
					Render("")

				row := emptyLabel + " "

				// Now indicator: place it at the sub-row corresponding to current minute
				nowIndicatorRow := -1
				if h == now.Hour() {
					for _, day := range tg.days {
						if util.SameDay(now, day) {
							// Map minute 0-59 to sub-rows 1..(currentRPH-1)
							nowIndicatorRow = (now.Minute() * (currentRPH - 1)) / 60
							if nowIndicatorRow < 1 {
								nowIndicatorRow = 1
							}
							break
						}
					}
				}

				if nowIndicatorRow == r {
					// Render now indicator line
					nowMin := now.Minute()
					indicator := fmt.Sprintf("%s%s",
						strings.Repeat(" ", timeLabelWidth+1),
						lipgloss.NewStyle().Foreground(tg.theme.Today).Render(
							fmt.Sprintf("──%02d──", nowMin)+strings.Repeat("─", tg.width-timeLabelWidth-8),
						),
					)
					lines = append(lines, indicator)
				} else {
					// Render continuation cells for each day column
					for _, day := range tg.days {
						key := day.Format("2006-01-02")
						dayEvents := tg.events[key]
						cellContent := tg.renderContinuationCell(h, day, dayEvents, colWidth)
						row += cellContent
					}
					lines = append(lines, row)
				}
			}
			rowsRendered++
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

		isSelectedEvent := tg.isSelectedEvent(eventInCell)

		// Check if this is the start hour of the event
		if eventInCell.Start.Hour() == hour {
			if isSelectedEvent {
				// Selected event: render with outline using Selected color border
				title := util.TruncateText(eventInCell.Summary, width-5)
				content := lipgloss.NewStyle().
					Background(color).
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true).
					Width(width-3).
					Padding(0, 1).
					Render(title)
				leftBorder := lipgloss.NewStyle().
					Foreground(tg.theme.Selected).
					Render("▐")
				rightBorder := lipgloss.NewStyle().
					Foreground(tg.theme.Selected).
					Render("▌")
				return borderStyle + leftBorder + content + rightBorder
			}

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
		if isSelectedEvent {
			content := lipgloss.NewStyle().
				Background(color).
				Width(width - 3).
				Render(strings.Repeat(" ", width-3))
			leftBorder := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▐")
			rightBorder := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▌")
			return borderStyle + leftBorder + content + rightBorder
		}

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

// renderContinuationCell renders a sub-row cell within an hour slot.
// It shows event continuation (colored background) or an empty cell.
func (tg *TimeGrid) renderContinuationCell(hour int, day time.Time, events []*ical.Event, width int) string {
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

		isSelectedEvent := tg.isSelectedEvent(eventInCell)

		if isSelectedEvent {
			content := lipgloss.NewStyle().
				Background(color).
				Width(width - 3).
				Render(strings.Repeat(" ", width-3))
			leftBorder := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▐")
			rightBorder := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▌")
			return borderStyle + leftBorder + content + rightBorder
		}

		content := lipgloss.NewStyle().
			Background(color).
			Width(width - 1).
			Render(strings.Repeat(" ", width-1))
		return borderStyle + content
	}

	// Empty continuation cell
	emptyCell := lipgloss.NewStyle().
		Width(width - 1).
		Render(strings.Repeat(" ", width-1))
	return borderStyle + emptyCell
}

// isSelectedEvent returns true if the given event is the currently selected event.
func (tg *TimeGrid) isSelectedEvent(e *ical.Event) bool {
	sel := tg.SelectedEvent()
	if sel == nil || e == nil {
		return false
	}
	return sel.UID == e.UID
}
