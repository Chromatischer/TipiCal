package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/caldav"
	"github.com/tipical/tipical/config"
	"github.com/tipical/tipical/ical"
	"github.com/tipical/tipical/ui/components"
	"github.com/tipical/tipical/ui/editor"
	"github.com/tipical/tipical/ui/views"
)

// CalendarView is the interface all views must implement.
type CalendarView interface {
	View() string
	SetSize(w, h int)
	SelectedDate() time.Time
	SetDate(d time.Time)
	NextPeriod()
	PrevPeriod()
	MoveUp()
	MoveDown()
	MoveLeft()
	MoveRight()
}

// syncMsg triggers a sync operation.
type syncMsg struct{}

// syncResultMsg carries the result of a sync operation.
type syncResultMsg struct {
	err error
}

// upsyncResultMsg carries the result of a create/update/delete upsync operation.
type upsyncResultMsg struct {
	action string // "create", "update", "delete"
	err    error
}

// App is the root Bubble Tea model.
type App struct {
	cfg    *config.Config
	theme  *config.Theme
	styles *Styles
	store  *ical.Store

	// Sync
	syncMgr *caldav.Sync

	// Components
	header      *Header
	sidebar     *Sidebar
	statusbar   *StatusBar
	eventEditor *editor.EventEditor
	search      *components.Search

	// Views
	viewType     config.ViewType
	monthView    *views.MonthView
	weekView     *views.WeekView
	threeDayView *views.ThreeDayView
	dayView      *views.DayView
	agendaView   *views.AgendaView

	// Window size
	width  int
	height int

	// State
	showSidebar bool
	showHelp    bool
}

// currentDate returns today's date — used by sidebar.go and others in this package.
func currentDate() time.Time {
	return time.Now()
}

// NewApp creates the root application model.
func NewApp(cfg *config.Config, store *ical.Store, syncMgr *caldav.Sync) *App {
	theme := config.NewTheme(cfg)
	styles := NewStyles(theme)

	app := &App{
		cfg:         cfg,
		theme:       theme,
		styles:      styles,
		store:       store,
		syncMgr:     syncMgr,
		header:      NewHeader(styles, cfg.General.DefaultView, time.Now()),
		sidebar:     NewSidebar(styles, cfg, store),
		statusbar:   NewStatusBar(styles),
		eventEditor: editor.NewEventEditor(theme, store),
		search:      components.NewSearch(theme, store, cfg.Use24h()),
		viewType:    cfg.General.DefaultView,
		showSidebar: true,
	}

	// Initialize all views
	app.monthView = views.NewMonthView(theme, store, cfg)
	app.weekView = views.NewWeekView(theme, store, cfg)
	app.threeDayView = views.NewThreeDayView(theme, store, cfg)
	app.dayView = views.NewDayView(theme, store, cfg)
	app.agendaView = views.NewAgendaView(theme, store, cfg)

	return app
}

// activeView returns the currently active calendar view.
func (a *App) activeView() CalendarView {
	switch a.viewType {
	case config.ViewMonth:
		return a.monthView
	case config.ViewWeek:
		return a.weekView
	case config.ViewThreeDay:
		return a.threeDayView
	case config.ViewDay:
		return a.dayView
	case config.ViewAgenda:
		return a.agendaView
	default:
		return a.monthView
	}
}

// calendarNames returns the names of discovered calendars.
func (a *App) calendarNames() []string {
	if len(a.store.Calendars) > 0 {
		names := make([]string, len(a.store.Calendars))
		for i, c := range a.store.Calendars {
			names[i] = c.Name
		}
		return names
	}
	if len(a.cfg.Calendars) == 0 {
		return []string{"Work", "Personal"}
	}
	names := make([]string, len(a.cfg.Calendars))
	for i, c := range a.cfg.Calendars {
		names[i] = c.Name
	}
	return names
}

// isReadOnly returns true if the given event's calendar is read-only.
func (a *App) isReadOnly(event *ical.Event) bool {
	if event.CalendarID >= 0 && event.CalendarID < len(a.store.Calendars) {
		return a.store.Calendars[event.CalendarID].ReadOnly
	}
	return false
}

