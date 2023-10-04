package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/levigross/grequests"
	"github.com/spf13/cobra"
)

type ZendeskConfig struct {
	Domain   string
	Email    string
	Password string
	ApiToken string
}

func main() {
	var ticketId string

	var rootCmd = &cobra.Command{
		Use:   "zendesk",
		Short: "Zendesk fetcher",
		Run: func(cmd *cobra.Command, args []string) {
			zd := &ZendeskConfig{
				Domain:   os.Getenv("ZENDESK_DOMAIN"),
				Email:    os.Getenv("ZENDESK_EMAIL"),
				Password: os.Getenv("ZENDESK_PASSWORD"),
				ApiToken: os.Getenv("ZENDESK_API_TOKEN"),
			}

			if ticketId != "" {
				// Handle fetching a single ticket using ticketId.
				// You'll need to implement `zd.getTicketById(ticketId)`
				ticket := zd.getTicketById(ticketId)
				fmt.Printf("ID: %d, Status: %s, Created At: %s, Subject: %s\n", ticket.ID, ticket.Status, ticket.CreatedAt, ticket.Subject)
			} else {
				// Provide the appropriate start time (Unix epoch time) from when you want to start fetching.
				date := time.Now().AddDate(-4, 0, 0)
				tickets := zd.getIncrementalTickets(date.Unix())

				for _, ticket := range tickets {
					fmt.Printf("ID: %d, Status: %s, Created At: %s, Subject: %s\n", ticket.ID, ticket.Status, ticket.CreatedAt, ticket.Subject)
				}
			}
		},
	}

	rootCmd.Flags().StringVarP(&ticketId, "id", "i", "", "Specify a ticket ID to fetch")
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

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
	ID        int    `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Subject   string `json:"subject"`
}

// getIncrementalTickets fetches tickets that changed since the provided startTime using cursor-based incremental exports.
func (zd *ZendeskConfig) getIncrementalTickets(startTime int64) []Ticket {
	endpoint := fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?start_time=%d", zd.Domain, startTime)

	headers := map[string]string{
		"Authorization": "Basic " + basicAuth(zd.Email, zd.ApiToken),
	}

	var allTickets []Ticket

	for {
		response, err := grequests.Get(endpoint, &grequests.RequestOptions{
			Headers: headers,
		})
		if err != nil {
			panic(err)
		}

		var result struct {
			Tickets     []Ticket `json:"tickets"`
			AfterCursor string   `json:"after_cursor"`
			EndOfStream bool     `json:"end_of_stream"`
		}
		if err := response.JSON(&result); err != nil {
			panic(err)
		}

		allTickets = append(allTickets, result.Tickets...)

		// Check if end of stream is reached or if there's no cursor for the next set
		if result.EndOfStream || result.AfterCursor == "" {
			break
		}

		endpoint = fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?cursor=%s", zd.Domain, result.AfterCursor)
	}

	return allTickets
}

func (zd *ZendeskConfig) getOldTickets() []Ticket {
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	endpoint := fmt.Sprintf("%s/api/v2/tickets.json", zd.Domain)

	headers := map[string]string{
		"Authorization": "Basic " + basicAuth(zd.Email, zd.ApiToken),
	}

	var tickets []Ticket

	i := 0

	for {
		response, err := grequests.Get(endpoint, &grequests.RequestOptions{
			Headers: headers,
			// add sort=updated_at as query parameter
			Params: map[string]string{
				"sort_by":    "created_at",
				"sort_order": "asc",
			},
		})
		if err != nil {
			panic(err)
		}

		var result struct {
			Tickets  []Ticket `json:"tickets"`
			NextPage string   `json:"next_page"`
		}
		if err := response.JSON(&result); err != nil {
			panic(err)
		}

		for _, ticket := range result.Tickets {
			createdAt, err := time.Parse(time.RFC3339, ticket.CreatedAt)
			if err != nil {
				panic(err)
			}

			if createdAt.Before(oneYearAgo) {
				tickets = append(tickets, ticket)
			} else {
				tickets = append(tickets, ticket)
			}

		}

		if result.NextPage == "" {
			break
		}

		endpoint = result.NextPage
		i++

		if i > 5 {
			break
		}
	}

	return tickets
}

func basicAuth(username, password string) string {
	auth := username + "/token:" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
