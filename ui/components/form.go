package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
)

// FormFieldType represents the type of form field.
type FormFieldType int

const (
	FieldText FormFieldType = iota
	FieldTextArea
	FieldSelect
)

// FormField represents a single form field.
type FormField struct {
	Label       string
	Value       string
	Placeholder string
	Type        FormFieldType
	Options     []string // for select fields
	Selected    int      // for select fields
}

// fieldBound holds the inclusive row range of a field within the rendered form.
type fieldBound struct {
	top, bottom int // form-local rows, 0 = top border row
}

// Form is a reusable inline form component.
type Form struct {
	theme     *config.Theme
	title     string
	fields    []FormField
	focused   int
	cursor    int // cursor position within the current text field
	width     int
	height    int
	submitted bool
	cancelled bool

	// Cached layout info from last View() call.
	renderedWidth  int
	renderedHeight int
	fieldBounds    []fieldBound
	buttonRowY     int
}

// NewForm creates a new form.
func NewForm(theme *config.Theme, title string, fields []FormField) *Form {
	return &Form{
		theme:  theme,
		title:  title,
		fields: fields,
		width:  50,
	}
}

// SetSize sets the form dimensions.
func (f *Form) SetSize(w, h int) {
	f.width = w
	f.height = h
}

// Fields returns the form fields.
func (f *Form) Fields() []FormField {
	return f.fields
}

// Field returns a specific field.
func (f *Form) Field(index int) *FormField {
	if index >= 0 && index < len(f.fields) {
		return &f.fields[index]
	}
	return nil
}

// IsSubmitted returns true if the form was submitted.
func (f *Form) IsSubmitted() bool {
	return f.submitted
}

// IsCancelled returns true if the form was cancelled.
func (f *Form) IsCancelled() bool {
	return f.cancelled
}

// Reset resets the form state.
func (f *Form) Reset() {
	f.submitted = false
	f.cancelled = false
	f.focused = 0
	f.cursor = 0
}

// textAreaLineInfo returns the current line index, the column within that line,
// and the start offsets (in runes) of each line for the given value and cursor.
func textAreaLineInfo(value string, cursor int) (line, col int, lineStarts []int) {
	runes := []rune(value)
	lineStarts = []int{0}
	for i, r := range runes {
		if r == '\n' {
			lineStarts = append(lineStarts, i+1)
		}
	}
	for i := len(lineStarts) - 1; i >= 0; i-- {
		if cursor >= lineStarts[i] {
			line = i
			col = cursor - lineStarts[i]
			break
		}
	}
	return
}

// textAreaLineEnd returns the length (in runes) of the given line.
func textAreaLineEnd(value string, lineStarts []int, line int) int {
	runes := []rune(value)
	start := lineStarts[line]
	if line+1 < len(lineStarts) {
		return lineStarts[line+1] - 1 - start // exclude the \n
	}
	return len(runes) - start
}

// HandleKey processes a key press.
func (f *Form) HandleKey(key string) {
	isTextArea := f.focused < len(f.fields) && f.fields[f.focused].Type == FieldTextArea

	switch key {
	case "tab":
		f.focusNext()

	case "shift+tab":
		f.focusPrev()

	case "down":
		if isTextArea {
			field := &f.fields[f.focused]
			line, col, lineStarts := textAreaLineInfo(field.Value, f.cursor)
			if line+1 < len(lineStarts) {
				// Move to next line
				nextLineLen := textAreaLineEnd(field.Value, lineStarts, line+1)
				if col > nextLineLen {
					col = nextLineLen
				}
				f.cursor = lineStarts[line+1] + col
			} else {
				// At last line, move to next field
				f.focusNext()
			}
		} else {
			f.focusNext()
		}

	case "up":
		if isTextArea {
			field := &f.fields[f.focused]
			line, col, lineStarts := textAreaLineInfo(field.Value, f.cursor)
			if line > 0 {
				// Move to previous line
				prevLineLen := textAreaLineEnd(field.Value, lineStarts, line-1)
				if col > prevLineLen {
					col = prevLineLen
				}
				f.cursor = lineStarts[line-1] + col
			} else {
				// At first line, move to previous field
				f.focusPrev()
			}
		} else {
			f.focusPrev()
		}

	case "ctrl+s":
		f.submitted = true

	case "enter":
		if f.focused >= len(f.fields) {
			// On button row
			f.submitted = true
			return
		}
		if isTextArea {
			// Insert newline
			field := &f.fields[f.focused]
			runes := []rune(field.Value)
			newRunes := make([]rune, 0, len(runes)+1)
			newRunes = append(newRunes, runes[:f.cursor]...)
			newRunes = append(newRunes, '\n')
			newRunes = append(newRunes, runes[f.cursor:]...)
			field.Value = string(newRunes)
			f.cursor++
		} else {
			// Move to next field on enter
			f.focusNext()
		}

	case "esc":
		f.cancelled = true

	case "backspace":
		if f.focused < len(f.fields) {
			field := &f.fields[f.focused]
			if field.Type == FieldSelect {
				return
			}
			if f.cursor > 0 {
				runes := []rune(field.Value)
				field.Value = string(runes[:f.cursor-1]) + string(runes[f.cursor:])
				f.cursor--
			}
		}

	case "left":
		if f.focused < len(f.fields) {
			field := &f.fields[f.focused]
			if field.Type == FieldSelect {
				if field.Selected > 0 {
					field.Selected--
					field.Value = field.Options[field.Selected]
				}
				return
			}
			if f.cursor > 0 {
				f.cursor--
			}
		}

	case "right":
		if f.focused < len(f.fields) {
			field := &f.fields[f.focused]
			if field.Type == FieldSelect {
				if field.Selected < len(field.Options)-1 {
					field.Selected++
					field.Value = field.Options[field.Selected]
				}
				return
			}
			if f.cursor < len([]rune(field.Value)) {
				f.cursor++
			}
		}

	default:
		// Type character
		if len(key) == 1 || (len(key) > 0 && !strings.HasPrefix(key, "ctrl+")) {
			if f.focused < len(f.fields) {
				field := &f.fields[f.focused]
				if field.Type == FieldSelect {
					return
				}
				runes := []rune(field.Value)
				newRunes := make([]rune, 0, len(runes)+len(key))
				newRunes = append(newRunes, runes[:f.cursor]...)
				newRunes = append(newRunes, []rune(key)...)
				newRunes = append(newRunes, runes[f.cursor:]...)
				field.Value = string(newRunes)
				f.cursor += len([]rune(key))
			}
		}
	}
}

