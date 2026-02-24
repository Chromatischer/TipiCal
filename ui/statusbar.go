package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders the bottom bar with keybinding hints and status.
type StatusBar struct {
	styles    *Styles
	width     int
	syncState string // "", "syncing", "synced", "error"
	lastSync  time.Time
	message   string // temporary status message
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
	// Key hints
	hints := []struct {
		key  string
		desc string
	}{
		{" ", "nav (h/l)"},
		{" ", "move (j/k)"},
		{"󰃶 ", "today (t)"},
		{" ", "new (n)"},
		{" ", "edit (e)"},
		{" ", "delete (D)"},
		{" ", "search (/)"},
		{"󰋖", "help (?)"},
		{" ", "sidebar(b)"},
		{"q", "quit"},
	}

	var hintParts []string
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
	}
	hintLines := packParts(hintParts, sb.width-2, "  ")

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

// packParts fits parts onto lines of lineWidth columns,
// separating them with sep. Parts are never split across lines.
func packParts(parts []string, lineWidth int, sep string) []string {
	sepWidth := lipgloss.Width(sep)
	var lines []string
	currentLine := ""
	currentWidth := 0

	for _, part := range parts {
		partWidth := lipgloss.Width(part)
		if currentWidth == 0 {
			currentLine = part
			currentWidth = partWidth
		} else if currentWidth+sepWidth+partWidth <= lineWidth {
			currentLine += sep + part
			currentWidth += sepWidth + partWidth
		} else {
			lines = append(lines, currentLine)
			currentLine = part
			currentWidth = partWidth
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}
