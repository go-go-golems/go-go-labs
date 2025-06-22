# Building a GitHub Projects (Beta) GraphQL API Client in Go

GitHub’s **Projects (Beta)** (also known as **Projects v2**) can be managed programmatically through GitHub’s **GraphQL v4 API**. In this guide, we’ll build a comprehensive Go client to interact with the GitHub GraphQL API for Projects. We’ll cover secure authentication with tokens, querying project data (including custom fields and views), creating and updating project items, linking issues to project items, handling various field types (text, number, date, single-select, etc.), and programmatically managing custom fields and options. All examples will use idiomatic Go, focusing purely on API-level interactions (no front-end UI code). We’ll also provide tips on navigating the GraphQL schema in Go for effective development.

**Table of Contents:**

1. [Setting Up Go GraphQL Client and Authentication](#setting-up-go-graphql-client-and-authentication)
2. [Querying Projects and Fields via GraphQL](#querying-projects-and-fields-via-graphql)

   * 2.1 [Retrieving a Project by Owner and Number](#retrieving-a-project-by-owner-and-number)
   * 2.2 [Listing Project Fields and Field Types](#listing-project-fields-and-field-types)
   * 2.3 [Retrieving Project Views (Tables/Boards)](#retrieving-project-views)
3. [Querying Project Items and Field Values](#querying-project-items-and-field-values)
4. [Creating and Updating Project Items](#creating-and-updating-project-items)

   * 4.1 [Adding Existing Issues/PRs to a Project](#adding-existing-issuesprs-to-a-project)
   * 4.2 [Adding Draft Issues to a Project](#adding-draft-issues-to-a-project)
   * 4.3 [Updating Field Values for Project Items](#updating-field-values-for-project-items)
5. [Creating GitHub Issues and Linking to Projects](#creating-github-issues-and-linking-to-projects)
6. [Working with Custom Fields and Options](#working-with-custom-fields-and-options)

   * 6.1 [Updating Text, Number, and Date Fields](#updating-text-number-and-date-fields)
   * 6.2 [Updating Single-Select Fields](#updating-single-select-fields)
   * 6.3 [Updating Iteration Fields](#updating-iteration-fields)
   * 6.4 [Listing and Managing Single-Select Options](#listing-and-managing-single-select-options)
   * 6.5 [Creating New Custom Fields via API](#creating-new-custom-fields-via-api)
7. [Navigating the GraphQL Schema in Go](#navigating-the-graphql-schema-in-go)
8. [Conclusion](#conclusion)

## Setting Up Go GraphQL Client and Authentication

To interact with GitHub’s GraphQL API, you’ll need a GitHub access token with appropriate scopes and a GraphQL HTTP client in Go. GitHub’s GraphQL endpoint is `https://api.github.com/graphql`. We use a Personal Access Token (PAT) or GitHub App installation token for authentication, and **it must include the `read:project` scope (for read-only access) or `project` scope (for read-write)**. The token will be sent in the `Authorization` header as a Bearer token.

**Secure Token Handling:** Never hard-code the token in your code. Instead, store it securely (e.g., in an environment variable or a config file) and load it at runtime. In Go, you might retrieve it via `os.Getenv("GITHUB_TOKEN")` or use a secrets management solution.

For the GraphQL client, we can use a library for convenience. A popular choice is the `machinebox/graphql` package, a simple low-level GraphQL HTTP client. Another option is `hasura/go-graphql-client` (a fork of `shurcooL/graphql`) which provides a type-safe client. We’ll use `machinebox/graphql` for its straightforward request-building API.

**Initializing the GraphQL Client in Go:**

```go
import (
    "context"
    "log"
    "os"

    "github.com/machinebox/graphql"
)

func main() {
    // Load GitHub token from environment
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        log.Fatal("GitHub token not set")
    }

    // Create a new GraphQL client (can be reused for multiple requests)
    client := graphql.NewClient("https://api.github.com/graphql")

    // Prepare a context (for cancellation/timeouts)
    ctx := context.Background()

    // Use the client...
}
```

We haven’t sent any request yet. The `machinebox/graphql` client allows us to build a request, set variables and headers, and then execute it. We must include the Authorization header with our token on each request. We can set a header for each request, or use a custom http.Client that injects the header for all requests.

**Example: Setting the Authorization Header**

```go
req := graphql.NewRequest(`
    query { viewer { login } } 
`) 
req.Header.Set("Authorization", "Bearer "+token)

// Now execute the request (and handle errors)
var resp struct {
    Viewer struct {
        Login string
    }
}
if err := client.Run(ctx, req, &resp); err != nil {
    log.Fatalf("GraphQL query failed: %v", err)
}
log.Printf("Authenticated as GitHub user: %s\n", resp.Viewer.Login)
```

In this snippet, we run a simple GraphQL query to verify authentication (fetching the logged-in user’s login name). The key is adding `Authorization: Bearer <TOKEN>` in the request header. No additional Accept headers are needed for GraphQL v4 (unlike REST v3 API, where custom media types are sometimes required).

**Using OAuth2 library (optional):** Alternatively, you can integrate with `golang.org/x/oauth2` to automatically inject the token via an HTTP client. For example:

```go
src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
httpClient := oauth2.NewClient(ctx, src)
client := graphql.NewClient("https://api.github.com/graphql", httpClient)
```

This way, the OAuth2 client adds the Authorization header for you. Either approach is fine.

## Querying Projects and Fields via GraphQL

Once the client is set up, we can query GitHub Projects data. The **Projects Beta** are represented in the GraphQL schema by the `ProjectV2` type (and related types like `ProjectV2Item`, `ProjectV2Field`, etc.). The older `ProjectNext` type has been deprecated in favor of `ProjectV2`. We will use queries to retrieve project details and their fields.

### Retrieving a Project by Owner and Number

Each project is identified by an **owner** (an organization or user) and a **project number** (the number in the URL). For example, a project URL might be `https://github.com/orgs/<org>/projects/5` (here the owner is an org and project number is 5). We can query a project by number using the owner’s login and the project number.

**GraphQL Query – Get Project by Number:**

```graphql
query($org: String!, $number: Int!) {
  organization(login: $org) {
    projectV2(number: $number) {
      id
      title
      public
      shortDescription
      closed
      items(first: 0) { totalCount }
    }
  }
}
```

This query (which can also be adapted for a user-owned project by using `user(login: "username")` instead of `organization`) fetches the project’s **node ID**, title, visibility (`public` boolean), short description, open/closed status, and an item count. We use `items(first: 0) { totalCount }` to get the count without retrieving items. The **node ID** (`id`) is important for mutations, as many GraphQL mutations use global node IDs as input.

In Go, using the `machinebox/graphql` client, we can execute this query as follows:

```go
// Define the query string with variables
query := `
query($org: String!, $number: Int!) {
  organization(login: $org) {
    projectV2(number: $number) {
      id
      title
      public
      shortDescription
      closed
      items(first: 0) { totalCount }
    }
  }
}`

req := graphql.NewRequest(query)
req.Var("org", "my-org")       // substitute your org or user
req.Var("number", 5)          // substitute the project number
req.Header.Set("Authorization", "Bearer "+token)

var resp struct {
    Organization struct {
        ProjectV2 struct {
            Id               string
            Title            string
            Public           bool
            ShortDescription string
            Closed           bool
            Items struct {
                TotalCount int
            }
        }
    }
}
if err := client.Run(ctx, req, &resp); err != nil {
    log.Fatal(err)
}
project := resp.Organization.ProjectV2
fmt.Printf("Project '%s' (ID=%s) – %d items\n", project.Title, project.Id, project.Items.TotalCount)
```

This will print basic info about the project. The **project’s node ID** (e.g. `"PVT_kwDOBQfyVc0FoQ..."`) is crucial for subsequent operations like adding items or updating fields. If you already know the project’s node ID, you can query the project directly by node using the `node(id: "...")` field and an inline fragment on `ProjectV2`, but querying by owner and number as above is usually simpler for an initial lookup.

### Listing Project Fields and Field Types

Projects have multiple fields – both default fields (like “Title”, “Assignees”, “Status”, etc.) and custom fields you may have added. Each field has a unique **ID**, a **name**, and a data type. We should list the fields to understand what custom fields exist and to get the IDs needed to update them.

The GraphQL API provides a `fields` connection on `ProjectV2`. Each field can be one of several types:

* **Text, Number, or Date fields** – represented by the `ProjectV2Field` object.
* **Single-select fields** – represented by `ProjectV2SingleSelectField`, which includes a list of options.
* **Iteration fields** (date-based sprints) – represented by `ProjectV2IterationField`, with an iteration configuration.
* **(Others)** Assignees, Labels, Milestone, and Repository fields exist as built-in fields tied to issue/PR properties. These appear in `fields` but are not updatable via `updateProjectV2ItemFieldValue` (as they mirror issue properties).

**GraphQL Query – List First N Fields of a Project:**

```graphql
query($projectId: ID!) {
  node(id: $projectId) {
    ... on ProjectV2 {
      fields(first: 20) {
        nodes {
          __typename
          id
          name
          ... on ProjectV2SingleSelectField {
            options {
              id
              name
            }
          }
          ... on ProjectV2IterationField {
            configuration {
              iterations {
                id
                startDate
                title
              }
            }
          }
        }
      }
    }
  }
}
```

This query uses an inline fragment on each field type to retrieve additional details:

* For `ProjectV2SingleSelectField`, we fetch its `options` (each option has an ID and name, and possibly a color or description not shown above).
* For `ProjectV2IterationField`, we fetch the configured iterations (each iteration has an ID, start date, and title).

If we only need the field names and IDs, we could use the `ProjectV2FieldCommon` interface to simplify the query, but here we want options and iteration info as well.

**Example Response (simplified):**

```json
{
  "data": {
    "node": {
      "fields": {
        "nodes": [
          { "__typename": "ProjectV2Field", "id": "PVTF_abc123...", "name": "Title" },
          { "__typename": "ProjectV2Field", "id": "PVTF_def456...", "name": "Assignees" },
          { "__typename": "ProjectV2SingleSelectField", "id": "PVTSSF_gh789...", "name": "Status",
            "options": [
              { "id": "f75ad846", "name": "Todo" },
              { "id": "47fc9ee4", "name": "In Progress" },
              { "id": "98236657", "name": "Done" }
            ]
          },
          { "__typename": "ProjectV2IterationField", "id": "PVTIF_xyz111...", "name": "Iteration",
            "configuration": {
              "iterations": [
                { "id": "cfc16e4d", "startDate": "2022-05-29", "title": "Sprint 1" }
              ]
            }
          }
        ]
      }
    }
  }
}
```

In this example, the project has a **Status** field which is single-select (with options “Todo/In Progress/Done”), and an **Iteration** field with an iteration configuration. Every field has a unique opaque ID (e.g., `"PVTF_..."` for basic fields, `"PVTSSF_..."` for single-select, `"PVTIF_..."` for iteration field). We will need these IDs when updating field values on items.

Each field’s type can be discerned by the `__typename` or by the presence of specific sub-fields:

* **Single select fields** – have an `options` array (each option ID is needed to set that option on an item).
* **Iteration fields** – have a `configuration.iterations` list with iteration IDs (needed to set a specific iteration on an item).
* **Text/Number/Date fields** – appear simply as `ProjectV2Field` with just name and id. (The `ProjectV2Field` covers all three of these data types; you might need to know which one it is. The API provides a `dataType` for the field’s configuration – more on that in section 6.5).

Using Go, after retrieving `fields`, you could map field names to IDs for convenience. For example:

```go
fields := resp.Node.Fields.Nodes
for _, f := range fields {
    fmt.Printf("Field %q (Type: %s) ID=%s\n", f.Name, f.Typename, f.Id)
    if f.Typename == "ProjectV2SingleSelectField" {
        for _, opt := range f.Options {
            fmt.Printf("  - Option %q ID=%s\n", opt.Name, opt.Id)
        }
    }
}
```

This would list each field and any single-select options. Knowing the option IDs is crucial because updating a single-select field on an item requires passing the option’s ID (not its name).

### Retrieving Project Views

**Project Views** represent different saved views (Table layout, Board layout, etc., possibly with filters or grouping) in the project. The GraphQL API allows listing the views of a Project via the `views` connection on `ProjectV2`. Each view has its own ID and configuration (like layout type, sorting, grouping).

For instance, you can query:

```graphql
query($projectId: ID!) {
  node(id: $projectId) {
    ... on ProjectV2 {
      views(first: 10) {
        nodes {
          id
          name
          layout
          number
          filters # complex object, can be queried for filter criteria
          sortBy  # sort configuration
          groupBy # grouping configuration
        }
      }
    }
  }
}
```

A view’s `layout` will indicate if it’s a `TABLE_LAYOUT` or `BOARD_LAYOUT`. The `sortBy` and `groupBy` fields (if present) describe how the view is sorted or grouped (e.g., grouped by a single-select field). You can use the view ID and number purely to identify views, but note that as of now (2025) the API does **not support creating or modifying views via GraphQL** – you can only read them. So, for our client, views are mostly relevant if you want to fetch items in a specific view order or filter (which can also be done manually by query filters, though the API’s filtering capabilities are limited). In summary, you can list and inspect project views if needed, but project items are typically fetched directly (as we’ll do next) rather than through a view.

## Querying Project Items and Field Values

Project **items** are the entries in a project’s table/board. Each item can be one of:

* An **Issue** (linked to a GitHub issue)
* A **Pull Request**
* A **Draft Issue** (an issue draft that lives only in the project until converted)
* (If you lack permission, an item may appear as **Redacted**)

Each Project item has its own node ID and a content type. We might want to fetch items along with their linked content and the values of certain fields.

**GraphQL Query – Fetch Project Items with Fields:**

```graphql
query($projectId: ID!) {
  node(id: $projectId) {
    ... on ProjectV2 {
      items(first: 20) {
        nodes {
          id
          type
          content {
            __typename
            ... on Issue {
              title
              number
              url
              assignees(first: 5) { nodes { login } }
            }
            ... on PullRequest {
              title
              number
              url
              assignees(first: 5) { nodes { login } }
            }
            ... on DraftIssue {
              title
              body
            }
          }
          # Fetch values of first 5 custom fields (by index in project settings)
          fieldValues(first: 5) {
            nodes {
              __typename
              ... on ProjectV2ItemFieldTextValue {
                text
                field { ... on ProjectV2FieldCommon { name } }
              }
              ... on ProjectV2ItemFieldNumberValue {
                number
                field { ... on ProjectV2FieldCommon { name } }
              }
              ... on ProjectV2ItemFieldDateValue {
                date
                field { ... on ProjectV2FieldCommon { name } }
              }
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field { ... on ProjectV2FieldCommon { name } }
              }
              ... on ProjectV2ItemFieldIterationValue {
                title
                startDate
                field { ... on ProjectV2FieldCommon { name } }
              }
            }
          }
        }
      }
    }
  }
}
```

This query takes a project ID and returns up to 20 items. For each item:

* We get the item’s **ID** and **type** (the type will be `"ISSUE"`, `"PULL_REQUEST"`, `"DRAFT_ISSUE"`, or `"REDACTED"`).
* The `content` field is a union of Issue, PullRequest, or DraftIssue. We retrieve basic info like title, URL, number, and assignees for issues/PRs.
* `fieldValues(first: 5)` gives the first five field values for each item (you can adjust the number or use `fieldValueByName` to get specific fields by name). Each field value is also a union type:

  * `ProjectV2ItemFieldTextValue` with a `text` value.
  * `ProjectV2ItemFieldNumberValue` with a `number` value.
  * `ProjectV2ItemFieldDateValue` with a `date` (ISO date string).
  * `ProjectV2ItemFieldSingleSelectValue` with a selected option’s name (and you could also get the `optionId` if needed).
  * `ProjectV2ItemFieldIterationValue` with iteration details (like title and start date of the iteration).
  * There are also types for `LabelValue`, `MilestoneValue`, `RepositoryValue` for those built-in fields, not shown above.

In practice, you may want to fetch *all* field values for each item. The API might limit how many you can get in one go (the docs example fetched first 8 fields). You can paginate `fieldValues` if needed or query specific fields by name (using `fieldValueByName(name: "...")` which is convenient if you know the field name).

**Handling Pagination:** The query above uses `first: 20` for items. If you need to retrieve more items, use the `pageInfo` from the result to check `hasNextPage` and `endCursor`, then fetch the next page with `after: <cursor>` in the `items()` connection (GraphQL pagination). Similarly for `fieldValues` if needed.

**Processing Items in Go:** You’d define Go structs for the JSON shape or use an interface to handle the union types. If using a library like `hasura/go-graphql-client`, you could define separate struct fields with GraphQL inline fragment tags. With the `machinebox` client, you might retrieve `fieldValues` as `[]map[string]interface{}` or similar to inspect `__typename` and cast accordingly, or better, create strongly-typed structs for each possible field value type.

For example, a simplified Go structure:

```go
type ProjectItem struct {
    Id       string
    Type     string
    Content  struct {
        Typename    string `json:"__typename"`
        Title       string
        URL         string
        Number      int
        Assignees   struct{ Nodes []struct{ Login string } } `json:"assignees"`
        Body        string // for DraftIssue
    }
    FieldValues struct {
        Nodes []struct {
            Typename string `json:"__typename"`
            Text     *string `json:"text,omitempty"`
            Number   *float64 `json:"number,omitempty"`
            Date     *string `json:"date,omitempty"`
            Name     *string `json:"name,omitempty"` // for single select name
            Field    struct {
                Name string
            }
        }
    }
}
```

Then parse into `[]ProjectItem`. Each entry’s `Type` and `Content.Typename` tell you the item type. The fieldValues slice contains various value types; you may check `Typename` for `"ProjectV2ItemFieldSingleSelectValue"` etc., and then use the corresponding field.

## Creating and Updating Project Items

Now that we can read project info, let’s automate changes: adding items to projects and updating their field values. GitHub’s GraphQL API provides mutations for these tasks (all require the `project` scope on the token, since these are write operations).

**Important:** You **cannot** add an item and update its fields in the same GraphQL call – these must be separate operations. For example, if you add an issue to a project, you’ll get back an item ID, and then you must call a second mutation to update any custom fields on that item.

### Adding Existing Issues/PRs to a Project

To add an existing issue or pull request to a Project (v2), use the `addProjectV2ItemById` mutation. This requires the project’s ID and the content’s ID (the issue or PR node ID).

* You can obtain an issue’s node ID via a GraphQL query (e.g. query the issue by number) or via REST. For simplicity, if you have the repository and issue number, a GraphQL query like `repository(owner:"X", name:"Y"){ issue(number: N){ id } }` will give the issue’s node ID.
* Pull request IDs can be obtained similarly.

**GraphQL Mutation – Add Issue/PR to Project:**

```graphql
mutation($projectId: ID!, $contentId: ID!) {
  addProjectV2ItemById(input: { projectId: $projectId, contentId: $contentId }) {
    item {
      id
    }
  }
}
```

If successful, this returns the new **project item ID** (`item.id`). If the item was already in the project, the API will return the existing item’s ID instead.

In Go, using our client:

```go
addReq := graphql.NewRequest(`
mutation($project: ID!, $content: ID!) {
  addProjectV2ItemById(input:{ projectId: $project, contentId: $content }) {
    item { id }
  }
}`)
addReq.Var("project", projectId)
addReq.Var("content", issueId)
addReq.Header.Set("Authorization", "Bearer "+token)

var addResp struct {
    AddProjectV2ItemById struct {
        Item struct{ Id string }
    }
}
if err := client.Run(ctx, addReq, &addResp); err != nil {
    log.Fatalf("Failed to add item: %v", err)
}
newItemId := addResp.AddProjectV2ItemById.Item.Id
fmt.Println("Added item to project with item ID:", newItemId)
```

At this point, the issue or PR is now a project item. The returned `newItemId` is the ID of the project item (note: **not** the same as the issue ID, but an ID in the Project’s context, often prefixed with “PVTI…”). You’ll use this project item ID for updating any custom fields on it.

### Adding Draft Issues to a Project

A **Draft Issue** is an issue idea that exists only within the project until you promote it to a real GitHub issue. You can create a draft directly via GraphQL with `addProjectV2DraftIssue`. This will create a new draft item in the project with the given title (and optional body).

**GraphQL Mutation – Add Draft Issue:**

```graphql
mutation($projectId: ID!, $title: String!, $body: String) {
  addProjectV2DraftIssue(input: { projectId: $projectId, title: $title, body: $body }) {
    projectItem {
      id
    }
  }
}
```

This returns a `projectItem.id` for the new draft. The draft will also have a content ID for the draft itself (distinct from a normal Issue ID). You can later convert the draft into a full issue (via UI or possibly the `convertProjectV2DraftIssueToIssue` mutation, if available).

Using Go:

```go
draftReq := graphql.NewRequest(`
mutation($project: ID!, $title: String!, $body: String) {
  addProjectV2DraftIssue(input:{ projectId:$project, title:$title, body:$body }) {
    projectItem { id }
  }
}`)
draftReq.Var("project", projectId)
draftReq.Var("title", "New Idea: Improve Documentation")
draftReq.Var("body", "We should improve the README with ...")
draftReq.Header.Set("Authorization", "Bearer "+token)

var draftResp struct {
    AddProjectV2DraftIssue struct {
        ProjectItem struct{ Id string }
    }
}
err := client.Run(ctx, draftReq, &draftResp)
// handle err
draftItemId := draftResp.AddProjectV2DraftIssue.ProjectItem.Id
fmt.Println("Created draft issue with item ID:", draftItemId)
```

After this, the project has a new draft item. You could update its fields like any other item. (To later convert it to a real issue, you would use a different mutation not covered here, or do it manually in the UI.)

### Updating Field Values for Project Items

Once an item is in the project, you can set or update its field values using `updateProjectV2ItemFieldValue`. This mutation is quite flexible: it takes the project ID, the item ID, the field ID, and a value input that can be a Text, Number, Date, SingleSelectOptionId, or IterationId depending on the field’s type.

**General GraphQL Mutation – Update a Field Value:**

```graphql
mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $fieldValue: ProjectV2FieldValue!) {
  updateProjectV2ItemFieldValue(input: {
    projectId: $projectId
    itemId: $itemId
    fieldId: $fieldId
    value: $fieldValue
  }) {
    projectV2Item { id }
  }
}
```

Here, `ProjectV2FieldValue` is an input type that can contain one of: `text: String`, `number: Float`, `date: Date`, `singleSelectOptionId: String`, or `iterationId: String`. You supply the appropriate one based on the field’s data type.

**Examples:**

* For a text field, `value: { text: "New text value" }`.
* For a number field, `value: { number: 42 }`.
* For a date field, `value: { date: "2025-07-01" }` (dates are in ISO YYYY-MM-DD format).
* For a single-select field, `value: { singleSelectOptionId: "<option-id>" }`.
* For an iteration field, `value: { iterationId: "<iteration-id>" }`.

We will cover specific examples in the next section (Working with Custom Fields). The mutation returns the updated `projectV2Item` (often you only need to know the call succeeded – the returned item ID will match the input item).

In Go, you can execute this by constructing the mutation query string with variables, similar to above. For example, to set a text field:

```go
updateReq := graphql.NewRequest(`
mutation($proj: ID!, $item: ID!, $field: ID!, $val: String!) {
  updateProjectV2ItemFieldValue(input:{
    projectId: $proj, itemId: $item, fieldId: $field,
    value: { text: $val }
  }) {
    projectV2Item { id }
  }
}`)
updateReq.Var("proj", projectId)
updateReq.Var("item", newItemId)
updateReq.Var("field", textFieldId)
updateReq.Var("val", "Updated text value")
```

For a number or date, use appropriate GraphQL scalar types (`Float!` for number, `Date!` for date). For single select, you’d have `$opt: ID!` and use `value: { singleSelectOptionId: $opt }`.

**Note:** You cannot use this mutation to update built-in issue fields like Assignees, Labels, Milestone, or Repository via the project – those must be updated via their own mutations on the issue/pull request itself (e.g. `addAssigneesToAssignable`, `updateIssue`, etc.). The project will reflect those changes automatically.

After updating field values, you may want to verify by querying the item’s fieldValues again or handle any errors (for example, if you use an invalid option ID, the API would return an error).

In summary, the flow to add and update an item is:

1. Add the item to project (get itemId).
2. For each field you want to set, call updateProjectV2ItemFieldValue with that itemId and field’s id & appropriate value.

Next, we’ll go through concrete examples for each field type.

## Creating GitHub Issues and Linking to Projects

Often, you might want to create a new GitHub Issue and immediately add it to a project. This involves two steps: (a) create the issue in a repository, (b) add the new issue to the project.

GitHub’s GraphQL provides a `createIssue` mutation to open a new issue. To use it, you need the repository’s ID and the issue details.

**GraphQL Mutation – Create a new Issue:**

```graphql
mutation($repoId: ID!, $title: String!, $body: String) {
  createIssue(input: { repositoryId: $repoId, title: $title, body: $body }) {
    issue {
      id
      url
      number
    }
  }
}
```

This will create an issue and return its node `id`, URL, and number. (You can also specify other fields in `createIssue` input like `assigneeIds`, `labelIds`, etc., if desired.)

**Step 1: Create the Issue** (Go code):

```go
createIssueReq := graphql.NewRequest(`
mutation($repo: ID!, $title: String!, $body: String) {
  createIssue(input:{ repositoryId: $repo, title: $title, body: $body }) {
    issue { id number url }
  }
}`)
createIssueReq.Var("repo", repositoryId)
createIssueReq.Var("title", "New API Bug")
createIssueReq.Var("body", "Details about the bug...")
createIssueReq.Header.Set("Authorization", "Bearer "+token)

var issueResp struct {
    CreateIssue struct {
        Issue struct {
            Id     string
            Number int
            Url    string
        }
    }
}
if err := client.Run(ctx, createIssueReq, &issueResp); err != nil {
    log.Fatal("CreateIssue failed:", err)
}
newIssueId := issueResp.CreateIssue.Issue.Id
fmt.Printf("Created issue #%d (%s)\n", issueResp.CreateIssue.Issue.Number, issueResp.CreateIssue.Issue.Url)
```

To get `repositoryId`, you could query `repository(owner:"X", name:"Y"){ id }` beforehand (the Medium “cheatsheet” suggests doing that as Query #8).

**Step 2: Add the new Issue to the Project:**

Now use `addProjectV2ItemById` with the project ID and the `newIssueId` from above. This is exactly the same as described in section 4.1. You would run that mutation and link the issue. For example:

```go
addReq.Var("project", projectId)
addReq.Var("content", newIssueId)
// ... run the request ...
fmt.Println("Added new issue to project, item ID:", addResp.AddProjectV2ItemById.Item.Id)
```

At this point, you might also update some project fields on that item if needed (e.g., set the Status field to “Todo” or assign an estimate). You would use the returned project item ID for those field updates.

*Pro tip:* While GraphQL doesn’t support a single transaction combining createIssue and addProjectItem, you could perform them sequentially in a script or even concurrently (with proper handling). If you use GitHub Actions or similar, ensure the PAT used has both `repo` (or appropriate issue scope) and `project` scopes if doing both actions.

## Working with Custom Fields and Options

In this section, we focus on how to handle each type of custom field through the API – including updating their values on items and managing the fields themselves.

### Updating Text, Number, and Date Fields

Text, Number, and Date are straightforward scalar types in Projects. They are all treated similarly in the GraphQL schema (as `ProjectV2Field` for their field definition, and `ProjectV2ItemFieldTextValue`, `...NumberValue`, `...DateValue` for item values).

To update:

* **Text field**: use `value: { text: "your string" }`.
* **Number field**: use `value: { number: 123 }` (numbers are typically floats; use an integer or float as needed).
* **Date field**: use `value: { date: "YYYY-MM-DD" }` format (the GraphQL `Date` scalar expects ISO date strings).

**Example:** Update a text field “Notes” on a given project item:

```go
func updateTextField(projectId, itemId, fieldId string, text string) error {
    req := graphql.NewRequest(`
    mutation($proj: ID!, $item: ID!, $field: ID!, $text: String!) {
      updateProjectV2ItemFieldValue(input:{
        projectId: $proj, itemId: $item, fieldId: $field,
        value: { text: $text }
      }) {
        projectV2Item { id }
      }
    }`)
    req.Var("proj", projectId)
    req.Var("item", itemId)
    req.Var("field", fieldId)
    req.Var("text", text)
    req.Header.Set("Authorization", "Bearer "+token)
    return client.Run(ctx, req, nil)  // we can ignore the response if not needed
}
```

This uses the text value in the mutation. For a number field, you’d change the variable type to Float and use `value: { number: $num }`. For date, variable type would be `Date` (which in practice is just a string in YYYY-MM-DD format) and use `value: { date: $date }`.

After such a mutation, the API returns the `projectV2Item.id` if successful, which confirms the item was updated.

### Updating Single-Select Fields

Single-select fields (think of dropdown picklists) require an **option ID** to set their value. When we listed the fields earlier, we gathered the option IDs for each single-select field. To update a single-select field on an item, use the `singleSelectOptionId` in the value.

**Example:** If there’s a single-select field “Status” and we want to set it to “In Progress”, we need the option ID for “In Progress” (say it’s `"47fc9ee4"` from our fields query). The mutation would be:

```graphql
mutation {
  updateProjectV2ItemFieldValue(input:{
    projectId: "PVT_xxx", itemId: "PVTI_yyy", fieldId: "PVTSSF_zzz",
    value: { singleSelectOptionId: "47fc9ee4" }
  }) {
    projectV2Item { id }
  }
}
```

In code, with variables:

```go
req := graphql.NewRequest(`
mutation($proj: ID!, $item: ID!, $field: ID!, $opt: String!) {
  updateProjectV2ItemFieldValue(input:{
    projectId:$proj, itemId:$item, fieldId:$field,
    value:{ singleSelectOptionId: $opt }
  }) {
    projectV2Item { id }
  }
}`)
req.Var("proj", projectId)
req.Var("item", itemId)
req.Var("field", statusFieldId)
req.Var("opt", inProgressOptionId)
```

Notice `$opt` is a **String** (the option IDs are not full global IDs, but shorter strings – still treat them as opaque strings). The above sets the field to the desired option.

If you mistakenly use an option ID that doesn’t belong to that field or is invalid, you’ll get a GraphQL error. Always retrieve the latest option IDs if in doubt.

Single-select updates are confirmed by retrieving the item’s field value: it will show the `name` of the option and match the selection we made.

**TIP:** The API currently does not allow using the option’s name directly in the mutation – you must use the ID. Our client could abstract this by mapping option name to ID (from the earlier fields query) to make it easier for users to specify the value.

### Updating Iteration Fields

Iteration fields represent a date range (like a sprint). To set an iteration field on an item, you need the **iteration ID** for the specific iteration (e.g., the “Sprint 1” that spans certain dates). The field’s configuration (retrieved via the fields query) gives you all iteration IDs (active and past).

Using that ID, you call `updateProjectV2ItemFieldValue` with `value: { iterationId: "iteration-id-here" }`.

**Example Mutation:**

```graphql
mutation {
  updateProjectV2ItemFieldValue(input:{
    projectId: "PVT_xxx", itemId: "PVTI_yyy", fieldId: "PVTIF_zzz",
    value: { iterationId: "cfc16e4d" }
  }) {
    projectV2Item { id }
  }
}
```

Which in variables form:

```graphql
mutation($proj: ID!, $item: ID!, $field: ID!, $iter: String!) {
  updateProjectV2ItemFieldValue(input:{
    projectId:$proj, itemId:$item, fieldId:$field, value:{ iterationId: $iter }
  }) {
    projectV2Item { id }
  }
}
```

In Go, you’d supply the iteration ID string (e.g., `"cfc16e4d"` from our earlier example). After this, the item is assigned to that iteration (which affects the Project’s iteration field value). You can verify by querying the item’s `fieldValueByName(name:"Iteration")` or similar – it should show the iteration’s title/startDate now.

This mutation works for both active and completed iterations (you can set an item to a past iteration, presumably).

### Listing and Managing Single-Select Options

For single-select fields, you may want to programmatically get all possible options (we did this in the fields query). If you want to **update the options themselves** (e.g., add a new option or rename one), the GraphQL API now provides ways to do so.

* **Listing options**: as shown, `ProjectV2SingleSelectField.options` gives id and name for each option.
* **Adding an option**: There isn’t a dedicated “add single select option” mutation, but you can use `updateProjectV2Field` mutation on the field’s configuration. Specifically, `updateProjectV2FieldInput` allows you to specify new options for a single-select field. In practice, you might have to pass the entire new options list (including existing ones plus the new one). Alternatively, a more direct way is to use `createProjectV2Field` to create a brand new field with the desired options (but that’s for new fields, not updating existing).

As of the latest API:

* **CreateProjectV2Field**: lets you create a new custom field on a project with specified options if it’s a single-select.
* **UpdateProjectV2Field**: lets you rename a field or, in the case of single-select, possibly change its options. The input for update includes a field ID and likely similar parameters (though documentation is sparse on adding options via update, it may support providing a new set of `singleSelectOptions`). If the API doesn’t yet allow adding options via update, the workaround is manual or recreating the field.

Because managing field options is more of a project setup task, many clients simply document the option IDs or retrieve them. Our Go client could provide helper functionality like:

```go
// Pseudo-code
options := getSingleSelectOptions(projectId, fieldId)
if !options.Has("New Option") {
    err := addSingleSelectOption(projectId, fieldId, "New Option")
    // This might involve calling updateProjectV2Field with new options list
}
```

Ensure you check the GitHub docs or changelog for updates on option management. (By the time of writing, adding options via API was a requested feature and may be available via `updateProjectV2Field`.)

For completeness, to **rename a custom field**, you can use `updateProjectV2Field` with a new name in the input (and it returns the updated field object) (renaming built-in fields like Title is not applicable, but for a custom field you created, it’s possible).

### Creating New Custom Fields via API

If your client needs to set up new fields on a project (instead of using the UI to create them), the GraphQL API supports this via `createProjectV2Field`. This mutation requires the project ID, field name, data type, and for single-select, a list of option values (each with name, and optionally color). The `dataType` is an enum of `ProjectV2CustomFieldType` – likely values like `TEXT`, `NUMBER`, `DATE`, `SINGLE_SELECT`, `ITERATION`.

**Example:** Create a new single-select field "Priority" with options.

```graphql
mutation($project: ID!) {
  createProjectV2Field(input: {
    projectId: $project
    dataType: SINGLE_SELECT
    name: "Priority"
    singleSelectOptions: [
      { name: "Low" },
      { name: "Medium" },
      { name: "High" }
    ]
  }) {
    projectV2Field {
      id
      name
      dataType
    }
  }
}
```

This will add a new field to the project and return its configuration (the returned object is of type `ProjectV2FieldConfiguration`). You could then query its `options` if needed to get the IDs of the newly created options.

**Note:** Creating iteration fields via API may require providing an `iterationConfiguration` (like duration of iteration cycles). For text/number/date, no extra config is needed.

Deleting a field is also possible (`deleteProjectV2Field` mutation) if needed, and similarly `deleteProjectV2Item` to remove an item from a project.

*Caution:* Adding or removing fields is typically something done during project setup. If you’re building a tool to sync or migrate projects, these mutations are invaluable. Always ensure your token has permission (only project admins can add/remove fields).

## Navigating the GraphQL Schema in Go

Working with GitHub’s GraphQL can be daunting due to the rich schema. Here are some tips for schema navigation and making your Go development smoother:

* **Use the GitHub GraphQL Explorer:** GitHub provides a GraphQL Explorer (at [https://docs.github.com/en/graphql/overview/explorer](https://docs.github.com/en/graphql/overview/explorer)) where you can interactively query the schema, see documentation, and test queries. This is great for discovering field names, types, and testing queries before coding them.

* **Download the Schema**: GitHub’s GraphQL API schema can be retrieved via introspection. You could use a tool to download the schema JSON and then use GraphQL codegen tools. For example, `github.com/Khan/genqlient` can use `.graphql` query files and generate Go types for you, using the schema for type-checking.

* **Leverage GraphQL type-safe clients:** Libraries like `hasura/go-graphql-client` (inspired by ShurcooL’s library) allow you to define Go structs with tags corresponding to GraphQL queries. This can catch mistakes (like querying a non-existent field) at compile time. It uses Go’s reflection and struct tags for queries, which can be very convenient once you get used to it.

* **GraphQL Voyager or Schema Viewer:** There are tools such as [GraphQL Voyager](https://github.com/graphql-kit/graphql-voyager) that can visualize the schema as an interactive graph. The Apollo Studio link for GitHub’s schema indicates you can browse types like `ProjectV2` (which shows it has a `views` connection, etc.). These tools help understand relationships between types.

* **Use Interfaces and Unions in code:** GraphQL uses **interfaces** (e.g., `ProjectV2FieldCommon`) and **union types** (e.g., `ProjectV2ItemFieldValue` union of various value types). In Go, you might model this with interface types or simple type switching. For instance, after unmarshalling a `fieldValues.nodes`, you might switch on the `__typename` field to decide how to interpret the data. This is one reason some prefer generated clients – they can generate union type handling.

* **Refer to Documentation and Community:** The official GraphQL reference docs list all objects, inputs, and mutations (e.g., the Projects-related mutations we discussed are documented). If something isn’t working, checking if the field or mutation exists in the latest docs is a good idea. Community forums, discussions, and the GitHub API changelog can also provide insight into new features (for example, the ability to create fields via API was added after initial beta).

* **Testing with `gh api`:** If you have the GitHub CLI (`gh`) installed, you can also test GraphQL quickly with commands like `gh api graphql -f query='<query>'`. This uses your logged-in credentials. It’s another handy way to prototype queries/mutations (the GitHub docs often show `gh api graphql` examples which you can mimic).

By combining these strategies, you can efficiently navigate and utilize the GraphQL schema. When writing your Go client, make small test calls to ensure you have the correct IDs and query shapes before writing larger pieces of logic.

## Conclusion

Building a GitHub Projects (v2 Beta) GraphQL client in Go involves understanding both the GraphQL schema and using Go libraries to execute queries and mutations. We covered how to authenticate with a token (using appropriate scopes), query project details, list custom fields and options, and perform common mutations like adding items and updating field values. We also looked at handling different field types – text, number, date, single-select, iteration – and how to manage project configurations (views and fields) at the API level.

With the provided examples, you can create and link issues to project boards, automate project updates, and even extend your tool to set up projects or migrate data. The Go code can be further abstracted into utility functions or a higher-level API client structure (for instance, methods like `AddItem(projectID, contentID)`, `UpdateField(itemID, fieldID, value)` etc.). Remember to handle errors and rate limits – GraphQL has a rate limit similar to REST, and large projects might require pagination logic.

Finally, keep an eye on GitHub’s changelogs and docs for Projects API improvements. The Projects (v2) API is actively evolving (for example, new mutations for templates, status updates, or field management). By staying updated and using robust schema exploration techniques, your Go client will remain effective and up-to-date with GitHub’s latest features.

**Sources:**

* GitHub GraphQL Docs – *Using the API to manage Projects*
* GitHub Changelog – *Projects GraphQL API updates*
* Community & Developer Insights – *ProjectV2 field usage and queries*
* GraphQL API Reference – *Mutations and Inputs for Projects*

