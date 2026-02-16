package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/util"
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
// bg is an optional background color; pass "" to use no background.
func EventDot(theme *config.Theme, event *ical.Event, maxWidth int, use24h bool, bg ...lipgloss.Color) string {
	color := lipgloss.Color(event.Color)
	if event.Color == "" {
		color = theme.CalendarColor(event.CalendarID)
	}

	dotStyle := lipgloss.NewStyle().Foreground(color)
	timeStyle := lipgloss.NewStyle().Foreground(theme.TextMuted)
	titleStyle := lipgloss.NewStyle().Foreground(theme.Text)

	if len(bg) > 0 && bg[0] != "" {
		dotStyle = dotStyle.Background(bg[0])
		timeStyle = timeStyle.Background(bg[0])
		titleStyle = titleStyle.Background(bg[0])
	}

	dot := dotStyle.Render("●")

	timeStr := util.FormatTime(event.Start, use24h)
	title := util.TruncateText(event.Summary, maxWidth-len(timeStr)-3)

	sep := " "
	if len(bg) > 0 && bg[0] != "" {
		sep = lipgloss.NewStyle().Background(bg[0]).Render(" ")
	}

	return dot + sep + timeStyle.Render(timeStr) + sep + titleStyle.Render(title)
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
