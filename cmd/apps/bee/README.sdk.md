# Bee.computer API Go Client

This Go package provides a client for interacting with the bee.computer API. It includes methods for managing
conversations, facts, and todos associated with a user.

## Installation

To use this package, first install it using `go get`:

```sh
go get github.com/yourusername/yourrepo/cmd/apps/bee/pkg/bee
```

## Usage

Import the package into your Go project:

```go
import "github.com/yourusername/yourrepo/cmd/apps/bee/pkg/bee"
```

### Creating a Client

Create a new client by providing your API key:

```go
client := bee.NewClient("your-api-key")
```

You can also customize the client with additional options:

```go
httpClient := &http.Client{Timeout: time.Second * 10}
client := bee.NewClient(
    "your-api-key",
    bee.WithHTTPClient(httpClient),
    bee.WithBaseURL("https://custom.api.bee.computer/v1"),
    bee.WithRateLimit(1, 5),
)
```

### Conversations

Retrieve a list of conversations:

```go
conversations, err := client.GetConversations(ctx, userID, page, limit)
```

Get a specific conversation:

```go
conversation, err := client.GetConversation(ctx, userID, conversationID)
```

Delete a conversation:

```go
err := client.DeleteConversation(ctx, userID, conversationID)
```

End a conversation:

```go
err := client.EndConversation(ctx, userID, conversationID)
```

Retry a conversation:

```go
err := client.RetryConversation(ctx, userID, conversationID)
```

### Facts

Retrieve a list of facts:

```go
facts, err := client.GetFacts(ctx, userID, page, limit, confirmed)
```

Create a new fact:

```go
fact, err := client.CreateFact(ctx, userID, bee.FactInput{Text: "New Fact", Confirmed: true})
```

Get a specific fact:

```go
fact, err := client.GetFact(ctx, userID, factID)
```

Update a fact:

```go
fact, err := client.UpdateFact(ctx, userID, factID, bee.FactInput{Text: "Updated Fact", Confirmed: false})
```

Delete a fact:

```go
err := client.DeleteFact(ctx, userID, factID)
```

### Todos

Retrieve a list of todos:

```go
todos, err := client.GetTodos(ctx, userID, page, limit)
```

Create a new todo:

```go
todo, err := client.CreateTodo(ctx, userID, bee.TodoInput{Text: "New Todo", Completed: false})
```

Get a specific todo:

```go
todo, err := client.GetTodo(ctx, userID, todoID)
```

Update a todo:

```go
todo, err := client.UpdateTodo(ctx, userID, todoID, bee.TodoInput{Text: "Updated Todo", Completed: true})
```

Delete a todo:

```go
err := client.DeleteTodo(ctx, userID, todoID)
```

## Types

The package defines several types used in API responses and requests. Refer to the [types.go](https://github.com/yourusername/yourrepo/cmd/apps/bee/pkg/bee/types.go) file for the complete list of types.

## Error Handling

The client methods return errors that can be checked to handle different failure scenarios. For example:

```go
if err != nil {
    if apiErr, ok := err.(bee.APIError); ok {
        // Handle API error with specific status code and message
        fmt.Printf("API error occurred: %s\n", apiErr)
    } else {
        // Handle other types of errors
        fmt.Printf("An error occurred: %s\n", err)
    }
}
```

## Context Support

All client methods accept a `context.Context` parameter, allowing you to set deadlines, cancel requests, and pass other request-scoped values.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

conversations, err := client.GetConversations(ctx, userID, page, limit)
```

## Rate Limiting

The client supports rate limiting. You can configure the rate limit when creating the client:

```go
client := bee.NewClient(
    "your-api-key",
    bee.WithRateLimit(1, 5), // 1 request per second with a burst of 5
)
```

