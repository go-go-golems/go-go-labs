package services

import (
	"context"
	"fmt"
	"net/smtp"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/wesen/friday-talks/internal/models"
)

// NotificationService handles sending notifications to users
type NotificationService struct {
	userRepo       models.UserRepository
	talkRepo       models.TalkRepository
	attendanceRepo models.AttendanceRepository
	config         NotificationConfig
}

// NotificationConfig holds configuration for email sending
type NotificationConfig struct {
	Enabled    bool
	SMTPHost   string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
	SenderName string
	SenderMail string
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(
	userRepo models.UserRepository,
	talkRepo models.TalkRepository,
	attendanceRepo models.AttendanceRepository,
	config NotificationConfig,
) *NotificationService {
	return &NotificationService{
		userRepo:       userRepo,
		talkRepo:       talkRepo,
		attendanceRepo: attendanceRepo,
		config:         config,
	}
}

// SendEmail sends an email to a recipient
func (n *NotificationService) SendEmail(to, subject, body string) error {
	if !n.config.Enabled {
		// Just log the email if notifications are disabled
		fmt.Printf("Email to: %s\nSubject: %s\nBody: %s\n", to, subject, body)
		return nil
	}

	// Format the email
	from := fmt.Sprintf("%s <%s>", n.config.SenderName, n.config.SenderMail)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, to, subject, body))

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", n.config.SMTPUser, n.config.SMTPPass, n.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", n.config.SMTPHost, n.config.SMTPPort)

	if err := smtp.SendMail(addr, auth, n.config.SenderMail, []string{to}, msg); err != nil {
		return errors.Wrap(err, "error sending email")
	}

	return nil
}

// SendTalkConfirmationToSpeaker sends a confirmation email to a talk speaker
func (n *NotificationService) SendTalkConfirmationToSpeaker(ctx context.Context, talkID int) error {
	talk, err := n.talkRepo.FindByID(ctx, talkID)
	if err != nil {
		return errors.Wrapf(err, "error finding talk %d", talkID)
	}

	if talk.Speaker == nil {
		// Load speaker info if not already loaded
		speaker, err := n.userRepo.FindByID(ctx, talk.SpeakerID)
		if err != nil {
			return errors.Wrapf(err, "error finding speaker %d", talk.SpeakerID)
		}
		talk.Speaker = speaker
	}

	if talk.ScheduledDate == nil {
		return errors.New("talk has no scheduled date")
	}

	dateStr := talk.ScheduledDate.Format("Monday, January 2, 2006")
	subject := fmt.Sprintf("Your talk '%s' is confirmed for %s", talk.Title, dateStr)

	body := fmt.Sprintf(`Hello %s,

Your talk "%s" has been scheduled for %s.

Please prepare your materials and let us know if you need any special accommodations.

Friday Talks Team
`, talk.Speaker.Name, talk.Title, dateStr)

	return n.SendEmail(talk.Speaker.Email, subject, body)
}

// SendTalkAnnouncementToAll sends an announcement about a new scheduled talk to all users
func (n *NotificationService) SendTalkAnnouncementToAll(ctx context.Context, talkID int) error {
	talk, err := n.talkRepo.FindByID(ctx, talkID)
	if err != nil {
		return errors.Wrapf(err, "error finding talk %d", talkID)
	}

	if talk.Speaker == nil {
		// Load speaker info if not already loaded
		speaker, err := n.userRepo.FindByID(ctx, talk.SpeakerID)
		if err != nil {
			return errors.Wrapf(err, "error finding speaker %d", talk.SpeakerID)
		}
		talk.Speaker = speaker
	}

	if talk.ScheduledDate == nil {
		return errors.New("talk has no scheduled date")
	}

	// Get all users
	users, err := n.userRepo.List(ctx)
	if err != nil {
		return errors.Wrap(err, "error listing users")
	}

	dateStr := talk.ScheduledDate.Format("Monday, January 2, 2006")
	subject := fmt.Sprintf("New Friday Talk: '%s' on %s", talk.Title, dateStr)

	// Send email to each user
	for _, user := range users {
		// Skip the speaker (they already got a confirmation)
		if user.ID == talk.SpeakerID {
			continue
		}

		body := fmt.Sprintf(`Hello %s,

A new Friday Talk has been scheduled:

Title: %s
Speaker: %s
Date: %s

%s

Please let us know if you'll be attending!

Friday Talks Team
`, user.Name, talk.Title, talk.Speaker.Name, dateStr, talk.Description)

		if err := n.SendEmail(user.Email, subject, body); err != nil {
			// Log the error but continue with other users
			fmt.Printf("Error sending email to %s: %v\n", user.Email, err)
		}
	}

	return nil
}

