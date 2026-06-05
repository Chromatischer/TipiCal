package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/tipical/tipical/caldav"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
)

const (
	exitCodeSuccess   = 0
	exitCodeFailure   = 1
	exitCodeNoResults = 3
)

type cliContext struct {
	cfg     *config.Config
	theme   *config.Theme
	store   *ical.Store
	syncMgr *caldav.Sync
}

type eventRecord struct {
	Event       *ical.Event
	Calendar    ical.CalendarInfo
	Occurrences int
}

func loadCLIContext(loadDemoData bool) (*cliContext, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	theme := config.NewTheme(cfg)
	store := ical.NewStore()

	var syncMgr *caldav.Sync
	if len(cfg.Calendars) > 0 {
		cache, err := caldav.NewCache(cfg.Sync.CacheDir)
		if err != nil {
			cache, _ = caldav.NewCache(os.TempDir())
		}

		var clients []*caldav.Client
		for i := range cfg.Calendars {
			client, err := caldav.NewClient(&cfg.Calendars[i], i)
			if err != nil {
				continue
			}
			clients = append(clients, client)
		}

		if len(clients) > 0 {
			syncMgr = caldav.NewSync(clients, cache, store, theme, cfg)
			syncMgr.LoadFromCache()
		}
	}

	if loadDemoData && syncMgr == nil {
		store.LoadTestData()
	}

	return &cliContext{
		cfg:     cfg,
		theme:   theme,
		store:   store,
		syncMgr: syncMgr,
	}, nil
}

func ensureRemoteData(ctx *cliContext) error {
	if ctx.syncMgr == nil {
		return nil
	}
	if len(ctx.store.Calendars) > 0 {
		return nil
	}
	syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return ctx.syncMgr.SyncAll(syncCtx)
}

func calendarInfoFor(store *ical.Store, id int) ical.CalendarInfo {
	if id >= 0 && id < len(store.Calendars) {
		return store.Calendars[id]
	}
	return ical.CalendarInfo{
		ID:    id,
		Name:  "",
		Color: "",
	}
}

func sortEventRecords(events []eventRecord) {
	sort.Slice(events, func(i, j int) bool {
		a := events[i].Event
		b := events[j].Event
		if !a.Start.Equal(b.Start) {
			return a.Start.Before(b.Start)
		}
		if !a.End.Equal(b.End) {
			return a.End.Before(b.End)
		}
		if !strings.EqualFold(events[i].Calendar.Name, events[j].Calendar.Name) {
			return strings.ToLower(events[i].Calendar.Name) < strings.ToLower(events[j].Calendar.Name)
		}
		return strings.ToLower(a.Summary) < strings.ToLower(b.Summary)
	})
}

func uniqueRecurringEvents(events []eventRecord) []eventRecord {
	seen := make(map[string]bool, len(events))
	var result []eventRecord
	for _, rec := range events {
		key := fmt.Sprintf("%s|%s|%s|%d", rec.Event.UID, rec.Event.Start.UTC().Format(time.RFC3339), rec.Event.End.UTC().Format(time.RFC3339), rec.Event.CalendarID)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, rec)
	}
	return result
}

func normalizeRecurringEvents(events []eventRecord) []eventRecord {
	type aggregate struct {
		record eventRecord
	}

	agg := make(map[string]*aggregate)
	var singles []eventRecord

	for _, rec := range events {
		if !rec.Event.Recurring {
			singles = append(singles, rec)
			continue
		}

		key := fmt.Sprintf("%d|%s", rec.Event.CalendarID, rec.Event.UID)
		group := agg[key]
		if group == nil {
			copyEvent := *rec.Event
			agg[key] = &aggregate{
				record: eventRecord{
					Event:       &copyEvent,
					Calendar:    rec.Calendar,
					Occurrences: 1,
				},
			}
			continue
		}

		group.record.Occurrences++
		if rec.Event.Start.Before(group.record.Event.Start) {
			group.record.Event.Start = rec.Event.Start
		}
		if rec.Event.End.After(group.record.Event.End) {
			group.record.Event.End = rec.Event.End
		}
	}

	result := append([]eventRecord{}, singles...)
	for _, item := range agg {
		result = append(result, item.record)
	}
	return result
}

func generateUID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("tical-%d-%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

func parseISODate(value string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", value)
	}
	return t, nil
}

func parseEventTime(date time.Time, value string) (time.Time, error) {
	t, err := time.ParseInLocation("15:04", value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time %q, expected HH:MM", value)
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, time.Local), nil
}

func resolveCalendarByName(store *ical.Store, name string) (ical.CalendarInfo, error) {
	if strings.TrimSpace(name) == "" {
		return ical.CalendarInfo{}, errors.New("calendar name is required")
	}

	var matches []ical.CalendarInfo
	query := strings.ToLower(strings.TrimSpace(name))
	for _, cal := range store.Calendars {
		if strings.EqualFold(cal.Name, name) {
			return cal, nil
		}
		if strings.Contains(strings.ToLower(cal.Name), query) {
			matches = append(matches, cal)
		}
	}

	switch len(matches) {
	case 0:
		return ical.CalendarInfo{}, fmt.Errorf("calendar %q not found", name)
	case 1:
		return matches[0], nil
	default:
		var names []string
		for _, cal := range matches {
			names = append(names, cal.Name)
		}
		return ical.CalendarInfo{}, fmt.Errorf("calendar %q is ambiguous: %s", name, strings.Join(names, ", "))
	}
}
