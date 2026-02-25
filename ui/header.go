package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
)

// waveTickMsg triggers the wave animation update.
type waveTickMsg struct{}

// headerHitZone is a clickable region within the header bar.
type headerHitZone struct {
	x1, x2 int
	action  string
}

// Header renders the top bar with current date range and view switcher.
type Header struct {
	styles       *Styles
	width        int
	viewType     config.ViewType
	date         time.Time
	wavePosition int // Current position of the wave effect
	hitZones     []headerHitZone
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

// WaveTick returns a command that triggers the wave animation.
func (h *Header) WaveTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return waveTickMsg{}
	})
}

// UpdateWave advances the wave animation position.
func (h *Header) UpdateWave() {
	h.wavePosition = (h.wavePosition + 1) % 20 // Cycle through 20 positions
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
	}

	navTitle := fmt.Sprintf("  %s  %s  %s", "◀", dateStr, "▶")
	title := lipgloss.NewStyle().
		Foreground(h.styles.Theme.Text).
		Bold(true).
		Render(navTitle)

	// View tabs
	tabs, tabWidths := h.renderTabs()

	// Logo with wave effect
	logo := h.renderLogoWithWave()

	// Layout: logo | title ... tabs
	logoWidth := lipgloss.Width(logo)
	titleWidth := logoWidth + lipgloss.Width(title)
	tabsWidth := lipgloss.Width(tabs)
	gap := h.width - titleWidth - tabsWidth - 2
	if gap < 0 {
		gap = 0
	}

	// Populate hit zones (header has Padding(0,1) so content starts at screen col 1)
	const leftPad = 1
	navTitleStartX := leftPad + logoWidth
	// navTitle = "  ◀  dateStr  ▶": ◀ at offset 2, ▶ at last char
	arrowLeftX := navTitleStartX + 2
	arrowRightX := navTitleStartX + lipgloss.Width(title) - 1
	tabsStartX := h.width - leftPad - tabsWidth
	viewActions := []string{"view:month", "view:week", "view:threeday", "view:day", "view:agenda"}
	h.hitZones = []headerHitZone{
		{x1: arrowLeftX - 1, x2: arrowLeftX + 1, action: "prev"},
		{x1: arrowRightX - 1, x2: arrowRightX + 1, action: "next"},
	}
	tx := tabsStartX
	for i, w := range tabWidths {
		h.hitZones = append(h.hitZones, headerHitZone{x1: tx, x2: tx + w - 1, action: viewActions[i]})
		tx += w
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

// HitTest returns the action string for a click at terminal column x in the header row.
func (h *Header) HitTest(x int) string {
	for _, zone := range h.hitZones {
		if x >= zone.x1 && x <= zone.x2 {
			return zone.action
		}
	}
	return ""
}

// renderLogoWithWave renders the TipiCal logo with a wave animation effect.
func (h *Header) renderLogoWithWave() string {
	text := "  TipiCal"
	waveWidth := 3 // Width of the wave highlight

	var result strings.Builder
	for i, ch := range text {
		// Calculate distance from wave center
		dist := (i - h.wavePosition + len(text)) % len(text)

		// Check if this character is in the wave
		inWave := dist < waveWidth

		style := lipgloss.NewStyle().Bold(true)
		if inWave {
			// Brighter color for characters in the wave
			style = style.Foreground(lipgloss.Color("#D0D0D0"))
		} else {
			// Normal accent color
			style = style.Foreground(h.styles.Theme.Accent)
		}

		result.WriteString(style.Render(string(ch)))
	}

	return result.String()
}

func (h *Header) renderTabs() (string, []int) {
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
	}

	var parts []string
	var widths []int
	for _, t := range tabs {
		label := fmt.Sprintf("[%s]", t.key)
		var rendered string
		if h.viewType == t.vt {
			rendered = h.styles.ViewTabActive.Render(label)
		} else {
			rendered = h.styles.ViewTab.Render(label)
		}
		parts = append(parts, rendered)
		widths = append(widths, lipgloss.Width(rendered))
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...), widths
}
