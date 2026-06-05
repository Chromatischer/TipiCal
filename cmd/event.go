package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tipical/tipical/ical"
)

func runAgenda(args []string) int {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		args = append([]string{"--calendar", args[0]}, args[1:]...)
	}
	return runPrint(append([]string{"agenda"}, args...))
}

func runEvent(args []string) int {
	if len(args) == 0 {
		printEventHelp()
		return exitCodeFailure
	}

	switch args[0] {
	case "create":
		return runEventCreate(args[1:])
	case "update":
		return runEventUpdate(args[1:])
	case "move":
		return runEventMove(args[1:])
	case "delete":
		return runEventDelete(args[1:])
	case "help", "-h", "--help":
		printEventHelp()
		return exitCodeSuccess
	default:
		fmt.Fprintf(os.Stderr, "Unknown event command: %s\n", args[0])
		printEventHelp()
		return exitCodeFailure
	}
}

func printEventHelp() {
	fmt.Println("TipiCal Event Commands")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tipical event create --calendar NAME --summary TEXT --date YYYY-MM-DD [--start HH:MM --end HH:MM | --all-day]")
	fmt.Println("  tipical event update UID [field flags]")
	fmt.Println("  tipical event move UID --date YYYY-MM-DD [--start HH:MM] [--end HH:MM]")
	fmt.Println("  tipical event delete UID")
	fmt.Println()
}

type eventMutationOptions struct {
	calendar    string
	summary     string
	description string
	location    string
	status      string
	date        string
	start       string
	end         string
	allDay      bool
	showJSON    bool
}

func runEventCreate(args []string) int {
	opts, err := parseEventMutationOptions("create", args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}

	ctx, err := loadCLIContext(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return exitCodeFailure
	}
	if err := ensureRemoteData(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading calendars: %v\n", err)
		return exitCodeFailure
	}
	if ctx.syncMgr == nil {
		fmt.Fprintln(os.Stderr, "No CalDAV calendars configured")
		return exitCodeFailure
	}

	event, err := buildCreatedEvent(ctx.store, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}

	callCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ctx.syncMgr.CreateEvent(callCtx, event); err != nil {
		fmt.Fprintf(os.Stderr, "Create failed: %v\n", err)
		return exitCodeFailure
	}

	if opts.showJSON {
		payload, _ := json.MarshalIndent(map[string]any{"event": jsonEventFromRecord(eventRecord{Event: event, Calendar: calendarInfoFor(ctx.store, event.CalendarID), Occurrences: 1})}, "", "  ")
		fmt.Printf("%s\n", payload)
	} else {
		fmt.Printf("Created %s (%s)\n", event.Summary, event.UID)
	}
	return exitCodeSuccess
}

func runEventUpdate(args []string) int {
	uid, opts, err := parseUIDMutationOptions("update", args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}

	ctx, err := loadCLIContext(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return exitCodeFailure
	}
	if err := ensureRemoteData(ctx); err != nil && len(ctx.store.Events) == 0 {
		fmt.Fprintf(os.Stderr, "Error loading events: %v\n", err)
		return exitCodeFailure
	}
	if ctx.syncMgr == nil {
		fmt.Fprintln(os.Stderr, "No CalDAV calendars configured")
		return exitCodeFailure
	}

	original := ctx.store.FindEvent(uid)
	if original == nil {
		fmt.Fprintf(os.Stderr, "Event %q not found\n", uid)
		return exitCodeNoResults
	}

	updated, err := mutateEvent(ctx.store, original, opts, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}

	callCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ctx.syncMgr.UpdateEvent(callCtx, updated); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		return exitCodeFailure
	}

	if opts.showJSON {
		payload, _ := json.MarshalIndent(map[string]any{"event": jsonEventFromRecord(eventRecord{Event: updated, Calendar: calendarInfoFor(ctx.store, updated.CalendarID), Occurrences: 1})}, "", "  ")
		fmt.Printf("%s\n", payload)
	} else {
		fmt.Printf("Updated %s (%s)\n", updated.Summary, updated.UID)
	}
	return exitCodeSuccess
}

