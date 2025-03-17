package dsl

import (
	"io"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
)

// EmailMessage represents a fully fetched email message with all its data
type EmailMessage struct {
	UID        uint32
	SeqNum     uint32
	Envelope   *EmailEnvelope
	Flags      []string
	Size       uint32
	MimeParts  []MimePart
	RawContent map[string][]byte // Store different body sections by their part specifier
}

// EmailEnvelope contains the message envelope information
type EmailEnvelope struct {
	Subject string
	From    []EmailAddress
	To      []EmailAddress
	Date    time.Time
}

// EmailAddress represents an email address with optional name
type EmailAddress struct {
	Name    string
	Address string
}

// NewEmailMessageFromIMAP creates an EmailMessage from IMAP message data
func NewEmailMessageFromIMAP(msg *imapclient.FetchMessageBuffer, mimeParts []MimePart) (*EmailMessage, error) {
	// Convert flags to strings
	flags := make([]string, len(msg.Flags))
	for i, flag := range msg.Flags {
		flags[i] = string(flag)
	}

	email := &EmailMessage{
		UID:        uint32(msg.UID),
		SeqNum:     msg.SeqNum,
		Flags:      flags,
		Size:       uint32(msg.RFC822Size),
		MimeParts:  mimeParts,
		RawContent: make(map[string][]byte),
	}

	if msg.Envelope != nil {
		email.Envelope = &EmailEnvelope{
			Subject: msg.Envelope.Subject,
			Date:    msg.Envelope.Date,
		}

		// Convert From addresses
		if len(msg.Envelope.From) > 0 {
			email.Envelope.From = make([]EmailAddress, len(msg.Envelope.From))
			for i, addr := range msg.Envelope.From {
				email.Envelope.From[i] = EmailAddress{
					Name:    addr.Name,
					Address: addr.Mailbox + "@" + addr.Host,
				}
			}
		}

		// Convert To addresses
		if len(msg.Envelope.To) > 0 {
			email.Envelope.To = make([]EmailAddress, len(msg.Envelope.To))
			for i, addr := range msg.Envelope.To {
				email.Envelope.To[i] = EmailAddress{
					Name:    addr.Name,
					Address: addr.Mailbox + "@" + addr.Host,
				}
			}
		}
	}

	return email, nil
}

// fetchMimeParts extracts MIME part content types and content from the message
func fetchMimeParts(r io.Reader) ([]MimePart, error) {
	result := []MimePart{}

	mr, err := mail.CreateReader(r)
	if err != nil {
		return nil, err
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			contentType, params, _ := header.ContentType()
			content, err := io.ReadAll(part.Body)
			if err != nil {
				return nil, err
			}
			result = append(result, MimePart{
				Type:    contentType,
				Subtype: params["subtype"],
				Content: string(content),
				Size:    uint32(len(content)),
				Charset: params["charset"],
			})
		case *mail.AttachmentHeader:
			filename, _ := header.Filename()
			contentType, _, _ := header.ContentType()
			content, err := io.ReadAll(part.Body)
			if err != nil {
				return nil, err
			}
			result = append(result, MimePart{
				Filename: filename,
				Type:     contentType,
				Size:     uint32(len(content)),
			})
		}
	}

	return result, nil
}
