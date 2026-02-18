package util

import (
	"testing"
	"time"
)

func TestWeekStart(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name        string
		input       time.Time
		startMonday bool
		want        time.Time
	}{
		{
			name:        "Monday-start, Wednesday returns same-week Monday",
			input:       time.Date(2026, time.February, 18, 0, 0, 0, 0, loc), // Wed
			startMonday: true,
			want:        time.Date(2026, time.February, 16, 0, 0, 0, 0, loc), // Mon
		},
		{
			name:        "Monday-start, Sunday returns previous Monday",
			input:       time.Date(2026, time.February, 22, 0, 0, 0, 0, loc), // Sun
			startMonday: true,
			want:        time.Date(2026, time.February, 16, 0, 0, 0, 0, loc), // Mon
		},
		{
			name:        "Monday-start, Saturday returns same-week Monday",
			input:       time.Date(2026, time.February, 21, 0, 0, 0, 0, loc), // Sat
			startMonday: true,
			want:        time.Date(2026, time.February, 16, 0, 0, 0, 0, loc), // Mon
		},
		{
			name:        "Monday-start, Monday returns same day",
			input:       time.Date(2026, time.February, 16, 0, 0, 0, 0, loc), // Mon
			startMonday: true,
			want:        time.Date(2026, time.February, 16, 0, 0, 0, 0, loc),
		},
		{
			name:        "Sunday-start, Wednesday returns same-week Sunday",
			input:       time.Date(2026, time.February, 18, 0, 0, 0, 0, loc), // Wed
			startMonday: false,
			want:        time.Date(2026, time.February, 15, 0, 0, 0, 0, loc), // Sun
		},
		{
			name:        "Sunday-start, Saturday returns same-week Sunday",
			input:       time.Date(2026, time.February, 21, 0, 0, 0, 0, loc), // Sat
			startMonday: false,
			want:        time.Date(2026, time.February, 15, 0, 0, 0, 0, loc), // Sun
		},
		{
			name:        "Sunday-start, Sunday returns same day",
			input:       time.Date(2026, time.February, 15, 0, 0, 0, 0, loc), // Sun
			startMonday: false,
			want:        time.Date(2026, time.February, 15, 0, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WeekStart(tt.input, tt.startMonday)
			if !got.Equal(tt.want) {
				t.Errorf("WeekStart(%v, %v) = %v, want %v",
					tt.input.Format("2006-01-02 (Mon)"), tt.startMonday,
					got.Format("2006-01-02 (Mon)"), tt.want.Format("2006-01-02 (Mon)"))
			}
		})
	}
}

