import * as reduxjstoolkit from 'https://esm.run/@reduxjs/toolkit';

const { configureStore } = reduxjstoolkit;
import promptFragmentsSlice from './slices/promptFragmentsSlice.js';
import imagesSlice from './slices/imagesSlice.js';
import optionsSlice from './slices/optionsSlice.js';
import promptHistorySlice from './slices/promptHistorySlice.js';
import uiSlice from './slices/uiSlice.js';
import persistMiddleware from './persistMiddleware.js';

function loadState() {
    try {
        const serializedState = localStorage.getItem('midjourneyPromptState');
        if (serializedState === null) {
            return undefined;
        }
        return JSON.parse(serializedState);
    } catch (err) {
        console.error("Failed to load state from localStorage:", err);
        return undefined;
    }
}

const persistedState = loadState();

const store = configureStore({
    reducer: {
        promptFragments: promptFragmentsSlice,
        images: imagesSlice,
        options: optionsSlice,
        promptHistory: promptHistorySlice,
        ui: uiSlice,
    },
    preloadedState: persistedState,
    middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(persistMiddleware),
    devTools: window.__REDUX_DEVTOOLS_EXTENSION__ ? window.__REDUX_DEVTOOLS_EXTENSION__() : false,
});

export default store;