// SendReminderToSpeaker sends a reminder to the speaker a day before their talk
func (n *NotificationService) SendReminderToSpeaker(ctx context.Context, talkID int) error {
	talk, err := n.talkRepo.FindByID(ctx, talkID)
	if err != nil {
		return errors.Wrapf(err, "error finding talk %d", talkID)
	}

	if talk.Speaker == nil {
		// Load speaker info if not already loaded
		speaker, err := n.userRepo.FindByID(ctx, talk.SpeakerID)
		if err != nil {
			return errors.Wrapf(err, "error finding speaker %d", talk.SpeakerID)
		}
		talk.Speaker = speaker
	}

	if talk.ScheduledDate == nil {
		return errors.New("talk has no scheduled date")
	}

	dateStr := talk.ScheduledDate.Format("Monday, January 2, 2006")
	subject := fmt.Sprintf("Reminder: Your talk '%s' is tomorrow", talk.Title)

	body := fmt.Sprintf(`Hello %s,

This is a friendly reminder that your talk "%s" is scheduled for tomorrow, %s.

We're looking forward to your presentation!

Friday Talks Team
`, talk.Speaker.Name, talk.Title, dateStr)

	return n.SendEmail(talk.Speaker.Email, subject, body)
}

// SendReminderToAttendees sends a reminder to confirmed attendees a day before a talk
func (n *NotificationService) SendReminderToAttendees(ctx context.Context, talkID int) error {
	talk, err := n.talkRepo.FindByID(ctx, talkID)
	if err != nil {
		return errors.Wrapf(err, "error finding talk %d", talkID)
	}

	if talk.Speaker == nil {
		// Load speaker info if not already loaded
		speaker, err := n.userRepo.FindByID(ctx, talk.SpeakerID)
		if err != nil {
			return errors.Wrapf(err, "error finding speaker %d", talk.SpeakerID)
		}
		talk.Speaker = speaker
	}

	if talk.ScheduledDate == nil {
		return errors.New("talk has no scheduled date")
	}

	// Get all confirmed attendees
	attendances, err := n.attendanceRepo.ListByTalk(ctx, talkID)
	if err != nil {
		return errors.Wrapf(err, "error listing attendances for talk %d", talkID)
	}

	dateStr := talk.ScheduledDate.Format("Monday, January 2, 2006")
	subject := fmt.Sprintf("Reminder: '%s' talk is tomorrow", talk.Title)

	for _, attendance := range attendances {
		// Only send to confirmed attendees
		if attendance.Status != models.AttendanceStatusConfirmed {
			continue
		}

		// Load user info
		user, err := n.userRepo.FindByID(ctx, attendance.UserID)
		if err != nil {
			fmt.Printf("Error finding user %d: %v\n", attendance.UserID, err)
			continue
		}

		body := fmt.Sprintf(`Hello %s,

This is a friendly reminder that you're confirmed to attend "%s" by %s tomorrow, %s.

We're looking forward to seeing you there!

Friday Talks Team
`, user.Name, talk.Title, talk.Speaker.Name, dateStr)

		if err := n.SendEmail(user.Email, subject, body); err != nil {
			// Log the error but continue with other attendees
			fmt.Printf("Error sending email to %s: %v\n", user.Email, err)
		}
	}

	return nil
}

