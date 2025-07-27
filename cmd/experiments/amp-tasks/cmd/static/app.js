// amp-tasks Web Interface JavaScript

// Global variables
let isRefreshing = false;

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    console.log('amp-tasks web interface loaded');
    
    // Set up tooltips
    initializeTooltips();
    
    // Set up refresh functionality
    setupRefreshHandlers();
    
    // Set up keyboard shortcuts
    setupKeyboardShortcuts();
});

// Initialize Bootstrap tooltips
function initializeTooltips() {
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function(tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
}

// Set up refresh handlers
function setupRefreshHandlers() {
    // Refresh button click handler
    const refreshButtons = document.querySelectorAll('[onclick="refreshContent()"]');
    refreshButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            refreshContent();
        });
    });
}

// Set up keyboard shortcuts
function setupKeyboardShortcuts() {
    document.addEventListener('keydown', function(e) {
        // Ctrl+R or Cmd+R for refresh (prevent default browser refresh)
        if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
            e.preventDefault();
            refreshContent();
        }
        
        // F5 for refresh
        if (e.key === 'F5') {
            e.preventDefault();
            refreshContent();
        }
        
        // D key for dashboard
        if (e.key === 'd' && !e.ctrlKey && !e.metaKey && !isInputFocused()) {
            window.location.href = '/dashboard';
        }
        
        // R key for report
        if (e.key === 'r' && !e.ctrlKey && !e.metaKey && !isInputFocused()) {
            window.location.href = '/report';
        }
    });
}

// Check if an input element is focused
function isInputFocused() {
    const activeElement = document.activeElement;
    return activeElement && (
        activeElement.tagName === 'INPUT' ||
        activeElement.tagName === 'TEXTAREA' ||
        activeElement.contentEditable === 'true'
    );
}

// Main refresh function
function refreshContent() {
    if (isRefreshing) {
        console.log('Refresh already in progress, skipping...');
        return;
    }
    
    isRefreshing = true;
    console.log('Refreshing content...');
    
    // Show loading state
    showLoadingState();
    
    // Determine current page and refresh accordingly
    const currentPath = window.location.pathname;
    let endpoint = '/api/dashboard';
    let contentId = 'dashboard-content';
    
    if (currentPath.includes('/report')) {
        endpoint = '/api/report';
        contentId = 'report-content';
        
        // Include query parameters for report
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.toString()) {
            endpoint += '?' + urlParams.toString();
        }
    }
    
    // Fetch updated content
    fetch(endpoint)
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.text();
        })
        .then(html => {
            // Update content
            const contentElement = document.getElementById(contentId);
            if (contentElement) {
                contentElement.innerHTML = html;
                console.log('Content refreshed successfully');
                
                // Reinitialize tooltips for new content
                initializeTooltips();
                
                // Update last updated time
                updateLastUpdatedTime();
                
                // Show success feedback
                showRefreshSuccess();
            } else {
                console.error('Content element not found:', contentId);
            }
        })
        .catch(error => {
            console.error('Error refreshing content:', error);
            showRefreshError(error.message);
        })
        .finally(() => {
            hideLoadingState();
            isRefreshing = false;
        });
}

// Show loading state
function showLoadingState() {
    const refreshButtons = document.querySelectorAll('[onclick="refreshContent()"]');
    refreshButtons.forEach(button => {
        button.classList.add('loading');
        button.disabled = true;
        
        // Add spinner to button
        const icon = button.querySelector('i');
        if (icon) {
            icon.className = 'bi bi-arrow-clockwise';
            icon.style.animation = 'spin 1s linear infinite';
        }
    });
    
    // Update refresh status badge
    const refreshStatus = document.getElementById('refresh-status');
    if (refreshStatus) {
        refreshStatus.innerHTML = '<i class="bi bi-arrow-clockwise"></i> Refreshing...';
        refreshStatus.className = 'badge bg-warning';
    }
}

