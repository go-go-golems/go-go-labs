import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { Widget } from '../../services/widgetsApi';

interface FeaturedWidgetState {
  widget: Widget | null;
  isHighlighted: boolean;
}

const initialState: FeaturedWidgetState = {
  widget: null,
  isHighlighted: false,
};

export const featuredWidgetSlice = createSlice({
  name: 'featuredWidget',
  initialState,
  reducers: {
    setFeaturedWidget: (state, action: PayloadAction<Widget>) => {
      state.widget = action.payload;
    },
    clearFeaturedWidget: (state) => {
      state.widget = null;
    },
    toggleHighlight: (state) => {
      state.isHighlighted = !state.isHighlighted;
    },
    setHighlight: (state, action: PayloadAction<boolean>) => {
      state.isHighlighted = action.payload;
    },
  },
});

export const {
  setFeaturedWidget,
  clearFeaturedWidget,
  toggleHighlight,
  setHighlight,
} = featuredWidgetSlice.actions;

export default featuredWidgetSlice.reducer; 