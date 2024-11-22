// Message handler for managing chat state and message processing
export class MessageHandler {
    constructor() {
        this.messages = [];
        this.currentResponse = '';
        this.handlers = new Map();
    }

    // Add a message to the chat history
    addMessage(role, content) {
        this.messages.push({ role, content });
        this.notifyHandlers('messageAdded', { role, content });
    }

    // Add a token to the current response
    addToken(token) {
        this.currentResponse += token;
        this.notifyHandlers('tokenAdded', token);
    }

    // Finalize the current response
    finalizeResponse() {
        if (this.currentResponse) {
            this.addMessage('assistant', this.currentResponse);
            this.currentResponse = '';
            this.notifyHandlers('responseDone');
        }
    }

    // Register a handler for events
    on(event, handler) {
        if (!this.handlers.has(event)) {
            this.handlers.set(event, new Set());
        }
        this.handlers.get(event).add(handler);
    }

    // Remove a handler
    off(event, handler) {
        if (this.handlers.has(event)) {
            this.handlers.get(event).delete(handler);
        }
    }

    // Notify all handlers of an event
    notifyHandlers(event, data) {
        if (this.handlers.has(event)) {
            for (const handler of this.handlers.get(event)) {
                handler(data);
            }
        }
    }

    // Save conversation to a file
    async saveConversation(filename, clientId) {
        try {
            const response = await fetch('/api/conversation?op=save&filename=' + encodeURIComponent(filename), {
                method: 'POST',
                headers: {
                    'X-Client-ID': clientId
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to save conversation');
            }
            
            this.notifyHandlers('conversationSaved', filename);
        } catch (error) {
            console.error('Error saving conversation:', error);
            throw error;
        }
    }

    // Load conversation from a file
    async loadConversation(filename, clientId) {
        try {
            const response = await fetch('/api/conversation?op=load&filename=' + encodeURIComponent(filename), {
                method: 'POST',
                headers: {
                    'X-Client-ID': clientId
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to load conversation');
            }

            // Get the updated conversation
            const conversationResponse = await fetch('/api/conversation', {
                headers: {
                    'X-Client-ID': clientId
                }
            });

            if (!conversationResponse.ok) {
                throw new Error('Failed to get conversation');
            }

            const conversation = await conversationResponse.json();
            this.messages = conversation.messages || [];
            this.notifyHandlers('conversationLoaded', conversation);
        } catch (error) {
            console.error('Error loading conversation:', error);
            throw error;
        }
    }

    // Get current messages
    getMessages() {
        return this.messages;
    }

    // Clear all messages
    clear() {
        this.messages = [];
        this.currentResponse = '';
        this.notifyHandlers('cleared');
    }
}
