package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ParseTime parses a user-supplied time string flexibly.
// Supported formats:
//   - "HH:MM" or "H:MM" or "H:M" (e.g., "09:00", "9:30")
//   - "H" or "HH" (e.g., "9" → 09:00, "14" → 14:00)
//   - Natural language: "morning" (09:00), "noon"/"midday" (12:00),
//     "afternoon" (14:00), "evening" (18:00), "night" (20:00), "midnight" (00:00)
//
// Returns a time.Time with only the hour/minute fields meaningful.
func ParseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case "morning":
		return time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), nil
	case "noon", "midday":
		return time.Date(0, 1, 1, 12, 0, 0, 0, time.Local), nil
	case "afternoon":
		return time.Date(0, 1, 1, 14, 0, 0, 0, time.Local), nil
	case "evening":
		return time.Date(0, 1, 1, 18, 0, 0, 0, time.Local), nil
	case "night":
		return time.Date(0, 1, 1, 20, 0, 0, 0, time.Local), nil
	case "midnight":
		return time.Date(0, 1, 1, 0, 0, 0, 0, time.Local), nil
	}

	// Standard "HH:MM"
	if t, err := time.Parse("15:04", s); err == nil {
		return t, nil
	}

	// "H:MM" or "H:M" — manual split handles any digit count
	if idx := strings.Index(s, ":"); idx >= 0 {
		h, err1 := strconv.Atoi(s[:idx])
		m, err2 := strconv.Atoi(s[idx+1:])
		if err1 == nil && err2 == nil && h >= 0 && h <= 23 && m >= 0 && m <= 59 {
			return time.Date(0, 1, 1, h, m, 0, 0, time.Local), nil
		}
	}

	// Bare hour: "9" → 09:00 (only plain digits, no sign prefix)
	if h, err := strconv.Atoi(s); err == nil && h >= 0 && h <= 23 && s[0] != '+' && s[0] != '-' {
		return time.Date(0, 1, 1, h, 0, 0, 0, time.Local), nil
	}

	// Compound expression: <time>(+|-)<offset>, e.g. "09:00+1", "14:30-0:30", "morning+1:30"
	for i := 1; i < len(s); i++ {
		op := s[i]
		if op != '+' && op != '-' {
			continue
		}
		base, err := ParseTime(s[:i])
		if err != nil {
			continue
		}
		rest := s[i+1:]
		if rest == "" {
			continue
		}
		d, ok := parseTimeOffset(rest)
		if !ok {
			continue
		}
		if op == '-' {
			d = -d
		}
		return base.Add(d), nil
	}

	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}

// parseTimeOffset converts an offset string (without leading sign) to a duration.
// Delegates to ParseDuration by prepending "+".
func parseTimeOffset(s string) (time.Duration, bool) {
	ref := time.Date(0, 1, 1, 0, 0, 0, 0, time.Local)
	t, ok := ParseDuration("+"+s, ref)
	if !ok {
		return 0, false
	}
	return t.Sub(ref), true
}

// ParseDate parses a user-supplied date string flexibly, using ref for defaults.
// The primary format is DD.MM.YYYY (European style).
// Supported formats:
//   - "today" → today's date
//   - "tomorrow" → tomorrow's date
//   - "DD" → that day in ref's month and year
//   - "DD.MM" → that day and month in ref's year
//   - "DD.MM.YYYY" → full date
//
// Returns a time.Time with only the date fields meaningful (time zeroed).
func ParseDate(s string, ref time.Time) (time.Time, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case "today":
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), nil
	case "tomorrow":
		t := time.Now().AddDate(0, 0, 1)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
	}

	parts := strings.Split(s, ".")
	switch len(parts) {
	case 1:
		// Just a day number
		d, err := strconv.Atoi(parts[0])
		if err == nil && d >= 1 && d <= 31 {
			return time.Date(ref.Year(), ref.Month(), d, 0, 0, 0, 0, time.Local), nil
		}
	case 2:
		// DD.MM — assume ref's year
		d, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil && d >= 1 && d <= 31 && m >= 1 && m <= 12 {
			return time.Date(ref.Year(), time.Month(m), d, 0, 0, 0, 0, time.Local), nil
		}
	case 3:
		// DD.MM.YYYY
		d, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		y, err3 := strconv.Atoi(parts[2])
		if err1 == nil && err2 == nil && err3 == nil && d >= 1 && d <= 31 && m >= 1 && m <= 12 {
			return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local), nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse date %q", s)
}

// ParseDuration parses a "+..." or "-..." relative end offset and applies it to start.
// Supported formats:
//   - "+2"    → +2 hours
//   - "-2"    → -2 hours
//   - "+2h"   → +2 hours
//   - "+30m"  → +30 minutes
//   - "+2d"   → +2 days
//   - "+1:30" → +1 hour 30 minutes
//   - "+1.5"  → +1.5 hours (= 1h30m)
//
// Returns the computed end time and true, or zero and false if not a duration.
func ParseDuration(s string, start time.Time) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	sign := 1
	switch s[0] {
	case '+':
		s = strings.TrimSpace(s[1:])
	case '-':
		sign = -1
		s = strings.TrimSpace(s[1:])
	default:
		return time.Time{}, false
	}
	if s == "" {
		return time.Time{}, false
	}

	// "H:MM" — hours and minutes separated by colon
	if idx := strings.Index(s, ":"); idx >= 0 {
		h, err1 := strconv.Atoi(s[:idx])
		m, err2 := strconv.Atoi(s[idx+1:])
		if err1 == nil && err2 == nil && h >= 0 && m >= 0 && m < 60 {
			d := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute
			return start.Add(time.Duration(sign) * d), true
		}
		return time.Time{}, false
	}

	// "N.M" — decimal hours (e.g. "1.5" = 1h30m)
	if strings.Contains(s, ".") {
		hours, err := strconv.ParseFloat(s, 64)
		if err == nil && hours > 0 {
			mins := math.Round(hours * 60)
			return start.Add(time.Duration(sign) * time.Duration(mins) * time.Minute), true
		}
		return time.Time{}, false
	}

	// Unit suffixes: d, h, m
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err == nil && n > 0 {
			return start.AddDate(0, 0, sign*n), true
		}
		return time.Time{}, false
	}
	if strings.HasSuffix(s, "h") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "h"))
		if err == nil && n > 0 {
			return start.Add(time.Duration(sign) * time.Duration(n) * time.Hour), true
		}
		return time.Time{}, false
	}
	if strings.HasSuffix(s, "m") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "m"))
		if err == nil && n > 0 {
			return start.Add(time.Duration(sign) * time.Duration(n) * time.Minute), true
		}
		return time.Time{}, false
	}

	// Bare integer: "2" → 2 hours
	n, err := strconv.Atoi(s)
	if err == nil && n > 0 {
		return start.Add(time.Duration(sign) * time.Duration(n) * time.Hour), true
	}

	return time.Time{}, false
}
