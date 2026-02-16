package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tipical/tipical/caldav"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui"
)

func runApp(demo bool) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	theme := config.NewTheme(cfg)
	store := ical.NewStore()

	// Build CalDAV sync manager if calendars are configured
	var syncMgr *caldav.Sync
	if !demo && len(cfg.Calendars) > 0 {
		// Create cache
		cache, err := caldav.NewCache(cfg.Sync.CacheDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not create cache: %v\n", err)
			// Continue without cache - create a fallback in temp dir
			cache, _ = caldav.NewCache(os.TempDir())
		}

		// Create CalDAV clients
		var clients []*caldav.Client
		for i := range cfg.Calendars {
			client, err := caldav.NewClient(&cfg.Calendars[i], i)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: calendar %q: %v\n", cfg.Calendars[i].Name, err)
				continue
			}
			clients = append(clients, client)
		}

		if len(clients) > 0 {
			syncMgr = caldav.NewSync(clients, cache, store, theme, cfg)
			syncMgr.LoadFromCache()
		}
	}

	// If no CalDAV calendars configured, load test data for demo
	if syncMgr == nil {
		store.LoadTestData()
	}

	app := ui.NewApp(cfg, store, syncMgr)
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
