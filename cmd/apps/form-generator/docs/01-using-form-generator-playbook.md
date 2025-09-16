---
Title: Form Generator Playbook
Slug: form-generator-playbook
Short: Practical guide for listing, auditing, fetching, and generating Google Forms with form-generator.
Topics:
- google-forms
- automation
- playbook
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: Playbook
---

# Form Generator Playbook

This playbook walks through every `form-generator` capability, starting with discovery and ending with full regeneration. Follow it sequentially to explore Drive inventory, audit responses, export a wizard DSL, and push updates with confidence.

## Prerequisites

Before running any command, ensure you have:

- Go 1.24 or later (the module declares `go 1.24.3`).
- OAuth credentials with Google Forms API enabled (`~/.google-form/client_secret.json`) and a writable token store (`~/.google-form/token.json`).
- A Uhoh Wizard DSL file if you plan to generate or update forms.
- Network access allowed by your runtime environment.

All commands below were smoke-tested with `go run ./cmd/apps/form-generator <verb> --help` to confirm the CLI wiring loads successfully.

## Inspect Available Forms

Listing forms is the fastest way to confirm that credentials are valid and to gather the IDs you'll need for subsequent steps.

### Steps

1. Run the list command with the desired sort order:

    ```bash
    go run ./cmd/apps/form-generator list \
      --sort modified \
      --desc \
      --limit 20 \
      --output table
    ```

2. Approve the OAuth consent screen on first use. The command requests `drive.metadata.readonly` and Forms scopes.
3. Review the table output. Each row contains:
   - `index`: 1-based position in the listing.
   - `id`: the Drive file ID (pass this to other verbs).
   - `name`: form title.
   - `created_time` / `modified_time`: RFC3339 timestamps.
   - `owner_names` / `owner_emails`: owners separated by commas when multiple people manage the form.
   - `web_view_link`: direct link to the live form.
4. Use Glazed filters to slice data further, e.g. `--filter 'modified_time > "2024-12-01"'` or `--select id`.

**Example output (abbreviated):**

```text
INDEX  ID               NAME                    MODIFIED_TIME         WEB_VIEW_LINK
1      1a2b3cFormsID    Quarterly Survey 2024   2024-08-15T10:42:07Z  https://docs.google.com/forms/d/1a2b...
2      4d5e6fFormsID    Beta Signup Form        2024-07-29T18:11:53Z  https://docs.google.com/forms/d/4d5e...
```

## Review Form Submissions

Pulling submissions through Glazed makes it easy to export responses as JSON, CSV, or templated text for analysis or archival.

### Steps

1. Choose a form ID from the previous step.
2. Stream submissions in JSON for post-processing:

    ```bash
    go run ./cmd/apps/form-generator fetch-submissions \
      --form-id 1a2b3cFormsID \
      --output json \
      --debug=false
    ```

3. Each emitted row captures responder metadata plus a single answer. Expect fields like:
   - `response_id`, `submitted_at`, `respondent_email`.
   - `step_id`, `field_key`, and `field_title` populated from DSL metadata tags.
   - `value` or `values` when multiple selections are chosen.
   - `files` array with `file_id`, `file_name`, and `mime_type` for upload questions.
4. Combine rows with Glazed helpers, e.g. `--sort-by respondent_email` or `--select-template '{{.response_id}} -> {{.field_title}}: {{.value}}'`.

> **Tip:** Use `--output table` for a quick console audit or `--output csv --output-file submissions.csv` to ingest into spreadsheets.

## Export a Form to Wizard DSL

Exporting DSL is vital when you need to version-control a form or migrate it between environments.

### Steps

1. Supply the target form ID and, optionally, an output file:

    ```bash
    go run ./cmd/apps/form-generator fetch \
      --form-id 1a2b3cFormsID \
      --output team-onboarding.yaml
    ```

2. The command reconstructs the wizard YAML by reading form items, decoding metadata tags (step and field IDs), and mapping question types.
3. Inspect the resulting YAML to verify metadata, descriptions, and decision steps align with expectations.
4. Commit the exported file to source control so future changes can be diffs on the DSL instead of ad-hoc edits in the UI.

## Generate or Update a Form from DSL

Generation is how you materialise a wizard DSL as a new Google Form or align an existing form with the latest DSL state.

### Create a Brand-New Form

1. Prepare a DSL file (`wizard.yaml`).
2. Run:

    ```bash
    go run ./cmd/apps/form-generator generate \
      --wizard wizard.yaml \
      --create \
      --title "Team Onboarding Wizard" \
      --description "Collect onboarding context from every new teammate."
    ```

3. The command outputs the new `FormId` and responder URL. Save both for distribution and automation.

### Update an Existing Form

1. Ensure the DSL file reflects the desired state.
2. Synchronise the form:

    ```bash
    go run ./cmd/apps/form-generator generate \
      --wizard wizard.yaml \
      --form-id 1a2b3cFormsID \
      --title "Team Onboarding Wizard"
    ```

3. The tool deletes all existing items, recreates them from the DSL (preserving step/field metadata), and then updates title/description when provided.
4. Follow up with `form-generator fetch` if you want to confirm the live form now matches the DSL byte-for-byte.

## Troubleshooting and Tips

- **Authentication:** If commands stall waiting for OAuth, confirm `--server-port` is open and that the redirect page displays “Authentication completed”. Delete `~/.google-form/token.json` to force a refresh.
- **Scopes:** The listing command requires `drive.metadata.readonly`; submissions, fetch, and generate require `forms.body`. `buildFormsAuthenticator` already requests both when needed, so no manual edits are necessary.
- **Metadata collisions:** Always run `generate` after editing DSL. The embedded metadata ensures future exports map back to the original step IDs.
- **Dry runs:** Use `--print-parsed-parameters` to verify flag parsing without contacting Google APIs.
- **Automation:** Because all read commands implement `GlazeCommand`, you can combine them with Glazed features such as `--jq`, `--template`, and `--glazed-limit` for scripted pipelines.

## Next Steps

- Integrate `form-generator list` into scheduled audits to detect stale forms.
- Pipe `fetch-submissions` output into analytics or incident workflows.
- Store exported DSL alongside application code so reviews and deployments stay in lockstep with the UI.

This playbook should give you a dependable workflow for discovering, auditing, exporting, and regenerating Google Forms using `form-generator`.
