import { createSlice } from '@reduxjs/toolkit';
import type { PayloadAction } from '@reduxjs/toolkit';
import type { LatLngExpression } from 'leaflet';  // Leaflet's type for [lat, lng] tuple

interface MapState {
  center: LatLngExpression;  // [lat, lng] tuple
  zoom: number;
}

const initialState: MapState = {
  center: [37.7749, -122.4194], // default center (San Francisco city)
  zoom: 12                     // default zoom level
};

const mapSlice = createSlice({
  name: 'map',
  initialState,
  reducers: {
    setView(state, action: PayloadAction<{ center: LatLngExpression; zoom: number }>) {
      state.center = action.payload.center;
      state.zoom = action.payload.zoom;
    }
  }
});

export const { setView } = mapSlice.actions;
export default mapSlice.reducer;