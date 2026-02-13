# terminal-ical

## Gotchas

- Any user-controllable text (event summaries, locations, calendar names, etc.) may contain emojis which can mess with displayed-text lengths in unexpected ways. Emojis are 1 rune but 2 terminal columns wide. Always use `go-runewidth` (via `util.TruncateText`) for width measurement and truncation — never count runes or bytes to determine display width.
- For any visual/rendering issue, ask for a screenshot before attempting a fix. Don't try to reason about TUI layout from code alone.
