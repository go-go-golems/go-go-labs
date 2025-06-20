Thanks! I’ll look into how to programmatically add comments to issues, update their body, assign users, and add labels via GitHub’s APIs—especially in the context of a dashboard interface. I’ll clarify whether each of these actions is available through the GraphQL API, REST API, or both, and how to use them from Go.


# Managing GitHub Issues via API (GraphQL vs REST)

## Adding Comments to Issues

**GraphQL API:** The GraphQL v4 API supports adding comments using the `addComment` mutation. For example:

```graphql
mutation {
  addComment(input: { subjectId: "ISSUE_NODE_ID", body: "Comment text" }) {
    commentEdge {
      node { id, url, body } 
    }
  }
}
```

This mutation requires the issue’s global **node ID** as `subjectId` and the comment text. In Go, you can execute this mutation by sending an HTTP POST to the GraphQL endpoint (`https://api.github.com/graphql`) with the query, including an authorization header with your token, or use a GraphQL client (such as the `githubv4` Go client library) to construct the mutation call.

**REST API:** The GitHub REST API (v3) provides an endpoint to create issue comments: `POST /repos/{owner}/{repo}/issues/{issue_number}/comments`. The request body must include the comment text (e.g. `{"body": "Comment text"}`). For example, using cURL:

```bash
curl -X POST -H "Authorization: Bearer <TOKEN>" \
     -H "Accept: application/vnd.github+json" \
     https://api.github.com/repos/OWNER/REPO/issues/123/comments \
     -d '{"body": "Comment text"}'
```

In Go, the idiomatic approach is to use the official `google/go-github` client library. For instance, you can call `client.Issues.CreateComment(ctx, owner, repo, issueNumber, comment)`, where `comment` is a `github.IssueComment` with the Body field set. Ensure you authenticate using a personal access token (PAT) or GitHub App token – for example, by wrapping the HTTP client with an OAuth2 token source as shown in the snippet (loading the token from an environment variable). Both GraphQL and REST methods fully support issue comments; using REST with Go’s library is straightforward, while GraphQL allows you to combine comment creation with fetching other data in a single request if needed.

## Updating Issue Description (Body)

**GraphQL API:** You can update an issue’s body (description) using the `updateIssue` GraphQL mutation. This mutation accepts the issue’s Node ID and the fields to change. For example:

```graphql
mutation {
  updateIssue(input: { id: "ISSUE_NODE_ID", body: "Updated issue description" }) {
    issue {
      id
      body
    }
  }
}
```

The GraphQL input object supports updating the title, body, state, assignees, labels, etc., in one call. (For instance, you could close an issue and add labels at the same time.) In Go, if you use the GraphQL route, you can construct a mutation query (as a string or via a GraphQL client library) and send it via an HTTP request. Include the OAuth token in the `Authorization: Bearer <TOKEN>` header for authentication.

**REST API:** The REST API supports editing issues with a PATCH request: `PATCH /repos/{owner}/{repo}/issues/{issue_number}`. You supply a JSON body containing the fields to update – for example, `{"body": "Updated issue description"}` to change the description. The response will reflect the new issue content. In Go (using go-github), you can use `client.Issues.Edit` with a `github.IssueRequest` struct: for example, set the `Body` field and call `Issues.Edit(ctx, owner, repo, issueNumber, issueRequest)`. This will issue the PATCH request for you. Both GraphQL and REST can update the issue body. If you need to update multiple attributes at once (e.g. body, title, labels, assignees), both APIs support that as well – GraphQL’s `updateIssue` allows setting those fields in one mutation, and the REST API allows specifying multiple fields in one PATCH payload (for example, including `title`, `body`, `labels`, and `assignees` together). The choice often comes down to convenience: using the REST client in Go is very simple for this, whereas GraphQL might be useful if you want to fetch or modify related data in the same request.

## Adding Labels to Issues

**GraphQL API:** GitHub’s GraphQL API provides two ways to manage labels on an issue. You can either include a list of `labelIds` in the `updateIssue` mutation to set the exact labels for the issue, or use the dedicated `addLabelsToLabelable` mutation to append labels without removing existing ones. For example, to add labels via GraphQL:

```graphql
mutation {
  addLabelsToLabelable(input: { 
    labelableId: "ISSUE_OR_PR_NODE_ID", 
    labelIds: ["MDU6TGFiZWwzMzY5MzUxMDI2", "MDU6TGFiZWwzMzY5MzUxMDI1"] 
  }) {
    labelable {
      ... on Issue {
        id
        labels(first: 5) { nodes { name } }
      }
    }
  }
}
```

This requires the issue’s ID and the IDs of the labels you want to add. (You can obtain label node IDs by querying the repository’s labels, filtering by name.) In Go, using GraphQL would involve preparing this mutation and sending it via a GraphQL client; ensure your token has appropriate scopes (e.g. `repo` scope for private repos) and is included in the authorization header.

