# Textarea + Reachat Chatbot Integration Tutorial

Below is a thin-slice "hello-world" style tutorial that drops a **"Chat with this"** button beside any `<textarea>` and, on click, slides in a Reachat-powered chatbot that can read **and overwrite** the textarea's value. The sample uses React (19+), Vite, Reachat ^2.0, Tailwind CSS v3, and a simple Express backend.

## Goal

To demonstrate integrating a Reachat component with a standard HTML textarea, allowing an AI assistant (powered by OpenAI via a backend API) to interact with and modify the textarea's content based on user prompts within the chat interface.

## Developer Quick Start

1.  **Clone/Setup:** Ensure you have this project directory (`web/textarea-reachat`).
2.  **Backend API Key:** Create a `.env` file in `web/textarea-reachat/server/` with your OpenAI API key:
    ```.env
    OPENAI_API_KEY=your_openai_api_key_here
    # PORT=3001 # Optional: Define if you need a port other than 3001
    ```
3.  **Install Deps:**
    ```bash
    # From workspace root
    cd web/textarea-reachat && npm install
    cd server && npm install
    cd ..
    ```
4.  **Run Dev Servers (in separate terminals):**
    ```bash
    # Terminal 1: Start Backend (from workspace root)
    cd web/textarea-reachat/server
    npm run dev
    ```
    ```bash
    # Terminal 2: Start Frontend (from workspace root)
    cd web/textarea-reachat
    npm run dev
    ```
5.  **Open App:** Navigate to the URL provided by Vite (e.g., `http://localhost:5173`).

---

## 0 Prerequisites & Setup

This project uses Node.js, npm, Vite for the frontend, and Express for the backend.

### 0.1 Project Structure

```
web/textarea-reachat/
├── node_modules/
├── public/
├── server/
│   ├── node_modules/
│   ├── .env          # <-- Add your OPENAI_API_KEY here
│   ├── package.json
│   ├── server.ts
│   └── tsconfig.json
├── src/
│   ├── App.tsx
│   ├── ChatWithTextarea.tsx
│   ├── index.css
│   └── main.tsx
├── .gitignore
├── index.html
├── package.json
├── postcss.config.js
├── tailwind.config.js
└── vite.config.ts
```

### 0.2 Frontend Dependencies

From the `web/textarea-reachat` directory:

```bash
# Core UI lib + peer deps
npm i react react-dom reachat

# Dev Dependencies (Vite, TS, Tailwind v3)
npm i -D typescript @types/react @types/react-dom vite @vitejs/plugin-react tailwindcss@^3.0 postcss autoprefixer
```

### 0.3 Backend Dependencies

From the `web/textarea-reachat/server` directory:

```bash
# Server runtime deps
npm i express openai dotenv cors

# Dev Dependencies (TS support)
npm i -D typescript @types/node @types/express @types/cors ts-node-dev
```

### 0.4 Tailwind CSS v3 Configuration

1.  **Initialize Tailwind & PostCSS:** Run this in `web/textarea-reachat`:

    ```bash
    npx tailwindcss init -p
    ```

    This creates `tailwind.config.js` and `postcss.config.js`.

2.  **Install PostCSS Import:** We need this to properly handle the CSS imports:

    ```bash
    npm i -D postcss-import
    ```

3.  **Configure Template Paths** (`web/textarea-reachat/tailwind.config.js`):

    ```javascript
    /** @type {import('tailwindcss').Config} */
    module.exports = {
      content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}", // Scan relevant files
      ],
      theme: {
        extend: {},
      },
      plugins: [],
    };
    ```

4.  **Update PostCSS Config** (`web/textarea-reachat/postcss.config.js`):

    ```javascript
    module.exports = {
      plugins: {
        "postcss-import": {}, // Must come first to process @import statements
        tailwindcss: {},
        autoprefixer: {},
      },
    };
    ```

5.  **Add Tailwind Directives and Import Reachat CSS** (`web/textarea-reachat/src/index.css`):

    ```css
    @tailwind base;
    @import "reachat/index.css";
    @tailwind components;
    @tailwind utilities;
    ```

    > **Note**: This ordering is crucial. We import Reachat's CSS between Tailwind directives so that its `@layer` directives work correctly with Tailwind's base styles.

---

## 1 Scaffold the "ChatWithTextarea" Component

This component manages the textarea, the chat button, the chat panel's visibility, and communication with the backend.

