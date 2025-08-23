function copyToken() {
    const tokenInput = document.querySelector('input[readonly]');
    if (tokenInput) {
        tokenInput.select();
        tokenInput.setSelectionRange(0, 99999); // For mobile devices
        
        try {
            document.execCommand('copy');
            
            // Show feedback
            const button = document.querySelector('button[onclick="copyToken()"]');
            const originalText = button.textContent;
            button.textContent = 'Copied!';
            button.classList.remove('btn-outline-secondary');
            button.classList.add('btn-success');
            
            setTimeout(() => {
                button.textContent = originalText;
                button.classList.remove('btn-success');
                button.classList.add('btn-outline-secondary');
            }, 2000);
            
        } catch (err) {
            console.error('Failed to copy token:', err);
            alert('Failed to copy token to clipboard');
        }
    }
}
