# Bee.Computer SDK CLI

This CLI tool provides a command-line interface for interacting with the Bee.Computer API. It allows you to manage conversations, facts, and todos directly from your terminal. The CLI is built using the [glazed framework](https://github.com/go-go-golems/glazed), which provides powerful capabilities for formatting and modifying structured output.

## Setup

### Prerequisites

- Go 1.21 or later
- An API key from Bee.Computer

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/go-go-golems/go-go-labs.git
   ```

2. Build the CLI:
   ```
   go build -o bee ./cmd/apps/bee
   ```

3. Set up your API key:
   Export your Bee.Computer API key as an environment variable:
   ```
   export BEE_API_KEY=your_api_key_here
   ```

## Usage

The CLI is structured with three main commands: `conversation`, `fact`, and `todo`. Each of these has subcommands for
different operations. All commands support various flags provided by the glazed framework for formatting and modifying
the output.

### Conversations

- List conversations:
  ```
  ./bee conversation list [flags]
  ```

- Get a specific conversation:
  ```
  ./bee conversation get <conversation_id> [flags]
  ```

- Delete a conversation:
  ```
  ./bee conversation delete <conversation_id> [flags]
  ```

- End a conversation:
  ```
  ./bee conversation end <conversation_id> [flags]
  ```

- Retry a conversation:
  ```
  ./bee conversation retry <conversation_id> [flags]
  ```

### Facts

- List facts:
  ```
  ./bee fact list [flags]
  ```

- Create a new fact:
  ```
  ./bee fact create --text "Your fact text here" [flags]
  ```

- Get a specific fact:
  ```
  ./bee fact get <fact_id> [flags]
  ```

- Update a fact:
  ```
  ./bee fact update <fact_id> --text "Updated fact text" [flags]
  ```

- Delete a fact:
  ```
  ./bee fact delete <fact_id> [flags]
  ```

### Todos

- List todos:
  ```
  ./bee todo list [flags]
  ```

- Create a new todo:
  ```
  ./bee todo create --text "Your todo text here" [flags]
  ```

- Get a specific todo:
  ```
  ./bee todo get <todo_id> [flags]
  ```

- Update a todo:
  ```
  ./bee todo update <todo_id> --text "Updated todo text" [flags]
  ```

- Delete a todo:
  ```
  ./bee todo delete <todo_id> [flags]
  ```

## Glazed Framework Features

This CLI uses the glazed framework, which provides a rich set of features for working with structured data. Some key
features include:

- Multiple output formats (JSON, YAML, CSV, etc.)
- Data filtering and transformation
- Customizable table layouts

To see the full list of available flags and options, use the `--help` flag with any command:

```
./bee <command> <subcommand> --help
```

## Pagination

For commands that list multiple items (like `conversation list`, `fact list`, and `todo list`), you can use the `--page` and `--limit` flags to control pagination:

- `--page`: Specifies the page number (default is 1)
- `--limit`: Specifies the number of items per page (default is 10)

Example:
```
./bee fact list --page 2 --limit 20
```

This will retrieve the second page of facts, with 20 facts per page.
