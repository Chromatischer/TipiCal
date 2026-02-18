package editor

import (
	"strings"
	"testing"
	"time"

	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
)

func testEditorTheme() *config.Theme {
	return &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextFaint:  "#6C7086",
		TextMuted:  "#A6ADC8",
		Today:      "#F38BA8",
		Selected:   "#89B4FA",
		Surface:    "#313244",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}
}

func testStore() *ical.Store {
	store := ical.NewStore()
	store.RegisterCalendar(ical.CalendarInfo{
		ID:     0,
		Name:   "Work",
		Color:  "#3B82F6",
		Source: "Work Account",
	})
	store.RegisterCalendar(ical.CalendarInfo{
		ID:     1,
		Name:   "Personal",
		Color:  "#10B981",
		Source: "Personal Account",
	})
	return store
}

func TestEventEditorOpenCreate(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})

	if !editor.IsActive() {
		t.Error("editor should be active after OpenCreate")
	}
	if editor.mode != ModeCreate {
		t.Errorf("mode = %v, want ModeCreate", editor.mode)
	}
	if editor.form == nil {
		t.Error("form should be created")
	}

	fields := editor.form.Fields()
	if len(fields) == 0 || fields[0].Label != "Title" {
		t.Error("form should have Title field as first field")
	}
}

func TestEventEditorOpenCreateDefaultTimes(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})

	fields := editor.form.Fields()

	dateField := fields[2]
	if dateField.Value != "2026-02-18" {
		t.Errorf("date field = %q, want %q", dateField.Value, "2026-02-18")
	}

	startField := fields[3]
	if startField.Value != "09:00" {
		t.Errorf("start time = %q, want %q", startField.Value, "09:00")
	}

	endField := fields[4]
	if endField.Value != "10:00" {
		t.Errorf("end time = %q, want %q", endField.Value, "10:00")
	}
}

func TestEventEditorOpenEdit(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:         "test-uid-123",
		Summary:     "Test Event",
		Description: "Test Description",
		Location:    "Test Location",
		Start:       time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:         time.Date(2026, time.February, 18, 15, 30, 0, 0, time.UTC),
		AllDay:      false,
		CalendarID:  0,
	}

	editor.OpenEdit(event, []string{"Work", "Personal"})

	if !editor.IsActive() {
		t.Error("editor should be active after OpenEdit")
	}
	if editor.mode != ModeEdit {
		t.Errorf("mode = %v, want ModeEdit", editor.mode)
	}
	if editor.editingUID != "test-uid-123" {
		t.Errorf("editingUID = %q, want %q", editor.editingUID, "test-uid-123")
	}

	fields := editor.form.Fields()
	if fields[0].Value != "Test Event" {
		t.Errorf("title = %q, want %q", fields[0].Value, "Test Event")
	}
	if fields[5].Value != "Test Location" {
		t.Errorf("location = %q, want %q", fields[5].Value, "Test Location")
	}
}

func TestEventEditorOpenEditAllDay(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:        "allday-uid",
		Summary:    "All Day Event",
		Start:      time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 19, 0, 0, 0, 0, time.UTC),
		AllDay:     true,
		CalendarID: 0,
	}

	editor.OpenEdit(event, []string{"Work", "Personal"})

	fields := editor.form.Fields()
	allDayField := fields[1]
	if allDayField.Selected != 1 {
		t.Errorf("all-day selected = %d, want 1 (Yes)", allDayField.Selected)
	}
}

func TestEventEditorOpenDelete(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:     "delete-uid",
		Summary: "Event to Delete",
	}

	editor.OpenDelete(event)

	if !editor.IsActive() {
		t.Error("editor should be active after OpenDelete")
	}
	if editor.mode != ModeDelete {
		t.Errorf("mode = %v, want ModeDelete", editor.mode)
	}
	if editor.deleteModal == nil {
		t.Error("delete modal should be created")
	}
}

