package editor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui/components"
	"github.com/tipical/tipical/util"
)

// Mode represents the editor mode.
type Mode int

const (
	ModeCreate Mode = iota
	ModeEdit
	ModeDelete
	ModeDetail
)

// Detail view button indices.
const (
	detailBtnEdit = iota
	detailBtnDelete
	detailBtnCopyPlain
	detailBtnCopyJSON
	detailBtnCount
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

	// Detail mode state
	detailEvent    *ical.Event
	detailCalName  string
	detailFocused  int
	detailUse24h   bool
	statusMsg      string
}

// EditorResult holds the result of an editor action.
type EditorResult struct {
	Action  string // "create", "update", "delete", "cancel", "edit-from-detail", "delete-from-detail", "clipboard"
	Event   *ical.Event
	Message string
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
		{Label: "All Day", Type: components.FieldSelect, Options: []string{"No", "Yes"}},
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

	allDaySelected := 0
	if event.AllDay {
		allDaySelected = 1
	}

	// Find the calendar index in the calendarNames list
	calSelected := 0
	if event.CalendarID >= 0 && event.CalendarID < len(ee.store.Calendars) {
		eventCalName := ee.store.Calendars[event.CalendarID].Name
		for i, name := range calOptions {
			if name == eventCalName {
				calSelected = i
				break
			}
		}
	}

	fields := []components.FormField{
		{Label: "Title", Type: components.FieldText, Value: event.Summary},
		{Label: "All Day", Type: components.FieldSelect, Options: []string{"No", "Yes"}, Selected: allDaySelected},
		{Label: "Date", Type: components.FieldText, Value: event.Start.Format("2006-01-02")},
		{Label: "Start", Type: components.FieldText, Value: event.Start.Format("15:04")},
		{Label: "End", Type: components.FieldText, Value: event.End.Format("15:04")},
		{Label: "Location", Type: components.FieldText, Value: event.Location},
		{Label: "Calendar", Type: components.FieldSelect, Options: calOptions, Selected: calSelected},
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

// OpenDetail opens the detail view for an event.
func (ee *EventEditor) OpenDetail(event *ical.Event, calendarName string, use24h bool) {
	ee.mode = ModeDetail
	ee.detailEvent = event
	ee.detailCalName = calendarName
	ee.detailFocused = 0
	ee.detailUse24h = use24h
	ee.editingUID = event.UID
	ee.active = true
	ee.result = nil
	ee.statusMsg = ""
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

// StatusMessage returns a pending status message (e.g. clipboard feedback).
func (ee *EventEditor) StatusMessage() string {
	return ee.statusMsg
}

// ClearStatusMessage clears the status message.
func (ee *EventEditor) ClearStatusMessage() {
	ee.statusMsg = ""
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

	case ModeDetail:
		ee.handleDetailKey(key)
	}
}

func (ee *EventEditor) handleDetailKey(key string) {
	switch key {
	case "left", "h":
		if ee.detailFocused > 0 {
			ee.detailFocused--
		}
	case "right", "l":
		if ee.detailFocused < detailBtnCount-1 {
			ee.detailFocused++
		}
	case "tab":
		// Cycle through buttons
		ee.detailFocused = (ee.detailFocused + 1) % detailBtnCount
	case "enter":
		ee.executeDetailButton(ee.detailFocused)
	case "e":
		ee.result = &EditorResult{Action: "edit-from-detail", Event: ee.detailEvent}
		ee.active = false
	case "D":
		ee.result = &EditorResult{Action: "delete-from-detail", Event: ee.detailEvent}
		ee.active = false
	case "c":
		ee.copyPlainText()
	case "J":
		ee.copyJSON()
	case "q", "esc":
		ee.result = &EditorResult{Action: "cancel"}
		ee.active = false
	}
}

func (ee *EventEditor) executeDetailButton(idx int) {
	switch idx {
	case detailBtnEdit:
		ee.result = &EditorResult{Action: "edit-from-detail", Event: ee.detailEvent}
		ee.active = false
	case detailBtnDelete:
		ee.result = &EditorResult{Action: "delete-from-detail", Event: ee.detailEvent}
		ee.active = false
	case detailBtnCopyPlain:
		ee.copyPlainText()
	case detailBtnCopyJSON:
		ee.copyJSON()
	}
}

func (ee *EventEditor) copyPlainText() {
	text := ee.formatPlainText()
	if err := util.CopyToClipboard(text); err != nil {
		ee.statusMsg = "Copy failed: " + err.Error()
	} else {
		ee.statusMsg = "Copied to clipboard"
	}
}

func (ee *EventEditor) copyJSON() {
	text := ee.formatJSON()
	if err := util.CopyToClipboard(text); err != nil {
		ee.statusMsg = "Copy failed: " + err.Error()
	} else {
		ee.statusMsg = "Copied as JSON to clipboard"
	}
}

func (ee *EventEditor) formatPlainText() string {
	ev := ee.detailEvent
	var sb strings.Builder

	sb.WriteString(ev.Summary)
	sb.WriteString("\n")

	sb.WriteString("Date: ")
	sb.WriteString(ev.Start.Format("2006-01-02"))
	sb.WriteString("\n")

	if ev.AllDay {
		sb.WriteString("Time: All Day\n")
	} else {
		timeFmt := "15:04"
		if !ee.detailUse24h {
			timeFmt = "3:04 PM"
		}
		sb.WriteString("Time: ")
		sb.WriteString(ev.Start.Format(timeFmt))
		sb.WriteString(" - ")
		sb.WriteString(ev.End.Format(timeFmt))
		sb.WriteString("\n")
	}

	if ev.Location != "" {
		sb.WriteString("Location: ")
		sb.WriteString(ev.Location)
		sb.WriteString("\n")
	}

	if ee.detailCalName != "" {
		sb.WriteString("Calendar: ")
		sb.WriteString(ee.detailCalName)
		sb.WriteString("\n")
	}

	if ev.Status != "" {
		sb.WriteString("Status: ")
		sb.WriteString(ev.Status)
		sb.WriteString("\n")
	}

	if ev.Recurring {
		recDesc := ev.RecurrenceDescription()
		if recDesc == "" {
			recDesc = "Yes"
		}
		sb.WriteString("Recurring: ")
		sb.WriteString(recDesc)
		sb.WriteString("\n")
	}

	if ev.Description != "" {
		sb.WriteString("\n")
		sb.WriteString(ev.Description)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (ee *EventEditor) formatJSON() string {
	ev := ee.detailEvent
	timeFmt := "15:04"
	if !ee.detailUse24h {
		timeFmt = "3:04 PM"
	}

	data := map[string]interface{}{
		"summary":     ev.Summary,
		"date":        ev.Start.Format("2006-01-02"),
		"start":       ev.Start.Format(timeFmt),
		"end":         ev.End.Format(timeFmt),
		"location":    ev.Location,
		"calendar":    ee.detailCalName,
		"status":      ev.Status,
		"allDay":      ev.AllDay,
		"recurring":   ev.Recurring,
		"description": ev.Description,
	}

	// Add recurrence description if recurring
	if ev.Recurring {
		recDesc := ev.RecurrenceDescription()
		if recDesc != "" {
			data["recurrenceRule"] = recDesc
		}
	}

	b, _ := json.MarshalIndent(data, "", "  ")
	return string(b)
}

func (ee *EventEditor) buildEventFromForm() *ical.Event {
	fields := ee.form.Fields()

	title := fields[0].Value
	allDay := fields[1].Selected == 1 // 0=No, 1=Yes
	dateStr := fields[2].Value
	startStr := fields[3].Value
	endStr := fields[4].Value
	location := fields[5].Value
	calIdx := fields[6].Selected
	notes := fields[7].Value

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	var startTime, endTime time.Time
	if allDay {
		startTime = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		endTime = startTime.AddDate(0, 0, 1)
	} else {
		start, err := time.Parse("15:04", startStr)
		if err != nil {
			start = time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)
		}

		end, err := time.Parse("15:04", endStr)
		if err != nil {
			end = time.Date(0, 1, 1, 10, 0, 0, 0, time.Local)
		}

		startTime = time.Date(date.Year(), date.Month(), date.Day(),
			start.Hour(), start.Minute(), 0, 0, time.Local)
		endTime = time.Date(date.Year(), date.Month(), date.Day(),
			end.Hour(), end.Minute(), 0, 0, time.Local)
	}

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
		AllDay:      allDay,
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
	case ModeDetail:
		content = ee.renderDetail()
	}

	return lipgloss.Place(
		ee.width,
		ee.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (ee *EventEditor) renderDetail() string {
	ev := ee.detailEvent
	formWidth := 60

	titleStyle := lipgloss.NewStyle().
		Foreground(ee.theme.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(formWidth)

	labelStyle := lipgloss.NewStyle().
		Foreground(ee.theme.TextMuted).
		Width(14).
		Align(lipgloss.Right).
		MarginRight(1)

	valueStyle := lipgloss.NewStyle().
		Foreground(ee.theme.Text)

	var lines []string
	lines = append(lines, "")

	// Title (multi-line if long)
	titleLines := util.WrapText(ev.Summary, formWidth)
	for _, tl := range titleLines {
		lines = append(lines, titleStyle.Render(tl))
	}
	lines = append(lines, "")

	// Helper function to render field rows
	renderField := func(label, value string) {
		row := labelStyle.Render(label+":") + valueStyle.Render(value)
		lines = append(lines, row)
		lines = append(lines, "")
	}

	// Date
	renderField("Date", ev.Start.Format("Monday, January 2, 2006"))

	// Time
	if ev.AllDay {
		renderField("Time", "All Day")
	} else {
		timeFmt := "15:04"
		if !ee.detailUse24h {
			timeFmt = "3:04 PM"
		}
		renderField("Time", ev.Start.Format(timeFmt)+" - "+ev.End.Format(timeFmt))
	}

	// Location (multi-line if long)
	if ev.Location != "" {
		textWidth := formWidth - 14 - 1
		locLines := util.WrapText(ev.Location, textWidth)
		for i, ll := range locLines {
			var row string
			if i == 0 {
				row = labelStyle.Render("Location:") + valueStyle.Render(ll)
			} else {
				row = labelStyle.Render("") + valueStyle.Render(ll)
			}
			lines = append(lines, row)
		}
		lines = append(lines, "")
	}

	// Calendar
	if ee.detailCalName != "" {
		renderField("Calendar", ee.detailCalName)
	}

	// Status
	if ev.Status != "" {
		renderField("Status", ev.Status)
	}

	// Recurring
	if ev.Recurring {
		// Debug: Always show RecurrenceRule if present
		if ev.RecurrenceRule != "" {
			recDesc := ev.RecurrenceDescription()
			if recDesc == "" {
				recDesc = "Yes (RRULE: " + ev.RecurrenceRule + ")"
			}
			renderField("Recurring", recDesc)
		} else {
			// No RRULE stored - event needs to be reloaded/synced
			renderField("Recurring", "Yes (no interval data)")
		}
	}

	// Description (multi-line support with wrapping)
	if ev.Description != "" {
		// Available width for text: formWidth - labelWidth - marginRight
		textWidth := formWidth - 14 - 1
		descLines := util.WrapText(ev.Description, textWidth)
		for i, dl := range descLines {
			var row string
			if i == 0 {
				row = labelStyle.Render("Notes:") + valueStyle.Render(dl)
			} else {
				row = labelStyle.Render("") + valueStyle.Render(dl)
			}
			lines = append(lines, row)
		}
		lines = append(lines, "")
	}

	// Buttons - match form style
	type btn struct {
		label  string
		danger bool
	}
	buttons := []btn{
		{"Edit", false},
		{"Delete", true},
		{"Copy", false},
		{"Copy JSON", false},
	}

	var btnStrs []string
	for i, b := range buttons {
		style := lipgloss.NewStyle().Padding(0, 2).Bold(true)

		if i == ee.detailFocused {
			if b.danger {
				style = style.
					Background(ee.theme.Error).
					Foreground(lipgloss.Color("#FFFFFF"))
			} else {
				style = style.
					Background(ee.theme.Accent).
					Foreground(lipgloss.Color("#FFFFFF"))
			}
		} else {
			style = style.
				Background(ee.theme.SurfaceAlt).
				Foreground(ee.theme.Text)
		}

		btnStrs = append(btnStrs, style.Render(b.label))
	}

	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(ee.theme.TextMuted)

	buttonRow := lipgloss.JoinHorizontal(lipgloss.Left,
		strings.Join(btnStrs, "  "),
		"  ",
		cancelStyle.Render("Close (Esc)"),
	)

	// Center the button row
	buttonRowStyle := lipgloss.NewStyle().
		Width(formWidth).
		Align(lipgloss.Center)

	lines = append(lines, buttonRowStyle.Render(buttonRow))

	// Status message (clipboard feedback)
	if ee.statusMsg != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(ee.theme.Accent).
			Italic(true).
			Align(lipgloss.Center).
			Width(formWidth)
		lines = append(lines, "")
		lines = append(lines, msgStyle.Render(ee.statusMsg))
	}

	lines = append(lines, "")

	content := strings.Join(lines, "\n")

	modalStyle := lipgloss.NewStyle().
		Width(formWidth + 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ee.theme.Accent).
		Padding(1, 2)

	return modalStyle.Render(content)
}
