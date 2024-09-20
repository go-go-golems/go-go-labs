import * as reduxjstoolkit from 'https://esm.run/@reduxjs/toolkit';
const { createSlice } = reduxjstoolkit;

const initialState = {
    // Define UI-related state if needed
};

const uiSlice = createSlice({
    name: 'ui',
    initialState,
    reducers: {
        // Define UI-related reducers if needed
    }
});

export const {
    // Export UI-related actions if defined
} = uiSlice.actions;

export default uiSlice.reducer;