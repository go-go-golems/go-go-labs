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
      const res = await fetch("http://localhost:3001/api/chat", { // Assuming server runs on 3001
        method: "POST",
        headers: {
          "Content-Type": "application/json", // Add content type header
        },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        throw new Error(`HTTP error! status: ${res.status}`);
      }

      const { revisedText, assistant } = await res.json();

      // ② store conversation locally so Reachat shows it
      const activeSession = sessions.find((s) => s.id === activeId);
      if (!activeSession) return; // Guard against undefined session

      const newConv: Conversation = {
        id: `${activeSession.id}-${activeSession.conversations.length}`,
        question: message,
        response: assistant,
        createdAt: new Date(),
        updatedAt: new Date(),
      };
      const updatedSession = {
        ...activeSession,
        conversations: [...activeSession.conversations, newConv],
        updatedAt: new Date(), // Update session timestamp
      };
      setSessions([
        ...sessions.filter((s) => s.id !== activeId),
        updatedSession,
      ]);

      // ③ push the LLM's revised text into the textarea
      if (textAreaRef.current) textAreaRef.current.value = revisedText;
    } catch (error) { // Add error handling
      console.error("Failed to send message:", error);
      // Optionally, update UI to show error
    }
  };

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
    setActiveId(id);
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
        className="absolute right-2 bottom-2 bg-indigo-600 text-white px-3 py-1 rounded hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
      >
        Chat&nbsp;with&nbsp;this
      </button>

      {showChat && (
        <div className="fixed inset-y-0 right-0 w-[400px] bg-white shadow-xl z-50 flex flex-col"> {/* Added flex container */}
          <div className="flex justify-end p-2"> {/* Container for close button */}
            <button
              onClick={() => setShowChat(false)}
              className="text-gray-500 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div className="flex-grow overflow-hidden"> {/* Added flex-grow and overflow-hidden */}
            <Chat
              viewType="chat"
              sessions={sessions}
              activeSessionId={activeId}
              onNewSession={handleNewSession}
              onSelectSession={setActiveId}
              onSendMessage={handleSend}
              // Add onDeleteSession if needed
              // Add onUpdateSession if needed
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
        </div>
      )}
    </div>
  );
} 