func TestEventEditorOpenDetail(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:        "detail-uid",
		Summary:    "Event Details",
		Location:   "Conference Room",
		Start:      time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		AllDay:     false,
		CalendarID: 0,
	}

	editor.OpenDetail(event, "Work", true)

	if !editor.IsActive() {
		t.Error("editor should be active after OpenDetail")
	}
	if editor.mode != ModeDetail {
		t.Errorf("mode = %v, want ModeDetail", editor.mode)
	}
	if editor.detailEvent != event {
		t.Error("detailEvent should be set")
	}
}

func TestEventEditorHandleKeyCreateSubmit(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})

	editor.HandleKey("T")
	editor.HandleKey("e")
	editor.HandleKey("s")
	editor.HandleKey("t")

	for i := 0; i < 6; i++ {
		editor.HandleKey("tab")
	}
	editor.HandleKey("ctrl+s")

	if !editor.form.IsSubmitted() {
		t.Error("form should be submitted")
	}

	result := editor.Result()
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Action != "create" {
		t.Errorf("action = %q, want %q", result.Action, "create")
	}
	if result.Event == nil {
		t.Fatal("event should not be nil")
	}
	if result.Event.Summary != "Test" {
		t.Errorf("summary = %q, want %q", result.Event.Summary, "Test")
	}
}

func TestEventEditorHandleKeyCreateCancel(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})

	editor.HandleKey("esc")

	result := editor.Result()
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Action != "cancel" {
		t.Errorf("action = %q, want %q", result.Action, "cancel")
	}
}

func TestEventEditorHandleKeyDeleteConfirm(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	store.AddEvent(&ical.Event{
		UID:     "delete-test-uid",
		Summary: "Event to Delete",
	})
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:     "delete-test-uid",
		Summary: "Event to Delete",
	}

	editor.OpenDelete(event)

	editor.HandleKey("right")
	editor.HandleKey("enter")

	result := editor.Result()
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Action != "delete" {
		t.Errorf("action = %q, want %q", result.Action, "delete")
	}
}

func TestEventEditorHandleKeyDeleteCancel(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:     "delete-cancel-uid",
		Summary: "Event",
	}

	editor.OpenDelete(event)

	editor.HandleKey("esc")

	result := editor.Result()
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Action != "cancel" {
		t.Errorf("action = %q, want %q", result.Action, "cancel")
	}
}

func TestEventEditorHandleKeyDetailNavigation(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:        "detail-nav-uid",
		Summary:    "Event",
		Start:      time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		CalendarID: 0,
	}

	editor.OpenDetail(event, "Work", true)

	if editor.detailFocused != 0 {
		t.Errorf("initial detailFocused = %d, want 0", editor.detailFocused)
	}

	editor.HandleKey("right")
	if editor.detailFocused != 1 {
		t.Errorf("after right, detailFocused = %d, want 1", editor.detailFocused)
	}

	editor.HandleKey("left")
	if editor.detailFocused != 0 {
		t.Errorf("after left, detailFocused = %d, want 0", editor.detailFocused)
	}
}

func TestEventEditorHandleKeyDetailEdit(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:        "detail-edit-uid",
		Summary:    "Event",
		Start:      time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		CalendarID: 0,
	}

	editor.OpenDetail(event, "Work", true)

	editor.HandleKey("e")

	result := editor.Result()
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Action != "edit-from-detail" {
		t.Errorf("action = %q, want %q", result.Action, "edit-from-detail")
	}
}

func TestEventEditorHandleKeyDetailDelete(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:        "detail-delete-uid",
		Summary:    "Event",
		Start:      time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		CalendarID: 0,
	}

	editor.OpenDetail(event, "Work", true)

	editor.HandleKey("D")

	result := editor.Result()
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Action != "delete-from-detail" {
		t.Errorf("action = %q, want %q", result.Action, "delete-from-detail")
	}
}

func TestEventEditorHandleKeyDetailCancel(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:        "detail-cancel-uid",
		Summary:    "Event",
		Start:      time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		CalendarID: 0,
	}

	editor.OpenDetail(event, "Work", true)

	editor.HandleKey("esc")

	result := editor.Result()
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Action != "cancel" {
		t.Errorf("action = %q, want %q", result.Action, "cancel")
	}
}

