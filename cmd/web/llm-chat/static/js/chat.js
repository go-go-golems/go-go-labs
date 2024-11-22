import { LitElement, html, css } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

const DEBUG = true;

function debugLog(...args) {
    if (DEBUG) {
        console.log('[DEBUG]', ...args);
    }
}

class ChatApp extends LitElement {
    static properties = {
        messages: { type: Array },
        isThinking: { type: Boolean },
        currentResponse: { type: String },
        clientId: { type: String },
        connected: { type: Boolean },
    };

    static styles = css`
        #chat-container {
            background-color: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            height: 500px;
            display: flex;
            flex-direction: column;
        }
        #chat-messages {
            flex-grow: 1;
            overflow-y: auto;
            margin-bottom: 20px;
            padding: 10px;
        }
        .message {
            margin: 10px 0;
            padding: 10px;
            border-radius: 5px;
        }
        .user-message {
            background-color: #e3f2fd;
            margin-left: 20%;
        }
        .assistant-message {
            background-color: #f5f5f5;
            margin-right: 20%;
        }
        .thinking {
            color: #666;
            font-style: italic;
            margin: 10px 0;
            padding: 10px;
        }
        .system-message {
            color: #666;
            font-style: italic;
            margin: 10px 0;
            padding: 10px;
        }
        #input-container {
            display: flex;
            gap: 10px;
        }
        input {
            flex-grow: 1;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 16px;
        }
        button {
            padding: 10px 20px;
            background-color: #2196f3;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover {
            background-color: #1976d2;
        }
        button:disabled {
            background-color: #ccc;
            cursor: not-allowed;
        }
    `;

    constructor() {
        super();
        this.messages = [];
        this.isThinking = false;
        this.currentResponse = '';
        this.clientId = null;
        this.connected = false;
        this.eventSource = null;
        this.reconnectTimeout = null;
        debugLog('ChatApp initialized');
        this._initializeSSE();
    }

    disconnectedCallback() {
        super.disconnectedCallback();
        this._cleanupSSE();
    }

    _cleanupSSE() {
        if (this.eventSource) {
            debugLog('Cleaning up SSE connection');
            this.eventSource.close();
            this.eventSource = null;
        }
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }
    }

    _initializeSSE() {
        debugLog('Initializing SSE connection');
        // Clean up any existing connection first
        this._cleanupSSE();

        this.eventSource = new EventSource('/api/chat');

        this.eventSource.onopen = () => {
            debugLog('SSE connection opened');
            // Clear any pending reconnect timeout
            if (this.reconnectTimeout) {
                clearTimeout(this.reconnectTimeout);
                this.reconnectTimeout = null;
            }
        };

        this.eventSource.onmessage = (event) => {
            debugLog('Received SSE event:', event.data);
            const data = JSON.parse(event.data);
            
            switch (data.type) {
                case 'connected':
                    debugLog('Received client ID:', data.content);
                    this.clientId = data.content;
                    this.connected = true;
                    break;
                case 'thinking':
                    debugLog('Thinking state started');
                    this.isThinking = true;
                    break;
                case 'token':
                    this.currentResponse += data.content;
                    break;
                case 'done':
                    debugLog('Response complete');
                    this.isThinking = false;
                    this.messages = [...this.messages, {
                        role: 'assistant',
                        content: this.currentResponse
                    }];
                    this.currentResponse = '';
                    break;
            }
        };

        this.eventSource.onerror = (error) => {
            debugLog('SSE error:', error);
            this.connected = false;
            
            // Clean up the existing connection
            if (this.eventSource) {
                this.eventSource.close();
                this.eventSource = null;
            }

            // Only set up reconnect if we don't already have one pending
            if (!this.reconnectTimeout) {
                debugLog('Scheduling SSE reconnection');
                this.reconnectTimeout = setTimeout(() => {
                    debugLog('Attempting SSE reconnection');
                    this._initializeSSE();
                }, 5000);
            }
        };
    }

    async _sendMessage() {
        const input = this.shadowRoot.querySelector('input');
        const message = input.value.trim();
        if (!message || !this.connected) return;

        debugLog('Sending message:', message);

        // Add user message to conversation
        const userMessage = { role: 'user', content: message };
        this.messages = [...this.messages, userMessage];
        input.value = '';
        this.currentResponse = '';
        this.isThinking = true;

        // Send the message
        try {
            debugLog('Sending POST request');
            const response = await fetch('/api/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Client-ID': this.clientId,
                },
                body: JSON.stringify({ messages: this.messages }),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            debugLog('POST request successful');
        } catch (error) {
            debugLog('Error sending message:', error);
            this.isThinking = false;
            this.messages = [...this.messages, {
                role: 'assistant',
                content: 'Sorry, there was an error processing your message.'
            }];
        }
    }

    updated(changedProperties) {
        if (changedProperties.has('messages') || changedProperties.has('currentResponse')) {
            debugLog('State updated:', {
                messages: this.messages,
                currentResponse: this.currentResponse,
                isThinking: this.isThinking
            });
            // Scroll to bottom when messages update
            const chatMessages = this.shadowRoot.querySelector('#chat-messages');
            if (chatMessages) {
                chatMessages.scrollTop = chatMessages.scrollHeight;
            }
        }
    }

    _handleKeyPress(e) {
        if (e.key === 'Enter' && !this.isThinking && this.connected) {
            debugLog('Enter key pressed, sending message');
            this._sendMessage();
        }
    }

    render() {
        return html`
            <div id="chat-container">
                <div id="chat-messages">
                    ${!this.connected ? html`
                        <div class="message system-message">
                            Connecting to server...
                        </div>
                    ` : ''}
                    ${this.messages.map(msg => html`
                        <div class="message ${msg.role}-message">
                            ${msg.content}
                        </div>
                    `)}
                    ${this.currentResponse ? html`
                        <div class="message assistant-message">
                            ${this.currentResponse}
                        </div>
                    ` : ''}
                    ${this.isThinking ? html`
                        <div class="thinking">Thinking...</div>
                    ` : ''}
                </div>
                <div id="input-container">
                    <input type="text" 
                           @keypress=${this._handleKeyPress} 
                           placeholder="${this.connected ? 'Type your message...' : 'Connecting...'}"
                           ?disabled=${!this.connected || this.isThinking}>
                    <button @click=${this._sendMessage} 
                            ?disabled=${!this.connected || this.isThinking}>
                        Send
                    </button>
                </div>
            </div>
        `;
    }
}

customElements.define('chat-app', ChatApp);