func runEventMove(args []string) int {
	uid, opts, err := parseUIDMutationOptions("move", args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}
	if opts.date == "" {
		fmt.Fprintln(os.Stderr, "--date is required")
		return exitCodeFailure
	}

	ctx, err := loadCLIContext(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return exitCodeFailure
	}
	if err := ensureRemoteData(ctx); err != nil && len(ctx.store.Events) == 0 {
		fmt.Fprintf(os.Stderr, "Error loading events: %v\n", err)
		return exitCodeFailure
	}
	if ctx.syncMgr == nil {
		fmt.Fprintln(os.Stderr, "No CalDAV calendars configured")
		return exitCodeFailure
	}

	original := ctx.store.FindEvent(uid)
	if original == nil {
		fmt.Fprintf(os.Stderr, "Event %q not found\n", uid)
		return exitCodeNoResults
	}

	moved, err := mutateEvent(ctx.store, original, opts, true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}

	callCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ctx.syncMgr.UpdateEvent(callCtx, moved); err != nil {
		fmt.Fprintf(os.Stderr, "Move failed: %v\n", err)
		return exitCodeFailure
	}

	if opts.showJSON {
		payload, _ := json.MarshalIndent(map[string]any{"event": jsonEventFromRecord(eventRecord{Event: moved, Calendar: calendarInfoFor(ctx.store, moved.CalendarID), Occurrences: 1})}, "", "  ")
		fmt.Printf("%s\n", payload)
	} else {
		fmt.Printf("Moved %s (%s)\n", moved.Summary, moved.UID)
	}
	return exitCodeSuccess
}

func runEventDelete(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "delete requires a UID")
		return exitCodeFailure
	}
	uid := args[0]
	showJSON := len(args) > 1 && args[1] == "--json"

	ctx, err := loadCLIContext(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return exitCodeFailure
	}
	if err := ensureRemoteData(ctx); err != nil && len(ctx.store.Events) == 0 {
		fmt.Fprintf(os.Stderr, "Error loading events: %v\n", err)
		return exitCodeFailure
	}
	if ctx.syncMgr == nil {
		fmt.Fprintln(os.Stderr, "No CalDAV calendars configured")
		return exitCodeFailure
	}

	event := ctx.store.FindEvent(uid)
	if event == nil {
		fmt.Fprintf(os.Stderr, "Event %q not found\n", uid)
		return exitCodeNoResults
	}

	callCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ctx.syncMgr.DeleteEvent(callCtx, event); err != nil {
		fmt.Fprintf(os.Stderr, "Delete failed: %v\n", err)
		return exitCodeFailure
	}

	if showJSON {
		payload, _ := json.MarshalIndent(map[string]any{"deleted": uid}, "", "  ")
		fmt.Printf("%s\n", payload)
	} else {
		fmt.Printf("Deleted %s (%s)\n", event.Summary, event.UID)
	}
	return exitCodeSuccess
}

func parseEventMutationOptions(name string, args []string) (*eventMutationOptions, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	opts := &eventMutationOptions{}
	fs.StringVar(&opts.calendar, "calendar", "", "")
	fs.StringVar(&opts.summary, "summary", "", "")
	fs.StringVar(&opts.description, "description", "", "")
	fs.StringVar(&opts.location, "location", "", "")
	fs.StringVar(&opts.status, "status", "", "")
	fs.StringVar(&opts.date, "date", "", "")
	fs.StringVar(&opts.start, "start", "", "")
	fs.StringVar(&opts.end, "end", "", "")
	fs.BoolVar(&opts.allDay, "all-day", false, "")
	fs.BoolVar(&opts.showJSON, "json", false, "")
	return opts, fs.Parse(args)
}

