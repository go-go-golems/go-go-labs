package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
)

func (zd *ZendeskConfig) getTicketById(id string) Ticket {
	endpoint := fmt.Sprintf("%s/api/v2/tickets/%s.json", zd.Domain, id)

	headers := map[string]string{
		"Authorization": "Basic " + basicAuth(zd.Email, zd.ApiToken),
	}

	response, err := grequests.Get(endpoint, &grequests.RequestOptions{
		Headers: headers,
	})
	if err != nil {
		panic(err)
	}

	var result struct {
		Ticket Ticket `json:"ticket"`
	}
	if err := response.JSON(&result); err != nil {
		panic(err)
	}

	return result.Ticket
}

type Ticket struct {
	ID                   int                `json:"id"`
	Status               string             `json:"status"`
	CreatedAt            string             `json:"created_at"`
	Subject              string             `json:"subject"`
	AllowAttachments     bool               `json:"allow_attachments"`
	AllowChannelback     bool               `json:"allow_channelback"`
	AssigneeID           int                `json:"assignee_id"`
	BrandID              int                `json:"brand_id"`
	CollaboratorIDs      []int              `json:"collaborator_ids"`
	CustomFields         []CustomField      `json:"custom_fields"`
	CustomStatusID       int                `json:"custom_status_id"`
	Description          string             `json:"description"`
	DueAt                *string            `json:"due_at"`
	EmailCCIDs           []int              `json:"email_cc_ids"`
	ExternalID           *string            `json:"external_id"`
	Fields               []Field            `json:"fields"`
	FollowerIDs          []int              `json:"follower_ids"`
	FollowupIDs          []int              `json:"followup_ids"`
	ForumTopicID         *string            `json:"forum_topic_id"`
	FromMessagingChannel bool               `json:"from_messaging_channel"`
	GeneratedTimestamp   float64            `json:"generated_timestamp"`
	GroupID              int                `json:"group_id"`
	HasIncidents         bool               `json:"has_incidents"`
	IsPublic             bool               `json:"is_public"`
	OrganizationID       *int               `json:"organization_id"`
	Priority             *string            `json:"priority"`
	ProblemID            *int               `json:"problem_id"`
	RawSubject           string             `json:"raw_subject"`
	Recipient            *string            `json:"recipient"`
	RequesterID          int                `json:"requester_id"`
	SatisfactionRating   SatisfactionRating `json:"satisfaction_rating"`
	SharingAgreementIDs  []int              `json:"sharing_agreement_ids"`
	SubmitterID          int                `json:"submitter_id"`
	Tags                 []string           `json:"tags"`
	Type                 *string            `json:"type"`
	UpdatedAt            string             `json:"updated_at"`
	URL                  string             `json:"url"`
	Via                  Via                `json:"via"`
}

type CustomField struct {
	ID    int     `json:"id"`
	Value *string `json:"value"`
}

type Field struct {
	ID    int     `json:"id"`
	Value *string `json:"value"`
}

type SatisfactionRating struct {
	Comment string `json:"comment"`
	ID      int    `json:"id"`
	Score   string `json:"score"`
}

type Via struct {
	Channel string `json:"channel"`
	Source  Source `json:"source"`
}

type Source struct {
	From map[string]interface{} `json:"from"`
	Rel  *string                `json:"rel"`
	To   map[string]interface{} `json:"to"`
}

type Query struct {
	StartDate time.Time
	EndDate   time.Time
	Limit     int
	Callback  func(Ticket) error
}

