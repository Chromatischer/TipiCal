package caldav

import (
	"strings"
	"testing"
	"time"

	goical "github.com/emersion/go-ical"
	"github.com/tipical/tipical/ical"
)

func makeCalendar(veventLines string) *goical.Calendar {
	ics := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\n" +
		veventLines +
		"END:VEVENT\r\nEND:VCALENDAR\r\n"
	dec := goical.NewDecoder(strings.NewReader(ics))
	cal, err := dec.Decode()
	if err != nil {
		panic("bad test ical: " + err.Error())
	}
	return cal
}

func TestParseCalendarObject(t *testing.T) {
	tests := []struct {
		name          string
		vevent        string
		calendarIndex int
		wantLen       int
		checkFunc     func(t *testing.T, events []*ical.Event)
	}{
		{
			name: "timed event with all fields",
			vevent: "UID:test-uid\r\n" +
				"SUMMARY:Test Event\r\n" +
				"DESCRIPTION:Test Description\r\n" +
				"LOCATION:Test Location\r\n" +
				"STATUS:CONFIRMED\r\n" +
				"DTSTART:20260115T100000Z\r\n" +
				"DTEND:20260115T110000Z\r\n",
			calendarIndex: 3,
			wantLen:       1,
			checkFunc: func(t *testing.T, events []*ical.Event) {
				e := events[0]
				if e.UID != "test-uid" {
					t.Errorf("UID = %q, want %q", e.UID, "test-uid")
				}
				if e.Summary != "Test Event" {
					t.Errorf("Summary = %q, want %q", e.Summary, "Test Event")
				}
				if e.Description != "Test Description" {
					t.Errorf("Description = %q, want %q", e.Description, "Test Description")
				}
				if e.Location != "Test Location" {
					t.Errorf("Location = %q, want %q", e.Location, "Test Location")
				}
				if e.Status != "CONFIRMED" {
					t.Errorf("Status = %q, want %q", e.Status, "CONFIRMED")
				}
				if e.AllDay {
					t.Error("AllDay should be false for timed event")
				}
				if e.CalendarID != 3 {
					t.Errorf("CalendarID = %d, want 3", e.CalendarID)
				}
				wantStart := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
				if !e.Start.Equal(wantStart) {
					t.Errorf("Start = %v, want %v", e.Start, wantStart)
				}
				wantEnd := time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC)
				if !e.End.Equal(wantEnd) {
					t.Errorf("End = %v, want %v", e.End, wantEnd)
				}
			},
		},
		{
			name: "all-day event (VALUE=DATE)",
			vevent: "UID:allday-uid\r\n" +
				"SUMMARY:All Day Event\r\n" +
				"DTSTART;VALUE=DATE:20260115\r\n" +
				"DTEND;VALUE=DATE:20260116\r\n",
			wantLen: 1,
			checkFunc: func(t *testing.T, events []*ical.Event) {
				e := events[0]
				if !e.AllDay {
					t.Error("AllDay should be true")
				}
				if e.Start.Day() != 15 {
					t.Errorf("Start day = %d, want 15", e.Start.Day())
				}
				if e.End.Day() != 16 {
					t.Errorf("End day = %d, want 16", e.End.Day())
				}
			},
		},
		{
			name: "all-day event missing DTEND",
			vevent: "UID:allday-no-end\r\n" +
				"SUMMARY:All Day No End\r\n" +
				"DTSTART;VALUE=DATE:20260115\r\n",
			wantLen: 1,
			checkFunc: func(t *testing.T, events []*ical.Event) {
				e := events[0]
				wantEnd := e.Start.AddDate(0, 0, 1)
				if !e.End.Equal(wantEnd) {
					t.Errorf("End = %v, want %v (Start + 1 day)", e.End, wantEnd)
				}
			},
		},
		{
			name: "timed event missing DTEND",
			vevent: "UID:timed-no-end\r\n" +
				"SUMMARY:Timed No End\r\n" +
				"DTSTART:20260115T100000Z\r\n",
			wantLen: 1,
			checkFunc: func(t *testing.T, events []*ical.Event) {
				e := events[0]
				wantEnd := e.Start.Add(time.Hour)
				if !e.End.Equal(wantEnd) {
					t.Errorf("End = %v, want %v (Start + 1 hour)", e.End, wantEnd)
				}
			},
		},
		{
			name: "event with RRULE",
			vevent: "UID:recurring-uid\r\n" +
				"SUMMARY:Recurring Event\r\n" +
				"DTSTART:20260115T100000Z\r\n" +
				"DTEND:20260115T110000Z\r\n" +
				"RRULE:FREQ=DAILY\r\n",
			wantLen: 1,
			checkFunc: func(t *testing.T, events []*ical.Event) {
				e := events[0]
				if !e.Recurring {
					t.Error("Recurring should be true")
				}
				if e.RecurrenceRule != "FREQ=DAILY" {
					t.Errorf("RecurrenceRule = %q, want %q", e.RecurrenceRule, "FREQ=DAILY")
				}
			},
		},
		{
			name: "event with no UID",
			vevent: "SUMMARY:No UID Event\r\n" +
				"DTSTART:20260115T100000Z\r\n" +
				"DTEND:20260115T110000Z\r\n",
			wantLen: 0,
		},
		{
			name: "event with zero DTSTART",
			vevent: "UID:no-start-uid\r\n" +
				"SUMMARY:No Start Event\r\n" +
				"DTEND:20260115T110000Z\r\n",
			wantLen: 0,
		},
		{
			name:    "calendar with no VEVENT",
			vevent:  "",
			wantLen: 0,
		},
		{
			name: "escaped characters in SUMMARY",
			vevent: "UID:escaped-uid\r\n" +
				"SUMMARY:Meeting\\, Team\r\n" +
				"DTSTART:20260115T100000Z\r\n" +
				"DTEND:20260115T110000Z\r\n",
			wantLen: 1,
			checkFunc: func(t *testing.T, events []*ical.Event) {
				e := events[0]
				if e.Summary == "" {
					t.Error("Summary should not be empty")
				}
			},
		},
		{
			name: "calendarIndex passed through",
			vevent: "UID:index-test\r\n" +
				"SUMMARY:Index Test\r\n" +
				"DTSTART:20260115T100000Z\r\n" +
				"DTEND:20260115T110000Z\r\n",
			calendarIndex: 5,
			wantLen:       1,
			checkFunc: func(t *testing.T, events []*ical.Event) {
				if events[0].CalendarID != 5 {
					t.Errorf("CalendarID = %d, want 5", events[0].CalendarID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cal := makeCalendar(tt.vevent)
			events, err := parseCalendarObject(cal, tt.calendarIndex)
			if err != nil {
				t.Errorf("parseCalendarObject() error = %v", err)
				return
			}
			if len(events) != tt.wantLen {
				t.Errorf("parseCalendarObject() returned %d events, want %d", len(events), tt.wantLen)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, events)
			}
		})
	}
}