// selectedEvent returns the selected event for the active view.
func (a *App) selectedEvent() *ical.Event {
	// For views with event-level selection, use their cursor
	switch a.viewType {
	case config.ViewWeek:
		if ev := a.weekView.SelectedEvent(); ev != nil {
			return ev
		}
	case config.ViewThreeDay:
		if ev := a.threeDayView.SelectedEvent(); ev != nil {
			return ev
		}
	case config.ViewDay:
		if ev := a.dayView.SelectedEvent(); ev != nil {
			return ev
		}
	}

	// Fallback: first event on the selected day
	day := a.activeView().SelectedDate()
	events := a.store.EventsForDay(day)
	if len(events) > 0 {
		return events[0]
	}
	return nil
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.SetWindowTitle("TipiCal"),
		a.header.WaveTick(), // Start the wave animation
	}

	// If we have a sync manager, trigger initial sync
	if a.syncMgr != nil {
		a.statusbar.SetSyncState("syncing")
		cmds = append(cmds, a.doSync())
	}

	return tea.Batch(cmds...)
}

// doSync returns a tea.Cmd that performs a CalDAV sync in the background.
func (a *App) doSync() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := a.syncMgr.SyncAll(ctx)
		return syncResultMsg{err: err}
	}
}

// scheduleSyncTick returns a tea.Cmd that waits the configured interval then triggers a sync.
func (a *App) scheduleSyncTick() tea.Cmd {
	interval := time.Duration(a.cfg.Sync.IntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return syncMsg{}
	})
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateLayout()
		return a, nil

	case syncMsg:
		// Background sync tick fired
		if a.syncMgr != nil {
			a.statusbar.SetSyncState("syncing")
			return a, a.doSync()
		}
		return a, nil

	case syncResultMsg:
		if msg.err != nil {
			a.statusbar.SetSyncState("error")
			a.statusbar.SetMessage("Sync error: " + msg.err.Error())
		} else {
			a.statusbar.SetSyncState("synced")
			a.statusbar.SetMessage("")
			a.refreshTimeGridViews()
		}
		// Schedule next sync
		var cmd tea.Cmd
		if a.syncMgr != nil {
			cmd = a.scheduleSyncTick()
		}
		return a, cmd

	case upsyncResultMsg:
		if msg.err != nil {
			a.statusbar.SetMessage("Sync failed: " + msg.err.Error())
		} else {
			switch msg.action {
			case "create":
				a.statusbar.SetMessage("Event created and synced")
			case "update":
				a.statusbar.SetMessage("Event updated and synced")
			case "delete":
				a.statusbar.SetMessage("Event deleted and synced")
			}
		}
		return a, nil

	case waveTickMsg:
		// Advance the wave animation
		a.header.UpdateWave()
		return a, a.header.WaveTick()

	case tea.MouseMsg:
		if !a.search.IsActive() && !a.eventEditor.IsActive() && !a.showHelp {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				a.activeView().MoveUp()
				a.syncDateToHeader()
			case tea.MouseButtonWheelDown:
				a.activeView().MoveDown()
				a.syncDateToHeader()
			}
		}
		return a, nil

	case tea.KeyMsg:
		// If search is active, route keys to it
		if a.search.IsActive() {
			closed := a.search.HandleKey(msg.String())
			if closed {
				if event := a.search.SelectedEvent(); event != nil {
					// Navigate to the selected event's date
					a.activeView().SetDate(event.Start)
					a.syncDateToHeader()
				}
				a.search.Close()
			}
			return a, nil
		}

		// If editor is active, route all keys to it
		if a.eventEditor.IsActive() {
			a.eventEditor.HandleKey(msg.String())

			// Check for status message (e.g. clipboard feedback)
			if sm := a.eventEditor.StatusMessage(); sm != "" {
				a.statusbar.SetMessage(sm)
				a.eventEditor.ClearStatusMessage()
			}

			// Check for result
			if result := a.eventEditor.Result(); result != nil {
				cmd := a.handleEditorResult(result)
				a.eventEditor.ClearResult()
				return a, cmd
			}
			return a, nil
		}

		// If help is showing, any key closes it
		if a.showHelp {
			a.showHelp = false
			return a, nil
		}

		// Global keys
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit

		// View switching
		case "1", "m":
			a.switchView(config.ViewMonth)
		case "2", "w":
			a.switchView(config.ViewWeek)
		case "3":
			a.switchView(config.ViewThreeDay)
		case "4", "d":
			a.switchView(config.ViewDay)
		case "5", "a":
			a.switchView(config.ViewAgenda)

		// Navigation
		case "h", "left":
			a.activeView().MoveLeft()
			a.syncDateToHeader()
		case "l", "right":
			a.activeView().MoveRight()
			a.syncDateToHeader()
		case "j", "down":
			a.activeView().MoveDown()
			a.syncDateToHeader()
		case "k", "up":
			a.activeView().MoveUp()
			a.syncDateToHeader()

		// Period navigation
		case "H":
			a.activeView().PrevPeriod()
			a.syncDateToHeader()
		case "L":
			a.activeView().NextPeriod()
			a.syncDateToHeader()

		// Jump to today
		case "t":
			a.activeView().SetDate(time.Now())
			a.syncDateToHeader()

		// Toggle sidebar
		case "b":
			a.showSidebar = !a.showSidebar
			a.updateLayout()

		// Event management
		case "n":
			a.eventEditor.OpenCreate(a.activeView().SelectedDate(), a.calendarNames())
			a.eventEditor.SetSize(a.width, a.height-3)

		case "e":
			if event := a.selectedEvent(); event != nil {
				if a.isReadOnly(event) {
					a.statusbar.SetMessage("Calendar is read-only")
				} else {
					a.eventEditor.OpenEdit(event, a.calendarNames())
					a.eventEditor.SetSize(a.width, a.height-3)
				}
			} else {
				a.statusbar.SetMessage("No event selected")
			}

		case "D":
			if event := a.selectedEvent(); event != nil {
				if a.isReadOnly(event) {
					a.statusbar.SetMessage("Calendar is read-only")
				} else {
					a.eventEditor.OpenDelete(event)
					a.eventEditor.SetSize(a.width, a.height-3)
				}
			} else {
				a.statusbar.SetMessage("No event selected")
			}

		// Event detail
		case "enter":
			if a.viewType == config.ViewAgenda || a.viewType == config.ViewMonth {
				break
			}
			if event := a.selectedEvent(); event != nil {
				calName := a.store.CalendarName(event.CalendarID)
				a.eventEditor.OpenDetail(event, calName, a.cfg.Use24h())
				a.eventEditor.SetSize(a.width, a.height-3)
			}

		// Search
		case "/":
			a.search.SetSize(a.width, a.height-3)
			a.search.Open()

		// Help
		case "?":
			a.showHelp = !a.showHelp
		}
	}

	return a, nil
}