func TestEventEditorBuildEventFromFormTimed(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})

	fields := editor.form.Fields()
	fields[0].Value = "Meeting"
	fields[1].Selected = 0
	fields[2].Value = "2026-02-18"
	fields[3].Value = "14:00"
	fields[4].Value = "15:30"
	fields[5].Value = "Room 101"
	fields[6].Selected = 0
	fields[7].Value = "Discuss project"

	event := editor.buildEventFromForm()

	if event.Summary != "Meeting" {
		t.Errorf("summary = %q, want %q", event.Summary, "Meeting")
	}
	if event.AllDay {
		t.Error("AllDay should be false")
	}
	if event.Start.Hour() != 14 || event.Start.Minute() != 0 {
		t.Errorf("start time = %v, want 14:00", event.Start)
	}
	if event.End.Hour() != 15 || event.End.Minute() != 30 {
		t.Errorf("end time = %v, want 15:30", event.End)
	}
	if event.Location != "Room 101" {
		t.Errorf("location = %q, want %q", event.Location, "Room 101")
	}
	if event.Description != "Discuss project" {
		t.Errorf("description = %q, want %q", event.Description, "Discuss project")
	}
	if event.CalendarID != 0 {
		t.Errorf("calendarID = %d, want 0", event.CalendarID)
	}
}

func TestEventEditorBuildEventFromFormAllDay(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})

	fields := editor.form.Fields()
	fields[0].Value = "Holiday"
	fields[1].Selected = 1
	fields[2].Value = "2026-02-18"

	event := editor.buildEventFromForm()

	if !event.AllDay {
		t.Error("AllDay should be true")
	}
	if event.Start.Hour() != 0 || event.Start.Minute() != 0 {
		t.Errorf("all-day start should be at midnight, got %v", event.Start)
	}
	wantEnd := event.Start.AddDate(0, 0, 1)
	if !event.End.Equal(wantEnd) {
		t.Errorf("all-day end = %v, want %v", event.End, wantEnd)
	}
}

func TestEventEditorSetSize(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})

	editor.SetSize(80, 24)

	if editor.width != 80 {
		t.Errorf("width = %d, want 80", editor.width)
	}
	if editor.height != 24 {
		t.Errorf("height = %d, want 24", editor.height)
	}
}

func TestEventEditorView(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	if editor.View() != "" {
		t.Error("inactive editor View() should return empty string")
	}

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})
	editor.SetSize(80, 24)

	view := editor.View()
	if view == "" {
		t.Error("active editor View() should not return empty string")
	}
	if !strings.Contains(view, "New Event") {
		t.Error("create view should contain 'New Event' title")
	}
}

func TestEventEditorViewDelete(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{UID: "test", Summary: "Test"}
	editor.OpenDelete(event)
	editor.SetSize(80, 24)

	view := editor.View()
	if !strings.Contains(view, "Delete Event") {
		t.Error("delete view should contain 'Delete Event' title")
	}
}

func TestEventEditorViewDetail(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:        "test",
		Summary:    "Test Event",
		Location:   "Room 101",
		Start:      time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		CalendarID: 0,
	}
	editor.OpenDetail(event, "Work", true)
	editor.SetSize(80, 24)

	view := editor.View()
	if !strings.Contains(view, "Test Event") {
		t.Error("detail view should contain event summary")
	}
	if !strings.Contains(view, "Room 101") {
		t.Error("detail view should contain location")
	}
}

func TestEventEditorResultClear(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})
	editor.HandleKey("esc")

	if editor.Result() == nil {
		t.Fatal("result should not be nil after cancel")
	}

	editor.ClearResult()

	if editor.Result() != nil {
		t.Error("result should be nil after ClearResult")
	}
}

