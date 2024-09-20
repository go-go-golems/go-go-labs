export function escapeRegExp(string) {
    if (typeof string !== 'string') {
        console.warn('escapeRegExp received a non-string input:', string);
        return '';
    }
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export function highlightText(text, query) {
    if (!query) return text;
    const regex = new RegExp(`(${escapeRegExp(query)})`, 'gi');
    return text.replace(regex, '<mark>$1</mark>');
}

export function showConfirmation(message) {
    const confirmation = document.getElementById('confirmation-msg');
    confirmation.textContent = message;
    confirmation.style.display = 'block';
    setTimeout(() => {
        confirmation.style.display = 'none';
    }, 3000);
}

export function parsePromptOptions(prompt) {
    if (typeof prompt !== 'string') {
        console.error('Invalid prompt: expected string, got', typeof prompt);
        return { aspectRatio: null, modelVersion: null, cleanPrompt: '' };
    }

    const options = {
        aspectRatio: null,
        modelVersion: null,
        cleanPrompt: prompt
    };

    try {
        const arMatch = prompt.match(/--ar\s+(\d+:\d+)/i);
        if (arMatch) {
            options.aspectRatio = arMatch[1];
            options.cleanPrompt = options.cleanPrompt.replace(arMatch[0], '').trim();
        }

        const vMatch = prompt.match(/--v\s+(\w+)/i);
        if (vMatch) {
            options.modelVersion = vMatch[1];
            options.cleanPrompt = options.cleanPrompt.replace(vMatch[0], '').trim();
        }
    } catch (error) {
        console.error('Error parsing prompt options:', error);
    }

    return options;
}