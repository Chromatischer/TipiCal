package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
)

func testPrintTheme() *config.Theme {
	return &config.Theme{
		Accent:    lipgloss.Color("#7C3AED"),
		Text:      lipgloss.Color("#CDD6F4"),
		TextFaint: lipgloss.Color("#6C7086"),
		TextMuted: lipgloss.Color("#A6ADC8"),
		Today:     lipgloss.Color("#F38BA8"),
		Border:    lipgloss.Color("#585B70"),
		CalendarColors: []lipgloss.Color{
			lipgloss.Color("#3B82F6"),
			lipgloss.Color("#10B981"),
		},
	}
}

func testPrintConfig() *config.Config {
	return &config.Config{
		General: config.GeneralConfig{
			TimeFormat: "24h",
		},
	}
}

func TestRenderPrintAgendaEmpty(t *testing.T) {
	theme := testPrintTheme()
	cfg := testPrintConfig()
	store := ical.NewStore()

	output := renderPrintAgenda(store, theme, cfg, 2, false)

	if !strings.Contains(output, "No upcoming events") {
		t.Errorf("empty agenda should show 'No upcoming events', got: %s", output)
	}
}

func TestRenderPrintAgendaBasic(t *testing.T) {
	theme := testPrintTheme()
	cfg := testPrintConfig()
	store := ical.NewStore()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	store.AddEvent(&ical.Event{
		UID:        "test-1",
		Summary:    "Team Meeting",
		Start:      today.Add(10 * time.Hour),
		End:        today.Add(11 * time.Hour),
		CalendarID: 0,
		Color:      "#3B82F6",
	})

	output := renderPrintAgenda(store, theme, cfg, 2, false)

	if !strings.Contains(output, "Today") {
		t.Error("output should contain 'Today'")
	}
	if !strings.Contains(output, "Team Meeting") {
		t.Error("output should contain event summary")
	}
	if !strings.Contains(output, "10:00 - 11:00") {
		t.Error("output should contain formatted time range")
	}
}

func TestRenderPrintAgendaAllDay(t *testing.T) {
	theme := testPrintTheme()
	cfg := testPrintConfig()
	store := ical.NewStore()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	store.AddEvent(&ical.Event{
		UID:        "allday-1",
		Summary:    "Birthday",
		Start:      today,
		End:        today.AddDate(0, 0, 1),
		AllDay:     true,
		CalendarID: 0,
		Color:      "#7C3AED",
	})

	output := renderPrintAgenda(store, theme, cfg, 2, false)

	if !strings.Contains(output, "All day") {
		t.Errorf("all-day event should show 'All day', got: %s", output)
	}
	if !strings.Contains(output, "Birthday") {
		t.Error("output should contain all-day event summary")
	}
}

func TestRenderPrintAgendaFull(t *testing.T) {
	theme := testPrintTheme()
	cfg := testPrintConfig()
	store := ical.NewStore()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	store.AddEvent(&ical.Event{
		UID:         "full-1",
		Summary:     "Conference",
		Location:    "Room 101",
		Description: "Annual team conference",
		Start:       today.Add(14 * time.Hour),
		End:         today.Add(16 * time.Hour),
		CalendarID:  0,
		Color:       "#3B82F6",
	})

	output := renderPrintAgenda(store, theme, cfg, 2, true)

	if !strings.Contains(output, "Conference") {
		t.Error("full output should contain summary")
	}
	if !strings.Contains(output, "Room 101") {
		t.Error("full output should contain location")
	}
	if !strings.Contains(output, "Annual team conference") {
		t.Error("full output should contain description")
	}
}

func TestRenderPrintAgendaMultipleDays(t *testing.T) {
	theme := testPrintTheme()
	cfg := testPrintConfig()
	store := ical.NewStore()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	tomorrow := today.AddDate(0, 0, 1)

	store.AddEvent(&ical.Event{
		UID:        "today-event",
		Summary:    "Today Meeting",
		Start:      today.Add(10 * time.Hour),
		End:        today.Add(11 * time.Hour),
		CalendarID: 0,
	})

	store.AddEvent(&ical.Event{
		UID:        "tomorrow-event",
		Summary:    "Tomorrow Meeting",
		Start:      tomorrow.Add(14 * time.Hour),
		End:        tomorrow.Add(15 * time.Hour),
		CalendarID: 0,
	})

	output := renderPrintAgenda(store, theme, cfg, 2, false)

	if !strings.Contains(output, "Today") {
		t.Error("output should contain 'Today'")
	}
	if !strings.Contains(output, "Tomorrow") {
		t.Error("output should contain 'Tomorrow'")
	}
	if !strings.Contains(output, "Today Meeting") {
		t.Error("output should contain today's event")
	}
	if !strings.Contains(output, "Tomorrow Meeting") {
		t.Error("output should contain tomorrow's event")
	}
}

func TestRenderPrintAgendaWeek(t *testing.T) {
	theme := testPrintTheme()
	cfg := testPrintConfig()
	store := ical.NewStore()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	for i := 0; i < 7; i++ {
		day := today.AddDate(0, 0, i)
		store.AddEvent(&ical.Event{
			UID:        string(rune('a' + i)),
			Summary:    "Event Day " + string(rune('0'+i)),
			Start:      day.Add(10 * time.Hour),
			End:        day.Add(11 * time.Hour),
			CalendarID: 0,
		})
	}

	output := renderPrintAgenda(store, theme, cfg, 7, false)

	dayCount := strings.Count(output, "────────")
	if dayCount < 6 {
		t.Errorf("week output should have ~7 day sections, got %d dividers", dayCount)
	}
}

func TestRenderPrintAgendaSortedByTime(t *testing.T) {
	theme := testPrintTheme()
	cfg := testPrintConfig()
	store := ical.NewStore()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	store.AddEvent(&ical.Event{
		UID:        "late",
		Summary:    "Late Event",
		Start:      today.Add(16 * time.Hour),
		End:        today.Add(17 * time.Hour),
		CalendarID: 0,
	})

	store.AddEvent(&ical.Event{
		UID:        "early",
		Summary:    "Early Event",
		Start:      today.Add(8 * time.Hour),
		End:        today.Add(9 * time.Hour),
		CalendarID: 0,
	})

	output := renderPrintAgenda(store, theme, cfg, 2, false)

	earlyIdx := strings.Index(output, "Early Event")
	lateIdx := strings.Index(output, "Late Event")

	if earlyIdx == -1 || lateIdx == -1 {
		t.Fatal("output should contain both events")
	}
	if earlyIdx > lateIdx {
		t.Error("events should be sorted by time (early before late)")
	}
}

func TestRenderPrintAgendaFutureDay(t *testing.T) {
	theme := testPrintTheme()
	cfg := testPrintConfig()
	store := ical.NewStore()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	future := today.AddDate(0, 0, 5)

	store.AddEvent(&ical.Event{
		UID:        "future-event",
		Summary:    "Future Event",
		Start:      future.Add(10 * time.Hour),
		End:        future.Add(11 * time.Hour),
		CalendarID: 0,
	})

	output := renderPrintAgenda(store, theme, cfg, 7, false)

	if !strings.Contains(output, "Future Event") {
		t.Error("output should contain future event")
	}

	dayNames := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	hasDayName := false
	for _, day := range dayNames {
		if strings.Contains(output, day) {
			hasDayName = true
			break
		}
	}
	if !hasDayName {
		t.Error("future day should show day name (e.g., 'Monday')")
	}
}
