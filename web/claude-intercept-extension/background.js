// Dynamically load dependencies
async function loadDependencies() {
    // Load polyfill only in Firefox
    if (typeof browser === 'undefined') {
        console.log('[Background] Loading polyfill');
        await new Promise((resolve) => {
            console.log('[Background] Creating polyfill element');
            const script = document.createElement('script');
            script.src = 'browser-polyfill.min.js';
            script.onload = resolve;
            console.log('[Background] polyfill element created');
            document.head.appendChild(script);
        });
    }

    // Load JSZip
    await new Promise((resolve) => {
        console.log('[Background] Creating jszip element');
        const script = document.createElement('script');
        script.src = 'jszip.min.js';
        script.onload = resolve;
        console.log('[Background] jszip element created');
        document.head.appendChild(script);
    });
}

// Initialize after loading dependencies
loadDependencies().then(() => {
    // Use the browser API with polyfill fallback to chrome
    const browserAPI = typeof browser !== 'undefined' ? browser : chrome;

    console.log('[Background] Script loaded, using:', typeof browser !== 'undefined' ? 'Firefox (browser)' : 'Chrome');

    let currentConversationId = null;
    let organizationId = null;

    // Set to store URLs of requests we've already processed
    const processedRequests = new Set();

    // Regular expression for matching URLs
    const urlRegex = /\/api\/organizations\/([^/]+)\/chat_conversations\/([^/]+)/;

    browserAPI.webRequest.onBeforeRequest.addListener(
        function (details) {
            console.log('[Request Intercepted]', details.url);
            const match = details.url.match(urlRegex);
            if (match) {
                organizationId = match[1];
                currentConversationId = match[2];
                console.log('[Conversation Updated] ID:', currentConversationId, 'Org:', organizationId);
            }
            return {};
        },
        {urls: ["<all_urls>"]},
        ["blocking"]
    );

    browserAPI.runtime.onMessage.addListener(
        function(request, sender, sendResponse) {
            console.log('[Message Received]', request.action);
            
            if (request.action === "downloadConversation") {
                if (currentConversationId && organizationId) {
                    console.log('[Download] Starting conversation download', {currentConversationId, organizationId});
                    downloadConversation(organizationId, currentConversationId);
                    sendResponse({status: "Downloading conversation"});
                } else {
                    console.warn('[Download] No active conversation');
                    sendResponse({status: "No active conversation"});
                }
            } else if (request.action === "downloadMarkdown") {
                if (currentConversationId && organizationId) {
                    console.log('[Download] Starting markdown download', {currentConversationId, organizationId});
                    downloadMarkdown(organizationId, currentConversationId);
                    sendResponse({status: "Downloading markdown"});
                } else {
                    console.warn('[Download] No active conversation for markdown');
                    sendResponse({status: "No active conversation"});
                }
            } else if (request.action === "downloadLastArtifacts") {
                if (currentConversationId && organizationId) {
                    console.log('[Download] Starting artifacts download', {currentConversationId, organizationId});
                    downloadLastArtifacts(organizationId, currentConversationId);
                    sendResponse({status: "Downloading last message artifacts"});
                } else {
                    console.warn('[Download] No active conversation for artifacts');
                    sendResponse({status: "No active conversation"});
                }
            } else if (request.action === "getConversationId") {
                console.log('[Query] Conversation ID requested:', currentConversationId);
                sendResponse({conversationId: currentConversationId});
            }
            return true;
        }
    );

    function downloadMarkdown(orgId, convId) {
        console.log('[Markdown] Starting download process', {orgId, convId});
        const url = `https://claude.ai/api/organizations/${orgId}/chat_conversations/${convId}?tree=True&rendering_mode=raw`;

        fetch(url)
            .then(response => {
                console.log('[Markdown] Fetch response status:', response.status);
                return response.json();
            })
            .then(data => {
                console.log('[Markdown] Data received, generating content');
                const markdownContent = generateMarkdown(data);
                const blob = new Blob([markdownContent], {type: 'text/markdown'});
                const url = URL.createObjectURL(blob);
                browserAPI.downloads.download({
                    url: url,
                    filename: `conversation_${data.uuid}.md`,
                    saveAs: false
                }, function(downloadId) {
                    if (browserAPI.runtime.lastError) {
                        console.error('[Markdown] Download failed:', browserAPI.runtime.lastError);
                    } else {
                        console.log('[Markdown] File saved with ID:', downloadId);
                    }
                    URL.revokeObjectURL(url);
                });
            })
            .catch(error => console.error('[Markdown] Error:', error));
    }

    function downloadLastArtifacts(orgId, convId) {
        console.log('[Artifacts] Starting download process', {orgId, convId});
        const url = `https://claude.ai/api/organizations/${orgId}/chat_conversations/${convId}?tree=True&rendering_mode=raw`;

        fetch(url)
            .then(response => {
                console.log('[Artifacts] Fetch response status:', response.status);
                return response.json();
            })
            .then(data => {
                console.log('[Artifacts] Data received, processing last message');
                const lastMessage = data.chat_messages[data.chat_messages.length - 1];
                const artifacts = extractArtifacts([lastMessage]);

                if (artifacts.length === 0) {
                    console.warn('[Artifacts] No artifacts found in the last message');
                    return;
                }

                console.log('[Artifacts] Found artifacts:', artifacts.length);
                const zip = new JSZip();
                artifacts.forEach((artifact) => {
                    console.log('[Artifacts] Adding to zip:', artifact.filename);
                    zip.file(artifact.filename, artifact.content);
                });

                zip.generateAsync({type:"blob"})
                    .then(function(content) {
                        console.log('[Artifacts] Zip generated, starting download');
                        const url = URL.createObjectURL(content);
                        browserAPI.downloads.download({
                            url: url,
                            filename: `last_message_artifacts_${data.uuid}.zip`,
                            saveAs: false
                        }, function(downloadId) {
                            if (browserAPI.runtime.lastError) {
                                console.error('[Artifacts] Download failed:', browserAPI.runtime.lastError);
                            } else {
                                console.log('[Artifacts] Zip saved with ID:', downloadId);
                            }
                            URL.revokeObjectURL(url);
                        });
                    });
            })
            .catch(error => console.error('[Artifacts] Error:', error));
    }

    function downloadConversation(orgId, convId) {
        console.log('[Conversation] Starting download process', {orgId, convId});
        const url = `https://claude.ai/api/organizations/${orgId}/chat_conversations/${convId}?tree=True&rendering_mode=raw`;

        fetch(url)
            .then(response => {
                console.log('[Conversation] Fetch response status:', response.status);
                return response.json();
            })
            .then(data => {
                console.log('[Conversation] Data received, processing');
                processConversation(data);
            })
            .catch(error => console.error('[Conversation] Error:', error));
    }

    function processConversation(data) {
        console.log('[Process] Starting conversation processing');
        const zip = new JSZip();

        console.log('[Process] Adding JSON file');
        zip.file(`conversation_${data.uuid}.json`, JSON.stringify(data, null, 2));

        console.log('[Process] Generating and adding markdown');
        const markdownContent = generateMarkdown(data);
        zip.file(`conversation_${data.uuid}.md`, markdownContent);

        console.log('[Process] Extracting artifacts');
        const artifacts = extractArtifacts(data);
        artifacts.forEach((artifact) => {
            console.log('[Process] Adding artifact:', artifact.filename);
            zip.file(artifact.filename, artifact.content);
        });

        console.log('[Process] Generating final zip');
        zip.generateAsync({type:"blob"})
            .then(function(content) {
                console.log('[Process] Zip generated, starting download');
                const url = URL.createObjectURL(content);
                browserAPI.downloads.download({
                    url: url,
                    filename: `conversation_${data.uuid}.zip`,
                    saveAs: false
                }, function(downloadId) {
                    if (browserAPI.runtime.lastError) {
                        console.error('[Process] Download failed:', browserAPI.runtime.lastError);
                    } else {
                        console.log('[Process] Zip saved with ID:', downloadId);
                    }
                    URL.revokeObjectURL(url);
                });
            });
    }

    const artifactRegex = /<antArtifact([^>]*)>([\s\S]*?)<\/antArtifact>/g;

    function extractArtifacts(parsedResponse) {
        const artifacts = [];
        let artifactIndex = 1;
        let indexContent = "# Artifact Index\n\n";

        parsedResponse.chat_messages.forEach((message, messageIndex) => {
            let match;
            while ((match = artifactRegex.exec(message.text)) !== null) {
                const fullTag = match[1];
                const content = match[2];

                const identifierMatch = fullTag.match(/identifier="([^"]+)"/);
                const titleMatch = fullTag.match(/title="([^"]+)"/);
                const typeMatch = fullTag.match(/type="([^"]+)"/);

                const identifier = identifierMatch ? identifierMatch[1] : `artifact_${messageIndex + 1}`;
                const title = titleMatch ? titleMatch[1] : 'Untitled';
                const type = typeMatch ? typeMatch[1] : 'text/plain';

                let extension;
                switch (type) {
                    case 'application/vnd.ant.code':
                        extension = 'js';  // Assuming JavaScript, adjust as needed
                        break;
                    case 'text/markdown':
                        extension = 'md';
                        break;
                    case 'text/html':
                        extension = 'html';
                        break;
                    case 'image/svg+xml':
                        extension = 'svg';
                        break;
                    default:
                        extension = 'txt';
                }

                const filename = `${artifactIndex.toString().padStart(2, '0')}_${identifier}_${title.replace(/[^a-z0-9]/gi, '_').toLowerCase()}.${extension}`;

                artifacts.push({filename, content});

                // Add entry to index
                indexContent += `## ${artifactIndex}. ${title}\n\n`;
                indexContent += `Filename: ${filename}\n\n`;

                artifactIndex++;
            }
        });

        // Add the index file to the artifacts
        artifacts.push({
            filename: 'artifact-index.md',
            content: indexContent
        });

        return artifacts;
    }

    function generateMarkdown(parsedResponse) {
        let markdown = `# ${parsedResponse.name}\n\n`;
        markdown += `Project: ${parsedResponse.project?.name || 'Unnamed Project'}\n\n`;
        markdown += `Created: ${parsedResponse.created_at}\n`;
        markdown += `Updated: ${parsedResponse.updated_at}\n\n`;
        markdown += `## Messages\n\n`;

        parsedResponse.chat_messages.forEach((message, index) => {
            markdown += `### Message ${index + 1}\n\n`;
            markdown += `Sender: ${message.sender}\n`;
            markdown += `Created: ${message.created_at}\n\n`;
            markdown += `${message.text}\n\n`;
        });

        return markdown;
    }
});