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

### Examples

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

### Detailed flags

```
## Usage:

zendesk [command] 

## Available Commands:

- completion   Generate the autocompletion script for the specified shell
- delete-tickets Delete tickets in Zendesk
- get-tickets Fetch tickets from Zendesk  
- help        Help about any command or topic

## Flags:

-h, --help      help for zendesk

Use "zendesk [command] --help" for more information about a command.
```

```
❯ zendesk delete-tickets --help

delete-tickets - Delete tickets in Zendesk  

For more help, run: zendesk help delete-tickets

## Usage:  

zendesk delete-tickets [flags]

## Flags:

--api-token     Zendesk API token.
--domain        Zendesk domain. 
--email         Zendesk email.
-h, --help      help for delete-tickets
--ids           List of ticket IDs to delete. (default [])
--tickets-file  File containing a list of tickets to delete.
--workers       Number of workers to use. (default 8)
```

``` 
❯ zendesk get-tickets --help

get-tickets - Fetch tickets from Zendesk

For more help, run: zendesk help get-tickets  

## Usage:

zendesk get-tickets [flags]  

## Flags:
  
--api-token     Zendesk API token.
--domain        Zendesk domain.
--email         Zendesk email.  
--end-date      Specify the end time until when you want to fetch tickets.
-h, --help      help for get-tickets
--id            Specify a ticket ID to fetch.
--limit         Limit the number of tickets to fetch.
--start-date    Specify the start time from when you want to start fetching tickets.
```

The get-tickets command outputs structured ticket data using the https://github.com/go-go-golems/glazed package. This
allows piping the output to various destinations and applying additional processing and filtering. See the Glazed docs
for the full list of capabilities.

## Contributing

Contributions are welcome! Please open an issue or PR if you would like to contribute.

Some ideas for improvements:

- Add ability to update tickets
- Filter ticket exports

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
