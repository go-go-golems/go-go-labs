Your application is a web-based tool for creating and managing prompts for the Midjourney AI image generation system. It allows users to compose prompts by combining text fragments, manage image URLs, set options like aspect ratio and model version, and maintain a history of generated prompts. The application has been refactored to use Redux Toolkit for state management, which provides a more structured and efficient way to handle the application's state.

Here's an overview of the main components and their functionalities:

1. Store Configuration:
The Redux store is configured in the `store.js` file. It combines reducers from different slices and sets up middleware for state persistence.

2. State Slices:
The application's state is divided into several slices, each managing a specific part of the state:

- `promptFragmentsSlice.js`: Manages prompt fragments, checked fragments, and saved selections.
- `imagesSlice.js`: Handles the list of image URLs.
- `optionsSlice.js`: Manages options like aspect ratio and model version.
- `promptHistorySlice.js`: Handles the prompt history and current prompt.
- `uiSlice.js`: Reserved for UI-related state (currently empty).

3. Components:
The application is structured into several components, each responsible for a specific part of the UI:

- `FragmentsColumn`: Manages the list of prompt fragments and their interactions.
- `PromptColumn`: Handles the current prompt and image URL management.
- `OptionsColumn`: Manages aspect ratio and model version options.
- `HistorySection`: Displays and manages the prompt history.

4. Main Application Logic:
The `App` class in `app.js` serves as the main controller, initializing components and handling global actions like importing/exporting state.

5. Utility Functions:
The `utils.js` file contains helper functions for text manipulation, UI feedback, and prompt parsing.

6. Persistence:
State persistence is achieved using a custom Redux middleware that saves the state to localStorage after each action.


7. UI Rendering:
The application uses Lit HTML for efficient UI rendering, as seen in the component files.

8. Recent Refactoring:
The application has been refactored to use Redux Toolkit, which simplifies state management by:
- Using `createSlice` to define reducers and action creators in one place.
- Automatically generating action creators based on reducer names.
- Allowing direct state mutations in reducers (thanks to Immer).

9. Planned Enhancements:
There are several planned enhancements, as listed in the `tasks.md` file, including drag-and-drop functionality, fragment categories, and undo/redo functionality.
