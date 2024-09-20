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