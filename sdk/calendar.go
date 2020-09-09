package sdk

import "github.com/pinpt/integration-sdk/calendar"

// CalendarCalendar details for the given integration calendar
type CalendarCalendar = calendar.Calendar

// CalendarEvent event on a calendar
type CalendarEvent = calendar.Event

// CalendarEventEndDate represents the object structure for end_date
type CalendarEventEndDate = calendar.EventEndDate

// CalendarEventLocation represents the object structure for location
type CalendarEventLocation = calendar.EventLocation

// CalendarEventStartDate represents the object structure for start_date
type CalendarEventStartDate = calendar.EventStartDate

// CalendarEventStatus is the enumeration type for status
type CalendarEventStatus = calendar.EventStatus

// CalendarEventStatusConfirmed is the enumeration value for confirmed
const CalendarEventStatusConfirmed CalendarEventStatus = calendar.EventStatusConfirmed

// CalendarEventStatusTentative is the enumeration value for tentative
const CalendarEventStatusTentative CalendarEventStatus = calendar.EventStatusTentative

// CalendarEventStatusCancelled is the enumeration value for cancelled
const CalendarEventStatusCancelled CalendarEventStatus = calendar.EventStatusCancelled

// CalendarUser the calendar user
type CalendarUser = calendar.User

// CalendarNewCalendarID provides a template for generating an ID field for Calendar
func CalendarNewCalendarID(customerID string, refID string, refType string) string {
	return calendar.NewCalendarID(customerID, refType, refID)
}

// CalendarNewEventID provides a template for generating an ID field for Event
func CalendarNewEventID(customerID string, refID string, refType string, ownerRefID string) string {
	return calendar.NewEventID(customerID, refType, refID, ownerRefID)
}

// CalendarNewUserID provides a template for generating an ID field for User
func CalendarNewUserID(customerID string, refID string, refType string) string {
	return calendar.NewUserID(customerID, refType, refID)
}
