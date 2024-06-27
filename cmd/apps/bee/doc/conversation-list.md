---
Title: Querying Conversation Lists with Bee.Computer SDK CLI
Slug: querying-conversation-lists
Short: Learn how to query conversation lists using the Bee.Computer SDK CLI with practical examples.
Topics:
- CLI
- Conversations
- Querying
Commands:
- conversation list
Flags:
- --limit
- --output
- --page
- --sort-by
- --filter
- --fields
IsTopLevel: false
ShowPerDefault: false
SectionType: GeneralTopic
---

# Querying Conversation Lists with Bee.Computer SDK CLI

The Bee.Computer SDK CLI provides a command-line interface to interact with conversation data. This document will guide you through the process of querying conversation lists, offering practical examples to enhance your workflow.

## Overview

Querying conversation lists can be done using the `conversation list` command. This command supports various flags that allow you to customize the output, such as limiting the number of conversations, choosing the output format, and applying filters or sorting.

## Examples

Here are some examples of how you can use the `conversation list` command to query conversation lists effectively:

### Basic Listing

To list the first 10 conversations (default behavior):

```bash
bee conversation list
```

### Custom Limit and Pagination

To list the first 5 conversations:

```bash
bee conversation list --limit 5
```

To get the second page of conversations, assuming a limit of 5:

```bash
bee conversation list --page 2 --limit 5
```

### Output Formats

To output the list in YAML format:

```bash
bee conversation list --output yaml
```

To output the list in JSON format:

```bash
bee conversation list --output json
```

### Sorting

To sort the conversations by start time in descending order:

```bash
bee conversation list --sort-by=-start_time
```

### Filtering Fields

To include only specific fields in the output:

```bash
bee conversation list --fields id,start_time,short_summary
```

To exclude certain fields from the output:

```bash
bee conversation list --filter end_time,device_type
```

### Combining Flags

To list conversations sorted by start time with a custom limit, output as YAML, including only specific fields:

```bash
bee conversation list --sort-by=start_time --limit 3 --output yaml --fields id,start_time,short_summary
```

