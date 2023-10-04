# Zendesk CLI

This is a command line interface for managing Zendesk tickets.
It allows you to easily list and delete Zendesk tickets from the command line.

## Features

- Fetch tickets from Zendesk between specified dates
- Fetch a specific ticket by ID
- Delete tickets by ID
- Bulk delete tickets
- Multithreaded ticket deletion for improved performance
- Incremental ticket exports based on time for efficient fetching

## Usage

```
# Fetch tickets between two dates
zendesk get-tickets --start-date 2021-01-01 --end-date 2021-06-30

# Fetch a specific ticket
zendesk get-tickets --id 36001234567

# Delete a ticket
zendesk delete-tickets --ids 36001234567

# Bulk delete tickets from a file
zendesk delete-tickets --tickets-file tickets.json

# Multithreaded bulk ticket deletion (in chunks of 100)
zendesk delete-tickets --tickets-file tickets.json --workers 10
```

### Authentication

The tool looks for the following environment variables to authenticate with the Zendesk API:

- `ZENDESK_DOMAIN` - Your Zendesk domain
- `ZENDESK_EMAIL` - Your Zendesk email
- `ZENDESK_API_TOKEN` - Your API token

You can also pass these values as flags for one-off usage.

## Contributing

Contributions are welcome! Please open an issue or PR if you would like to contribute.

Some ideas for improvements:

- Add ability to update tickets
- Filter ticket exports
- Output tickets to JSON/CSV
- Improve test coverage

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