// SendFeedbackRequest sends a request for feedback after a talk is completed
func (n *NotificationService) SendFeedbackRequest(ctx context.Context, talkID int) error {
	talk, err := n.talkRepo.FindByID(ctx, talkID)
	if err != nil {
		return errors.Wrapf(err, "error finding talk %d", talkID)
	}

	if talk.Speaker == nil {
		// Load speaker info if not already loaded
		speaker, err := n.userRepo.FindByID(ctx, talk.SpeakerID)
		if err != nil {
			return errors.Wrapf(err, "error finding speaker %d", talk.SpeakerID)
		}
		talk.Speaker = speaker
	}

	// Get all attendees
	attendances, err := n.attendanceRepo.ListByTalk(ctx, talkID)
	if err != nil {
		return errors.Wrapf(err, "error listing attendances for talk %d", talkID)
	}

	subject := fmt.Sprintf("Please share your feedback on '%s'", talk.Title)

	for _, attendance := range attendances {
		// Only send to those who attended
		if attendance.Status != models.AttendanceStatusAttended {
			continue
		}

		// Load user info
		user, err := n.userRepo.FindByID(ctx, attendance.UserID)
		if err != nil {
			fmt.Printf("Error finding user %d: %v\n", attendance.UserID, err)
			continue
		}

		body := fmt.Sprintf(`Hello %s,

Thank you for attending "%s" by %s.

We'd love to hear your thoughts! Please take a moment to share your feedback on the talk.

Friday Talks Team
`, user.Name, talk.Title, talk.Speaker.Name)

		if err := n.SendEmail(user.Email, subject, body); err != nil {
			// Log the error but continue with other attendees
			fmt.Printf("Error sending email to %s: %v\n", user.Email, err)
		}
	}

	return nil
}

// SendWeeklyTalkDigest sends a weekly digest of upcoming talks
func (n *NotificationService) SendWeeklyTalkDigest(ctx context.Context) error {
	// Get upcoming talks for the next 4 weeks
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 28) // 4 weeks ahead

	// Get all scheduled talks
	talks, err := n.talkRepo.ListByStatus(ctx, models.TalkStatusScheduled)
	if err != nil {
		return errors.Wrap(err, "error listing scheduled talks")
	}

	// Filter to only include upcoming talks within our time window
	var upcomingTalks []*models.Talk
	for _, talk := range talks {
		if talk.ScheduledDate != nil &&
			!talk.ScheduledDate.Before(startDate) &&
			!talk.ScheduledDate.After(endDate) {
			upcomingTalks = append(upcomingTalks, talk)
		}
	}

	// No upcoming talks, no need to send digest
	if len(upcomingTalks) == 0 {
		return nil
	}

	// Sort talks by date
	sort.Slice(upcomingTalks, func(i, j int) bool {
		return upcomingTalks[i].ScheduledDate.Before(*upcomingTalks[j].ScheduledDate)
	})

	// Build the digest content
	subject := "Friday Talks: Upcoming Schedule"

	// Get all users
	users, err := n.userRepo.List(ctx)
	if err != nil {
		return errors.Wrap(err, "error listing users")
	}

	// Send to each user
	for _, user := range users {
		var talksList string
		for _, talk := range upcomingTalks {
			if talk.Speaker == nil {
				// Load speaker info if not already loaded
				speaker, err := n.userRepo.FindByID(ctx, talk.SpeakerID)
				if err != nil {
					return errors.Wrapf(err, "error finding speaker %d", talk.SpeakerID)
				}
				talk.Speaker = speaker
			}

			dateStr := talk.ScheduledDate.Format("Monday, January 2")
			talksList += fmt.Sprintf("- %s: '%s' by %s\n", dateStr, talk.Title, talk.Speaker.Name)
		}

		body := fmt.Sprintf(`Hello %s,

Here's your weekly update on upcoming Friday Talks:

%s

To change your attendance status or see more details, please visit the website.

Friday Talks Team
`, user.Name, talksList)

		if err := n.SendEmail(user.Email, subject, body); err != nil {
			// Log the error but continue with other users
			fmt.Printf("Error sending digest to %s: %v\n", user.Email, err)
		}
	}

	return nil
}
