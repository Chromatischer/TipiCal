package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

type calendarsOptions struct {
	showJSON bool
	search   string
}

func runCalendars(args []string) int {
	opts, err := parseCalendarsOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}
	if opts == nil {
		return exitCodeSuccess
	}

	ctx, err := loadCLIContext(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return exitCodeFailure
	}
	if err := ensureRemoteData(ctx); err != nil && len(ctx.store.Calendars) == 0 {
		fmt.Fprintf(os.Stderr, "Error loading calendars: %v\n", err)
		return exitCodeFailure
	}

	var calendars []map[string]any
	for _, cal := range ctx.store.Calendars {
		if opts.search != "" && !strings.Contains(strings.ToLower(cal.Name), strings.ToLower(opts.search)) && !strings.Contains(strings.ToLower(cal.Source), strings.ToLower(opts.search)) {
			continue
		}
		calendars = append(calendars, map[string]any{
			"name":      cal.Name,
			"source":    cal.Source,
			"color":     cal.Color,
			"readonly":  cal.ReadOnly,
			"calendar":  cal.CalPath,
			"config_id": cal.ConfigIndex,
		})
	}

	sort.Slice(calendars, func(i, j int) bool {
		return fmt.Sprint(calendars[i]["name"]) < fmt.Sprint(calendars[j]["name"])
	})

	if len(calendars) == 0 {
		if opts.showJSON {
			payload, _ := json.MarshalIndent(map[string]any{"calendars": []any{}}, "", "  ")
			fmt.Printf("%s\n", payload)
		} else {
			fmt.Println("No calendars found")
		}
		return exitCodeNoResults
	}

	if opts.showJSON {
		payload, _ := json.MarshalIndent(map[string]any{"calendars": calendars}, "", "  ")
		fmt.Printf("%s\n", payload)
		return exitCodeSuccess
	}

	for _, cal := range calendars {
		name := fmt.Sprint(cal["name"])
		source := fmt.Sprint(cal["source"])
		if source != "" {
			fmt.Printf("%s (%s)\n", name, source)
		} else {
			fmt.Println(name)
		}
	}
	return exitCodeSuccess
}

func parseCalendarsOptions(args []string) (*calendarsOptions, error) {
	fs := flag.NewFlagSet("calendars", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	opts := &calendarsOptions{}
	var help bool
	fs.BoolVar(&opts.showJSON, "json", false, "")
	fs.BoolVar(&help, "help", false, "")
	fs.BoolVar(&help, "h", false, "")
	fs.StringVar(&opts.search, "search", "", "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if help {
		fmt.Println("Usage: tipical calendars [--json] [--search TEXT]")
		return nil, nil
	}
	return opts, nil
}
