import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Agent } from '@/types/api';

interface UiState {
  selectedAgent: Agent | null;
  activeTab: 'fleet' | 'updates' | 'tasks';
  connectionStatus: 'connected' | 'disconnected' | 'reconnecting';
  loading: boolean;
  error: string | null;
  agentDetailModalVisible: boolean;
}

const initialState: UiState = {
  selectedAgent: null,
  activeTab: 'fleet',
  connectionStatus: 'disconnected',
  loading: false,
  error: null,
  agentDetailModalVisible: false,
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    setSelectedAgent: (state, action: PayloadAction<Agent | null>) => {
      state.selectedAgent = action.payload;
    },
    setActiveTab: (state, action: PayloadAction<'fleet' | 'updates' | 'tasks'>) => {
      state.activeTab = action.payload;
    },
    setConnectionStatus: (state, action: PayloadAction<'connected' | 'disconnected' | 'reconnecting'>) => {
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
  },
});

export const {
  setSelectedAgent,
  setActiveTab,
  setConnectionStatus,
  setLoading,
  setError,
  setAgentDetailModalVisible,
} = uiSlice.actions;

export default uiSlice.reducer;
