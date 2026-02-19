# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added

- **Overlap rendering is now fully correct across all real-world event configurations.** Two fixes landed together:
  - Events with the same start time are now rendered consistently on every row: the longer event occupies the full-width content area and the shorter event collapses into the sidebar marker (▌) for the entire duration of the overlap — not just on the first row.
  - Selection outlines on overlapping same-start-time events now track the correct event across all continuation rows. Previously the outline would jump between the primary and secondary event on every sub-row after the first, making keyboard selection unreliable.

### Fixed

- `+1` badge now shows correctly for same-calendar events with the same start time, including when one event ends before the other. Badge shifts inward when the event is selected to avoid crowding the selection marker.
- Calendar names from CalDAV discovery are now correctly restored from cache on startup, including when calendars have no events.
- App no longer falls back to "Work"/"Personal" placeholder calendars when a background sync fails (e.g. offline) — the cached calendar list is preserved.
- `source_name` (account name) is now persisted in `cache_index.json` so the sidebar source headers survive a restart.
- Agenda view scrolling no longer jumps when long descriptions or locations wrap across multiple lines.
- Agenda hyperlinks are shortened and tinted blue for readability, without breaking click targets.

## 0.1.0-beta.0

- Initial public beta.
- CalDAV sync, multi-view calendar, and event editor UI.
