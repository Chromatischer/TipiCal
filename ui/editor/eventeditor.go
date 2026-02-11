package editor

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ical"
	"github.com/terminal-ical/terminal-ical/ui/components"
)

// Mode represents the editor mode.
type Mode int

const (
	ModeCreate Mode = iota
	ModeEdit
	ModeDelete
)

// EventEditor manages the event creation/editing form.
type EventEditor struct {
	theme       *config.Theme
	store       *ical.Store
	form        *components.Form
	deleteModal *components.Modal
	mode        Mode
	editingUID  string
	calendarID  int
	date        time.Time
	width       int
	height      int
	active      bool
	result      *EditorResult
}

// EditorResult holds the result of an editor action.
type EditorResult struct {
	Action string // "create", "update", "delete", "cancel"
	Event  *ical.Event
}

// NewEventEditor creates a new event editor.
func NewEventEditor(theme *config.Theme, store *ical.Store) *EventEditor {
	return &EventEditor{
		theme: theme,
		store: store,
	}
}

// OpenCreate opens the editor in create mode.
func (ee *EventEditor) OpenCreate(date time.Time, calendarNames []string) {
	ee.mode = ModeCreate
	ee.date = date
	ee.active = true
	ee.result = nil

	defaultStart := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.Local)
	defaultEnd := defaultStart.Add(time.Hour)

	calOptions := calendarNames
	if len(calOptions) == 0 {
		calOptions = []string{"Work", "Personal"}
	}

	fields := []components.FormField{
		{Label: "Title", Type: components.FieldText, Placeholder: "Event title"},
		{Label: "Date", Type: components.FieldText, Value: date.Format("2006-01-02")},
		{Label: "Start", Type: components.FieldText, Value: defaultStart.Format("15:04")},
		{Label: "End", Type: components.FieldText, Value: defaultEnd.Format("15:04")},
		{Label: "Location", Type: components.FieldText, Placeholder: "Location"},
		{Label: "Calendar", Type: components.FieldSelect, Options: calOptions},
		{Label: "Notes", Type: components.FieldTextArea, Placeholder: "Description"},
	}

	ee.form = components.NewForm(ee.theme, "New Event", fields)
}

// OpenEdit opens the editor in edit mode for an existing event.
func (ee *EventEditor) OpenEdit(event *ical.Event, calendarNames []string) {
	ee.mode = ModeEdit
	ee.editingUID = event.UID
	ee.active = true
	ee.result = nil

	calOptions := calendarNames
	if len(calOptions) == 0 {
		calOptions = []string{"Work", "Personal"}
	}

	fields := []components.FormField{
		{Label: "Title", Type: components.FieldText, Value: event.Summary},
		{Label: "Date", Type: components.FieldText, Value: event.Start.Format("2006-01-02")},
		{Label: "Start", Type: components.FieldText, Value: event.Start.Format("15:04")},
		{Label: "End", Type: components.FieldText, Value: event.End.Format("15:04")},
		{Label: "Location", Type: components.FieldText, Value: event.Location},
		{Label: "Calendar", Type: components.FieldSelect, Options: calOptions, Selected: event.CalendarID},
		{Label: "Notes", Type: components.FieldTextArea, Value: event.Description},
	}

	ee.form = components.NewForm(ee.theme, "Edit Event", fields)
}

// OpenDelete opens the delete confirmation modal.
func (ee *EventEditor) OpenDelete(event *ical.Event) {
	ee.mode = ModeDelete
	ee.editingUID = event.UID
	ee.active = true
	ee.result = nil

	ee.deleteModal = components.NewModal(ee.theme,
		"Delete Event",
		fmt.Sprintf("Are you sure you want to delete \"%s\"?", event.Summary),
		[]components.ModalAction{
			{Label: "Cancel", Key: "esc"},
			{Label: "Delete", Key: "d", Danger: true},
		},
	)
}

// IsActive returns true if the editor is open.
func (ee *EventEditor) IsActive() bool {
	return ee.active
}

// Result returns the editor result (non-nil when a decision was made).
func (ee *EventEditor) Result() *EditorResult {
	return ee.result
}

// ClearResult clears the result.
func (ee *EventEditor) ClearResult() {
	ee.result = nil
}

// SetSize sets the editor dimensions.
func (ee *EventEditor) SetSize(w, h int) {
	ee.width = w
	ee.height = h
	if ee.form != nil {
		ee.form.SetSize(w/2, h-4)
	}
}

// HandleKey processes a key press.
func (ee *EventEditor) HandleKey(key string) {
	if !ee.active {
		return
	}

	switch ee.mode {
	case ModeCreate, ModeEdit:
		ee.form.HandleKey(key)

		if ee.form.IsSubmitted() {
			event := ee.buildEventFromForm()
			if ee.mode == ModeCreate {
				ee.result = &EditorResult{Action: "create", Event: event}
			} else {
				ee.result = &EditorResult{Action: "update", Event: event}
			}
			ee.active = false
		} else if ee.form.IsCancelled() {
			ee.result = &EditorResult{Action: "cancel"}
			ee.active = false
		}

	case ModeDelete:
		action := ee.deleteModal.HandleKey(key)
		switch action {
		case 0: // Cancel
			ee.result = &EditorResult{Action: "cancel"}
			ee.active = false
		case 1: // Delete
			existing := ee.store.FindEvent(ee.editingUID)
			ee.result = &EditorResult{Action: "delete", Event: existing}
			ee.active = false
		case -1: // Esc
			ee.result = &EditorResult{Action: "cancel"}
			ee.active = false
		}
	}
}

func (ee *EventEditor) buildEventFromForm() *ical.Event {
	fields := ee.form.Fields()

	title := fields[0].Value
	dateStr := fields[1].Value
	startStr := fields[2].Value
	endStr := fields[3].Value
	location := fields[4].Value
	calIdx := fields[5].Selected
	notes := fields[6].Value

	// Parse date and time
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	start, err := time.Parse("15:04", startStr)
	if err != nil {
		start = time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)
	}

	end, err := time.Parse("15:04", endStr)
	if err != nil {
		end = time.Date(0, 1, 1, 10, 0, 0, 0, time.Local)
	}

	startTime := time.Date(date.Year(), date.Month(), date.Day(),
		start.Hour(), start.Minute(), 0, 0, time.Local)
	endTime := time.Date(date.Year(), date.Month(), date.Day(),
		end.Hour(), end.Minute(), 0, 0, time.Local)

	uid := ee.editingUID
	if uid == "" {
		uid = fmt.Sprintf("tical-%d", time.Now().UnixNano())
	}

	return &ical.Event{
		UID:         uid,
		Summary:     title,
		Description: notes,
		Location:    location,
		Start:       startTime,
		End:         endTime,
		CalendarID:  calIdx,
		Status:      "CONFIRMED",
	}
}

// View renders the editor.
func (ee *EventEditor) View() string {
	if !ee.active {
		return ""
	}

	var content string
	switch ee.mode {
	case ModeCreate, ModeEdit:
		content = ee.form.View()
	case ModeDelete:
		content = ee.deleteModal.View()
	}

	return lipgloss.Place(
		ee.width,
		ee.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
