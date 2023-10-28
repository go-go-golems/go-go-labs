# Differential DSL - DSL for Code Modifications

## Overview

This JSON DSL (Domain Specific Language) specifies a series of code modifications within a file, leveraging a content-contextual approach. Each operation within the DSL is documented with a comment explaining its purpose, prioritizing clarity on the intent behind each change.

## Structure of the DSL

The DSL is structured as a JSON object containing the path to the target source file and a list of change operations to be applied. Each operation begins with a comment field to explain its intent, followed by specific action details.

### General JSON Structure

```json
{
  "path": "<path_to_source_file>",
  "changes": [
    {
      "comment": "<explanation_for_the_change>",
      "action": "<action_type>",
      ...<additional_fields_based_on_action_type>
    },
    ...<more_operations_as_needed>
  ]
}
```

### Fields

- `path`: A string indicating the path to the source file on which to apply the changes.
- `changes`: An array of change operations, where each operation is an object containing:
    - `comment`: A string providing a brief explanation of the reason for the change, placed as the first field for clarity and emphasis.
    - `action`: A string specifying the type of change ("insert", "delete", "move", or "replace").
    - Additional fields corresponding to the specified action type, detailed in the subsequent sections.

## Action Types and Corresponding Fields

Each action type within the "changes" array starts with a "comment" field, followed by the type of action and the relevant details for that action. Below are the specifics for each action type.

### 1. Replace

- `comment`: Explanation for why the replacement is occurring.
- `action`: Must be "replace".
- `old`: The exact block of code that will be replaced.
- `new`: The new code that will replace the old segment.

### 2. Insert

- `comment`: Reason for the insertion of new code.
- `action`: Must be "insert".
- `content`: The new code to be inserted.
- `destination_above` or `destination_below`: Context specifying where the new content will be placed relative to existing code.

### 3. Delete

- `comment`: Explanation for the deletion.
- `action`: Must be "delete".
- `content`: The exact block of code to be deleted.

### 4. Move

- `comment`: Rationale behind moving a particular code block.
- `action`: Must be "move".
- `content`: The exact block of code to be moved.
- `destination_above` or `destination_below`: Context indicating the new location relative to existing code.

## Example of DSL Usage

Here is an example illustrating the usage of various actions, each starting with a comment:

```json
{
  "path": "source_file.py",
  "changes": [
    {
      "comment": "Refactor function for better performance",
      "action": "replace",
      "old": "def outdated_function():\n    pass",
      "new": "def updated_function():\n    print('Enhanced functionality')"
    },
    {
      "comment": "Extend application capability",
      "action": "insert",
      "content": "def auxiliary_function():\n    print('Supportive operation')",
      "destination_below": "def primary_function():"
    },
    {
      "comment": "Remove deprecated function",
      "action": "delete",
      "content": "def unnecessary_function():\n    pass"
    },
    {
      "comment": "Improve code organization",
      "action": "move",
      "content": "def misplaced_function():\n    print('Operational')",
      "destination_above": "def target_location():"
    }
  ]
}
```