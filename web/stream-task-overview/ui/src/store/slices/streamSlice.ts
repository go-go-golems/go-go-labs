import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface StreamInfo {
  title: string;
  description: string;
  startTime: string;
  language: string;
  githubRepo: string;
  currentTask: string;
  viewerCount: number;
}

interface StreamState {
  info: StreamInfo;
  completedSteps: string[];
  activeStep: string;
  upcomingSteps: string[];
  isEditing: boolean;
  isLoggedIn: boolean;
}

const initialState: StreamState = {
  info: {
    title: "Building a React Component Library",
    description: "Creating reusable UI components with TailwindCSS",
    startTime: new Date().toISOString(),
    language: "JavaScript/React",
    githubRepo: "https://github.com/yourusername/component-library",
    currentTask: "",
    viewerCount: 42,
  },
  completedSteps: [
    "Project setup and initialization",
    "Design system planning"
  ],
  activeStep: "Setting up component architecture",
  upcomingSteps: [
    "Implement Button component",
    "Create Card component",
    "Build Form elements",
    "Add dark mode toggle"
  ],
  isEditing: false,
  isLoggedIn: true
};

export const streamSlice = createSlice({
  name: 'stream',
  initialState,
  reducers: {
    setStreamInfo: (state, action: PayloadAction<StreamInfo>) => {
      state.info = action.payload;
    },
    toggleEditMode: (state) => {
      state.isEditing = !state.isEditing;
    },
    toggleLoggedIn: (state) => {
      state.isLoggedIn = !state.isLoggedIn;
      // If logging out, cancel any editing mode
      if (!state.isLoggedIn) {
        state.isEditing = false;
      }
    },
    resetTimer: (state) => {
      state.info.startTime = new Date().toISOString();
    },
    addUpcomingStep: (state, action: PayloadAction<string>) => {
      state.upcomingSteps.push(action.payload);
    },
    setNewActiveTopic: (state, action: PayloadAction<string>) => {
      if (state.activeStep) {
        state.completedSteps.push(state.activeStep);
      }
      state.activeStep = action.payload;
    },
    completeCurrentStep: (state) => {
      if (state.activeStep) {
        state.completedSteps.push(state.activeStep);
        if (state.upcomingSteps.length > 0) {
          state.activeStep = state.upcomingSteps[0];
          state.upcomingSteps.splice(0, 1);
        } else {
          state.activeStep = "";
        }
      }
    },
    makeStepActive: (state, action: PayloadAction<{step: string, source: 'upcoming' | 'completed'}>) => {
      const { step, source } = action.payload;
      
      if (state.activeStep) {
        state.completedSteps.push(state.activeStep);
      }
      
      state.activeStep = step;
      
      if (source === 'upcoming') {
        state.upcomingSteps = state.upcomingSteps.filter(s => s !== step);
      } else if (source === 'completed') {
        state.completedSteps = state.completedSteps.filter(s => s !== step);
      }
    }
  }
});

export const { 
  setStreamInfo, 
  toggleEditMode, 
  toggleLoggedIn,
  resetTimer, 
  addUpcomingStep, 
  setNewActiveTopic, 
  completeCurrentStep, 
  makeStepActive 
} = streamSlice.actions;

export default streamSlice.reducer;