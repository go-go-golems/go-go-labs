package services

import (
	"context"
	"sort"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/models"
	"github.com/pkg/errors"
)

// SchedulerService handles talk scheduling logic
type SchedulerService struct {
	talkRepo       models.TalkRepository
	voteRepo       models.VoteRepository
	attendanceRepo models.AttendanceRepository
}

// NewSchedulerService creates a new SchedulerService
func NewSchedulerService(
	talkRepo models.TalkRepository,
	voteRepo models.VoteRepository,
	attendanceRepo models.AttendanceRepository,
) *SchedulerService {
	return &SchedulerService{
		talkRepo:       talkRepo,
		voteRepo:       voteRepo,
		attendanceRepo: attendanceRepo,
	}
}

// TalkRanking represents a talk with its computed ranking score
type TalkRanking struct {
	Talk           *models.Talk
	InterestScore  int     // Sum of interest levels
	AvailableUsers int     // Number of users available on the proposed date
	FinalScore     float64 // Calculated final score for ranking
}

// FindNextFriday returns the date of the next Friday after the given date
func FindNextFriday(after time.Time) time.Time {
	// Start from the day after the given date
	date := after.AddDate(0, 0, 1)

	// Find the next Friday (weekday 5)
	daysUntilFriday := (5 - int(date.Weekday()) + 7) % 7

	return date.AddDate(0, 0, daysUntilFriday)
}

// GetUpcomingFridays returns a list of upcoming Fridays
func (s *SchedulerService) GetUpcomingFridays(weeks int) []time.Time {
	fridays := make([]time.Time, weeks)

	// Start with next Friday
	friday := FindNextFriday(time.Now())

	for i := 0; i < weeks; i++ {
		fridays[i] = friday
		friday = friday.AddDate(0, 0, 7) // Add a week
	}

	return fridays
}

// IsDateScheduled checks if a date already has a scheduled talk
func (s *SchedulerService) IsDateScheduled(ctx context.Context, date time.Time) (bool, error) {
	// Get all scheduled talks
	talks, err := s.talkRepo.ListByStatus(ctx, models.TalkStatusScheduled)
	if err != nil {
		return false, errors.Wrap(err, "error listing scheduled talks")
	}

	// Format the date string for consistent comparison (without time component)
	dateStr := date.Format("2006-01-02")

	// Check if any scheduled talk falls on the given date
	for _, talk := range talks {
		if talk.ScheduledDate != nil {
			talkDateStr := talk.ScheduledDate.Format("2006-01-02")
			if talkDateStr == dateStr {
				return true, nil
			}
		}
	}

	return false, nil
}

// FindBestTalksForDate finds the best candidate talks for a given date
func (s *SchedulerService) FindBestTalksForDate(ctx context.Context, date time.Time) ([]*TalkRanking, error) {
	// 1. Find all proposed talks that have this date as a preferred date
	candidateTalks, err := s.talkRepo.FindProposedWithPreferredDate(ctx, date)
	if err != nil {
		return nil, errors.Wrap(err, "error finding talks with preferred date")
	}

	if len(candidateTalks) == 0 {
		return nil, nil
	}

	// 2. Rank the talks
	rankedTalks, err := s.rankTalks(ctx, candidateTalks, date)
	if err != nil {
		return nil, errors.Wrap(err, "error ranking talks")
	}

	// 3. Sort by final score (descending)
	sort.Slice(rankedTalks, func(i, j int) bool {
		return rankedTalks[i].FinalScore > rankedTalks[j].FinalScore
	})

	return rankedTalks, nil
}

