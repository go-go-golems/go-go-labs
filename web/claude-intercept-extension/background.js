let currentConversationId = null;
let organizationId = null;

// Set to store URLs of requests we've already processed
const processedRequests = new Set();

// Regular expression for matching URLs
const urlRegex = /\/api\/organizations\/([^/]+)\/chat_conversations\/([^/]+)/;

chrome.webRequest.onBeforeRequest.addListener(
    function (details) {
        const match = details.url.match(urlRegex);
        if (match) {
            organizationId = match[1];
            currentConversationId = match[2];
            console.log('Current conversation ID:', currentConversationId);
        }
        return {};
    },
    {urls: ["<all_urls>"]},
    ["blocking"]
);

chrome.runtime.onMessage.addListener(
    function(request, sender, sendResponse) {
        if (request.action === "downloadConversation") {
            if (currentConversationId && organizationId) {
                downloadConversation(organizationId, currentConversationId);
                sendResponse({status: "Downloading conversation"});
            } else {
                sendResponse({status: "No active conversation"});
            }
        } else if (request.action === "downloadMarkdown") {
            if (currentConversationId && organizationId) {
                downloadMarkdown(organizationId, currentConversationId);
                sendResponse({status: "Downloading markdown"});
            } else {
                sendResponse({status: "No active conversation"});
            }
        } else if (request.action === "downloadLastArtifacts") {
            if (currentConversationId && organizationId) {
                downloadLastArtifacts(organizationId, currentConversationId);
                sendResponse({status: "Downloading last message artifacts"});
            } else {
                sendResponse({status: "No active conversation"});
            }
        } else if (request.action === "getConversationId") {
            sendResponse({conversationId: currentConversationId});
        }
        return true;
    }
);

function downloadMarkdown(orgId, convId) {
    const url = `https://claude.ai/api/organizations/${orgId}/chat_conversations/${convId}?tree=True&rendering_mode=raw`;

    fetch(url)
        .then(response => response.json())
        .then(data => {
            const markdownContent = generateMarkdown(data);
            const blob = new Blob([markdownContent], {type: 'text/markdown'});
            const url = URL.createObjectURL(blob);
            chrome.downloads.download({
                url: url,
                filename: `conversation_${data.uuid}.md`,
                saveAs: false
            }, function(downloadId) {
                if (chrome.runtime.lastError) {
                    console.error('Download failed:', chrome.runtime.lastError);
                } else {
                    console.log('Markdown file saved with ID:', downloadId);
                }
                URL.revokeObjectURL(url);
            });
        })
        .catch(error => console.error('Error:', error));
}


function downloadLastArtifacts(orgId, convId) {
    const url = `https://claude.ai/api/organizations/${orgId}/chat_conversations/${convId}?tree=True&rendering_mode=raw`;

    fetch(url)
        .then(response => response.json())
        .then(data => {
            const lastMessage = data.chat_messages[data.chat_messages.length - 1];
            const artifacts = extractArtifacts([lastMessage]);

            if (artifacts.length === 0) {
                console.log('No artifacts found in the last message');
                return;
            }

            const zip = new JSZip();
            artifacts.forEach((artifact) => {
                zip.file(artifact.filename, artifact.content);
            });

            zip.generateAsync({type:"blob"})
                .then(function(content) {
                    const url = URL.createObjectURL(content);
                    chrome.downloads.download({
                        url: url,
                        filename: `last_message_artifacts_${data.uuid}.zip`,
                        saveAs: false
                    }, function(downloadId) {
                        if (chrome.runtime.lastError) {
                            console.error('Download failed:', chrome.runtime.lastError);
                        } else {
                            console.log('Last message artifacts zip saved with ID:', downloadId);
                        }
                        URL.revokeObjectURL(url);
                    });
                });
        })
        .catch(error => console.error('Error:', error));
}


function downloadConversation(orgId, convId) {
    const url = `https://claude.ai/api/organizations/${orgId}/chat_conversations/${convId}?tree=True&rendering_mode=raw`;

    fetch(url)
        .then(response => response.json())
        .then(data => processConversation(data))
        .catch(error => console.error('Error:', error));
}

function processConversation(data) {
    // Create a new JSZip instance
    const zip = new JSZip();

    // Add JSON file to zip
    zip.file(`conversation_${data.uuid}.json`, JSON.stringify(data, null, 2));

    // Generate and add markdown file to zip
    const markdownContent = generateMarkdown(data);
    zip.file(`conversation_${data.uuid}.md`, markdownContent);

    // Extract and add artifacts to zip
    const artifacts = extractArtifacts(data);
    artifacts.forEach((artifact) => {
        zip.file(artifact.filename, artifact.content);
    });

    // Generate zip file
    zip.generateAsync({type:"blob"})
        .then(function(content) {
            // Use chrome.downloads.download() to save the zip file
            const url = URL.createObjectURL(content);
            chrome.downloads.download({
                url: url,
                filename: `conversation_${data.uuid}.zip`,
                saveAs: false
            }, function(downloadId) {
                if (chrome.runtime.lastError) {
                    console.error('Download failed:', chrome.runtime.lastError);
                } else {
                    console.log('Zip file saved with ID:', downloadId);
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