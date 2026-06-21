package mcpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/tipical/tipical/ical"
)

// eventJSON is the wire shape for an event returned by tools.
type eventJSON struct {
	UID         string `json:"uid"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	Calendar    string `json:"calendar"`
	CalendarID  int    `json:"calendar_id"`
	Start       string `json:"start"`
	End         string `json:"end"`
	AllDay      bool   `json:"all_day"`
	Status      string `json:"status,omitempty"`
	Recurring   bool   `json:"recurring,omitempty"`
	Recurrence  string `json:"recurrence,omitempty"`
}

const (
	dateLayout     = "2006-01-02"
	dateTimeLayout = "2006-01-02 15:04"
)

func (s *Server) toEventJSON(e *ical.Event) eventJSON {
	startLayout, endLayout := dateTimeLayout, dateTimeLayout
	if e.AllDay {
		startLayout, endLayout = dateLayout, dateLayout
	}
	return eventJSON{
		UID:         e.UID,
		Summary:     e.Summary,
		Description: e.Description,
		Location:    e.Location,
		Calendar:    s.store.CalendarName(e.CalendarID),
		CalendarID:  e.CalendarID,
		Start:       e.Start.Format(startLayout),
		End:         e.End.Format(endLayout),
		AllDay:      e.AllDay,
		Status:      e.Status,
		Recurring:   e.Recurring,
		Recurrence:  e.RecurrenceDescription(),
	}
}

// eventListResult wraps a slice in an object; the MCP spec requires a tool's
// structuredContent to be a JSON object, not a bare array.
type eventListResult struct {
	Events []eventJSON `json:"events"`
	Count  int         `json:"count"`
}

func (s *Server) eventListJSON(events []*ical.Event) (*mcp.CallToolResult, error) {
	out := make([]eventJSON, 0, len(events))
	for _, e := range events {
		out = append(out, s.toEventJSON(e))
	}
	return mcp.NewToolResultJSON(eventListResult{Events: out, Count: len(out)})
}

func limitArg(req mcp.CallToolRequest, def int) int {
	n := req.GetInt("limit", def)
	if n <= 0 {
		return def
	}
	return n
}

// resolveCalendar maps a calendar name or numeric id to a registered id.
func (s *Server) resolveCalendar(nameOrID string) (int, error) {
	nameOrID = strings.TrimSpace(nameOrID)
	if nameOrID == "" {
		if len(s.store.Calendars) == 1 {
			return s.store.Calendars[0].ID, nil
		}
		return -1, fmt.Errorf("calendar is required (%s)", s.calendarChoices())
	}
	if id, err := strconv.Atoi(nameOrID); err == nil {
		if id >= 0 && id < len(s.store.Calendars) {
			return id, nil
		}
	}
	for _, c := range s.store.Calendars {
		if c.Name == nameOrID {
			return c.ID, nil
		}
	}
	return -1, fmt.Errorf("unknown calendar %q (%s)", nameOrID, s.calendarChoices())
}

func (s *Server) calendarChoices() string {
	if len(s.store.Calendars) == 0 {
		return "no calendars synced"
	}
	parts := make([]string, 0, len(s.store.Calendars))
	for _, c := range s.store.Calendars {
		parts = append(parts, fmt.Sprintf("%d=%s", c.ID, c.Name))
	}
	return "available: " + strings.Join(parts, ", ")
}

// parseWhen parses a date or date-time in the local timezone, reporting whether
// the input was a bare date.
func parseWhen(s string) (time.Time, bool, error) {
	if t, err := time.ParseInLocation(dateLayout, s, time.Local); err == nil {
		return t, true, nil
	}
	for _, layout := range []string{dateTimeLayout, "2006-01-02T15:04", "2006-01-02 15:04:05", time.RFC3339} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, false, nil
		}
	}
	return time.Time{}, false, fmt.Errorf("invalid date/time %q (use YYYY-MM-DD or \"YYYY-MM-DD HH:MM\")", s)
}

func newEventUID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("tical-%d-%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

func sortByStart(events []*ical.Event) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].Start.Before(events[j].Start)
	})
}

// --- read handlers ---

func (s *Server) handleListCalendars(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type calJSON struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Source   string `json:"source,omitempty"`
		Color    string `json:"color,omitempty"`
		ReadOnly bool   `json:"read_only"`
	}
	out := make([]calJSON, 0, len(s.store.Calendars))
	for _, c := range s.store.Calendars {
		out = append(out, calJSON{
			ID:       c.ID,
			Name:     c.Name,
			Source:   c.Source,
			Color:    c.Color,
			ReadOnly: c.ReadOnly,
		})
	}
	return mcp.NewToolResultJSON(struct {
		Calendars []calJSON `json:"calendars"`
		Count     int       `json:"count"`
	}{Calendars: out, Count: len(out)})
}

func (s *Server) handleListEvents(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if v := strings.TrimSpace(req.GetString("from", "")); v != "" {
		t, _, err := parseWhen(v)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("from", err), nil
		}
		start = t
	}

	days := req.GetInt("days", 7)
	if days <= 0 {
		days = 7
	}
	end := start.AddDate(0, 0, days)
	if v := strings.TrimSpace(req.GetString("to", "")); v != "" {
		t, _, err := parseWhen(v)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("to", err), nil
		}
		end = t
	}

	calFilter := -1
	if v := strings.TrimSpace(req.GetString("calendar", "")); v != "" {
		id, err := s.resolveCalendar(v)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("calendar", err), nil
		}
		calFilter = id
	}

	limit := limitArg(req, defaultLimit)
	matches := make([]*ical.Event, 0)
	for _, e := range s.store.EventsInRange(start, end) {
		if calFilter >= 0 && e.CalendarID != calFilter {
			continue
		}
		matches = append(matches, e)
	}
	sortByStart(matches)
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return s.eventListJSON(matches)
}

func (s *Server) handleSearchEvents(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("query", err), nil
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return mcp.NewToolResultError("query must not be empty"), nil
	}

	var (
		start, end time.Time
		haveStart  bool
		haveEnd    bool
	)
	if v := strings.TrimSpace(req.GetString("from", "")); v != "" {
		t, _, perr := parseWhen(v)
		if perr != nil {
			return mcp.NewToolResultErrorFromErr("from", perr), nil
		}
		start, haveStart = t, true
	}
	if v := strings.TrimSpace(req.GetString("to", "")); v != "" {
		t, _, perr := parseWhen(v)
		if perr != nil {
			return mcp.NewToolResultErrorFromErr("to", perr), nil
		}
		end, haveEnd = t, true
	}

	calFilter := -1
	if v := strings.TrimSpace(req.GetString("calendar", "")); v != "" {
		id, cerr := s.resolveCalendar(v)
		if cerr != nil {
			return mcp.NewToolResultErrorFromErr("calendar", cerr), nil
		}
		calFilter = id
	}

	limit := limitArg(req, defaultLimit)
	matches := make([]*ical.Event, 0)
	for _, e := range s.store.Events {
		if calFilter >= 0 && e.CalendarID != calFilter {
			continue
		}
		if haveStart && !e.End.After(start) {
			continue
		}
		if haveEnd && !e.Start.Before(end) {
			continue
		}
		hay := strings.ToLower(e.Summary + "\n" + e.Location + "\n" + e.Description)
		if !strings.Contains(hay, q) {
			continue
		}
		matches = append(matches, e)
	}
	sortByStart(matches)
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return s.eventListJSON(matches)
}

func (s *Server) handleGetEvent(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uid, err := req.RequireString("uid")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("uid", err), nil
	}
	e := s.store.FindEvent(uid)
	if e == nil {
		return mcp.NewToolResultError(fmt.Sprintf("no event with uid %q", uid)), nil
	}
	return mcp.NewToolResultJSON(s.toEventJSON(e))
}

// --- write handlers ---

func (s *Server) handleCreateEvent(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.requireSync(); err != nil {
		return mcp.NewToolResultErrorFromErr("create_event", err), nil
	}
	calArg, err := req.RequireString("calendar")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("calendar", err), nil
	}
	calID, err := s.resolveCalendar(calArg)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("calendar", err), nil
	}
	if s.store.Calendars[calID].ReadOnly {
		return mcp.NewToolResultError(fmt.Sprintf("calendar %q is read-only", s.store.CalendarName(calID))), nil
	}
	title, err := req.RequireString("title")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("title", err), nil
	}
	startStr, err := req.RequireString("start")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("start", err), nil
	}
	startTime, isDate, err := parseWhen(startStr)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("start", err), nil
	}
	allDay := req.GetBool("all_day", isDate)

	var endTime time.Time
	if v := strings.TrimSpace(req.GetString("end", "")); v != "" {
		t, _, perr := parseWhen(v)
		if perr != nil {
			return mcp.NewToolResultErrorFromErr("end", perr), nil
		}
		endTime = t
	} else if allDay {
		endTime = startTime.AddDate(0, 0, 1)
	} else {
		endTime = startTime.Add(time.Hour)
	}

	event := &ical.Event{
		UID:         newEventUID(),
		Summary:     title,
		Description: strings.TrimSpace(req.GetString("description", "")),
		Location:    strings.TrimSpace(req.GetString("location", "")),
		Start:       startTime,
		End:         endTime,
		AllDay:      allDay,
		CalendarID:  calID,
		Status:      "CONFIRMED",
		CalPath:     s.store.Calendars[calID].CalPath,
	}

	s.store.AddEvent(event)
	ctx, cancel := writeCtx()
	defer cancel()
	if err := s.sync.CreateEvent(ctx, event); err != nil {
		s.store.RemoveEvent(event.UID)
		return mcp.NewToolResultErrorFromErr("creating event", err), nil
	}
	return mcp.NewToolResultJSON(s.toEventJSON(event))
}

func (s *Server) handleUpdateEvent(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.requireSync(); err != nil {
		return mcp.NewToolResultErrorFromErr("update_event", err), nil
	}
	uid, err := req.RequireString("uid")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("uid", err), nil
	}
	orig := s.store.FindEvent(uid)
	if orig == nil {
		return mcp.NewToolResultError(fmt.Sprintf("no event with uid %q", uid)), nil
	}
	if orig.CalendarID >= 0 && orig.CalendarID < len(s.store.Calendars) && s.store.Calendars[orig.CalendarID].ReadOnly {
		return mcp.NewToolResultError(fmt.Sprintf("calendar %q is read-only", s.store.CalendarName(orig.CalendarID))), nil
	}

	// Copy so a failed sync doesn't leave the store mutated.
	updated := *orig

	if v, ok := optionalString(req, "title"); ok {
		updated.Summary = v
	}
	if v, ok := optionalString(req, "location"); ok {
		updated.Location = v
	}
	if v, ok := optionalString(req, "description"); ok {
		updated.Description = v
	}
	if v, ok := optionalString(req, "start"); ok {
		t, isDate, perr := parseWhen(v)
		if perr != nil {
			return mcp.NewToolResultErrorFromErr("start", perr), nil
		}
		updated.Start = t
		updated.AllDay = isDate
	}
	if v, ok := optionalString(req, "end"); ok {
		t, _, perr := parseWhen(v)
		if perr != nil {
			return mcp.NewToolResultErrorFromErr("end", perr), nil
		}
		updated.End = t
	}
	if req.GetArguments()["all_day"] != nil {
		updated.AllDay = req.GetBool("all_day", updated.AllDay)
	}

	s.store.RemoveEvent(uid)
	s.store.AddEvent(&updated)
	ctx, cancel := writeCtx()
	defer cancel()
	if err := s.sync.UpdateEvent(ctx, &updated); err != nil {
		// Restore the original on failure.
		s.store.RemoveEvent(uid)
		s.store.AddEvent(orig)
		return mcp.NewToolResultErrorFromErr("updating event", err), nil
	}
	return mcp.NewToolResultJSON(s.toEventJSON(&updated))
}

func (s *Server) handleDeleteEvent(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.requireSync(); err != nil {
		return mcp.NewToolResultErrorFromErr("delete_event", err), nil
	}
	uid, err := req.RequireString("uid")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("uid", err), nil
	}
	event := s.store.FindEvent(uid)
	if event == nil {
		return mcp.NewToolResultError(fmt.Sprintf("no event with uid %q", uid)), nil
	}
	if event.CalendarID >= 0 && event.CalendarID < len(s.store.Calendars) && s.store.Calendars[event.CalendarID].ReadOnly {
		return mcp.NewToolResultError(fmt.Sprintf("calendar %q is read-only", s.store.CalendarName(event.CalendarID))), nil
	}

	ctx, cancel := writeCtx()
	defer cancel()
	if err := s.sync.DeleteEvent(ctx, event); err != nil {
		return mcp.NewToolResultErrorFromErr("deleting event", err), nil
	}
	s.store.RemoveEvent(uid)
	return mcp.NewToolResultText(fmt.Sprintf("Deleted event %q (%s).", event.Summary, uid)), nil
}

// optionalString returns the trimmed value of an argument and whether it was
// present in the request at all (so callers can distinguish "unset" from
// "explicitly empty").
func optionalString(req mcp.CallToolRequest, key string) (string, bool) {
	args := req.GetArguments()
	if _, ok := args[key]; !ok {
		return "", false
	}
	return strings.TrimSpace(req.GetString(key, "")), true
}
