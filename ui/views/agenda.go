package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ical"
	"github.com/terminal-ical/terminal-ical/util"
)

// AgendaView renders a chronological list of upcoming events.
type AgendaView struct {
	theme       *config.Theme
	store       *ical.Store
	selectedDay time.Time
	width       int
	height      int
	use24h      bool
	agendaDays  int
	scrollY     int
	cursor      int
}

// NewAgendaView creates a new agenda view.
func NewAgendaView(theme *config.Theme, store *ical.Store, cfg *config.Config) *AgendaView {
	return &AgendaView{
		theme:       theme,
		store:       store,
		selectedDay: time.Now(),
		use24h:      cfg.Use24h(),
		agendaDays:  cfg.General.AgendaDays,
	}
}

// SetSize sets the view dimensions.
func (av *AgendaView) SetSize(w, h int) {
	av.width = w
	av.height = h
}

// SelectedDate returns the currently selected date.
func (av *AgendaView) SelectedDate() time.Time {
	return av.selectedDay
}

// SetDate navigates to a specific date.
func (av *AgendaView) SetDate(d time.Time) {
	av.selectedDay = d
	av.scrollY = 0
}

// NextPeriod advances the agenda start by a week.
func (av *AgendaView) NextPeriod() {
	av.selectedDay = av.selectedDay.AddDate(0, 0, 7)
}

// PrevPeriod goes back a week.
func (av *AgendaView) PrevPeriod() {
	av.selectedDay = av.selectedDay.AddDate(0, 0, -7)
}

// MoveUp moves cursor up.
func (av *AgendaView) MoveUp() {
	if av.scrollY > 0 {
		av.scrollY--
	}
}

// MoveDown moves cursor down.
func (av *AgendaView) MoveDown() {
	av.scrollY++
}

// MoveLeft is a no-op for agenda.
func (av *AgendaView) MoveLeft() {
	av.PrevPeriod()
}

// MoveRight is a no-op for agenda.
func (av *AgendaView) MoveRight() {
	av.NextPeriod()
}

// View renders the agenda view.
func (av *AgendaView) View() string {
	startDate := time.Date(av.selectedDay.Year(), av.selectedDay.Month(), av.selectedDay.Day(),
		0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 0, av.agendaDays)
	now := time.Now()

	var lines []string
	lineCount := 0

	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		events := av.store.EventsForDay(d)
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

		headerStyle := lipgloss.NewStyle().Bold(true)
		if isToday {
			headerStyle = headerStyle.Foreground(av.theme.Today)
		} else {
			headerStyle = headerStyle.Foreground(av.theme.Accent)
		}

		dateHeader := headerStyle.Render(fmt.Sprintf("  %s — %s",
			dayLabel, d.Format("January 2")))

		if lineCount >= av.scrollY {
			lines = append(lines, dateHeader)
			divider := lipgloss.NewStyle().
				Foreground(av.theme.Border).
				Render("  " + strings.Repeat("─", av.width-6))
			lines = append(lines, divider)
		}
		lineCount += 2

		// Events
		for _, e := range events {
			if lineCount < av.scrollY {
				lineCount++
				continue
			}

			color := lipgloss.Color(e.Color)
			if e.Color == "" {
				color = av.theme.CalendarColor(e.CalendarID)
			}

			dot := lipgloss.NewStyle().Foreground(color).Render("●")
			timeStr := util.FormatTimeRange(e.Start, e.End, av.use24h)
			timeRendered := lipgloss.NewStyle().
				Foreground(av.theme.TextMuted).
				Width(14).
				Render(timeStr)
			title := lipgloss.NewStyle().
				Foreground(av.theme.Text).
				Bold(true).
				Render(e.Summary)

			eventLine := fmt.Sprintf("  %s %s  %s", dot, timeRendered, title)

			if e.Location != "" {
				loc := lipgloss.NewStyle().
					Foreground(av.theme.TextFaint).
					Render(fmt.Sprintf("  %s", e.Location))
				eventLine += loc
			}

			lines = append(lines, eventLine)
			lineCount++

			if len(lines) >= av.height-2 {
				break
			}
		}

		if lineCount >= av.scrollY {
			lines = append(lines, "") // spacing between days
		}
		lineCount++

		if len(lines) >= av.height-2 {
			break
		}
	}

	if len(lines) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(av.theme.TextFaint).
			Render("\n  No upcoming events")
		return emptyMsg
	}

	return strings.Join(lines, "\n")
}
