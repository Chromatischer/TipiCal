package ical

import (
	"testing"
	"time"
)

func TestOverlapsWith(t *testing.T) {
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := func(hour int) time.Time { return day.Add(time.Duration(hour) * time.Hour) }
	hm := func(hour, min int) time.Time {
		return day.Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute)
	}

	tests := []struct {
		name     string
		eStart   time.Time
		eEnd     time.Time
		qStart   time.Time
		qEnd     time.Time
		overlaps bool
	}{
		{"fully before", h(8), h(9), h(10), h(11), false},
		{"fully after", h(12), h(13), h(10), h(11), false},
		{"partial overlap at start", hm(9, 30), h(11), h(10), h(12), true},
		{"partial overlap at end", h(11), hm(12, 30), h(10), h(12), true},
		{"event contains range", h(9), h(13), h(10), h(12), true},
		{"range contains event", hm(10, 30), hm(11, 30), h(10), h(12), true},
		{"exact match", h(10), h(12), h(10), h(12), true},
		{"adjacent before (touching)", h(8), h(10), h(10), h(12), false},
		{"adjacent after (touching)", h(12), h(14), h(10), h(12), false},
		{"zero-duration event inside range", h(11), h(11), h(10), h(12), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Event{Start: tt.eStart, End: tt.eEnd}
			got := e.OverlapsWith(tt.qStart, tt.qEnd)
			if got != tt.overlaps {
				t.Errorf("OverlapsWith(%v, %v) = %v, want %v", tt.qStart, tt.qEnd, got, tt.overlaps)
			}
		})
	}
}

func TestIsMultiDay(t *testing.T) {
	tests := []struct {
		name     string
		allDay   bool
		start    time.Time
		end      time.Time
		multiDay bool
	}{
		{
			name:     "single-day timed (same day)",
			allDay:   false,
			start:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 1, 17, 0, 0, 0, time.UTC),
			multiDay: false,
		},
		{
			name:     "multi-day timed (different days)",
			allDay:   false,
			start:    time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 2, 2, 0, 0, 0, time.UTC),
			multiDay: true,
		},
		{
			name:     "single all-day (End = Start+1day)",
			allDay:   true,
			start:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			multiDay: false,
		},
		{
			name:     "multi-day all-day",
			allDay:   true,
			start:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
			multiDay: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Event{AllDay: tt.allDay, Start: tt.start, End: tt.end}
			got := e.IsMultiDay()
			if got != tt.multiDay {
				t.Errorf("IsMultiDay() = %v, want %v", got, tt.multiDay)
			}
		})
	}
}

func TestAllDaySpanDays(t *testing.T) {
	tests := []struct {
		name   string
		allDay bool
		start  time.Time
		end    time.Time
		want   int
	}{
		{
			name:   "non-all-day event",
			allDay: false,
			start:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:    time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			want:   0,
		},
		{
			name:   "single all-day",
			allDay: true,
			start:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:    time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			want:   1,
		},
		{
			name:   "multi-day all-day (3 days)",
			allDay: true,
			start:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:    time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC),
			want:   3,
		},
		{
			name:   "same start and end (degenerate)",
			allDay: true,
			start:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			want:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Event{AllDay: tt.allDay, Start: tt.start, End: tt.end}
			got := e.AllDaySpanDays()
			if got != tt.want {
				t.Errorf("AllDaySpanDays() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRecurrenceDescription(t *testing.T) {
	tests := []struct {
		name  string
		event *Event
		want  string
	}{
		{"empty rule", &Event{Recurring: true, RecurrenceRule: ""}, ""},
		{"not recurring", &Event{Recurring: false, RecurrenceRule: "FREQ=DAILY"}, ""},
		{"FREQ=DAILY", &Event{Recurring: true, RecurrenceRule: "FREQ=DAILY"}, "Daily"},
		{"FREQ=DAILY;INTERVAL=3", &Event{Recurring: true, RecurrenceRule: "FREQ=DAILY;INTERVAL=3"}, "Every 3 days"},
		{"FREQ=WEEKLY", &Event{Recurring: true, RecurrenceRule: "FREQ=WEEKLY"}, "Weekly"},
		{"FREQ=WEEKLY;INTERVAL=2", &Event{Recurring: true, RecurrenceRule: "FREQ=WEEKLY;INTERVAL=2"}, "Every 2 weeks"},
		{"FREQ=WEEKLY;BYDAY=MO,WE,FR", &Event{Recurring: true, RecurrenceRule: "FREQ=WEEKLY;BYDAY=MO,WE,FR"}, "Weekly on Mon, Wed and Fri"},
		{"FREQ=WEEKLY;BYDAY=MO", &Event{Recurring: true, RecurrenceRule: "FREQ=WEEKLY;BYDAY=MO"}, "Weekly on Mon"},
		{"FREQ=MONTHLY", &Event{Recurring: true, RecurrenceRule: "FREQ=MONTHLY"}, "Monthly"},
		{"FREQ=YEARLY", &Event{Recurring: true, RecurrenceRule: "FREQ=YEARLY"}, "Yearly"},
		{"FREQ=DAILY;COUNT=5", &Event{Recurring: true, RecurrenceRule: "FREQ=DAILY;COUNT=5"}, "Daily (5 times)"},
		{"FREQ=WEEKLY;UNTIL=20261231T000000Z", &Event{Recurring: true, RecurrenceRule: "FREQ=WEEKLY;UNTIL=20261231T000000Z"}, "Weekly (until 20261231T000000Z)"},
		{"FREQ=DAILY;INTERVAL=2;COUNT=10", &Event{Recurring: true, RecurrenceRule: "FREQ=DAILY;INTERVAL=2;COUNT=10"}, "Every 2 days (10 times)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.RecurrenceDescription()
			if got != tt.want {
				t.Errorf("RecurrenceDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}
