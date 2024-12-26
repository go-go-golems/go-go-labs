// Use the browser API with polyfill fallback to chrome
const browserAPI = typeof browser !== 'undefined' ? browser : chrome;

document.addEventListener('DOMContentLoaded', function() {
    // Get current conversation ID
    browserAPI.runtime.sendMessage({action: "getConversationId"}, function(response) {
        const conversationId = response.conversationId;
        document.getElementById('conversationId').textContent = conversationId || 'No active conversation';
    });

    // Download conversation button
    document.getElementById('downloadConversation').addEventListener('click', function() {
        browserAPI.runtime.sendMessage({action: "downloadConversation"}, function(response) {
            document.getElementById('status').textContent = response.status;
        });
    });

    // Download markdown button
    document.getElementById('downloadMarkdown').addEventListener('click', function() {
        browserAPI.runtime.sendMessage({action: "downloadMarkdown"}, function(response) {
            document.getElementById('status').textContent = response.status;
        });
    });

    // Download last artifacts button
    document.getElementById('downloadLastArtifacts').addEventListener('click', function() {
        browserAPI.runtime.sendMessage({action: "downloadLastArtifacts"}, function(response) {
            document.getElementById('status').textContent = response.status;
        });
    });
});