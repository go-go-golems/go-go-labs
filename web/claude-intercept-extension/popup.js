// Use the browser API with polyfill fallback to chrome
const browserAPI = typeof browser !== 'undefined' ? browser : chrome;

console.log('[Popup] Script loaded, using:', typeof browser !== 'undefined' ? 'Firefox (browser)' : 'Chrome');

document.addEventListener('DOMContentLoaded', function() {
    console.log('[Popup] DOM loaded, initializing...');
    
    // Get current conversation ID
    browserAPI.runtime.sendMessage({action: "getConversationId"}, function(response) {
        console.log('[Popup] Got conversation ID response:', response);
        const conversationId = response.conversationId;
        document.getElementById('conversationId').textContent = conversationId || 'No active conversation';
    });

    // Download conversation button
    document.getElementById('downloadConversation').addEventListener('click', function() {
        console.log('[Popup] Download conversation clicked');
        browserAPI.runtime.sendMessage({action: "downloadConversation"}, function(response) {
            console.log('[Popup] Download conversation response:', response);
            document.getElementById('status').textContent = response.status;
        });
    });

    // Download markdown button
    document.getElementById('downloadMarkdown').addEventListener('click', function() {
        console.log('[Popup] Download markdown clicked');
        browserAPI.runtime.sendMessage({action: "downloadMarkdown"}, function(response) {
            console.log('[Popup] Download markdown response:', response);
            document.getElementById('status').textContent = response.status;
        });
    });

    // Download last artifacts button
    document.getElementById('downloadLastArtifacts').addEventListener('click', function() {
        console.log('[Popup] Download artifacts clicked');
        browserAPI.runtime.sendMessage({action: "downloadLastArtifacts"}, function(response) {
            console.log('[Popup] Download artifacts response:', response);
            document.getElementById('status').textContent = response.status;
        });
    });
});