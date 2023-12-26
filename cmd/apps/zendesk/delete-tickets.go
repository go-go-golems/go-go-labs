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

var _ cmds.WriterCommand = (*DeleteTicketsCommand)(nil)

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

type DeleteTicketSettings struct {
	Ids         []int                    `glazed.parameter:"ids"`
	TicketsFile []map[string]interface{} `glazed.parameter:"tickets-file"`
	Domain      string                   `glazed.parameter:"domain"`
	Email       string                   `glazed.parameter:"email"`
	APIToken    string                   `glazed.parameter:"api-token"`
	Workers     int                      `glazed.parameter:"workers"`
}

func (c *DeleteTicketsCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &DeleteTicketSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	for _, ticket_ := range s.TicketsFile {
		ticketId, ok := ticket_["id"].(float64)
		if !ok {
			return fmt.Errorf("could not convert ticket ID to int")
		}
		s.Ids = append(s.Ids, int(ticketId))
	}

	if len(s.Ids) == 0 {
		return fmt.Errorf("no ticket IDs specified")
	}

	// If flags are not set, use environment variables
	if s.Domain == "" {
		s.Domain = os.Getenv("ZENDESK_DOMAIN")
	}
	if s.Email == "" {
		s.Email = os.Getenv("ZENDESK_EMAIL")
	}
	if s.APIToken == "" {
		s.APIToken = os.Getenv("ZENDESK_API_TOKEN")
	}

	zd := &ZendeskConfig{
		Domain:   s.Domain,
		Email:    s.Email,
		ApiToken: s.APIToken,
	}

	// split ticketIds in group of 100 tickets
	var ticketIdGroups [][]int
	for i := 0; i < len(s.Ids); i += 100 {
		end := i + 100
		if end > len(s.Ids) {
			end = len(s.Ids)
		}
		ticketIdGroups = append(ticketIdGroups, s.Ids[i:end])
	}

	fmt.Printf("Using %d workers\n", s.Workers)
	pool := workerpool.New(s.Workers) // For example, 5 workers
	pool.Start()

	total := 0
	for _, ticketIdGroup := range ticketIdGroups {
		ticketIdGroupCopy := ticketIdGroup // Create a copy to avoid closure over the loop variable
		job := func() error {
			firstId := ticketIdGroupCopy[0]
			lastId := ticketIdGroupCopy[len(ticketIdGroupCopy)-1]
			fmt.Printf("About to delete %d (%d/%d) tickets, first: %d, last: %d\n",
				len(ticketIdGroupCopy),
				total, len(s.Ids),
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
