import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface ScrollState {
  offset: number;
  height: number;
  contentHeight: number;
  isAutoScrollEnabled: boolean;
}

const initialState: ScrollState = {
  offset: 0,
  height: 0,
  contentHeight: 0,
  isAutoScrollEnabled: true,
};

export const scrollSlice = createSlice({
  name: 'scroll',
  initialState,
  reducers: {
    setOffset: (state, action: PayloadAction<number>) => {
      state.offset = action.payload;
      // Disable auto-scroll if user manually scrolls up
      if (action.payload < state.contentHeight - state.height) {
        state.isAutoScrollEnabled = false;
      }
      // Re-enable auto-scroll if user manually scrolls to bottom
      if (action.payload >= state.contentHeight - state.height) {
        state.isAutoScrollEnabled = true;
      }
    },
    setDimensions: (state, action: PayloadAction<{ height: number; contentHeight: number }>) => {
      state.height = action.payload.height;
      state.contentHeight = action.payload.contentHeight;
      // If auto-scroll is enabled, scroll to bottom when content height changes
      if (state.isAutoScrollEnabled) {
        state.offset = Math.max(0, state.contentHeight - state.height);
      }
    },
    scrollToBottom: (state) => {
      state.offset = Math.max(0, state.contentHeight - state.height);
      state.isAutoScrollEnabled = true;
    },
    toggleAutoScroll: (state) => {
      state.isAutoScrollEnabled = !state.isAutoScrollEnabled;
    }
  },
});

export const { setOffset, setDimensions, scrollToBottom, toggleAutoScroll } = scrollSlice.actions;
export default scrollSlice.reducer; 