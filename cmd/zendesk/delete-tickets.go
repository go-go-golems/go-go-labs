package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/clay/pkg/workerpool"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"io"
	"os"
	"time"
)

type DeleteTicketsCommand struct {
	*cmds.CommandDescription
}

func NewDeleteTicketsCommand() (*DeleteTicketsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &DeleteTicketsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete-tickets",
			cmds.WithShort("Delete tickets in Zendesk"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"ids",
					parameters.ParameterTypeIntegerList,
					parameters.WithHelp("List of ticket IDs to delete."),
				),
				parameters.NewParameterDefinition(
					"tickets-file",
					parameters.ParameterTypeObjectListFromFile,
					parameters.WithHelp("File containing a list of tickets to delete."),
				),
				parameters.NewParameterDefinition(
					"domain",
					parameters.ParameterTypeString,
					parameters.WithHelp("Zendesk domain."),
				),
				parameters.NewParameterDefinition(
					"email",
					parameters.ParameterTypeString,
					parameters.WithHelp("Zendesk email."),
				),
				parameters.NewParameterDefinition(
					"api-token",
					parameters.ParameterTypeString,
					parameters.WithHelp("Zendesk API token."),
				),
				parameters.NewParameterDefinition(
					"workers",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Number of workers to use."),
					parameters.WithDefault(8),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *DeleteTicketsCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	w io.Writer,
) error {
	// Extract flags from ps map
	ticketIds_, ok := ps["ids"]
	var ticketIds []int
	if ok {
		ticketIds = ticketIds_.([]int)
	}

	ticketsFromFile_, ok := ps["tickets-file"]
	var ticketsFromFile []interface{}
	if ok {
		ticketsFromFile = ticketsFromFile_.([]interface{})
		for _, ticket_ := range ticketsFromFile {
			ticket, ok := ticket_.(map[string]interface{})
			if !ok {
				return fmt.Errorf("could not convert ticket to map[string]interface{}")
			}
			ticketId, ok := ticket["id"].(float64)
			if !ok {
				return fmt.Errorf("could not convert ticket ID to int")
			}
			ticketIds = append(ticketIds, int(ticketId))
		}
	}

	if len(ticketIds) == 0 {
		return fmt.Errorf("no ticket IDs specified")
	}

	domain_, ok := ps["domain"]
	var domain string
	if ok {
		domain = domain_.(string)
	}
	email_, ok := ps["email"]
	var email string
	if ok {
		email = email_.(string)
	}
	apiToken_, ok := ps["api-token"]
	var apiToken string
	if ok {
		apiToken = apiToken_.(string)
	}

	// If flags are not set, use environment variables
	if domain == "" {
		domain = os.Getenv("ZENDESK_DOMAIN")
	}
	if email == "" {
		email = os.Getenv("ZENDESK_EMAIL")
	}
	if apiToken == "" {
		apiToken = os.Getenv("ZENDESK_API_TOKEN")
	}

	zd := &ZendeskConfig{
		Domain:   domain,
		Email:    email,
		ApiToken: apiToken,
	}

	// split ticketIds in group of 100 tickets
	var ticketIdGroups [][]int
	for i := 0; i < len(ticketIds); i += 100 {
		end := i + 100
		if end > len(ticketIds) {
			end = len(ticketIds)
		}
		ticketIdGroups = append(ticketIdGroups, ticketIds[i:end])
	}

	workers := ps["workers"].(int)
	fmt.Printf("Using %d workers\n", workers)
	pool := workerpool.New(workers) // For example, 5 workers
	pool.Start()

	total := 0
	for _, ticketIdGroup := range ticketIdGroups {
		ticketIdGroupCopy := ticketIdGroup // Create a copy to avoid closure over the loop variable
		job := func() error {
			firstId := ticketIdGroupCopy[0]
			lastId := ticketIdGroupCopy[len(ticketIdGroupCopy)-1]
			fmt.Printf("About to delete %d (%d/%d) tickets, first: %d, last: %d\n",
				len(ticketIdGroupCopy),
				total, len(ticketIds),
				firstId, lastId)

			jobStatus, err := zd.bulkDeleteTickets(ticketIdGroupCopy)
			if err != nil {
				return err
			}

			retryCnt := 0
			// poll jobStatus and print to the writer (assuming w is your writer)
			for {
				jobStatus, err := zd.getJobStatus(jobStatus.ID)
				if err != nil {
					if retryCnt < 3 {
						retryCnt++
						continue
					}
					return errors.Wrapf(err, "failed to get job status for job ID %s", jobStatus.ID)
				}

				_, _ = fmt.Fprintf(w, "Job ID: %s, Status: %s, Progress: %d%%, Message: %s\n",
					jobStatus.ID, jobStatus.Status, jobStatus.Progress, jobStatus.Message)

				if jobStatus.Status == "completed" || jobStatus.Status == "failed" {
					total += len(ticketIdGroupCopy)
					break
				}
				time.Sleep(time.Second) // Optional: Add a sleep for regular intervals to avoid aggressive polling
			}

			return nil
		}

		pool.AddJob(job)
	}
	pool.Close()

	return nil
}