```tsx
// web/textarea-reachat/src/ChatWithTextarea.tsx
import { useState, useRef } from "react";
import {
  Chat,
  SessionMessages,
  SessionMessagePanel,
  ChatInput,
  SessionsList,
  NewSessionButton,
  SessionGroups,
} from "reachat";

// Define the Session type based on usage
interface Conversation {
  id: string;
  question: string;
  response: string;
  createdAt: Date;
  updatedAt: Date;
}

interface Session {
  id: string;
  title: string;
  createdAt: Date;
  updatedAt: Date;
  conversations: Conversation[];
}

export default function ChatWithTextarea() {
  const [showChat, setShowChat] = useState(false);
  const [sessions, setSessions] = useState<Session[]>([]); // Use the Session type
  const [activeId, setActiveId] = useState<string | undefined>(undefined); // Match type
  const textAreaRef = useRef<HTMLTextAreaElement>(null);

  /* -- handlers --------------------------------------------------------- */

  const handleSend = async (message: string) => {
    // ① send message + textarea content to backend LLM
    const body = {
      prompt: message,
      currentText: textAreaRef.current?.value ?? "",
    };
    try {
      // Connect to the backend server (running on port 3001)
      const res = await fetch("http://localhost:3001/api/chat", {
        method: "POST",
        headers: {
          "Content-Type": "application/json", // Essential header
        },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        // Basic error handling for network issues
        throw new Error(`HTTP error! status: ${res.status}`);
      }

      const { revisedText, assistant } = await res.json();

      // ② store conversation locally so Reachat shows it
      const activeSession = sessions.find((s) => s.id === activeId);
      if (!activeSession) return; // Should not happen if UI logic is correct

      const newConv: Conversation = {
        id: `${activeSession.id}-${activeSession.conversations.length}`,
        question: message,
        response: assistant, // The AI's response for the chat log
        createdAt: new Date(),
        updatedAt: new Date(),
      };
      const updatedSession = {
        ...activeSession,
        conversations: [...activeSession.conversations, newConv],
        updatedAt: new Date(), // Update session timestamp
      };
      // Update the sessions state, replacing the old session with the updated one
      setSessions([
        ...sessions.filter((s) => s.id !== activeId),
        updatedSession,
      ]);

      // ③ push the LLM's revised text directly into the textarea
      if (textAreaRef.current) textAreaRef.current.value = revisedText;
    } catch (error) {
      // Catch fetch errors or issues during processing
      console.error("Failed to send message:", error);
      // TODO: Implement user-facing error display
    }
  };

  // Handler to create a new chat session
  const handleNewSession = () => {
    const id = (sessions.length + 1).toString();
    const newSession: Session = {
      id,
      title: `Session ${id}`,
      createdAt: new Date(),
      updatedAt: new Date(),
      conversations: [],
    };
    setSessions([...sessions, newSession]);
    setActiveId(id); // Activate the new session
  };

  /* -- ui ---------------------------------------------------------------- */

  return (
    <div className="relative">
      {" "}
      {/* Container for positioning */}
      <textarea
        ref={textAreaRef}
        className="w-full h-40 p-2 border rounded"
        placeholder="Write something…"
      />
      {/* Button positioned absolutely within the relative parent */}
      <button
        onClick={() => setShowChat(true)}
        className="absolute right-2 bottom-2 bg-indigo-600 text-white px-3 py-1 rounded hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
      >
        Chat&nbsp;with&nbsp;this
      </button>
      {/* Conditional rendering of the chat panel */}
      {showChat && (
        <div className="fixed inset-y-0 right-0 w-[400px] bg-white shadow-xl z-50 flex flex-col">
          {" "}
          {/* Panel styling and layout */}
          {/* Close Button Area */}
          <div className="flex justify-end p-2">
            <button
              onClick={() => setShowChat(false)}
              className="text-gray-500 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500"
              aria-label="Close chat panel"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
          {/* Reachat Component Area */}
          <div className="flex-grow overflow-hidden">
            {" "}
            {/* Allows chat to scroll */}
            <Chat
              viewType="chat" // Renders only the chat column
              sessions={sessions} // Pass local session state
              activeSessionId={activeId} // Control the active session
              onNewSession={handleNewSession} // Callback for new session button
              onSelectSession={setActiveId} // Callback when user selects a session
              onSendMessage={handleSend} // Callback when user sends a message
              // TODO: Implement onDeleteSession and onUpdateSession if needed
            >
              <SessionsList>
                {" "}
                {/* Left panel for session list */}
                <NewSessionButton />
                <SessionGroups />
              </SessionsList>

              <SessionMessagePanel>
                {" "}
                {/* Right panel for messages */}
                <SessionMessages /> {/* Displays the conversation */}
                <ChatInput /> {/* Input field for the user */}
              </SessionMessagePanel>
            </Chat>
          </div>
        </div>
      )}
    </div>
  );
}
```

