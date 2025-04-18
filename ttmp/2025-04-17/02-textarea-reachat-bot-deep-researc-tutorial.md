Great — I’ll put together a tutorial showing how to integrate a ReaChat-powered chatbot into a React app. The chatbot will pop open from a button placed next to a textarea and support general conversation, with the added ability to edit or update the textarea content (e.g., by parsing the chatbot output or interpreting tool calls).

I’ll include code examples for:
- Rendering the textarea with a "Chat with this" button
- Opening the ReaChat widget on button click
- Sending and receiving messages from ReaChat
- Handling structured output (like inline tool calls or text replacements)

I’ll get this ready for you shortly.

# Integrating a ReaChat Chatbot into a React App (Step-by-Step Tutorial)

 ([reachat - Build Chat Experiences in Hours, Not Weeks.](https://reachat.dev/)) *Example of a ReaChat-powered chat widget in a React application. The following tutorial will guide you through building a similar interface.* 

1. **Set Up React and Install ReaChat**  
   Start by creating a new React project (using Create React App, Vite, Next.js, etc.) and installing ReaChat. ReaChat is provided as an NPM package ([Setup](https://reachat.dev/docs/getting-started/setup#:~:text=1)). You’ll also need Tailwind CSS for styles and ReaBlocks (the theming provider used by ReaChat). Run the following in your project directory:  
   
    ```bash
    npx create-react-app my-chatbot-app
    cd my-chatbot-app
    npm install reachat reablocks tailwindcss @tailwindcss/forms @tailwindcss/typography
    ```  
   This installs **ReaChat**, **ReaBlocks**, and Tailwind CSS (with some useful plugins for forms/typography). ReaChat’s UI is built with Tailwind, so including it is important for proper styling.

2. **Configure Tailwind CSS and ReaChat Theme**  
   Next, initialize Tailwind in your project and set up the ReaBlocks theme provider. Generate a Tailwind config (if using CRA or Vite, you can run `npx tailwindcss init`). Then enable Tailwind in your CSS and wrap your app with the theme provider:  

    - **Add Tailwind Directives:** In your main CSS file (e.g. `src/index.css` or `src/App.css`), import Tailwind’s base styles. For example:  
      ```css
      @tailwind base;
      @tailwind components;
      @tailwind utilities;
      ```  
      This includes Tailwind’s styles in your app.  
    - **Wrap with ThemeProvider:** In your root component (e.g. `src/App.js`), use ReaBlocks’ `ThemeProvider` to provide default theming to ReaChat components ([Setup](https://reachat.dev/docs/getting-started/setup#:~:text=import%20,from%20%27reablocks)). For example:  
      ```jsx
      // App.js
      import React from 'react';
      import { ThemeProvider, theme } from 'reablocks';  // import ReaBlocks theme
      import ChatWidget from './ChatWidget';            // (we'll create ChatWidget next)
      
      function App() {
        return (
          <ThemeProvider theme={theme}>
            <ChatWidget />
          </ThemeProvider>
        );
      }
      export default App;
      ```  
      Here, `ThemeProvider` is given a default `theme` from ReaBlocks. This ensures the ReaChat UI adopts the expected styling and theming.

3. **Build the Textarea and Button UI**  
   Create a component (e.g. `ChatWidget.jsx`) that will contain the text editor (textarea) and the chat widget. In this component, set up a textarea for the user’s content and a “Chat with this” button:  

    ```jsx
    // ChatWidget.jsx
    import React, { useState } from 'react';
    import { Chat } from 'reachat';
    // (We will add more imports from reachat as needed)
    
    export default function ChatWidget() {
      const [text, setText] = useState('');       // state for textarea content
      const [chatOpen, setChatOpen] = useState(false);  // whether chat widget is open
    
      return (
        <div className="chat-container" style={{ position: 'relative', padding: '1rem' }}>
          {/* Textarea for content input */}
          <textarea 
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Type or paste your text here..." 
            rows={6} cols={60}
            className="text-area"
          />
          {/* Button to open the chat widget */}
          <button onClick={() => setChatOpen(true)} className="open-chat-btn">
            Chat with this
          </button>
          
          {/* Chat widget will be conditionally rendered below */}
          { /* ...chat UI goes here... */ }
        </div>
      );
    }
    ```  

   In the code above, we maintain state for the textarea (`text`) and a boolean `chatOpen` to track if the chat popup is visible. The textarea allows the user to input or edit content, and the button toggles the chat interface. You can style the elements with Tailwind classes or custom CSS as needed (e.g., add some margin or styling to the button and textarea for a clean UI).

4. **Embed the ReaChat Widget**  
   Now, integrate the ReaChat chat interface into the component. Using ReaChat’s provided components, we’ll render a chat window when `chatOpen` is true. We’ll use the **`<Chat>`** component with a *“chat” view type* so that it displays a single chat session without the session sidebar ([Components](https://reachat.dev/docs/api/components#:~:text=,console)). Inside `<Chat>`, include the message list and input field components. For example:  

    ```jsx
    import { Chat, SessionMessagePanel, SessionMessages, SessionMessage, ChatInput } from 'reachat';
    // ... inside ChatWidget component's return:
          {chatOpen && (
            <div className="chat-widget-panel" style={{ position: 'absolute', top: 0, right: 0, width: '400px', height: '500px', border: '1px solid #ccc' }}>
              <Chat viewType="chat" sessions={sessions} activeSessionId={activeSessionId} 
                    isLoading={loading} className="reachat-chat">
                <SessionMessagePanel allowBack={false /* hide back button since no sidebar */}>
                  <SessionMessages>
                    {(conversations) =>
                      conversations.map(conv => (
                        <SessionMessage key={conv.id} conversation={conv} />
                      ))
                    }
                  </SessionMessages>
                  <ChatInput placeholder="Ask the chatbot..." />
                </SessionMessagePanel>
              </Chat>
            </div>
          )}
    ```  

   In this snippet, we render the `<Chat>` component (only when `chatOpen` is true) inside an absolutely-positioned panel (so it appears as an overlay panel – you can adjust styling or use a modal as desired). Key parts to note:  
   - We set `viewType="chat"` to show only the chat interface without the session list ([Components](https://reachat.dev/docs/api/components#:~:text=,console)) (ideal for a pop-up widget).  
   - We include a `<SessionMessagePanel>` containing `<SessionMessages>` and a `<ChatInput>` field. These are ReaChat sub-components that display the conversation history and the input box respectively. We pass `allowBack={false}` to the panel to disable any “back” button (since we’re not using multiple sessions here).  
   - We will manage the `sessions`, `activeSessionId`, and `loading` state in our component (not shown in this snippet yet). These will be used to control the chat data and loading state. For now, ensure you’ve imported all needed components from `reachat`.  

5. **Initialize Chat Session with Textarea Content**  
   To give the chatbot context of the textarea’s content, we’ll start a new chat session that includes the current text. In the component state, set up a `sessions` array with one session that has a conversation entry containing the user’s text. For example:  

    ```jsx
    // Inside ChatWidget component (above the return statement)
    const [sessions, setSessions] = useState([
      {
        id: 'session1',
        title: 'Editing Session',
        createdAt: new Date(),
        updatedAt: new Date(),
        conversations: []  // we will populate this when chat opens
      }
    ]);
    const [activeSessionId, setActiveSessionId] = useState('session1');
    const [loading, setLoading] = useState(false);
    
    // When chatOpen becomes true, initialize the session with the current text content
    useEffect(() => {
      if (chatOpen) {
        setSessions(prev => prev.map(session => {
          if (session.id === 'session1') {
            // Add an initial conversation with the user's text as context
            const userMessage = {
              id: 'msg1',
              question: "Here is my text:\n\n" + text,  // user's content as the message
              response: undefined,
              createdAt: new Date()
            };
            return { ...session, conversations: [ userMessage ] };
          }
          return session;
        }));
      }
    }, [chatOpen]);
    ```  

   In this code, when the chat is opened (`chatOpen` switches to true), we update the session’s conversations to include a first message from the “user” that contains the textarea content. We format it as if the user sent a message: “Here is my text: …” followed by the content. This message will appear in the chat window, providing context to the chatbot. (In a real scenario, you might designate this as a system or context message, but for simplicity we treat it as a user message.) The ReaChat data model expects each conversation entry to have a **`question`** (user prompt) and eventually a **`response`** (AI reply) ([Models](https://reachat.dev/docs/api/models#:~:text=%2F,question%3A%20string)). We leave `response` undefined for now since the AI hasn’t responded yet.  

   > **Note:** We use `useEffect` to update the session when chat opens. This ensures the content is captured at the moment the user clicks the button. We also set `activeSessionId` to our session’s ID so that the Chat knows which session’s conversations to display.

6. **Send User Prompts to the Chatbot (Integrate the AI Backend)**  
   With the UI in place and the session initialized, we need to handle the chatbot responses. ReaChat is frontend-only – you must connect it to an AI model or backend API of your choice ([Providers](https://reachat.dev/docs/examples/providers#:~:text=Providers)). For example, you can use OpenAI’s ChatGPT API to generate responses. When the user types a message in the ChatInput and hits send, we should call our AI API with the conversation (which includes the textarea content and the user’s latest question) and then update the chat with the AI’s answer. 

   **Example using OpenAI API:** Suppose we have an endpoint or function to get the AI response. We can intercept the sending of a message and use `setLoading(true)` while the request is in progress, then append the AI’s answer to the conversations. One approach is to listen for new user messages and handle them in a `useEffect`:  

    ```jsx
    useEffect(() => {
      // Whenever conversations change, if the last message has no response, call the AI
      const session = sessions.find(s => s.id === activeSessionId);
      if (!session) return;
      const convos = session.conversations;
      if (!loading && convos.length > 0) {
        const lastConv = convos[convos.length - 1];
        // If the last conversation entry has no response (meaning it's a user prompt awaiting answer)
        if (lastConv.response === undefined) {
          setLoading(true);
          // Call AI API with the conversation history
          fetch('/api/generate', { 
            method: 'POST', 
            body: JSON.stringify({ messages: convos.map(c => c.question) }) 
          })
            .then(res => res.json())
            .then(data => {
              const aiText = data.reply;  // The AI's reply text from your API
              // Update the last conversation entry with the AI's response
              setSessions(prev => prev.map(s => {
                if (s.id !== activeSessionId) return s;
                return {
                  ...s,
                  conversations: s.conversations.map(c => 
                    c.id === lastConv.id ? { ...c, response: aiText, updatedAt: new Date() } : c
                  )
                };
              }));
            })
            .catch(err => {
              console.error("Error fetching AI response", err);
            })
            .finally(() => setLoading(false));
        }
      }
    }, [sessions, activeSessionId]);
    ```  

   In this effect, whenever the `sessions` state updates, we check if the latest conversation entry is missing a `response`. If so, we call an API (here a placeholder `POST /api/generate`) to get a reply. This should be replaced with your actual AI call – e.g., an OpenAI Chat Completion API call with the messages (you’d include the user messages and possibly a system prompt instructing the AI it can edit text). Once the API returns, we update the session state by filling in the `response` field of the last conversation. This causes the ReaChat UI to display the assistant’s answer below the user’s question. We also manage a `loading` state to indicate to the Chat component when a response is in progress (`isLoading={loading}` prop on `<Chat>` will show a typing indicator or spinner).

   > **Tip:** Instead of manually intercepting `sessions` changes, you can also use ReaChat’s context utilities. The `useSessionContext()` hook provides a `sendMessage()` function that will add a user message to the chat for you ([Custom Components](https://reachat.dev/docs/customization/custom#:~:text=import%20,reachat)). You could create a custom handler on the ChatInput to call `sendMessage` and then perform the API call. In this tutorial, we manually manage state for clarity, but advanced implementations can leverage ReaChat’s context methods.

7. **Update the Textarea Based on Chatbot Responses**  
   The chatbot can now have a general conversation about the text. We also want it to be able to **edit or rewrite the textarea content** either when the user explicitly asks or via a special instruction. To achieve this, we need to detect when an AI response is meant to update the text content and then set our `text` state accordingly. There are a couple of ways to do this: 

   - **By User Instruction:** The simplest scenario is when the user explicitly asks the bot to modify the text (e.g., "Please rewrite the text in a formal tone"). The assistant’s reply will likely contain the revised text. We can detect such a case by the context or format of the reply. For example, if the assistant response is just a large block of text (and not a general answer), we can assume it’s the new version of the textarea content. In our code, after updating the sessions with the AI response, we can add:  
     ```jsx
     // After setting the response in the sessions (inside .then of fetch):
     if (aiText && lastConv.question.toLowerCase().includes("rewrite")) {
       setText(aiText);  // replace textarea content with the AI's revised text
     }
     ```  
     This simplistic check looks if the user’s prompt included "rewrite" and then replaces the original text with the AI’s output. In a real app, you might use a more robust trigger or even a user confirmation before replacing the text.  

   - **Structured Output (Tool Usage):** A more advanced approach is to have the AI respond with a special formatted output when it intends to perform an action (like updating the text). For instance, you could design the bot to output a JSON or a markdown code block containing the new text. Then your frontend can parse that format and update the textarea accordingly. ReaChat supports custom Markdown parsing via **remark plugins** ([Markdown Plugins](https://reachat.dev/docs/customization/markdown#:~:text=reachat%20uses%20remark%20to%20parse,should%20work%20here%20as%20well)), so you could even have the bot output a token like `<UPDATE>new text here</UPDATE>` and use a plugin or regex to intercept it. For example:  
     ```jsx
     // Pseudocode: parse AI response for structured update command
     if (aiText.startsWith("UPDATE_TEXT:")) {
       const newText = aiText.replace("UPDATE_TEXT:", "").trim();
       setText(newText);
     }
     ```  
     In this approach, the assistant would be prompted (in the system instruction) to prefix any direct text edit suggestions with `"UPDATE_TEXT:"`. The code checks for that prefix and updates the textarea when it appears. You can get creative with such conventions. Just ensure the format is agreed upon in the prompt design for the AI.  

   After updating the textarea state (`text`), the UI will reflect the changes immediately (because the textarea’s value is bound to `text`). This means the user’s text is now replaced or modified as per the chatbot’s suggestion. The conversation in the ReaChat widget can continue – for example, the assistant might say “I've updated the text above.” and the user can further ask questions or make more modifications. 

8. **Polish the UI and Configuration**  
   Finally, refine the interface for a better user experience: 
   - You might want to add a close button for the chat widget so users can hide it after use (e.g., an “×” in the corner that sets `setChatOpen(false)`).  
   - Customize the **ChatInput** placeholder or allowed file types if needed. For instance, we set a custom placeholder "Ask the chatbot..." in the code above. You can also disable file attachments by excluding or customizing the attach icon (via `ChatInput` props).  
   - Adjust styling of the chat panel. You can apply Tailwind classes or custom CSS to `.chat-widget-panel` for a nicer look (e.g. a drop shadow, rounded corners, etc.). Since ReaChat components use Tailwind, you can also override theme colors via the `theme` prop or customize the default theme if desired.  
   - **ReaChat Config Options:** Remember that ReaChat offers other view types and settings. We used `viewType="chat"` for a single-session widget, but there is also a `"companion"` view that is a compact sidebar-style chat, and a `"console"` view for a full-page chat UI ([Components](https://reachat.dev/docs/api/components#:~:text=,console)). Choose what fits your app’s design. Additionally, the `remarkPlugins` prop on `<Chat>` can enable rich rendering or custom parsing – for example, you could use it to render special responses (like the structured outputs mentioned) in a custom way ([Markdown Plugins](https://reachat.dev/docs/customization/markdown#:~:text=reachat%20uses%20remark%20to%20parse,should%20work%20here%20as%20well)).  

   With these steps, you have a React component that contains a textarea for text editing and a ReaChat-powered chatbot that can discuss and even modify that text. The user can click **“Chat with this”**, get a conversational assistant to help with their content, and apply changes suggested by the AI directly to the text area. This provides a smooth, interactive experience for editing or refining text with the help of a chatbot. Enjoy your new integrated ReaChat widget!

**Sources:** The ReaChat documentation and examples were referenced for setup and component usage ([Setup](https://reachat.dev/docs/getting-started/setup#:~:text=import%20,from%20%27reablocks)) ([Components](https://reachat.dev/docs/api/components#:~:text=,console)). ReaChat is an open-source React library for chat UIs, which uses Tailwind CSS for styling and offers customizable components ([reachat - Build Chat Experiences in Hours, Not Weeks.](https://reachat.dev/#:~:text=Chat%20UIs%20without%20all%20the,baggage)) ([Components](https://reachat.dev/docs/api/components#:~:text=Chat)). The integration approach above combines the ReaChat UI components with an AI backend (e.g. OpenAI) to achieve a functional chatbot that can update app state.