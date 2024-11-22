import { LitElement, html, css } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
import { SSEClient } from './sse-client.js';
import { MessageHandler } from './message-handler.js';

export class ChatWidget extends LitElement {
    static properties = {
        messages: { type: Array },
        currentResponse: { type: String },
        isThinking: { type: Boolean },
        clientId: { type: String },
        isSaving: { type: Boolean },
        isLoading: { type: Boolean }
    };

    static styles = css`
        :host {
            display: block;
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }

        .chat-container {
            border: 1px solid #ccc;
            border-radius: 5px;
            padding: 20px;
            height: 500px;
            display: flex;
            flex-direction: column;
        }

        .messages {
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

        .user {
            background-color: #e3f2fd;
            margin-left: 20%;
        }

        .assistant {
            background-color: #f5f5f5;
            margin-right: 20%;
        }

        .input-container {
            display: flex;
            gap: 10px;
            margin-top: 10px;
        }

        input {
            flex-grow: 1;
            padding: 10px;
            border: 1px solid #ccc;
            border-radius: 5px;
        }

        button {
            padding: 10px 20px;
            background-color: #2196f3;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
        }

        button:disabled {
            background-color: #ccc;
            cursor: not-allowed;
        }

        .controls {
            display: flex;
            gap: 10px;
            margin-top: 10px;
        }

        .thinking {
            font-style: italic;
            color: #666;
        }

        .file-input {
            display: none;
        }
    `;

    constructor() {
        super();
        this.messages = [];
        this.currentResponse = '';
        this.isThinking = false;
        this.isSaving = false;
        this.isLoading = false;
        this.clientId = '';
        this.messageHandler = new MessageHandler();

        this.sseClient = new SSEClient('/api/chat', {
            onConnected: (clientId) => {
                this.clientId = clientId;
            },
            onThinking: () => {
                this.isThinking = true;
            },
            onToken: (token) => {
                this.messageHandler.addToken(token);
            },
            onDone: () => {
                this.messageHandler.finalizeResponse();
            },
            onDisconnected: () => {
                console.warn('Disconnected from SSE server');
            }
        });

        this._setupEventHandlers();
        this.sseClient.connect();
    }

    _setupEventHandlers() {
        // Message handler events
        this.messageHandler.on('messageAdded', ({ role, content }) => {
            this.messages = [...this.messages, { role, content }];
        });

        this.messageHandler.on('tokenAdded', (token) => {
            this.currentResponse = this.messageHandler.currentResponse;
        });

        this.messageHandler.on('responseDone', () => {
            this.currentResponse = '';
            this.isThinking = false;
        });

        this.messageHandler.on('conversationSaved', () => {
            this.isSaving = false;
        });

        this.messageHandler.on('conversationLoaded', () => {
            this.messages = this.messageHandler.getMessages();
            this.isLoading = false;
        });
    }

    async handleSubmit(e) {
        e.preventDefault();
        const input = this.shadowRoot.querySelector('input');
        const message = input.value.trim();
        
        if (message) {
            input.value = '';
            this.messageHandler.addMessage('user', message);
            
            try {
                await this.sseClient.sendMessage({
                    messages: this.messageHandler.getMessages()
                });
            } catch (error) {
                console.error('Error sending message:', error);
            }
        }
    }

    async handleSave() {
        const filename = prompt('Enter filename to save conversation:', 'conversation.json');
        if (filename) {
            this.isSaving = true;
            try {
                await this.messageHandler.saveConversation(filename, this.clientId);
            } catch (error) {
                alert('Error saving conversation: ' + error.message);
                this.isSaving = false;
            }
        }
    }

    async handleLoad() {
        const filename = prompt('Enter filename to load conversation:', 'conversation.json');
        if (filename) {
            this.isLoading = true;
            try {
                await this.messageHandler.loadConversation(filename, this.clientId);
            } catch (error) {
                alert('Error loading conversation: ' + error.message);
                this.isLoading = false;
            }
        }
    }

    handleClear() {
        if (confirm('Are you sure you want to clear the conversation?')) {
            this.messageHandler.clear();
            this.messages = [];
            this.currentResponse = '';
        }
    }

    render() {
        return html`
            <div class="chat-container">
                <div class="messages">
                    ${this.messages.map(msg => html`
                        <div class="message ${msg.role}">
                            ${msg.content}
                        </div>
                    `)}
                    ${this.currentResponse && html`
                        <div class="message assistant">
                            ${this.currentResponse}
                        </div>
                    `}
                    ${this.isThinking && html`
                        <div class="message thinking">
                            Thinking...
                        </div>
                    ` || ''}
                </div>
                <form @submit=${this.handleSubmit}>
                    <div class="input-container">
                        <input type="text" placeholder="Type your message..." ?disabled=${this.isThinking}>
                        <button type="submit" ?disabled=${this.isThinking}>Send</button>
                    </div>
                </form>
                <div class="controls">
                    <button @click=${this.handleSave} ?disabled=${this.isSaving || this.isThinking}>
                        ${this.isSaving ? 'Saving...' : 'Save'}
                    </button>
                    <button @click=${this.handleLoad} ?disabled=${this.isLoading || this.isThinking}>
                        ${this.isLoading ? 'Loading...' : 'Load'}
                    </button>
                    <button @click=${this.handleClear} ?disabled=${this.isThinking}>Clear</button>
                </div>
            </div>
        `;
    }
}

customElements.define('chat-widget', ChatWidget);