// rankTalks ranks talks based on various factors
func (s *SchedulerService) rankTalks(ctx context.Context, talks []*models.Talk, date time.Time) ([]*TalkRanking, error) {
	var rankings []*TalkRanking

	for _, talk := range talks {
		// Get interest level for this talk
		interestCount, err := s.voteRepo.GetTalkInterestCount(ctx, talk.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting interest count for talk %d", talk.ID)
		}

		// Get availability for this date
		availability, err := s.voteRepo.GetAvailabilityForDate(ctx, talk.ID, date)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting availability for talk %d", talk.ID)
		}

		// Calculate available users
		availableUsers := 0
		for _, available := range availability {
			if available {
				availableUsers++
			}
		}

		// Calculate final score - we prioritize talks with more interest and availability
		// The exact formula can be adjusted based on specific requirements
		finalScore := float64(interestCount*2) + float64(availableUsers)

		rankings = append(rankings, &TalkRanking{
			Talk:           talk,
			InterestScore:  interestCount,
			AvailableUsers: availableUsers,
			FinalScore:     finalScore,
		})
	}

	return rankings, nil
}

// SuggestNextFriday finds the best date and talk for the next Friday sessions
func (s *SchedulerService) SuggestNextFriday(ctx context.Context, weeksAhead int) (time.Time, []*TalkRanking, error) {
	// Get upcoming Fridays
	fridays := s.GetUpcomingFridays(weeksAhead)

	var bestDate time.Time
	var bestTalks []*TalkRanking
	var bestScore float64 = -1

	// For each Friday, check if it's already scheduled and find the best talks
	for _, friday := range fridays {
		// Skip if already scheduled
		scheduled, err := s.IsDateScheduled(ctx, friday)
		if err != nil {
			return time.Time{}, nil, errors.Wrapf(err, "error checking if date is scheduled: %s", friday.Format("2006-01-02"))
		}

		if scheduled {
			continue
		}

		// Find the best talks for this date
		rankedTalks, err := s.FindBestTalksForDate(ctx, friday)
		if err != nil {
			return time.Time{}, nil, errors.Wrapf(err, "error finding best talks for date: %s", friday.Format("2006-01-02"))
		}

		// Skip if no talks available
		if len(rankedTalks) == 0 {
			continue
		}

		// Use the highest scored talk to compare dates
		topScore := rankedTalks[0].FinalScore

		// If this date has a better top talk, use it
		if topScore > bestScore {
			bestScore = topScore
			bestDate = friday
			bestTalks = rankedTalks
		}
	}

	// If we found a suitable date and talk
	if !bestDate.IsZero() && len(bestTalks) > 0 {
		return bestDate, bestTalks, nil
	}

	// If no ideal match was found, just return the next available Friday and any proposed talks
	for _, friday := range fridays {
		scheduled, err := s.IsDateScheduled(ctx, friday)
		if err != nil {
			return time.Time{}, nil, errors.Wrapf(err, "error checking if date is scheduled: %s", friday.Format("2006-01-02"))
		}

		if !scheduled {
			// Get any proposed talks
			talks, err := s.talkRepo.ListByStatus(ctx, models.TalkStatusProposed)
			if err != nil {
				return time.Time{}, nil, errors.Wrap(err, "error listing proposed talks")
			}

			var rankedTalks []*TalkRanking
			for _, talk := range talks {
				rankedTalks = append(rankedTalks, &TalkRanking{
					Talk:           talk,
					InterestScore:  0,
					AvailableUsers: 0,
					FinalScore:     0,
				})
			}

			return friday, rankedTalks, nil
		}
	}

	return time.Time{}, nil, nil
}

// ScheduleTalk schedules a talk for a specific date
func (s *SchedulerService) ScheduleTalk(ctx context.Context, talkID int, date time.Time) error {
	// Get the talk
	talk, err := s.talkRepo.FindByID(ctx, talkID)
	if err != nil {
		return errors.Wrapf(err, "error finding talk %d", talkID)
	}

	// Update the talk status and scheduled date
	talk.Status = models.TalkStatusScheduled
	talk.ScheduledDate = &date

	// Save the changes
	if err := s.talkRepo.Update(ctx, talk); err != nil {
		return errors.Wrapf(err, "error updating talk %d", talkID)
	}

	return nil
}
