package util

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		input   string
		wantH   int
		wantM   int
		wantErr bool
	}{
		{"09:00", 9, 0, false},
		{"9:00", 9, 0, false},
		{"9:30", 9, 30, false},
		{"9:5", 9, 5, false},
		{"14:00", 14, 0, false},
		{"9", 9, 0, false},
		{"14", 14, 0, false},
		{"0", 0, 0, false},
		{"23", 23, 0, false},
		{"morning", 9, 0, false},
		{"noon", 12, 0, false},
		{"midday", 12, 0, false},
		{"afternoon", 14, 0, false},
		{"evening", 18, 0, false},
		{"night", 20, 0, false},
		{"midnight", 0, 0, false},
		// Case insensitive
		{"Morning", 9, 0, false},
		{"NOON", 12, 0, false},
		// Compound: base+offset / base-offset
		{"09:00+1", 10, 0, false},
		{"09:00+1:30", 10, 30, false},
		{"09:00+1.5", 10, 30, false},
		{"09:00+30m", 9, 30, false},
		{"14:30-0:30", 14, 0, false},
		{"14:30-1", 13, 30, false},
		{"morning+1", 10, 0, false},
		{"9:30+1", 10, 30, false},
		// Bare +N must NOT be treated as a time (would silently give N:00am)
		{"+1", 0, 0, true},
		{"+8", 0, 0, true},
		// Invalid
		{"25:00", 0, 0, true},
		{"abc", 0, 0, true},
		{"9:60", 0, 0, true},
		{"24", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTime(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseTime(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Hour() != tt.wantH || got.Minute() != tt.wantM {
				t.Errorf("ParseTime(%q) = %02d:%02d, want %02d:%02d",
					tt.input, got.Hour(), got.Minute(), tt.wantH, tt.wantM)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	ref := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.Local)

	tests := []struct {
		input     string
		wantYear  int
		wantMonth time.Month
		wantDay   int
		wantErr   bool
	}{
		// Full DD.MM.YYYY
		{"18.02.2026", 2026, time.February, 18, false},
		{"1.1.2026", 2026, time.January, 1, false},
		{"31.12.2025", 2025, time.December, 31, false},
		// DD.MM — assumes ref year (2026)
		{"18.02", 2026, time.February, 18, false},
		{"1.3", 2026, time.March, 1, false},
		// DD — assumes ref month (February) and year (2026)
		{"18", 2026, time.February, 18, false},
		{"1", 2026, time.February, 1, false},
		// Invalid
		{"32.01.2026", 0, 0, 0, true},
		{"18.13.2026", 0, 0, 0, true},
		{"abc", 0, 0, 0, true},
		{"0", 0, 0, 0, true},
		{"32", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDate(tt.input, ref)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDate(%q, ref) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseDate(%q, ref) unexpected error: %v", tt.input, err)
				return
			}
			y, mo, d := got.Date()
			if y != tt.wantYear || mo != tt.wantMonth || d != tt.wantDay {
				t.Errorf("ParseDate(%q, ref) = %04d-%02d-%02d, want %04d-%02d-%02d",
					tt.input, y, mo, d, tt.wantYear, tt.wantMonth, tt.wantDay)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	start := time.Date(2026, time.February, 25, 14, 0, 0, 0, time.Local)

	tests := []struct {
		input    string
		wantH    int
		wantM    int
		wantDay  int // expected day (for +Nd tests)
		wantOK   bool
	}{
		{"+2", 16, 0, 25, true},
		{"+2h", 16, 0, 25, true},
		{"+30m", 14, 30, 25, true},
		{"+1:30", 15, 30, 25, true},
		{"+1.5", 15, 30, 25, true},
		{"+0:45", 14, 45, 25, true},
		// Days — result time is same H:M but +N days
		{"+2d", 14, 0, 27, true},
		{"+1d", 14, 0, 26, true},
		// Negative offsets
		{"-1", 13, 0, 25, true},
		{"-1:30", 12, 30, 25, true},
		{"-30m", 13, 30, 25, true},
		{"-2d", 14, 0, 23, true},
		// Not a duration
		{"15:30", 0, 0, 0, false},
		{"", 0, 0, 0, false},
		{"+", 0, 0, 0, false},
		{"-", 0, 0, 0, false},
		{"abc", 0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseDuration(tt.input, start)
			if ok != tt.wantOK {
				t.Errorf("ParseDuration(%q) ok=%v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if !ok {
				return
			}
			if got.Hour() != tt.wantH || got.Minute() != tt.wantM {
				t.Errorf("ParseDuration(%q) time = %02d:%02d, want %02d:%02d",
					tt.input, got.Hour(), got.Minute(), tt.wantH, tt.wantM)
			}
			if got.Day() != tt.wantDay {
				t.Errorf("ParseDuration(%q) day = %d, want %d", tt.input, got.Day(), tt.wantDay)
			}
		})
	}
}

func TestParseDateRelative(t *testing.T) {
	ref := time.Now()

	t.Run("today", func(t *testing.T) {
		got, err := ParseDate("today", ref)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		now := time.Now()
		if !SameDay(got, now) {
			t.Errorf("ParseDate(\"today\") = %v, want today (%v)", got.Format("2006-01-02"), now.Format("2006-01-02"))
		}
	})

	t.Run("tomorrow", func(t *testing.T) {
		got, err := ParseDate("tomorrow", ref)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Now().AddDate(0, 0, 1)
		if !SameDay(got, want) {
			t.Errorf("ParseDate(\"tomorrow\") = %v, want tomorrow (%v)", got.Format("2006-01-02"), want.Format("2006-01-02"))
		}
	})
}
