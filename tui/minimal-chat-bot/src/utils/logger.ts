import fs from 'fs';
import path from 'path';

// Log file path
const LOG_FILE = '/tmp/chatbot.log';

// Ensure log directory exists (not needed for /tmp but included for completeness)
try {
  fs.mkdirSync(path.dirname(LOG_FILE), { recursive: true });
} catch (err) {
  // Ignore errors if directory already exists
}

// Log levels
export enum LogLevel {
  DEBUG = 'DEBUG',
  INFO = 'INFO',
  WARN = 'WARN',
  ERROR = 'ERROR',
}

type LogEntry = {
  timestamp: string;
  level: LogLevel;
  component: string;
  message: string;
  data?: any;
};

/**
 * Structured logger
 */
export class Logger {
  private component: string;

  constructor(component: string) {
    this.component = component;
  }

  /**
   * Log a message with structured data
   */
  private log(level: LogLevel, message: string, data?: any): void {
    const timestamp = new Date().toISOString();
    
    const logEntry: LogEntry = {
      timestamp,
      level,
      component: this.component,
      message,
      ...(data !== undefined ? { data } : {}),
    };
    
    const logString = JSON.stringify(logEntry) + '\n';
    
    // Append to log file
    try {
      fs.appendFileSync(LOG_FILE, logString);
    } catch (err) {
      // If writing to file fails, fall back to console
      console.error('Failed to write to log file:', err);
      console.error(logString);
    }
  }

  debug(message: string, data?: any): void {
    this.log(LogLevel.DEBUG, message, data);
  }

  info(message: string, data?: any): void {
    this.log(LogLevel.INFO, message, data);
  }

  warn(message: string, data?: any): void {
    this.log(LogLevel.WARN, message, data);
  }

  error(message: string, data?: any): void {
    this.log(LogLevel.ERROR, message, data);
  }
}

/**
 * Create a logger for a specific component
 */
export function createLogger(component: string): Logger {
  return new Logger(component);
} 