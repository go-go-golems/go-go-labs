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

var _ cmds.GlazeCommand = (*GetTicketsCommand)(nil)

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
					"start-date",
					parameters.ParameterTypeDate,
					parameters.WithHelp("Specify the start time from when you want to start fetching tickets."),
					parameters.WithDefault(time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)),
				),
				parameters.NewParameterDefinition(
					"end-date",
					parameters.ParameterTypeDate,
					parameters.WithHelp("Specify the end time until when you want to fetch tickets."),
					parameters.WithDefault(time.Now()),
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

type GetSettings struct {
	StartDate time.Time `glazed.parameter:"start-date"`
	EndDate   time.Time `glazed.parameter:"end-date"`
	Domain    string    `glazed.parameter:"domain"`
	Email     string    `glazed.parameter:"email"`
	ApiToken  string    `glazed.parameter:"api-token"`
	Id        string    `glazed.parameter:"id"`
	Limit     int       `glazed.parameter:"limit"`
}

type ErrFinish struct{}

func (e ErrFinish) Error() string {
	return "finish"
}

func (c *GetTicketsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &GetSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	// If flags are not set, use environment variables
	if s.Domain == "" {
		s.Domain = os.Getenv("ZENDESK_DOMAIN")
	}
	if s.Email == "" {
		s.Email = os.Getenv("ZENDESK_EMAIL")
	}
	if s.ApiToken == "" {
		s.ApiToken = os.Getenv("ZENDESK_API_TOKEN")
	}

	// Set up the ZendeskConfig with the parsed flags
	zd := &ZendeskConfig{
		Domain:   s.Domain,
		Email:    s.Email,
		ApiToken: s.ApiToken,
	}

	count := 0

	addTicketRow := func(ticket Ticket) error {
		row := types.NewRow(
			types.MRP("id", ticket.ID),
			types.MRP("status", ticket.Status),
			types.MRP("created_at", ticket.CreatedAt),
			types.MRP("subject", ticket.Subject),
			types.MRP("allow_attachments", ticket.AllowAttachments),
			types.MRP("allow_channelback", ticket.AllowChannelback),
			types.MRP("assignee_id", ticket.AssigneeID),
			types.MRP("brand_id", ticket.BrandID),
			types.MRP("collaborator_ids", ticket.CollaboratorIDs),
			types.MRP("custom_fields", ticket.CustomFields),
			types.MRP("custom_status_id", ticket.CustomStatusID),
			types.MRP("description", ticket.Description),
			types.MRP("due_at", ticket.DueAt),
			types.MRP("email_cc_ids", ticket.EmailCCIDs),
			types.MRP("external_id", ticket.ExternalID),
			types.MRP("fields", ticket.Fields),
			types.MRP("follower_ids", ticket.FollowerIDs),
			types.MRP("followup_ids", ticket.FollowupIDs),
			types.MRP("forum_topic_id", ticket.ForumTopicID),
			types.MRP("from_messaging_channel", ticket.FromMessagingChannel),
			types.MRP("generated_timestamp", ticket.GeneratedTimestamp),
			types.MRP("group_id", ticket.GroupID),
			types.MRP("has_incidents", ticket.HasIncidents),
			types.MRP("is_public", ticket.IsPublic),
			types.MRP("organization_id", ticket.OrganizationID),
			types.MRP("priority", ticket.Priority),
			types.MRP("problem_id", ticket.ProblemID),
			types.MRP("raw_subject", ticket.RawSubject),
			types.MRP("recipient", ticket.Recipient),
			types.MRP("requester_id", ticket.RequesterID),
			types.MRP("satisfaction_rating", ticket.SatisfactionRating),
			types.MRP("sharing_agreement_ids", ticket.SharingAgreementIDs),
			types.MRP("submitter_id", ticket.SubmitterID),
			types.MRP("tags", ticket.Tags),
			types.MRP("type", ticket.Type),
			types.MRP("updated_at", ticket.UpdatedAt),
			types.MRP("url", ticket.URL),
			types.MRP("via", ticket.Via),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}

		count++

		if s.Limit > 0 && count >= s.Limit {
			return ErrFinish{}
		}
		return nil
	}

	if s.Id != "" {
		ticket := zd.getTicketById(s.Id)
		err := addTicketRow(ticket)
		if err != nil {
			return err
		}
	} else {
		_, err := zd.getIncrementalTickets(Query{
			StartDate: s.StartDate,
			EndDate:   s.EndDate,
			Limit:     s.Limit,
			Callback:  addTicketRow,
		})

		if err != nil && err.Error() != "finish" {
			return err
		}
	}

	return nil
}
