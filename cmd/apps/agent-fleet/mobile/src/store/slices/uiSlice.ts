import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { Agent } from "@/types/api";

export interface Notification {
  id: string;
  type: "question" | "warning" | "error" | "info";
  agentId: string;
  agentName: string;
  title: string;
  message: string;
  timestamp: string;
  dismissed: boolean;
}

interface UiState {
  selectedAgent: Agent | null;
  activeTab: "fleet" | "updates" | "tasks";
  connectionStatus: "connected" | "disconnected" | "reconnecting";
  loading: boolean;
  error: string | null;
  agentDetailModalVisible: boolean;
  notifications: Notification[];
  feedbackModalVisible: boolean;
  feedbackTargetAgent: Agent | null;
}

const initialState: UiState = {
  selectedAgent: null,
  activeTab: "fleet",
  connectionStatus: "disconnected",
  loading: false,
  error: null,
  agentDetailModalVisible: false,
  notifications: [],
  feedbackModalVisible: false,
  feedbackTargetAgent: null,
};

const uiSlice = createSlice({
  name: "ui",
  initialState,
  reducers: {
    setSelectedAgent: (state, action: PayloadAction<Agent | null>) => {
      state.selectedAgent = action.payload;
    },
    setActiveTab: (
      state,
      action: PayloadAction<"fleet" | "updates" | "tasks">
    ) => {
      state.activeTab = action.payload;
    },
    setConnectionStatus: (
      state,
      action: PayloadAction<"connected" | "disconnected" | "reconnecting">
    ) => {
      state.connectionStatus = action.payload;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    setAgentDetailModalVisible: (state, action: PayloadAction<boolean>) => {
      state.agentDetailModalVisible = action.payload;
    },
    addNotification: (state, action: PayloadAction<Notification>) => {
      // Add to beginning so newest are first
      state.notifications.unshift(action.payload);
      // Keep only last 20 notifications
      if (state.notifications.length > 20) {
        state.notifications = state.notifications.slice(0, 20);
      }
    },
    dismissNotification: (state, action: PayloadAction<string>) => {
      const notification = state.notifications.find(
        (n) => n.id === action.payload
      );
      if (notification) {
        notification.dismissed = true;
      }
    },
    removeNotification: (state, action: PayloadAction<string>) => {
      state.notifications = state.notifications.filter(
        (n) => n.id !== action.payload
      );
    },
    clearNotifications: (state) => {
      state.notifications = [];
    },
    setFeedbackModalVisible: (state, action: PayloadAction<boolean>) => {
      state.feedbackModalVisible = action.payload;
    },
    setFeedbackTargetAgent: (state, action: PayloadAction<Agent | null>) => {
      state.feedbackTargetAgent = action.payload;
    },
  },
});

export const {
  setSelectedAgent,
  setActiveTab,
  setConnectionStatus,
  setLoading,
  setError,
  setAgentDetailModalVisible,
  addNotification,
  dismissNotification,
  removeNotification,
  clearNotifications,
  setFeedbackModalVisible,
  setFeedbackTargetAgent,
} = uiSlice.actions;

export default uiSlice.reducer;
