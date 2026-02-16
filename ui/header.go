package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
)

// Header renders the top bar with current date range and view switcher.
type Header struct {
	styles   *Styles
	width    int
	viewType config.ViewType
	date     time.Time
}

// NewHeader creates a new header component.
func NewHeader(styles *Styles, viewType config.ViewType, date time.Time) *Header {
	return &Header{
		styles:   styles,
		viewType: viewType,
		date:     date,
	}
}

// SetWidth sets the header width.
func (h *Header) SetWidth(w int) {
	h.width = w
}

// SetView updates the current view type.
func (h *Header) SetView(v config.ViewType) {
	h.viewType = v
}

// SetDate updates the displayed date.
func (h *Header) SetDate(d time.Time) {
	h.date = d
}

// View renders the header.
func (h *Header) View() string {
	// Date title based on current view
	var dateStr string
	switch h.viewType {
	case config.ViewMonth:
		dateStr = h.date.Format("January 2006")
	case config.ViewWeek:
		dateStr = fmt.Sprintf("Week of %s", h.date.Format("Jan 2, 2006"))
	case config.ViewThreeDay:
		end := h.date.AddDate(0, 0, 2)
		dateStr = fmt.Sprintf("%s - %s", h.date.Format("Jan 2"), end.Format("Jan 2, 2006"))
	case config.ViewDay:
		dateStr = h.date.Format("Monday, January 2, 2006")
	case config.ViewAgenda:
		dateStr = "Upcoming Events"
	case config.ViewStacked:
		dateStr = "Events"
	}

	navTitle := fmt.Sprintf("  %s  %s  %s", "◀", dateStr, "▶")
	title := lipgloss.NewStyle().
		Foreground(h.styles.Theme.Text).
		Bold(true).
		Render(navTitle)

	// View tabs
	tabs := h.renderTabs()

	// Logo
	logo := lipgloss.NewStyle().
		Foreground(h.styles.Theme.Accent).
		Bold(true).
		Render("  ical")

	// Layout: logo | title ... tabs
	titleWidth := lipgloss.Width(logo) + lipgloss.Width(title)
	tabsWidth := lipgloss.Width(tabs)
	gap := h.width - titleWidth - tabsWidth - 2
	if gap < 0 {
		gap = 0
	}

	bar := lipgloss.JoinHorizontal(
		lipgloss.Center,
		logo,
		title,
		strings.Repeat(" ", gap),
		tabs,
	)

	return h.styles.Header.Width(h.width).Render(bar)
}

func (h *Header) renderTabs() string {
	type tab struct {
		key   string
		label string
		vt    config.ViewType
	}

	tabs := []tab{
		{"M", "Month", config.ViewMonth},
		{"W", "Week", config.ViewWeek},
		{"3", "3-Day", config.ViewThreeDay},
		{"D", "Day", config.ViewDay},
		{"A", "Agenda", config.ViewAgenda},
		{"S", "Stack", config.ViewStacked},
	}

	var parts []string
	for _, t := range tabs {
		label := fmt.Sprintf("[%s]", t.key)
		if h.viewType == t.vt {
			parts = append(parts, h.styles.ViewTabActive.Render(label))
		} else {
			parts = append(parts, h.styles.ViewTab.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}
