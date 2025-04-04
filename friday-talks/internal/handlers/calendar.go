package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/wesen/friday-talks/internal/auth"
	"github.com/wesen/friday-talks/internal/models"
	"github.com/wesen/friday-talks/internal/templates"
)

// CalendarHandler handles calendar-related routes
type CalendarHandler struct {
	talkRepo models.TalkRepository
}

// NewCalendarHandler creates a new CalendarHandler
func NewCalendarHandler(talkRepo models.TalkRepository) *CalendarHandler {
	return &CalendarHandler{
		talkRepo: talkRepo,
	}
}

// HandleCalendar renders the calendar page
func (h *CalendarHandler) HandleCalendar(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := auth.UserFromContext(r.Context())

	// Get month and year parameters from query
	var month time.Month
	var year int
	now := time.Now()

	monthStr := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")

	if monthStr != "" && yearStr != "" {
		monthInt, err := strconv.Atoi(monthStr)
		if err == nil && monthInt >= 1 && monthInt <= 12 {
			month = time.Month(monthInt)
		} else {
			month = now.Month()
		}

		yearInt, err := strconv.Atoi(yearStr)
		if err == nil && yearInt >= 2000 && yearInt <= 2100 {
			year = yearInt
		} else {
			year = now.Year()
		}
	} else {
		month = now.Month()
		year = now.Year()
	}

	// Create calendar data for current and next month
	currentMonthData := generateCalendarMonth(month, year, now, h.talkRepo, r.Context())

	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	nextMonthData := generateCalendarMonth(nextMonth, nextYear, now, h.talkRepo, r.Context())

	// Render calendar page
	templates.Calendar(user, []templates.CalendarMonth{currentMonthData, nextMonthData}, month, year).Render(r.Context(), w)
}

// generateCalendarMonth creates a calendar month structure for the given month and year
func generateCalendarMonth(month time.Month, year int, now time.Time, talkRepo models.TalkRepository, ctx context.Context) templates.CalendarMonth {
	// Get first day of the month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)

	// Get last day of the month
	lastDay := firstDay.AddDate(0, 1, -1)

	// Get number of days in the month
	daysInMonth := lastDay.Day()

	// Get scheduled talks for this month
	talks, _ := talkRepo.ListByStatus(ctx, models.TalkStatusScheduled)

	// Create a map of dates to talks
	talksByDate := make(map[string]*models.Talk)
	for _, talk := range talks {
		if talk.ScheduledDate != nil {
			date := talk.ScheduledDate.Format("2006-01-02")
			talksByDate[date] = talk
		}
	}

	// Create calendar days
	var days []templates.CalendarDay

	// Get the day of the week for the first day of the month (0 = Sunday, 6 = Saturday)
	firstDayOfWeek := int(firstDay.Weekday())

	// Add days from previous month to fill the first week
	prevMonth := month - 1
	prevYear := year
	if prevMonth < 1 {
		prevMonth = 12
		prevYear--
	}
	prevMonthLastDay := time.Date(prevYear, prevMonth, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, -1)
	prevMonthDays := prevMonthLastDay.Day()

	for i := 0; i < firstDayOfWeek; i++ {
		day := prevMonthDays - firstDayOfWeek + i + 1
		date := time.Date(prevYear, prevMonth, day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")

		days = append(days, templates.CalendarDay{
			Date:         date,
			IsOtherMonth: true,
			IsToday:      date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day(),
			Talk:         talksByDate[dateStr],
			HasTalk:      talksByDate[dateStr] != nil,
		})
	}

	// Add days for the current month
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")

		days = append(days, templates.CalendarDay{
			Date:    date,
			IsToday: date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day(),
			Talk:    talksByDate[dateStr],
			HasTalk: talksByDate[dateStr] != nil,
		})
	}

	// Add days from next month to complete the last week
	lastDayOfWeek := int(lastDay.Weekday())
	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}

	for i := lastDayOfWeek + 1; i < 7; i++ {
		day := i - lastDayOfWeek
		date := time.Date(nextYear, nextMonth, day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")

		days = append(days, templates.CalendarDay{
			Date:         date,
			IsOtherMonth: true,
			IsToday:      date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day(),
			Talk:         talksByDate[dateStr],
			HasTalk:      talksByDate[dateStr] != nil,
		})
	}

	// Arrange days into weeks
	var weeks [][7]templates.CalendarDay
	numWeeks := len(days) / 7

	for i := 0; i < numWeeks; i++ {
		var week [7]templates.CalendarDay
		for j := 0; j < 7; j++ {
			week[j] = days[i*7+j]
		}
		weeks = append(weeks, week)
	}

	return templates.CalendarMonth{
		Month: month,
		Year:  year,
		Days:  days,
		Weeks: weeks,
	}
}
