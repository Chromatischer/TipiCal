package views

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

// StackedView renders events as full-width cards stacked vertically.
type StackedView struct {
	theme       *config.Theme
	store       *ical.Store
	selectedDay time.Time
	width       int
	height      int
	use24h      bool
	agendaDays  int
	scrollY     int
}

// NewStackedView creates a new stacked view.
func NewStackedView(theme *config.Theme, store *ical.Store, cfg *config.Config) *StackedView {
	return &StackedView{
		theme:       theme,
		store:       store,
		selectedDay: time.Now(),
		use24h:      cfg.Use24h(),
		agendaDays:  cfg.General.AgendaDays,
	}
}

// SetSize sets the view dimensions.
func (sv *StackedView) SetSize(w, h int) {
	sv.width = w
	sv.height = h
}

// SelectedDate returns the currently selected date.
func (sv *StackedView) SelectedDate() time.Time {
	return sv.selectedDay
}

// SetDate navigates to a specific date.
func (sv *StackedView) SetDate(d time.Time) {
	sv.selectedDay = d
	sv.scrollY = 0
}

// NextPeriod advances by a week.
func (sv *StackedView) NextPeriod() {
	sv.selectedDay = sv.selectedDay.AddDate(0, 0, 7)
}

// PrevPeriod goes back a week.
func (sv *StackedView) PrevPeriod() {
	sv.selectedDay = sv.selectedDay.AddDate(0, 0, -7)
}

// MoveUp scrolls up.
func (sv *StackedView) MoveUp() {
	if sv.scrollY > 0 {
		sv.scrollY--
	}
}

// MoveDown scrolls down.
func (sv *StackedView) MoveDown() {
	sv.scrollY++
}

// MoveLeft goes to previous day.
func (sv *StackedView) MoveLeft() {
	sv.PrevPeriod()
}

// MoveRight goes to next day.
func (sv *StackedView) MoveRight() {
	sv.NextPeriod()
}

// View renders the stacked view.
func (sv *StackedView) View() string {
	startDate := time.Date(sv.selectedDay.Year(), sv.selectedDay.Month(), sv.selectedDay.Day(),
		0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 0, sv.agendaDays)
	now := time.Now()

	cardWidth := sv.width - 4
	if cardWidth < 20 {
		cardWidth = 20
	}

	var allLines []string

	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		events := sv.store.EventsForDay(d)
		if len(events) == 0 {
			continue
		}

		sort.Slice(events, func(i, j int) bool {
			return events[i].Start.Before(events[j].Start)
		})

		// Day header
		var dayLabel string
		if util.SameDay(d, now) {
			dayLabel = "Today"
		} else if util.SameDay(d, now.AddDate(0, 0, 1)) {
			dayLabel = "Tomorrow"
		} else {
			dayLabel = d.Format("Monday, Jan 2")
		}

		isToday := util.SameDay(d, now)
		headerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
		if isToday {
			headerStyle = headerStyle.Foreground(sv.theme.Today)
		} else {
			headerStyle = headerStyle.Foreground(sv.theme.Accent)
		}

		allLines = append(allLines, headerStyle.Render(dayLabel))

		// Event cards
		for _, e := range events {
			card := sv.renderCard(e, cardWidth)
			allLines = append(allLines, card)
		}

		allLines = append(allLines, "") // spacing
	}

	// Apply scroll
	if sv.scrollY >= len(allLines) {
		sv.scrollY = len(allLines) - 1
		if sv.scrollY < 0 {
			sv.scrollY = 0
		}
	}

	visibleLines := allLines
	if sv.scrollY < len(allLines) {
		visibleLines = allLines[sv.scrollY:]
	}

	// Trim to height
	result := strings.Join(visibleLines, "\n")
	outputLines := strings.Split(result, "\n")
	if len(outputLines) > sv.height-2 {
		outputLines = outputLines[:sv.height-2]
	}

	if len(allLines) == 0 {
		return lipgloss.NewStyle().
			Foreground(sv.theme.TextFaint).
			Render("\n  No upcoming events")
	}

	return strings.Join(outputLines, "\n")
}

func (sv *StackedView) renderCard(e *ical.Event, width int) string {
	color := lipgloss.Color(e.Color)
	if e.Color == "" {
		color = sv.theme.CalendarColor(e.CalendarID)
	}

	bar := lipgloss.NewStyle().Foreground(color).Render("▌")

	title := lipgloss.NewStyle().
		Foreground(sv.theme.Text).
		Bold(true).
		Render(util.TruncateText(e.Summary, width-4))

	timeStr := util.FormatTimeRange(e.Start, e.End, sv.use24h)
	timeRendered := lipgloss.NewStyle().
		Foreground(sv.theme.TextMuted).
		Render(timeStr)

	// First line: bar + title + time right-aligned
	titleWidth := lipgloss.Width(title)
	timeWidth := lipgloss.Width(timeRendered)
	gap := width - titleWidth - timeWidth - 6
	if gap < 1 {
		gap = 1
	}

	line1 := fmt.Sprintf(" %s %s%s%s", bar, title, strings.Repeat(" ", gap), timeRendered)

	var contentLines []string
	contentLines = append(contentLines, line1)

	if e.Location != "" {
		loc := lipgloss.NewStyle().
			Foreground(sv.theme.TextFaint).
			Render(fmt.Sprintf("     %s", util.TruncateText(e.Location, width-8)))
		contentLines = append(contentLines, loc)
	}

	if e.Description != "" {
		desc := lipgloss.NewStyle().
			Foreground(sv.theme.TextFaint).
			Render(fmt.Sprintf("     %s", util.TruncateText(e.Description, width-8)))
		contentLines = append(contentLines, desc)
	}

	content := strings.Join(contentLines, "\n")

	cardStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 0)

	return "  " + cardStyle.Render(content)
}
