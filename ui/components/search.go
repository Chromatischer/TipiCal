package components

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ical"
	"github.com/terminal-ical/terminal-ical/util"
)

// Search provides event search functionality.
type Search struct {
	theme   *config.Theme
	store   *ical.Store
	query   string
	results []*ical.Event
	cursor  int
	width   int
	height  int
	active  bool
	use24h  bool
}

// NewSearch creates a new search component.
func NewSearch(theme *config.Theme, store *ical.Store, use24h bool) *Search {
	return &Search{
		theme:  theme,
		store:  store,
		use24h: use24h,
	}
}

// Open activates the search.
func (s *Search) Open() {
	s.active = true
	s.query = ""
	s.results = nil
	s.cursor = 0
}

// Close deactivates the search.
func (s *Search) Close() {
	s.active = false
}

// IsActive returns whether search is active.
func (s *Search) IsActive() bool {
	return s.active
}

// SelectedEvent returns the currently highlighted event.
func (s *Search) SelectedEvent() *ical.Event {
	if s.cursor >= 0 && s.cursor < len(s.results) {
		return s.results[s.cursor]
	}
	return nil
}

// SetSize sets the search panel dimensions.
func (s *Search) SetSize(w, h int) {
	s.width = w
	s.height = h
}

// HandleKey processes a key press.
func (s *Search) HandleKey(key string) bool {
	switch key {
	case "esc":
		s.Close()
		return true
	case "enter":
		// Select current result
		return true
	case "up", "ctrl+k":
		if s.cursor > 0 {
			s.cursor--
		}
		return false
	case "down", "ctrl+j":
		if s.cursor < len(s.results)-1 {
			s.cursor++
		}
		return false
	case "backspace":
		if len(s.query) > 0 {
			s.query = s.query[:len(s.query)-1]
			s.search()
		}
		return false
	default:
		if len(key) == 1 || (len(key) > 0 && !strings.HasPrefix(key, "ctrl+")) {
			s.query += key
			s.search()
		}
		return false
	}
}

func (s *Search) search() {
	if s.query == "" {
		s.results = nil
		s.cursor = 0
		return
	}

	queryLower := strings.ToLower(s.query)
	s.results = nil

	for _, e := range s.store.Events {
		if strings.Contains(strings.ToLower(e.Summary), queryLower) ||
			strings.Contains(strings.ToLower(e.Location), queryLower) ||
			strings.Contains(strings.ToLower(e.Description), queryLower) {
			s.results = append(s.results, e)
		}
	}

	// Sort by start time
	sort.Slice(s.results, func(i, j int) bool {
		return s.results[i].Start.Before(s.results[j].Start)
	})

	s.cursor = 0
}

// View renders the search panel.
func (s *Search) View() string {
	searchWidth := s.width - 4
	if searchWidth < 30 {
		searchWidth = 30
	}
	if searchWidth > 80 {
		searchWidth = 80
	}

	// Search input
	promptStyle := lipgloss.NewStyle().
		Foreground(s.theme.Accent).
		Bold(true)
	inputStyle := lipgloss.NewStyle().
		Foreground(s.theme.Text)

	prompt := promptStyle.Render(" /") + " " + inputStyle.Render(s.query+"█")

	inputBar := lipgloss.NewStyle().
		Width(searchWidth).
		Background(s.theme.Surface).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(s.theme.Accent).
		Render(prompt)

	var lines []string
	lines = append(lines, inputBar)

	// Results
	now := time.Now()
	maxResults := s.height - 6
	if maxResults < 1 {
		maxResults = 1
	}

	if len(s.results) == 0 && s.query != "" {
		noResults := lipgloss.NewStyle().
			Foreground(s.theme.TextFaint).
			Padding(1, 1).
			Render("No events found")
		lines = append(lines, noResults)
	}

	for i, event := range s.results {
		if i >= maxResults {
			remaining := len(s.results) - maxResults
			more := lipgloss.NewStyle().
				Foreground(s.theme.TextFaint).
				Padding(0, 1).
				Render(strings.Repeat(" ", 2) + "..." + strings.Repeat(" ", 2) + string(rune('0'+remaining%10)) + " more")
			_ = more
			break
		}

		isSelected := i == s.cursor
		color := lipgloss.Color(event.Color)
		if event.Color == "" {
			color = s.theme.CalendarColor(event.CalendarID)
		}

		dot := lipgloss.NewStyle().Foreground(color).Render("●")

		// Date
		var dateStr string
		if util.SameDay(event.Start, now) {
			dateStr = "Today"
		} else if util.SameDay(event.Start, now.AddDate(0, 0, 1)) {
			dateStr = "Tomorrow"
		} else {
			dateStr = event.Start.Format("Jan 2")
		}

		timeStr := util.FormatTime(event.Start, s.use24h)
		title := util.TruncateText(event.Summary, searchWidth-25)

		style := lipgloss.NewStyle().Width(searchWidth).Padding(0, 1)
		if isSelected {
			style = style.Background(s.theme.SurfaceAlt)
		}

		dateStyle := lipgloss.NewStyle().Foreground(s.theme.TextMuted).Width(10)
		timeStyle := lipgloss.NewStyle().Foreground(s.theme.TextFaint).Width(6)
		titleStyle := lipgloss.NewStyle().Foreground(s.theme.Text)
		if isSelected {
			titleStyle = titleStyle.Bold(true)
		}

		line := style.Render(
			dot + " " +
				dateStyle.Render(dateStr) +
				timeStyle.Render(timeStr) + " " +
				titleStyle.Render(title),
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	modalStyle := lipgloss.NewStyle().
		Width(searchWidth+4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.Accent).
		Padding(0, 1)

	return lipgloss.Place(
		s.width,
		s.height,
		lipgloss.Center,
		lipgloss.Top,
		modalStyle.Render(content),
	)
}
