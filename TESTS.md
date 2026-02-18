# Test Writing Guide

## Completed Tests

All core tests are implemented:

- `util/timeutil_test.go` — WeekStart, MonthGrid, SameDay, FormatTime/Range, TruncateText, HoursBetween
- `util/text_test.go` — WrapText
- `config/config_test.go` — DefaultConfig, CalendarColor
- `ical/event_test.go` — OverlapsWith, IsMultiDay, AllDaySpanDays, RecurrenceDescription
- `ical/store_test.go` — EventsForDay, EventsInRange, AddEvent/RemoveEvent/FindEvent
- `ical/recurrence_test.go` — ExpandRecurrence
- `caldav/sync_test.go` — parseCalendarObject, eventToIcal round-trip

---

## Additional Tests to Consider

### UI Components (`ui/components/`)

**MiniCalendar**
- Navigation: prev/next month, clicking date selects it
- Highlighting: today, selected day, days with events
- Week start preference (Monday vs Sunday)

**Modal**
- Open/close behavior
- Focus management
- Escape key handling

### Editor (`ui/editor/`)

**Event Editor**
- Creating new event: all fields save correctly
- Editing existing event: changes persist
- Validation: required fields, time constraints (End after Start)
- All-day toggle: hides time inputs, adjusts end date
- Recurrence rule: UI updates when rule changes
- Cancel: discards changes without saving

### Views (`ui/views/`)

**DayView / WeekView**
- Event rendering: position, height based on time
- Multi-day events span correctly
- All-day events render in header area
- Click to select, keyboard navigation
- Scrolling behavior

**MonthView**
- Grid rendering: 6 rows, correct dates
- Week start preference
- Event indicators on days
- Navigation between months

### Sync (`caldav/`)

**Sync**
- `SyncAll`: handles client failures gracefully, preserves cache when offline
- `LoadFromCache`: populates store from cached data
- `CreateEvent` / `UpdateEvent` / `DeleteEvent`: round-trip to server
- Cache invalidation on successful operations

**Client**
- Authentication handling
- Calendar discovery
- Event fetching with date range

---

## Gotchas

**All-day end is exclusive.** A single-day all-day event on Jan 15 has `Start=Jan 15` and `End=Jan 16`.

**rrule.Between is inclusive at both ends.** `rule.Between(start, end, true)` includes both endpoints.

**Use time.UTC for test times** unless specifically testing timezone behavior.

**parseCalendarObject and eventToIcal are unexported** — test file must be in `package caldav`.

---

## Running Tests

```
go test ./... -v
go test ./ical/ ./caldav/ -v
```
