# Building an Application to Schedule Friday Talks

## Introduction

In many organizations and communities, knowledge sharing through regular talks is a valuable practice. This document outlines a plan to build an application specifically designed to schedule and manage weekly Friday talks for a small group of enthusiasts. The application will streamline the process of proposing, scheduling, and managing these 1-hour sessions.

## Overview and Requirements

The application should fulfill the following requirements:

1. Allow users to propose topics for talks
2. Manage a calendar of scheduled talks
3. Provide a simple interface for viewing upcoming talks
4. Send reminders to speakers and potential attendees
5. Store historical data of past talks
6. Support basic user management to identify speakers and attendees

## Architecture Overview

We'll design this as a modern web application with a clean separation of concerns:

- **Backend**: Go-based API server
- **Frontend**: Simple yet elegant UI using HTMX and Bootstrap
- **Templates**: Using the templ templating language
- **Database**: SQLite for simplicity, can be upgraded to PostgreSQL if needed
- **CLI**: Cobra-based command line interface for administrative tasks

### System Components

```
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│                  │       │                  │       │                  │
│  Web Interface   │◄─────►│  API Server      │◄─────►│  Database        │
│  (HTMX/Bootstrap)│       │  (Go)            │       │  (SQLite)        │
│                  │       │                  │       │                  │
└──────────────────┘       └──────────────────┘       └──────────────────┘
                                    ▲
                                    │
                                    ▼
                           ┌──────────────────┐
                           │                  │
                           │  CLI Admin Tool  │
                           │  (Cobra)         │
                           │                  │
                           └──────────────────┘
```

## Data Model

### Core Entities

1. **User**
   - ID
   - Name
   - Email
   - Role (Admin, Speaker, Attendee)

2. **Talk**
   - ID
   - Title
   - Description
   - SpeakerID
   - ScheduledDate
   - Duration (default: 1h)
   - Status (Proposed, Scheduled, Completed, Cancelled)
   - Materials (links to slides, code, etc.)

3. **Attendance**
   - TalkID
   - UserID
   - Status (Attending, Maybe, Not Attending)

## Implementation Plan

### 1. Project Setup

```go
// Directory structure
talks-scheduler/
├── cmd/
│   └── scheduler/
│       └── main.go
├── internal/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   └── service/
├── templates/
├── static/
├── migrations/
└── go.mod
```

### 2. Database Schema and Models

```go
// internal/models/user.go
type User struct {
    ID        int64  `db:"id"`
    Name      string `db:"name"`
    Email     string `db:"email"`
    Role      string `db:"role"`
    CreatedAt time.Time `db:"created_at"`
}

// internal/models/talk.go
type Talk struct {
    ID           int64     `db:"id"`
    Title        string    `db:"title"`
    Description  string    `db:"description"`
    SpeakerID    int64     `db:"speaker_id"`
    ScheduledDate time.Time `db:"scheduled_date"`
    Duration     int       `db:"duration"` // in minutes
    Status       string    `db:"status"`
    Materials    string    `db:"materials"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}

