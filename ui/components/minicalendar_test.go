package components

import (
	"strings"
	"testing"
	"time"

	"github.com/tipical/tipical/config"
)

func testTheme() *config.Theme {
	return &config.Theme{
		IsDark:     true,
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextFaint:  "#6C7086",
		TextMuted:  "#A6ADC8",
		Today:      "#F38BA8",
		Selected:   "#89B4FA",
		SurfaceAlt: "#45475A",
	}
}

func TestMiniCalendarSetDate(t *testing.T) {
	theme := testTheme()
	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	mc := NewMiniCalendar(theme, date, true)

	if mc.year != 2026 {
		t.Errorf("year = %d, want 2026", mc.year)
	}
	if mc.month != time.February {
		t.Errorf("month = %v, want February", mc.month)
	}
	if !mc.selectedDay.Equal(date) {
		t.Errorf("selectedDay = %v, want %v", mc.selectedDay, date)
	}

	newDate := time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)
	mc.SetDate(newDate)

	if mc.year != 2026 {
		t.Errorf("year after SetDate = %d, want 2026", mc.year)
	}
	if mc.month != time.March {
		t.Errorf("month after SetDate = %v, want March", mc.month)
	}
	if !mc.selectedDay.Equal(newDate) {
		t.Errorf("selectedDay after SetDate = %v, want %v", mc.selectedDay, newDate)
	}
}

func TestMiniCalendarWeekStartMonday(t *testing.T) {
	theme := testTheme()
	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	mc := NewMiniCalendar(theme, date, true)

	view := mc.View()
	if !strings.Contains(view, "Mo") {
		t.Error("Monday-start calendar should show 'Mo' as first weekday")
	}
	if !strings.Contains(view, "Tu") {
		t.Error("Calendar should show 'Tu' as second weekday")
	}
}

func TestMiniCalendarWeekStartSunday(t *testing.T) {
	theme := testTheme()
	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	mc := NewMiniCalendar(theme, date, false)

	view := mc.View()
	if !strings.Contains(view, "Su") {
		t.Error("Sunday-start calendar should show 'Su' as first weekday")
	}
	lines := strings.Split(view, "\n")
	weekdayLine := lines[1]
	if !strings.HasPrefix(strings.TrimSpace(weekdayLine), "Su") {
		t.Errorf("First weekday should be 'Su', got %q", weekdayLine[:15])
	}
}

func TestMiniCalendarMonthYearHeader(t *testing.T) {
	theme := testTheme()
	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	mc := NewMiniCalendar(theme, date, true)

	view := mc.View()
	if !strings.Contains(view, "Feb 2026") {
		t.Error("Calendar should show 'Feb 2026' in header")
	}

	mc.SetDate(time.Date(2025, time.December, 25, 0, 0, 0, 0, time.UTC))
	view = mc.View()
	if !strings.Contains(view, "Dec 2025") {
		t.Error("Calendar should show 'Dec 2025' in header")
	}
}

func TestMiniCalendarTodayHighlight(t *testing.T) {
	theme := testTheme()
	now := time.Now()
	mc := NewMiniCalendar(theme, now, true)

	view := mc.View()
	dayNum := now.Day()

	dayStr := ""
	if dayNum < 10 {
		dayStr = " " + string(rune('0'+dayNum))
	} else {
		dayStr = string(rune('0'+dayNum/10)) + string(rune('0'+dayNum%10))
	}

	if !strings.Contains(view, dayStr) {
		t.Errorf("Calendar should contain today's day number %q", dayStr)
	}
}

func TestMiniCalendarSelectedDayHighlight(t *testing.T) {
	theme := testTheme()
	selected := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	mc := NewMiniCalendar(theme, selected, true)

	if !mc.selectedDay.Equal(selected) {
		t.Errorf("selectedDay = %v, want %v", mc.selectedDay, selected)
	}
}

func TestMiniCalendarOtherMonthDays(t *testing.T) {
	theme := testTheme()
	date := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	mc := NewMiniCalendar(theme, date, true)

	view := mc.View()

	if !strings.Contains(view, "26") {
		t.Error("February 2026 calendar should show Jan 26 (previous month overflow)")
	}
	if !strings.Contains(view, " 1") {
		t.Error("February 2026 calendar should show Mar 1 (next month overflow)")
	}
}

func TestMiniCalendarWidth(t *testing.T) {
	theme := testTheme()
	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	mc := NewMiniCalendar(theme, date, true)

	view := mc.View()
	lines := strings.Split(view, "\n")

	for i, line := range lines {
		width := len(line)
		if width > 25 {
			t.Errorf("Line %d width = %d, should be <= 25 (mini calendar is ~20 chars wide)", i, width)
		}
	}
}

func TestMiniCalendarSixWeeks(t *testing.T) {
	theme := testTheme()
	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	mc := NewMiniCalendar(theme, date, true)

	view := mc.View()
	lines := strings.Split(view, "\n")

	dataLines := 0
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) != "" {
			dataLines++
		}
	}

	if dataLines < 4 || dataLines > 6 {
		t.Errorf("Calendar should have 4-6 weeks of data, got %d lines", dataLines)
	}
}
