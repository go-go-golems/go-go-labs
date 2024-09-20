# Midjourney Prompt Remix Tool Specification

## Table of Contents

1. [Overview](#overview)
2. [Purpose](#purpose)
3. [Scope](#scope)
4. [User Interface](#user-interface)
    - [Layout](#layout)
    - [Components](#components)
5. [Functional Requirements](#functional-requirements)
    - [Prompt Management](#prompt-management)
    - [Prompt Fragment Management](#prompt-fragment-management)
    - [Image URL Management](#image-url-management)
    - [Options Management](#options-management)
    - [Prompt History and Search](#prompt-history-and-search)
    - [Randomization Feature](#randomization-feature)
    - [Copy to Clipboard with Options](#copy-to-clipboard-with-options)
    - [Import/Export Feature](#import/export-feature)
    - [Image Thumbnails](#image-thumbnails)
6. [Non-Functional Requirements](#non-functional-requirements)
    - [Performance](#performance)
    - [Usability](#usability)
    - [Reliability](#reliability)
    - [Maintainability](#maintainability)
7. [Data Model](#data-model)
8. [State Management](#state-management)
9. [Persistence](#persistence)
10. [User Interactions](#user-interactions)
11. [Error Handling](#error-handling)
12. [Security Considerations](#security-considerations)
13. [Future Enhancements](#future-enhancements)
14. [Appendix](#appendix)

---

## Overview

The **Midjourney Prompt Remix Tool** is a web-based application designed to assist users in creating, managing, and refining prompts for the Midjourney AI image generation system. The tool provides a user-friendly interface for composing prompts by remixing various text fragments and images, offering customizable options, and maintaining a history of generated prompts.

---

## Purpose

The primary purpose of this application is to streamline the process of generating complex and effective prompts for Midjourney by allowing users to:

- **Reuse and manage prompt fragments**
- **Incorporate images into prompts**
- **Customize options such as aspect ratio and model version**
- **Maintain and search through a history of previously created prompts**
- **Randomize prompt fragment selection for creative variations**

---

## Scope

This specification covers the functionality, user interface, data management, and technical requirements of the Midjourney Prompt Remix Tool. It is intended for developers tasked with building, maintaining, or enhancing the application.

---

## User Interface

### Layout

The application is structured as a single-page interface divided into three main columns and a bottom section:

- **Left Column:** Prompt Fragments
- **Center Column:** Current Prompt and Images
- **Right Column:** Options
- **Bottom Section:** Prompt History

### Components

#### 1. Left Column: Prompt Fragments

- **Title:** "Prompt Fragments"
- **Fragment List:** Displays a list of reusable prompt fragments, each accompanied by a checkbox and a delete button.
- **Buttons:**
    - **Add New Fragment:** Opens a prompt dialog to add a new fragment.
    - **Randomize:** Selects a random subset of fragments and adds them to the current prompt.

#### 2. Center Column: Current Prompt and Images

- **Title:** "Current Prompt"
- **Textarea:** Editable area where users can manually edit the current prompt.
- **Buttons:**
    - **Copy to Clipboard:** Copies the current prompt (including options) to the clipboard.
    - **Add Image URL:** Opens a modal to input a new image URL.
- **Images List:** Displays a grid of thumbnails for added image URLs, each with an "Add to Prompt" and a delete button.

#### 3. Right Column: Options

- **Title:** "Options"
- **Aspect Ratio:**
    - **Radio Buttons:** Options for selecting aspect ratios (`1:1`, `16:9`, `4:3`).
- **Model Version:**
    - **Dropdown Menu:** Options for selecting model versions (`v5`, `v4`, `v3`).
- **New Functionality (Below Options):**
    - **Seed Input:** Allow users to specify a seed value for reproducible results.
    - **Style Presets:** Provide a dropdown of predefined style options (e.g., "photorealistic", "anime", "oil painting").
    - **Negative Prompts:** Add a textarea for users to input things they want to exclude from the generated image.
    - **Image Count:** Allow users to specify how many images they want to generate.
    - **Custom Parameters:** Provide input fields for advanced users to add custom Midjourney parameters.

#### 4. Bottom Section: Prompt History

- **Title:** "Prompt History"
- **Search Input:** Allows users to filter history based on search queries.
- **History List:** Displays a list of previously created prompts, each clickable to load into the current prompt.

#### 5. Modal: Add Image URL

- **Title:** "Add Image URL"
- **Input Field:** Text input for pasting the image URL.
- **Buttons:**
    - **Add:** Adds the entered URL to the images list.
    - **Cancel:** Closes the modal without adding.

#### 6. Confirmation Message

- **Location:** Fixed at the bottom center of the screen.
- **Purpose:** Displays temporary confirmation messages (e.g., "Prompt copied to clipboard!").

#### 7. Import/Export Buttons

- **Location:** Top navigation or settings area
- **Buttons:**
    - **Export:** Triggers the download of the current application state as a JSON file.
    - **Import:** Opens a file dialog to select and import a JSON state file.

---

## Functional Requirements

### Prompt Management

#### Editable Current Prompt

- **Description:** Users can manually edit the current prompt directly within the textarea.
- **Behavior:** 
    - Any manual changes are saved and reflected in the prompt history upon copying or adding fragments/images.

### Copy to Clipboard with Options

- **Description:** When users click the "Copy to Clipboard" button, the current prompt is appended with the selected aspect ratio (`--ar XX:XX`) and model version (`--v X`) before being copied.
- **Behavior:**
    - The appended options are **not** visible in the textarea.
    - The textarea remains clean, allowing for manual edits without automatic option additions.

### Prompt Fragment Management

#### Adding New Fragments

- **Description:** Users can add new prompt fragments via the "Add New Fragment" button.
- **Behavior:**
    - Opens a prompt dialog to input the new fragment.
    - The new fragment is added to the fragments list with a checkbox and the fragment text.

#### Deleting Fragments

- **Description:** Each prompt fragment has a delete button to remove it from the list.
- **Behavior:**
    - Upon deletion, the fragment is removed from the current prompt if present.

#### Using Fragments

- **Description:** Users can add fragments to the current prompt by clicking on the fragment text.
- **Behavior:**
    - Clicked fragments are appended to the current prompt, separated by commas.

#### Selecting Fragments for Randomization

- **Description:** Users can select multiple fragments using checkboxes to mark them for potential randomization.
- **Behavior:**
    - Checking a fragment's checkbox marks it as available for random selection.

### Randomization Feature

#### Randomize Button

- **Description:** A "Randomize" button allows users to select a random subset of the checked prompt fragments and add them to the current prompt.
- **Behavior:**
    - Only fragments with checked checkboxes are considered for randomization.
    - The number of fragments selected can vary (e.g., between 1 and all checked fragments).
    - Selected fragments are appended to the current prompt, separated by commas.

### Image URL Management

#### Adding Image URLs

- **Description:** Users can add new image URLs via the "Add Image URL" button, which opens a modal.
- **Behavior:**
    - Users input the URL in the modal and confirm to add it to the images list.

#### Deleting Image URLs

- **Description:** Each image URL has a delete button to remove it from the list.
- **Behavior:**
    - Upon deletion, the image URL is removed from the current prompt if present.

#### Using Image URLs

- **Description:** Users can add an image URL to the current prompt by clicking on it or using the "Add to Prompt" button.
- **Behavior:**
    - The image URL is prepended to the current prompt, separated by a comma.

### Options Management

#### Aspect Ratio

- **Description:** Users can select an aspect ratio using radio buttons.
- **Behavior:**
    - The selected ratio is appended as `--ar XX:XX` when copying the prompt.
    - Changes to the aspect ratio do not alter the current prompt directly.

#### Model Version

- **Description:** Users can select a model version from a dropdown menu.
- **Behavior:**
    - The selected version is appended as `--v X` when copying the prompt.
    - Changes to the model version do not alter the current prompt directly.

#### New Functionality (Below Options)

- **Description:** Additional options or features can be added below the existing options in the options column.
- **Potential Features:**
    - **Seed Input:** Allow users to specify a seed value for reproducible results.
    - **Style Presets:** Provide a dropdown of predefined style options (e.g., "photorealistic", "anime", "oil painting").
    - **Negative Prompts:** Add a textarea for users to input things they want to exclude from the generated image.
    - **Image Count:** Allow users to specify how many images they want to generate.
    - **Custom Parameters:** Provide input fields for advanced users to add custom Midjourney parameters.

### Prompt History and Search

#### Automatic History

- **Description:** Every unique prompt is automatically added to the history upon creation or modification.
- **Behavior:**
    - Duplicates are avoided if the prompt is identical to the most recent entry.

#### Searching History

- **Description:** Users can search through the prompt history using the search input.
- **Behavior:**
    - The history list filters in real-time based on the search query.
    - Matching terms within prompts are highlighted for visibility.

#### Using Historical Prompts

- **Description:** Clicking on a historical prompt loads it into the current prompt and updates the options accordingly.
- **Behavior:**
    - The aspect ratio and model version from the historical prompt are parsed and set as current selections.

### Prompt Fragments Separated by Commas

- **Description:** When prompt fragments are added to the current prompt, they are separated by commas for clarity and structure.
- **Behavior:**
    - Ensures that the prompt maintains a readable and organized format.

### Import/Export Feature

#### Exporting State

- **Description:** Users can export their entire application state (including prompt fragments, images, options, and history) as a JSON file.
- **Behavior:**
    - An "Export" button is added to the UI, likely in the top navigation or settings area.
    - Clicking the button triggers a download of a JSON file containing the current application state.

#### Importing State

- **Description:** Users can import a previously exported JSON file to restore their application state.
- **Behavior:**
    - An "Import" button is added next to the "Export" button.
    - Clicking the button opens a file dialog to select a JSON file.
    - The imported state replaces the current application state, updating all UI components.

### Image Thumbnails

#### Displaying Thumbnails

- **Description:** Instead of showing plain URLs, the application now displays thumbnail previews for added image URLs.
- **Behavior:**
    - When an image URL is added, the application attempts to load and display a thumbnail.
    - Thumbnails are shown in the Images List in the Center Column.
    - Each thumbnail is accompanied by the "Add to Prompt" and delete buttons.

#### Fallback for Invalid Images

- **Description:** If an image fails to load or the URL is invalid, a placeholder image or icon is displayed.
- **Behavior:**
    - The application attempts to load the image asynchronously.
    - If loading fails, a default "broken image" icon is shown instead.

---

## Non-Functional Requirements

### Performance

- **Responsiveness:** The application should respond to user interactions without noticeable delays.
- **Efficiency:** State updates and UI re-rendering should be optimized for smooth performance, even with a large number of prompt fragments or history entries.

### Usability

- **Intuitive Interface:** The layout and controls should be straightforward, minimizing the learning curve for new users.
- **Accessibility:** Ensure that the application is accessible to users with disabilities (e.g., keyboard navigation, screen reader compatibility).

### Reliability

- **Data Integrity:** Ensure that all state changes are accurately reflected in `localStorage` to prevent data loss.
- **Error Handling:** Gracefully handle errors, such as invalid image URLs or issues with clipboard access.

### Maintainability

- **Code Organization:** Structure the codebase in a modular and readable manner to facilitate future updates.
- **Documentation:** Provide clear comments and documentation within the code to aid understanding and maintenance.

---

## Data Model

### State Structure

The application state is managed as a JavaScript object and persisted in `localStorage`. The state structure is as follows:

```javascript
{
    prompt_fragments: [ "fragment1", "fragment2", ... ],
    images: [
        { url: "https://image1.url", thumbnail: "data:image/jpeg;base64,...", alt: "Description" },
        { url: "https://image2.url", thumbnail: "data:image/jpeg;base64,...", alt: "Description" },
        ...
    ],
    options: {
        aspect_ratio: "16:9", // Possible values: "1:1", "16:9", "4:3"
        model_version: "v5"    // Possible values: "v5", "v4", "v3"
    },
    current_prompt: "user's current prompt text",
    prompt_history: [ "history_prompt1", "history_prompt2", ... ],
    search_query: "" // Current search term for filtering history
}
```

### Components

- **prompt_fragments:** An array of strings representing reusable prompt components.
- **images:** An array of objects containing image URLs, thumbnails, and alt text.
- **options:** An object holding the current selections for aspect ratio and model version.
- **current_prompt:** A string representing the prompt being edited or composed.
- **prompt_history:** An array of strings representing previously created prompts.
- **search_query:** A string representing the current search term for filtering the prompt history.

---

## State Management

- **Initialization:** On application load, the state is retrieved from `localStorage`. If no previous state exists, the application initializes with default values.
- **Updates:** Any user interaction that modifies the state triggers an update to the state object, followed by a UI re-render and saving the updated state to `localStorage`.
- **Synchronization:** Ensure that all state changes are consistently reflected across all relevant UI components.

---

## Persistence

- **LocalStorage Usage:** The entire application state is stored in the browser's `localStorage` under the key `midjourneyPromptState`.
- **Data Retrieval:** Upon loading the application, the state is fetched from `localStorage`. If absent, default initial values are used.
- **Data Saving:** After every state change, the updated state object is serialized and saved back to `localStorage` to ensure persistence across sessions.

---

## User Interactions

### Adding a New Prompt Fragment

1. **Action:** Click the "Add New Fragment" button.
2. **Behavior:** A prompt dialog appears asking for the new fragment text.
3. **Result:** Upon confirmation, the new fragment is added to the fragments list with a checkbox and delete button.

### Selecting and Adding Prompt Fragments

1. **Action:** Check the boxes next to desired prompt fragments.
2. **Behavior:** Selected fragments are identified for addition.
3. **Result:** Fragments are appended to the current prompt, separated by commas.

### Randomizing Prompt Fragments

1. **Action:** Click the "Randomize" button.
2. **Behavior:** A random subset of available prompt fragments is selected.
3. **Result:** Selected fragments are appended to the current prompt, separated by commas.

### Editing the Current Prompt

1. **Action:** Manually type or modify text within the current prompt textarea.
2. **Behavior:** Changes are immediately reflected in the state.
3. **Result:** The prompt history is updated upon copying or adding fragments/images.

### Adding an Image URL

1. **Action:** Click the "Add Image URL" button.
2. **Behavior:** A modal appears with an input field for the image URL.
3. **Result:** Upon confirmation, the image URL is added to the images list, and a thumbnail is generated and displayed.

### Deleting a Prompt Fragment or Image URL

1. **Action:** Click the "Delete" button next to a fragment or image URL.
2. **Behavior:** The selected item is removed from its respective list.
3. **Result:** If the deleted item was part of the current prompt, it is also removed from there.

### Copying the Current Prompt

1. **Action:** Click the "Copy to Clipboard" button.
2. **Behavior:** The current prompt is appended with the selected aspect ratio and model version before being copied.
3. **Result:** A confirmation message appears, and the prompt is saved to history.

### Searching Prompt History

1. **Action:** Enter a search term in the search input field within the prompt history section.
2. **Behavior:** The history list filters in real-time based on the entered term.
3. **Result:** Only prompts containing the search term are displayed, with matching text highlighted.

### Using a Historical Prompt

1. **Action:** Click on a prompt within the filtered history list.
2. **Behavior:** The selected prompt loads into the current prompt textarea, and options are updated accordingly.
3. **Result:** The aspect ratio and model version are set based on the selected prompt's parameters.

### Exporting Application State

1. **Action:** Click the "Export" button.
2. **Behavior:** The current application state is serialized into a JSON file.
3. **Result:** A JSON file containing the application state is downloaded to the user's device.

### Importing Application State

1. **Action:** Click the "Import" button and select a JSON file.
2. **Behavior:** The selected file is parsed and validated.
3. **Result:** If valid, the imported state replaces the current application state, and the UI is updated accordingly.

---

## Future Enhancements

1. **Drag and Drop Functionality:**
    - Allow users to reorder prompt fragments and images via drag-and-drop interactions.
4. **Advanced Search:**
    - Implement advanced filtering options (e.g., by date, tags).
8. **Responsive Design Enhancements:**
    - Optimize the UI for mobile and tablet devices to enhance accessibility.