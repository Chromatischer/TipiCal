package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
)

// buttonHitZone holds the x range of a button within the rendered modal.
type buttonHitZone struct {
	x1, x2 int
}

// Modal renders a centered modal overlay.
type Modal struct {
	theme   *config.Theme
	title   string
	content string
	width   int
	actions []ModalAction
	focused int

	// Cached layout info from last View() call.
	renderedWidth  int
	renderedHeight int
	buttonRowY     int
	buttonZones    []buttonHitZone
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

	rendered := modalStyle.Render(content)
	m.renderedWidth = lipgloss.Width(rendered)
	m.renderedHeight = lipgloss.Height(rendered)

	// Buttons are the 3rd row from the bottom: border(1) + padding(1) from bottom.
	m.buttonRowY = m.renderedHeight - 3

	// Compute button x positions within the rendered modal.
	// From modal left edge: border(1) + padding left(2) = 3 cols offset.
	const innerLeft = 3
	m.buttonZones = make([]buttonHitZone, len(m.actions))
	pos := innerLeft
	for i, action := range m.actions {
		w := 2 + lipgloss.Width(action.Label) + 2 // Padding(0, 2)
		m.buttonZones[i] = buttonHitZone{x1: pos, x2: pos + w - 1}
		pos += w + 2 // +2 for the "  " separator
	}

	return rendered
}

// HandleMouse processes a left-click at (x, y) relative to the container's top-left
// corner. containerWidth and containerHeight are the dimensions of the space in which
// the modal is centered. Returns the index of the clicked button (0+), or -2 for no hit.
func (m *Modal) HandleMouse(x, y, containerWidth, containerHeight int) int {
	if m.renderedWidth == 0 || m.renderedHeight == 0 {
		return -2
	}

	// Modal is centered in the container by lipgloss.Place.
	modalLeft := (containerWidth - m.renderedWidth) / 2
	modalTop := (containerHeight - m.renderedHeight) / 2

	// Convert to modal-local coordinates.
	mx := x - modalLeft
	my := y - modalTop

	if mx < 0 || mx >= m.renderedWidth || my < 0 || my >= m.renderedHeight {
		return -2
	}

	if my != m.buttonRowY {
		return -2
	}

	for i, zone := range m.buttonZones {
		if mx >= zone.x1 && mx <= zone.x2 {
			m.focused = i
			return i
		}
	}
	return -2
}
