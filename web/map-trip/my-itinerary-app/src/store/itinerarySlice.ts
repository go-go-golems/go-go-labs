import { createSlice } from '@reduxjs/toolkit';
import type { PayloadAction } from '@reduxjs/toolkit';

interface ItineraryState {
  list: Itinerary[];        // all itineraries loaded
  selectedId: string | null; // id of the currently selected itinerary
}

const initialState: ItineraryState = {
  list: [],
  selectedId: null
};

const itinerarySlice = createSlice({
  name: 'itineraries',
  initialState,
  reducers: {
    setItineraries(state, action: PayloadAction<Itinerary[]>) {
      state.list = action.payload;
      // Auto-select the first itinerary by default, if not already selected
      if (state.list.length > 0 && state.selectedId === null) {
        state.selectedId = state.list[0].id;
      }
    },
    selectItinerary(state, action: PayloadAction<string>) {
      state.selectedId = action.payload;
    }
  }
});

export const { setItineraries, selectItinerary } = itinerarySlice.actions;
export default itinerarySlice.reducer;