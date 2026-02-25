package views

import (
	"time"

	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui/components"
)

// DayView renders a single day with detailed time slots.
type DayView struct {
	theme       *config.Theme
	store       *ical.Store
	grid        *components.TimeGrid
	selectedDay time.Time
	width       int
	height      int
	use24h      bool
}

// NewDayView creates a new day view.
func NewDayView(theme *config.Theme, store *ical.Store, cfg *config.Config) *DayView {
	now := time.Now()
	dv := &DayView{
		theme:       theme,
		store:       store,
		selectedDay: now,
		use24h:      cfg.Use24h(),
	}
	dv.rebuildGrid()
	return dv
}

// RefreshEvents rebuilds the time grid from the current store state.
func (dv *DayView) RefreshEvents() {
	dv.rebuildGrid()
}

func (dv *DayView) rebuildGrid() {
	days := []time.Time{dv.selectedDay}
	dv.grid = components.NewTimeGrid(dv.theme, days, dv.store, dv.use24h)
	dv.grid.SetSelected(dv.selectedDay)
	dv.grid.ResetEventSelection()
	if dv.width > 0 && dv.height > 0 {
		dv.grid.SetSize(dv.width, dv.height)
	}
}

// SetSize sets the view dimensions.
func (dv *DayView) SetSize(w, h int) {
	dv.width = w
	dv.height = h
	if dv.grid != nil {
		dv.grid.SetSize(w, h)
	}
}

// SelectedDate returns the currently selected date.
func (dv *DayView) SelectedDate() time.Time {
	return dv.selectedDay
}

// SetDate navigates to a specific date.
func (dv *DayView) SetDate(d time.Time) {
	dv.selectedDay = d
	dv.rebuildGrid()
}

// NextPeriod advances by 1 day.
func (dv *DayView) NextPeriod() {
	dv.selectedDay = dv.selectedDay.AddDate(0, 0, 1)
	dv.rebuildGrid()
}

// PrevPeriod goes back 1 day.
func (dv *DayView) PrevPeriod() {
	dv.selectedDay = dv.selectedDay.AddDate(0, 0, -1)
	dv.rebuildGrid()
}

// MoveUp selects the previous event.
func (dv *DayView) MoveUp() {
	dv.grid.SelectPrevEvent()
}

// MoveDown selects the next event.
func (dv *DayView) MoveDown() {
	dv.grid.SelectNextEvent()
}

// MoveLeft goes to previous day.
func (dv *DayView) MoveLeft() {
	dv.PrevPeriod()
}

// MoveRight goes to next day.
func (dv *DayView) MoveRight() {
	dv.NextPeriod()
}

// View renders the day view.
func (dv *DayView) View() string {
	if dv.grid == nil {
		return ""
	}
	return dv.grid.View()
}

// SelectedEvent returns the currently selected event, or nil.
func (dv *DayView) SelectedEvent() *ical.Event {
	if dv.grid == nil {
		return nil
	}
	return dv.grid.SelectedEvent()
}

// HitTestAt returns the event/day at position (relX, relY) within the view content area.
func (dv *DayView) HitTestAt(relX, relY int) (*ical.Event, time.Time, string) {
	if dv.grid == nil {
		return nil, time.Time{}, ""
	}
	return dv.grid.HitTestAt(relX, relY)
}

// SetSelectedDay changes the selected day without rebuilding the grid.
func (dv *DayView) SetSelectedDay(day time.Time) {
	dv.selectedDay = day
	if dv.grid != nil {
		dv.grid.SetSelectedDay(day)
	}
}

// SelectEventByUID selects an event in the current day by UID.
func (dv *DayView) SelectEventByUID(uid string) {
	if dv.grid != nil {
		dv.grid.SelectEventByUID(uid)
	}
}
