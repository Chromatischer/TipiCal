package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/tipical/tipical/caldav"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
)

// loadStore loads the configuration and builds an event store backed by the
// configured CalDAV calendars. It mirrors the setup performed by the TUI and
// print command so the CLI and MCP server share one code path.
//
// The returned Sync is nil when no calendars are configured (or none could be
// constructed); callers that only need to read cached data can still use the
// store. When syncNow is true a best-effort live SyncAll is performed so the
// store reflects the server; on failure the cached data already loaded is kept.
func loadStore(syncNow bool) (*config.Config, *ical.Store, *caldav.Sync, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("loading config: %w", err)
	}

	theme := config.NewTheme(cfg)
	store := ical.NewStore()

	if len(cfg.Calendars) == 0 {
		return cfg, store, nil, nil
	}

	cache, err := caldav.NewCache(cfg.Sync.CacheDir)
	if err != nil {
		cache, _ = caldav.NewCache(os.TempDir())
	}

	var clients []*caldav.Client
	for i := range cfg.Calendars {
		client, err := caldav.NewClient(&cfg.Calendars[i], i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: calendar %q: %v\n", cfg.Calendars[i].Name, err)
			continue
		}
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return cfg, store, nil, nil
	}

	syncMgr := caldav.NewSync(clients, cache, store, theme, cfg)
	syncMgr.LoadFromCache()

	if syncNow {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := syncMgr.SyncAll(ctx); err != nil {
			// Keep whatever was loaded from cache; surface a warning only.
			fmt.Fprintf(os.Stderr, "Warning: sync failed, using cached data: %v\n", err)
		}
	}

	return cfg, store, syncMgr, nil
}

// resolveCalendar maps a calendar name (or numeric ID) to a registered
// calendar ID in the store. Matching is case-sensitive on the discovered
// calendar name; a bare integer is treated as a direct ID.
func resolveCalendar(store *ical.Store, nameOrID string) (int, error) {
	if nameOrID == "" {
		if len(store.Calendars) == 1 {
			return store.Calendars[0].ID, nil
		}
		return -1, fmt.Errorf("calendar is required (%s)", calendarChoices(store))
	}
	// Try numeric ID first.
	var id int
	if _, err := fmt.Sscanf(nameOrID, "%d", &id); err == nil {
		if id >= 0 && id < len(store.Calendars) {
			return id, nil
		}
	}
	for _, c := range store.Calendars {
		if c.Name == nameOrID {
			return c.ID, nil
		}
	}
	return -1, fmt.Errorf("unknown calendar %q (%s)", nameOrID, calendarChoices(store))
}

func calendarChoices(store *ical.Store) string {
	if len(store.Calendars) == 0 {
		return "no calendars synced"
	}
	out := "available: "
	for i, c := range store.Calendars {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%d=%s", c.ID, c.Name)
	}
	return out
}

// newEventUID generates a unique event UID matching the format used by the TUI
// editor.
func newEventUID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("tical-%d-%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

// parseDateTime parses a date or date-time in a few common layouts. The
// returned bool reports whether the input was a date only (no time component),
// which the caller can treat as an all-day boundary.
func parseDateTime(s string) (time.Time, bool, error) {
	dateLayouts := []string{"2006-01-02"}
	for _, layout := range dateLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true, nil
		}
	}
	dtLayouts := []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range dtLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, false, nil
		}
	}
	return time.Time{}, false, fmt.Errorf("invalid date/time %q (use YYYY-MM-DD or \"YYYY-MM-DD HH:MM\")", s)
}
