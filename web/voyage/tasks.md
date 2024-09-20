# Todo

## Implement Option Parsing for Historical Prompts

- [x] Create a function to parse `--v` and `--ar` options from a prompt string
- [x] Modify the historical prompt loading function to use the parsing function
- [x] Update the current prompt state with the cleaned prompt (without options)
- [x] Set the aspect ratio option based on the parsed `--ar` value
- [x] Set the model version option based on the parsed `--v` value
- [x] Update the UI to reflect the new aspect ratio and model version selections
- [x] Add error handling for cases where options are not found or are invalid

## Implement Toggle Functionality for Prompt Fragments

- [x] Modify the click event handler for prompt fragments
- [x] Create a function to check if a fragment is present in the current prompt
- [x] Implement logic to append the fragment if not present
- [x] Implement logic to remove the fragment (and associated comma) if already present
- [x] Update the current prompt state after toggling
- [x] Ensure proper comma handling (e.g., don't leave trailing commas, handle spaces correctly)
- [x] Update the UI to reflect changes in the current prompt
- [x] Add visual feedback to indicate which fragments are currently in use

## Future Enhancements

- [ ] Implement drag-and-drop functionality for reordering prompt fragments
- [ ] Add support for fragment categories or tags
- [ ] Implement undo/redo functionality for prompt changes
- [ ] Add a feature to save and load custom fragment sets