func TestEventToIcalRoundTrip(t *testing.T) {
	t.Run("timed event", func(t *testing.T) {
		original := &ical.Event{
			UID:         "round-trip-uid",
			Summary:     "Round Trip Test",
			Description: "Some description",
			Location:    "Room 42",
			Status:      "CONFIRMED",
			Start:       time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			End:         time.Date(2026, 1, 15, 11, 30, 0, 0, time.UTC),
			AllDay:      false,
		}

		cal := eventToIcal(original)
		events, err := parseCalendarObject(cal, 0)
		if err != nil {
			t.Fatalf("parseCalendarObject() error = %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		e := events[0]
		if e.UID != original.UID {
			t.Errorf("UID = %q, want %q", e.UID, original.UID)
		}
		if e.Summary != original.Summary {
			t.Errorf("Summary = %q, want %q", e.Summary, original.Summary)
		}
		if e.Description != original.Description {
			t.Errorf("Description = %q, want %q", e.Description, original.Description)
		}
		if e.Location != original.Location {
			t.Errorf("Location = %q, want %q", e.Location, original.Location)
		}
		if e.Status != original.Status {
			t.Errorf("Status = %q, want %q", e.Status, original.Status)
		}
		if !e.Start.Equal(original.Start) {
			t.Errorf("Start = %v, want %v", e.Start, original.Start)
		}
		if !e.End.Equal(original.End) {
			t.Errorf("End = %v, want %v", e.End, original.End)
		}
		if e.AllDay {
			t.Error("AllDay should be false")
		}
	})

	t.Run("all-day event", func(t *testing.T) {
		original := &ical.Event{
			UID:     "allday-round-trip",
			Summary: "All Day Round Trip",
			Start:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			End:     time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC),
			AllDay:  true,
		}

		cal := eventToIcal(original)
		events, err := parseCalendarObject(cal, 0)
		if err != nil {
			t.Fatalf("parseCalendarObject() error = %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		e := events[0]
		if !e.AllDay {
			t.Error("AllDay should be true")
		}
		if e.Start.Year() != 2026 || e.Start.Month() != time.January || e.Start.Day() != 15 {
			t.Errorf("Start date = %v, want 2026-01-15", e.Start)
		}
		if e.End.Year() != 2026 || e.End.Month() != time.January || e.End.Day() != 16 {
			t.Errorf("End date = %v, want 2026-01-16", e.End)
		}
	})

	t.Run("minimal event (no optional fields)", func(t *testing.T) {
		original := &ical.Event{
			UID:     "minimal-uid",
			Summary: "Minimal Event",
			Start:   time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			End:     time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
		}

		cal := eventToIcal(original)
		events, err := parseCalendarObject(cal, 0)
		if err != nil {
			t.Fatalf("parseCalendarObject() error = %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		e := events[0]
		if e.UID != "minimal-uid" {
			t.Errorf("UID = %q, want %q", e.UID, "minimal-uid")
		}
		if e.Summary != "Minimal Event" {
			t.Errorf("Summary = %q, want %q", e.Summary, "Minimal Event")
		}
	})

	t.Run("recurring event preserves RRULE", func(t *testing.T) {
		original := &ical.Event{
			UID:            "recurring-round-trip",
			Summary:        "Recurring Round Trip",
			Start:          time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			End:            time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
			Recurring:      true,
			RecurrenceRule: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE",
		}

		cal := eventToIcal(original)
		prop := cal.Children[0].Props.Get(goical.PropRecurrenceRule)
		if prop == nil {
			t.Fatal("RRULE was not emitted")
		}
		if prop.Value != original.RecurrenceRule {
			t.Errorf("RRULE = %q, want %q", prop.Value, original.RecurrenceRule)
		}

		events, err := parseCalendarObject(cal, 0)
		if err != nil {
			t.Fatalf("parseCalendarObject() error = %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		e := events[0]
		if !e.Recurring {
			t.Error("Recurring should be true")
		}
		if e.RecurrenceRule != original.RecurrenceRule {
			t.Errorf("RecurrenceRule = %q, want %q", e.RecurrenceRule, original.RecurrenceRule)
		}
	})

	t.Run("event with no Description/Location", func(t *testing.T) {
		original := &ical.Event{
			UID:     "no-optional-uid",
			Summary: "No Optional Fields",
			Start:   time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			End:     time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
		}

		cal := eventToIcal(original)
		events, err := parseCalendarObject(cal, 0)
		if err != nil {
			t.Fatalf("parseCalendarObject() error = %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		e := events[0]
		if e.Description != "" {
			t.Errorf("Description = %q, want empty", e.Description)
		}
		if e.Location != "" {
			t.Errorf("Location = %q, want empty", e.Location)
		}
	})
}