**Key points:**

- **State Management:** Uses `useState` for chat visibility (`showChat`), session data (`sessions`), and the currently active session ID (`activeId`).
- **Textarea Reference:** Uses `useRef` (`textAreaRef`) to directly access and update the textarea's value.
- **Type Safety:** Defines `Conversation` and `Session` interfaces for better code clarity and type checking.
- **`handleSend` Logic:** Constructs the request body, fetches data from the `http://localhost:3001/api/chat` backend endpoint, updates the local session state with the new conversation, and directly sets the textarea's value with `revisedText` from the response.
- **Error Handling:** Includes basic `try...catch` for the fetch operation.
- **`handleNewSession`:** Creates a new session object, adds it to the state, and sets it as active.
- **UI:**
  - A `relative` container holds the `textarea` and the absolutely positioned chat button.
  - The chat panel is conditionally rendered using `showChat` state.
  - The panel uses `fixed` positioning to overlay the page content.
  - A close button (`X`) is added to hide the panel.
  - Flexbox (`flex flex-col`) is used for the panel layout.
  - `viewType="chat"` renders Reachat in a compact, single-column mode suitable for side panels.
  - Callbacks (`onNewSession`, `onSelectSession`, `onSendMessage`) connect UI actions to state handlers.

---

## 2 Minimal Backend Server (Express + OpenAI)

This simple Express server listens for POST requests on `/api/chat`, interacts with the OpenAI API, and returns the modified text.

```typescript
// web/textarea-reachat/server/server.ts
import express, { Request, Response } from "express";
import OpenAI from "openai";
import dotenv from "dotenv";
import cors from "cors"; // Import cors

dotenv.config(); // Load .env file variables (OPENAI_API_KEY)

const app = express();
const port = process.env.PORT || 3001;

// Enable CORS for all origins (suitable for development)
// For production, configure specific origins: app.use(cors({ origin: 'YOUR_FRONTEND_URL' }));
app.use(cors());

// Middleware to parse JSON request bodies
app.use(express.json());

// Initialize OpenAI client
const openai = new OpenAI({
  apiKey: process.env.OPENAI_API_KEY, // Ensure key is in .env
});

if (!process.env.OPENAI_API_KEY) {
  console.warn("OPENAI_API_KEY not found in .env file. API calls will fail.");
}

// Define the chat endpoint
app.post("/api/chat", async (req: Request, res: Response) => {
  const { prompt, currentText } = req.body;

  // Basic validation
  if (!prompt || typeof currentText !== "string") {
    return res.status(400).json({ error: "Missing prompt or currentText" });
  }

  try {
    const completion = await openai.chat.completions.create({
      model: "gpt-4o-mini", // Or your preferred model
      messages: [
        {
          role: "system",
          // Instruct the AI on its role and expected output format
          content:
            "You are a helpful writing assistant. You will be given the current text in a textarea and an edit request. Your goal is to apply the edit request to the current text and return ONLY the full, revised text, without any extra explanations, commentary, or conversational filler. If the request is ambiguous or cannot be applied, ask for clarification.",
        },
        {
          role: "user",
          content: `Here is the current text:\n\n${currentText}`,
        },
        {
          role: "user",
          content: `Please apply this edit request: ${prompt}`,
        },
      ],
      temperature: 0.7, // Adjust creativity vs. predictability
    });

    const revisedText = completion.choices[0]?.message?.content?.trim();

    // Handle potential issues with the OpenAI response
    if (revisedText === undefined) {
      console.error("OpenAI response was empty or malformed:", completion);
      return res.status(500).json({ error: "Failed to get response from AI" });
    }

    // Return the revised text for the textarea and the same text as the 'assistant' response for the chat log
    res.json({ revisedText, assistant: revisedText });
  } catch (error) {
    console.error("Error calling OpenAI:", error);
    // Handle potential errors during the API call
    res.status(500).json({ error: "Failed to process chat request" });
  }
});

app.listen(port, () => {
  console.log(`Server listening on http://localhost:${port}`);
});
```

**Key points:**

- **Dependencies:** Uses `express`, `openai`, `dotenv`, and `cors`.
- **Configuration:** Loads the `OPENAI_API_KEY` from a `.env` file.
- **CORS:** Uses the `cors` middleware to allow requests from the frontend (running on a different port during development).
- **Endpoint:** Listens for `POST` requests at `/api/chat`.
- **OpenAI Call:** Constructs the messages array (including a system prompt defining the AI's task and the user's text/prompt) and calls `openai.chat.completions.create`.
- **Response:** Extracts the AI's response (`revisedText`) and sends it back in a JSON object `{ revisedText, assistant }`. Both fields contain the same modified text in this simple setup.
- **Error Handling:** Includes basic `try...catch` for the OpenAI API call and checks for empty responses.

---

## 3 Frontend Entry Point & Mounting

Standard Vite React setup.

### 3.1 Main Component (`App.tsx`)

```tsx
// web/textarea-reachat/src/App.tsx
import ChatWithTextarea from "./ChatWithTextarea";
import "./index.css"; // Import Tailwind and Reachat styles

