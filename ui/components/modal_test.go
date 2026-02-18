package components

import (
	"strings"
	"testing"

	"github.com/tipical/tipical/config"
)

func TestModalHandleKeyNavigation(t *testing.T) {
	theme := &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextMuted:  "#A6ADC8",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}

	actions := []ModalAction{
		{Label: "Cancel", Key: "esc"},
		{Label: "Delete", Key: "d", Danger: true},
	}
	modal := NewModal(theme, "Test Modal", "Test content", actions)

	if modal.focused != 0 {
		t.Errorf("initial focused = %d, want 0", modal.focused)
	}

	result := modal.HandleKey("right")
	if result != -2 {
		t.Errorf("HandleKey(right) = %d, want -2 (no action)", result)
	}
	if modal.focused != 1 {
		t.Errorf("after right, focused = %d, want 1", modal.focused)
	}

	result = modal.HandleKey("right")
	if modal.focused != 1 {
		t.Errorf("focused should stay at 1 (already at last action), got %d", modal.focused)
	}

	result = modal.HandleKey("left")
	if modal.focused != 0 {
		t.Errorf("after left, focused = %d, want 0", modal.focused)
	}

	result = modal.HandleKey("left")
	if modal.focused != 0 {
		t.Errorf("focused should stay at 0 (already at first action), got %d", modal.focused)
	}
}

func TestModalHandleKeyHAndL(t *testing.T) {
	theme := &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextMuted:  "#A6ADC8",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}

	actions := []ModalAction{
		{Label: "Cancel", Key: "esc"},
		{Label: "Delete", Key: "d", Danger: true},
	}
	modal := NewModal(theme, "Test", "Test", actions)

	modal.HandleKey("l")
	if modal.focused != 1 {
		t.Errorf("after 'l', focused = %d, want 1", modal.focused)
	}

	modal.HandleKey("h")
	if modal.focused != 0 {
		t.Errorf("after 'h', focused = %d, want 0", modal.focused)
	}
}

func TestModalHandleKeyEnter(t *testing.T) {
	theme := &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextMuted:  "#A6ADC8",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}

	actions := []ModalAction{
		{Label: "Cancel", Key: "esc"},
		{Label: "Delete", Key: "d", Danger: true},
	}
	modal := NewModal(theme, "Test", "Test", actions)

	result := modal.HandleKey("enter")
	if result != 0 {
		t.Errorf("HandleKey(enter) with focused=0 = %d, want 0", result)
	}

	modal.HandleKey("right")
	result = modal.HandleKey("enter")
	if result != 1 {
		t.Errorf("HandleKey(enter) with focused=1 = %d, want 1", result)
	}
}

func TestModalHandleKeyEscape(t *testing.T) {
	theme := &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextMuted:  "#A6ADC8",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}

	actions := []ModalAction{
		{Label: "Cancel", Key: "esc"},
		{Label: "Delete", Key: "d", Danger: true},
	}
	modal := NewModal(theme, "Test", "Test", actions)

	result := modal.HandleKey("esc")
	if result != -1 {
		t.Errorf("HandleKey(esc) = %d, want -1 (cancel)", result)
	}
}

func TestModalHandleKeyShortcut(t *testing.T) {
	theme := &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextMuted:  "#A6ADC8",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}

	actions := []ModalAction{
		{Label: "Cancel", Key: "c"},
		{Label: "Delete", Key: "d", Danger: true},
	}
	modal := NewModal(theme, "Test", "Test", actions)

	result := modal.HandleKey("d")
	if result != 1 {
		t.Errorf("HandleKey(d) shortcut = %d, want 1 (Delete action)", result)
	}

	result = modal.HandleKey("c")
	if result != 0 {
		t.Errorf("HandleKey(c) shortcut = %d, want 0 (Cancel action)", result)
	}
}

func TestModalSetWidth(t *testing.T) {
	theme := &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextMuted:  "#A6ADC8",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}

	actions := []ModalAction{
		{Label: "OK", Key: "enter"},
	}
	modal := NewModal(theme, "Test", "Content", actions)

	if modal.width != 50 {
		t.Errorf("default width = %d, want 50", modal.width)
	}

	modal.SetWidth(60)
	if modal.width != 60 {
		t.Errorf("after SetWidth(60), width = %d, want 60", modal.width)
	}
}

func TestModalView(t *testing.T) {
	theme := &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextMuted:  "#A6ADC8",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}

	actions := []ModalAction{
		{Label: "Cancel", Key: "esc"},
		{Label: "Delete", Key: "d", Danger: true},
	}
	modal := NewModal(theme, "Delete Event", "Are you sure?", actions)

	view := modal.View()

	if !strings.Contains(view, "Delete Event") {
		t.Error("Modal view should contain title 'Delete Event'")
	}
	if !strings.Contains(view, "Are you sure?") {
		t.Error("Modal view should contain content 'Are you sure?'")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("Modal view should contain action 'Cancel'")
	}
	if !strings.Contains(view, "Delete") {
		t.Error("Modal view should contain action 'Delete'")
	}
}

func TestModalMultipleActions(t *testing.T) {
	theme := &config.Theme{
		Accent:     "#7C3AED",
		Text:       "#CDD6F4",
		TextMuted:  "#A6ADC8",
		SurfaceAlt: "#45475A",
		Error:      "#F38BA8",
	}

	actions := []ModalAction{
		{Label: "Save", Key: "s"},
		{Label: "Don't Save", Key: "n"},
		{Label: "Cancel", Key: "esc"},
	}
	modal := NewModal(theme, "Save Changes", "Save before closing?", actions)

	result := modal.HandleKey("right")
	if modal.focused != 1 {
		t.Errorf("after right, focused = %d, want 1", modal.focused)
	}

	modal.HandleKey("right")
	if modal.focused != 2 {
		t.Errorf("after second right, focused = %d, want 2", modal.focused)
	}

	modal.HandleKey("right")
	if modal.focused != 2 {
		t.Errorf("focused should stay at 2 (last action), got %d", modal.focused)
	}

	result = modal.HandleKey("s")
	if result != 0 {
		t.Errorf("HandleKey(s) shortcut = %d, want 0 (Save action)", result)
	}

	result = modal.HandleKey("n")
	if result != 1 {
		t.Errorf("HandleKey(n) shortcut = %d, want 1 (Don't Save action)", result)
	}
}
