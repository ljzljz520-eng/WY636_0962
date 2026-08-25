package planning

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type DayWindow struct {
	Weekday time.Weekday
	Open    time.Duration
	Close   time.Duration
	Closed  bool
}
type Calendar struct {
	Name     string
	Location string
	Windows  map[time.Weekday]DayWindow
	Holidays map[string]string
}
type Availability struct {
	Start     time.Time
	End       time.Time
	Available bool
	Reason    string
}

func NewCalendar(name, location string) Calendar {
	return Calendar{Name: strings.TrimSpace(name), Location: strings.TrimSpace(location), Windows: make(map[time.Weekday]DayWindow), Holidays: make(map[string]string)}
}
func (c *Calendar) SetWindow(window DayWindow) error {
	if window.Open < 0 || window.Close > 24*time.Hour || window.Open >= window.Close {
		return fmt.Errorf("invalid day window")
	}
	c.Windows[window.Weekday] = window
	return nil
}
func (c *Calendar) AddHoliday(date time.Time, label string) {
	if c.Holidays == nil {
		c.Holidays = make(map[string]string)
	}
	c.Holidays[date.Format("2006-01-02")] = label
}
func (c Calendar) IsHoliday(at time.Time) bool {
	_, ok := c.Holidays[at.Format("2006-01-02")]
	return ok
}
func (c Calendar) WindowFor(at time.Time) (DayWindow, bool) {
	window, ok := c.Windows[at.Weekday()]
	return window, ok
}
func (c Calendar) IsOpen(at time.Time) bool {
	if c.IsHoliday(at) {
		return false
	}
	window, ok := c.WindowFor(at)
	if !ok || window.Closed {
		return false
	}
	sinceMidnight := time.Duration(at.Hour())*time.Hour + time.Duration(at.Minute())*time.Minute + time.Duration(at.Second())*time.Second
	return sinceMidnight >= window.Open && sinceMidnight < window.Close
}
func (c Calendar) Check(start, end time.Time) Availability {
	if !start.Before(end) {
		return Availability{Start: start, End: end, Reason: "end must be after start"}
	}
	for cursor := start; cursor.Before(end); cursor = cursor.Add(30 * time.Minute) {
		if !c.IsOpen(cursor) {
			return Availability{Start: start, End: end, Reason: "calendar closed at " + cursor.Format(time.RFC3339)}
		}
	}
	return Availability{Start: start, End: end, Available: true, Reason: "calendar open"}
}
func NextOpen(c Calendar, from time.Time, limit int) (time.Time, bool) {
	candidate := from
	for i := 0; i < limit; i++ {
		if c.IsOpen(candidate) {
			return candidate, true
		}
		candidate = candidate.Add(30 * time.Minute)
	}
	return time.Time{}, false
}
func WindowSummary(c Calendar) []DayWindow {
	result := make([]DayWindow, 0, len(c.Windows))
	for _, window := range c.Windows {
		result = append(result, window)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Weekday < result[j].Weekday })
	return result
}
