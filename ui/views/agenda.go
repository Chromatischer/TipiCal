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

func styleWrap(color lipgloss.Color) (string, string) {
	rendered := lipgloss.NewStyle().Foreground(color).Render("X")
	parts := strings.SplitN(rendered, "X", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
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

	indent := 20
	textWidth := av.width - indent
	if textWidth < 20 {
		textWidth = 20
	}
	maxURLLen := textWidth
	if maxURLLen > 40 {
		maxURLLen = 40
	}
	linkPrefix, _ := styleWrap(av.theme.Link)
	locPrefix, _ := styleWrap(av.theme.TextFaint)
	descPrefix, _ := styleWrap(av.theme.TextMuted)

	total := 0
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		events := av.store.EventsForDay(d)
		if len(events) == 0 {
			continue
		}

		total += 2
		for _, e := range events {
			total++
			if e.Location != "" {
				locFormatted := util.FormatWithHyperlinksStyled(e.Location, maxURLLen, linkPrefix, locPrefix)
				wrapped := util.WrapTextANSI(locFormatted, textWidth)
				total += len(wrapped)
			}
			if e.Description != "" {
				descFormatted := util.FormatWithHyperlinksStyled(e.Description, maxURLLen, linkPrefix, descPrefix)
				wrapped := util.WrapTextANSI(descFormatted, textWidth)
				total += len(wrapped)
			}
		}
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
	linkPrefix, _ := styleWrap(av.theme.Link)
	locPrefix, locSuffix := styleWrap(av.theme.TextFaint)
	descPrefix, descSuffix := styleWrap(av.theme.TextMuted)

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

		lines = append(lines, dateHeader)
		divider := lipgloss.NewStyle().
			Foreground(av.theme.Border).
			Render("  " + strings.Repeat("─", av.width-6))
		lines = append(lines, divider)

		for _, e := range events {
			color := lipgloss.Color(e.Color)
			if e.Color == "" {
				color = av.theme.CalendarColor(e.CalendarID)
			}

			dot := lipgloss.NewStyle().Foreground(color).Render("●")
			var timeStr string
			if e.AllDay {
				timeStr = "All day"
			} else {
				timeStr = util.FormatTimeRange(e.Start, e.End, av.use24h)
			}
			timeRendered := lipgloss.NewStyle().
				Foreground(av.theme.TextMuted).
				Width(14).
				Render(timeStr)
			title := lipgloss.NewStyle().
				Foreground(av.theme.Text).
				Bold(true).
				Render(e.Summary)

			eventLine := fmt.Sprintf("  %s %s  %s", dot, timeRendered, title)

			lines = append(lines, eventLine)

			indent := 20
			textWidth := av.width - indent
			if textWidth < 20 {
				textWidth = 20
			}
			maxURLLen := textWidth
			if maxURLLen > 40 {
				maxURLLen = 40
			}

			if e.Location != "" {
				locFormatted := util.FormatWithHyperlinksStyled(e.Location, maxURLLen, linkPrefix, locPrefix)
				wrapped := util.WrapTextANSI(locFormatted, textWidth)

				for _, line := range wrapped {
					indentedLine := strings.Repeat(" ", indent) + locPrefix + line + locSuffix
					lines = append(lines, indentedLine)
				}
			}

			if e.Description != "" {
				descFormatted := util.FormatWithHyperlinksStyled(e.Description, maxURLLen, linkPrefix, descPrefix)
				wrapped := util.WrapTextANSI(descFormatted, textWidth)

				for _, line := range wrapped {
					indentedLine := strings.Repeat(" ", indent) + descPrefix + line + descSuffix
					lines = append(lines, indentedLine)
				}
			}
		}

		lines = append(lines, "")
	}

	if len(lines) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(av.theme.TextFaint).
			Render("\n  No upcoming events")
		return emptyMsg
	}

	start := av.scrollY
	if start > len(lines) {
		start = len(lines)
	}
	end := start + av.height
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[start:end]

	return lipgloss.NewStyle().
		Width(av.width).
		Height(av.height).
		MaxHeight(av.height).
		Render(strings.Join(visible, "\n"))
}
