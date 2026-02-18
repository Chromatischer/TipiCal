package ical

import (
	"testing"
	"time"
)

func timedEventSimple(uid string, startH, endH int) *Event {
	d := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	return &Event{
		UID:   uid,
		Start: d.Add(time.Duration(startH) * time.Hour),
		End:   d.Add(time.Duration(endH) * time.Hour),
	}
}

func allDayEvent(uid string, startDay, endDay int) *Event {
	d := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &Event{
		UID:    uid,
		AllDay: true,
		Start:  d.AddDate(0, 0, startDay),
		End:    d.AddDate(0, 0, endDay),
	}
}

func timedEventOnDay(uid string, day int, startH, endH int) *Event {
	d := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, day)
	return &Event{
		UID:   uid,
		Start: d.Add(time.Duration(startH) * time.Hour),
		End:   d.Add(time.Duration(endH) * time.Hour),
	}
}

func timedEventEndsAtMidnight(uid string, day int, hour int) *Event {
	d := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, day)
	return &Event{
		UID:   uid,
		Start: d.Add(time.Duration(hour) * time.Hour),
		End:   d,
	}
}

func timedEventStartsAtMidnight(uid string) *Event {
	d := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	return &Event{
		UID:   uid,
		Start: d,
		End:   d.Add(time.Hour),
	}
}

func timedEventWithMinutes(uid string, startH, startM, endH, endM int) *Event {
	d := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	return &Event{
		UID:   uid,
		Start: d.Add(time.Duration(startH)*time.Hour + time.Duration(startM)*time.Minute),
		End:   d.Add(time.Duration(endH)*time.Hour + time.Duration(endM)*time.Minute),
	}
}

func TestEventsForDay(t *testing.T) {
	day := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		events    []*Event
		wantCount int
		wantUIDs  []string
	}{
		{
			name:      "timed event on that day",
			events:    []*Event{timedEventSimple("e1", 9, 17)},
			wantCount: 1,
			wantUIDs:  []string{"e1"},
		},
		{
			name:      "timed event on different day",
			events:    []*Event{timedEventOnDay("e1", 16, 9, 17)},
			wantCount: 0,
			wantUIDs:  nil,
		},
		{
			name:      "all-day event covering that day",
			events:    []*Event{allDayEvent("e1", 14, 15)},
			wantCount: 1,
			wantUIDs:  []string{"e1"},
		},
		{
			name:      "multi-day all-day spanning that day",
			events:    []*Event{allDayEvent("e1", 10, 20)},
			wantCount: 1,
			wantUIDs:  []string{"e1"},
		},
		{
			name:      "all-day event ending on that day",
			events:    []*Event{allDayEvent("e1", 13, 14)},
			wantCount: 0,
			wantUIDs:  nil,
		},
		{
			name:      "midnight edge: event ending at midnight of day",
			events:    []*Event{timedEventEndsAtMidnight("e1", 15, 0)},
			wantCount: 0,
			wantUIDs:  nil,
		},
		{
			name:      "midnight edge: event starting at midnight of day",
			events:    []*Event{timedEventStartsAtMidnight("e1")},
			wantCount: 1,
			wantUIDs:  []string{"e1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{Events: tt.events}
			got := s.EventsForDay(day)
			if len(got) != tt.wantCount {
				t.Errorf("EventsForDay() returned %d events, want %d", len(got), tt.wantCount)
				return
			}
			if tt.wantUIDs != nil {
				gotUIDs := make(map[string]bool)
				for _, e := range got {
					gotUIDs[e.UID] = true
				}
				for _, uid := range tt.wantUIDs {
					if !gotUIDs[uid] {
						t.Errorf("EventsForDay() missing event with UID %q", uid)
					}
				}
			}
		})
	}
}