func (a *App) handleEditorResult(result *editor.EditorResult) tea.Cmd {
	switch result.Action {
	case "create":
		if result.Event != nil {
			calID := result.Event.CalendarID
			if calID >= 0 && calID < len(a.store.Calendars) {
				if a.store.Calendars[calID].ReadOnly {
					a.statusbar.SetMessage("Calendar is read-only")
					return nil
				}
				result.Event.CalPath = a.store.Calendars[calID].CalPath
			}
			a.store.AddEvent(result.Event)
			a.refreshTimeGridViews()
			a.statusbar.SetMessage("Creating event...")
			if a.syncMgr != nil {
				event := result.Event
				return func() tea.Msg {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					err := a.syncMgr.CreateEvent(ctx, event)
					return upsyncResultMsg{action: "create", err: err}
				}
			}
			a.statusbar.SetMessage("Event created: " + result.Event.Summary)
		}
	case "update":
		if result.Event != nil {
			// Preserve CalDAV paths from the original event
			if orig := a.store.FindEvent(result.Event.UID); orig != nil {
				result.Event.CalPath = orig.CalPath
				result.Event.ResourcePath = orig.ResourcePath
			}
			a.store.RemoveEvent(result.Event.UID)
			a.store.AddEvent(result.Event)
			a.refreshTimeGridViews()
			a.statusbar.SetMessage("Updating event...")
			if a.syncMgr != nil {
				event := result.Event
				return func() tea.Msg {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					err := a.syncMgr.UpdateEvent(ctx, event)
					return upsyncResultMsg{action: "update", err: err}
				}
			}
			a.statusbar.SetMessage("Event updated: " + result.Event.Summary)
		}
	case "delete":
		if result.Event != nil {
			a.store.RemoveEvent(result.Event.UID)
			a.refreshTimeGridViews()
			a.statusbar.SetMessage("Deleting event...")
			if a.syncMgr != nil {
				event := result.Event
				return func() tea.Msg {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					err := a.syncMgr.DeleteEvent(ctx, event)
					return upsyncResultMsg{action: "delete", err: err}
				}
			}
			a.statusbar.SetMessage("Event deleted: " + result.Event.Summary)
		}
	case "edit-from-detail":
		if result.Event != nil {
			a.eventEditor.OpenEdit(result.Event, a.calendarNames())
			a.eventEditor.SetSize(a.width, a.height-3)
		}
	case "delete-from-detail":
		if result.Event != nil {
			a.eventEditor.OpenDelete(result.Event)
			a.eventEditor.SetSize(a.width, a.height-3)
		}
	case "cancel":
		a.statusbar.SetMessage("")
	}
	return nil
}

