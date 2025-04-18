Below is a thin‑slice “hello‑world” style tutorial that drops a **“Chat with this”** button beside any `<textarea>` and, on click, slides in a Reachat‑powered chatbot that can read **and overwrite** the textarea’s value. The sample uses plain React (18+), Reachat ^1.3, and Tailwind for the few utility classes Reachat expects, but nothing prevents you from keeping your existing Bootstrap styles.

---

## 0  Prereqs

```bash
# Core UI lib + peer deps
npm i reachat react react-dom
# Tailwind (just for Reachat defaults – skip if you’ve already got it)
npm i -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

Add the Tailwind directives to `src/index.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

Reachat ships un‑scoped classes, so no further setup is required citeturn0search0.

---

## 1  Scaffold a “ChatWithTextarea” component

```tsx
// ChatWithTextarea.tsx
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

export default function ChatWithTextarea() {
  const [showChat, setShowChat] = useState(false);
  const [sessions, setSessions] = useState([]);
  const [activeId, setActiveId] = useState<string>();
  const textAreaRef = useRef<HTMLTextAreaElement>(null);

  /* -- handlers --------------------------------------------------------- */

  const handleSend = async (message: string) => {
    // ① send message + textarea content to backend LLM
    const body = {
      prompt: message,
      currentText: textAreaRef.current?.value ?? "",
    };
    const res = await fetch("/api/chat", {
      method: "POST",
      body: JSON.stringify(body),
    });
    const { revisedText, assistant } = await res.json();

    // ② store conversation locally so Reachat shows it
    const s = sessions.find((s) => s.id === activeId)!;
    const newConv = {
      id: `${s.id}-${s.conversations.length}`,
      question: message,
      response: assistant,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    const updated = { ...s, conversations: [...s.conversations, newConv] };
    setSessions([...sessions.filter((s) => s.id !== activeId), updated]);

    // ③ push the LLM’s revised text into the textarea
    if (textAreaRef.current) textAreaRef.current.value = revisedText;
  };

  /* -- ui ---------------------------------------------------------------- */

  return (
    <div className="relative">
      <textarea
        ref={textAreaRef}
        className="w-full h-40 p-2 border rounded"
        placeholder="Write something…"
      />

      <button
        onClick={() => setShowChat(true)}
        className="absolute right-2 bottom-2 bg-indigo-600 text-white px-3 py-1 rounded"
      >
        Chat&nbsp;with&nbsp;this
      </button>

      {showChat && (
        <div className="fixed inset-y-0 right-0 w-[400px] bg-white shadow-xl z-50">
          <Chat
            viewType="chat"
            sessions={sessions}
            activeSessionId={activeId}
            onNewSession={() => {
              const id = (sessions.length + 1).toString();
              setSessions([
                ...sessions,
                {
                  id,
                  title: `Session ${id}`,
                  createdAt: new Date(),
                  updatedAt: new Date(),
                  conversations: [],
                },
              ]);
              setActiveId(id);
            }}
            onSelectSession={setActiveId}
            onSendMessage={handleSend}
          >
            <SessionsList>
              <NewSessionButton />
              <SessionGroups />
            </SessionsList>

            <SessionMessagePanel>
              <SessionMessages />
              <ChatInput />
            </SessionMessagePanel>
          </Chat>
        </div>
      )}
    </div>
  );
}
```

Key points:

- **`viewType="chat"`** renders _just_ the chat column—perfect for a slide‑in drawer citeturn1view0.
- The `handleSend` callback bundles the user prompt **plus the current textarea value** so your backend can return a revised version. When the backend answers, we drop that new text straight into the textarea and log the conversation for Reachat to display citeturn2view0.

---

## 2  Minimal backend stub (Next.js / Express—take your pick)

```ts
// /api/chat.ts
import OpenAI from "openai";
const openai = new OpenAI();

export default async function handler(req, res) {
  const { prompt, currentText } = JSON.parse(req.body);
  const completion = await openai.chat.completions.create({
    model: "gpt-4o-mini",
    messages: [
      { role: "system", content: "You're a helpful writing assistant." },
      { role: "user", content: `Here is the current text:\n\n${currentText}` },
      { role: "user", content: `Please apply this edit request: ${prompt}` },
    ],
  });

  const revisedText = completion.choices[0].message.content.trim();
  res.json({ revisedText, assistant: revisedText });
}
```

The backend flattens to two items:

1. **`revisedText`** – what we’ll stuff back into the `<textarea>`.
2. **`assistant`** – echoed into the chat transcript (often identical, but you could make it more verbose).

---

## 3  Mount it anywhere

```tsx
import ChatWithTextarea from "./ChatWithTextarea";
import "reachat/styles.css"; // glob‑import Reachat defaults
import "./index.css"; // Tailwind

function App() {
  return (
    <main className="max-w-3xl mx-auto py-10">
      <h1 className="text-2xl font-bold mb-4">Blog post draft</h1>
      <ChatWithTextarea />
    </main>
  );
}
```

Run `npm run dev`, type into the box, click **Chat with this**, and iterate on your text through the side‑panel chatbot.

---

### Where to go from here

- **Styling** – Reachat ships a full theming object you can merge with Tailwind classes for perfect visual parity with Bootstrap or your design system citeturn1view0.
- **Multi‑text binding** – pass a textarea identifier into the chat so the backend knows which one you’re editing if you have several on the page.
- **Streaming** – wire `onStopMessage` and stream partial responses for a live‑typing effect.

That’s the entire plumbing required to bolt a Reachat chatbot onto any free‑form text input in your app—no hand‑rolled chat widgets, just the bits you care about.
