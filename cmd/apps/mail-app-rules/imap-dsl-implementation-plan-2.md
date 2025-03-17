# IMAP DSL Implementation Plan - Phase 2

This document outlines the plan for implementing additional features from the IMAP DSL specification that were not covered in the initial implementation.

## Overview

Phase 1 implemented basic search criteria (date-based and "from" searches) and output formatting. Phase 2 will expand the functionality to include:

1. Additional search criteria
2. Complex search conditions
3. Actions on messages
4. Enhanced output options
5. Error handling improvements

## Implementation Tasks

### 1. Expand Search Criteria

- [x] Refactor *Config to have a Validate() method that checks if the config is valid, and returns clear readable explanations if not

- [x] Implement header-based search
  - [x] Add support for `to`, `cc`, `bcc` search criteria
  - [x] Add support for `subject` and `subject_contains` search criteria
  - [x] Add support for arbitrary header search via `header.name` and `header.value`

- [x] Implement content-based search
  - [x] Add support for `body_contains` search criteria
  - [x] Add support for `text` search criteria (searches in headers and body)

- [x] Implement flag-based search
  - [x] Add support for `flags.has` criteria
  - [x] Add support for `flags.not_has` criteria

- [x] Implement size-based search
  - [x] Add support for `size.larger_than` criteria with unit parsing (B, K, M, G)
  - [x] Add support for `size.smaller_than` criteria with unit parsing

### 2. Implement Complex Search Conditions

- [ ] Add support for logical operators in search criteria
  - [ ] Implement `operator: and` with nested conditions
  - [ ] Implement `operator: or` with nested conditions
  - [ ] Implement `operator: not` with nested conditions
  - [ ] Support nested operators for complex queries

### 3. Implement Actions on Messages

- [ ] Implement flag operations
  - [ ] Add support for `actions.flags.add` to add flags to messages
  - [ ] Add support for `actions.flags.remove` to remove flags from messages

- [ ] Implement mailbox operations
  - [ ] Add support for `actions.move_to` to move messages to another mailbox
  - [ ] Add support for `actions.copy_to` to copy messages to another mailbox

- [ ] Implement delete operations
  - [ ] Add support for `actions.delete` to delete messages
  - [ ] Add support for `actions.delete.trash` to move messages to trash

- [ ] Implement export operations
  - [ ] Add support for `actions.export.format` (eml, mbox)
  - [ ] Add support for `actions.export.directory` for export location
  - [ ] Add support for `actions.export.filename_template` with variable substitution

### 4. Enhance Output Options

- [x] Improve message body handling
  - [x] Add support for `min_length` option in body retrieval
  - [x] Merge `body` and `mime_parts` fields into a unified content handling system
  - [ ] Add support for `strip_quotes` option in body retrieval
  - [ ] Implement proper MIME part selection for multipart messages

- [x] Implement MIME parts listing
  - [x] Add support for `mime_parts` field to list content types of message parts
  - [x] Add support for filtering MIME parts by mode (`text_only`, `full`, `filter`)
  - [x] Add support for filtering MIME parts by specific types
  - [x] Add support for showing MIME part content with `show_content` option
  - [x] Add support for content length control with `max_length` and `min_length`

- [ ] Implement header selection
  - [ ] Add support for `headers.include` to select specific headers

- [ ] Implement attachment handling
  - [ ] Add support for `attachments.list` to list attachments
  - [ ] Add support for `attachments.download` to download attachments
  - [ ] Add support for `attachments.save_path` to specify download location
  - [ ] Add support for `attachments.types` to filter by MIME type

- [ ] Improve table output format
  - [ ] Implement proper table formatting with column alignment
  - [ ] Add support for custom column headers

### 5. Improve Error Handling and Validation

- [ ] Enhance YAML validation
  - [ ] Add comprehensive validation for all new fields
  - [ ] Provide clear error messages for invalid configurations

- [ ] Implement better error reporting
  - [ ] Add detailed error messages for IMAP server errors
  - [ ] Add context to error messages (e.g., which message caused the error)

- [ ] Add validation for actions
  - [ ] Validate mailbox existence for move/copy operations
  - [ ] Validate flag names against IMAP standards
  - [ ] Validate file paths for export operations

### 6. Update Core Components

- [ ] Update data structures
  - [ ] Extend `Rule` struct to include actions
  - [ ] Add new structs for complex search conditions
  - [ ] Add structs for action configurations

- [ ] Update parser
  - [ ] Enhance YAML parsing to handle new fields
  - [ ] Add validation for new fields

- [ ] Update processor
  - [ ] Modify `ProcessRule` to execute actions after search/output
  - [ ] Implement action execution logic

### 7. Documentation and Examples

- [ ] Update documentation
  - [ ] Update README.md with new features
  - [ ] Add documentation for new YAML fields

- [ ] Create new examples
  - [ ] Add example for complex search conditions
  - [ ] Add example for each type of action
  - [ ] Add comprehensive example using multiple features

### 8. Testing

- [ ] Write unit tests for new functionality
  - [ ] Test complex search condition parsing
  - [ ] Test action execution
  - [ ] Test enhanced output options

- [ ] Write integration tests
  - [ ] Test end-to-end workflow with actions
  - [ ] Test with real IMAP server

## Implementation Strategy

1. Start with expanding search criteria, as this builds on existing functionality
2. Implement complex search conditions next, as this affects the search process
3. Add actions functionality, which is a major new feature
4. Enhance output options to provide more flexibility
5. Improve error handling throughout the codebase
6. Update documentation and examples last

## Dependencies

- go-imap/v2 library for IMAP operations
- yaml.v3 for YAML parsing
- Additional libraries may be needed for:
  - MIME handling for attachments
  - Template processing for filename templates
  - Table formatting for improved output

## Timeline Estimate

- Expanding search criteria: 2-3 days
- Complex search conditions: 2-3 days
- Actions implementation: 3-4 days
- Output enhancements: 2-3 days
- Error handling improvements: 1-2 days
- Documentation and examples: 1-2 days
- Testing: 2-3 days

Total estimated time: 2-3 weeks of development effort 