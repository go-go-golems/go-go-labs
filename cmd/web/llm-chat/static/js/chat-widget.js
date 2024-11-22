import { LitElement, html, css } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
import { SSEClient } from './sse-client.js';
import { MessageHandler } from './message-handler.js';

const DEBUG = true;

function debugLog(...args) {
    if (DEBUG) {
        console.log('[DEBUG]', ...args);
    }
}

export class ChatWidget extends LitElement {
    static properties = {
        messages: { type: Array },
        isThinking: { type: Boolean },
        currentResponse: { type: String },
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
        this.connected = false;

        // Initialize message handler
        this.messageHandler = new MessageHandler();
        this.messageHandler.setHandlers({
            onStateChange: (state) => {
                this.messages = state.messages;
                this.currentResponse = state.currentResponse;
                this.isThinking = state.isThinking;
            },
            onError: (error) => {
                debugLog('Error in message handler:', error);
            },
        });

        // Initialize SSE client
        this.sseClient = new SSEClient('/api/chat', {
            onConnected: () => {
                this.connected = true;
            },
            onThinking: () => {
                this.messageHandler.handleThinking();
            },
            onToken: (token) => {
                this.messageHandler.handleToken(token);
            },
            onDone: () => {
                this.messageHandler.handleDone();
            },
            onDisconnected: () => {
                this.connected = false;
            },
        });

        debugLog('ChatWidget initialized');
        this.sseClient.connect();
    }

    disconnectedCallback() {
        super.disconnectedCallback();
        this.sseClient.disconnect();
    }

    async _sendMessage() {
        const input = this.shadowRoot.querySelector('input');
        const message = input.value.trim();
        if (!message || !this.connected) return;

        debugLog('Sending message:', message);
        input.value = '';

        try {
            this.messageHandler.handleUserMessage(message);
            await this.sseClient.sendMessage(this.messages);
        } catch (error) {
            debugLog('Error sending message:', error);
            this.messageHandler.handleError(error);
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

customElements.define('chat-widget', ChatWidget);
