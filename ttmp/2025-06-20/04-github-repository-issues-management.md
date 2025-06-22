Here’s the short list of GraphQL building-blocks you need, plus minimal copy-/-paste examples.

---

## 1. Skip the draft & open a *real* Issue directly

Use the **`createIssue`** mutation.  Everything hangs off a single `CreateIssueInput` object where you pass the repo ID, title, body, etc.  Note the handy `labelIds` and `projectIds` arrays. ([docs.github.com][1], [docs.github.com][2])

```graphql
mutation CreateIssue(
  $repositoryId: ID!
  $title: String!
  $body: String
  $labelIds: [ID!]
  $projectIds: [ID!]   # optional – auto-adds the new issue to a Project V2
) {
  createIssue(
    input: {
      repositoryId: $repositoryId
      title: $title
      body: $body
      labelIds: $labelIds       # any repo label IDs
      projectIds: $projectIds   # the Project V2 ID if you want it linked immediately
    }
  ) {
    issue {
      id
      number
      url
    }
  }
}
```

*Tip*: If you don’t supply `projectIds`, you can attach the issue afterwards with `addProjectV2ItemById`.

---

## 2. Convert an existing **draft item** inside a Project V2

Draft items live only inside the project; to turn one into a full Issue use **`convertProjectV2DraftIssueItemToIssue`**.  You give it the project-wide *item* ID (not the draft’s content ID) plus the target repo. ([docs.github.com][1])

```graphql
mutation ConvertDraft(
  $projectId: ID!
  $itemId: ID!           # the PVTI_… ID of the draft item in the project
  $repositoryId: ID!     # where the new issue should live
) {
  convertProjectV2DraftIssueItemToIssue(
    input: {
      projectId: $projectId
      itemId: $itemId
      repositoryId: $repositoryId
    }
  ) {
    item {                # same project item, now pointing at a real Issue
      id
      content {
        ... on Issue {
          id
          number
          url
        }
      }
    }
  }
}
```

GitHub keeps the new Issue linked to the original project item for you.

---

## 3. Adding / removing labels after creation

* Add labels: **`addLabelsToLabelable`**
* Remove labels: **`removeLabelsFromLabelable`**

Both take the Issue’s node ID (`labelableId`) plus a list of label IDs. ([docs.github.com][1])

```graphql
mutation AddLabels($issueId: ID!, $labelIds: [ID!]!) {
  addLabelsToLabelable(
    input: { labelableId: $issueId, labelIds: $labelIds }
  ) {
    labelable {
      ... on Issue {
        id
        labels(first: 10) { nodes { name } }
      }
    }
  }
}
```

---

## 4. Discovering the labels you can use

```graphql
query RepoLabels($owner: String!, $name: String!, $after: String) {
  repository(owner: $owner, name: $name) {
    labels(first: 100, after: $after) {
      nodes { id name color description }
      pageInfo { hasNextPage endCursor }
    }
  }
}
```

Paginate with `after` until `hasNextPage` is false.

If you need to mint a brand-new label first, there’s **`createLabel`** (same repo ID + name + color). ([docs.github.com][1])

---

### Putting it together

1. **Get IDs you’ll need once**
   *repoId, projectId, labelIds* etc. via simple queries (GraphQL explorer or `gh api graphql`).

2. **Create the Issue** with `createIssue` *or* raise a draft then `convertProjectV2DraftIssueItemToIssue`.

3. **Tweak metadata** (labels, assignees, milestones, custom project fields) via the relevant mutations.

That’s it—you’re no longer limited to draft items and you can stay 100 % in GraphQL. Happy automating!

[1]: https://docs.github.com/en/graphql/reference/mutations "Mutations - GitHub Docs"
[2]: https://docs.github.com/en/graphql/reference/input-objects?utm_source=chatgpt.com "Input objects - GitHub Docs"

