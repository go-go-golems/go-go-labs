import * as reduxjstoolkit from 'https://esm.run/@reduxjs/toolkit';
const { createSlice } = reduxjstoolkit;

const initialState = {
    aspect_ratio: "16:9",
    model_version: "v5"
};

const optionsSlice = createSlice({
    name: 'options',
    initialState,
    reducers: {
        setAspectRatio: (state, action) => {
            state.aspect_ratio = action.payload;
        },
        setModelVersion: (state, action) => {
            state.model_version = action.payload;
        },
        replaceOptions: (state, action) => {
            return action.payload;
        }
    }
});

export const { setAspectRatio, setModelVersion, replaceOptions } = optionsSlice.actions;

export default optionsSlice.reducer;