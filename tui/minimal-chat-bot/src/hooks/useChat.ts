import { useCallback } from 'react';
import { v4 as uuidv4 } from 'uuid';
import { MessageRole } from '../types/index.js';
import Anthropic from '@anthropic-ai/sdk';
import { useAppDispatch, useAppSelector } from '../store/hooks.js';
import { addMessage, setLoading, setError } from '../store/chatSlice.js';
import type { ChatState, ChatMessage } from '../store/chatSlice.js';
import { createLogger } from '../utils/index.js';

// Create a logger for this hook
const logger = createLogger('useChat');

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
    };
    
    logger.debug('Creating new message', { messageId: newMessage.id, role });
    dispatch(addMessage(newMessage));
    return newMessage;
  }, [dispatch]);

  // Main function to send a message and get a response
  const sendMessage = useCallback(async (content: string) => {
    if (!content.trim()) {
      logger.debug('Empty message received, ignoring');
      return null;
    }
    
    // Add user message to the conversation
    logger.info('Processing user message', { messageLength: content.length });
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
        logger.error('Missing API key', { env: Object.keys(process.env).filter(k => k.startsWith('ANTHR')) });
        throw new Error('ANTHROPIC_API_KEY is not set in environment variables');
      }
      
      logger.info('Calling LLM API', { 
        model: 'claude-3-sonnet-20240229', 
        messageCount: apiMessages.length 
      });
      
      // Call the LLM API
      const response = await anthropic.messages.create({
        model: 'claude-3-sonnet-20240229',
        max_tokens: 1000,
        messages: apiMessages,
      });
      
      logger.info('Received LLM response', { 
        contentLength: response.content[0].text.length,
        usage: response.usage
      });
      
      // Add the assistant response to our conversation
      const assistantMessage = createMessage('assistant', response.content[0].text);
      dispatch(setLoading(false));
      return assistantMessage;
      
    } catch (err) {
      // Handle errors gracefully
      dispatch(setLoading(false));
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      logger.error('Error in chat interaction', { 
        error: errorMessage,
        stack: err instanceof Error ? err.stack : undefined 
      });
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