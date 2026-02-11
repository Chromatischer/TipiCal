package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-ical/terminal-ical/config"
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

// HandleKey processes a key press.
func (f *Form) HandleKey(key string) {
	switch key {
	case "tab", "down":
		f.focused++
		if f.focused >= len(f.fields)+1 { // +1 for buttons
			f.focused = 0
		}
		f.cursor = len(f.fields[f.safeFieldIndex()].Value)

	case "shift+tab", "up":
		f.focused--
		if f.focused < 0 {
			f.focused = len(f.fields)
		}
		if f.focused < len(f.fields) {
			f.cursor = len(f.fields[f.focused].Value)
		}

	case "ctrl+s", "enter":
		if f.focused >= len(f.fields) {
			// On button row
			f.submitted = true
			return
		}
		// Move to next field on enter
		f.focused++
		if f.focused >= len(f.fields)+1 {
			f.submitted = true
		}
		if f.focused < len(f.fields) {
			f.cursor = len(f.fields[f.focused].Value)
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
			if f.cursor < len(field.Value) {
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

	for i, field := range f.fields {
		label := labelStyle.Render(field.Label + ":")

		var input string
		isFocused := i == f.focused

		switch field.Type {
		case FieldSelect:
			input = f.renderSelect(field, inputWidth, isFocused)
		default:
			input = f.renderTextInput(field, inputWidth, isFocused)
		}

		lines = append(lines, label+input)
		lines = append(lines, "")
	}

	// Buttons
	lines = append(lines, "")
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

	return modalStyle.Render(content)
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
			Foreground(f.theme.Text).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(f.theme.Accent)

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
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(f.theme.Accent)
	} else {
		style = style.
			Background(f.theme.SurfaceAlt).
			Foreground(f.theme.Text)
	}

	return style.Render(indicator)
}
