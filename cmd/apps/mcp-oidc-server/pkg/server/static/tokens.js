function escapeHtml(text) {
    if (text == null || text == undefined) return '';
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return String(text).replace(/[&<>"']/g, function(m) { return map[m]; });
}

function formatDate(dateStr) {
    if (!dateStr) return 'N/A';
    try {
        return new Date(dateStr).toLocaleString();
    } catch (e) {
        return dateStr;
    }
}

async function refreshTokens() {
    const loading = document.getElementById('loading');
    const table = document.getElementById('tokenTable');
    const noTokens = document.getElementById('noTokens');
    const tbody = document.getElementById('tokenTableBody');
    
    loading.style.display = 'block';
    table.style.display = 'none';
    noTokens.style.display = 'none';
    
    try {
        const response = await fetch('/api/tokens');
        const data = await response.json();
        
        // Handle error responses
        if (!response.ok || data.error) {
            throw new Error(data.error || `HTTP ${response.status}: ${response.statusText}`);
        }
        
        const tokens = Array.isArray(data) ? data : [];
        tbody.innerHTML = '';
        
        if (tokens.length === 0) {
            loading.style.display = 'none';
            noTokens.style.display = 'block';
            return;
        }
        
        tokens.forEach(token => {
            const row = document.createElement('tr');
            // Handle both lowercase and capitalized field names from Go struct
            const tokenValue = token.token || token.Token || '';
            const clientId = token.client_id || token.ClientID || '';
            const scopes = token.scopes || token.Scopes || [];
            const expiresAt = token.expires_at || token.ExpiresAt || '';
            const scopesStr = Array.isArray(scopes) ? scopes.join(', ') : String(scopes || '');
            
            row.innerHTML = `
                <td><code class="text-break">${escapeHtml(tokenValue)}</code></td>
                <td>${escapeHtml(clientId)}</td>
                <td>${escapeHtml(scopesStr)}</td>
                <td>${formatDate(expiresAt)}</td>
                <td>
                    <button class="btn btn-sm btn-outline-secondary me-2" data-token="${escapeHtml(tokenValue)}" onclick="copyTokenValue(this)">Copy</button>
                    <button class="btn btn-sm btn-outline-danger" onclick="deleteToken('${escapeHtml(tokenValue)}')">
                        Delete
                    </button>
                </td>
            `;
            tbody.appendChild(row);
        });
        
        loading.style.display = 'none';
        table.style.display = 'block';
        
    } catch (error) {
        console.error('Error fetching tokens:', error);
        loading.style.display = 'none';
        tbody.innerHTML = '<tr><td colspan="5" class="text-center text-danger">Error loading tokens</td></tr>';
        table.style.display = 'block';
    }
}

async function createToken() {
    const createBtn = document.getElementById('createBtn');
    const originalText = createBtn.textContent;
    
    createBtn.disabled = true;
    createBtn.textContent = 'Creating...';
    
    try {
        const clientId = document.getElementById('client_id').value.trim() || 'manual-client';
        const scopesStr = document.getElementById('scopes').value.trim() || 'openid,profile';
        const ttl = document.getElementById('ttl').value.trim() || '24h';
        
        const scopes = scopesStr.split(',').map(s => s.trim()).filter(Boolean);
        
        const response = await fetch('/api/tokens', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                client_id: clientId,
                scopes: scopes,
                ttl: ttl
            })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        await refreshTokens();
        
        // Reset form
        document.getElementById('client_id').value = 'manual-client';
        document.getElementById('scopes').value = 'openid,profile';
        document.getElementById('ttl').value = '24h';
        
    } catch (error) {
        console.error('Error creating token:', error);
        alert('Failed to create token: ' + error.message);
    } finally {
        createBtn.disabled = false;
        createBtn.textContent = originalText;
    }
}

async function deleteToken(token) {
    if (!confirm('Are you sure you want to delete this token?')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/tokens?token=${encodeURIComponent(token)}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        await refreshTokens();
        
    } catch (error) {
        console.error('Error deleting token:', error);
        alert('Failed to delete token: ' + error.message);
    }
}

// Initialize page
document.addEventListener('DOMContentLoaded', function() {
    refreshTokens();
    
    document.getElementById('createBtn').addEventListener('click', createToken);
    
    // Allow Enter key to create token
    document.getElementById('tokenForm').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            createToken();
        }
    });
});

async function copyTokenValue(btn){
    try{
        const token = btn.getAttribute('data-token') || '';
        await navigator.clipboard.writeText(token);
        const old = btn.textContent;
        btn.textContent = 'Copied!';
        btn.classList.remove('btn-outline-secondary');
        btn.classList.add('btn-success');
        setTimeout(()=>{ btn.textContent = old; btn.classList.remove('btn-success'); btn.classList.add('btn-outline-secondary'); }, 1200);
    }catch(err){
        console.error('Copy failed', err);
        alert('Failed to copy token');
    }
}