func (f *Form) focusNext() {
	f.focused++
	if f.focused >= len(f.fields)+1 { // +1 for buttons
		f.focused = 0
	}
	if f.focused < len(f.fields) {
		f.cursor = len([]rune(f.fields[f.focused].Value))
	}
}

func (f *Form) focusPrev() {
	f.focused--
	if f.focused < 0 {
		f.focused = len(f.fields)
	}
	if f.focused < len(f.fields) {
		f.cursor = len([]rune(f.fields[f.focused].Value))
	}
}

func (f *Form) safeFieldIndex() int {
	if f.focused >= len(f.fields) {
		return len(f.fields) - 1
	}
	if f.focused < 0 {
		return 0
	}
	return f.focused
}

// View renders the form.
func (f *Form) View() string {
	formWidth := f.width - 4
	if formWidth < 30 {
		formWidth = 30
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(f.theme.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(formWidth)

	labelStyle := lipgloss.NewStyle().
		Foreground(f.theme.TextMuted).
		Width(12).
		Align(lipgloss.Right).
		MarginRight(1)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render(f.title))
	lines = append(lines, "")

	inputWidth := formWidth - 16
	if inputWidth < 10 {
		inputWidth = 10
	}

	// labelWidth = label Width(12) + MarginRight(1)
	const labelWidth = 13
	underlineStyle := lipgloss.NewStyle().Foreground(f.theme.Accent)
	underlineLine := strings.Repeat(" ", labelWidth) + underlineStyle.Render(strings.Repeat("─", inputWidth))

	f.fieldBounds = make([]fieldBound, len(f.fields))
	for i, field := range f.fields {
		isFocused := i == f.focused

		// Record the top row of this field. Content lines start at form row 1
		// (row 0 is the border top), so the next line is at index len(lines)+1.
		topRow := len(lines) + 1

		switch field.Type {
		case FieldSelect:
			label := labelStyle.Render(field.Label + ":")
			input := f.renderSelect(field, inputWidth, isFocused)
			lines = append(lines, label+input)
			f.fieldBounds[i] = fieldBound{top: topRow, bottom: topRow}
			if isFocused {
				lines = append(lines, underlineLine)
			}
			lines = append(lines, "")
		case FieldTextArea:
			rendered := f.renderTextArea(field, inputWidth, isFocused, labelStyle)
			lines = append(lines, rendered...)
			f.fieldBounds[i] = fieldBound{top: topRow, bottom: topRow + len(rendered) - 1}
			if isFocused {
				lines = append(lines, underlineLine)
			}
			lines = append(lines, "")
		default:
			label := labelStyle.Render(field.Label + ":")
			input := f.renderTextInput(field, inputWidth, isFocused)
			lines = append(lines, label+input)
			f.fieldBounds[i] = fieldBound{top: topRow, bottom: topRow}
			if isFocused {
				lines = append(lines, underlineLine)
			}
			lines = append(lines, "")
		}
	}

	// Buttons
	lines = append(lines, "")
	// Record button row: the next line added will be at this form-local row.
	f.buttonRowY = len(lines) + 1

	saveStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true)
	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2)

	if f.focused >= len(f.fields) {
		saveStyle = saveStyle.
			Background(f.theme.Accent).
			Foreground(lipgloss.Color("#FFFFFF"))
		cancelStyle = cancelStyle.
			Foreground(f.theme.TextMuted)
	} else {
		saveStyle = saveStyle.
			Background(f.theme.SurfaceAlt).
			Foreground(f.theme.Text)
		cancelStyle = cancelStyle.
			Foreground(f.theme.TextMuted)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		"  ",
		saveStyle.Render("Save"),
		"  ",
		cancelStyle.Render("Cancel (Esc)"),
	)
	lines = append(lines, buttons)
	lines = append(lines, "")

	content := strings.Join(lines, "\n")

	modalStyle := lipgloss.NewStyle().
		Width(formWidth+4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(f.theme.Accent).
		Padding(0, 1)

	rendered := modalStyle.Render(content)
	f.renderedWidth = lipgloss.Width(rendered)
	f.renderedHeight = lipgloss.Height(rendered)
	return rendered
}

