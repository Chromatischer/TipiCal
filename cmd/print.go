package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/util"
	"golang.org/x/term"
)

type printOptions struct {
	mode               string
	numDays            int
	showFull           bool
	showJSON           bool
	noColor            bool
	calendarFilters    []string
	search             string
	from               time.Time
	to                 time.Time
	hasFrom            bool
	hasTo              bool
	formatFields       []string
	dedupeRecurring    bool
	normalizeRecurring bool
}

type jsonEvent struct {
	UID             string `json:"uid"`
	Summary         string `json:"summary"`
	Description     string `json:"description,omitempty"`
	Location        string `json:"location,omitempty"`
	Start           string `json:"start"`
	End             string `json:"end"`
	AllDay          bool   `json:"all_day"`
	Color           string `json:"color,omitempty"`
	Status          string `json:"status,omitempty"`
	Recurring       bool   `json:"recurring,omitempty"`
	Calendar        string `json:"calendar,omitempty"`
	CalendarSource  string `json:"calendar_source,omitempty"`
	OccurrenceCount int    `json:"occurrence_count,omitempty"`
}

type jsonDay struct {
	Date   string      `json:"date"`
	Events []jsonEvent `json:"events"`
}

type jsonOutput struct {
	Generated string      `json:"generated"`
	Days      []jsonDay   `json:"days,omitempty"`
	Events    []jsonEvent `json:"events,omitempty"`
}

func runPrint(args []string) int {
	opts, err := parsePrintOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}
	if opts == nil {
		return exitCodeSuccess
	}

	ctx, err := loadCLIContext(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return exitCodeFailure
	}

	theme := ctx.theme
	if os.Getenv("NO_COLOR") != "" || opts.noColor || opts.showJSON || len(opts.formatFields) > 0 {
		theme = config.NewMonoTheme()
	}

	events, err := filteredEvents(ctx.store, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFailure
	}

	if len(events) == 0 {
		if !opts.showJSON && len(opts.formatFields) == 0 {
			fmt.Print(lipgloss.NewStyle().Foreground(theme.TextFaint).Render("No matching events") + "\n")
		} else if opts.showJSON {
			fmt.Print(renderPrintJSON(events, opts))
		}
		return exitCodeNoResults
	}

	switch {
	case opts.showJSON:
		fmt.Print(renderPrintJSON(events, opts))
	case len(opts.formatFields) > 0:
		fmt.Print(renderFormattedEvents(events, ctx.cfg.Use24h(), opts.formatFields))
	default:
		fmt.Print(renderPrintAgenda(events, theme, ctx.cfg.Use24h(), opts.showFull))
	}
	return exitCodeSuccess
}