func TestMonthGrid(t *testing.T) {
	// February 2026: starts on Sunday
	t.Run("Feb 2026 Monday-start: grid begins on previous Monday", func(t *testing.T) {
		grid := MonthGrid(2026, time.February, true)
		// Feb 1 is Sunday → week starts Jan 26 (Mon)
		wantFirst := time.Date(2026, time.January, 26, 0, 0, 0, 0, time.Local)
		if !SameDay(grid[0][0], wantFirst) {
			t.Errorf("grid[0][0] = %v, want %v", grid[0][0].Format("2006-01-02 (Mon)"), wantFirst.Format("2006-01-02 (Mon)"))
		}
		// grid[0][6] should be Feb 1 (Sunday)
		wantSun := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.Local)
		if !SameDay(grid[0][6], wantSun) {
			t.Errorf("grid[0][6] = %v, want %v", grid[0][6].Format("2006-01-02"), wantSun.Format("2006-01-02"))
		}
	})

	t.Run("Feb 2026 Sunday-start: grid begins on Feb 1", func(t *testing.T) {
		grid := MonthGrid(2026, time.February, false)
		// Feb 1 is Sunday → grid starts on Feb 1 itself
		wantFirst := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.Local)
		if !SameDay(grid[0][0], wantFirst) {
			t.Errorf("grid[0][0] = %v, want %v", grid[0][0].Format("2006-01-02"), wantFirst.Format("2006-01-02"))
		}
		// grid[0][6] should be Feb 7
		wantLast := time.Date(2026, time.February, 7, 0, 0, 0, 0, time.Local)
		if !SameDay(grid[0][6], wantLast) {
			t.Errorf("grid[0][6] = %v, want %v", grid[0][6].Format("2006-01-02"), wantLast.Format("2006-01-02"))
		}
	})

	t.Run("Feb 2024 leap year: Feb 29 appears in grid", func(t *testing.T) {
		// Feb 1, 2024 is Thursday → gridStart = Jan 29 (Mon)
		grid := MonthGrid(2024, time.February, true)
		wantFirst := time.Date(2024, time.January, 29, 0, 0, 0, 0, time.Local)
		if !SameDay(grid[0][0], wantFirst) {
			t.Errorf("grid[0][0] = %v, want %v", grid[0][0].Format("2006-01-02"), wantFirst.Format("2006-01-02"))
		}
		// Feb 29 is 31 days after Jan 29 → row 4, col 3
		feb29 := time.Date(2024, time.February, 29, 0, 0, 0, 0, time.Local)
		if !SameDay(grid[4][3], feb29) {
			t.Errorf("grid[4][3] = %v, want Feb 29", grid[4][3].Format("2006-01-02"))
		}
	})

	t.Run("July 2023 Monday-start: Jul 31 appears in row 5", func(t *testing.T) {
		// Jul 1, 2023 is Saturday → gridStart = Jun 26 (Mon)
		grid := MonthGrid(2023, time.July, true)
		wantFirst := time.Date(2023, time.June, 26, 0, 0, 0, 0, time.Local)
		if !SameDay(grid[0][0], wantFirst) {
			t.Errorf("grid[0][0] = %v, want %v", grid[0][0].Format("2006-01-02"), wantFirst.Format("2006-01-02"))
		}
		// Jul 31 = Jun 26 + 35 days → grid[5][0]
		jul31 := time.Date(2023, time.July, 31, 0, 0, 0, 0, time.Local)
		if !SameDay(grid[5][0], jul31) {
			t.Errorf("grid[5][0] = %v, want Jul 31", grid[5][0].Format("2006-01-02"))
		}
	})

	t.Run("grid is always 6 rows of 7 days = 42 consecutive days", func(t *testing.T) {
		grid := MonthGrid(2026, time.February, true)
		for i := 0; i < 41; i++ {
			row, col := i/7, i%7
			nRow, nCol := (i+1)/7, (i+1)%7
			cur := grid[row][col]
			next := grid[nRow][nCol]
			diff := next.Sub(cur)
			if diff != 24*time.Hour {
				t.Errorf("cell %d→%d: gap = %v, want 24h", i, i+1, diff)
			}
		}
	})
}

