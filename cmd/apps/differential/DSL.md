# Differential DSL - DSL for Code Modifications

## Overview

This YAML DSL (Domain Specific Language) specifies a series of code modifications within a file, leveraging a content-contextual approach. Each operation within the DSL is documented with a comment explaining its purpose, prioritizing clarity on the intent behind each change.

## Structure of the DSL

The DSL is structured as a YAML document containing the path to the target source file and a list of change operations to be applied. Each operation begins with a comment field to explain its intent, followed by specific action details.

### General YAML Structure

```yaml
path: <path_to_source_file>
changes:
  - comment: <explanation_for_the_change>
    action: <action_type>
    # ...<additional_fields_based_on_action_type>

  # ...<more_operations_as_needed>
```

### Fields

- `path`: A string indicating the path to the source file on which to apply the changes.
- `changes`: A list of change operations, where each operation is a dictionary containing:
  - `comment`: A string providing a brief explanation of the reason for the change, placed as the first field for clarity and emphasis.
  - `action`: A string specifying the type of change ("insert", "delete", "move", "replace", "prepend", or "append").
  - Additional fields corresponding to the specified action type, detailed in the subsequent sections.

## Action Types and Corresponding Fields

Each action type within the "changes" list starts with a "comment" field, followed by the type of action and the relevant details for that action. Below are the specifics for each action type.

### 1. Replace

- `comment`: Explanation for why the replacement is occurring.
- `action`: Must be "replace".
- `old`: |
  The exact block of code that will be replaced.
- `new`: |
  The new code that will replace the old segment.

### 2. Insert

- `comment`: Reason for the insertion of new code.
- `action`: Must be "insert".
- `content`: |
  The new code to be inserted.
- `above`: Context specifying where the new content will be placed relative to existing lines.
  - Be sure that `above` refers to text lines, not code elements.

### 3. Delete

- `comment`: Explanation for the deletion.
- `action`: Must be "delete".
- `content`: |
  The exact block of code to be deleted.

### 4. Move

- `comment`: Rationale behind moving a particular code block.
- `action`: Must be "move".
- `content`: |
  The exact block of code to be moved.
- `above`: Context specifying where the content will be placed relative to existing lines.
  - Be sure that `above` refers to text lines, not code elements.

### 5. Prepend

- `comment`: Reason for adding new code at the beginning of the file.
- `action`: Must be "prepend".
- `content`: |
  The new code to be added at the top of the file.

### 6. Append

- `comment`: Reason for adding new code at the end of the file.
- `action`: Must be "append".
- `content`: |
  The new code to be added at the end of the file.

## Example of DSL Usage

Here is an example illustrating the usage of various actions, each starting with a comment:

```yaml
path: source_file.py
changes:
  - comment: Refactor function for better performance
    action: replace
    old: |
      def outdated_function():
          pass
    new: |
      def updated_function():
          print('Enhanced functionality')

  - comment: Add initialization function
    action: prepend
    content: |
      def initialize():
          print('Starting application')

  - comment: Add cleanup function at the end
    action: append
    content: |
      def cleanup():
          print('Shutting down')

  - comment: Remove deprecated function
    action: delete
    content: |
      def unnecessary_function():
          pass

  - comment: Improve code organization
    action: move
    content: |
      def misplaced_function():
          print('Operational')
    above: |
      def target_location():
      
  - comment: Add new functionality
    action: insert
    content: |
      def new_function():
          print('New functionality')
    above: |
      def existing_function():
```
