import { useCallback } from 'react';
import { v4 as uuidv4 } from 'uuid';
import { MessageRole } from '../types/index.js';
import Anthropic from '@anthropic-ai/sdk';
import { useAppDispatch, useAppSelector } from '../store/hooks.js';
import { addMessage, setLoading, setError } from '../store/chatSlice.js';
import type { ChatState, ChatMessage } from '../store/chatSlice.js';

// Initialize your LLM client - will need an API key to work
const anthropic = new Anthropic({
  apiKey: process.env.ANTHROPIC_API_KEY || '',
});

export function useChat() {
  const dispatch = useAppDispatch();
  const messages = useAppSelector((state: { chat: ChatState }) => state.chat.messages);
  const isLoading = useAppSelector((state: { chat: ChatState }) => state.chat.isLoading);
  const error = useAppSelector((state: { chat: ChatState }) => state.chat.error);
  
  // Helper function to add a new message to the conversation
  const createMessage = useCallback((role: MessageRole, content: string) => {
    const newMessage: ChatMessage = {
      id: uuidv4(),
      role,
      content,
      timestamp: new Date()
    };
    
    dispatch(addMessage(newMessage));
    return newMessage;
  }, [dispatch]);

  // Main function to send a message and get a response
  const sendMessage = useCallback(async (content: string) => {
    if (!content.trim()) return null;
    
    // Add user message to the conversation
    createMessage('user', content);
    
    try {
      dispatch(setLoading(true));
      dispatch(setError(null));
      
      // Format messages for Anthropic API
      const apiMessages = messages.map((msg: ChatMessage) => ({
        role: msg.role,
        content: msg.content
      }));
      
      // Add the current message
      apiMessages.push({
        role: 'user' as const,
        content
      });
      
      // Check for API key
      if (!process.env.ANTHROPIC_API_KEY) {
        throw new Error('ANTHROPIC_API_KEY is not set in environment variables');
      }
      
      // Call the LLM API
      const response = await anthropic.messages.create({
        model: 'claude-3-sonnet-20240229',
        max_tokens: 1000,
        messages: apiMessages,
      });
      
      // Add the assistant response to our conversation
      const assistantMessage = createMessage('assistant', response.content[0].text);
      dispatch(setLoading(false));
      return assistantMessage;
      
    } catch (err) {
      // Handle errors gracefully
      dispatch(setLoading(false));
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      dispatch(setError(errorMessage));
      return null;
    }
  }, [messages, createMessage, dispatch]);
  
  // Return everything needed by components
  return {
    messages,
    isLoading,
    error,
    sendMessage
  };
} 