// internal/models/attendance.go
type Attendance struct {
    TalkID    int64  `db:"talk_id"`
    UserID    int64  `db:"user_id"`
    Status    string `db:"status"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}
```

### 3. Service Layer

```go
// internal/service/talk_service.go
type TalkService interface {
    ListUpcomingTalks(ctx context.Context) ([]Talk, error)
    GetTalk(ctx context.Context, id int64) (*Talk, error)
    ProposeTalk(ctx context.Context, talk Talk) (int64, error)
    ScheduleTalk(ctx context.Context, id int64, date time.Time) error
    CancelTalk(ctx context.Context, id int64) error
    RecordAttendance(ctx context.Context, talkID, userID int64, status string) error
}
```

### 4. API Handlers

```go
// internal/handlers/talk_handlers.go
func (h *TalkHandler) ListTalks(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    talks, err := h.service.ListUpcomingTalks(ctx)
    if err != nil {
        http.Error(w, "Failed to fetch talks", http.StatusInternalServerError)
        return
    }
    
    component := templates.TalksList(talks)
    component.Render(ctx, w)
}

func (h *TalkHandler) ProposeTalk(w http.ResponseWriter, r *http.Request) {
    // Handle form submission to propose a new talk
    // ...
}
```

### 5. Templates with Templ

```go
// templates/talks.templ
package templates

import "yourmodule/internal/models"

templ TalksList(talks []models.Talk) {
    <div class="talks-list">
        <h2>Upcoming Talks</h2>
        <div class="list-group">
            for _, talk := range talks {
                <div class="list-group-item">
                    <h5 class="mb-1">{ talk.Title }</h5>
                    <p class="mb-1">{ talk.Description }</p>
                    <small>
                        { talk.ScheduledDate.Format("January 2, 2006") } • 
                        { talk.Duration } minutes • 
                        Speaker: { talk.SpeakerName }
                    </small>
                </div>
            }
        </div>
    </div>
}
```

### 6. CLI Tools

```go
// cmd/scheduler/main.go
func main() {
    var rootCmd = &cobra.Command{
        Use:   "talks-scheduler",
        Short: "A tool for managing Friday talks",
    }
    
    var serveCmd = &cobra.Command{
        Use:   "serve",
        Short: "Start the web server",
        Run: func(cmd *cobra.Command, args []string) {
            // Initialize and start the server
        },
    }
    
    var addUserCmd = &cobra.Command{
        Use:   "add-user",
        Short: "Add a new user",
        Run: func(cmd *cobra.Command, args []string) {
            // Add user implementation
        },
    }
    
    rootCmd.AddCommand(serveCmd, addUserCmd)
    rootCmd.Execute()
}
```

## Key Features Implementation

### User Authentication

We'll use a simple session-based authentication system:

```go
// internal/middleware/auth.go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session, err := store.Get(r, "talks-session")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        userID, ok := session.Values["user_id"]
        if !ok {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        
        ctx := context.WithValue(r.Context(), "user_id", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Scheduling Algorithm

For a small group, we can implement a simple scheduling system that follows these rules:

1. Only one talk per Friday
2. Talks are scheduled at least 1 week in advance
3. Priority given to speakers who haven't presented recently

```go
// internal/service/scheduler.go
func (s *SchedulerService) SuggestNextAvailableFriday(ctx context.Context) (time.Time, error) {
    // Find next available Friday
    now := time.Now()
    daysUntilFriday := (5 - int(now.Weekday()) + 7) % 7
    nextFriday := now.AddDate(0, 0, daysUntilFriday)
    
    // Look one week ahead for scheduling buffer
    schedulingBuffer := nextFriday.AddDate(0, 0, 7)
    
    // Check if there's already a talk scheduled
    bookedDates, err := s.talkRepo.GetBookedDates(ctx, schedulingBuffer, schedulingBuffer.AddDate(0, 1, 0))
    if err != nil {
        return time.Time{}, err
    }
    
    // Find the first available Friday
    candidateDate := schedulingBuffer
    for len(bookedDates) > 0 {
        isBooked := false
        for _, date := range bookedDates {
            if date.Year() == candidateDate.Year() && 
               date.Month() == candidateDate.Month() && 
               date.Day() == candidateDate.Day() {
                isBooked = true
                break
            }
        }
        
        if !isBooked {
            return candidateDate, nil
        }
        
        candidateDate = candidateDate.AddDate(0, 0, 7)
    }
    
    return schedulingBuffer, nil
}
```

### Email Notifications

```go
// internal/service/notification.go
type NotificationService struct {
    emailClient EmailClient
}

func (s *NotificationService) SendTalkReminder(ctx context.Context, talk Talk, users []User) error {
    // Send reminder to the speaker
    speakerEmail := Email{
        To:      talk.Speaker.Email,
        Subject: fmt.Sprintf("Reminder: Your talk '%s' is scheduled for this Friday", talk.Title),
        Body:    fmt.Sprintf("Don't forget to prepare for your talk on %s", talk.ScheduledDate.Format("January 2")),
    }
    
    if err := s.emailClient.Send(ctx, speakerEmail); err != nil {
        return err
    }
    
    // Send notification to attendees
    for _, user := range users {
        if user.ID == talk.SpeakerID {
            continue // Skip the speaker
        }
        
        attendeeEmail := Email{
            To:      user.Email,
            Subject: fmt.Sprintf("Upcoming talk: %s", talk.Title),
            Body:    fmt.Sprintf("There will be a talk by %s on %s", talk.Speaker.Name, talk.ScheduledDate.Format("January 2")),
        }
        
        if err := s.emailClient.Send(ctx, attendeeEmail); err != nil {
            return err
        }
    }
    
    return nil
}
```

## UI Design

The UI will be clean and functional, with these main views:

1. **Home/Calendar View**: Shows upcoming talks in a calendar format
2. **Talk Detail View**: Information about a specific talk
3. **Propose Talk Form**: Form for proposing new talks
4. **Admin Dashboard**: For managing users and talk proposals

The UI will use Bootstrap for styling and HTMX for interactive elements without requiring a heavy JavaScript framework.

## Deployment Options

1. **Self-Hosted**: Deploy on a small VPS or internal server
2. **Docker**: Package the application as a Docker container for easy deployment
3. **Cloud Services**: Deploy on services like Railway, Fly.io, or Heroku for minimal maintenance

## Future Enhancements

1. **Integration with Video Conferencing**: Automatically create Zoom/Google Meet links
2. **Feedback System**: Allow attendees to provide feedback on talks
3. **Content Repository**: Store slides and materials from past talks
4. **Automatic Recording**: Integration with recording tools
5. **Topic Suggestion System**: Let attendees suggest topics they'd like to hear about

## Conclusion

This application provides a focused solution for managing your Friday talks. The architecture is intentionally kept simple while still following good software design practices. By using Go with HTMX and Bootstrap, we create a responsive application that's easy to maintain and extend.

The focus on simplicity means you can get this up and running quickly, while the modular design allows for adding more sophisticated features as your needs evolve. 