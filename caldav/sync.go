package caldav

import (
	"context"
	"fmt"
	"strings"
	"time"

	goical "github.com/emersion/go-ical"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ical"
)

// Sync manages syncing between CalDAV servers and the local event store.
type Sync struct {
	clients []*Client
	cache   *Cache
	store   *ical.Store
	theme   *config.Theme
	cfg     *config.Config
}

// NewSync creates a new sync manager.
func NewSync(clients []*Client, cache *Cache, store *ical.Store, theme *config.Theme, cfg *config.Config) *Sync {
	return &Sync{
		clients: clients,
		cache:   cache,
		store:   store,
		theme:   theme,
		cfg:     cfg,
	}
}

// SyncAll syncs all configured calendars. It clears the store first
// to avoid stale events from previous syncs.
func (s *Sync) SyncAll(ctx context.Context) error {
	// Collect all new events, then replace the store atomically
	var allEvents []*ical.Event

	var syncErrors []error
	for _, client := range s.clients {
		events, err := s.syncClient(ctx, client)
		if err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("syncing calendar %d: %w", client.CalendarIndex(), err))
			continue
		}
		allEvents = append(allEvents, events...)
	}

	// Replace store contents atomically
	s.store.Events = allEvents

	// Return first error if any, but only after we've loaded what we can
	if len(syncErrors) > 0 {
		return syncErrors[0]
	}
	return nil
}

func (s *Sync) syncClient(ctx context.Context, client *Client) ([]*ical.Event, error) {
	// Discover calendars
	calendars, err := client.DiscoverCalendars(ctx)
	if err != nil {
		return nil, err
	}

	// Determine the color for this calendar
	calIdx := client.CalendarIndex()
	calColor := string(s.theme.CalendarColor(calIdx))
	if calIdx < len(s.cfg.Calendars) && s.cfg.Calendars[calIdx].Color != "" {
		calColor = s.cfg.Calendars[calIdx].Color
	}

	// Fetch events from each calendar
	now := time.Now()
	start := now.AddDate(0, -1, 0) // 1 month back
	end := now.AddDate(0, 3, 0)    // 3 months forward

	var events []*ical.Event
	for _, cal := range calendars {
		objects, err := client.FetchEvents(ctx, cal.Path, start, end)
		if err != nil {
			return nil, fmt.Errorf("fetching from %s: %w", cal.Path, err)
		}

		for _, obj := range objects {
			// Cache the raw data
			s.cache.Put(cal.Path, CacheEntry{
				Path:      obj.Path,
				ETag:      obj.ETag,
				FetchedAt: time.Now(),
			})

			// Parse events
			parsed, err := parseCalendarObject(obj.Data, calIdx)
			if err != nil {
				continue // Skip unparseable events
			}

			for _, event := range parsed {
				// Assign calendar color
				event.Color = calColor

				// Expand recurring events
				if event.Recurring {
					expanded := s.expandRecurring(event, obj.Data, start, end)
					events = append(events, expanded...)
				} else {
					events = append(events, event)
				}
			}
		}
	}

	return events, nil
}

// expandRecurring extracts the RRULE from the raw iCal data and expands it.
func (s *Sync) expandRecurring(event *ical.Event, cal *goical.Calendar, rangeStart, rangeEnd time.Time) []*ical.Event {
	// Find the RRULE string from the raw calendar data
	for _, comp := range cal.Children {
		if comp.Name != goical.CompEvent {
			continue
		}
		if prop := comp.Props.Get(goical.PropRecurrenceRule); prop != nil {
			return ical.ExpandRecurrence(event, prop.Value, rangeStart, rangeEnd)
		}
	}
	// No RRULE found, return as-is
	return []*ical.Event{event}
}

// parseCalendarObject converts a go-ical Calendar into our Event model.
func parseCalendarObject(cal *goical.Calendar, calendarIndex int) ([]*ical.Event, error) {
	var events []*ical.Event

	for _, comp := range cal.Children {
		if comp.Name != goical.CompEvent {
			continue
		}

		event := &ical.Event{
			CalendarID: calendarIndex,
		}

		// UID
		if prop := comp.Props.Get(goical.PropUID); prop != nil {
			event.UID = prop.Value
		}

		// Summary
		if prop := comp.Props.Get(goical.PropSummary); prop != nil {
			event.Summary = prop.Value
		}

		// Description
		if prop := comp.Props.Get(goical.PropDescription); prop != nil {
			event.Description = prop.Value
		}

		// Location
		if prop := comp.Props.Get(goical.PropLocation); prop != nil {
			event.Location = prop.Value
		}

		// Status
		if prop := comp.Props.Get(goical.PropStatus); prop != nil {
			event.Status = prop.Value
		}

		// Start time
		if prop := comp.Props.Get(goical.PropDateTimeStart); prop != nil {
			t, err := prop.DateTime(nil)
			if err == nil {
				event.Start = t
			}
			// Check if all-day
			if v := prop.Params.Get("VALUE"); v != "" && strings.EqualFold(v, "DATE") {
				event.AllDay = true
			}
		}

		// End time
		if prop := comp.Props.Get(goical.PropDateTimeEnd); prop != nil {
			t, err := prop.DateTime(nil)
			if err == nil {
				event.End = t
			}
		} else if event.AllDay {
			// All-day events without DTEND default to 1 day
			event.End = event.Start.AddDate(0, 0, 1)
		} else {
			// Default to 1 hour
			event.End = event.Start.Add(time.Hour)
		}

		// RRULE check
		if prop := comp.Props.Get(goical.PropRecurrenceRule); prop != nil {
			event.Recurring = true
		}

		if event.UID != "" && !event.Start.IsZero() {
			events = append(events, event)
		}
	}

	return events, nil
}