// HandleMouse processes a left-click at (x, y) relative to the container's top-left.
// containerWidth and containerHeight are the dimensions of the centered overlay space.
func (f *Form) HandleMouse(x, y, containerWidth, containerHeight int) {
	if f.renderedWidth == 0 || f.renderedHeight == 0 {
		return
	}

	// Form is centered in container by lipgloss.Place.
	formLeft := (containerWidth - f.renderedWidth) / 2
	formTop := (containerHeight - f.renderedHeight) / 2

	// Check bounds.
	if x < formLeft || x >= formLeft+f.renderedWidth || y < formTop || y >= formTop+f.renderedHeight {
		return
	}

	// Convert to form-local coordinates (0 = left/top border row/col).
	localX := x - formLeft
	localY := y - formTop

	// Check button row.
	// Button layout from form left: border(1) + padding(1) = 2 offset, then
	// JoinHorizontal("  ", Save(8), "  ", Cancel(16)):
	// "  " at [2,3], Save at [4,11], "  " at [12,13], Cancel at [14,29].
	if localY == f.buttonRowY {
		if localX >= 4 && localX <= 11 {
			f.submitted = true
		} else if localX >= 14 && localX <= 29 {
			f.cancelled = true
		} else {
			// Generic click on button area — focus button row.
			f.focused = len(f.fields)
		}
		return
	}

	// Check field rows.
	for i, bounds := range f.fieldBounds {
		if localY >= bounds.top && localY <= bounds.bottom {
			f.focused = i
			f.cursor = len([]rune(f.fields[i].Value))
			return
		}
	}
}

func (f *Form) renderTextInput(field FormField, width int, focused bool) string {
	value := field.Value
	if value == "" && !focused {
		value = field.Placeholder
		return lipgloss.NewStyle().
			Foreground(f.theme.TextFaint).
			Width(width).
			Padding(0, 1).
			Background(f.theme.SurfaceAlt).
			Render(value)
	}

	style := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1)

	if focused {
		style = style.
			Background(f.theme.Surface).
			Foreground(f.theme.Text)

		// Show cursor
		runes := []rune(value)
		if f.cursor >= len(runes) {
			value = value + "█"
		} else {
			value = string(runes[:f.cursor]) + "█" + string(runes[f.cursor:])
		}
	} else {
		style = style.
			Background(f.theme.SurfaceAlt).
			Foreground(f.theme.Text)
	}

	return style.Render(value)
}

func (f *Form) renderTextArea(field FormField, width int, focused bool, labelStyle lipgloss.Style) []string {
	value := field.Value
	valueLines := strings.Split(value, "\n")

	// Determine which line the cursor is on and the column within it
	cursorLine, cursorCol := 0, 0
	if focused {
		cursorLine, cursorCol, _ = textAreaLineInfo(value, f.cursor)
	}

	var result []string
	for i, vl := range valueLines {
		// Label only on first line
		var label string
		if i == 0 {
			label = labelStyle.Render(field.Label + ":")
		} else {
			label = labelStyle.Render("")
		}

		displayLine := vl

		style := lipgloss.NewStyle().
			Width(width).
			Padding(0, 1)

		if focused {
			style = style.
				Background(f.theme.Surface).
				Foreground(f.theme.Text)

			// Show cursor on the active line
			if i == cursorLine {
				runes := []rune(displayLine)
				if cursorCol >= len(runes) {
					displayLine = displayLine + "█"
				} else {
					displayLine = string(runes[:cursorCol]) + "█" + string(runes[cursorCol:])
				}
			}
		} else {
			style = style.
				Background(f.theme.SurfaceAlt).
				Foreground(f.theme.Text)
		}

		if !focused && value == "" && i == 0 {
			style = style.Foreground(f.theme.TextFaint)
			displayLine = field.Placeholder
		}

		result = append(result, label+style.Render(displayLine))
	}

	return result
}

func (f *Form) renderSelect(field FormField, width int, focused bool) string {
	if len(field.Options) == 0 {
		return ""
	}

	value := field.Options[field.Selected]
	indicator := fmt.Sprintf("◀ %s ▶", value)

	style := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1)

	if focused {
		style = style.
			Background(f.theme.Surface).
			Foreground(f.theme.Accent).
			Bold(true)
	} else {
		style = style.
			Background(f.theme.SurfaceAlt).
			Foreground(f.theme.Text)
	}

	return style.Render(indicator)
}
