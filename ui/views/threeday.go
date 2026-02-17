package views

import (
	"time"

	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui/components"
)

// ThreeDayView renders a 3-day view with hourly time slots.
type ThreeDayView struct {
	theme       *config.Theme
	store       *ical.Store
	grid        *components.TimeGrid
	selectedDay time.Time
	width       int
	height      int
	use24h      bool
}

// NewThreeDayView creates a new 3-day view.
func NewThreeDayView(theme *config.Theme, store *ical.Store, cfg *config.Config) *ThreeDayView {
	now := time.Now()
	td := &ThreeDayView{
		theme:       theme,
		store:       store,
		selectedDay: now,
		use24h:      cfg.Use24h(),
	}
	td.rebuildGrid()
	return td
}

// RefreshEvents rebuilds the time grid from the current store state.
func (td *ThreeDayView) RefreshEvents() {
	td.rebuildGrid()
}

func (td *ThreeDayView) rebuildGrid() {
	days := make([]time.Time, 3)
	days[0] = td.selectedDay.AddDate(0, 0, -1) // yesterday
	days[1] = td.selectedDay                   // today
	days[2] = td.selectedDay.AddDate(0, 0, 1)  // tomorrow
	td.grid = components.NewTimeGrid(td.theme, days, td.store, td.use24h)
	td.grid.SetSelected(td.selectedDay)
	td.grid.ResetEventSelection()
	if td.width > 0 && td.height > 0 {
		td.grid.SetSize(td.width, td.height)
	}
}

// SetSize sets the view dimensions.
func (td *ThreeDayView) SetSize(w, h int) {
	td.width = w
	td.height = h
	if td.grid != nil {
		td.grid.SetSize(w, h)
	}
}

// SelectedDate returns the currently selected date.
func (td *ThreeDayView) SelectedDate() time.Time {
	return td.selectedDay
}

// SetDate navigates to a specific date.
func (td *ThreeDayView) SetDate(d time.Time) {
	td.selectedDay = d
	td.rebuildGrid()
}

// NextPeriod advances by 3 days.
func (td *ThreeDayView) NextPeriod() {
	td.selectedDay = td.selectedDay.AddDate(0, 0, 3)
	td.rebuildGrid()
}

// PrevPeriod goes back 3 days.
func (td *ThreeDayView) PrevPeriod() {
	td.selectedDay = td.selectedDay.AddDate(0, 0, -3)
	td.rebuildGrid()
}

// MoveUp selects the previous event on the selected day.
func (td *ThreeDayView) MoveUp() {
	td.grid.SelectPrevEvent()
}

// MoveDown selects the next event on the selected day.
func (td *ThreeDayView) MoveDown() {
	td.grid.SelectNextEvent()
}

// MoveLeft moves to previous day.
func (td *ThreeDayView) MoveLeft() {
	td.selectedDay = td.selectedDay.AddDate(0, 0, -1)
	td.rebuildGrid()
}

// MoveRight moves to next day.
func (td *ThreeDayView) MoveRight() {
	td.selectedDay = td.selectedDay.AddDate(0, 0, 1)
	td.rebuildGrid()
}

// View renders the 3-day view.
func (td *ThreeDayView) View() string {
	if td.grid == nil {
		return ""
	}
	return td.grid.View()
}

// SelectedEvent returns the currently selected event, or nil.
func (td *ThreeDayView) SelectedEvent() *ical.Event {
	if td.grid == nil {
		return nil
	}
	return td.grid.SelectedEvent()
}
