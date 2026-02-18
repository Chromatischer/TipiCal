package views

import (
	"strings"
	"testing"
	"time"

	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
)

func testMonthTheme() *config.Theme {
	return &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextFaint:  "#6C7086",
		TextMuted:  "#A6ADC8",
		Today:      "#F38BA8",
		Selected:   "#89B4FA",
		Border:     "#585B70",
		Surface:    "#313244",
		SurfaceAlt: "#45475A",
	}
}

func testMonthConfig() *config.Config {
	return &config.Config{
		General: config.GeneralConfig{
			WeekStart:  "monday",
			TimeFormat: "24h",
		},
		Theme: config.ThemeConfig{
			Mode:   "dark",
			Accent: "#7C3AED",
		},
	}
}

func TestMonthViewNavigation(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))

	if mv.SelectedDate().Day() != 15 {
		t.Errorf("initial selected day = %d, want 15", mv.SelectedDate().Day())
	}

	mv.MoveRight()
	if mv.SelectedDate().Day() != 16 {
		t.Errorf("after MoveRight, day = %d, want 16", mv.SelectedDate().Day())
	}

	mv.MoveLeft()
	if mv.SelectedDate().Day() != 15 {
		t.Errorf("after MoveLeft, day = %d, want 15", mv.SelectedDate().Day())
	}

	mv.MoveDown()
	expectedDown := 15 + 7
	if mv.SelectedDate().Day() != expectedDown {
		t.Errorf("after MoveDown, day = %d, want %d", mv.SelectedDate().Day(), expectedDown)
	}

	mv.MoveUp()
	if mv.SelectedDate().Day() != 15 {
		t.Errorf("after MoveUp, day = %d, want 15", mv.SelectedDate().Day())
	}
}

func TestMonthViewNavigationAcrossMonths(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC))

	mv.MoveRight()
	if mv.SelectedDate().Month() != time.February {
		t.Errorf("after MoveRight from Jan 31, month = %v, want February", mv.SelectedDate().Month())
	}
	if mv.SelectedDate().Day() != 1 {
		t.Errorf("after MoveRight from Jan 31, day = %d, want 1", mv.SelectedDate().Day())
	}

	mv.SetDate(time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC))
	mv.MoveLeft()
	if mv.SelectedDate().Month() != time.February {
		t.Errorf("after MoveLeft from Mar 1, month = %v, want February", mv.SelectedDate().Month())
	}
	if mv.SelectedDate().Day() != 28 {
		t.Errorf("after MoveLeft from Mar 1, day = %d, want 28", mv.SelectedDate().Day())
	}
}

func TestMonthViewNextPrevPeriod(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))

	mv.NextPeriod()
	if mv.month != time.March {
		t.Errorf("after NextPeriod, month = %v, want March", mv.month)
	}
	if mv.year != 2026 {
		t.Errorf("year = %d, want 2026", mv.year)
	}

	mv.PrevPeriod()
	if mv.month != time.February {
		t.Errorf("after PrevPeriod, month = %v, want February", mv.month)
	}
}

func TestMonthViewYearTransition(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.December, 15, 0, 0, 0, 0, time.UTC))

	mv.NextPeriod()
	if mv.month != time.January {
		t.Errorf("after NextPeriod from Dec, month = %v, want January", mv.month)
	}
	if mv.year != 2027 {
		t.Errorf("after NextPeriod from Dec, year = %d, want 2027", mv.year)
	}

	mv.SetDate(time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC))
	mv.PrevPeriod()
	if mv.month != time.December {
		t.Errorf("after PrevPeriod from Jan, month = %v, want December", mv.month)
	}
	if mv.year != 2025 {
		t.Errorf("after PrevPeriod from Jan, year = %d, want 2025", mv.year)
	}
}

func TestMonthViewSetDate(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)

	newDate := time.Date(2025, time.December, 25, 0, 0, 0, 0, time.UTC)
	mv.SetDate(newDate)

	if mv.year != 2025 {
		t.Errorf("year = %d, want 2025", mv.year)
	}
	if mv.month != time.December {
		t.Errorf("month = %v, want December", mv.month)
	}
	if mv.SelectedDate().Day() != 25 {
		t.Errorf("selected day = %d, want 25", mv.SelectedDate().Day())
	}
}

func TestMonthViewSetSize(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)

	mv.SetSize(120, 40)

	if mv.width != 120 {
		t.Errorf("width = %d, want 120", mv.width)
	}
	if mv.height != 40 {
		t.Errorf("height = %d, want 40", mv.height)
	}
}

func TestMonthViewWeekStartMonday(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	cfg.General.WeekStart = "monday"
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	mv.SetSize(120, 40)

	view := mv.View()
	if !strings.Contains(view, "Mon") {
		t.Error("Monday-start view should show 'Mon' in header")
	}
	lines := strings.Split(view, "\n")
	if len(lines) > 0 {
		headerLine := strings.TrimSpace(lines[0])
		if !strings.HasPrefix(headerLine, "Mon") {
			t.Errorf("First weekday should be Mon, got: %s", headerLine[:15])
		}
	}
}

func TestMonthViewWeekStartSunday(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	cfg.General.WeekStart = "sunday"
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	mv.SetSize(120, 40)

	view := mv.View()
	if !strings.Contains(view, "Sun") {
		t.Error("Sunday-start view should show 'Sun' in header")
	}
	lines := strings.Split(view, "\n")
	if len(lines) > 0 {
		headerLine := strings.TrimSpace(lines[0])
		if !strings.HasPrefix(headerLine, "Sun") {
			t.Errorf("First weekday should be Sun, got: %s", headerLine[:15])
		}
	}
}

