const persistMiddleware = store => next => action => {
    const result = next(action);
    const state = store.getState();
    localStorage.setItem('midjourneyPromptState', JSON.stringify(state));
    return result;
};

export default persistMiddleware;