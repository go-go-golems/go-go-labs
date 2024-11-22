const DEBUG = true;

function debugLog(...args) {
    if (DEBUG) {
        console.log('[DEBUG]', ...args);
    }
}

export class SSEClient {
    constructor(url, options = {}) {
        this.url = url;
        this.eventSource = null;
        this.reconnectTimeout = null;
        this.handlers = {
            onConnected: options.onConnected || (() => {}),
            onThinking: options.onThinking || (() => {}),
            onToken: options.onToken || (() => {}),
            onDone: options.onDone || (() => {}),
            onDisconnected: options.onDisconnected || (() => {}),
        };
        this.clientId = null;
        this.connected = false;
    }

    connect() {
        debugLog('Initializing SSE connection');
        this._cleanup();

        this.eventSource = new EventSource(this.url);

        this.eventSource.onopen = () => {
            debugLog('SSE connection opened');
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
                    this.handlers.onConnected(data.content);
                    break;
                case 'thinking':
                    debugLog('Thinking state started');
                    this.handlers.onThinking();
                    break;
                case 'token':
                    this.handlers.onToken(data.content);
                    break;
                case 'done':
                    debugLog('Response complete');
                    this.handlers.onDone();
                    break;
            }
        };

        this.eventSource.onerror = (error) => {
            debugLog('SSE error:', error);
            this.connected = false;
            this.handlers.onDisconnected();
            
            if (this.eventSource) {
                this.eventSource.close();
                this.eventSource = null;
            }

            if (!this.reconnectTimeout) {
                debugLog('Scheduling SSE reconnection');
                this.reconnectTimeout = setTimeout(() => {
                    debugLog('Attempting SSE reconnection');
                    this.connect();
                }, 5000);
            }
        };
    }

    async sendMessage(messages) {
        if (!this.connected || !this.clientId) {
            throw new Error('Not connected to server');
        }

        debugLog('Sending POST request');
        const response = await fetch(this.url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Client-ID': this.clientId,
            },
            body: JSON.stringify({ messages }),
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        debugLog('POST request successful');
    }

    disconnect() {
        this._cleanup();
    }

    _cleanup() {
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
}
