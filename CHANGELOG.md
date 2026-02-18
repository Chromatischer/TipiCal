# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Fixed

- `+1` badge now shows correctly for same-calendar events with the same start time, including when one event ends before the other. Badge shifts inward when the event is selected to avoid crowding the selection marker.
- Calendar names from CalDAV discovery are now correctly restored from cache on startup, including when calendars have no events.
- App no longer falls back to "Work"/"Personal" placeholder calendars when a background sync fails (e.g. offline) — the cached calendar list is preserved.
- `source_name` (account name) is now persisted in `cache_index.json` so the sidebar source headers survive a restart.

## 0.1.0-beta.0

- Initial public beta.
- CalDAV sync, multi-view calendar, and event editor UI.
