package caldav

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/terminal-ical/terminal-ical/config"
)

// Client wraps the CalDAV client with convenience methods.
type Client struct {
	cfg      *config.CalendarConfig
	client   *caldav.Client
	calIndex int
}

// NewClient creates a CalDAV client for the given calendar config.
func NewClient(calCfg *config.CalendarConfig, index int) (*Client, error) {
	password, err := resolvePassword(calCfg)
	if err != nil {
		return nil, fmt.Errorf("resolving password: %w", err)
	}

	httpClient := webdav.HTTPClientWithBasicAuth(&http.Client{
		Timeout: 30 * time.Second,
	}, calCfg.Username, password)

	client, err := caldav.NewClient(httpClient, calCfg.URL)
	if err != nil {
		return nil, fmt.Errorf("creating caldav client: %w", err)
	}

	return &Client{
		cfg:      calCfg,
		client:   client,
		calIndex: index,
	}, nil
}

// DiscoverCalendars finds all calendars accessible on this server.
func (c *Client) DiscoverCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	principal, err := c.client.FindCurrentUserPrincipal(ctx)
	if err != nil {
		// Fall back to empty principal
		principal = ""
	}

	homeSet, err := c.client.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return nil, fmt.Errorf("finding calendar home set: %w", err)
	}

	calendars, err := c.client.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("finding calendars: %w", err)
	}

	return calendars, nil
}

// FetchEvents fetches events in the given time range from the specified calendar path.
func (c *Client) FetchEvents(ctx context.Context, calPath string, start, end time.Time) ([]caldav.CalendarObject, error) {
	query := &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{{
				Name:  "VEVENT",
				Start: start,
				End:   end,
			}},
		},
	}

	objects, err := c.client.QueryCalendar(ctx, calPath, query)
	if err != nil {
		return nil, fmt.Errorf("querying calendar: %w", err)
	}

	return objects, nil
}

// PutEvent creates or updates an event on the server.
func (c *Client) PutEvent(ctx context.Context, path string, obj *caldav.CalendarObject) (*caldav.CalendarObject, error) {
	result, err := c.client.PutCalendarObject(ctx, path, obj.Data)
	if err != nil {
		return nil, fmt.Errorf("putting calendar object: %w", err)
	}
	return result, nil
}

// DeleteEvent removes an event from the server.
func (c *Client) DeleteEvent(ctx context.Context, path string) error {
	err := c.client.RemoveAll(ctx, path)
	if err != nil {
		return fmt.Errorf("deleting calendar object: %w", err)
	}
	return nil
}

// CalendarIndex returns the index of this calendar in the config.
func (c *Client) CalendarIndex() int {
	return c.calIndex
}

// resolvePassword gets the password either directly or via command.
func resolvePassword(calCfg *config.CalendarConfig) (string, error) {
	if calCfg.Password != "" {
		return calCfg.Password, nil
	}

	if calCfg.PasswordCmd != "" {
		parts := strings.Fields(calCfg.PasswordCmd)
		if len(parts) == 0 {
			return "", fmt.Errorf("empty password_cmd")
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("running password_cmd %q: %w", calCfg.PasswordCmd, err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	return "", nil
}
