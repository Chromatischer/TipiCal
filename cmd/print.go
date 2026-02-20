package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/caldav"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/util"
	"golang.org/x/term"
)

func runPrint(args []string) {
	var showWeek, showFull bool
	remaining := args[:0]
	for _, arg := range args {
		switch arg {
		case "--week", "-w":
			showWeek = true
		case "--full", "-f":
			showFull = true
		case "--no-color":
			os.Setenv("NO_COLOR", "1")
		case "--help", "-h":
			printPrintHelp()
			return
		default:
			remaining = append(remaining, arg)
		}
	}

	if len(remaining) > 0 && remaining[0] == "agenda" {
		remaining = remaining[1:]
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	theme := config.NewTheme(cfg)
	if os.Getenv("NO_COLOR") != "" {
		theme = config.NewMonoTheme()
	}

	store := ical.NewStore()

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
			syncMgr := caldav.NewSync(clients, cache, store, theme, cfg)
			syncMgr.LoadFromCache()
		}
	}

	if len(store.Events) == 0 {
		store.LoadTestData()
	}

	days := 2
	if showWeek {
		days = 7
	}

	output := renderPrintAgenda(store, theme, cfg, days, showFull)
	fmt.Print(output)
}

func printPrintHelp() {
	fmt.Println("TipiCal Print - Output agenda to console")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tipical print [flags]")
	fmt.Println("  tipical print agenda [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --week, -w      Print next 7 days (default: today + tomorrow)")
	fmt.Println("  --full, -f      Include location and description")
	fmt.Println("  --no-color      Disable colored output")
	fmt.Println("  --help, -h      Show this help")
	fmt.Println()
}

func renderPrintAgenda(store *ical.Store, theme *config.Theme, cfg *config.Config, numDays int, full bool) string {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 0, numDays)

	width := terminalWidth()
	if width < 40 {
		width = 40
	}
	if width > 60 {
		width = 60
	}

	var lines []string
	use24h := cfg.Use24h()

	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		events := store.EventsForDay(d)
		if len(events) == 0 {
			continue
		}

		sort.Slice(events, func(i, j int) bool {
			return events[i].Start.Before(events[j].Start)
		})

		var dayLabel string
		if util.SameDay(d, now) {
			dayLabel = "Today"
		} else if util.SameDay(d, now.AddDate(0, 0, 1)) {
			dayLabel = "Tomorrow"
		} else {
			dayLabel = d.Format("Monday")
		}

		isToday := util.SameDay(d, now)

		headerStyle := lipgloss.NewStyle().Bold(true)
		if isToday {
			headerStyle = headerStyle.Foreground(theme.Today)
		} else {
			headerStyle = headerStyle.Foreground(theme.Accent)
		}

		dateStr := d.Format("January 2")
		header := headerStyle.Render(fmt.Sprintf("%s — %s", dayLabel, dateStr))
		lines = append(lines, header)

		divider := lipgloss.NewStyle().
			Foreground(theme.Border).
			Render(strings.Repeat("─", width))
		lines = append(lines, divider)

		for _, e := range events {
			color := lipgloss.Color(e.Color)
			if e.Color == "" {
				color = theme.CalendarColor(e.CalendarID)
			}

			dot := lipgloss.NewStyle().Foreground(color).Render("●")

			var timeStr string
			if e.AllDay {
				timeStr = "All day"
			} else {
				timeStr = util.FormatTimeRange(e.Start, e.End, use24h)
			}

			timeRendered := lipgloss.NewStyle().
				Foreground(theme.TextMuted).
				Width(14).
				Render(timeStr)

			title := lipgloss.NewStyle().
				Foreground(theme.Text).
				Bold(true).
				Render(e.Summary)

			eventLine := fmt.Sprintf("  %s %s  %s", dot, timeRendered, title)
			lines = append(lines, eventLine)

			if full {
				indent := "                    "
				if e.Location != "" {
					loc := lipgloss.NewStyle().
						Foreground(theme.TextFaint).
						Italic(true).
						Render(e.Location)
					lines = append(lines, indent+loc)
				}
				if e.Description != "" {
					desc := lipgloss.NewStyle().
						Foreground(theme.TextMuted).
						Render(e.Description)
					lines = append(lines, indent+desc)
				}
			}
		}

		lines = append(lines, "")
	}

	if len(lines) == 0 {
		return lipgloss.NewStyle().
			Foreground(theme.TextFaint).
			Render("No upcoming events") + "\n"
	}

	return strings.Join(lines, "\n")
}

func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 {
		return 60
	}
	return width
}
