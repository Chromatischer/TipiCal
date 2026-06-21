package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/util"
)

func handleEvents(args []string) {
	if len(args) == 0 {
		printEventsHelp()
		return
	}

	switch args[0] {
	case "list", "ls":
		eventsList(args[1:])
	case "show", "get":
		eventsShow(args[1:])
	case "add", "create":
		eventsAdd(args[1:])
	case "delete", "rm", "del":
		eventsDelete(args[1:])
	case "help", "--help", "-h":
		printEventsHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown events command: %s\n\n", args[0])
		printEventsHelp()
		os.Exit(1)
	}
}

func printEventsHelp() {
	fmt.Println("TipiCal Events - Inspect and manage calendar events")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tipical events list [flags]      List events in a date range")
	fmt.Println("  tipical events show --uid UID     Show a single event")
	fmt.Println("  tipical events add [flags]        Create a new event")
	fmt.Println("  tipical events delete --uid UID   Delete an event")
	fmt.Println()
	fmt.Println("List flags:")
	fmt.Println("  --from DATE       Start of range (YYYY-MM-DD, default today)")
	fmt.Println("  --to DATE         End of range, exclusive (default --from + --days)")
	fmt.Println("  --days N          Days from --from when --to omitted (default 7)")
	fmt.Println("  --calendar NAME   Restrict to a calendar (name or id)")
	fmt.Println("  --search TEXT     Substring filter on title/location/description")
	fmt.Println("  --json            Output JSON")
	fmt.Println()
	fmt.Println("Add flags:")
	fmt.Println("  --calendar NAME   Target calendar (name or id, required)")
	fmt.Println("  --title TEXT      Event title (required)")
	fmt.Println("  --start WHEN      Start (YYYY-MM-DD or \"YYYY-MM-DD HH:MM\", required)")
	fmt.Println("  --end WHEN        End, exclusive (default +1h, or +1 day if all-day)")
	fmt.Println("  --all-day         Make an all-day event")
	fmt.Println("  --location TEXT   Location")
	fmt.Println("  --description TEXT Description / notes")
	fmt.Println()
}

func eventsList(args []string) {
	fs := flag.NewFlagSet("events list", flag.ExitOnError)
	from := fs.String("from", "", "start of range")
	to := fs.String("to", "", "end of range (exclusive)")
	days := fs.Int("days", 7, "days from --from")
	calendar := fs.String("calendar", "", "calendar name or id")
	search := fs.String("search", "", "substring filter")
	asJSON := fs.Bool("json", false, "output JSON")
	fs.Parse(args)

	_, store, _, err := loadStore(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if *from != "" {
		t, _, err := parseDateTime(*from)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		start = t
	}
	end := start.AddDate(0, 0, *days)
	if *to != "" {
		t, _, err := parseDateTime(*to)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		end = t
	}

	calFilter := -1
	if *calendar != "" {
		id, err := resolveCalendar(store, *calendar)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		calFilter = id
	}

	q := strings.ToLower(strings.TrimSpace(*search))
	var events []*ical.Event
	for _, e := range store.EventsInRange(start, end) {
		if calFilter >= 0 && e.CalendarID != calFilter {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Summary + "\n" + e.Location + "\n" + e.Description)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		events = append(events, e)
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Start.Before(events[j].Start)
	})

	if *asJSON {
		printEventsJSON(store, events)
		return
	}

	if len(events) == 0 {
		fmt.Println("No events in range.")
		return
	}
	for _, e := range events {
		fmt.Println(formatEventLine(store, e))
	}
}

