package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// statusHintZone is a clickable hint region within the status bar content row.
type statusHintZone struct {
	x1, x2 int
	row    int // row within the status bar content (0 = first line)
	action string
}

// StatusBar renders the bottom bar with keybinding hints and status.
type StatusBar struct {
	styles    *Styles
	width     int
	syncState string // "", "syncing", "synced", "error"
	lastSync  time.Time
	message   string // temporary status message
	hintZones []statusHintZone
}

// NewStatusBar creates a new status bar.
func NewStatusBar(styles *Styles) *StatusBar {
	return &StatusBar{
		styles:    styles,
		syncState: "",
	}
}

// SetWidth sets the status bar width.
func (sb *StatusBar) SetWidth(w int) {
	sb.width = w
}

// SetMessage sets a temporary status message.
func (sb *StatusBar) SetMessage(msg string) {
	sb.message = msg
}

// SetSyncState sets the sync status.
func (sb *StatusBar) SetSyncState(state string) {
	sb.syncState = state
	if state == "synced" {
		sb.lastSync = time.Now()
	}
}

// View renders the status bar.
func (sb *StatusBar) View() string {
	// Key hints with associated click actions
	type hintDef struct {
		key    string
		desc   string
		action string
	}
	hints := []hintDef{
		{" ", "nav (h/l)", ""},
		{" ", "move (j/k)", ""},
		{"󰃶 ", "today (t)", "today"},
		{" ", "new (n)", "new"},
		{" ", "edit (e)", "edit"},
		{" ", "delete (D)", "delete"},
		{" ", "search (/)", "search"},
		{"󰋖", "help (?)", "help"},
		{" ", "sidebar(b)", "sidebar"},
		{"q", "quit", "quit"},
	}

	var hintParts []string
	var hintActions []string
	keyStyle := lipgloss.NewStyle().
		Foreground(sb.styles.Theme.Accent).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(sb.styles.Theme.TextFaint)

	for _, h := range hints {
		hintParts = append(hintParts, fmt.Sprintf("%s %s",
			keyStyle.Render(h.key),
			descStyle.Render(h.desc),
		))
		hintActions = append(hintActions, h.action)
	}

	hintLines, hintPositions := packPartsWithPositions(hintParts, sb.width-2, "  ")

	// Record clickable zones (content starts at screen col 1 due to Padding(0,1))
	const leftPad = 1
	sb.hintZones = sb.hintZones[:0]
	for i, pos := range hintPositions {
		if hintActions[i] == "" {
			continue
		}
		partWidth := lipgloss.Width(hintParts[i])
		sb.hintZones = append(sb.hintZones, statusHintZone{
			x1:     leftPad + pos.x,
			x2:     leftPad + pos.x + partWidth - 1,
			row:    pos.row,
			action: hintActions[i],
		})
	}

	// Right side: sync status or message
	var rightStr string
	if sb.message != "" {
		rightStr = lipgloss.NewStyle().
			Foreground(sb.styles.Theme.TextMuted).
			Render(sb.message)
	} else {
		switch sb.syncState {
		case "syncing":
			rightStr = lipgloss.NewStyle().
				Foreground(sb.styles.Theme.Warning).
				Render("⟳ Syncing...")
		case "synced":
			rightStr = lipgloss.NewStyle().
				Foreground(sb.styles.Theme.Success).
				Render(fmt.Sprintf("● Synced %s", sb.lastSync.Format("15:04")))
		case "error":
			rightStr = lipgloss.NewStyle().
				Foreground(sb.styles.Theme.Error).
				Render("✗ Sync error")
		}
	}

	rightWidth := lipgloss.Width(rightStr)

	last := hintLines[len(hintLines)-1]
	if lipgloss.Width(last)+1+rightWidth > sb.width-2 {
		hintLines = append(hintLines, "")
	}

	var barLines []string
	for i, line := range hintLines {
		if i == len(hintLines)-1 {
			lineWidth := lipgloss.Width(line)
			gap := sb.width - lineWidth - rightWidth - 2
			if gap < 1 {
				gap = 1
			}
			barLines = append(barLines, line+strings.Repeat(" ", gap)+rightStr)
		} else {
			barLines = append(barLines, line)
		}
	}

	bar := strings.Join(barLines, "\n")

	return sb.styles.StatusBar.Width(sb.width).Render(bar)
}

// HitTest returns the action string for a click at terminal (x, y).
// termHeight is the total terminal height; the content row is at termHeight-1.
func (sb *StatusBar) HitTest(x, y, termHeight int) string {
	// Status bar border is at termHeight-2, content starts at termHeight-1.
	if y != termHeight-1 {
		return ""
	}
	for _, zone := range sb.hintZones {
		if zone.row == 0 && x >= zone.x1 && x <= zone.x2 {
			return zone.action
		}
	}
	return ""
}

// packParts fits parts onto lines of lineWidth columns,
// separating them with sep. Parts are never split across lines.
func packParts(parts []string, lineWidth int, sep string) []string {
	lines, _ := packPartsWithPositions(parts, lineWidth, sep)
	return lines
}

// packPartsWithPositions is like packParts but also returns the (x, row) position
// of each part's start within the packed layout.
func packPartsWithPositions(parts []string, lineWidth int, sep string) ([]string, []struct{ x, row int }) {
	sepWidth := lipgloss.Width(sep)
	var lines []string
	var positions []struct{ x, row int }
	currentLine := ""
	currentWidth := 0
	currentRow := 0

	for _, part := range parts {
		partWidth := lipgloss.Width(part)
		if currentWidth == 0 {
			positions = append(positions, struct{ x, row int }{0, currentRow})
			currentLine = part
			currentWidth = partWidth
		} else if currentWidth+sepWidth+partWidth <= lineWidth {
			positions = append(positions, struct{ x, row int }{currentWidth + sepWidth, currentRow})
			currentLine += sep + part
			currentWidth += sepWidth + partWidth
		} else {
			lines = append(lines, currentLine)
			currentRow++
			positions = append(positions, struct{ x, row int }{0, currentRow})
			currentLine = part
			currentWidth = partWidth
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines, positions
}
