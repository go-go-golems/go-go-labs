import * as SQLite from 'expo-sqlite';
import { Roll, ProcessStep, TimerState } from '../types';
import rollsData from '../data/rolls.json';
import stepsData from '../data/steps.json';

let db: SQLite.SQLiteDatabase;

export const initializeDatabase = async (): Promise<void> => {
  db = await SQLite.openDatabaseAsync('filmDevTimer.db');
  
  // Create tables
  await db.execAsync(`
    CREATE TABLE IF NOT EXISTS rolls (
      id INTEGER PRIMARY KEY,
      roll INTEGER,
      film TEXT,
      pushPull TEXT,
      devTime TEXT,
      notes TEXT
    );
  `);
  
  await db.execAsync(`
    CREATE TABLE IF NOT EXISTS process_steps (
      id TEXT PRIMARY KEY,
      name TEXT,
      time TEXT,
      description TEXT
    );
  `);
  
  await db.execAsync(`
    CREATE TABLE IF NOT EXISTS timer_states (
      id INTEGER PRIMARY KEY,
      currentRoll INTEGER,
      currentStep INTEGER,
      timeLeft INTEGER,
      isRunning INTEGER,
      isPaused INTEGER,
      completedSteps TEXT,
      lastUpdated INTEGER
    );
  `);
  
  await db.execAsync(`
    CREATE TABLE IF NOT EXISTS settings (
      key TEXT PRIMARY KEY,
      value TEXT
    );
  `);
  
  // Check if data exists, if not, seed it
  const rollCount = await db.getFirstAsync('SELECT COUNT(*) as count FROM rolls') as { count: number };
  if (rollCount.count === 0) {
    await seedData();
  }
};

const seedData = async (): Promise<void> => {
  // Insert rolls
  for (const roll of rollsData) {
    await db.runAsync(
      'INSERT INTO rolls (id, roll, film, pushPull, devTime, notes) VALUES (?, ?, ?, ?, ?, ?)',
      [roll.id, roll.roll, roll.film, roll.pushPull, roll.devTime, roll.notes]
    );
  }
  
  // Insert process steps
  for (const step of stepsData) {
    await db.runAsync(
      'INSERT INTO process_steps (id, name, time, description) VALUES (?, ?, ?, ?)',
      [step.id, step.name, step.time, step.description]
    );
  }
};

export const getRolls = async (): Promise<Roll[]> => {
  const result = await db.getAllAsync('SELECT * FROM rolls ORDER BY roll');
  return result as Roll[];
};

export const getProcessSteps = async (): Promise<ProcessStep[]> => {
  const result = await db.getAllAsync('SELECT * FROM process_steps');
  return result as ProcessStep[];
};

export const saveTimerState = async (state: TimerState): Promise<void> => {
  const completedStepsJson = JSON.stringify(state.completedSteps);
  await db.runAsync(
    `INSERT OR REPLACE INTO timer_states 
     (id, currentRoll, currentStep, timeLeft, isRunning, isPaused, completedSteps, lastUpdated) 
     VALUES (1, ?, ?, ?, ?, ?, ?, ?)`,
    [
      state.currentRoll,
      state.currentStep,
      state.timeLeft,
      state.isRunning ? 1 : 0,
      state.isPaused ? 1 : 0,
      completedStepsJson,
      Date.now()
    ]
  );
};

export const getTimerState = async (): Promise<TimerState | null> => {
  const result = await db.getFirstAsync('SELECT * FROM timer_states WHERE id = 1') as any;
  if (!result) return null;
  
  return {
    currentRoll: result.currentRoll,
    currentStep: result.currentStep,
    timeLeft: result.timeLeft,
    isRunning: result.isRunning === 1,
    isPaused: result.isPaused === 1,
    completedSteps: JSON.parse(result.completedSteps || '[]')
  };
};

export const saveSetting = async (key: string, value: string): Promise<void> => {
  await db.runAsync(
    'INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)',
    [key, value]
  );
};

export const getSetting = async (key: string): Promise<string | null> => {
  const result = await db.getFirstAsync('SELECT value FROM settings WHERE key = ?', [key]) as { value: string } | null;
  return result?.value || null;
};