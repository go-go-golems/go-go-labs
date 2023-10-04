package main

import (
	"encoding/base64"
	"fmt"
	"github.com/levigross/grequests"
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

// getIncrementalTickets fetches tickets that changed since the provided startTime using cursor-based incremental exports.
func (zd *ZendeskConfig) getIncrementalTickets(startTime int64, limit int) []Ticket {
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
			panic(err)
		}

		//if count == 0 {
		//	var result struct {
		//		Tickets     []interface{} `json:"tickets"`
		//		AfterCursor string        `json:"after_cursor"`
		//		EndOfStream bool          `json:"end_of_stream"`
		//	}
		//	err := response.JSON(&result)
		//	if err != nil {
		//		panic(err)
		//	}
		//	fmt.Println(result.Tickets[0])
		//	continue
		//}

		var result struct {
			Tickets     []Ticket `json:"tickets"`
			AfterCursor string   `json:"after_cursor"`
			EndOfStream bool     `json:"end_of_stream"`
		}
		if err := response.JSON(&result); err != nil {
			panic(err)
		}

		allTickets = append(allTickets, result.Tickets...)

		count += len(result.Tickets)
		if limit > 0 && count >= limit {
			break
		}

		// Check if end of stream is reached or if there's no cursor for the next set
		if result.EndOfStream || result.AfterCursor == "" {
			break
		}

		endpoint = fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?cursor=%s", zd.Domain, result.AfterCursor)
	}

	return allTickets
}

func basicAuth(username, password string) string {
	auth := username + "/token:" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
