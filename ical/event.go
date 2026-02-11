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
	sy, sm, sd := e.Start.Date()
	ey, em, ed := e.End.Date()
	return sy != ey || sm != em || sd != ed
}

// OverlapsWith returns true if this event overlaps the given time range.
func (e *Event) OverlapsWith(start, end time.Time) bool {
	return e.Start.Before(end) && e.End.After(start)
}

// Store holds events in memory and provides filtering.
type Store struct {
	Events []*Event
}

// NewStore creates an empty event store.
func NewStore() *Store {
	return &Store{}
}

// EventsForDay returns all events that occur on the given day.
func (s *Store) EventsForDay(day time.Time) []*Event {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	var result []*Event
	for _, e := range s.Events {
		if e.OverlapsWith(dayStart, dayEnd) {
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
	}

	s.Events = testEvents
}
