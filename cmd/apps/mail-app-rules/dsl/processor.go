package dsl

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// ProcessRule executes an IMAP rule
func ProcessRule(client *imapclient.Client, rule *Rule) error {
	// 1. Build search criteria
	criteria, err := BuildSearchCriteria(rule.Search)
	if err != nil {
		return fmt.Errorf("failed to build search criteria: %w", err)
	}

	// 2. Execute search
	searchCmd := client.Search(criteria, nil)
	searchData, err := searchCmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to execute search: %w", err)
	}

	// 3. Check if we have results
	seqNums := searchData.AllSeqNums()
	if len(seqNums) == 0 {
		fmt.Println("No messages found matching the criteria")
		return nil
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
		return fmt.Errorf("failed to build fetch options: %w", err)
	}

	bodySection := &imap.FetchItemBodySection{
		Part: []int{1},
		Peek: true,
	}
	fetchOptions.BodySection = []*imap.FetchItemBodySection{bodySection}

	// 6. First fetch: get metadata and structure
	messages, err := client.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	// 7. Process each message
	count := 0
	for _, msg := range messages {
		// Determine required body sections based on structure
		bodySections, err := determineRequiredBodySections(msg.BodyStructure, rule.Output)
		if err != nil {
			return fmt.Errorf("failed to determine required body sections: %w", err)
		}

		var bodySectionData imapclient.FetchItemDataBodySection

		// If we need body sections, do a second fetch
		if len(bodySections) > 0 {
			// Create a sequence set for just this message
			msgSeqSet := imap.SeqSetNum(msg.SeqNum)

			// Second fetch: get required body sections
			bodyFetchOptions, err := BuildFetchOptions(rule.Output)
			if err != nil {
				return fmt.Errorf("failed to build fetch options: %w", err)
			}
			bodyFetchOptions.BodySection = bodySections

			fetchCmd := client.Fetch(msgSeqSet, bodyFetchOptions)
			defer fetchCmd.Close()

			msg := fetchCmd.Next()
			if msg == nil {
				return fmt.Errorf("failed to fetch message body: %w", err)
			}

			var found bool
			for {
				item := msg.Next()
				if item == nil {
					break
				}

				if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
					bodySectionData = data
					found = true
					break
				}
			}
			// msg_, err := msg.Collect()
			// _ = msg_
			// if err != nil {
			// 	return fmt.Errorf("failed to collect message body: %w", err)
			// }

			// var found bool
			// for _, item := range msg_.BodySection {
			// 	if data, ok := item.Section.
			// 		bodySectionData = data
			// 		found = true
			// 		// break
			// 	}
			// }

			err = fetchCmd.Close()
			if err != nil {
				return fmt.Errorf("failed to close fetch command: %w", err)
			}

			if !found {
				return fmt.Errorf("failed to find body section data")
			}
		}

		// Format and output the message
		output, err := FormatOutput(msg, rule.Output, bodySectionData)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		count++
		if count > 1 {
			fmt.Println("----------------------------------------")
		}
		fmt.Println(output)
	}

	// 8. Print summary
	fmt.Printf("\nFound %d message(s) matching the criteria\n", count)

	return nil
}