// refreshTimeGridViews forces WeekView, ThreeDayView, and DayView to re-snapshot
// events from the store so changes appear immediately without waiting for a sync.
func (a *App) refreshTimeGridViews() {
	a.weekView.RefreshEvents()
	a.threeDayView.RefreshEvents()
	a.dayView.RefreshEvents()
}

func (a *App) switchView(vt config.ViewType) {
	currentDate := a.activeView().SelectedDate()
	a.viewType = vt
	a.activeView().SetDate(currentDate)
	a.header.SetView(vt)
	a.header.SetDate(currentDate)
	a.updateLayout()
}

func (a *App) syncDateToHeader() {
	d := a.activeView().SelectedDate()
	a.header.SetDate(d)
	a.sidebar.miniCal.SetDate(d)
}

func (a *App) updateLayout() {
	sidebarWidth := 0
	sidebarRenderedWidth := 0 // includes border
	if a.showSidebar {
		sidebarWidth = 24
		sidebarRenderedWidth = 25 // 24 content + 1 right border
	}

	contentWidth := a.width - sidebarRenderedWidth
	contentHeight := a.height - 4 // header (1 line + border) + statusbar (border + 1 line)

	a.header.SetWidth(a.width)
	a.statusbar.SetWidth(a.width)
	a.sidebar.SetSize(sidebarWidth, contentHeight)

	// Update all views with content size
	a.monthView.SetSize(contentWidth, contentHeight)
	a.weekView.SetSize(contentWidth, contentHeight)
	a.threeDayView.SetSize(contentWidth, contentHeight)
	a.dayView.SetSize(contentWidth, contentHeight)
	a.agendaView.SetSize(contentWidth, contentHeight)

	// Update editor size
	a.eventEditor.SetSize(a.width, contentHeight)
}

// View implements tea.Model.
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	// Header
	header := a.header.View()

	// Content area
	var content string
	viewContent := a.activeView().View()

	if a.showSidebar {
		sidebarContent := a.sidebar.View()
		content = lipgloss.JoinHorizontal(lipgloss.Top, sidebarContent, viewContent)
	} else {
		content = viewContent
	}

	// Event editor overlay
	if a.eventEditor.IsActive() {
		content = a.eventEditor.View()
	}

	// Search overlay
	if a.search.IsActive() {
		content = a.search.View()
	}

	// Help overlay
	if a.showHelp {
		content = a.renderHelp()
	}

	// Status bar
	statusbar := a.statusbar.View()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusbar)
}

func (a *App) renderHelp() string {
	helpStyle := a.styles.Modal.Width(50)
	titleStyle := a.styles.ModalTitle.Width(46)

	keyStyle := lipgloss.NewStyle().
		Foreground(a.theme.Accent).
		Background(a.theme.Surface).
		Bold(true).
		Width(12)
	descStyle := lipgloss.NewStyle().
		Foreground(a.theme.Text).
		Background(a.theme.Surface).
		Width(34)
	emptyStyle := lipgloss.NewStyle().
		Background(a.theme.Surface).
		Width(46)

	bindings := []struct {
		key  string
		desc string
	}{
		{"1/m", "Month view"},
		{"2/w", "Week view"},
		{"3", "3-Day view"},
		{"4/d", "Day view"},
		{"5/a", "Agenda view"},
		{"", ""},
		{"h/l ←/→", "Move left/right"},
		{"j/k ↑/↓", "Move up/down"},
		{"H/L", "Prev/next period"},
		{"t", "Jump to today"},
		{"b", "Toggle sidebar"},
		{"", ""},
		{"n", "New event"},
		{"e", "Edit event"},
		{"D", "Delete event"},
		{"/", "Search"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Keyboard Shortcuts"))
	lines = append(lines, emptyStyle.Render(""))

	for _, b := range bindings {
		if b.key == "" {
			lines = append(lines, emptyStyle.Render(""))
			continue
		}
		lines = append(lines, keyStyle.Render(b.key)+descStyle.Render(b.desc))
	}

	helpContent := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Center in the content area
	return lipgloss.Place(
		a.width,
		a.height-3,
		lipgloss.Center,
		lipgloss.Center,
		helpStyle.Render(helpContent),
	)
}
