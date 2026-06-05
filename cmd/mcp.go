package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// runMCP starts an MCP (Model Context Protocol) server over stdio, exposing
// TipiCal's core calendar operations as tools an MCP client can call.
func runMCP(args []string) int {
	for _, a := range args {
		switch a {
		case "-h", "--help", "help":
			printMCPHelp()
			return exitCodeSuccess
		}
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "tipical",
		Title:   "TipiCal Calendar",
		Version: Version,
	}, nil)

	registerMCPTools(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		return exitCodeFailure
	}
	return exitCodeSuccess
}

func printMCPHelp() {
	fmt.Println("TipiCal MCP - Model Context Protocol server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tipical mcp")
	fmt.Println()
	fmt.Println("Starts an MCP server over stdio so an MCP client (e.g. Claude) can")
	fmt.Println("drive TipiCal's calendar operations as tools. It reads the same config")
	fmt.Println("(~/.config/tipical/config.toml) and CalDAV credentials as the other")
	fmt.Println("commands. Run it as a stdio MCP server from your client configuration.")
	fmt.Println()
	fmt.Println("Tools: list_calendars, list_events, get_event, create_event,")
	fmt.Println("       update_event, delete_event")
	fmt.Println()
}

// --- Tool input/output types ---

type mcpEvent struct {
	UID         string `json:"uid"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	Start       string `json:"start"`
	End         string `json:"end"`
	AllDay      bool   `json:"all_day"`
	Status      string `json:"status,omitempty"`
	Recurring   bool   `json:"recurring,omitempty"`
	Calendar    string `json:"calendar,omitempty"`
	Source      string `json:"source,omitempty"`
}

func mcpEventFrom(rec eventRecord) mcpEvent {
	e := rec.Event
	startStr := e.Start.Format(time.RFC3339)
	endStr := e.End.Format(time.RFC3339)
	if e.AllDay {
		startStr = e.Start.Format("2006-01-02")
		endStr = e.End.Format("2006-01-02")
	}
	return mcpEvent{
		UID:         e.UID,
		Summary:     e.Summary,
		Description: e.Description,
		Location:    e.Location,
		Start:       startStr,
		End:         endStr,
		AllDay:      e.AllDay,
		Status:      e.Status,
		Recurring:   e.Recurring,
		Calendar:    rec.Calendar.Name,
		Source:      rec.Calendar.Source,
	}
}

type listCalendarsInput struct {
	Search string `json:"search,omitempty" jsonschema:"optional case-insensitive filter on calendar name or source"`
}

type mcpCalendar struct {
	Name     string `json:"name"`
	Source   string `json:"source,omitempty"`
	Color    string `json:"color,omitempty"`
	ReadOnly bool   `json:"readonly"`
}

type listCalendarsOutput struct {
	Calendars []mcpCalendar `json:"calendars"`
}

type listEventsInput struct {
	From     string `json:"from,omitempty" jsonschema:"inclusive start date in YYYY-MM-DD; defaults to today"`
	To       string `json:"to,omitempty" jsonschema:"inclusive end date in YYYY-MM-DD; defaults to 14 days after from"`
	Calendar string `json:"calendar,omitempty" jsonschema:"optional calendar name filter (substring match)"`
	Search   string `json:"search,omitempty" jsonschema:"optional case-insensitive text filter on summary, description, and location"`
}

type listEventsOutput struct {
	Events []mcpEvent `json:"events"`
}

type getEventInput struct {
	UID string `json:"uid" jsonschema:"the unique identifier (UID) of the event"`
}

type createEventInput struct {
	Calendar    string `json:"calendar" jsonschema:"target calendar name (must not be read-only)"`
	Summary     string `json:"summary" jsonschema:"event title"`
	Date        string `json:"date" jsonschema:"event date in YYYY-MM-DD"`
	Start       string `json:"start,omitempty" jsonschema:"start time HH:MM (24h); defaults to 09:00, ignored when all_day"`
	End         string `json:"end,omitempty" jsonschema:"end time HH:MM (24h); defaults to 10:00, ignored when all_day"`
	AllDay      bool   `json:"all_day,omitempty" jsonschema:"set true for an all-day event"`
	Description string `json:"description,omitempty" jsonschema:"optional event description"`
	Location    string `json:"location,omitempty" jsonschema:"optional event location"`
}

type updateEventInput struct {
	UID         string `json:"uid" jsonschema:"the UID of the event to update"`
	Calendar    string `json:"calendar,omitempty" jsonschema:"move event to this calendar (must not be read-only)"`
	Summary     string `json:"summary,omitempty" jsonschema:"new title"`
	Description string `json:"description,omitempty" jsonschema:"new description"`
	Location    string `json:"location,omitempty" jsonschema:"new location"`
	Status      string `json:"status,omitempty" jsonschema:"new status, e.g. CONFIRMED, TENTATIVE, CANCELLED"`
	Date        string `json:"date,omitempty" jsonschema:"new date YYYY-MM-DD (preserves duration)"`
	Start       string `json:"start,omitempty" jsonschema:"new start time HH:MM (24h)"`
	End         string `json:"end,omitempty" jsonschema:"new end time HH:MM (24h)"`
}

type deleteEventInput struct {
	UID string `json:"uid" jsonschema:"the UID of the event to delete"`
}

type mutationOutput struct {
	Event mcpEvent `json:"event"`
}

type deleteOutput struct {
	Deleted string `json:"deleted"`
}

// registerMCPTools wires all TipiCal tools onto the server.
func registerMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_calendars",
		Description: "List the calendars configured in TipiCal (from CalDAV or demo data).",
	}, handleListCalendars)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_events",
		Description: "List or query events within a date range, optionally filtered by calendar and text. Defaults to the next 14 days from today.",
	}, handleListEvents)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_event",
		Description: "Get the full details of a single event by its UID.",
	}, handleGetEvent)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_event",
		Description: "Create a new event on a writable CalDAV calendar.",
	}, handleCreateEvent)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_event",
		Description: "Update fields of an existing event identified by UID, and persist via CalDAV.",
	}, handleUpdateEvent)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_event",
		Description: "Delete an event by UID from its CalDAV calendar. This is irreversible.",
	}, handleDeleteEvent)
}

// loadMCPContext loads config + CalDAV data the same way the CLI commands do.
// When requireSync is true it errors if no CalDAV calendars are configured.
func loadMCPContext(requireSync bool) (*cliContext, error) {
	ctx, err := loadCLIContext(true)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	if err := ensureRemoteData(ctx); err != nil && len(ctx.store.Events) == 0 && len(ctx.store.Calendars) == 0 {
		return nil, fmt.Errorf("loading calendars: %w", err)
	}
	if requireSync && ctx.syncMgr == nil {
		return nil, fmt.Errorf("no CalDAV calendars configured; mutating operations are unavailable")
	}
	return ctx, nil
}

func handleListCalendars(_ context.Context, _ *mcp.CallToolRequest, in listCalendarsInput) (*mcp.CallToolResult, listCalendarsOutput, error) {
	ctx, err := loadMCPContext(false)
	if err != nil {
		return nil, listCalendarsOutput{}, err
	}

	search := strings.ToLower(strings.TrimSpace(in.Search))
	out := listCalendarsOutput{Calendars: []mcpCalendar{}}
	for _, cal := range ctx.store.Calendars {
		if search != "" &&
			!strings.Contains(strings.ToLower(cal.Name), search) &&
			!strings.Contains(strings.ToLower(cal.Source), search) {
			continue
		}
		out.Calendars = append(out.Calendars, mcpCalendar{
			Name:     cal.Name,
			Source:   cal.Source,
			Color:    cal.Color,
			ReadOnly: cal.ReadOnly,
		})
	}
	return nil, out, nil
}

func handleListEvents(_ context.Context, _ *mcp.CallToolRequest, in listEventsInput) (*mcp.CallToolResult, listEventsOutput, error) {
	ctx, err := loadMCPContext(false)
	if err != nil {
		return nil, listEventsOutput{}, err
	}

	opts := &printOptions{mode: "agenda", numDays: 14}
	if in.From != "" {
		from, err := parseISODate(in.From)
		if err != nil {
			return nil, listEventsOutput{}, err
		}
		opts.from = from
		opts.hasFrom = true
	}
	if in.To != "" {
		to, err := parseISODate(in.To)
		if err != nil {
			return nil, listEventsOutput{}, err
		}
		opts.to = to.AddDate(0, 0, 1)
		opts.hasTo = true
	}
	if opts.hasFrom && opts.hasTo && !opts.from.Before(opts.to) {
		return nil, listEventsOutput{}, fmt.Errorf("from must be on or before to")
	}
	if c := strings.TrimSpace(in.Calendar); c != "" {
		opts.calendarFilters = []string{c}
	}
	opts.search = in.Search

	records, err := filteredEvents(ctx.store, opts)
	if err != nil {
		return nil, listEventsOutput{}, err
	}

	out := listEventsOutput{Events: []mcpEvent{}}
	for _, rec := range records {
		out.Events = append(out.Events, mcpEventFrom(rec))
	}
	return nil, out, nil
}

func handleGetEvent(_ context.Context, _ *mcp.CallToolRequest, in getEventInput) (*mcp.CallToolResult, mcpEvent, error) {
	if strings.TrimSpace(in.UID) == "" {
		return nil, mcpEvent{}, fmt.Errorf("uid is required")
	}
	ctx, err := loadMCPContext(false)
	if err != nil {
		return nil, mcpEvent{}, err
	}
	e := ctx.store.FindEvent(in.UID)
	if e == nil {
		return nil, mcpEvent{}, fmt.Errorf("event %q not found", in.UID)
	}
	rec := eventRecord{Event: e, Calendar: calendarInfoFor(ctx.store, e.CalendarID), Occurrences: 1}
	return nil, mcpEventFrom(rec), nil
}

func handleCreateEvent(_ context.Context, _ *mcp.CallToolRequest, in createEventInput) (*mcp.CallToolResult, mutationOutput, error) {
	ctx, err := loadMCPContext(true)
	if err != nil {
		return nil, mutationOutput{}, err
	}

	event, err := buildCreatedEvent(ctx.store, &eventMutationOptions{
		calendar:    in.Calendar,
		summary:     in.Summary,
		description: in.Description,
		location:    in.Location,
		date:        in.Date,
		start:       in.Start,
		end:         in.End,
		allDay:      in.AllDay,
	})
	if err != nil {
		return nil, mutationOutput{}, err
	}

	callCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ctx.syncMgr.CreateEvent(callCtx, event); err != nil {
		return nil, mutationOutput{}, fmt.Errorf("create failed: %w", err)
	}

	rec := eventRecord{Event: event, Calendar: calendarInfoFor(ctx.store, event.CalendarID), Occurrences: 1}
	return nil, mutationOutput{Event: mcpEventFrom(rec)}, nil
}

func handleUpdateEvent(_ context.Context, _ *mcp.CallToolRequest, in updateEventInput) (*mcp.CallToolResult, mutationOutput, error) {
	if strings.TrimSpace(in.UID) == "" {
		return nil, mutationOutput{}, fmt.Errorf("uid is required")
	}
	ctx, err := loadMCPContext(true)
	if err != nil {
		return nil, mutationOutput{}, err
	}

	original := ctx.store.FindEvent(in.UID)
	if original == nil {
		return nil, mutationOutput{}, fmt.Errorf("event %q not found", in.UID)
	}

	updated, err := mutateEvent(ctx.store, original, &eventMutationOptions{
		calendar:    in.Calendar,
		summary:     in.Summary,
		description: in.Description,
		location:    in.Location,
		status:      in.Status,
		date:        in.Date,
		start:       in.Start,
		end:         in.End,
	}, false)
	if err != nil {
		return nil, mutationOutput{}, err
	}

	callCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ctx.syncMgr.UpdateEvent(callCtx, updated); err != nil {
		return nil, mutationOutput{}, fmt.Errorf("update failed: %w", err)
	}

	rec := eventRecord{Event: updated, Calendar: calendarInfoFor(ctx.store, updated.CalendarID), Occurrences: 1}
	return nil, mutationOutput{Event: mcpEventFrom(rec)}, nil
}

func handleDeleteEvent(_ context.Context, _ *mcp.CallToolRequest, in deleteEventInput) (*mcp.CallToolResult, deleteOutput, error) {
	if strings.TrimSpace(in.UID) == "" {
		return nil, deleteOutput{}, fmt.Errorf("uid is required")
	}
	ctx, err := loadMCPContext(true)
	if err != nil {
		return nil, deleteOutput{}, err
	}

	event := ctx.store.FindEvent(in.UID)
	if event == nil {
		return nil, deleteOutput{}, fmt.Errorf("event %q not found", in.UID)
	}

	callCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ctx.syncMgr.DeleteEvent(callCtx, event); err != nil {
		return nil, deleteOutput{}, fmt.Errorf("delete failed: %w", err)
	}
	return nil, deleteOutput{Deleted: event.UID}, nil
}
