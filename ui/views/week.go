package views

import (
	"time"

	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui/components"
	"github.com/tipical/tipical/util"
)

// WeekView renders a 7-day view with hourly time slots.
type WeekView struct {
	theme       *config.Theme
	store       *ical.Store
	grid        *components.TimeGrid
	selectedDay time.Time
	width       int
	height      int
	use24h      bool
	startMonday bool
}

// NewWeekView creates a new week view.
func NewWeekView(theme *config.Theme, store *ical.Store, cfg *config.Config) *WeekView {
	now := time.Now()
	wv := &WeekView{
		theme:       theme,
		store:       store,
		selectedDay: now,
		use24h:      cfg.Use24h(),
		startMonday: cfg.StartMonday(),
	}
	wv.rebuildGrid()
	return wv
}

func (wv *WeekView) rebuildGrid() {
	weekStart := util.WeekStart(wv.selectedDay, wv.startMonday)
	days := make([]time.Time, 7)
	for i := 0; i < 7; i++ {
		days[i] = weekStart.AddDate(0, 0, i)
	}
	wv.grid = components.NewTimeGrid(wv.theme, days, wv.store, wv.use24h)
	wv.grid.SetSelected(wv.selectedDay)
	wv.grid.ResetEventSelection()
	if wv.width > 0 && wv.height > 0 {
		wv.grid.SetSize(wv.width, wv.height)
	}
}

// SetSize sets the view dimensions.
func (wv *WeekView) SetSize(w, h int) {
	wv.width = w
	wv.height = h
	if wv.grid != nil {
		wv.grid.SetSize(w, h)
	}
}

// SelectedDate returns the currently selected date.
func (wv *WeekView) SelectedDate() time.Time {
	return wv.selectedDay
}

// SetDate navigates to a specific date.
func (wv *WeekView) SetDate(d time.Time) {
	wv.selectedDay = d
	wv.rebuildGrid()
}

// NextPeriod advances to the next week.
func (wv *WeekView) NextPeriod() {
	wv.selectedDay = wv.selectedDay.AddDate(0, 0, 7)
	wv.rebuildGrid()
}

// PrevPeriod goes to the previous week.
func (wv *WeekView) PrevPeriod() {
	wv.selectedDay = wv.selectedDay.AddDate(0, 0, -7)
	wv.rebuildGrid()
}

// MoveUp selects the previous event on the selected day.
func (wv *WeekView) MoveUp() {
	wv.grid.SelectPrevEvent()
}

// MoveDown selects the next event on the selected day.
func (wv *WeekView) MoveDown() {
	wv.grid.SelectNextEvent()
}

// MoveLeft moves to previous day.
func (wv *WeekView) MoveLeft() {
	wv.selectedDay = wv.selectedDay.AddDate(0, 0, -1)
	wv.rebuildGrid()
}

// MoveRight moves to next day.
func (wv *WeekView) MoveRight() {
	wv.selectedDay = wv.selectedDay.AddDate(0, 0, 1)
	wv.rebuildGrid()
}

// View renders the week view.
func (wv *WeekView) View() string {
	if wv.grid == nil {
		return ""
	}
	return wv.grid.View()
}

// SelectedEvent returns the currently selected event, or nil.
func (wv *WeekView) SelectedEvent() *ical.Event {
	if wv.grid == nil {
		return nil
	}
	return wv.grid.SelectedEvent()
}
