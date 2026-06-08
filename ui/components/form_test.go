package components

import (
	"strings"
	"testing"

	"github.com/tipical/tipical/config"
)

func testFormTheme() *config.Theme {
	return &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextFaint:  "#6C7086",
		TextMuted:  "#A6ADC8",
		Surface:    "#313244",
		SurfaceAlt: "#45475A",
	}
}

func TestFormNavigation(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: ""},
		{Label: "Location", Type: FieldText, Value: ""},
		{Label: "All Day", Type: FieldSelect, Options: []string{"No", "Yes"}},
	}
	form := NewForm(theme, "Test Form", fields)

	if form.focused != 0 {
		t.Errorf("initial focused = %d, want 0", form.focused)
	}

	form.HandleKey("tab")
	if form.focused != 1 {
		t.Errorf("after tab, focused = %d, want 1", form.focused)
	}

	form.HandleKey("tab")
	if form.focused != 2 {
		t.Errorf("after second tab, focused = %d, want 2", form.focused)
	}

	form.HandleKey("tab")
	if form.focused != 3 {
		t.Errorf("after third tab (to buttons), focused = %d, want 3", form.focused)
	}

	form.HandleKey("tab")
	if form.focused != 0 {
		t.Errorf("after fourth tab (wrap), focused = %d, want 0", form.focused)
	}
}

func TestFormShiftTabNavigation(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: ""},
		{Label: "Location", Type: FieldText, Value: ""},
	}
	form := NewForm(theme, "Test Form", fields)

	form.HandleKey("shift+tab")
	if form.focused != 2 {
		t.Errorf("after shift+tab from 0, focused = %d, want 2 (button row)", form.focused)
	}

	form.HandleKey("shift+tab")
	if form.focused != 1 {
		t.Errorf("after shift+tab, focused = %d, want 1", form.focused)
	}
}

func TestFormTextInput(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: ""},
	}
	form := NewForm(theme, "Test Form", fields)

	form.HandleKey("H")
	form.HandleKey("e")
	form.HandleKey("l")
	form.HandleKey("l")
	form.HandleKey("o")

	if fields[0].Value != "Hello" {
		t.Errorf("field value = %q, want %q", fields[0].Value, "Hello")
	}

	if form.cursor != 5 {
		t.Errorf("cursor = %d, want 5", form.cursor)
	}
}

func TestFormBackspace(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "Hello"},
	}
	form := NewForm(theme, "Test Form", fields)
	form.cursor = 5

	form.HandleKey("backspace")

	if fields[0].Value != "Hell" {
		t.Errorf("after backspace, value = %q, want %q", fields[0].Value, "Hell")
	}
	if form.cursor != 4 {
		t.Errorf("after backspace, cursor = %d, want 4", form.cursor)
	}

	form.HandleKey("backspace")
	form.HandleKey("backspace")
	form.HandleKey("backspace")
	form.HandleKey("backspace")

	if fields[0].Value != "" {
		t.Errorf("after all backspaces, value = %q, want empty", fields[0].Value)
	}
}

func TestFormLeftRightCursor(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "Hello"},
	}
	form := NewForm(theme, "Test Form", fields)
	form.cursor = 5

	form.HandleKey("left")
	if form.cursor != 4 {
		t.Errorf("after left, cursor = %d, want 4", form.cursor)
	}

	form.HandleKey("left")
	form.HandleKey("left")
	form.HandleKey("left")
	form.HandleKey("left")
	if form.cursor != 0 {
		t.Errorf("after 4 lefts, cursor = %d, want 0", form.cursor)
	}

	form.HandleKey("left")
	if form.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", form.cursor)
	}

	form.HandleKey("right")
	if form.cursor != 1 {
		t.Errorf("after right, cursor = %d, want 1", form.cursor)
	}
}

func TestFormSelectField(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "All Day", Type: FieldSelect, Options: []string{"No", "Yes"}, Selected: 0},
	}
	form := NewForm(theme, "Test Form", fields)

	if fields[0].Selected != 0 {
		t.Errorf("initial selected = %d, want 0", fields[0].Selected)
	}

	form.HandleKey("right")
	if fields[0].Selected != 1 {
		t.Errorf("after right on select, selected = %d, want 1", fields[0].Selected)
	}
	if fields[0].Value != "Yes" {
		t.Errorf("after right on select, value = %q, want %q", fields[0].Value, "Yes")
	}

	form.HandleKey("right")
	if fields[0].Selected != 1 {
		t.Errorf("selected should stay at 1 (last option), got %d", fields[0].Selected)
	}

	form.HandleKey("left")
	if fields[0].Selected != 0 {
		t.Errorf("after left on select, selected = %d, want 0", fields[0].Selected)
	}
}

