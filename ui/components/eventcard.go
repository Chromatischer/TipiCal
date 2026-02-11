package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ical"
	"github.com/terminal-ical/terminal-ical/util"
)

// EventCard renders a single event as a styled card.
type EventCard struct {
	theme  *config.Theme
	event  *ical.Event
	width  int
	use24h bool
}

// NewEventCard creates a new event card.
func NewEventCard(theme *config.Theme, event *ical.Event, width int, use24h bool) *EventCard {
	return &EventCard{
		theme:  theme,
		event:  event,
		width:  width,
		use24h: use24h,
	}
}

// View renders the event card.
func (ec *EventCard) View() string {
	e := ec.event
	color := lipgloss.Color(e.Color)
	if e.Color == "" {
		color = ec.theme.CalendarColor(e.CalendarID)
	}

	// Color bar on left
	bar := lipgloss.NewStyle().
		Foreground(color).
		Render("▌")

	// Title
	title := lipgloss.NewStyle().
		Foreground(ec.theme.Text).
		Bold(true).
		Render(util.TruncateText(e.Summary, ec.width-4))

	// Time
	timeStr := util.FormatTimeRange(e.Start, e.End, ec.use24h)
	timeRendered := lipgloss.NewStyle().
		Foreground(ec.theme.TextMuted).
		Render(timeStr)

	// Location
	var locLine string
	if e.Location != "" {
		locLine = lipgloss.NewStyle().
			Foreground(ec.theme.TextFaint).
			Render(fmt.Sprintf("  %s", util.TruncateText(e.Location, ec.width-6)))
	}

	// Build card content
	content := fmt.Sprintf("%s %s  %s", bar, title, timeRendered)
	if locLine != "" {
		content += "\n" + locLine
	}

	cardStyle := lipgloss.NewStyle().
		Width(ec.width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1)

	return cardStyle.Render(content)
}

// EventDot renders a compact inline event indicator for month view cells.
func EventDot(theme *config.Theme, event *ical.Event, maxWidth int, use24h bool) string {
	color := lipgloss.Color(event.Color)
	if event.Color == "" {
		color = theme.CalendarColor(event.CalendarID)
	}

	dot := lipgloss.NewStyle().Foreground(color).Render("●")

	timeStr := util.FormatTime(event.Start, use24h)
	title := util.TruncateText(event.Summary, maxWidth-len(timeStr)-3)

	return fmt.Sprintf("%s %s %s",
		dot,
		lipgloss.NewStyle().Foreground(theme.TextMuted).Render(timeStr),
		lipgloss.NewStyle().Foreground(theme.Text).Render(title),
	)
}

// EventBlock renders an event as a colored block for time grid views.
func EventBlock(theme *config.Theme, event *ical.Event, width int, use24h bool) string {
	color := lipgloss.Color(event.Color)
	if event.Color == "" {
		color = theme.CalendarColor(event.CalendarID)
	}

	title := util.TruncateText(event.Summary, width-2)
	timeStr := util.FormatTimeRange(event.Start, event.End, use24h)

	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Render(title)

	timeRendered := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Render(timeStr)

	block := lipgloss.NewStyle().
		Background(color).
		Foreground(lipgloss.Color("#FFFFFF")).
		Width(width).
		Padding(0, 1).
		Render(content + "\n" + timeRendered)

	return block
}