// Hide loading state
function hideLoadingState() {
    const refreshButtons = document.querySelectorAll('[onclick="refreshContent()"]');
    refreshButtons.forEach(button => {
        button.classList.remove('loading');
        button.disabled = false;
        
        // Remove spinner from button
        const icon = button.querySelector('i');
        if (icon) {
            icon.style.animation = '';
        }
    });
}

// Update last updated time
function updateLastUpdatedTime() {
    const lastUpdatedElement = document.getElementById('last-updated');
    if (lastUpdatedElement) {
        lastUpdatedElement.textContent = new Date().toLocaleTimeString();
    }
}

// Show refresh success feedback
function showRefreshSuccess() {
    const refreshStatus = document.getElementById('refresh-status');
    if (refreshStatus) {
        const originalContent = refreshStatus.innerHTML;
        const originalClass = refreshStatus.className;
        
        refreshStatus.innerHTML = '<i class="bi bi-check-circle"></i> Updated';
        refreshStatus.className = 'badge bg-success';
        
        // Restore original content after 2 seconds
        setTimeout(() => {
            refreshStatus.innerHTML = originalContent;
            refreshStatus.className = originalClass;
        }, 2000);
    }
}

// Show refresh error feedback
function showRefreshError(message) {
    const refreshStatus = document.getElementById('refresh-status');
    if (refreshStatus) {
        const originalContent = refreshStatus.innerHTML;
        const originalClass = refreshStatus.className;
        
        refreshStatus.innerHTML = '<i class="bi bi-exclamation-triangle"></i> Error';
        refreshStatus.className = 'badge bg-danger';
        
        // Restore original content after 5 seconds
        setTimeout(() => {
            refreshStatus.innerHTML = originalContent;
            refreshStatus.className = originalClass;
        }, 5000);
    }
    
    // Show toast notification
    showToast('Refresh Error', message, 'error');
}

// Show toast notification
function showToast(title, message, type = 'info') {
    // Create toast element
    const toastId = 'toast-' + Date.now();
    const toastHtml = `
        <div id="${toastId}" class="toast" role="alert" aria-live="assertive" aria-atomic="true">
            <div class="toast-header">
                <i class="bi bi-${getToastIcon(type)} me-2"></i>
                <strong class="me-auto">${title}</strong>
                <button type="button" class="btn-close" data-bs-dismiss="toast" aria-label="Close"></button>
            </div>
            <div class="toast-body">
                ${message}
            </div>
        </div>
    `;
    
    // Get or create toast container
    let toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toast-container';
        toastContainer.className = 'toast-container position-fixed top-0 end-0 p-3';
        toastContainer.style.zIndex = '1055';
        document.body.appendChild(toastContainer);
    }
    
    // Add toast to container
    toastContainer.insertAdjacentHTML('beforeend', toastHtml);
    
    // Initialize and show toast
    const toastElement = document.getElementById(toastId);
    const toast = new bootstrap.Toast(toastElement);
    toast.show();
    
    // Clean up after toast is hidden
    toastElement.addEventListener('hidden.bs.toast', function() {
        toastElement.remove();
    });
}

// Get icon for toast type
function getToastIcon(type) {
    switch (type) {
        case 'success':
            return 'check-circle';
        case 'error':
            return 'exclamation-triangle';
        case 'warning':
            return 'exclamation-circle';
        default:
            return 'info-circle';
    }
}

// Utility function to format time ago
function formatTimeAgo(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now - date;
    
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (seconds < 60) {
        return 'just now';
    } else if (minutes < 60) {
        return `${minutes}m ago`;
    } else if (hours < 24) {
        return `${hours}h ago`;
    } else if (days === 1) {
        return '1 day ago';
    } else {
        return `${days} days ago`;
    }
}

// Animation helpers
function animateElement(element, animation) {
    element.style.animation = animation;
    element.addEventListener('animationend', function() {
        element.style.animation = '';
    }, { once: true });
}

// Export functions for global access
window.refreshContent = refreshContent;
window.showToast = showToast;
window.formatTimeAgo = formatTimeAgo;
