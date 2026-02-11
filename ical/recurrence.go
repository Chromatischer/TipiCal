package ical

import (
	"time"

	"github.com/teambition/rrule-go"
)

// ExpandRecurrence expands a recurring event into concrete instances
// within the given date range.
func ExpandRecurrence(event *Event, rruleStr string, rangeStart, rangeEnd time.Time) []*Event {
	if rruleStr == "" {
		return []*Event{event}
	}

	rule, err := rrule.StrToRRule(rruleStr)
	if err != nil {
		// If we can't parse the rule, return the original event
		return []*Event{event}
	}

	rule.DTStart(event.Start)

	instances := rule.Between(rangeStart, rangeEnd, true)
	duration := event.End.Sub(event.Start)

	var expanded []*Event
	for _, instanceStart := range instances {
		instance := &Event{
			UID:         event.UID,
			Summary:     event.Summary,
			Description: event.Description,
			Location:    event.Location,
			Start:       instanceStart,
			End:         instanceStart.Add(duration),
			AllDay:      event.AllDay,
			CalendarID:  event.CalendarID,
			Color:       event.Color,
			Status:      event.Status,
			Recurring:   true,
		}
		expanded = append(expanded, instance)
	}

	return expanded
}