func TestFormSelectFieldNoTextInput(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "All Day", Type: FieldSelect, Options: []string{"No", "Yes"}, Selected: 0},
	}
	form := NewForm(theme, "Test Form", fields)

	form.HandleKey("a")
	form.HandleKey("b")
	form.HandleKey("c")

	if fields[0].Value != "" && fields[0].Value != "No" {
		t.Errorf("select field should not accept text input, got %q", fields[0].Value)
	}
}

func TestFormSubmit(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "Test"},
	}
	form := NewForm(theme, "Test Form", fields)

	if form.IsSubmitted() {
		t.Error("form should not be submitted initially")
	}

	form.HandleKey("ctrl+s")
	if !form.IsSubmitted() {
		t.Error("form should be submitted after ctrl+s")
	}
}

func TestFormSubmitFromButtonRow(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "Test"},
	}
	form := NewForm(theme, "Test Form", fields)

	form.focused = 1

	form.HandleKey("enter")
	if !form.IsSubmitted() {
		t.Error("form should be submitted after enter on button row")
	}
}

func TestFormCancel(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "Test"},
	}
	form := NewForm(theme, "Test Form", fields)

	if form.IsCancelled() {
		t.Error("form should not be cancelled initially")
	}

	form.HandleKey("esc")
	if !form.IsCancelled() {
		t.Error("form should be cancelled after esc")
	}
}

func TestFormReset(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "Test"},
	}
	form := NewForm(theme, "Test Form", fields)

	form.HandleKey("esc")
	form.HandleKey("tab")
	form.HandleKey("tab")

	form.Reset()

	if form.IsSubmitted() {
		t.Error("form should not be submitted after reset")
	}
	if form.IsCancelled() {
		t.Error("form should not be cancelled after reset")
	}
	if form.focused != 0 {
		t.Errorf("focused = %d after reset, want 0", form.focused)
	}
	if form.cursor != 0 {
		t.Errorf("cursor = %d after reset, want 0", form.cursor)
	}
}

func TestFormTextAreaEnter(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Notes", Type: FieldTextArea, Value: "Line1"},
	}
	form := NewForm(theme, "Test Form", fields)
	form.cursor = 5

	form.HandleKey("enter")
	form.HandleKey("L")
	form.HandleKey("i")
	form.HandleKey("n")
	form.HandleKey("e")
	form.HandleKey("2")

	want := "Line1\nLine2"
	if fields[0].Value != want {
		t.Errorf("textarea value = %q, want %q", fields[0].Value, want)
	}
}

func TestFormTextAreaNavigation(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Notes", Type: FieldTextArea, Value: "Line1\nLine2\nLine3"},
	}
	form := NewForm(theme, "Test Form", fields)

	form.cursor = 0

	form.HandleKey("down")
	if form.cursor != 6 {
		t.Errorf("after down from line 0, cursor = %d, want 6", form.cursor)
	}

	form.HandleKey("up")
	if form.cursor != 0 {
		t.Errorf("after up from line 1, cursor = %d, want 0", form.cursor)
	}
}

func TestFormDownOnNonTextArea(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: ""},
		{Label: "Location", Type: FieldText, Value: ""},
	}
	form := NewForm(theme, "Test Form", fields)

	form.HandleKey("down")
	if form.focused != 1 {
		t.Errorf("after down on text field, focused = %d, want 1", form.focused)
	}
}

func TestFormUpOnNonTextArea(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: ""},
		{Label: "Location", Type: FieldText, Value: ""},
	}
	form := NewForm(theme, "Test Form", fields)
	form.focused = 1

	form.HandleKey("up")
	if form.focused != 0 {
		t.Errorf("after up on text field, focused = %d, want 0", form.focused)
	}
}

func TestFormFieldsAccessor(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "Test"},
		{Label: "Location", Type: FieldText, Value: "Room"},
	}
	form := NewForm(theme, "Test Form", fields)

	gotFields := form.Fields()
	if len(gotFields) != 2 {
		t.Errorf("Fields() returned %d fields, want 2", len(gotFields))
	}

	field := form.Field(0)
	if field == nil || field.Label != "Title" {
		t.Error("Field(0) should return Title field")
	}

	field = form.Field(-1)
	if field != nil {
		t.Error("Field(-1) should return nil")
	}

	field = form.Field(100)
	if field != nil {
		t.Error("Field(100) should return nil")
	}
}

