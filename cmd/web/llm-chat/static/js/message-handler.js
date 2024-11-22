const DEBUG = true;

function debugLog(...args) {
    if (DEBUG) {
        console.log('[DEBUG]', ...args);
    }
}

export class MessageHandler {
    constructor() {
        this.messages = [];
        this.currentResponse = '';
        this.isThinking = false;
        this.handlers = {
            onStateChange: () => {},
            onError: () => {},
        };
    }

    setHandlers(handlers) {
        this.handlers = { ...this.handlers, ...handlers };
    }

    handleUserMessage(message) {
        debugLog('Adding user message:', message);
        this.messages = [...this.messages, { role: 'user', content: message }];
        this.currentResponse = '';
        this.isThinking = true;
        this._notifyStateChange();
    }

    handleThinking() {
        debugLog('Setting thinking state');
        this.isThinking = true;
        this._notifyStateChange();
    }

    handleToken(token) {
        this.currentResponse += token;
        this._notifyStateChange();
    }

    handleDone() {
        debugLog('Handling done state');
        this.isThinking = false;
        if (this.currentResponse) {
            this.messages = [...this.messages, {
                role: 'assistant',
                content: this.currentResponse
            }];
        }
        this.currentResponse = '';
        this._notifyStateChange();
    }

    handleError(error) {
        debugLog('Handling error:', error);
        this.isThinking = false;
        this.messages = [...this.messages, {
            role: 'assistant',
            content: 'Sorry, there was an error processing your message.'
        }];
        this.currentResponse = '';
        this._notifyStateChange();
        this.handlers.onError(error);
    }

    getState() {
        return {
            messages: this.messages,
            currentResponse: this.currentResponse,
            isThinking: this.isThinking,
        };
    }

    _notifyStateChange() {
        this.handlers.onStateChange(this.getState());
    }
}
