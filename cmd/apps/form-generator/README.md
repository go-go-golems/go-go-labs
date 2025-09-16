# Google Forms Generator

A command-line tool that generates Google Forms from Uhoh Wizard DSL files. The generator transforms multi-step wizard definitions into structured Google Forms with proper sections, question types, and validation rules.

## Overview

The form-generator bridges the gap between declarative form definitions and Google Forms creation. Instead of manually building forms through the Google Forms UI, you can define your form structure in YAML using the Uhoh Wizard DSL, then automatically generate the corresponding Google Form with proper question types, sections, and formatting.

**Key capabilities:**
- Converts Uhoh Wizard steps into Google Forms sections with page breaks
- Maps DSL field types to appropriate Google Forms question types
- Supports creating new forms, updating them in place, and exporting a form back to the Wizard DSL
- Handles authentication via OAuth2 with token persistence
- Preserves field descriptions, required flags, and validation rules

## Installation

Build the form-generator from source:

```bash
go build ./go-go-labs/cmd/apps/form-generator
```

## Authentication Setup

Before using the form-generator, you need to set up Google OAuth2 credentials:

1. **Create a Google Cloud Project** and enable the Google Forms API
2. **Create OAuth2 credentials** (Desktop application type)
3. **Download the credentials JSON** and save it as `~/.google-form/client_secret.json`

The tool will automatically handle the OAuth flow on first use, storing your token at `~/.google-form/token.json` for future runs.

## Usage

### Creating a New Form

Generate a new Google Form from a wizard definition:

```bash
form-generator generate --wizard survey.yaml --create
```

### Updating an Existing Form

Replace the current form items with the content defined in the DSL:

```bash
form-generator generate --wizard survey.yaml --form-id 1ABC123def456GHI
```

The command removes any existing items in the target form before recreating them according to the DSL, so you always end up with an exact match.

### Fetching an Existing Form

Export an existing Google Form back into a Wizard DSL YAML document:

```bash
form-generator fetch --form-id 1ABC123def456GHI --output survey.yaml
```

The resulting YAML can be inspected or used as a starting point for subsequent edits.

### Fetching Form Submissions

Download all submissions for a form and stream them as structured rows that can be formatted by Glazed:

```bash
form-generator fetch-submissions --form-id 1ABC123def456GHI --output json
```

Use the standard Glazed formatting flags (`--output`, `--output-file`, etc.) to emit tables, CSV, JSON, or templated reports. Each row includes the submission metadata, the corresponding wizard step and field identifiers, the raw answer value(s), and uploaded file metadata when available.

### Command Options

- `--wizard`: Path to Uhoh wizard YAML file (required)
- `--create`: Create a new form (mutually exclusive with `--form-id`)
- `--form-id`: Update existing form with this ID
- `--title`: Override form title
- `--description`: Override form description
- `--debug`: Enable debug logging

## Wizard DSL to Google Forms Mapping

The form-generator maps Uhoh Wizard concepts to Google Forms elements:

### Step Types

- **Info steps** → Static text items with page breaks
- **Form steps** → Question sections with page breaks
- **Decision steps** → Single-choice radio questions

### Field Types

- `input` → Short answer text question
- `text` → Paragraph text question
- `select` → Single-choice radio buttons
- `multiselect` → Multiple-choice checkboxes
- `confirm` → Yes/No radio buttons

### Field Properties

- `title` → Question title
- `description` → Question help text
- `required` → Required field validation
- `options` → Choice options for select/multiselect

## Example Wizard File

```yaml
name: Customer Feedback Survey
description: Gather feedback about our services
theme: Default

steps:
  - id: welcome
    type: info
    title: Welcome
    content: |
      Thank you for taking our customer feedback survey.
      Your responses help us improve our services.

  - id: feedback
    type: form
    title: Your Feedback
    form:
      groups:
        - name: Basic Information
          fields:
            - type: input
              key: customer_name
              title: Your Name
              required: true
              
            - type: select
              key: satisfaction
              title: Overall Satisfaction
              required: true
              options:
                - label: "Very Satisfied"
                  value: "very_satisfied"
                - label: "Satisfied"
                  value: "satisfied"
                - label: "Neutral"
                  value: "neutral"
                - label: "Dissatisfied"
                  value: "dissatisfied"
                  
            - type: text
              key: comments
              title: Additional Comments
              description: Please share any specific feedback
```

This wizard generates a Google Form with:
- A welcome section with introductory text
- A page break before the feedback section
- A required short-answer field for the customer name
- A required single-choice question for satisfaction rating
- An optional paragraph field for additional comments

## Output

After successful generation, the tool outputs:
- The Google Form ID
- A direct link to fill out the form
- Debug information (when `--debug` is enabled)

```
Created form: 1ABC123def456GHI789jkl
Fill-in link: https://docs.google.com/forms/d/e/1FAIpQLSd.../viewform
```

## Limitations

- File upload questions require Google Workspace accounts
- Complex conditional logic is not supported
- Form themes and advanced styling are not configurable
- Validation rules are limited to required/optional fields

## Troubleshooting

**Authentication errors:** Ensure your credentials file is properly configured and the Google Forms API is enabled in your Google Cloud project.

**Permission errors:** Verify that your OAuth2 application has the correct scopes (`https://www.googleapis.com/auth/forms.body`).

**Port conflicts:** The OAuth callback server uses port 8080 by default. Ensure this port is available during authentication.

For detailed error information, run with the `--debug` flag to see the complete request/response flow.