function App() {
  return (
    <main className="max-w-3xl mx-auto py-10 px-4">
      {" "}
      {/* Basic page layout */}
      <h1 className="text-2xl font-bold mb-4">Blog post draft</h1>
      <ChatWithTextarea />
      {/* You could add more instances here if needed */}
    </main>
  );
}

export default App;
```

### 3.2 Root Rendering (`main.tsx`)

```tsx
// web/textarea-reachat/src/main.tsx
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
// Note: All styles are imported in App.tsx

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

### 3.3 HTML Host (`index.html`)

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Reachat Textarea Example</title>
  </head>
  <body>
    <div id="root"></div>
    {/* React app mounts here */}
    <script type="module" src="/src/main.tsx"></script>
    {/* Entry point */}
  </body>
</html>
```

### 3.4 Vite Configuration (`vite.config.ts`)

```typescript
// web/textarea-reachat/vite.config.ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()], // Basic React plugin
  // No specific Tailwind plugin needed for v3 with PostCSS
});
```

---

## 4 Running the Application

1.  **Ensure `.env**:\*\* Place your `OPENAI_API_KEY` in `web/textarea-reachat/server/.env`.
2.  **Start Backend Server:**
    ```bash
    cd web/textarea-reachat/server
    npm run dev
    # Should log: Server listening on http://localhost:3001
    ```
3.  **Start Frontend Dev Server (in a new terminal):**
    ```bash
    cd web/textarea-reachat
    npm run dev
    # Should log the local URL, e.g., http://localhost:5173
    ```
4.  **Open in Browser:** Navigate to the frontend URL (e.g., `http://localhost:5173`).

Type into the textarea, click **Chat with this**, and use the chat panel to send edit prompts to the AI.

### 4.1 Troubleshooting CSS Integration

If you encounter CSS errors related to Tailwind layers like:

```
[postcss] `@layer base` is used but no matching `@tailwind base` directive is present.
```

This happens because Reachat's CSS includes Tailwind layer directives that need to be processed in the same pass as your Tailwind directives. The solution uses these key techniques:

1. **Order matters**: The CSS import must appear _between_ Tailwind directives (after `@tailwind base`).
2. **Use postcss-import**: This plugin lets PostCSS inline the imported CSS before Tailwind processes it.
3. **Single import source**: Only import styles from the main CSS file, not directly in components.

Our setup in sections 0.4 and 3.1 follows these best practices, so the styles work correctly.

## 5 Where to Go From Here

This is a basic implementation. Potential enhancements include:

- **Styling/Theming:** Customize Reachat's appearance to match your application using its theming capabilities or by overriding Tailwind classes. Check the Reachat documentation for theming options.
- **Error Handling:** Display errors from the backend API call more gracefully to the user within the chat interface.
- **Loading Indicators:** Show a loading state in the chat while waiting for the backend response.
- **Streaming Responses:** Implement streaming from the OpenAI API (using `stream: true` and handling SSE on the frontend) for a more interactive, real-time typing effect in the chat response.
- **Multi-Textarea Support:** Modify `handleSend` to include an identifier for the specific textarea being edited if multiple `ChatWithTextarea` components are on the same page.
- **More Sophisticated Backend Logic:** Enhance the system prompt or add more complex logic on the server to handle different types of requests or maintain context across multiple turns.
- **Session Persistence:** Save/load chat sessions (e.g., to local storage or a database) so they persist across page reloads.
- **Production Build & Deployment:** Configure CORS properly for production, build the frontend (`npm run build`), and deploy the frontend static files and the backend server.
