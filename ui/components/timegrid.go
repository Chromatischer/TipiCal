package components

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/util"
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
		sort.Slice(events[key], func(i, j int) bool {
			return events[key][i].Start.Before(events[key][j].Start)
		})
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
	headerLines := 2 + tg.allDayRowCount()
	availableRows := tg.height - headerLines
	if availableRows < 1 {
		availableRows = 1
	}
	visibleHours := availableRows / rph
	if visibleHours < 1 {
		visibleHours = 1
	}

	// All-day events don't need scroll adjustment
	if ev.AllDay {
		return
	}

	eventHour := tg.eventEffectiveHour(ev, tg.selected)
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

// SetSelectedDay changes the selected day column without rebuilding the grid.
func (tg *TimeGrid) SetSelectedDay(day time.Time) {
	tg.selected = day
	tg.selectedEventIdx = -1
}

// SelectEventByUID selects an event in the currently selected day by UID.
func (tg *TimeGrid) SelectEventByUID(uid string) {
	key := tg.selected.Format("2006-01-02")
	for i, e := range tg.events[key] {
		if e.UID == uid {
			tg.selectedEventIdx = i
			return
		}
	}
}

// HitTestAt returns the event and/or day at the given position relative to the time grid's
// content area (relX, relY are 0-indexed from the top-left of the grid). hitType is one of
// "event", "allday", "day-header", "empty", or "" (outside clickable area).
func (tg *TimeGrid) HitTestAt(relX, relY int) (event *ical.Event, day time.Time, hitType string) {
	numDays := len(tg.days)
	if numDays == 0 {
		return nil, time.Time{}, ""
	}

	timeLabelWidth := 6
	colWidth := (tg.width - timeLabelWidth - 1) / numDays
	if colWidth < 10 {
		colWidth = 10
	}
	contentStartX := timeLabelWidth + 1

	// Determine which day column was clicked
	if relX < contentStartX {
		return nil, time.Time{}, "" // clicked on time label area
	}
	colIdx := (relX - contentStartX) / colWidth
	if colIdx < 0 || colIdx >= numDays {
		return nil, time.Time{}, ""
	}
	day = tg.days[colIdx]
	key := day.Format("2006-01-02")

	headerLines := 2 + tg.allDayRowCount()

	// Row 0: day header
	if relY == 0 {
		return nil, day, "day-header"
	}
	// Row 1: thin separator
	if relY == 1 {
		return nil, time.Time{}, ""
	}
	// All-day section
	if relY < headerLines {
		// Last row of the allday section is the thick separator
		if relY == headerLines-1 {
			return nil, time.Time{}, ""
		}
		allDayRow := relY - 2
		// Collect visible all-day events (up to maxVisibleRows)
		const maxVisibleRows = 3
		count := 0
		for _, e := range tg.events[key] {
			if e.AllDay {
				if count == allDayRow {
					return e, day, "allday"
				}
				count++
				if count >= maxVisibleRows {
					break
				}
			}
		}
		return nil, day, "allday"
	}

	// Time slot area
	timeRow := relY - headerLines
	if timeRow < 0 {
		return nil, time.Time{}, ""
	}

	// Map the clicked row to (hour, subRow) using the same variable row
	// distribution as View() (some early hours may get an extra row).
	rphBase := tg.rowsPerHour()
	availableRows := tg.height - headerLines
	if availableRows < 1 {
		availableRows = 1
	}
	if timeRow >= availableRows {
		return nil, time.Time{}, ""
	}

	visibleHours := availableRows / rphBase
	if visibleHours < 1 {
		visibleHours = 1
	}
	totalHours := tg.endHour - tg.startHour - tg.scrollY
	if totalHours < 0 {
		totalHours = 0
	}
	if totalHours > visibleHours {
		totalHours = visibleHours
	}
	if totalHours == 0 {
		return nil, time.Time{}, ""
	}

	remainder := availableRows - (rphBase * totalHours)
	hour := -1
	subRow := 0
	currentRPH := rphBase
	row := timeRow
	for hourIndex := 0; hourIndex < totalHours; hourIndex++ {
		rowsThisHour := rphBase
		if hourIndex < remainder {
			rowsThisHour = rphBase + 1
		}
		if row < rowsThisHour {
			hour = tg.startHour + tg.scrollY + hourIndex
			subRow = row
			currentRPH = rowsThisHour
			break
		}
		row -= rowsThisHour
	}
	if hour < tg.startHour || hour >= tg.endHour {
		return nil, time.Time{}, ""
	}

	dayEvents := tg.events[key]
	oInfo := tg.findOverlapInfo(day, dayEvents)
	primary, secondary := tg.findCellEvents(hour, day, dayEvents, oInfo)

	// Apply sub-row filtering (same logic as renderHourCell)
	subRowStartMinute := (subRow * 60) / currentRPH
	subRowEndMinute := ((subRow + 1) * 60) / currentRPH
	subRowStart := time.Date(day.Year(), day.Month(), day.Day(), hour, subRowStartMinute, 0, 0, day.Location())
	subRowEnd := time.Date(day.Year(), day.Month(), day.Day(), hour, subRowEndMinute, 0, 0, day.Location())

	if primary != nil && util.SameDay(primary.Start, day) && !primary.Start.Before(subRowEnd) {
		primary = nil
	}
	if secondary != nil && util.SameDay(secondary.Start, day) && !secondary.Start.Before(subRowEnd) {
		secondary = nil
	}
	if primary != nil && !primary.End.After(subRowStart) {
		primary = nil
	}
	if secondary != nil && !secondary.End.After(subRowStart) {
		secondary = nil
	}

	if primary == nil && secondary == nil {
		return nil, day, "empty"
	}

	// When two events overlap, determine which one occupies the content area at the
	// clicked x position using the same logic as renderHourCell.
	if secondary != nil {
		colStartX := contentStartX + colIdx*colWidth
		xInCol := relX - colStartX

		// The sidebar bar (1 col wide) starts just after the border char (col 0 of the column)
		if xInCol <= 1 && primary != nil {
			return primary, day, "event"
		}

		// Determine content event (mirrors renderHourCell logic)
		contentEvent := secondary
		isPriStart := primary != nil && util.SameDay(primary.Start, day) &&
			!primary.Start.Before(subRowStart) && primary.Start.Before(subRowEnd)
		isSecStart := util.SameDay(secondary.Start, day) &&
			!secondary.Start.Before(subRowStart) && secondary.Start.Before(subRowEnd)
		if !util.SameDay(secondary.Start, day) {
			isSecStart = hour == tg.startHour && subRow == 0
		}
		if primary != nil && !util.SameDay(primary.Start, day) {
			isPriStart = hour == tg.startHour && subRow == 0
		}
		if isPriStart && !isSecStart {
			contentEvent = primary
		} else if isPriStart && isSecStart && primary != nil && !primary.Start.After(secondary.Start) {
			contentEvent = primary
		} else if !isPriStart && !isSecStart && primary != nil && primary.Start.Equal(secondary.Start) {
			contentEvent = primary
		}
		return contentEvent, day, "event"
	}

	return primary, day, "event"
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
	headerLines := 2 + tg.allDayRowCount()
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
	headerLines := 2 + tg.allDayRowCount() // day header + separator + all-day section
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

	// All-day events section
	allDayLines := tg.renderAllDaySection(colWidth, timeLabelWidth)
	if len(allDayLines) > 0 {
		lines = append(lines, allDayLines...)
		allDaySep := lipgloss.NewStyle().
			Foreground(tg.theme.Border).
			Render(strings.Repeat("━", tg.width))
		lines = append(lines, allDaySep)
	}

	// Pre-compute overlap info per day
	overlapInfoByDay := make(map[string]map[string]overlapInfo)
	for _, day := range tg.days {
		key := day.Format("2006-01-02")
		overlapInfoByDay[key] = tg.findOverlapInfo(day, tg.events[key])
	}

	// Time rows
	rph := tg.rowsPerHour()
	headerLines := 2 + tg.allDayRowCount()
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
					cellContent := tg.renderHourCell(h, day, dayEvents, colWidth, overlapInfoByDay[key], 0, currentRPH)
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
						cellContent := tg.renderContinuationCell(h, day, dayEvents, colWidth, overlapInfoByDay[key], r, currentRPH)
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

func (tg *TimeGrid) renderHourCell(hour int, day time.Time, events []*ical.Event, width int, oInfo map[string]overlapInfo, subRow int, rph int) string {
	primary, secondary := tg.findCellEvents(hour, day, events, oInfo)

	// Calculate sub-row time range for filtering events
	subRowStartMinute := (subRow * 60) / rph
	subRowEndMinute := ((subRow + 1) * 60) / rph
	subRowStart := time.Date(day.Year(), day.Month(), day.Day(), hour, subRowStartMinute, 0, 0, day.Location())
	subRowEnd := time.Date(day.Year(), day.Month(), day.Day(), hour, subRowEndMinute, 0, 0, day.Location())

	// Filter events not yet active at this sub-row
	if primary != nil && util.SameDay(primary.Start, day) && !primary.Start.Before(subRowEnd) {
		primary = nil
	}
	if secondary != nil && util.SameDay(secondary.Start, day) && !secondary.Start.Before(subRowEnd) {
		secondary = nil
	}
	// Filter events that have already ended before this sub-row
	if primary != nil && !primary.End.After(subRowStart) {
		primary = nil
	}
	if secondary != nil && !secondary.End.After(subRowStart) {
		secondary = nil
	}

	borderStyle := lipgloss.NewStyle().
		Foreground(tg.theme.Border).
		Render("│")

	// Overlap/indented rendering: secondary is present and column wide enough
	if secondary != nil && width >= 6 {
		// Check if events start within this sub-row
		isSecStart := util.SameDay(secondary.Start, day) &&
			!secondary.Start.Before(subRowStart) && secondary.Start.Before(subRowEnd)
		isPriStart := primary != nil && util.SameDay(primary.Start, day) &&
			!primary.Start.Before(subRowStart) && primary.Start.Before(subRowEnd)

		// Multi-day events: show title at top of grid
		if !util.SameDay(secondary.Start, day) {
			isSecStart = hour == tg.startHour && subRow == 0
		}
		if primary != nil && !util.SameDay(primary.Start, day) {
			isPriStart = hour == tg.startHour && subRow == 0
		}

		// Determine which event gets the content area vs the sidebar marker.
		// Default: secondary (indented, later-starting) takes the content area so its
		// title is visible during its overlap window; primary shows as a sidebar bar.
		// Exception 1: primary starts here but secondary doesn't — show primary title.
		// Exception 2: both start at the same time — primary (longer) always gets content.
		contentEvent := secondary
		sidebarEvent := primary
		isContentStart := isSecStart
		if isPriStart && !isSecStart {
			contentEvent = primary
			sidebarEvent = secondary
			isContentStart = isPriStart
		} else if isPriStart && isSecStart && primary != nil && !primary.Start.After(secondary.Start) {
			contentEvent = primary
			sidebarEvent = secondary
			isContentStart = isPriStart
		} else if !isPriStart && !isSecStart && primary != nil && primary.Start.Equal(secondary.Start) {
			// Continuation rows of a same-start-time pair: keep primary in content for consistency.
			contentEvent = primary
			sidebarEvent = secondary
		}

		contentColor := tg.eventColor(contentEvent)
		isContentSelected := tg.isSelectedEvent(contentEvent)
		isSidebarSelected := sidebarEvent != nil && tg.isSelectedEvent(sidebarEvent)

		// Compute extra sidebar markers for nested overlaps
		extraStr, extraCount := tg.extraSidebars(secondary, hour, day, events, oInfo)

		var sidebars string
		var overlapWidth int
		if sidebarEvent != nil {
			sidebarColor := tg.eventColor(sidebarEvent)
			sidebars = lipgloss.NewStyle().
				Foreground(sidebarColor).
				Render("▌")
			overlapWidth = 1
		}

		// Badge: +N on the bottom-most row of the content event when a same-calendar
		// event with the same start time is hidden in the sidebar.
		var overlapBadge string
		if sidebarEvent != nil && sidebarEvent.CalendarID == contentEvent.CalendarID &&
			sidebarEvent.Start.Equal(contentEvent.Start) &&
			!contentEvent.End.After(subRowEnd) {
			overlapBadge = "+1"
		} else if sidebarEvent == nil && !contentEvent.End.After(subRowEnd) {
			// Content event is alone in this cell (its overlap partner ended earlier)
			// but a same-calendar same-start-time event may still overlap it.
			for _, e := range events {
				if !e.AllDay && e.UID != contentEvent.UID &&
					e.CalendarID == contentEvent.CalendarID &&
					e.Start.Equal(contentEvent.Start) {
					overlapBadge = "+1"
					break
				}
			}
		}

		// Sidebar event selected: selection marker on left, sidebars, then content
		if isSidebarSelected {
			selLeft := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▐")
			contentWidth := width - 2 - overlapWidth - extraCount
			if contentWidth < 1 {
				contentWidth = 1
			}
			return borderStyle + selLeft + sidebars + extraStr +
				tg.buildOverlapContent(contentEvent, contentColor, isContentStart, contentWidth, overlapBadge, false)
		}

		// Content event selected
		if isContentSelected {
			contentWidth := width - 3 - overlapWidth - extraCount
			if contentWidth < 1 {
				contentWidth = 1
			}
			selLeft := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▐")
			selRight := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▌")
			return borderStyle + sidebars + extraStr + selLeft +
				tg.buildOverlapContent(contentEvent, contentColor, isContentStart, contentWidth, overlapBadge, true) +
				selRight
		}

		// No selection on either
		contentWidth := width - 1 - overlapWidth - extraCount
		if contentWidth < 1 {
			contentWidth = 1
		}
		return borderStyle + sidebars + extraStr +
			tg.buildOverlapContent(contentEvent, contentColor, isContentStart, contentWidth, overlapBadge, false)
	}

	// Single event rendering (no overlap)
	eventInCell := primary
	if eventInCell == nil {
		eventInCell = secondary
	}

	if eventInCell == nil {
		// Empty cell
		dotBorder := lipgloss.NewStyle().
			Foreground(tg.theme.Border).
			Render(strings.Repeat("·", width-1))
		return borderStyle + dotBorder
	}

	color := tg.eventColor(eventInCell)
	isSelectedEvent := tg.isSelectedEvent(eventInCell)

	// Check if event starts within this sub-row
	isEventStart := util.SameDay(eventInCell.Start, day) &&
		!eventInCell.Start.Before(subRowStart) && eventInCell.Start.Before(subRowEnd)
	// Multi-day events: show title at top of grid
	if !util.SameDay(eventInCell.Start, day) {
		isEventStart = hour == tg.startHour && subRow == 0
	}

	// Badge on the bottom row when a same-calendar same-start-time event also occupies this block
	var badge string
	if !eventInCell.End.After(subRowEnd) {
		for _, e := range events {
			if !e.AllDay && e.UID != eventInCell.UID &&
				e.CalendarID == eventInCell.CalendarID &&
				e.Start.Equal(eventInCell.Start) {
				badge = "+1"
				break
			}
		}
	}

	if isSelectedEvent {
		leftBorder := lipgloss.NewStyle().
			Foreground(tg.theme.Selected).
			Render("▐")
		rightBorder := lipgloss.NewStyle().
			Foreground(tg.theme.Selected).
			Render("▌")
		return borderStyle + leftBorder +
			tg.buildOverlapContent(eventInCell, color, isEventStart, width-3, badge, true) +
			rightBorder
	}

	return borderStyle + tg.buildOverlapContent(eventInCell, color, isEventStart, width-1, badge, false)
}

// renderContinuationCell renders a sub-row cell within an hour slot.
// It shows event continuation (colored background) or an empty cell.
func (tg *TimeGrid) renderContinuationCell(hour int, day time.Time, events []*ical.Event, width int, oInfo map[string]overlapInfo, subRow int, rph int) string {
	primary, secondary := tg.findCellEvents(hour, day, events, oInfo)

	// Calculate sub-row time range for filtering events
	subRowStartMinute := (subRow * 60) / rph
	subRowEndMinute := ((subRow + 1) * 60) / rph
	subRowStart := time.Date(day.Year(), day.Month(), day.Day(), hour, subRowStartMinute, 0, 0, day.Location())
	subRowEnd := time.Date(day.Year(), day.Month(), day.Day(), hour, subRowEndMinute, 0, 0, day.Location())

	// Filter events not yet active at this sub-row
	if primary != nil && util.SameDay(primary.Start, day) && !primary.Start.Before(subRowEnd) {
		primary = nil
	}
	if secondary != nil && util.SameDay(secondary.Start, day) && !secondary.Start.Before(subRowEnd) {
		secondary = nil
	}
	// Filter events that have already ended before this sub-row
	if primary != nil && !primary.End.After(subRowStart) {
		primary = nil
	}
	if secondary != nil && !secondary.End.After(subRowStart) {
		secondary = nil
	}

	borderStyle := lipgloss.NewStyle().
		Foreground(tg.theme.Border).
		Render("│")

	// Overlap/indented rendering
	if secondary != nil && width >= 6 {
		// Check if events start within this sub-row
		isSecStart := util.SameDay(secondary.Start, day) &&
			!secondary.Start.Before(subRowStart) && secondary.Start.Before(subRowEnd)
		isPriStart := primary != nil && util.SameDay(primary.Start, day) &&
			!primary.Start.Before(subRowStart) && primary.Start.Before(subRowEnd)

		// Multi-day events: show title at top of grid
		if !util.SameDay(secondary.Start, day) {
			isSecStart = hour == tg.startHour && subRow == 0
		}
		if primary != nil && !util.SameDay(primary.Start, day) {
			isPriStart = hour == tg.startHour && subRow == 0
		}

		// Determine which event gets the content area vs the sidebar marker.
		// Default: secondary (indented, later-starting) takes the content area so its
		// title is visible during its overlap window; primary shows as a sidebar bar.
		// Exception 1: primary starts here but secondary doesn't — show primary title.
		// Exception 2: both start at the same time — primary (longer) always gets content.
		contentEvent := secondary
		sidebarEvent := primary
		isContentStart := isSecStart
		if isPriStart && !isSecStart {
			contentEvent = primary
			sidebarEvent = secondary
			isContentStart = isPriStart
		} else if isPriStart && isSecStart && primary != nil && !primary.Start.After(secondary.Start) {
			contentEvent = primary
			sidebarEvent = secondary
			isContentStart = isPriStart
		} else if !isPriStart && !isSecStart && primary != nil && primary.Start.Equal(secondary.Start) {
			// Continuation rows of a same-start-time pair: keep primary in content for consistency.
			contentEvent = primary
			sidebarEvent = secondary
		}

		contentColor := tg.eventColor(contentEvent)
		isContentSelected := tg.isSelectedEvent(contentEvent)
		isSidebarSelected := sidebarEvent != nil && tg.isSelectedEvent(sidebarEvent)

		// Compute extra sidebar markers for nested overlaps
		extraStr, extraCount := tg.extraSidebars(secondary, hour, day, events, oInfo)

		var sidebars string
		var overlapWidth int
		if sidebarEvent != nil {
			sidebarColor := tg.eventColor(sidebarEvent)
			sidebars = lipgloss.NewStyle().
				Foreground(sidebarColor).
				Render("▌")
			overlapWidth = 1
		}

		// Badge: +N on the bottom-most row of the content event when a same-calendar
		// event with the same start time is hidden in the sidebar.
		var overlapBadge string
		if sidebarEvent != nil && sidebarEvent.CalendarID == contentEvent.CalendarID &&
			sidebarEvent.Start.Equal(contentEvent.Start) &&
			!contentEvent.End.After(subRowEnd) {
			overlapBadge = "+1"
		} else if sidebarEvent == nil && !contentEvent.End.After(subRowEnd) {
			for _, e := range events {
				if !e.AllDay && e.UID != contentEvent.UID &&
					e.CalendarID == contentEvent.CalendarID &&
					e.Start.Equal(contentEvent.Start) {
					overlapBadge = "+1"
					break
				}
			}
		}

		// Sidebar event selected: selection marker on left, sidebars, then content
		if isSidebarSelected {
			selLeft := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▐")
			contentWidth := width - 2 - overlapWidth - extraCount
			if contentWidth < 1 {
				contentWidth = 1
			}
			return borderStyle + selLeft + sidebars + extraStr +
				tg.buildOverlapContent(contentEvent, contentColor, isContentStart, contentWidth, overlapBadge, false)
		}

		// Content event selected
		if isContentSelected {
			contentWidth := width - 3 - overlapWidth - extraCount
			if contentWidth < 1 {
				contentWidth = 1
			}
			selLeft := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▐")
			selRight := lipgloss.NewStyle().
				Foreground(tg.theme.Selected).
				Render("▌")
			return borderStyle + sidebars + extraStr + selLeft +
				tg.buildOverlapContent(contentEvent, contentColor, isContentStart, contentWidth, overlapBadge, true) +
				selRight
		}

		// No selection
		contentWidth := width - 1 - overlapWidth - extraCount
		if contentWidth < 1 {
			contentWidth = 1
		}
		return borderStyle + sidebars + extraStr +
			tg.buildOverlapContent(contentEvent, contentColor, isContentStart, contentWidth, overlapBadge, false)
	}

	// Single event rendering
	eventInCell := primary
	if eventInCell == nil {
		eventInCell = secondary
	}

	if eventInCell == nil {
		// Empty continuation cell
		emptyCell := lipgloss.NewStyle().
			Width(width - 1).
			Render(strings.Repeat(" ", width-1))
		return borderStyle + emptyCell
	}

	color := tg.eventColor(eventInCell)
	isSelectedEvent := tg.isSelectedEvent(eventInCell)

	// Check if event starts within this sub-row
	isEventStart := util.SameDay(eventInCell.Start, day) &&
		!eventInCell.Start.Before(subRowStart) && eventInCell.Start.Before(subRowEnd)
	// Multi-day events: show title at top of grid
	if !util.SameDay(eventInCell.Start, day) {
		isEventStart = hour == tg.startHour && subRow == 0
	}

	// Badge on the bottom row when a same-calendar same-start-time event also occupies this block
	var badge string
	if !eventInCell.End.After(subRowEnd) {
		for _, e := range events {
			if !e.AllDay && e.UID != eventInCell.UID &&
				e.CalendarID == eventInCell.CalendarID &&
				e.Start.Equal(eventInCell.Start) {
				badge = "+1"
				break
			}
		}
	}

	if isSelectedEvent {
		leftBorder := lipgloss.NewStyle().
			Foreground(tg.theme.Selected).
			Render("▐")
		rightBorder := lipgloss.NewStyle().
			Foreground(tg.theme.Selected).
			Render("▌")
		return borderStyle + leftBorder +
			tg.buildOverlapContent(eventInCell, color, isEventStart, width-3, badge, true) +
			rightBorder
	}

	return borderStyle + tg.buildOverlapContent(eventInCell, color, isEventStart, width-1, badge, false)
}

// renderAllDaySection renders all-day events in a compact bar above the time grid.
func (tg *TimeGrid) renderAllDaySection(colWidth, timeLabelWidth int) []string {
	const maxVisibleRows = 3

	// Collect all-day events per day, separating single-day from multi-day
	hasAny := false
	allDayByDay := make(map[string][]*ical.Event)    // visible events
	strippedByDay := make(map[string]int)            // count of stripped multi-day events
	strippedColorsByDay := make(map[string][]string) // colors of stripped events

	for _, day := range tg.days {
		key := day.Format("2006-01-02")
		var allDay []*ical.Event
		for _, e := range tg.events[key] {
			if e.AllDay {
				allDay = append(allDay, e)
				hasAny = true
			}
		}

		if len(allDay) > maxVisibleRows {
			// Too many events: keep all single-day events, strip multi-day
			// ones first (they appear on other days too).
			var visible, stripped []*ical.Event
			for _, e := range allDay {
				if len(visible) < maxVisibleRows {
					visible = append(visible, e)
				} else if e.AllDaySpanDays() > 1 {
					stripped = append(stripped, e)
				} else {
					// Single-day event that doesn't fit: swap out the last
					// multi-day event in visible to make room.
					swapped := false
					for i := len(visible) - 1; i >= 0; i-- {
						if visible[i].AllDaySpanDays() > 1 {
							stripped = append(stripped, visible[i])
							visible[i] = e
							swapped = true
							break
						}
					}
					if !swapped {
						// All visible are single-day; just strip this one
						stripped = append(stripped, e)
					}
				}
			}
			allDayByDay[key] = visible
			strippedByDay[key] = len(stripped)
			for _, e := range stripped {
				c := e.Color
				if c == "" {
					c = string(tg.theme.CalendarColor(e.CalendarID))
				}
				strippedColorsByDay[key] = append(strippedColorsByDay[key], c)
			}
		} else {
			allDayByDay[key] = allDay
		}
	}
	if !hasAny {
		return nil
	}

	// Find the max number of visible all-day events on any single day
	maxEvents := 0
	hasStripped := false
	for _, day := range tg.days {
		key := day.Format("2006-01-02")
		if len(allDayByDay[key]) > maxEvents {
			maxEvents = len(allDayByDay[key])
		}
		if strippedByDay[key] > 0 {
			hasStripped = true
		}
	}

	borderStyle := lipgloss.NewStyle().
		Foreground(tg.theme.Border).
		Render("│")

	var rows []string
	for row := 0; row < maxEvents; row++ {
		labelStyle := lipgloss.NewStyle().
			Width(timeLabelWidth).
			Align(lipgloss.Right).
			Foreground(tg.theme.TextFaint)
		var label string
		if row == 0 {
			label = "ALL"
		}
		line := labelStyle.Render(label) + " "

		for _, day := range tg.days {
			key := day.Format("2006-01-02")
			evts := allDayByDay[key]
			if row < len(evts) {
				e := evts[row]
				color := lipgloss.Color(e.Color)
				if e.Color == "" {
					color = tg.theme.CalendarColor(e.CalendarID)
				}
				isSelected := tg.isSelectedEvent(e)
				if isSelected {
					title := util.TruncateText(e.Summary, colWidth-5)
					content := lipgloss.NewStyle().
						Background(color).
						Foreground(lipgloss.Color("#FFFFFF")).
						Bold(true).
						Width(colWidth-3).
						Padding(0, 1).
						Render(title)
					leftBorder := lipgloss.NewStyle().
						Foreground(tg.theme.Selected).
						Render("▐")
					rightBorder := lipgloss.NewStyle().
						Foreground(tg.theme.Selected).
						Render("▌")
					line += borderStyle + leftBorder + content + rightBorder
				} else {
					title := util.TruncateText(e.Summary, colWidth-3)
					content := lipgloss.NewStyle().
						Background(color).
						Foreground(lipgloss.Color("#FFFFFF")).
						Bold(true).
						Width(colWidth-1).
						Padding(0, 1).
						Render(title)
					line += borderStyle + content
				}
			} else {
				emptyCell := lipgloss.NewStyle().
					Width(colWidth - 1).
					Render(strings.Repeat(" ", colWidth-1))
				line += borderStyle + emptyCell
			}
		}
		rows = append(rows, line)
	}

	// Add indicator row for stripped multi-day events
	if hasStripped {
		labelStyle := lipgloss.NewStyle().
			Width(timeLabelWidth).
			Align(lipgloss.Right).
			Foreground(tg.theme.TextFaint)
		line := labelStyle.Render("") + " "

		for _, day := range tg.days {
			key := day.Format("2006-01-02")
			count := strippedByDay[key]
			if count > 0 {
				// Build colored dots for each stripped event + count.
				// Cap the number of dots so the indicator fits within the column.
				countStr := fmt.Sprintf(" +%d", count)
				maxDots := colWidth - 1 - len(countStr)
				if maxDots < 1 {
					maxDots = 1
				}
				var dots []string
				colors := strippedColorsByDay[key]
				for i, c := range colors {
					if i >= maxDots {
						break
					}
					dots = append(dots, lipgloss.NewStyle().
						Foreground(lipgloss.Color(c)).
						Render("●"))
				}
				indicator := fmt.Sprintf("%s%s", strings.Join(dots, ""), countStr)
				content := lipgloss.NewStyle().
					Foreground(tg.theme.TextFaint).
					Width(colWidth - 1).
					Align(lipgloss.Center).
					Render(indicator)
				line += borderStyle + content
			} else {
				emptyCell := lipgloss.NewStyle().
					Width(colWidth - 1).
					Render(strings.Repeat(" ", colWidth-1))
				line += borderStyle + emptyCell
			}
		}
		rows = append(rows, line)
	}

	return rows
}

// allDayRowCount returns the number of rows consumed by the all-day section (including separator).
func (tg *TimeGrid) allDayRowCount() int {
	const maxVisibleRows = 3
	maxEvents := 0
	hasStripped := false
	for _, day := range tg.days {
		key := day.Format("2006-01-02")
		count := 0
		for _, e := range tg.events[key] {
			if e.AllDay {
				count++
			}
		}
		if count > maxVisibleRows {
			hasStripped = true
			count = maxVisibleRows
		}
		if count > maxEvents {
			maxEvents = count
		}
	}
	if maxEvents == 0 {
		return 0
	}
	rows := maxEvents + 1 // +1 for the thick separator line
	if hasStripped {
		rows++ // +1 for the indicator row
	}
	return rows
}

// eventEffectiveHour returns the hour at which an event effectively starts on
// the given day. For single-day events this is Start.Hour(). For multi-day
// events viewed on a day after their start date, this returns startHour (top
// of the grid) since the event spans the entire visible range.
func (tg *TimeGrid) eventEffectiveHour(e *ical.Event, day time.Time) int {
	if util.SameDay(e.Start, day) {
		return e.Start.Hour()
	}
	// Event started on a previous day — it runs from midnight on this day,
	// so anchor it at the top of the grid.
	return tg.startHour
}

// isSelectedEvent returns true if the given event is the currently selected event.
func (tg *TimeGrid) isSelectedEvent(e *ical.Event) bool {
	sel := tg.SelectedEvent()
	if sel == nil || e == nil {
		return false
	}
	return sel.UID == e.UID
}

// eventColor returns the display color for an event.
func (tg *TimeGrid) eventColor(e *ical.Event) lipgloss.Color {
	if e.Color != "" {
		return lipgloss.Color(e.Color)
	}
	return tg.theme.CalendarColor(e.CalendarID)
}

// overlapInfo tracks whether an event should be rendered indented due to overlap.
type overlapInfo struct {
	indented         bool
	overlappingColor lipgloss.Color // color of the primary event it overlaps with (for sidebar)
	overlappingUID   string         // UID of the primary event it overlaps with
	// Extra overlaps with other indented events (for nested overlap rendering).
	// When a third event overlaps with both a primary and a secondary, these
	// track the secondary's color/UID so an additional sidebar marker can be drawn.
	extraOverlapColors []lipgloss.Color
	extraOverlapUIDs   []string
}

// buildOverlapContent renders the content area for the "content" event in an overlap
// cell. badge is non-empty when a +N indicator should appear at the right edge.
// isStart indicates this sub-row is the event's first (shows title); otherwise a
// plain coloured background is rendered.
func (tg *TimeGrid) buildOverlapContent(e *ical.Event, color lipgloss.Color, isStart bool, width int, badge string, rightPad bool) string {
	if badge != "" {
		badge = "  " + badge
		if rightPad {
			badge += " "
		}
	}
	badgeW := len(badge) // badge is always ASCII
	if badgeW > 0 && width < badgeW+3 {
		badge = ""
		badgeW = 0
	}
	if isStart {
		innerW := width - 2 // subtract Padding(0, 1)
		if badgeW > 0 {
			innerW -= badgeW
		}
		if innerW < 0 {
			innerW = 0
		}
		title := util.TruncateText(e.Summary, innerW)
		var inner string
		if badgeW > 0 {
			spaces := width - 2 - util.DisplayWidth(title) - badgeW
			if spaces < 0 {
				spaces = 0
			}
			inner = title + strings.Repeat(" ", spaces) + badge
		} else {
			inner = title
		}
		return lipgloss.NewStyle().
			Background(color).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Width(width).
			Padding(0, 1).
			Render(inner)
	}
	// Continuation row (no title)
	if badgeW > 0 {
		spaces := width - badgeW
		if spaces < 0 {
			spaces = 0
		}
		return lipgloss.NewStyle().
			Background(color).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Width(width).
			Render(strings.Repeat(" ", spaces) + badge)
	}
	return lipgloss.NewStyle().
		Background(color).
		Width(width).
		Render(strings.Repeat(" ", width))
}

// findOverlapInfo pre-computes overlap relationships for a day's events.
// An event is marked "indented" if it starts while another event is ongoing.
// Once indented, it stays indented for its entire duration.
func (tg *TimeGrid) findOverlapInfo(day time.Time, events []*ical.Event) map[string]overlapInfo {
	info := make(map[string]overlapInfo)

	var timed []*ical.Event
	for _, e := range events {
		if !e.AllDay {
			timed = append(timed, e)
		}
	}

	// First pass: mark indented status based on overlap with earlier-starting event.
	for i := 0; i < len(timed); i++ {
		for j := i + 1; j < len(timed); j++ {
			a, b := timed[i], timed[j]
			if !a.Start.Before(b.End) || !b.Start.Before(a.End) {
				continue
			}
			// Determine which event is secondary (indented) based on actual start time.
			// When start times are equal, the shorter event goes to the sidebar.
			var secondary *ical.Event
			if a.Start.Equal(b.Start) {
				if a.End.Before(b.End) {
					secondary = a
				} else {
					secondary = b
				}
			} else if !a.Start.After(b.Start) {
				secondary = b
			} else {
				secondary = a
			}
			if _, exists := info[secondary.UID]; !exists {
				info[secondary.UID] = overlapInfo{indented: true}
			}
		}
	}

	// Second pass: set overlappingColor from the first primary event each secondary overlaps.
	for _, e := range timed {
		oi, isIndented := info[e.UID]
		if !isIndented {
			continue
		}
		for _, candidate := range timed {
			if info[candidate.UID].indented {
				continue
			}
			if !candidate.Start.Before(e.End) || !e.Start.Before(candidate.End) {
				continue
			}
			oi.overlappingColor = tg.eventColor(candidate)
			oi.overlappingUID = candidate.UID
			info[e.UID] = oi
			break
		}
	}

	// Third pass: detect overlaps between indented events.
	// When two indented events overlap, the later-starting one gets an extra
	// sidebar marker for the earlier indented event's color.
	for i := 0; i < len(timed); i++ {
		for j := i + 1; j < len(timed); j++ {
			a, b := timed[i], timed[j]
			aInfo, aIndented := info[a.UID]
			bInfo, bIndented := info[b.UID]
			if !aIndented || !bIndented {
				continue
			}
			if !a.Start.Before(b.End) || !b.Start.Before(a.End) {
				continue
			}
			if !a.Start.After(b.Start) {
				bInfo.extraOverlapColors = append(bInfo.extraOverlapColors, tg.eventColor(a))
				bInfo.extraOverlapUIDs = append(bInfo.extraOverlapUIDs, a.UID)
				info[b.UID] = bInfo
			} else {
				aInfo.extraOverlapColors = append(aInfo.extraOverlapColors, tg.eventColor(b))
				aInfo.extraOverlapUIDs = append(aInfo.extraOverlapUIDs, b.UID)
				info[a.UID] = aInfo
			}
		}
	}

	return info
}

// isEventActiveInCell checks whether the event with the given UID is active
// (overlapping) in the specified hour cell.
func (tg *TimeGrid) isEventActiveInCell(uid string, hour int, day time.Time, events []*ical.Event) bool {
	cellStart := time.Date(day.Year(), day.Month(), day.Day(), hour, 0, 0, 0, day.Location())
	cellEnd := cellStart.Add(time.Hour)
	for _, e := range events {
		if e.UID == uid && !e.AllDay && e.OverlapsWith(cellStart, cellEnd) {
			return true
		}
	}
	return false
}

// extraSidebars computes extra sidebar marker strings for a secondary event that
// overlaps with other indented events. Returns the combined sidebar string and
// the number of extra markers rendered (to adjust content width).
func (tg *TimeGrid) extraSidebars(secondary *ical.Event, hour int, day time.Time, events []*ical.Event, oInfo map[string]overlapInfo) (string, int) {
	oi := oInfo[secondary.UID]
	if len(oi.extraOverlapUIDs) == 0 {
		return "", 0
	}
	var sidebars string
	count := len(oi.extraOverlapUIDs)
	sel := tg.SelectedEvent()
	for i, uid := range oi.extraOverlapUIDs {
		if tg.isEventActiveInCell(uid, hour, day, events) {
			isMarkerSelected := sel != nil && sel.UID == uid
			if isMarkerSelected {
				sidebars += lipgloss.NewStyle().
					Foreground(tg.theme.Selected).
					Render("▐")
				count++
			}
			sidebars += lipgloss.NewStyle().
				Foreground(oi.extraOverlapColors[i]).
				Render("▌")
		} else {
			sidebars += " "
		}
	}
	return sidebars, count
}

// findCellEvents finds the primary (non-indented) and secondary (indented) events
// overlapping the given hour cell.
func (tg *TimeGrid) findCellEvents(hour int, day time.Time, events []*ical.Event, oInfo map[string]overlapInfo) (primary, secondary *ical.Event) {
	cellStart := time.Date(day.Year(), day.Month(), day.Day(), hour, 0, 0, 0, day.Location())
	cellEnd := cellStart.Add(time.Hour)
	for _, e := range events {
		if e.AllDay || !e.OverlapsWith(cellStart, cellEnd) {
			continue
		}
		if oInfo[e.UID].indented {
			if secondary == nil {
				secondary = e
			} else if tg.eventEffectiveHour(e, day) == hour && tg.eventEffectiveHour(secondary, day) != hour {
				// Prefer the event starting at this hour so its title is shown;
				// the other indented event is just a continuation.
				secondary = e
			}
		} else {
			if primary == nil {
				primary = e
			}
		}
	}
	return
}