func TestSameDay(t *testing.T) {
	t.Run("same day different times returns true", func(t *testing.T) {
		a := time.Date(2026, time.February, 18, 9, 0, 0, 0, time.UTC)
		b := time.Date(2026, time.February, 18, 23, 59, 59, 0, time.UTC)
		if !SameDay(a, b) {
			t.Error("expected SameDay to be true for same date different times")
		}
	})

	t.Run("different days returns false", func(t *testing.T) {
		a := time.Date(2026, time.February, 18, 12, 0, 0, 0, time.UTC)
		b := time.Date(2026, time.February, 19, 12, 0, 0, 0, time.UTC)
		if SameDay(a, b) {
			t.Error("expected SameDay to be false for different dates")
		}
	})

	t.Run("same calendar date different years returns false", func(t *testing.T) {
		a := time.Date(2025, time.February, 18, 12, 0, 0, 0, time.UTC)
		b := time.Date(2026, time.February, 18, 12, 0, 0, 0, time.UTC)
		if SameDay(a, b) {
			t.Error("expected SameDay to be false for same date in different years")
		}
	})

	t.Run("different timezones same absolute moment may differ by calendar day", func(t *testing.T) {
		// UTC midnight Feb 18 = Feb 17 11pm in UTC-1
		utc := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
		west := utc.In(time.FixedZone("UTC-1", -3600)) // still Feb 17 in UTC-1
		if SameDay(utc, west) {
			t.Error("expected SameDay to differ for times in different timezones that span a date boundary")
		}
	})
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name   string
		t      time.Time
		use24h bool
		want   string
	}{
		{
			name:   "24h afternoon",
			t:      time.Date(2026, 1, 1, 15, 30, 0, 0, time.UTC),
			use24h: true,
			want:   "15:30",
		},
		{
			name:   "12h afternoon",
			t:      time.Date(2026, 1, 1, 15, 30, 0, 0, time.UTC),
			use24h: false,
			want:   "3:30 PM",
		},
		{
			name:   "24h midnight",
			t:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			use24h: true,
			want:   "00:00",
		},
		{
			name:   "12h midnight",
			t:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			use24h: false,
			want:   "12:00 AM",
		},
		{
			name:   "24h noon",
			t:      time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			use24h: true,
			want:   "12:00",
		},
		{
			name:   "12h noon",
			t:      time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			use24h: false,
			want:   "12:00 PM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTime(tt.t, tt.use24h)
			if got != tt.want {
				t.Errorf("FormatTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTimeRange(t *testing.T) {
	start := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 17, 30, 0, 0, time.UTC)

	t.Run("24h range", func(t *testing.T) {
		got := FormatTimeRange(start, end, true)
		want := "09:00 - 17:30"
		if got != want {
			t.Errorf("FormatTimeRange() = %q, want %q", got, want)
		}
	})

	t.Run("12h range", func(t *testing.T) {
		got := FormatTimeRange(start, end, false)
		want := "9:00 AM - 5:30 PM"
		if got != want {
			t.Errorf("FormatTimeRange() = %q, want %q", got, want)
		}
	})
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "shorter than max returns unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length returns unchanged",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "maxLen=0 returns empty string",
			input:  "hello",
			maxLen: 0,
			want:   "",
		},
		{
			name:   "maxLen=1 returns ellipsis",
			input:  "hello world",
			maxLen: 1,
			want:   "…",
		},
		{
			name:   "ASCII truncated with ellipsis",
			input:  "Hello",
			maxLen: 4,
			want:   "Hel…", // 3 chars + ellipsis = 4 cols
		},
		{
			name:   "emoji: 2-col wide characters counted correctly",
			input:  "🎉🎉🎉",
			maxLen: 5,
			want:   "🎉🎉…", // 2+2+1 = 5 cols
		},
		{
			name:   "CJK: 2-col wide characters",
			input:  "日本語",
			maxLen: 3,
			want:   "日…", // 2+1 = 3 cols
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateText(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateText(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestHoursBetween(t *testing.T) {
	day := func(h, m int) time.Time {
		return time.Date(2026, 1, 1, h, m, 0, 0, time.UTC)
	}

	event := func(startH, startM, endH, endM int) struct{ Start, End time.Time } {
		return struct{ Start, End time.Time }{day(startH, startM), day(endH, endM)}
	}

	t.Run("no events returns defaults", func(t *testing.T) {
		minH, maxH := HoursBetween(nil, 8, 20)
		if minH != 8 || maxH != 20 {
			t.Errorf("HoursBetween(nil, 8, 20) = (%d, %d), want (8, 20)", minH, maxH)
		}
	})

	t.Run("event within defaults: no change", func(t *testing.T) {
		events := []struct{ Start, End time.Time }{event(9, 0, 17, 0)}
		minH, maxH := HoursBetween(events, 8, 20)
		if minH != 8 || maxH != 20 {
			t.Errorf("HoursBetween() = (%d, %d), want (8, 20)", minH, maxH)
		}
	})

	t.Run("early start expands minH", func(t *testing.T) {
		events := []struct{ Start, End time.Time }{event(6, 0, 10, 0)}
		minH, maxH := HoursBetween(events, 8, 20)
		if minH != 6 || maxH != 20 {
			t.Errorf("HoursBetween() = (%d, %d), want (6, 20)", minH, maxH)
		}
	})

	t.Run("late end expands maxH", func(t *testing.T) {
		events := []struct{ Start, End time.Time }{event(9, 0, 22, 0)}
		minH, maxH := HoursBetween(events, 8, 20)
		if minH != 8 || maxH != 22 {
			t.Errorf("HoursBetween() = (%d, %d), want (8, 22)", minH, maxH)
		}
	})

	t.Run("non-zero end minutes rounds up", func(t *testing.T) {
		// event ends at 17:30 → end hour should be 18
		events := []struct{ Start, End time.Time }{event(9, 0, 17, 30)}
		minH, maxH := HoursBetween(events, 8, 17)
		if minH != 8 || maxH != 18 {
			t.Errorf("HoursBetween() = (%d, %d), want (8, 18)", minH, maxH)
		}
	})

	t.Run("exact hour end does not round up", func(t *testing.T) {
		events := []struct{ Start, End time.Time }{event(9, 0, 17, 0)}
		minH, maxH := HoursBetween(events, 8, 17)
		if minH != 8 || maxH != 17 {
			t.Errorf("HoursBetween() = (%d, %d), want (8, 17)", minH, maxH)
		}
	})
}
