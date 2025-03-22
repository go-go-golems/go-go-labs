import chalk from 'chalk';

// Simple markdown parser for terminal output
export function applyMarkdown(text: string): string {
  // Bold
  text = text.replace(/\*\*(.*?)\*\*/g, (_, match) => chalk.bold(match));
  
  // Italic
  text = text.replace(/\*(.*?)\*/g, (_, match) => chalk.italic(match));
  
  // Code
  text = text.replace(/`(.*?)`/g, (_, match) => chalk.bgBlack.gray(` ${match} `));
  
  return text;
} 