func parsePrintOptions(args []string) (*printOptions, error) {
	opts := &printOptions{mode: "agenda"}
	args = append([]string{}, args...)

	if len(args) > 0 {
		switch args[0] {
		case "agenda", "week", "3day":
			opts.mode = args[0]
			args = args[1:]
		}
	}

	fs := flag.NewFlagSet("print", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var calendarArg string
	var fromArg string
	var toArg string
	var formatArg string

	fs.BoolVar(&opts.showFull, "full", false, "")
	fs.BoolVar(&opts.showFull, "f", false, "")
	fs.BoolVar(&opts.showJSON, "json", false, "")
	fs.BoolVar(&opts.showJSON, "j", false, "")
	fs.BoolVar(&opts.noColor, "no-color", false, "")
	fs.BoolVar(&opts.dedupeRecurring, "dedupe-recurring", false, "")
	fs.BoolVar(&opts.normalizeRecurring, "normalize-recurring", false, "")
	fs.IntVar(&opts.numDays, "days", 0, "")
	fs.IntVar(&opts.numDays, "d", 0, "")
	fs.StringVar(&calendarArg, "calendar", "", "")
	fs.StringVar(&opts.search, "search", "", "")
	fs.StringVar(&fromArg, "from", "", "")
	fs.StringVar(&toArg, "to", "", "")
	fs.StringVar(&formatArg, "format", "", "")

	weekLong := fs.Bool("week", false, "")
	weekShort := fs.Bool("w", false, "")
	helpLong := fs.Bool("help", false, "")
	helpShort := fs.Bool("h", false, "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if *helpLong || *helpShort {
		printPrintHelp()
		return nil, nil
	}
	if *weekLong || *weekShort {
		opts.mode = "week"
	}
	if opts.numDays < 0 {
		return nil, fmt.Errorf("--days must be >= 0")
	}
	if opts.dedupeRecurring && opts.normalizeRecurring {
		return nil, fmt.Errorf("--dedupe-recurring and --normalize-recurring are mutually exclusive")
	}

	switch {
	case opts.numDays > 0:
	case opts.mode == "week":
		opts.numDays = 7
	case opts.mode == "3day":
		opts.numDays = 3
	default:
		opts.numDays = 2
	}

	if fromArg != "" {
		from, err := parseISODate(fromArg)
		if err != nil {
			return nil, err
		}
		opts.from = from
		opts.hasFrom = true
	}
	if toArg != "" {
		to, err := parseISODate(toArg)
		if err != nil {
			return nil, err
		}
		opts.to = to.AddDate(0, 0, 1)
		opts.hasTo = true
	}
	if opts.hasFrom && opts.hasTo && !opts.from.Before(opts.to) {
		return nil, fmt.Errorf("--from must be on or before --to")
	}

	if calendarArg != "" {
		for _, part := range strings.Split(calendarArg, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				opts.calendarFilters = append(opts.calendarFilters, part)
			}
		}
	}

	if formatArg != "" {
		for _, field := range strings.Split(formatArg, ",") {
			field = strings.TrimSpace(strings.ToLower(field))
			if field == "" {
				continue
			}
			if !isSupportedFormatField(field) {
				return nil, fmt.Errorf("unsupported --format field %q", field)
			}
			opts.formatFields = append(opts.formatFields, field)
		}
	}

	return opts, nil
}

func printPrintHelp() {
	fmt.Println("TipiCal Print - Output agenda to console")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tipical print [flags]")
	fmt.Println("  tipical print agenda [flags]")
	fmt.Println("  tipical print week [flags]")
	fmt.Println("  tipical print 3day [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --week, -w                Print next 7 days")
	fmt.Println("  --days N, -d N            Print next N days")
	fmt.Println("  --calendar NAME[,NAME]    Filter by calendar name")
	fmt.Println("  --search TEXT             Filter summary, description, and location")
	fmt.Println("  --from YYYY-MM-DD         Inclusive start date")
	fmt.Println("  --to YYYY-MM-DD           Inclusive end date")
	fmt.Println("  --format FIELDS           Tabular output fields, e.g. summary,start,end,calendar")
	fmt.Println("  --json, -j                Output matching events as JSON")
	fmt.Println("  --full, -f                Include location and description in agenda view")
	fmt.Println("  --dedupe-recurring        Remove duplicate recurring instances")
	fmt.Println("  --normalize-recurring     Collapse recurring instances by series")
	fmt.Println("  --no-color                Disable colored output")
	fmt.Println("  --help, -h                Show this help")
	fmt.Println()
	fmt.Printf("Exit codes: %d=success, %d=failure, %d=no results\n", exitCodeSuccess, exitCodeFailure, exitCodeNoResults)
	fmt.Println()
}

func filteredEvents(store *ical.Store, opts *printOptions) ([]eventRecord, error) {
	now := time.Now()
	rangeStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	rangeEnd := rangeStart.AddDate(0, 0, opts.numDays)
	if opts.hasFrom {
		rangeStart = opts.from
	}
	if opts.hasTo {
		rangeEnd = opts.to
	}

	var results []eventRecord
	search := strings.ToLower(strings.TrimSpace(opts.search))

	for _, e := range store.Events {
		if e.AllDay {
			if !dateOverlaps(e.Start, e.End, rangeStart, rangeEnd) {
				continue
			}
		} else if !e.OverlapsWith(rangeStart, rangeEnd) {
			continue
		}

		cal := calendarInfoFor(store, e.CalendarID)
		if len(opts.calendarFilters) > 0 && !matchesCalendarFilter(cal, opts.calendarFilters) {
			continue
		}
		if search != "" && !matchesTextFilter(e, search) {
			continue
		}

		results = append(results, eventRecord{
			Event:       e,
			Calendar:    cal,
			Occurrences: 1,
		})
	}

	if opts.dedupeRecurring {
		results = uniqueRecurringEvents(results)
	}
	if opts.normalizeRecurring {
		results = normalizeRecurringEvents(results)
	}

	sortEventRecords(results)
	return results, nil
}

