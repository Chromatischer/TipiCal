package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
)

// Modal renders a centered modal overlay.
type Modal struct {
	theme   *config.Theme
	title   string
	content string
	width   int
	actions []ModalAction
	focused int
}

// ModalAction represents a button in the modal.
type ModalAction struct {
	Label  string
	Key    string
	Danger bool
}

// NewModal creates a new modal.
func NewModal(theme *config.Theme, title, content string, actions []ModalAction) *Modal {
	return &Modal{
		theme:   theme,
		title:   title,
		content: content,
		width:   50,
		actions: actions,
	}
}

// SetWidth sets the modal width.
func (m *Modal) SetWidth(w int) {
	m.width = w
}

// HandleKey processes a key press.
func (m *Modal) HandleKey(key string) int {
	switch key {
	case "left", "h":
		if m.focused > 0 {
			m.focused--
		}
	case "right", "l":
		if m.focused < len(m.actions)-1 {
			m.focused++
		}
	case "enter":
		return m.focused
	case "esc":
		return -1
	}

	// Check for action key shortcuts
	for i, a := range m.actions {
		if key == a.Key {
			return i
		}
	}

	return -2 // no action
}

// View renders the modal.
func (m *Modal) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width - 6)

	contentStyle := lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Width(m.width-6).
		Padding(1, 0)

	var lines []string
	lines = append(lines, titleStyle.Render(m.title))
	lines = append(lines, contentStyle.Render(m.content))

	// Action buttons
	var buttons []string
	for i, action := range m.actions {
		style := lipgloss.NewStyle().Padding(0, 2)

		if i == m.focused {
			if action.Danger {
				style = style.
					Background(m.theme.Error).
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true)
			} else {
				style = style.
					Background(m.theme.Accent).
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true)
			}
		} else {
			style = style.
				Background(m.theme.SurfaceAlt).
				Foreground(m.theme.Text)
		}

		buttons = append(buttons, style.Render(action.Label))
	}

	lines = append(lines, strings.Join(buttons, "  "))

	content := strings.Join(lines, "\n")

	modalStyle := lipgloss.NewStyle().
		Width(m.width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
		Padding(1, 2)

	return modalStyle.Render(content)
}
