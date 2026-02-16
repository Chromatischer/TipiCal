# TipiCal - Agent Notes

## Build & Test

```
go build .
go test ./...
```

## Styling Gotchas

### Lipgloss Background Colors
Background colors do **not** cascade to child elements. Each rendered text element must have its own background set:

```go
// WRONG - parent background won't show through
container := lipgloss.NewStyle().Background(bg).Render(
    text.Render("hello") // no background here
)

// RIGHT - set background on inner elements
text := lipgloss.NewStyle().Background(bg).Render("hello")
container := lipgloss.NewStyle().Background(bg).Render(text)
```

### Text Width & Emojis
See CLAUDE.md - user text may contain emojis (2 terminal columns wide). Always use `util.TruncateText` for width measurement.

## Debugging Visual Issues

When the user reports a visual/rendering issue in a specific component, **read that component first**. Don't tour the entire codebase.

## Key Files

- `ui/app.go` - Root Bubble Tea model, routing, main layout
- `ui/styles.go` - All lipgloss style definitions
- `config/theme.go` - Color palette (dark/light modes)
- `ui/views/` - Calendar view implementations
- `ui/components/` - Reusable UI components (MiniCalendar, Modal, etc.)
- `ui/editor/` - Event editor modal
