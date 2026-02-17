package caldav

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	goical "github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
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
	// Clear calendar registry before rebuilding
	s.store.ClearCalendars()

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

	// Get the config calendar name for grouping
	calIdx := client.CalendarIndex()
	configName := ""
	configColor := ""
	if calIdx < len(s.cfg.Calendars) {
		configName = s.cfg.Calendars[calIdx].Name
		configColor = s.cfg.Calendars[calIdx].Color
	}

	// Fetch events from each calendar
	now := time.Now()
	start := now.AddDate(0, -1, 0) // 1 month back
	end := now.AddDate(0, 3, 0)    // 3 months forward

	var events []*ical.Event
	for _, cal := range calendars {
		// Register each discovered sub-calendar with a unique ID
		calColor := configColor
		if calColor == "" {
			calColor = string(s.theme.CalendarColor(len(s.store.Calendars)))
		}

		calName := cal.Name
		if calName == "" {
			calName = cal.Path
		}

		readOnly := false
		if calIdx < len(s.cfg.Calendars) {
			readOnly = s.cfg.Calendars[calIdx].ReadOnly
		}
		calID := s.store.RegisterCalendar(ical.CalendarInfo{
			Name:        calName,
			Color:       calColor,
			Source:      configName,
			CalPath:     cal.Path,
			ConfigIndex: calIdx,
			ReadOnly:    readOnly,
		})

		objects, err := client.FetchEvents(ctx, cal.Path, start, end)
		if err != nil {
			return nil, fmt.Errorf("fetching from %s: %w", cal.Path, err)
		}

		// Store calendar metadata in cache
		s.cache.SetIndexMeta(cal.Path, calName, calColor, calIdx)

		for _, obj := range objects {
			// Serialize iCal data for caching
			var buf bytes.Buffer
			if err := goical.NewEncoder(&buf).Encode(obj.Data); err == nil {
				s.cache.Put(cal.Path, CacheEntry{
					Path:      obj.Path,
					ETag:      obj.ETag,
					Data:      buf.String(),
					FetchedAt: time.Now(),
				})
			}

			// Parse events
			parsed, err := parseCalendarObject(obj.Data, calID)
			if err != nil {
				continue // Skip unparseable events
			}

			for _, event := range parsed {
				// Assign calendar color and CalDAV paths
				event.Color = calColor
				event.CalPath = cal.Path
				event.ResourcePath = obj.Path

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

// LoadFromCache populates the store from cached iCal data for instant startup.
func (s *Sync) LoadFromCache() {
	indexes := s.cache.AllIndexes()
	if len(indexes) == 0 {
		return
	}

	now := time.Now()
	start := now.AddDate(0, -1, 0)
	end := now.AddDate(0, 3, 0)

	var allEvents []*ical.Event

	for _, idx := range indexes {
		if len(idx.Entries) == 0 {
			continue
		}

		readOnly := false
		if idx.ConfigIndex < len(s.cfg.Calendars) {
			readOnly = s.cfg.Calendars[idx.ConfigIndex].ReadOnly
		}
		calID := s.store.RegisterCalendar(ical.CalendarInfo{
			Name:        idx.CalendarName,
			Color:       idx.CalendarColor,
			CalPath:     idx.CalendarURL,
			ConfigIndex: idx.ConfigIndex,
			ReadOnly:    readOnly,
		})

		for _, entry := range idx.Entries {
			if entry.Data == "" {
				continue
			}

			dec := goical.NewDecoder(strings.NewReader(entry.Data))
			cal, err := dec.Decode()
			if err != nil {
				continue
			}

			parsed, err := parseCalendarObject(cal, calID)
			if err != nil {
				continue
			}

			for _, event := range parsed {
				event.Color = idx.CalendarColor
				event.CalPath = idx.CalendarURL
				event.ResourcePath = entry.Path

				if event.Recurring {
					expanded := s.expandRecurring(event, cal, start, end)
					allEvents = append(allEvents, expanded...)
				} else {
					allEvents = append(allEvents, event)
				}
			}
		}
	}

	s.store.Events = allEvents
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
			if text, err := prop.Text(); err == nil {
				event.Summary = text
			} else {
				event.Summary = prop.Value
			}
		}

		// Description
		if prop := comp.Props.Get(goical.PropDescription); prop != nil {
			if text, err := prop.Text(); err == nil {
				event.Description = text
			} else {
				event.Description = prop.Value
			}
		}

		// Location
		if prop := comp.Props.Get(goical.PropLocation); prop != nil {
			if text, err := prop.Text(); err == nil {
				event.Location = text
			} else {
				event.Location = prop.Value
			}
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
			event.RecurrenceRule = prop.Value
		}

		if event.UID != "" && !event.Start.IsZero() {
			events = append(events, event)
		}
	}

	return events, nil
}

// clientForCalendar returns the Client responsible for the given CalendarID.
func (s *Sync) clientForCalendar(calendarID int) *Client {
	if calendarID < 0 || calendarID >= len(s.store.Calendars) {
		return nil
	}
	configIdx := s.store.Calendars[calendarID].ConfigIndex
	for _, c := range s.clients {
		if c.CalendarIndex() == configIdx {
			return c
		}
	}
	return nil
}

// eventToIcal converts an ical.Event to a go-ical Calendar object ready for PUT.
func eventToIcal(event *ical.Event) *goical.Calendar {
	cal := goical.NewCalendar()
	cal.Props.SetText(goical.PropVersion, "2.0")
	cal.Props.SetText(goical.PropProductID, "-//TipiCal//TipiCal//EN")

	comp := goical.NewComponent(goical.CompEvent)
	comp.Props.SetText(goical.PropUID, event.UID)
	comp.Props.SetText(goical.PropSummary, event.Summary)

	if event.Description != "" {
		comp.Props.SetText(goical.PropDescription, event.Description)
	}
	if event.Location != "" {
		comp.Props.SetText(goical.PropLocation, event.Location)
	}
	if event.Status != "" {
		comp.Props.SetText(goical.PropStatus, event.Status)
	}

	now := time.Now().UTC()
	dtstamp := goical.NewProp(goical.PropDateTimeStamp)
	dtstamp.SetDateTime(now)
	comp.Props.Set(dtstamp)

	if event.AllDay {
		dtstart := goical.NewProp(goical.PropDateTimeStart)
		dtstart.Params.Set("VALUE", "DATE")
		dtstart.Value = event.Start.Format("20060102")
		comp.Props.Set(dtstart)

		dtend := goical.NewProp(goical.PropDateTimeEnd)
		dtend.Params.Set("VALUE", "DATE")
		dtend.Value = event.End.Format("20060102")
		comp.Props.Set(dtend)
	} else {
		dtstart := goical.NewProp(goical.PropDateTimeStart)
		dtstart.SetDateTime(event.Start.UTC())
		comp.Props.Set(dtstart)

		dtend := goical.NewProp(goical.PropDateTimeEnd)
		dtend.SetDateTime(event.End.UTC())
		comp.Props.Set(dtend)
	}

	cal.Children = append(cal.Children, comp)
	return cal
}

// CreateEvent pushes a newly created event to the CalDAV server.
func (s *Sync) CreateEvent(ctx context.Context, event *ical.Event) error {
	client := s.clientForCalendar(event.CalendarID)
	if client == nil {
		return fmt.Errorf("no client found for calendar %d", event.CalendarID)
	}

	calPath := event.CalPath
	if calPath == "" {
		if event.CalendarID < len(s.store.Calendars) {
			calPath = s.store.Calendars[event.CalendarID].CalPath
		}
	}
	if calPath == "" {
		return fmt.Errorf("no calendar path for calendar %d", event.CalendarID)
	}

	resourcePath := calPath
	if len(resourcePath) > 0 && resourcePath[len(resourcePath)-1] != '/' {
		resourcePath += "/"
	}
	resourcePath += event.UID + ".ics"

	cal := eventToIcal(event)
	obj := &caldav.CalendarObject{Data: cal}

	result, err := client.PutEvent(ctx, resourcePath, obj)
	if err != nil {
		return fmt.Errorf("creating event on server: %w", err)
	}
	if result != nil && result.Path != "" {
		event.ResourcePath = result.Path
	} else {
		event.ResourcePath = resourcePath
	}
	event.CalPath = calPath

	var buf bytes.Buffer
	if err := goical.NewEncoder(&buf).Encode(cal); err == nil {
		etag := ""
		if result != nil {
			etag = result.ETag
		}
		s.cache.Put(calPath, CacheEntry{
			Path:      event.ResourcePath,
			ETag:      etag,
			Data:      buf.String(),
			FetchedAt: time.Now(),
		})
	}

	return nil
}

// UpdateEvent pushes a modified event to the CalDAV server.
func (s *Sync) UpdateEvent(ctx context.Context, event *ical.Event) error {
	client := s.clientForCalendar(event.CalendarID)
	if client == nil {
		return fmt.Errorf("no client found for calendar %d", event.CalendarID)
	}
	if event.ResourcePath == "" {
		return s.CreateEvent(ctx, event)
	}

	calPath := event.CalPath
	if calPath == "" {
		if event.CalendarID < len(s.store.Calendars) {
			calPath = s.store.Calendars[event.CalendarID].CalPath
		}
	}

	cal := eventToIcal(event)
	obj := &caldav.CalendarObject{Data: cal}

	result, err := client.PutEvent(ctx, event.ResourcePath, obj)
	if err != nil {
		return fmt.Errorf("updating event on server: %w", err)
	}

	var buf bytes.Buffer
	if err := goical.NewEncoder(&buf).Encode(cal); err == nil {
		etag := ""
		if result != nil {
			etag = result.ETag
		}
		s.cache.Put(calPath, CacheEntry{
			Path:      event.ResourcePath,
			ETag:      etag,
			Data:      buf.String(),
			FetchedAt: time.Now(),
		})
	}

	return nil
}

// DeleteEvent removes an event from the CalDAV server.
func (s *Sync) DeleteEvent(ctx context.Context, event *ical.Event) error {
	if event.ResourcePath == "" {
		return nil
	}
	client := s.clientForCalendar(event.CalendarID)
	if client == nil {
		return fmt.Errorf("no client found for calendar %d", event.CalendarID)
	}
	if err := client.DeleteEvent(ctx, event.ResourcePath); err != nil {
		return fmt.Errorf("deleting event from server: %w", err)
	}

	calPath := event.CalPath
	if calPath == "" {
		if event.CalendarID < len(s.store.Calendars) {
			calPath = s.store.Calendars[event.CalendarID].CalPath
		}
	}
	if calPath != "" {
		s.cache.Remove(calPath, event.ResourcePath)
	}

	return nil
}
