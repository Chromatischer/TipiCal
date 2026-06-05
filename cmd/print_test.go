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

func testStore() *ical.Store {
	store := ical.NewStore()
	store.RegisterCalendar(ical.CalendarInfo{ID: 0, Name: "Work", Source: "Main"})
	store.RegisterCalendar(ical.CalendarInfo{ID: 1, Name: "Uni", Source: "Main"})

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	store.AddEvent(&ical.Event{
		UID:            "rec-1",
		Summary:        "Lecture",
		Description:    "Systemprogrammierung",
		Location:       "Auditorium",
		Start:          today.Add(9 * time.Hour),
		End:            today.Add(10 * time.Hour),
		CalendarID:     1,
		Status:         "CONFIRMED",
		Recurring:      true,
		RecurrenceRule: "FREQ=DAILY",
	})
	store.AddEvent(&ical.Event{
		UID:            "rec-1",
		Summary:        "Lecture",
		Description:    "Systemprogrammierung",
		Location:       "Auditorium",
		Start:          today.AddDate(0, 0, 1).Add(9 * time.Hour),
		End:            today.AddDate(0, 0, 1).Add(10 * time.Hour),
		CalendarID:     1,
		Status:         "CONFIRMED",
		Recurring:      true,
		RecurrenceRule: "FREQ=DAILY",
	})
	store.AddEvent(&ical.Event{
		UID:         "work-1",
		Summary:     "Standup",
		Description: "Daily sync",
		Start:       today.Add(11 * time.Hour),
		End:         today.Add(1130 * time.Minute),
		CalendarID:  0,
		Status:      "CONFIRMED",
	})
	store.AddEvent(&ical.Event{
		UID:        "all-day-1",
		Summary:    "Holiday",
		Start:      today,
		End:        today.AddDate(0, 0, 1),
		AllDay:     true,
		CalendarID: 0,
		Status:     "CONFIRMED",
	})
	return store
}

func TestParsePrintOptions(t *testing.T) {
	opts, err := parsePrintOptions([]string{"week", "--calendar", "Uni", "--search", "system", "--from", "2026-04-10", "--to", "2026-04-12", "--format", "summary,start,calendar"})
	if err != nil {
		t.Fatalf("parsePrintOptions() error = %v", err)
	}
	if opts.mode != "week" {
		t.Fatalf("mode = %q, want week", opts.mode)
	}
	if opts.numDays != 7 {
		t.Fatalf("numDays = %d, want 7", opts.numDays)
	}
	if len(opts.calendarFilters) != 1 || opts.calendarFilters[0] != "Uni" {
		t.Fatalf("calendarFilters = %#v, want [Uni]", opts.calendarFilters)
	}
	if len(opts.formatFields) != 3 {
		t.Fatalf("formatFields = %#v, want 3 fields", opts.formatFields)
	}
}

func TestFilteredEventsCalendarAndSearch(t *testing.T) {
	store := testStore()
	opts := &printOptions{
		numDays:         2,
		calendarFilters: []string{"uni"},
		search:          "systemprogrammierung",
	}

	events, err := filteredEvents(store, opts)
	if err != nil {
		t.Fatalf("filteredEvents() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	for _, rec := range events {
		if rec.Calendar.Name != "Uni" {
			t.Fatalf("calendar = %q, want Uni", rec.Calendar.Name)
		}
	}
}

func TestFilteredEventsNormalizeRecurring(t *testing.T) {
	store := testStore()
	opts := &printOptions{
		numDays:            2,
		calendarFilters:    []string{"uni"},
		normalizeRecurring: true,
	}

	events, err := filteredEvents(store, opts)
	if err != nil {
		t.Fatalf("filteredEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].Occurrences != 2 {
		t.Fatalf("occurrences = %d, want 2", events[0].Occurrences)
	}
}

func TestRenderFormattedEvents(t *testing.T) {
	store := testStore()
	events, err := filteredEvents(store, &printOptions{numDays: 1, calendarFilters: []string{"work"}})
	if err != nil {
		t.Fatalf("filteredEvents() error = %v", err)
	}

	output := renderFormattedEvents(events, true, []string{"summary", "calendar", "time"})
	if !strings.Contains(output, "Standup\tWork\t11:00 - 18:50") && !strings.Contains(output, "Standup\tWork\t11:00 - 11:30") {
		t.Fatalf("formatted output missing expected row: %q", output)
	}
}

func TestRenderPrintJSONFlatWhenFiltered(t *testing.T) {
	store := testStore()
	events, err := filteredEvents(store, &printOptions{numDays: 2, search: "lecture"})
	if err != nil {
		t.Fatalf("filteredEvents() error = %v", err)
	}

	output := renderPrintJSON(events, &printOptions{numDays: 2, search: "lecture", showJSON: true})
	if !strings.Contains(output, "\"events\"") {
		t.Fatalf("json output should contain flat events list: %s", output)
	}
	if strings.Contains(output, "\"days\"") {
		t.Fatalf("json output should not contain day wrappers when filtered: %s", output)
	}
}

func TestRenderPrintAgendaIncludesCalendar(t *testing.T) {
	store := testStore()
	events, err := filteredEvents(store, &printOptions{numDays: 1})
	if err != nil {
		t.Fatalf("filteredEvents() error = %v", err)
	}

	output := renderPrintAgenda(events, testPrintTheme(), true, false)
	if !strings.Contains(output, "[Work]") {
		t.Fatalf("agenda output should include calendar name, got: %s", output)
	}
}

func TestBuildCreatedEvent(t *testing.T) {
	store := testStore()
	event, err := buildCreatedEvent(store, &eventMutationOptions{
		calendar: "Work",
		summary:  "New Event",
		date:     "2026-04-13",
		start:    "09:00",
		end:      "10:30",
	})
	if err != nil {
		t.Fatalf("buildCreatedEvent() error = %v", err)
	}
	if event.CalendarID != 0 {
		t.Fatalf("CalendarID = %d, want 0", event.CalendarID)
	}
	if event.UID == "" {
		t.Fatal("UID should be generated")
	}
}

func TestMutateEventMove(t *testing.T) {
	store := testStore()
	original := store.FindEvent("work-1")
	moved, err := mutateEvent(store, original, &eventMutationOptions{
		date:  "2026-04-20",
		start: "14:00",
		end:   "15:00",
	}, true)
	if err != nil {
		t.Fatalf("mutateEvent() error = %v", err)
	}
	if moved.Start.Format("2006-01-02 15:04") != "2026-04-20 14:00" {
		t.Fatalf("moved start = %s", moved.Start.Format("2006-01-02 15:04"))
	}
	if moved.End.Format("15:04") != "15:00" {
		t.Fatalf("moved end = %s, want 15:00", moved.End.Format("15:04"))
	}
}