func TestEventsInRange(t *testing.T) {
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	rangeStart := day.Add(10 * time.Hour)
	rangeEnd := day.Add(12 * time.Hour)

	tests := []struct {
		name      string
		events    []*Event
		wantCount int
		wantUIDs  []string
	}{
		{
			name:      "event fully inside range",
			events:    []*Event{timedEventWithMinutes("e1", 10, 30, 11, 30)},
			wantCount: 1,
			wantUIDs:  []string{"e1"},
		},
		{
			name:      "event fully outside range (before)",
			events:    []*Event{timedEventSimple("e1", 8, 9)},
			wantCount: 0,
			wantUIDs:  nil,
		},
		{
			name:      "event fully outside range (after)",
			events:    []*Event{timedEventSimple("e1", 13, 14)},
			wantCount: 0,
			wantUIDs:  nil,
		},
		{
			name:      "event partially overlapping start of range",
			events:    []*Event{timedEventSimple("e1", 9, 11)},
			wantCount: 1,
			wantUIDs:  []string{"e1"},
		},
		{
			name:      "event partially overlapping end of range",
			events:    []*Event{timedEventSimple("e1", 11, 13)},
			wantCount: 1,
			wantUIDs:  []string{"e1"},
		},
		{
			name:      "empty store",
			events:    []*Event{},
			wantCount: 0,
			wantUIDs:  nil,
		},
		{
			name: "multiple events, only some overlap",
			events: []*Event{
				timedEventSimple("e1", 8, 9),
				timedEventWithMinutes("e2", 10, 30, 11, 30),
				timedEventSimple("e3", 13, 14),
				timedEventSimple("e4", 11, 12),
			},
			wantCount: 2,
			wantUIDs:  []string{"e2", "e4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{Events: tt.events}
			got := s.EventsInRange(rangeStart, rangeEnd)
			if len(got) != tt.wantCount {
				t.Errorf("EventsInRange() returned %d events, want %d", len(got), tt.wantCount)
				return
			}
			if tt.wantUIDs != nil {
				gotUIDs := make(map[string]bool)
				for _, e := range got {
					gotUIDs[e.UID] = true
				}
				for _, uid := range tt.wantUIDs {
					if !gotUIDs[uid] {
						t.Errorf("EventsInRange() missing event with UID %q", uid)
					}
				}
			}
		})
	}
}

func TestAddEvent(t *testing.T) {
	s := NewStore()
	e := &Event{UID: "test-uid", Summary: "Test Event"}

	s.AddEvent(e)

	if len(s.Events) != 1 {
		t.Errorf("AddEvent() expected 1 event, got %d", len(s.Events))
	}
	if s.Events[0].UID != "test-uid" {
		t.Errorf("AddEvent() event UID = %q, want %q", s.Events[0].UID, "test-uid")
	}
}

func TestFindEvent(t *testing.T) {
	t.Run("AddEvent then FindEvent returns same pointer", func(t *testing.T) {
		s := NewStore()
		e := &Event{UID: "test-uid", Summary: "Test Event"}
		s.AddEvent(e)

		got := s.FindEvent("test-uid")
		if got != e {
			t.Errorf("FindEvent() = %p, want %p", got, e)
		}
	})

	t.Run("FindEvent on empty store returns nil", func(t *testing.T) {
		s := NewStore()
		got := s.FindEvent("nonexistent")
		if got != nil {
			t.Errorf("FindEvent() = %v, want nil", got)
		}
	})

	t.Run("FindEvent with wrong UID returns nil", func(t *testing.T) {
		s := NewStore()
		s.AddEvent(&Event{UID: "test-uid"})
		got := s.FindEvent("wrong-uid")
		if got != nil {
			t.Errorf("FindEvent() = %v, want nil", got)
		}
	})

	t.Run("FindEvent when multiple events exist returns the right one", func(t *testing.T) {
		s := NewStore()
		e1 := &Event{UID: "uid-1", Summary: "Event 1"}
		e2 := &Event{UID: "uid-2", Summary: "Event 2"}
		e3 := &Event{UID: "uid-3", Summary: "Event 3"}
		s.AddEvent(e1)
		s.AddEvent(e2)
		s.AddEvent(e3)

		got := s.FindEvent("uid-2")
		if got != e2 {
			t.Errorf("FindEvent() = %p, want %p", got, e2)
		}
	})
}

func TestRemoveEvent(t *testing.T) {
	t.Run("RemoveEvent removes correctly, FindEvent after returns nil", func(t *testing.T) {
		s := NewStore()
		s.AddEvent(&Event{UID: "test-uid"})
		s.RemoveEvent("test-uid")

		got := s.FindEvent("test-uid")
		if got != nil {
			t.Errorf("FindEvent() after RemoveEvent = %v, want nil", got)
		}
	})

	t.Run("RemoveEvent on non-existent UID is a no-op", func(t *testing.T) {
		s := NewStore()
		s.AddEvent(&Event{UID: "test-uid"})
		s.RemoveEvent("nonexistent")

		if len(s.Events) != 1 {
			t.Errorf("RemoveEvent() non-existent should not change store length, got %d", len(s.Events))
		}
	})

	t.Run("After RemoveEvent, len(s.Events) decreases by 1", func(t *testing.T) {
		s := NewStore()
		s.AddEvent(&Event{UID: "uid-1"})
		s.AddEvent(&Event{UID: "uid-2"})
		initialLen := len(s.Events)

		s.RemoveEvent("uid-1")

		if len(s.Events) != initialLen-1 {
			t.Errorf("RemoveEvent() expected len %d, got %d", initialLen-1, len(s.Events))
		}
	})
}
