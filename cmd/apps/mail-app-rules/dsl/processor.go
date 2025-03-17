package dsl

import (
	"fmt"
	"io"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/rs/zerolog/log"
)

// FetchMessages retrieves messages from IMAP server based on the rule
func (rule *Rule) FetchMessages(client *imapclient.Client) ([]*EmailMessage, error) {
	// 1. Build search criteria
	criteria, err := BuildSearchCriteria(rule.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to build search criteria: %w", err)
	}

	// 2. Execute search
	searchCmd := client.Search(criteria, nil)
	searchData, err := searchCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	// 3. Check if we have results
	seqNums := searchData.AllSeqNums()
	if len(seqNums) == 0 {
		return nil, nil
	}

	// 4. Create sequence set from results, respecting the limit if set
	var seqSet imap.SeqSet
	limit := len(seqNums)
	if rule.Output.Limit > 0 && rule.Output.Limit < limit {
		limit = rule.Output.Limit
	}

	// Use the most recent messages first (highest sequence numbers)
	startIdx := len(seqNums) - 1
	endIdx := len(seqNums) - limit
	if endIdx < 0 {
		endIdx = 0
	}

	for i := startIdx; i >= endIdx; i-- {
		seqSet.AddNum(seqNums[i])
	}

	// 5. Build initial fetch options for metadata and structure
	fetchOptions, err := BuildFetchOptions(rule.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to build fetch options: %w", err)
	}

	fetchOptions.BodySection = []*imap.FetchItemBodySection{}

	// 6. First fetch: get metadata and structure
	messages, err := client.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	// 7. Process each message
	result := make([]*EmailMessage, 0, len(messages))
	for _, msg := range messages {
		// Determine required body sections based on structure
		bodyStructure := msg.BodyStructure
		mimePartMetadata, err := determineRequiredBodySections(bodyStructure, rule.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to determine required body sections: %w", err)
		}

		var mimeParts []MimePart
		var fetchSections []*imap.FetchItemBodySection

		// Collect all fetch sections
		for _, metadata := range mimePartMetadata {
			fetchSections = append(fetchSections, metadata.FetchSection)
		}

		// If we need body sections, do a second fetch
		if len(fetchSections) > 0 {
			// Create a sequence set for just this message
			msgSeqSet := imap.SeqSetNum(msg.SeqNum)

			// Second fetch: get required body sections
			bodyFetchOptions, err := BuildFetchOptions(rule.Output)
			if err != nil {
				return nil, fmt.Errorf("failed to build fetch options: %w", err)
			}
			bodyFetchOptions.BodyStructure = &imap.FetchItemBodyStructure{}
			bodyFetchOptions.BodySection = fetchSections

			fetchCmd := client.Fetch(msgSeqSet, bodyFetchOptions)
			defer fetchCmd.Close()

			msg := fetchCmd.Next()
			if msg == nil {
				return nil, fmt.Errorf("failed to fetch message body")
			}

			// Create a map to store content for each path
			contentMap := make(map[string][]byte)

			for {
				item := msg.Next()
				if item == nil {
					break
				}

				if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
					// Read the body content
					content, err := io.ReadAll(data.Literal)
					if err != nil {
						return nil, fmt.Errorf("failed to read body section: %w", err)
					}

					// Create a key from the section
					pathKey := fmt.Sprintf("%v", data.Section.Part)
					contentMap[pathKey] = content
				}
			}

			err = fetchCmd.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to close fetch command: %w", err)
			}

			// Create MimeParts using metadata and content
			for _, metadata := range mimePartMetadata {
				pathKey := fmt.Sprintf("%v", metadata.Path)
				content := contentMap[pathKey]

				mimePart := MimePart{
					Type:     metadata.Type,
					Subtype:  metadata.Subtype,
					Content:  string(content),
					Size:     uint32(len(content)),
					Charset:  metadata.Params["charset"],
					Filename: metadata.Filename,
				}
				mimeParts = append(mimeParts, mimePart)
			}
		}

		// Convert to our internal format
		email, err := NewEmailMessageFromIMAP(msg, mimeParts)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}
		result = append(result, email)
	}

	return result, nil
}

// ProcessRule executes an IMAP rule
func ProcessRule(client *imapclient.Client, rule *Rule) error {
	// 1. Fetch messages
	messages, err := rule.FetchMessages(client)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		log.Warn().Msg("No messages found matching the criteria")
		return nil
	}

	// 2. Output messages
	err = OutputMessages(messages, rule.Output)
	if err != nil {
		return fmt.Errorf("failed to output messages: %w", err)
	}

	return nil
}
