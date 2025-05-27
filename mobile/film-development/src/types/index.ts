export interface Roll {
  id: number;
  roll: number;
  film: string;
  pushPull: string;
  devTime: string;
  notes: string;
}

export interface ProcessStep {
  id: string;
  name: string;
  time: string;
  description: string;
}

export interface TimerState {
  currentRoll: number;
  currentStep: number;
  timeLeft: number;
  isRunning: boolean;
  isPaused: boolean;
  completedSteps: number[];
}

export interface NotificationData {
  id: string;
  stepName: string;
  rollNumber: number;
  scheduledTime: number;
}