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

func (dv *DayView) rebuildGrid() {
	days := []time.Time{dv.selectedDay}
	dv.grid = components.NewTimeGrid(dv.theme, days, dv.store, dv.use24h)
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

// MoveUp scrolls up.
func (dv *DayView) MoveUp() {
	dv.grid.ScrollUp()
}

// MoveDown scrolls down.
func (dv *DayView) MoveDown() {
	dv.grid.ScrollDown()
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