**REST API:** The REST v3 equivalent is `POST /repos/{owner}/{repo}/issues/{issue_number}/labels`, which adds one or more labels to an issue. The JSON body takes an array of label names, for example: `{"labels": ["bug", "enhancement"]}`. This call will append those labels to the issue’s existing labels. (If you provide an empty array, it removes all labels. If a provided label name does not exist in the repository, the API will return a 404 error – you would need to create the label first using the labels API.) In Go, you can simply call `client.Issues.AddLabelsToIssue(ctx, owner, repo, issueNumber, []string{"bug","enhancement"})` using the go-github library, which wraps this REST endpoint. Alternatively, you could include a `labels` array in a PATCH to `/issues/{number}` to replace the set of labels on the issue. Typically, the “add labels” POST endpoint is the safest when you just want to append new labels without affecting existing ones. Both GraphQL and REST cover this operation. GraphQL requires you to use label IDs (and has no mutation to create new labels), whereas the REST approach uses label names directly. If you already have a Go REST client, using `AddLabelsToIssue` is straightforward; if you’re using GraphQL, you might use `addLabelsToLabelable` for a non-destructive addition.

## Assigning Users to Issues

**GraphQL API:** Assigning users is supported via the `addAssigneesToAssignable` mutation, which lets you add one or more assignees to an issue (or pull request) by their user IDs. For example:

```graphql
mutation {
  addAssigneesToAssignable(input: { 
    assignableId: "ISSUE_NODE_ID", 
    assigneeIds: ["MDQ6VXNlcjU4MzI1ODQ=", "MDQ6VXNlcjEwMjM0NTY="] 
  }) {
    assignable { 
      ... on Issue {
        id
        assignees(first: 10) { nodes { login } }
      }
    }
  }
}
```

This will add those users to the issue’s assignees (any already-assigned users remain assigned). Similarly, GraphQL’s `updateIssue` mutation can set the full assignee list via an `assigneeIds` array, which replaces the assignees with the specified set. In a Go backend, you could execute an `addAssigneesToAssignable` call by sending a GraphQL mutation (using a library or raw HTTP) with the proper token.

**REST API:** The REST approach provides an endpoint to add assignees: `POST /repos/{owner}/{repo}/issues/{issue_number}/assignees`. The body should include an `"assignees"` array of usernames, e.g. `{"assignees":["octocat","hubot"]}`. This call adds the specified users to the issue without removing existing assignees (note that only users with push access to the repo can be assigned; any usernames that don’t have access or are already assigned will be silently ignored). In Go, the go-github client offers `Issues.AddAssigneesToIssue()` (or you can use `Issues.Edit` with the Assignees field to replace all assignees). For example, to append an assignee without removing others, you can call `client.Issues.AddAssigneesToIssue(ctx, owner, repo, issueNumber, []string{"hubot"})`. Under the hood, this hits the above POST endpoint. Both GraphQL and REST allow assigning users. If you want to set a precise list of assignees (overwriting existing assignments), you can use GraphQL’s `updateIssue` with `assigneeIds` or the REST PATCH method with an `"assignees"` list (which replaces all assignees on the issue). If you just want to **add** a user or two, using the dedicated add-assignees mutation/endpoint is convenient. In practice, using the REST API via the Go client is very straightforward for this task, whereas GraphQL might be chosen if you’re already using it or need to bundle multiple operations in one call.

## Authentication and Security

Whether you choose REST or GraphQL, make sure to authenticate requests securely. In all cases, you must supply a valid token with appropriate scopes (for example, a personal access token with the `repo` scope for private repositories or `public_repo` for public repos). In Go, avoid hard-coding the token; instead, read it from an environment variable or configuration and use it to create an OAuth2 client. The `go-github` library does not handle token auth automatically, but you can easily integrate it with `golang.org/x/oauth2` as shown in the example code (using `oauth2.StaticTokenSource` and `oauth2.NewClient`). This will inject the `Authorization: token ...` header on each request. For GraphQL, you similarly include the token in the `Authorization: Bearer <TOKEN>` header of your HTTP requests. By following these patterns, your Go backend (dashboard) will interact with GitHub’s API in an idiomatic and secure way.

**Choosing GraphQL vs REST:** All the operations above are supported in both GraphQL (v4) and REST (v3) APIs. There isn’t a strict requirement to use one over the other – it often depends on your needs and existing workflow. The REST API, combined with GitHub’s Go SDK, is very convenient for common tasks (comments, updating issues, labels, assignees) with minimal setup. The GraphQL API offers more flexibility if you need to batch queries or retrieve exactly the data you want in one round-trip. For example, you might use GraphQL to add a comment and simultaneously query the issue’s updated status or other fields in a single request. However, for a typical dashboard backend, the REST endpoints are usually sufficient and quicker to implement. In terms of capabilities, both APIs are largely on par for issue management today – GraphQL even allows bulk updates like adding labels and assignees in one mutation. In summary, use the REST API (and Go client) for simplicity and clarity, or GraphQL if you require its querying power; both will let you add comments, edit issue descriptions, label issues, and assign users effectively in your Go application.

**Sources:**

1. GitHub GraphQL API Reference – Mutations (Issue operations)
2. GitHub GraphQL API Reference – UpdateIssue mutation input fields
3. GitHub REST API Reference – Issues (create comment, update issue, labels, assignees)
4. GitHub REST API Reference – Issues (parameters for labels and assignees)
5. Stack Overflow – Using go-github for issue comments (Go code sample)

