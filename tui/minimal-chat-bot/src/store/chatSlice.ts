import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { createLogger, LogLevel } from '../utils/logger.js';

// Create a logger for this slice
const logger = createLogger('chatSlice', LogLevel.DEBUG);

// Log slice initialization
logger.info('Initializing chat slice');

export interface ChatMessage {
  id: string;
  content: string;
  role: 'user' | 'assistant';
}

export interface ChatState {
  messages: ChatMessage[];
  isLoading: boolean;
  error: string | null;
}

const initialState: ChatState = {
  messages: [],
  isLoading: false,
  error: null,
};

export const chatSlice = createSlice({
  name: 'chat',
  initialState,
  reducers: {
    setLoading: (state, action: PayloadAction<boolean>) => {
      logger.debug('Setting loading state', { 
        previousState: state.isLoading,
        newState: action.payload 
      });
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      logger.info('Setting error state', { 
        previousError: state.error,
        newError: action.payload 
      });
      state.error = action.payload;
    },
    addMessage: (state, action: PayloadAction<ChatMessage>) => {
      logger.debug('Adding message', { 
        messageId: action.payload.id,
        role: action.payload.role,
        contentLength: action.payload.content.length,
        timestamp: new Date().toISOString()
      });
      state.messages.push(action.payload);
      logger.debug('Messages updated', { 
        totalMessages: state.messages.length,
        lastMessageId: action.payload.id
      });
    },
    clearMessages: (state) => {
      logger.info('Clearing all messages', {
        messageCount: state.messages.length,
        timestamp: new Date().toISOString()
      });
      state.messages = [];
    },
  },
});

// Log slice creation
logger.info('Chat slice created');

export const { setLoading, setError, addMessage, clearMessages } = chatSlice.actions;

export default chatSlice.reducer; 