package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"os"
	"time"
)

type GetTicketsCommand struct {
	*cmds.CommandDescription
}

func NewGetTicketsCommand() (*GetTicketsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &GetTicketsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get-tickets",
			cmds.WithShort("Fetch tickets from Zendesk"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"start-time",
					parameters.ParameterTypeDate,
					parameters.WithHelp("Specify the start time from when you want to start fetching tickets."),
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
					"id",
					parameters.ParameterTypeString,
					parameters.WithHelp("Specify a ticket ID to fetch."),
				),
				parameters.NewParameterDefinition(
					"limit",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Limit the number of tickets to fetch."),
					parameters.WithDefault(0),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *GetTicketsCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	// Extract flags from ps map
	startTime_, ok := ps["start-time"]

	var startTime time.Time
	if ok {
		startTime = startTime_.(time.Time)
	} else {
		// set to 2010-01-01 per default
		startTime = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
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

	ticketId_, ok := ps["id"]
	var ticketId string
	if ok {
		ticketId = ticketId_.(string)
	}
	limit := ps["limit"].(int)

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

	// Set up the ZendeskConfig with the parsed flags
	zd := &ZendeskConfig{
		Domain:   domain,
		Email:    email,
		ApiToken: apiToken,
	}

	// Logic for fetching ticket data
	var tickets []Ticket
	if ticketId != "" {
		ticket := zd.getTicketById(ticketId)
		tickets = append(tickets, ticket)
	} else {
		// Convert startTime to Unix timestamp, as required by getIncrementalTickets
		date := startTime.Unix()
		tickets = zd.getIncrementalTickets(date, limit)
	}

	count := 0
	// Create and add rows of ticket data using the GlazeProcessor
	for _, ticket := range tickets {
		row := types.NewRow(
			types.MRP("id", ticket.ID),
			types.MRP("status", ticket.Status),
			types.MRP("created_at", ticket.CreatedAt),
			types.MRP("subject", ticket.Subject),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
		count++
		if count >= limit {
			break
		}
	}

	return nil
}
