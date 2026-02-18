package ical

import (
	"testing"
	"time"
)

func TestExpandRecurrence(t *testing.T) {
	base := &Event{
		UID:     "test-uid",
		Summary: "Daily Standup",
		Start:   time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
		End:     time.Date(2026, 1, 1, 9, 30, 0, 0, time.UTC),
	}

	janEnd := func(day int) time.Time {
		return time.Date(2026, 1, day, 23, 59, 59, 0, time.UTC)
	}

	janStart := func(day int) time.Time {
		return time.Date(2026, 1, day, 0, 0, 0, 0, time.UTC)
	}

	tests := []struct {
		name       string
		event      *Event
		rruleStr   string
		rangeStart time.Time
		rangeEnd   time.Time
		wantLen    int
		wantDates  []time.Time
		checkFunc  func(t *testing.T, instances []*Event)
	}{
		{
			name:       "empty rule returns original",
			event:      base,
			rruleStr:   "",
			rangeStart: janStart(1),
			rangeEnd:   janEnd(10),
			wantLen:    1,
			checkFunc: func(t *testing.T, instances []*Event) {
				if instances[0] != base {
					t.Error("empty rule should return original event pointer")
				}
			},
		},
		{
			name:       "invalid rule returns original",
			event:      base,
			rruleStr:   "NOT_A_RULE",
			rangeStart: janStart(1),
			rangeEnd:   janEnd(10),
			wantLen:    1,
			checkFunc: func(t *testing.T, instances []*Event) {
				if instances[0] != base {
					t.Error("invalid rule should return original event pointer")
				}
			},
		},
		{
			name:       "daily, 3 days",
			event:      base,
			rruleStr:   "FREQ=DAILY",
			rangeStart: janStart(1),
			rangeEnd:   janEnd(3),
			wantLen:    3,
			wantDates:  []time.Time{janStart(1), janStart(2), janStart(3)},
		},
		{
			name:       "weekly BYDAY MO,WE for 2 weeks",
			event:      base,
			rruleStr:   "FREQ=WEEKLY;BYDAY=MO,WE",
			rangeStart: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2026, 1, 18, 23, 59, 59, 0, time.UTC),
			wantLen:    4,
			wantDates: []time.Time{
				time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "COUNT limits instances",
			event:      base,
			rruleStr:   "FREQ=DAILY;COUNT=3",
			rangeStart: janStart(1),
			rangeEnd:   janEnd(10),
			wantLen:    3,
		},
		{
			name:       "UNTIL limits instances",
			event:      base,
			rruleStr:   "FREQ=DAILY;UNTIL=20260103T235959Z",
			rangeStart: janStart(1),
			rangeEnd:   janEnd(10),
			wantLen:    3,
		},
		{
			name:       "range start after event start",
			event:      base,
			rruleStr:   "FREQ=DAILY",
			rangeStart: janStart(5),
			rangeEnd:   janEnd(7),
			wantLen:    3,
			wantDates:  []time.Time{janStart(5), janStart(6), janStart(7)},
		},
		{
			name:       "duration preserved",
			event:      base,
			rruleStr:   "FREQ=DAILY",
			rangeStart: janStart(1),
			rangeEnd:   janEnd(3),
			checkFunc: func(t *testing.T, instances []*Event) {
				for i, inst := range instances {
					dur := inst.End.Sub(inst.Start)
					if dur != 30*time.Minute {
						t.Errorf("instance %d: duration = %v, want 30m", i, dur)
					}
				}
			},
		},
		{
			name:       "metadata copied",
			event:      base,
			rruleStr:   "FREQ=DAILY",
			rangeStart: janStart(1),
			rangeEnd:   janEnd(1),
			wantLen:    1,
			checkFunc: func(t *testing.T, instances []*Event) {
				inst := instances[0]
				if inst.UID != base.UID {
					t.Errorf("instance.UID = %q, want %q", inst.UID, base.UID)
				}
				if inst.Summary != base.Summary {
					t.Errorf("instance.Summary = %q, want %q", inst.Summary, base.Summary)
				}
			},
		},
		{
			name:       "empty range (start==end)",
			event:      base,
			rruleStr:   "FREQ=DAILY",
			rangeStart: janStart(5),
			rangeEnd:   janStart(4),
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandRecurrence(tt.event, tt.rruleStr, tt.rangeStart, tt.rangeEnd)

			if tt.wantLen > 0 && len(got) != tt.wantLen {
				t.Errorf("ExpandRecurrence() returned %d instances, want %d", len(got), tt.wantLen)
			}

			if tt.wantDates != nil {
				for i, wantDate := range tt.wantDates {
					if i >= len(got) {
						t.Errorf("missing instance %d for date %v", i, wantDate)
						continue
					}
					gotDate := got[i].Start
					if !sameDay(gotDate, wantDate) {
						t.Errorf("instance %d: date = %v, want %v", i, gotDate.Format("2006-01-02"), wantDate.Format("2006-01-02"))
					}
				}
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

func sameDay(a, b time.Time) bool {
	aYear, aMonth, aDay := a.Date()
	bYear, bMonth, bDay := b.Date()
	return aYear == bYear && aMonth == bMonth && aDay == bDay
}