func eventsShow(args []string) {
	fs := flag.NewFlagSet("events show", flag.ExitOnError)
	uid := fs.String("uid", "", "event UID")
	asJSON := fs.Bool("json", false, "output JSON")
	fs.Parse(args)

	if *uid == "" {
		fmt.Fprintln(os.Stderr, "Error: --uid is required")
		os.Exit(1)
	}

	_, store, _, err := loadStore(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	e := store.FindEvent(*uid)
	if e == nil {
		fmt.Fprintf(os.Stderr, "Error: no event with uid %q\n", *uid)
		os.Exit(1)
	}

	if *asJSON {
		printEventsJSON(store, []*ical.Event{e})
		return
	}

	fmt.Printf("Title:       %s\n", e.Summary)
	fmt.Printf("UID:         %s\n", e.UID)
	fmt.Printf("Calendar:    %s\n", store.CalendarName(e.CalendarID))
	if e.AllDay {
		fmt.Printf("When:        %s – %s (all day)\n", e.Start.Format("2006-01-02"), e.End.Format("2006-01-02"))
	} else {
		fmt.Printf("When:        %s – %s\n", e.Start.Format("2006-01-02 15:04"), e.End.Format("2006-01-02 15:04"))
	}
	if e.Location != "" {
		fmt.Printf("Location:    %s\n", e.Location)
	}
	if e.Description != "" {
		fmt.Printf("Description: %s\n", e.Description)
	}
	if e.Recurring {
		fmt.Printf("Recurrence:  %s\n", e.RecurrenceDescription())
	}
	if e.Status != "" {
		fmt.Printf("Status:      %s\n", e.Status)
	}
}

func eventsAdd(args []string) {
	fs := flag.NewFlagSet("events add", flag.ExitOnError)
	calendar := fs.String("calendar", "", "target calendar (name or id)")
	title := fs.String("title", "", "event title")
	startStr := fs.String("start", "", "start time")
	endStr := fs.String("end", "", "end time (exclusive)")
	allDay := fs.Bool("all-day", false, "all-day event")
	location := fs.String("location", "", "location")
	description := fs.String("description", "", "description")
	fs.Parse(args)

	if *title == "" || *startStr == "" {
		fmt.Fprintln(os.Stderr, "Error: --title and --start are required")
		os.Exit(1)
	}

	_, store, syncMgr, err := loadStore(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if syncMgr == nil {
		fmt.Fprintln(os.Stderr, "Error: no CalDAV calendars configured; run 'tipical auth add' first")
		os.Exit(1)
	}

	calID, err := resolveCalendar(store, *calendar)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if store.Calendars[calID].ReadOnly {
		fmt.Fprintf(os.Stderr, "Error: calendar %q is read-only\n", store.CalendarName(calID))
		os.Exit(1)
	}

	start, isDate, err := parseDateTime(*startStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	isAllDay := *allDay || isDate

	var end time.Time
	if *endStr != "" {
		t, _, err := parseDateTime(*endStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		end = t
	} else if isAllDay {
		end = start.AddDate(0, 0, 1)
	} else {
		end = start.Add(time.Hour)
	}

	event := &ical.Event{
		UID:         newEventUID(),
		Summary:     *title,
		Description: *description,
		Location:    *location,
		Start:       start,
		End:         end,
		AllDay:      isAllDay,
		CalendarID:  calID,
		Status:      "CONFIRMED",
		CalPath:     store.Calendars[calID].CalPath,
	}
	store.AddEvent(event)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := syncMgr.CreateEvent(ctx, event); err != nil {
		fmt.Fprintf(os.Stderr, "Error: creating event: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created event %q on %s\n", event.Summary, store.CalendarName(calID))
	fmt.Printf("UID: %s\n", event.UID)
}

func eventsDelete(args []string) {
	fs := flag.NewFlagSet("events delete", flag.ExitOnError)
	uid := fs.String("uid", "", "event UID")
	fs.Parse(args)

	if *uid == "" {
		fmt.Fprintln(os.Stderr, "Error: --uid is required")
		os.Exit(1)
	}

	_, store, syncMgr, err := loadStore(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if syncMgr == nil {
		fmt.Fprintln(os.Stderr, "Error: no CalDAV calendars configured")
		os.Exit(1)
	}

	event := store.FindEvent(*uid)
	if event == nil {
		fmt.Fprintf(os.Stderr, "Error: no event with uid %q\n", *uid)
		os.Exit(1)
	}
	if event.CalendarID >= 0 && event.CalendarID < len(store.Calendars) && store.Calendars[event.CalendarID].ReadOnly {
		fmt.Fprintf(os.Stderr, "Error: calendar %q is read-only\n", store.CalendarName(event.CalendarID))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := syncMgr.DeleteEvent(ctx, event); err != nil {
		fmt.Fprintf(os.Stderr, "Error: deleting event: %v\n", err)
		os.Exit(1)
	}
	store.RemoveEvent(*uid)

	fmt.Printf("Deleted event %q\n", event.Summary)
}

func formatEventLine(store *ical.Store, e *ical.Event) string {
	var when string
	if e.AllDay {
		when = e.Start.Format("2006-01-02") + " all-day"
	} else {
		when = e.Start.Format("2006-01-02 15:04") + "–" + e.End.Format("15:04")
	}
	// Calendar names may contain emojis (2 columns wide), so pad by display
	// width rather than rune/byte count to keep columns aligned.
	cal := util.TruncateText(store.CalendarName(e.CalendarID), 18)
	line := fmt.Sprintf("%-22s  %s  %s", when, padDisplay(cal, 18), e.Summary)
	if e.Location != "" {
		line += "  @ " + e.Location
	}
	return line + fmt.Sprintf("  [%s]", e.UID)
}

// padDisplay right-pads s with spaces to the given display width (columns),
// accounting for wide runes such as emojis.
func padDisplay(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func printEventsJSON(store *ical.Store, events []*ical.Event) {
	type ev struct {
		UID         string `json:"uid"`
		Summary     string `json:"summary"`
		Description string `json:"description,omitempty"`
		Location    string `json:"location,omitempty"`
		Calendar    string `json:"calendar"`
		CalendarID  int    `json:"calendar_id"`
		Start       string `json:"start"`
		End         string `json:"end"`
		AllDay      bool   `json:"all_day"`
		Recurrence  string `json:"recurrence,omitempty"`
	}
	out := make([]ev, 0, len(events))
	for _, e := range events {
		startLayout, endLayout := "2006-01-02 15:04", "2006-01-02 15:04"
		if e.AllDay {
			startLayout, endLayout = "2006-01-02", "2006-01-02"
		}
		out = append(out, ev{
			UID:         e.UID,
			Summary:     e.Summary,
			Description: e.Description,
			Location:    e.Location,
			Calendar:    store.CalendarName(e.CalendarID),
			CalendarID:  e.CalendarID,
			Start:       e.Start.Format(startLayout),
			End:         e.End.Format(endLayout),
			AllDay:      e.AllDay,
			Recurrence:  e.RecurrenceDescription(),
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}