func TestEventEditorFormatPlainText(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		Summary:     "Test Event",
		Location:    "Room 101",
		Description: "Event description",
		Start:       time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:         time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		Status:      "CONFIRMED",
	}

	editor.OpenDetail(event, "Work", true)
	text := editor.formatPlainText()

	if !strings.Contains(text, "Test Event") {
		t.Error("plain text should contain summary")
	}
	if !strings.Contains(text, "Room 101") {
		t.Error("plain text should contain location")
	}
	if !strings.Contains(text, "Event description") {
		t.Error("plain text should contain description")
	}
	if !strings.Contains(text, "Work") {
		t.Error("plain text should contain calendar name")
	}
}

func TestEventEditorFormatJSON(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		Summary:     "Test Event",
		Location:    "Room 101",
		Description: "Event description",
		Start:       time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:         time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		Status:      "CONFIRMED",
	}

	editor.OpenDetail(event, "Work", true)
	json := editor.formatJSON()

	if !strings.Contains(json, `"summary": "Test Event"`) {
		t.Error("JSON should contain summary field")
	}
	if !strings.Contains(json, `"location": "Room 101"`) {
		t.Error("JSON should contain location field")
	}
	if !strings.Contains(json, `"calendar": "Work"`) {
		t.Error("JSON should contain calendar field")
	}
}

func TestEventEditorStatusMessage(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	editor.statusMsg = "Test message"
	if editor.StatusMessage() != "Test message" {
		t.Errorf("StatusMessage() = %q, want %q", editor.StatusMessage(), "Test message")
	}

	editor.ClearStatusMessage()
	if editor.StatusMessage() != "" {
		t.Errorf("StatusMessage() after clear = %q, want empty", editor.StatusMessage())
	}
}

func TestEventEditorEditExistingPreservesUID(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	originalUID := "original-uid-12345"
	event := &ical.Event{
		UID:        originalUID,
		Summary:    "Original Event",
		Start:      time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		CalendarID: 0,
	}

	editor.OpenEdit(event, []string{"Work", "Personal"})

	fields := editor.form.Fields()
	fields[0].Value = "Updated Event"

	for i := 0; i < 6; i++ {
		editor.HandleKey("tab")
	}
	editor.HandleKey("ctrl+s")

	result := editor.Result()
	if result.Event.UID != originalUID {
		t.Errorf("updated event UID = %q, want %q", result.Event.UID, originalUID)
	}
}

func TestEventEditorCreateGeneratesUID(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	date := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	editor.OpenCreate(date, []string{"Work", "Personal"})

	fields := editor.form.Fields()
	fields[0].Value = "New Event"

	for i := 0; i < 6; i++ {
		editor.HandleKey("tab")
	}
	editor.HandleKey("ctrl+s")

	result := editor.Result()
	if result.Event.UID == "" {
		t.Error("new event should have generated UID")
	}
	if !strings.Contains(result.Event.UID, "tical-") {
		t.Errorf("UID should start with 'tical-', got %q", result.Event.UID)
	}
}

func TestEventEditorHandleKeyInactive(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	editor.HandleKey("a")
	editor.HandleKey("b")
	editor.HandleKey("enter")

	if editor.Result() != nil {
		t.Error("inactive editor should not produce result from key events")
	}
}

func TestEventEditorDetailTabNavigation(t *testing.T) {
	theme := testEditorTheme()
	store := testStore()
	editor := NewEventEditor(theme, store)

	event := &ical.Event{
		UID:        "tab-nav-uid",
		Summary:    "Event",
		Start:      time.Date(2026, time.February, 18, 14, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.February, 18, 15, 0, 0, 0, time.UTC),
		CalendarID: 0,
	}

	editor.OpenDetail(event, "Work", true)

	editor.HandleKey("tab")
	if editor.detailFocused != 1 {
		t.Errorf("after tab, detailFocused = %d, want 1", editor.detailFocused)
	}

	for i := 0; i < 10; i++ {
		editor.HandleKey("tab")
	}
	if editor.detailFocused >= detailBtnCount {
		t.Errorf("detailFocused should wrap around, got %d", editor.detailFocused)
	}
}
