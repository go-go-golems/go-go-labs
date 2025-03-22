// Define message types
export type MessageRole = 'user' | 'assistant';

export interface Message {
  id: string;
  role: MessageRole;
  content: string;
  timestamp: Date;
}

// Theme interface for consistent styling
export interface Theme {
  primary: string;
  secondary: string;
  accent: string;
  error: string;
  text: string;
  dimText: string;
  border: string;
} 