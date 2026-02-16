package util

import (
	"time"

	"github.com/mattn/go-runewidth"
)

// WeekStart returns the Monday of the week containing t.
func WeekStart(t time.Time, startMonday bool) time.Time {
	weekday := t.Weekday()
	if startMonday {
		if weekday == time.Sunday {
			weekday = 7
		}
		return t.AddDate(0, 0, -int(weekday-time.Monday))
	}
	return t.AddDate(0, 0, -int(weekday))
}

// MonthGrid returns a 6x7 grid of dates for the month view.
// Each row is a week, starting from the configured week start day.
func MonthGrid(year int, month time.Month, startMonday bool) [6][7]time.Time {
	var grid [6][7]time.Time

	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	gridStart := WeekStart(firstOfMonth, startMonday)

	// If grid starts on the 1st but shouldn't, go back a week
	if gridStart.After(firstOfMonth) {
		gridStart = gridStart.AddDate(0, 0, -7)
	}

	for row := 0; row < 6; row++ {
		for col := 0; col < 7; col++ {
			grid[row][col] = gridStart.AddDate(0, 0, row*7+col)
		}
	}
	return grid
}

// SameDay returns true if a and b are on the same calendar day.
func SameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// FormatTime formats a time according to 24h or 12h preference.
func FormatTime(t time.Time, use24h bool) string {
	if use24h {
		return t.Format("15:04")
	}
	return t.Format("3:04 PM")
}

// FormatTimeRange formats a start-end time range.
func FormatTimeRange(start, end time.Time, use24h bool) string {
	return FormatTime(start, use24h) + " - " + FormatTime(end, use24h)
}

// DaysInMonth returns the number of days in the given month.
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}

// TruncateText truncates text to maxLen display columns, adding ellipsis if needed.
func TruncateText(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return runewidth.Truncate(s, maxLen, "…")
}

// WrapText wraps text to the specified width (in display columns), respecting word boundaries
// and preserving existing newlines. Returns a slice of lines.
func WrapText(s string, width int) []string {
	if width <= 0 {
		return []string{""}
	}

	var result []string
	lines := splitLines(s)

	for _, line := range lines {
		if runewidth.StringWidth(line) <= width {
			result = append(result, line)
			continue
		}

		// Wrap long line
		wrapped := wrapLine(line, width)
		result = append(result, wrapped...)
	}

	return result
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" || len(lines) == 0 {
		lines = append(lines, current)
	}
	return lines
}

func wrapLine(line string, width int) []string {
	var result []string
	runes := []rune(line)

	for len(runes) > 0 {
		// Find the longest prefix that fits
		end := 0
		currentWidth := 0
		lastSpace := -1

		for i, r := range runes {
			w := runewidth.RuneWidth(r)
			if currentWidth+w > width {
				break
			}
			currentWidth += w
			end = i + 1

			if r == ' ' || r == '\t' {
				lastSpace = i
			}
		}

		if end == 0 {
			// Single character is too wide - force it anyway
			end = 1
		} else if end < len(runes) && lastSpace > 0 {
			// Break at last space if not at end
			end = lastSpace + 1
		}

		// Extract line and trim trailing space if we broke at a space
		lineText := string(runes[:end])
		if end < len(runes) && lastSpace > 0 && lastSpace+1 == end {
			lineText = string(runes[:lastSpace])
			// Skip the space for next line
		} else if end < len(runes) {
			// Didn't break at space, don't skip anything
		}

		result = append(result, lineText)
		runes = runes[end:]
	}

	if len(result) == 0 {
		result = append(result, "")
	}

	return result
}

// HoursBetween returns start and end hours covering a set of events.
func HoursBetween(events []struct{ Start, End time.Time }, defaultStart, defaultEnd int) (int, int) {
	minH := defaultStart
	maxH := defaultEnd
	for _, e := range events {
		h := e.Start.Hour()
		if h < minH {
			minH = h
		}
		h = e.End.Hour()
		if e.End.Minute() > 0 {
			h++
		}
		if h > maxH {
			maxH = h
		}
	}
	if maxH > 24 {
		maxH = 24
	}
	return minH, maxH
}
