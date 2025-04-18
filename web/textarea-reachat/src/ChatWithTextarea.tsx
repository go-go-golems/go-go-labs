import React, { useState, useRef, useEffect } from "react";
import {
  Chat,
  SessionMessages,
  SessionMessagePanel,
  ChatInput,
  SessionMessage,
} from "reachat";

// Define the Session type based on usage
interface Conversation {
  id: string;
  question: string;
  response?: string;
  createdAt: Date;
  updatedAt?: Date;
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
  const [sessions, setSessions] = useState<Session[]>([
    {
      id: 'session1',
      title: 'Editing Session',
      createdAt: new Date(),
      updatedAt: new Date(),
      conversations: []
    }
  ]);
  const [activeId, setActiveId] = useState<string>('session1');
  const [loading, setLoading] = useState(false);
  const textAreaRef = useRef<HTMLTextAreaElement>(null);

  /* -- Effects ---------------------------------------------------------- */

  useEffect(() => {
    if (showChat) {
      const currentText = textAreaRef.current?.value ?? '';
      setSessions(prev => prev.map(session => {
        if (session.id === activeId) {
          const initialMessage: Conversation = {
            id: `${session.id}-msg-0`,
            question: "Here is my text:\n\n" + currentText,
            response: undefined,
            createdAt: new Date(),
            updatedAt: new Date(),
          };
          return { ...session, conversations: [ initialMessage ] };
        }
        return session;
      }));
    }
  }, [showChat, activeId]);

  /* -- handlers --------------------------------------------------------- */

  const handleSend = async (message: string) => {
    const sessionId = activeId;
    if (!sessionId) {
      console.error("No active session ID found");
      return;
    }

    const activeSession = sessions.find((s) => s.id === sessionId);
    if (!activeSession) return;

    const newConv: Conversation = {
      id: `${sessionId}-msg-${activeSession.conversations.length}`,
      question: message,
      response: undefined,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    setSessions(prev => prev.map(s =>
      s.id === sessionId
        ? { ...s, conversations: [...s.conversations, newConv], updatedAt: new Date() }
        : s
    ));

    setLoading(true);
    const currentTextForAPI = textAreaRef.current?.value ?? '';

    try {
      const res = await fetch("http://localhost:3001/api/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: message, currentText: currentTextForAPI })
      });

      if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
      const { revisedText, assistant } = await res.json();

      setSessions(prev => prev.map(s => {
        if (s.id !== sessionId) return s;
        return {
          ...s,
          conversations: s.conversations.map(c =>
            c.id === newConv.id ? { ...c, response: assistant, updatedAt: new Date() } : c
          ),
          updatedAt: new Date()
        };
      }));

      if (textAreaRef.current && revisedText !== undefined) {
        textAreaRef.current.value = revisedText;
      }

    } catch (err: any) {
      console.error("Error fetching AI response", err);
      setSessions(prev => prev.map(s => {
        if (s.id !== sessionId) return s;
        return {
          ...s,
          conversations: s.conversations.map(c =>
            c.id === newConv.id ? { ...c, response: `Error: ${err.message}`, updatedAt: new Date() } : c
          ),
          updatedAt: new Date()
        };
      }));
    } finally {
      setLoading(false);
    }
  };

  /* -- ui ---------------------------------------------------------------- */

  return (
    <div className="relative p-4">
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
        <div
          className="chat-widget-panel fixed inset-y-0 right-0 w-[400px] h-full bg-white shadow-xl z-50 flex flex-col border-l border-gray-200"
        >
          <div className="flex justify-end p-2 border-b border-gray-200">
            <button
              onClick={() => setShowChat(false)}
              className="text-gray-500 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500"
              aria-label="Close chat panel"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div className="flex-grow overflow-hidden">
            <Chat
              viewType="chat"
              sessions={sessions}
              activeSessionId={activeId}
              isLoading={loading}
              onSendMessage={handleSend}
            >
              <SessionMessagePanel allowBack={false}>
                <SessionMessages>
                  {(conversations: Conversation[]) =>
                    conversations.map(conv => (
                      <SessionMessage key={conv.id} conversation={conv} />
                    ))
                  }
                </SessionMessages>
                <div className="border-t p-2">
                  <ChatInput 
                    placeholder="Ask the chatbot..."
                  />
                </div>
              </SessionMessagePanel>
            </Chat>
          </div>
        </div>
      )}
    </div>
  );
} 