// Package mcpserver exposes TipiCal's calendars over the Model Context Protocol
// so that LLM agents can list, search, create, update, and delete events
// through the same CalDAV sync layer the TUI uses.
//
// The server speaks JSON-RPC over stdio; stdout is reserved for the protocol,
// so nothing here may print to stdout. Read tools serve from the in-memory
// store populated at startup (cache + a best-effort live sync); write tools
// push to the CalDAV server on demand and mirror the change back into the
// store and cache, matching the TUI's behaviour.
package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/tipical/tipical/caldav"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
)

const (
	serverName    = "tipical"
	serverVersion = "0.1.0"
	defaultLimit  = 100
)

// Server wires the event store and CalDAV sync manager into MCP tool handlers.
type Server struct {
	cfg   *config.Config
	store *ical.Store
	sync  *caldav.Sync
}

// New creates a Server backed by the given config, store, and sync manager. The
// sync manager may be nil when no calendars are configured, in which case write
// tools return an error and read tools serve whatever is in the store.
func New(cfg *config.Config, store *ical.Store, sync *caldav.Sync) *Server {
	return &Server{cfg: cfg, store: store, sync: sync}
}

// Serve builds the MCP server and serves it over stdio until stdin closes.
func Serve(cfg *config.Config, store *ical.Store, sync *caldav.Sync) error {
	return server.ServeStdio(New(cfg, store, sync).MCPServer())
}

// MCPServer constructs the MCP server with all tools registered.
func (s *Server) MCPServer() *server.MCPServer {
	srv := server.NewMCPServer(
		serverName, serverVersion,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
		server.WithInstructions(
			"TipiCal calendar access. Read tools (list_calendars, list_events, "+
				"search_events, get_event) serve from a local store and are safe to call "+
				"freely. Write tools (create_event, update_event, delete_event) change the "+
				"real CalDAV calendars; confirm intent with the user before calling them. "+
				"Dates use the format YYYY-MM-DD for all-day boundaries or "+
				"\"YYYY-MM-DD HH:MM\" for timed events, interpreted in the local timezone.",
		),
	)

	s.registerReadTools(srv)
	s.registerWriteTools(srv)
	return srv
}

func (s *Server) registerReadTools(srv *server.MCPServer) {
	srv.AddTool(mcp.NewTool("list_calendars",
		mcp.WithDescription("List the synced calendars with their id, name, source account, "+
			"color, and read-only status. Use a calendar id or name with create_event."),
	), s.handleListCalendars)

	srv.AddTool(mcp.NewTool("list_events",
		mcp.WithDescription("List events in a date range, soonest first. Defaults to the next "+
			"7 days starting today."),
		mcp.WithString("from", mcp.Description("Start of range (YYYY-MM-DD or \"YYYY-MM-DD HH:MM\"). Defaults to now.")),
		mcp.WithString("to", mcp.Description("End of range (exclusive). Defaults to 'from' plus 'days'.")),
		mcp.WithNumber("days", mcp.Description("Number of days from 'from' when 'to' is omitted (default 7).")),
		mcp.WithString("calendar", mcp.Description("Restrict to this calendar (id or name).")),
		mcp.WithNumber("limit", mcp.Description("Max events to return (default 100).")),
	), s.handleListEvents)

	srv.AddTool(mcp.NewTool("search_events",
		mcp.WithDescription("Case-insensitive substring search over event summary, location, "+
			"and description. Optionally bounded by a date range."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Text to search for.")),
		mcp.WithString("from", mcp.Description("Restrict to events ending on or after this date.")),
		mcp.WithString("to", mcp.Description("Restrict to events starting before this date.")),
		mcp.WithString("calendar", mcp.Description("Restrict to this calendar (id or name).")),
		mcp.WithNumber("limit", mcp.Description("Max results to return (default 100).")),
	), s.handleSearchEvents)

	srv.AddTool(mcp.NewTool("get_event",
		mcp.WithDescription("Get a single event by its UID, including all details."),
		mcp.WithString("uid", mcp.Required(), mcp.Description("Event UID.")),
	), s.handleGetEvent)
}

func (s *Server) registerWriteTools(srv *server.MCPServer) {
	srv.AddTool(mcp.NewTool("create_event",
		mcp.WithDescription("Create an event on a CalDAV calendar."),
		mcp.WithString("calendar", mcp.Required(), mcp.Description("Target calendar (id or name).")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Event summary/title.")),
		mcp.WithString("start", mcp.Required(), mcp.Description("Start (YYYY-MM-DD for all-day, or \"YYYY-MM-DD HH:MM\").")),
		mcp.WithString("end", mcp.Description("End (exclusive). Defaults to +1h for timed events, +1 day for all-day.")),
		mcp.WithBoolean("all_day", mcp.Description("Force an all-day event regardless of 'start' format.")),
		mcp.WithString("location", mcp.Description("Event location.")),
		mcp.WithString("description", mcp.Description("Event description/notes.")),
	), s.handleCreateEvent)

	srv.AddTool(mcp.NewTool("update_event",
		mcp.WithDescription("Update an existing event identified by UID. Only the provided "+
			"fields are changed."),
		mcp.WithString("uid", mcp.Required(), mcp.Description("Event UID.")),
		mcp.WithString("title", mcp.Description("New summary/title.")),
		mcp.WithString("start", mcp.Description("New start.")),
		mcp.WithString("end", mcp.Description("New end (exclusive).")),
		mcp.WithBoolean("all_day", mcp.Description("Set the all-day flag.")),
		mcp.WithString("location", mcp.Description("New location.")),
		mcp.WithString("description", mcp.Description("New description/notes.")),
	), s.handleUpdateEvent)

	srv.AddTool(mcp.NewTool("delete_event",
		mcp.WithDescription("Delete an event identified by UID from its CalDAV calendar."),
		mcp.WithString("uid", mcp.Required(), mcp.Description("Event UID.")),
	), s.handleDeleteEvent)
}

// writeCtx returns a context with a timeout suitable for CalDAV writes.
func writeCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// requireSync returns the sync manager or an error result if none is available.
func (s *Server) requireSync() error {
	if s.sync == nil {
		return fmt.Errorf("no CalDAV calendars configured; run 'tipical auth add' first")
	}
	return nil
}
