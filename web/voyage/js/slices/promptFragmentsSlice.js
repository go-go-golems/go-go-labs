import * as reduxjstoolkit from 'https://esm.run/@reduxjs/toolkit';
const { createSlice } = reduxjstoolkit;

const initialState = {
    prompt_fragments: [
        "a majestic lion",
        "in a lush jungle",
        "with vibrant colors",
        "photorealistic style"
    ],
    checked_fragments: [],
    saved_selections: []
};

const promptFragmentsSlice = createSlice({
    name: 'promptFragments',
    initialState,
    reducers: {
        addFragment: (state, action) => {
            state.prompt_fragments.push(action.payload);
        },
        deleteFragment: (state, action) => {
            state.prompt_fragments.splice(action.payload, 1);
            state.checked_fragments = state.checked_fragments
                .filter(index => index !== action.payload)
                .map(index => (index > action.payload ? index - 1 : index));
        },
        toggleCheckedFragment: (state, action) => {
            const index = action.payload;
            if (state.checked_fragments.includes(index)) {
                state.checked_fragments = state.checked_fragments.filter(i => i !== index);
            } else {
                state.checked_fragments.push(index);
            }
        },
        unselectAllFragments: (state) => {
            state.checked_fragments = [];
        },
        saveSelection: (state, action) => {
            state.saved_selections.push(action.payload);
        },
        deleteSavedSelection: (state, action) => {
            state.saved_selections.splice(action.payload, 1);
        },
        replacePromptFragments: (state, action) => {
            return action.payload;
        }
    }
});

export const {
    addFragment,
    deleteFragment,
    toggleCheckedFragment,
    unselectAllFragments,
    saveSelection,
    deleteSavedSelection,
    replacePromptFragments
} = promptFragmentsSlice.actions;

export default promptFragmentsSlice.reducer;