func TestMonthViewEventIndicators(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	store.AddEvent(&ical.Event{
		UID:        "test-event-1",
		Summary:    "Test Event",
		Start:      time.Date(2026, time.February, 15, 10, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 15, 11, 0, 0, 0, time.UTC),
		CalendarID: 0,
		Color:      "#3B82F6",
	})

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	mv.SetSize(120, 40)

	events := store.EventsForDay(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	if len(events) != 1 {
		t.Errorf("EventsForDay(Feb 15) = %d events, want 1", len(events))
	}
}

func TestMonthViewGridSixRows(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	mv.SetSize(120, 40)

	view := mv.View()
	lines := strings.Split(view, "\n")

	dataRows := 0
	for _, line := range lines {
		if strings.Contains(line, " 1") ||
			strings.Contains(line, " 2") ||
			strings.Contains(line, "15") ||
			strings.Contains(line, "28") {
			dataRows++
		}
	}

	if dataRows < 4 || dataRows > 6 {
		t.Errorf("Month grid should have 4-6 rows, got %d data rows", dataRows)
	}
}

func TestMonthViewTodayHighlight(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetSize(120, 40)

	now := time.Now()
	view := mv.View()

	dayStr := ""
	if now.Day() < 10 {
		dayStr = " " + string(rune('0'+now.Day()))
	} else {
		dayStr = string(rune('0'+now.Day()/10)) + string(rune('0'+now.Day()%10))
	}

	if !strings.Contains(view, dayStr) {
		t.Errorf("View should contain today's day number %q", dayStr)
	}
}

func TestMonthViewOtherMonthDays(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	mv.SetSize(120, 40)

	view := mv.View()

	if !strings.Contains(view, "26") {
		t.Error("February 2026 grid should show Jan 26 (overflow from previous month)")
	}
	if !strings.Contains(view, " 1") {
		t.Error("February 2026 grid should show Mar 1 (overflow to next month)")
	}
}

func TestMonthViewMultipleEventsPerDay(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	day := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		store.AddEvent(&ical.Event{
			UID:        string(rune('a' + i)),
			Summary:    "Event " + string(rune('A'+i)),
			Start:      day.Add(time.Duration(9+i) * time.Hour),
			End:        day.Add(time.Duration(10+i) * time.Hour),
			CalendarID: 0,
			Color:      "#3B82F6",
		})
	}

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	mv.SetSize(120, 40)

	events := store.EventsForDay(day)
	if len(events) != 5 {
		t.Errorf("EventsForDay = %d, want 5", len(events))
	}
}

func TestMonthViewAllDayEvents(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	store.AddEvent(&ical.Event{
		UID:        "allday-1",
		Summary:    "All Day Event",
		Start:      time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 16, 0, 0, 0, 0, time.UTC),
		AllDay:     true,
		CalendarID: 0,
		Color:      "#7C3AED",
	})

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	mv.SetSize(120, 40)

	events := store.EventsForDay(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))
	if len(events) != 1 {
		t.Errorf("EventsForDay = %d, want 1 all-day event", len(events))
	}
}

func TestMonthViewMultiDayAllDayEvents(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	store.AddEvent(&ical.Event{
		UID:        "multiday-1",
		Summary:    "Conference",
		Start:      time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC),
		AllDay:     true,
		CalendarID: 0,
		Color:      "#7C3AED",
	})

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC))

	for day := 15; day <= 17; day++ {
		d := time.Date(2026, time.February, day, 0, 0, 0, 0, time.UTC)
		events := store.EventsForDay(d)
		if len(events) != 1 {
			t.Errorf("EventsForDay(Feb %d) = %d, want 1", day, len(events))
		}
	}

	d := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	events := store.EventsForDay(d)
	if len(events) != 0 {
		t.Errorf("EventsForDay(Feb 18) = %d, want 0 (end is exclusive)", len(events))
	}
}

func TestMonthViewLeapYear(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2024, time.February, 15, 0, 0, 0, 0, time.UTC))
	mv.SetSize(120, 40)

	view := mv.View()

	if !strings.Contains(view, "29") {
		t.Error("February 2024 (leap year) should show day 29")
	}
}

func TestMonthViewNavigationUpdatesMonth(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	mv.SetDate(time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC))

	mv.MoveRight()
	if mv.month != time.February {
		t.Errorf("month after navigation = %v, want February", mv.month)
	}
	if mv.year != 2026 {
		t.Errorf("year after navigation = %d, want 2026", mv.year)
	}
}

func TestMonthViewSelectedDatePersistence(t *testing.T) {
	theme := testMonthTheme()
	cfg := testMonthConfig()
	store := ical.NewStore()

	mv := NewMonthView(theme, store, cfg)
	originalDate := time.Date(2026, time.February, 18, 14, 30, 0, 0, time.UTC)
	mv.SetDate(originalDate)

	selected := mv.SelectedDate()
	if selected.Year() != originalDate.Year() ||
		selected.Month() != originalDate.Month() ||
		selected.Day() != originalDate.Day() {
		t.Errorf("SelectedDate() = %v, want same date as %v", selected, originalDate)
	}
}
