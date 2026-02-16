package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/tipical/tipical/config"
)

// Styles holds all the shared lip gloss styles.
type Styles struct {
	Theme *config.Theme

	// Layout
	App       lipgloss.Style
	Header    lipgloss.Style
	Sidebar   lipgloss.Style
	Content   lipgloss.Style
	StatusBar lipgloss.Style

	// View switcher
	ViewTab       lipgloss.Style
	ViewTabActive lipgloss.Style

	// Calendar grid
	DayCell         lipgloss.Style
	DayCellToday    lipgloss.Style
	DayCellSelected lipgloss.Style
	DayCellOther    lipgloss.Style // days from other months
	DayNumber       lipgloss.Style
	DayNumberToday  lipgloss.Style
	WeekdayHeader   lipgloss.Style

	// Events
	EventDot   lipgloss.Style
	EventTitle lipgloss.Style
	EventTime  lipgloss.Style
	EventCard  lipgloss.Style

	// Time grid
	TimeLabel lipgloss.Style
	TimeSlot  lipgloss.Style
	TimeLine  lipgloss.Style
	NowLine   lipgloss.Style

	// Modal / form
	Modal      lipgloss.Style
	ModalTitle lipgloss.Style
	FormLabel  lipgloss.Style
	FormInput  lipgloss.Style
	FormButton lipgloss.Style

	// Misc
	Title   lipgloss.Style
	Help    lipgloss.Style
	Error   lipgloss.Style
	Success lipgloss.Style
}

// NewStyles creates a Styles instance from the given theme.
func NewStyles(theme *config.Theme) *Styles {
	s := &Styles{Theme: theme}

	s.App = lipgloss.NewStyle().
		Background(theme.Background)

	s.Header = lipgloss.NewStyle().
		Foreground(theme.Text).
		Bold(true).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Border)

	s.Sidebar = lipgloss.NewStyle().
		Foreground(theme.Text).
		Padding(1, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderRight(true).
		BorderForeground(theme.Border)

	s.Content = lipgloss.NewStyle().
		Padding(0, 0)

	s.StatusBar = lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(theme.Border)

	// View tabs
	s.ViewTab = lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Padding(0, 1).
		MarginRight(1)

	s.ViewTabActive = lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Padding(0, 1).
		MarginRight(1).
		Underline(true)

	// Day cells
	s.DayCell = lipgloss.NewStyle().
		Foreground(theme.Text).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border)

	s.DayCellToday = lipgloss.NewStyle().
		Foreground(theme.Text).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Today).
		Bold(true)

	s.DayCellSelected = lipgloss.NewStyle().
		Foreground(theme.Text).
		Background(theme.SurfaceAlt).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Selected)

	s.DayCellOther = lipgloss.NewStyle().
		Foreground(theme.TextFaint).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border)

	s.DayNumber = lipgloss.NewStyle().
		Foreground(theme.Text).
		Padding(0, 0)

	s.DayNumberToday = lipgloss.NewStyle().
		Foreground(theme.Today).
		Bold(true)

	s.WeekdayHeader = lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Bold(true).
		Align(lipgloss.Center)

	// Events
	s.EventDot = lipgloss.NewStyle()

	s.EventTitle = lipgloss.NewStyle().
		Foreground(theme.Text)

	s.EventTime = lipgloss.NewStyle().
		Foreground(theme.TextMuted)

	s.EventCard = lipgloss.NewStyle().
		Foreground(theme.Text).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border)

	// Time grid
	s.TimeLabel = lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Width(6).
		Align(lipgloss.Right)

	s.TimeSlot = lipgloss.NewStyle().
		Foreground(theme.Text).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Border)

	s.TimeLine = lipgloss.NewStyle().
		Foreground(theme.Border)

	s.NowLine = lipgloss.NewStyle().
		Foreground(theme.Today)

	// Modal
	s.Modal = lipgloss.NewStyle().
		Foreground(theme.Text).
		Background(theme.Surface).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent)

	s.ModalTitle = lipgloss.NewStyle().
		Foreground(theme.Accent).
		Background(theme.Surface).
		Bold(true).
		Align(lipgloss.Center)

	s.FormLabel = lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Width(12).
		Align(lipgloss.Right).
		MarginRight(1)

	s.FormInput = lipgloss.NewStyle().
		Foreground(theme.Text).
		Background(theme.SurfaceAlt).
		Padding(0, 1)

	s.FormButton = lipgloss.NewStyle().
		Foreground(theme.Text).
		Background(theme.Accent).
		Padding(0, 2).
		Bold(true)

	// Misc
	s.Title = lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true)

	s.Help = lipgloss.NewStyle().
		Foreground(theme.TextFaint)

	s.Error = lipgloss.NewStyle().
		Foreground(theme.Error)

	s.Success = lipgloss.NewStyle().
		Foreground(theme.Success)

	return s
}