func dateOverlaps(start, end, rangeStart, rangeEnd time.Time) bool {
	sy, sm, sd := start.Date()
	ey, em, ed := end.Date()
	startDate := time.Date(sy, sm, sd, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(ey, em, ed, 0, 0, 0, 0, time.UTC)

	ry, rm, rd := rangeStart.Date()
	rey, rem, red := rangeEnd.Date()
	filterStart := time.Date(ry, rm, rd, 0, 0, 0, 0, time.UTC)
	filterEnd := time.Date(rey, rem, red, 0, 0, 0, 0, time.UTC)
	return startDate.Before(filterEnd) && endDate.After(filterStart)
}

func matchesCalendarFilter(cal ical.CalendarInfo, filters []string) bool {
	name := strings.ToLower(cal.Name)
	source := strings.ToLower(cal.Source)
	for _, filter := range filters {
		f := strings.ToLower(filter)
		if strings.Contains(name, f) || (source != "" && strings.Contains(source, f)) {
			return true
		}
	}
	return false
}

func matchesTextFilter(e *ical.Event, search string) bool {
	return strings.Contains(strings.ToLower(e.Summary), search) ||
		strings.Contains(strings.ToLower(e.Description), search) ||
		strings.Contains(strings.ToLower(e.Location), search)
}

func renderPrintJSON(events []eventRecord, opts *printOptions) string {
	now := time.Now()

	if len(opts.formatFields) > 0 || opts.search != "" || len(opts.calendarFilters) > 0 || opts.hasFrom || opts.hasTo || opts.normalizeRecurring || opts.dedupeRecurring {
		output := jsonOutput{Generated: now.Format(time.RFC3339)}
		for _, rec := range events {
			output.Events = append(output.Events, jsonEventFromRecord(rec))
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		return string(b) + "\n"
	}

	grouped := make(map[string][]jsonEvent)
	var order []string
	for _, rec := range events {
		day := rec.Event.Start.In(time.Local).Format("2006-01-02")
		if _, ok := grouped[day]; !ok {
			order = append(order, day)
		}
		grouped[day] = append(grouped[day], jsonEventFromRecord(rec))
	}
	sort.Strings(order)

	output := jsonOutput{Generated: now.Format(time.RFC3339)}
	for _, day := range order {
		output.Days = append(output.Days, jsonDay{
			Date:   day,
			Events: grouped[day],
		})
	}
	b, _ := json.MarshalIndent(output, "", "  ")
	return string(b) + "\n"
}

func jsonEventFromRecord(rec eventRecord) jsonEvent {
	e := rec.Event
	startStr := e.Start.Format(time.RFC3339)
	endStr := e.End.Format(time.RFC3339)
	if e.AllDay {
		startStr = e.Start.Format("2006-01-02")
		endStr = e.End.Format("2006-01-02")
	}
	return jsonEvent{
		UID:             e.UID,
		Summary:         e.Summary,
		Description:     e.Description,
		Location:        e.Location,
		Start:           startStr,
		End:             endStr,
		AllDay:          e.AllDay,
		Color:           e.Color,
		Status:          e.Status,
		Recurring:       e.Recurring,
		Calendar:        rec.Calendar.Name,
		CalendarSource:  rec.Calendar.Source,
		OccurrenceCount: rec.Occurrences,
	}
}

func renderPrintAgenda(events []eventRecord, theme *config.Theme, use24h bool, full bool) string {
	width := terminalWidth()
	if width < 40 {
		width = 40
	}
	if width > 60 {
		width = 60
	}

	grouped := make(map[string][]eventRecord)
	var order []string
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	tomorrow := today.AddDate(0, 0, 1)

	for _, rec := range events {
		var firstDay time.Time
		if rec.Event.AllDay {
			firstDay = time.Date(rec.Event.Start.Year(), rec.Event.Start.Month(), rec.Event.Start.Day(), 0, 0, 0, 0, time.Local)
		} else {
			firstDay = time.Date(rec.Event.Start.In(time.Local).Year(), rec.Event.Start.In(time.Local).Month(), rec.Event.Start.In(time.Local).Day(), 0, 0, 0, 0, time.Local)
		}
		key := firstDay.Format("2006-01-02")
		if _, ok := grouped[key]; !ok {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], rec)
	}

	sort.Strings(order)
	var lines []string

	for _, key := range order {
		d, _ := time.ParseInLocation("2006-01-02", key, time.Local)

		dayLabel := d.Format("Monday")
		switch {
		case util.SameDay(d, today):
			dayLabel = "Today"
		case util.SameDay(d, tomorrow):
			dayLabel = "Tomorrow"
		}

		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
		if util.SameDay(d, today) {
			headerStyle = headerStyle.Foreground(theme.Today)
		}

		lines = append(lines, headerStyle.Render(fmt.Sprintf("%s - %s", dayLabel, d.Format("January 2"))))
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.Border).Render(strings.Repeat("─", width)))

		dayEvents := grouped[key]
		sortEventRecords(dayEvents)
		for _, rec := range dayEvents {
			e := rec.Event
			color := lipgloss.Color(e.Color)
			if color == "" {
				color = theme.CalendarColor(e.CalendarID)
			}

			dot := lipgloss.NewStyle().Foreground(color).Render("●")
			timeStr := "All day"
			if !e.AllDay {
				timeStr = util.FormatTimeRange(e.Start, e.End, use24h)
			}
			if rec.Occurrences > 1 {
				timeStr += fmt.Sprintf("  x%d", rec.Occurrences)
			}

			timeRendered := lipgloss.NewStyle().Foreground(theme.TextMuted).Width(16).Render(timeStr)
			title := lipgloss.NewStyle().Foreground(theme.Text).Bold(true).Render(e.Summary)
			calendarName := ""
			if rec.Calendar.Name != "" {
				calendarName = lipgloss.NewStyle().Foreground(theme.TextFaint).Render(" [" + rec.Calendar.Name + "]")
			}
			lines = append(lines, fmt.Sprintf("  %s %s %s%s", dot, timeRendered, title, calendarName))

			if full {
				indent := "                      "
				if e.Location != "" {
					lines = append(lines, indent+lipgloss.NewStyle().Foreground(theme.TextFaint).Italic(true).Render(e.Location))
				}
				if e.Description != "" {
					lines = append(lines, indent+lipgloss.NewStyle().Foreground(theme.TextMuted).Render(e.Description))
				}
			}
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func renderFormattedEvents(events []eventRecord, use24h bool, fields []string) string {
	var lines []string
	for _, rec := range events {
		var values []string
		for _, field := range fields {
			values = append(values, formatFieldValue(rec, field, use24h))
		}
		lines = append(lines, strings.Join(values, "\t"))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func isSupportedFormatField(field string) bool {
	switch field {
	case "uid", "summary", "description", "location", "start", "end", "date", "time", "calendar", "source", "status", "all_day", "recurring", "occurrences":
		return true
	default:
		return false
	}
}

func formatFieldValue(rec eventRecord, field string, use24h bool) string {
	e := rec.Event
	switch field {
	case "uid":
		return e.UID
	case "summary":
		return e.Summary
	case "description":
		return strings.ReplaceAll(e.Description, "\n", " ")
	case "location":
		return e.Location
	case "start":
		if e.AllDay {
			return e.Start.Format("2006-01-02")
		}
		return e.Start.Format(time.RFC3339)
	case "end":
		if e.AllDay {
			return e.End.Format("2006-01-02")
		}
		return e.End.Format(time.RFC3339)
	case "date":
		return e.Start.Format("2006-01-02")
	case "time":
		if e.AllDay {
			return "all-day"
		}
		return util.FormatTimeRange(e.Start, e.End, use24h)
	case "calendar":
		return rec.Calendar.Name
	case "source":
		return rec.Calendar.Source
	case "status":
		return e.Status
	case "all_day":
		if e.AllDay {
			return "true"
		}
		return "false"
	case "recurring":
		if e.Recurring {
			return "true"
		}
		return "false"
	case "occurrences":
		return fmt.Sprintf("%d", rec.Occurrences)
	default:
		return ""
	}
}

func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 {
		return 60
	}
	return width
}