func parseUIDMutationOptions(name string, args []string) (string, *eventMutationOptions, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("%s requires a UID", name)
	}
	opts, err := parseEventMutationOptions(name, args[1:])
	if err != nil {
		return "", nil, err
	}
	return args[0], opts, nil
}

func buildCreatedEvent(store *ical.Store, opts *eventMutationOptions) (*ical.Event, error) {
	if opts.summary == "" {
		return nil, fmt.Errorf("--summary is required")
	}
	if opts.calendar == "" {
		return nil, fmt.Errorf("--calendar is required")
	}
	if opts.date == "" {
		return nil, fmt.Errorf("--date is required")
	}

	cal, err := resolveCalendarByName(store, opts.calendar)
	if err != nil {
		return nil, err
	}
	if cal.ReadOnly {
		return nil, fmt.Errorf("calendar %q is read-only", cal.Name)
	}

	date, err := parseISODate(opts.date)
	if err != nil {
		return nil, err
	}

	event := &ical.Event{
		UID:         generateUID(),
		Summary:     opts.summary,
		Description: opts.description,
		Location:    opts.location,
		Status:      "CONFIRMED",
		CalendarID:  cal.ID,
		CalPath:     cal.CalPath,
	}
	if opts.status != "" {
		event.Status = strings.ToUpper(opts.status)
	}

	if opts.allDay {
		event.AllDay = true
		event.Start = date
		event.End = date.AddDate(0, 0, 1)
		return event, nil
	}

	startValue := opts.start
	if startValue == "" {
		startValue = "09:00"
	}
	endValue := opts.end
	if endValue == "" {
		endValue = "10:00"
	}

	start, err := parseEventTime(date, startValue)
	if err != nil {
		return nil, err
	}
	end, err := parseEventTime(date, endValue)
	if err != nil {
		return nil, err
	}
	if !end.After(start) {
		return nil, fmt.Errorf("end time must be after start time")
	}
	event.Start = start
	event.End = end
	return event, nil
}

func mutateEvent(store *ical.Store, original *ical.Event, opts *eventMutationOptions, moveOnly bool) (*ical.Event, error) {
	event := *original

	if !moveOnly && opts.summary != "" {
		event.Summary = opts.summary
	}
	if !moveOnly && opts.description != "" {
		event.Description = opts.description
	}
	if !moveOnly && opts.location != "" {
		event.Location = opts.location
	}
	if !moveOnly && opts.status != "" {
		event.Status = strings.ToUpper(opts.status)
	}
	if !moveOnly && opts.calendar != "" {
		cal, err := resolveCalendarByName(store, opts.calendar)
		if err != nil {
			return nil, err
		}
		if cal.ReadOnly {
			return nil, fmt.Errorf("calendar %q is read-only", cal.Name)
		}
		event.CalendarID = cal.ID
		event.CalPath = cal.CalPath
	}

	if opts.allDay {
		event.AllDay = true
	}

	if opts.date != "" {
		date, err := parseISODate(opts.date)
		if err != nil {
			return nil, err
		}
		if event.AllDay || opts.allDay {
			event.AllDay = true
			event.Start = date
			event.End = date.AddDate(0, 0, 1)
		} else {
			duration := event.End.Sub(event.Start)
			event.Start = time.Date(date.Year(), date.Month(), date.Day(), event.Start.Hour(), event.Start.Minute(), 0, 0, time.Local)
			event.End = event.Start.Add(duration)
		}
	}

	if !event.AllDay && opts.start != "" {
		start, err := parseEventTime(event.Start, opts.start)
		if err != nil {
			return nil, err
		}
		duration := event.End.Sub(event.Start)
		event.Start = start
		event.End = start.Add(duration)
	}

	if !event.AllDay && opts.end != "" {
		end, err := parseEventTime(event.Start, opts.end)
		if err != nil {
			return nil, err
		}
		event.End = end
	}

	if !event.AllDay && !event.End.After(event.Start) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	return &event, nil
}
