package ical

import (
	"time"
)

// Event represents a calendar event.
type Event struct {
	UID         string
	Summary     string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	AllDay      bool
	CalendarID  int // index into config calendar list
	Color       string
	Status      string // CONFIRMED, TENTATIVE, CANCELLED
	Recurring   bool
}

// Duration returns the duration of the event.
func (e *Event) Duration() time.Duration {
	return e.End.Sub(e.Start)
}

// DurationHours returns the duration in fractional hours.
func (e *Event) DurationHours() float64 {
	return e.Duration().Hours()
}

// IsMultiDay returns true if the event spans multiple days.
func (e *Event) IsMultiDay() bool {
	if e.AllDay {
		// For all-day events, End is exclusive (midnight of next day).
		// A single-day all-day event has End = Start + 1 day, so compare
		// using date components to avoid timezone confusion.
		return e.AllDaySpanDays() > 1
	}
	sy, sm, sd := e.Start.Date()
	ey, em, ed := e.End.Date()
	return sy != ey || sm != em || sd != ed
}

// AllDaySpanDays returns how many calendar days an all-day event covers.
// Returns 0 for non-all-day events.
func (e *Event) AllDaySpanDays() int {
	if !e.AllDay {
		return 0
	}
	sy, sm, sd := e.Start.Date()
	ey, em, ed := e.End.Date()
	startDate := time.Date(sy, sm, sd, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(ey, em, ed, 0, 0, 0, 0, time.UTC)
	days := int(endDate.Sub(startDate).Hours() / 24)
	if days < 1 {
		days = 1
	}
	return days
}

// OverlapsWith returns true if this event overlaps the given time range.
func (e *Event) OverlapsWith(start, end time.Time) bool {
	return e.Start.Before(end) && e.End.After(start)
}

// CalendarInfo represents a discovered sub-calendar at runtime.
type CalendarInfo struct {
	ID     int    // unique sequential ID
	Name   string // discovered calendar name (from CalDAV)
	Color  string // resolved hex color
	Source string // parent config calendar name (account name)
}

// Store holds events in memory and provides filtering.
type Store struct {
	Events    []*Event
	Calendars []CalendarInfo
}

// NewStore creates an empty event store.
func NewStore() *Store {
	return &Store{}
}

// RegisterCalendar adds a calendar to the registry and returns its ID.
func (s *Store) RegisterCalendar(info CalendarInfo) int {
	info.ID = len(s.Calendars)
	s.Calendars = append(s.Calendars, info)
	return info.ID
}

// ClearCalendars resets the calendar registry.
func (s *Store) ClearCalendars() {
	s.Calendars = nil
}

// CalendarName returns the name of the calendar with the given ID.
func (s *Store) CalendarName(id int) string {
	if id >= 0 && id < len(s.Calendars) {
		return s.Calendars[id].Name
	}
	return ""
}

// EventsForDay returns all events that occur on the given day.
func (s *Store) EventsForDay(day time.Time) []*Event {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	var result []*Event
	for _, e := range s.Events {
		if e.AllDay {
			// For all-day events, compare date components only to avoid
			// timezone mismatches (all-day times are often in UTC while
			// the day is in local time).
			dy, dm, dd := day.Date()
			sy, sm, sd := e.Start.Date()
			ey, em, ed := e.End.Date()
			dayNorm := time.Date(dy, dm, dd, 0, 0, 0, 0, time.UTC)
			startNorm := time.Date(sy, sm, sd, 0, 0, 0, 0, time.UTC)
			endNorm := time.Date(ey, em, ed, 0, 0, 0, 0, time.UTC)
			// End is exclusive: day must be >= start and < end
			if !dayNorm.Before(startNorm) && dayNorm.Before(endNorm) {
				result = append(result, e)
			}
		} else if e.OverlapsWith(dayStart, dayEnd) {
			result = append(result, e)
		}
	}
	return result
}

// EventsInRange returns all events overlapping [start, end).
func (s *Store) EventsInRange(start, end time.Time) []*Event {
	var result []*Event
	for _, e := range s.Events {
		if e.OverlapsWith(start, end) {
			result = append(result, e)
		}
	}
	return result
}

// AddEvent adds an event to the store.
func (s *Store) AddEvent(e *Event) {
	s.Events = append(s.Events, e)
}

// RemoveEvent removes an event by UID.
func (s *Store) RemoveEvent(uid string) {
	for i, e := range s.Events {
		if e.UID == uid {
			s.Events = append(s.Events[:i], s.Events[i+1:]...)
			return
		}
	}
}

// FindEvent finds an event by UID.
func (s *Store) FindEvent(uid string) *Event {
	for _, e := range s.Events {
		if e.UID == uid {
			return e
		}
	}
	return nil
}

// LoadTestData populates the store with sample events for development.
func (s *Store) LoadTestData() {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// Register test calendars
	s.ClearCalendars()
	s.RegisterCalendar(CalendarInfo{
		Name:   "Work",
		Color:  "#3B82F6",
		Source: "Work Account",
	})
	s.RegisterCalendar(CalendarInfo{
		Name:   "Personal",
		Color:  "#10B981",
		Source: "Personal Account",
	})

	testEvents := []*Event{
		{
			UID:        "test-1",
			Summary:    "Team Standup",
			Location:   "Room 3A",
			Start:      today.Add(9 * time.Hour),
			End:        today.Add(9*time.Hour + 30*time.Minute),
			CalendarID: 0,
			Color:      "#3B82F6",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-2",
			Summary:    "Lunch with Sarah",
			Location:   "Cafe Roma",
			Start:      today.Add(12 * time.Hour),
			End:        today.Add(13 * time.Hour),
			CalendarID: 1,
			Color:      "#10B981",
			Status:     "CONFIRMED",
		},
		{
			UID:         "test-3",
			Summary:     "Project Review",
			Description: "Q1 review of all active projects and roadmap discussion",
			Location:    "Board Room",
			Start:       today.Add(14 * time.Hour),
			End:         today.Add(15*time.Hour + 30*time.Minute),
			CalendarID:  0,
			Color:       "#3B82F6",
			Status:      "CONFIRMED",
		},
		{
			UID:        "test-4",
			Summary:    "1:1 with Manager",
			Location:   "Virtual",
			Start:      today.AddDate(0, 0, 1).Add(10 * time.Hour),
			End:        today.AddDate(0, 0, 1).Add(11 * time.Hour),
			CalendarID: 0,
			Color:      "#3B82F6",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-5",
			Summary:    "Sprint Planning",
			Location:   "Board Room",
			Start:      today.AddDate(0, 0, 1).Add(16 * time.Hour),
			End:        today.AddDate(0, 0, 1).Add(17 * time.Hour),
			CalendarID: 0,
			Color:      "#3B82F6",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-6",
			Summary:    "Dentist Appointment",
			Location:   "Downtown Dental",
			Start:      today.AddDate(0, 0, 2).Add(11 * time.Hour),
			End:        today.AddDate(0, 0, 2).Add(12 * time.Hour),
			CalendarID: 1,
			Color:      "#10B981",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-7",
			Summary:    "Team Lunch",
			Start:      today.AddDate(0, 0, 3).Add(12 * time.Hour),
			End:        today.AddDate(0, 0, 3).Add(13*time.Hour + 30*time.Minute),
			CalendarID: 0,
			Color:      "#3B82F6",
			Status:     "CONFIRMED",
		},
		{
			UID:         "test-8",
			Summary:     "Design Workshop",
			Description: "UX design workshop for new features",
			Location:    "Creative Space",
			Start:       today.AddDate(0, 0, -1).Add(13 * time.Hour),
			End:         today.AddDate(0, 0, -1).Add(16 * time.Hour),
			CalendarID:  0,
			Color:       "#3B82F6",
			Status:      "CONFIRMED",
		},
		{
			UID:        "test-9",
			Summary:    "Yoga Class",
			Location:   "Studio B",
			Start:      today.AddDate(0, 0, -2).Add(7 * time.Hour),
			End:        today.AddDate(0, 0, -2).Add(8 * time.Hour),
			CalendarID: 1,
			Color:      "#10B981",
			Status:     "CONFIRMED",
		},
		{
			UID:         "test-10",
			Summary:     "Product Launch",
			Description: "Major product launch event",
			Start:       today.AddDate(0, 0, 5).Add(9 * time.Hour),
			End:         today.AddDate(0, 0, 5).Add(17 * time.Hour),
			CalendarID:  0,
			Color:       "#7C3AED",
			Status:      "CONFIRMED",
		},
		{
			UID:        "test-11",
			Summary:    "Birthday Party",
			Location:   "Home",
			Start:      today.AddDate(0, 0, 6).Add(18 * time.Hour),
			End:        today.AddDate(0, 0, 6).Add(22 * time.Hour),
			CalendarID: 1,
			Color:      "#10B981",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-12",
			Summary:    "Code Review",
			Start:      today.Add(16 * time.Hour),
			End:        today.Add(16*time.Hour + 30*time.Minute),
			CalendarID: 0,
			Color:      "#3B82F6",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-13",
			Summary:    "All Hands Meeting",
			Location:   "Auditorium",
			Start:      today.AddDate(0, 0, 4).Add(10 * time.Hour),
			End:        today.AddDate(0, 0, 4).Add(11 * time.Hour),
			CalendarID: 0,
			Color:      "#3B82F6",
			Status:     "CONFIRMED",
		},
		{
			UID:         "test-multiday",
			Summary:     "Hackathon",
			Description: "48-hour company hackathon",
			Location:    "Office",
			Start:       today.Add(14 * time.Hour),
			End:         today.AddDate(0, 0, 2).Add(14 * time.Hour),
			CalendarID:  1,
			Color:       "#F59E0B",
			Status:      "CONFIRMED",
		},
		{
			UID:        "test-14",
			Summary:    "Company Holiday",
			Start:      today,
			End:        today.AddDate(0, 0, 1),
			AllDay:     true,
			CalendarID: 0,
			Color:      "#7C3AED",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-15",
			Summary:    "Team Offsite",
			Start:      today.AddDate(0, 0, 1),
			End:        today.AddDate(0, 0, 2),
			AllDay:     true,
			CalendarID: 0,
			Color:      "#3B82F6",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-16",
			Summary:    "Conf Week",
			Start:      today,
			End:        today.AddDate(0, 0, 5),
			AllDay:     true,
			CalendarID: 1,
			Color:      "#F59E0B",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-17",
			Summary:    "Sprint Review",
			Start:      today,
			End:        today.AddDate(0, 0, 1),
			AllDay:     true,
			CalendarID: 0,
			Color:      "#EF4444",
			Status:     "CONFIRMED",
		},
		{
			UID:        "test-18",
			Summary:    "Maintenance Window",
			Start:      today,
			End:        today.AddDate(0, 0, 3),
			AllDay:     true,
			CalendarID: 0,
			Color:      "#6366F1",
			Status:     "CONFIRMED",
		},
	}

	s.Events = testEvents
}
