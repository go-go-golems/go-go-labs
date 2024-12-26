document.addEventListener('DOMContentLoaded', function() {
    var downloadBtn = document.getElementById('downloadBtn');
    var downloadMarkdownBtn = document.getElementById('downloadMarkdownBtn');
    var downloadLastArtifactsBtn = document.getElementById('downloadLastArtifactsBtn');
    var statusText = document.getElementById('status');
    var conversationIdSpan = document.getElementById('conversationId');

    // Request the current conversation ID when popup opens
    chrome.runtime.sendMessage({action: "getConversationId"}, function(response) {
        if (response.conversationId) {
            conversationIdSpan.textContent = response.conversationId;
            statusText.textContent = "Conversation active";
            downloadBtn.disabled = false;
            downloadMarkdownBtn.disabled = false;
            downloadLastArtifactsBtn.disabled = false;
        } else {
            conversationIdSpan.textContent = "None";
            statusText.textContent = "No active conversation";
            downloadBtn.disabled = true;
            downloadMarkdownBtn.disabled = true;
            downloadLastArtifactsBtn.disabled = true;
        }
    });

    // Listen for updates to the conversation ID
    chrome.runtime.onMessage.addListener(function(request, sender, sendResponse) {
        if (request.action === "updateConversationId") {
            conversationIdSpan.textContent = request.conversationId;
            statusText.textContent = "Conversation active";
            downloadBtn.disabled = false;
            downloadMarkdownBtn.disabled = false;
            downloadLastArtifactsBtn.disabled = false;
        }
    });

    downloadBtn.addEventListener('click', function() {
        chrome.runtime.sendMessage({action: "downloadConversation"}, function(response) {
            statusText.textContent = response.status;
        });
    });

    downloadMarkdownBtn.addEventListener('click', function() {
        chrome.runtime.sendMessage({action: "downloadMarkdown"}, function(response) {
            statusText.textContent = response.status;
        });
    });

    downloadLastArtifactsBtn.addEventListener('click', function() {
        chrome.runtime.sendMessage({action: "downloadLastArtifacts"}, function(response) {
            statusText.textContent = response.status;
        });
    });
});