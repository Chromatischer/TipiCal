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
		{"h/l", "nav"},
		{"j/k", "move"},
		{"t", "today"},
		{"n", "new"},
		{"e", "edit"},
		{"D", "delete"},
		{"/", "search"},
		{"?", "help"},
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
	hintsStr := strings.Join(hintParts, "  ")

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

	hintsWidth := lipgloss.Width(hintsStr)
	rightWidth := lipgloss.Width(rightStr)
	gap := sb.width - hintsWidth - rightWidth - 2
	if gap < 0 {
		gap = 0
	}

	bar := lipgloss.JoinHorizontal(
		lipgloss.Center,
		hintsStr,
		strings.Repeat(" ", gap),
		rightStr,
	)

	return sb.styles.StatusBar.Width(sb.width).Render(bar)
}
