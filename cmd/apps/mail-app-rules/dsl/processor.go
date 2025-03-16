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

	// 4. Create sequence set from results
	var seqSet imap.SeqSet
	for _, num := range seqNums {
		seqSet.AddNum(num)
	}

	// 5. Build fetch options
	fetchOptions, err := BuildFetchOptions(rule.Output)
	if err != nil {
		return fmt.Errorf("failed to build fetch options: %w", err)
	}

	// 6. Fetch messages
	fetchCmd := client.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	// 7. Process and output messages
	count := 0
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		msgData, err := msg.Collect()
		if err != nil {
			return fmt.Errorf("failed to collect message data: %w", err)
		}

		output, err := FormatOutput(msgData, rule.Output)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		count++
		if count > 1 {
			fmt.Println("----------------------------------------")
		}
		fmt.Println(output)
	}

	if err := fetchCmd.Close(); err != nil {
		return fmt.Errorf("failed to close fetch command: %w", err)
	}

	// 8. Print summary
	fmt.Printf("\nFound %d message(s) matching the criteria\n", count)

	return nil
}
