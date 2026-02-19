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

type AgendaView struct {
	theme       *config.Theme
	store       *ical.Store
	selectedDay time.Time
	width       int
	height      int
	use24h      bool
	agendaDays  int
	scrollY     int
}

func NewAgendaView(theme *config.Theme, store *ical.Store, cfg *config.Config) *AgendaView {
	return &AgendaView{
		theme:       theme,
		store:       store,
		selectedDay: time.Now(),
		use24h:      cfg.Use24h(),
		agendaDays:  cfg.General.AgendaDays,
	}
}

func (av *AgendaView) SetSize(w, h int) {
	av.width = w
	av.height = h
}

func (av *AgendaView) SelectedDate() time.Time {
	return av.selectedDay
}

func (av *AgendaView) SetDate(d time.Time) {
	av.selectedDay = d
	av.scrollY = 0
}

func (av *AgendaView) NextPeriod() {
	av.selectedDay = av.selectedDay.AddDate(0, 0, 7)
	av.scrollY = 0
}

func (av *AgendaView) PrevPeriod() {
	av.selectedDay = av.selectedDay.AddDate(0, 0, -7)
	av.scrollY = 0
}

func (av *AgendaView) MoveUp() {
	if av.scrollY > 0 {
		av.scrollY--
	}
}

func (av *AgendaView) MoveDown() {
	maxScroll := av.maxScroll()
	if av.scrollY < maxScroll {
		av.scrollY++
	}
}

func (av *AgendaView) MoveLeft() {
	av.PrevPeriod()
}

func (av *AgendaView) MoveRight() {
	av.NextPeriod()
}

func (av *AgendaView) maxScroll() int {
	total := av.countTotalLines()
	maxScroll := total - av.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

func (av *AgendaView) countTotalLines() int {
	startDate := time.Date(av.selectedDay.Year(), av.selectedDay.Month(), av.selectedDay.Day(),
		0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 0, av.agendaDays)

	total := 0
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		events := av.store.EventsForDay(d)
		if len(events) == 0 {
			continue
		}

		total += 2
		total += len(events)
		total++
	}
	return total
}

func (av *AgendaView) View() string {
	startDate := time.Date(av.selectedDay.Year(), av.selectedDay.Month(), av.selectedDay.Day(),
		0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 0, av.agendaDays)
	now := time.Now()

	var lines []string
	lineCount := 0

dayLoop:
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		events := av.store.EventsForDay(d)
		if len(events) == 0 {
			continue
		}

		sort.Slice(events, func(i, j int) bool {
			return events[i].Start.Before(events[j].Start)
		})

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

			if len(lines) >= av.height {
				break dayLoop
			}
		}

		if lineCount >= av.scrollY {
			lines = append(lines, "")
		}
		lineCount++

		if len(lines) >= av.height {
			break dayLoop
		}
	}

	if len(lines) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(av.theme.TextFaint).
			Render("\n  No upcoming events")
		return emptyMsg
	}

	return lipgloss.NewStyle().
		Width(av.width).
		Height(av.height).
		MaxHeight(av.height).
		Render(strings.Join(lines, "\n"))
}