// getIncrementalTickets fetches tickets that changed since the provided startTime using cursor-based incremental exports.
func (zd *ZendeskConfig) getIncrementalTickets(query Query) ([]Ticket, error) {
	startTime := query.StartDate.Unix()
	endpoint := fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?start_time=%d", zd.Domain, startTime)

	headers := map[string]string{
		"Authorization": "Basic " + basicAuth(zd.Email, zd.ApiToken),
	}

	var allTickets []Ticket
	count := 0

	for {
		response, err := grequests.Get(endpoint, &grequests.RequestOptions{
			Headers: headers,
		})
		if err != nil {
			return nil, err
		}

		var result struct {
			Tickets     []Ticket `json:"tickets"`
			AfterCursor string   `json:"after_cursor"`
			EndOfStream bool     `json:"end_of_stream"`
		}
		if err := response.JSON(&result); err != nil {
			return nil, err
		}

		oneSkipped := false
		for _, ticket := range result.Tickets {
			createdAt, err1 := time.Parse(time.RFC3339, ticket.CreatedAt)
			updatedAt, err2 := time.Parse(time.RFC3339, ticket.UpdatedAt)

			if err1 != nil || err2 != nil {
				// Handle the error, e.g., skip the ticket or return an error
				continue
			}

			if createdAt.Before(query.StartDate) && updatedAt.Before(query.StartDate) {
				continue
			}
			if createdAt.After(query.EndDate) && updatedAt.After(query.EndDate) {
				oneSkipped = true
				continue
			}

			if query.Callback != nil {
				err := query.Callback(ticket)
				if err != nil {
					return nil, err
				}
				continue
			}

			allTickets = append(allTickets, ticket)
		}

		count += len(result.Tickets)
		if query.Limit > 0 && count >= query.Limit {
			break
		}

		if oneSkipped {
			log.Warn().Msg("One or more tickets were skipped due to being outside the specified time range")
			break
		}

		// Check if end of stream is reached or if there's no cursor for the next set
		if result.EndOfStream || result.AfterCursor == "" {
			break
		}

		endpoint = fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?cursor=%s", zd.Domain, result.AfterCursor)
	}

	return allTickets, nil
}

func basicAuth(username, password string) string {
	auth := username + "/token:" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type JobStatus struct {
	ID       string `json:"id"`
	Message  string `json:"message"`
	Progress int    `json:"progress"`
	Status   string `json:"status"`
	Total    int    `json:"total"`
	URL      string `json:"url"`
}

func (zd *ZendeskConfig) getJobStatus(jobID string) (*JobStatus, error) {
	endpoint := fmt.Sprintf("%s/api/v2/job_statuses/%s.json", zd.Domain, jobID)
	headers := map[string]string{
		"Authorization": "Basic " + basicAuth(zd.Email, zd.ApiToken),
	}

	for {
		response, err := grequests.Get(endpoint, &grequests.RequestOptions{
			Headers: headers,
		})
		if err != nil {
			return nil, err
		}

		if response.StatusCode != 200 {
			return nil, fmt.Errorf("failed to fetch job status. Status: %s", response.String())
		}

		var result struct {
			JobStatus JobStatus `json:"job_status"`
		}

		if err := json.Unmarshal(response.Bytes(), &result); err != nil {
			return nil, err
		}

		switch result.JobStatus.Status {
		case "completed", "failed":
			return &result.JobStatus, nil
		case "queued", "working":
			// Polling interval. Adjust as needed.
			time.Sleep(5 * time.Second)
		default:
			return nil, errors.New("unexpected job status")
		}
	}
}

func (zd *ZendeskConfig) bulkDeleteTickets(ticketIds []int) (*JobStatus, error) {
	endpoint := fmt.Sprintf("%s/api/v2/tickets/destroy_many.json?ids=%s", zd.Domain, strings.Join(convertIntsToStrings(ticketIds), ","))

	headers := map[string]string{
		"Authorization": "Basic " + basicAuth(zd.Email, zd.ApiToken),
	}

	response, err := grequests.Delete(endpoint, &grequests.RequestOptions{
		Headers: headers,
	})
	if err != nil {
		return nil, err
	}

	// Check and handle rate limit
	if response.StatusCode == 429 {
		retryAfter, _ := strconv.Atoi(response.Header.Get("Retry-After"))
		fmt.Printf("Rate limit exceeded. Retrying after %d seconds.\n", retryAfter)
		time.Sleep(time.Duration(retryAfter) * time.Second)
		return zd.bulkDeleteTickets(ticketIds)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to bulk delete tickets. Status: %s", response.String())
	}

	var respBody struct {
		JobStatus *JobStatus `json:"job_status"`
	}
	if err := json.Unmarshal(response.Bytes(), &respBody); err != nil {
		return nil, err
	}

	return respBody.JobStatus, nil
}

func convertIntsToStrings(ints []int) []string {
	strs := make([]string, len(ints))
	for i, v := range ints {
		strs[i] = fmt.Sprintf("%d", v)
	}
	return strs
}
