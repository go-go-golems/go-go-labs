import { createSlice, PayloadAction } from '@reduxjs/toolkit';

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
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    addMessage: (state, action: PayloadAction<ChatMessage>) => {
      state.messages.push(action.payload);
    },
    clearMessages: (state) => {
      state.messages = [];
    },
  },
});

export const { setLoading, setError, addMessage, clearMessages } = chatSlice.actions;

export default chatSlice.reducer; 