func TestFormSetSize(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: ""},
	}
	form := NewForm(theme, "Test Form", fields)

	form.SetSize(80, 24)

	if form.width != 80 {
		t.Errorf("width = %d, want 80", form.width)
	}
	if form.height != 24 {
		t.Errorf("height = %d, want 24", form.height)
	}
}

func TestFormView(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "My Event"},
		{Label: "All Day", Type: FieldSelect, Options: []string{"No", "Yes"}, Selected: 0},
	}
	form := NewForm(theme, "New Event", fields)

	view := form.View()

	if !strings.Contains(view, "New Event") {
		t.Error("Form view should contain title 'New Event'")
	}
	if !strings.Contains(view, "Title:") {
		t.Error("Form view should contain field label 'Title:'")
	}
	if !strings.Contains(view, "All Day:") {
		t.Error("Form view should contain field label 'All Day:'")
	}
	if !strings.Contains(view, "Save") {
		t.Error("Form view should contain 'Save' button")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("Form view should contain 'Cancel' button")
	}
}

func TestFormCheckboxOptionsClickable(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "All Day", Type: FieldCheckbox, Selected: 0},
	}
	form := NewForm(theme, "New Event", fields)
	form.SetSize(50, 20)
	form.View()

	rowY := form.fieldBounds[0].top

	allDayX := 2 + formLabelWidth + 9
	form.HandleMouse(allDayX, rowY, form.renderedWidth, form.renderedHeight)
	if fields[0].Selected != 1 {
		t.Fatalf("clicking All Day option should select 1, got %d", fields[0].Selected)
	}

	timedX := 2 + formLabelWidth + 1
	form.HandleMouse(timedX, rowY, form.renderedWidth, form.renderedHeight)
	if fields[0].Selected != 0 {
		t.Fatalf("clicking Timed option should select 0, got %d", fields[0].Selected)
	}
}

func TestFormCheckboxOptionsClickableAtMinimumWidth(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "All Day", Type: FieldCheckbox, Selected: 0},
	}
	form := NewForm(theme, "New Event", fields)
	form.SetSize(34, 20)
	form.View()

	rowY := form.fieldBounds[0].top
	inputWidth := form.inputWidth()
	pad := checkboxSegmentPadding(inputWidth)
	offWidth := len("Timed") + 2*pad
	onWidth := len("All Day") + 2*pad

	allDayX := 2 + formLabelWidth + 1 + offWidth + 1
	form.HandleMouse(allDayX, rowY, form.renderedWidth, form.renderedHeight)
	if fields[0].Selected != 1 {
		t.Fatalf("minimum-width All Day click should select 1, got %d", fields[0].Selected)
	}

	timedX := 2 + formLabelWidth + 1
	form.HandleMouse(timedX, rowY, form.renderedWidth, form.renderedHeight)
	if fields[0].Selected != 0 {
		t.Fatalf("minimum-width Timed click should select 0, got %d", fields[0].Selected)
	}

	if offWidth+1+onWidth > inputWidth {
		t.Fatalf("toggle width %d should fit inside input width %d", offWidth+1+onWidth, inputWidth)
	}
}

func TestFormEnterOnTextFieldMovesToNext(t *testing.T) {
	theme := testFormTheme()
	fields := []FormField{
		{Label: "Title", Type: FieldText, Value: "Test"},
		{Label: "Location", Type: FieldText, Value: ""},
	}
	form := NewForm(theme, "Test Form", fields)

	form.HandleKey("enter")
	if form.focused != 1 {
		t.Errorf("after enter on text field, focused = %d, want 1", form.focused)
	}
}

func TestFormTextAreaLineInfo(t *testing.T) {
	tests := []struct {
		value      string
		cursor     int
		wantLine   int
		wantCol    int
		wantStarts int
	}{
		{"Hello", 3, 0, 3, 1},
		{"Line1\nLine2", 0, 0, 0, 2},
		{"Line1\nLine2", 5, 0, 5, 2},
		{"Line1\nLine2", 6, 1, 0, 2},
		{"Line1\nLine2", 8, 1, 2, 2},
		{"A\nB\nC", 0, 0, 0, 3},
		{"A\nB\nC", 2, 1, 0, 3},
		{"A\nB\nC", 4, 2, 0, 3},
	}

	for _, tt := range tests {
		line, col, starts := textAreaLineInfo(tt.value, tt.cursor)
		if line != tt.wantLine || col != tt.wantCol || len(starts) != tt.wantStarts {
			t.Errorf("textAreaLineInfo(%q, %d) = (line=%d, col=%d, starts=%d), want (line=%d, col=%d, starts=%d)",
				tt.value, tt.cursor, line, col, len(starts), tt.wantLine, tt.wantCol, tt.wantStarts)
		}
	}
}
