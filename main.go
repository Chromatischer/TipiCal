package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-ical/terminal-ical/caldav"
	"github.com/terminal-ical/terminal-ical/config"
	"github.com/terminal-ical/terminal-ical/ical"
	"github.com/terminal-ical/terminal-ical/setup"
	"github.com/terminal-ical/terminal-ical/ui"
)

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--setup" {
			if err := setup.RunWizard(); err != nil {
				fmt.Fprintf(os.Stderr, "Setup error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	theme := config.NewTheme(cfg)
	store := ical.NewStore()

	// Build CalDAV sync manager if calendars are configured
	var syncMgr *caldav.Sync
	if len(cfg.Calendars) > 0 {
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
