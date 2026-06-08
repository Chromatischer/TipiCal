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

// RefreshEvents rebuilds the time grid from the current store state.
func (wv *WeekView) RefreshEvents() {
	wv.rebuildGrid()
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
	if wv.grid == nil {
		wv.selectedDay = wv.selectedDay.AddDate(0, 0, -1)
		wv.rebuildGrid()
		return
	}

	refMinutes := wv.grid.VisibleStartHour() * 60
	if ev := wv.grid.SelectedEvent(); ev != nil && !ev.AllDay {
		// Keep focus near the selected event's visual start.
		if util.SameDay(ev.Start, wv.selectedDay) {
			refMinutes = ev.Start.Hour()*60 + ev.Start.Minute()
		} else {
			refMinutes = 7 * 60
		}
	}

	newDay := wv.selectedDay.AddDate(0, 0, -1)
	oldWeekStart := util.WeekStart(wv.selectedDay, wv.startMonday)
	newWeekStart := util.WeekStart(newDay, wv.startMonday)
	oldScroll := wv.grid.ScrollY()

	wv.selectedDay = newDay
	if !util.SameDay(oldWeekStart, newWeekStart) {
		wv.rebuildGrid()
		wv.grid.SetScrollY(oldScroll)
	} else {
		wv.grid.SetSelectedDay(newDay)
	}

	wv.grid.SetSelectedDay(newDay)
	wv.grid.SelectClosestEvent(refMinutes)
}

// MoveRight moves to next day.
func (wv *WeekView) MoveRight() {
	if wv.grid == nil {
		wv.selectedDay = wv.selectedDay.AddDate(0, 0, 1)
		wv.rebuildGrid()
		return
	}

	refMinutes := wv.grid.VisibleStartHour() * 60
	if ev := wv.grid.SelectedEvent(); ev != nil && !ev.AllDay {
		if util.SameDay(ev.Start, wv.selectedDay) {
			refMinutes = ev.Start.Hour()*60 + ev.Start.Minute()
		} else {
			refMinutes = 7 * 60
		}
	}

	newDay := wv.selectedDay.AddDate(0, 0, 1)
	oldWeekStart := util.WeekStart(wv.selectedDay, wv.startMonday)
	newWeekStart := util.WeekStart(newDay, wv.startMonday)
	oldScroll := wv.grid.ScrollY()

	wv.selectedDay = newDay
	if !util.SameDay(oldWeekStart, newWeekStart) {
		wv.rebuildGrid()
		wv.grid.SetScrollY(oldScroll)
	} else {
		wv.grid.SetSelectedDay(newDay)
	}

	wv.grid.SetSelectedDay(newDay)
	wv.grid.SelectClosestEvent(refMinutes)
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

// HitTestAt returns the event/day at position (relX, relY) within the view content area.
func (wv *WeekView) HitTestAt(relX, relY int) (*ical.Event, time.Time, string) {
	if wv.grid == nil {
		return nil, time.Time{}, ""
	}
	return wv.grid.HitTestAt(relX, relY)
}

// SetSelectedDay changes the selected day without rebuilding the grid.
func (wv *WeekView) SetSelectedDay(day time.Time) {
	wv.selectedDay = day
	if wv.grid != nil {
		wv.grid.SetSelectedDay(day)
	}
}

// SelectEventByUID selects an event in the current day by UID.
func (wv *WeekView) SelectEventByUID(uid string) {
	if wv.grid != nil {
		wv.grid.SelectEventByUID(uid)
	}
}
