# Claude Conversation Downloader

A Chrome extension to download Claude AI conversations, including generated artifacts.

## Installation

1. Download the extension files.
2. Open Chrome and go to `chrome://extensions/`.
3. Enable "Developer mode" in the top right.
4. Click "Load unpacked" and select the extension folder.

## Usage

1. Go to [Claude AI](https://claude.ai) and open a conversation.
2. Click the extension icon in the Chrome toolbar.
3. The popup will show the current conversation ID if one is detected.
4. You have three download options:
    - Click "Download Full Conversation" to save the complete conversation data.
    - Click "Download Markdown Only" to save just the markdown summary.
    - Click "Download Last Message Artifacts" to save only the artifacts from the last message.

## What's Downloaded

### Full Conversation
The extension creates a zip file containing:
- JSON file with raw conversation data
- Markdown summary of the conversation
- Individual files for each generated artifact
- An index of all artifacts

### Markdown Only
This option downloads a single markdown file summarizing the conversation.

### Last Message Artifacts
This option downloads a zip file containing only the artifacts from the last message in the conversation.

Find the downloaded files in your Chrome downloads folder.

## Troubleshooting

- If no conversation is detected, refresh the Claude AI page.
- Check the Chrome DevTools console for any error messages.
- If the "Download Last Message Artifacts" button doesn't produce a download, it's possible that the last message doesn't contain any artifacts.

## Creating a Distribution Zip

To create a zip file of the extension for distribution:

1. Ensure you have `make` installed on your system.
2. Open a terminal and navigate to the extension's directory.
3. Run the command:

   ```
   make
   ```

This will create a file named `claude_conversation_downloader.zip` containing all the necessary extension files.

To remove the created zip file, run:

```
make clean
```