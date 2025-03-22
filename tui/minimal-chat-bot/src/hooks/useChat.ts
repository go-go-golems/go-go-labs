import { useState, useCallback } from 'react';
import { v4 as uuidv4 } from 'uuid';
import { Message, MessageRole } from '../types/index.js';
import Anthropic from '@anthropic-ai/sdk';

// Initialize your LLM client - will need an API key to work
const anthropic = new Anthropic({
  apiKey: process.env.ANTHROPIC_API_KEY || '',
});

export function useChat() {
  // State for messages, loading status, and errors
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Helper function to add a new message to the conversation
  const addMessage = useCallback((role: MessageRole, content: string) => {
    const newMessage: Message = {
      id: uuidv4(),
      role,
      content,
      timestamp: new Date()
    };
    
    setMessages(prevMessages => [...prevMessages, newMessage]);
    return newMessage;
  }, []);

  // Main function to send a message and get a response
  const sendMessage = useCallback(async (content: string) => {
    if (!content.trim()) return null;
    
    // Add user message to the conversation
    addMessage('user', content);
    
    try {
      setIsLoading(true);
      setError(null);
      
      // Format messages for Anthropic API
      const apiMessages = messages.map(msg => ({
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
      const assistantMessage = addMessage('assistant', response.content[0].text);
      setIsLoading(false);
      return assistantMessage;
      
    } catch (err) {
      // Handle errors gracefully
      setIsLoading(false);
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      return null;
    }
  }, [messages, addMessage]);
  
  // Return everything needed by components
  return {
    messages,
    isLoading,
    error,
    sendMessage,
    addMessage
  };
} 