import * as reduxjstoolkit from 'https://esm.run/@reduxjs/toolkit';
const { createSlice } = reduxjstoolkit;

const initialState = {
    prompt_history: [
        "a serene lake at sunset",
        "cyberpunk cityscape with neon lights",
        "abstract geometric patterns in pastel colors"
    ],
    search_query: "",
    current_prompt: ""
};

const promptHistorySlice = createSlice({
    name: 'promptHistory',
    initialState,
    reducers: {
        addToHistory: (state, action) => {
            const prompt = action.payload;
            if (prompt && prompt !== state.prompt_history[0]) {
                state.prompt_history.unshift(prompt);
            }
        },
        setSearchQuery: (state, action) => {
            state.search_query = action.payload;
        },
        setCurrentPrompt: (state, action) => {
            state.current_prompt = action.payload;
        },
        replacePromptHistory: (state, action) => {
            return action.payload;
        }
    }
});

export const { addToHistory, setSearchQuery, setCurrentPrompt, replacePromptHistory } = promptHistorySlice.actions;

export default promptHistorySlice.reducer;