# Scroll Tweet Saver – Full Build Guide (Updated)

> **Goal**: Build a Firefox WebExtension that automatically captures every tweet (now _post_) you scroll past on `twitter.com`/`x.com`, stores the structured data in `browser.storage.local`, and lets you view stats and export the archive as JSON from a tiny toolbar popup.

---

## 1. Prerequisites & Tooling

Before you begin, you'll need a few tools installed. These are minimal and don't require any complex build pipeline.

### Required Tools:

- **Firefox Developer Edition**

  - Offers live extension reload, console tools, and access to upcoming APIs.
  - [Download here](https://www.mozilla.org/firefox/developer/)

- **Node.js (version ≥ 18)**

  - Provides access to the `web-ext` CLI utility, which simplifies development.
  - [Download Node.js](https://nodejs.org/)

- **web-ext CLI**
  - Firefox's official command-line tool for developing WebExtensions.
  - Install via npm:
    ```bash
    npm install -g web-ext
    ```
  - [web-ext documentation](https://developer.mozilla.org/en-US/Add-ons/WebExtensions/Getting_started_with_web-ext)

These tools help create a fast feedback loop while building the extension and allow you to test features instantly.

---

## 2. Architecture Overview (Revised for Export Fix)

Here's how the system is structured, including the updated export flow:

```
┌─────────────────────────────┐ 1  DOM tweets
│  twitter.com (page)         │◀───────────────┐
└─────────────────────────────┘               │
       ▲  injects                            │ MutationObserver
       │ content‑script.js                   │
┌──────┴──────────────────────┐              ▼ sendMessage("SAVE_TWEET")
│  Scroll Tweet Saver         │      ┌────────────────────────┐
│  content‑script             │────▶│  background.js          │
└─────────────────────────────┘ 2    │  (persistent page)     │◀ ─ ─ ─ ┐ TWEET_SAVED msg
                 ▲ storage.get/set() │  🔑 owns storage quota  │        │ (on save)
                 │                   │  + Handles Export Logic │        │
                 │                   └───────────┬────────────┘        │
                 │                               │                         │
                 ▼                               │ runtime API             │
        popup/popup.html/js                      │                         │
        (toolbar action) 3  GET_TWEETS       ◀───┘                         │
                           EXPORT_TWEETS    ───▶                           │
                           ◀─────────────────────┘ listens for TWEET_SAVED
```

### Roles:

1.  **Content Script (`content-script.js`)**: Injected into Twitter/X tabs. Uses `MutationObserver` to detect new tweet elements (`article[data-testid="tweet"]`). Extracts structured data (`extractTweet`), deduplicates using `window.__seenTweets`, and sends `"SAVE_TWEET"` messages to the background script.
2.  **Background Script (`background.js`)**: The central coordinator and data handler.
    - **Storage**: Manages the tweet list in `browser.storage.local`, enforcing `TWEET_LIMIT`.
    - **Message Handling**: Listens for messages:
      - `"SAVE_TWEET"` (from Content Script): Saves the tweet data if not a duplicate. On success, sends `"TWEET_SAVED"` to potentially update the popup.
      - `"GET_TWEETS"` (from Popup): Retrieves the current list of saved tweets from storage and sends it back.
      - `"EXPORT_TWEETS"` (from Popup): **Handles the entire export process:** Fetches tweets, creates a `Blob` and `ObjectURL` _within the background context_, initiates the download using `browser.downloads.download`, and manages the `ObjectURL` lifecycle using `browser.downloads.onChanged` to prevent memory leaks.
3.  **Popup UI (`popup/popup.html`, `popup/popup.js`)**: The user interface in the browser toolbar.
    - **Display**: Shows statistics (total tweets, estimated export size) and a list of recently saved tweets upon opening (by sending `"GET_TWEETS"`).
    - **Auto-Refresh**: Listens for `"TWEET_SAVED"` messages from the background script and refreshes the displayed list and stats.
    - **Export Trigger**: Contains the "Export JSON" button. When clicked, it _only_ sends an `"EXPORT_TWEETS"` message to the background script, delegating the actual export work.

This separation ensures that UI, scraping, and data logic remain modular and easy to maintain.

---

## 3. Project Directory Structure

Create your extension in a directory like this:

```
tweet-saver/
├── manifest.json
├── background.js
├── content-script.js
├── icons/
│   ├── icon48.png
│   └── icon128.png
└── popup/
    ├── popup.html
    └── popup.js
```

You can use any 48x48 and 128x128 PNG icons for now (or download placeholders from sites like [iconify.design](https://iconify.design)).

Version control is strongly recommended — initialize a Git repo to track changes:

```bash
git init
git add .
git commit -m "Initial tweet saver extension"
```

---

## 4. `manifest.json`: Declaring the Extension

This file describes the structure and permissions of your extension. It uses **Manifest Version 3**, which is now supported in Firefox without service workers (unlike in Chrome).

Here's a minimal, working `manifest.json`:

```json
{
  "manifest_version": 3,
  "name": "Scroll Tweet Saver",
  "version": "0.1.0",
  "description": "Automatically saves every tweet you scroll past.",

  "permissions": ["storage", "downloads"],
  "host_permissions": [
    "https://twitter.com/*",
    "https://x.com/*",
    "https://mobile.twitter.com/*"
  ],

  "background": {
    "scripts": ["background.js"],
    "type": "module"
  },

  "content_scripts": [
    {
      "matches": [
        "https://twitter.com/*",
        "https://x.com/*",
        "https://mobile.twitter.com/*"
      ],
      "js": ["content-script.js"],
      "run_at": "document_idle"
    }
  ],

  "action": {
    "default_popup": "popup/popup.html"
  },

  "icons": {
    "48": "icons/icon48.png",
    "128": "icons/icon128.png"
  }
}
```

### Notes:

- `storage` permission is required for `browser.storage.local`.
- `downloads` lets us programmatically trigger file exports.
- `background` specifies our persistent logic file.
- `content_scripts` define which URLs inject tweet-scraping logic.

---

## 5. Writing the Content Script (`content-script.js`)

The content script scans for tweets using the Twitter/X DOM structure, deduplicates them, and sends tweet objects to the background script. This version is more robust against DOM changes and uses a `Set` to track seen tweets.

```js
/*  ╭─────────────  Configure what we want to pull  ─────────────╮ */
function extractTweet(article) {
  // 1️⃣ canonical tweet URL & ID
  const statusLink = [...article.querySelectorAll("a")].find((a) =>
    /\/status\/\d+/.test(a.href)
  );
  if (!statusLink) return null; // safety
  const url = statusLink.href;
  const id = url.match(/status\/(\d+)/)[1];

  // 2️⃣ author (username + display name if present)
  const authorLink = [...article.querySelectorAll("a")].find((a) =>
    /^https?:\/\/(?:twitter|x)\.com\/[^\/?]+$/.test(a.href)
  );
  const username = authorLink ? authorLink.href.split("/").pop() : null;
  const author = authorLink?.innerText || username;

  // 3️⃣ full tweet text (handles multiple blocks & emojis)
  const textEls = article.querySelectorAll('[data-testid="tweetText"]');
  const text = [...textEls].map((el) => el.innerText).join("\n");

  // 4️⃣ ISO timestamp
  const ts = article.querySelector("time")?.getAttribute("datetime") || null;

  return { id, url, author, username, text, timestamp: ts };
}

/*  ╭─────────────  Live capture as you keep scrolling  ─────────╮ */
window.__seenTweets = new Set();

function handleNewArticle(article) {
  const t = extractTweet(article);
  if (!t || window.__seenTweets.has(t.id)) return;
  window.__seenTweets.add(t.id);
  console.log("[Content Script] Found and sending tweet:", t.id, t.author);
  // Send the tweet to the background script
  browser.runtime
    .sendMessage({ type: "SAVE_TWEET", tweet: t })
    .catch((err) =>
      console.error("[Content Script] Error sending tweet:", err)
    );
}

// Use MutationObserver to detect tweets added to the DOM
const observer = new MutationObserver((muts) => {
  muts.forEach((m) => {
    m.addedNodes.forEach((node) => {
      if (node.nodeType !== 1) return; // Only process element nodes
      // Check if the added node itself is a tweet
      if (node.matches?.('article[data-testid="tweet"]')) {
        handleNewArticle(node);
      }
      // Check if the added node *contains* tweets (important for batch loads)
      node
        .querySelectorAll?.('article[data-testid="tweet"]')
        .forEach(handleNewArticle);
    });
  });
});

console.log("[Content Script] Initializing observer...");
// Observe the entire document for additions to the DOM tree
observer.observe(document, { childList: true, subtree: true });
console.log("[Content Script] Observer attached.");
```

**Key improvements:**

- Uses specific selectors like `[data-testid="tweet"]` which are less likely to break than class names.
- Checks both added nodes and their children for tweets.
- Uses `window.__seenTweets` (a `Set`) to efficiently prevent sending duplicate tweet data.
- Includes logging to help debug in the Browser Console (for the Twitter tab).

---

## 6. Creating the Background Script (`background.js`) (Revised for Export)

This script now handles the export logic directly.

```js
const TWEET_LIMIT = 5000; // Max tweets to store

async function saveTweet(tweet) {
  console.log("[Background Script] Received SAVE_TWEET for:", tweet.id);
  try {
    const { tweets = [] } = await browser.storage.local.get("tweets");
    // Check if tweet already exists
    if (tweets.some((t) => t.id === tweet.id)) {
      console.log(
        "[Background Script] Tweet already exists, skipping:",
        tweet.id
      );
      return false; // Indicate no save occurred
    }
    // Add new tweet and manage list size
    tweets.push(tweet);
    if (tweets.length > TWEET_LIMIT) {
      tweets.shift(); // Remove the oldest tweet
    }
    await browser.storage.local.set({ tweets });
    console.log("[Background Script] Tweet saved successfully:", tweet.id);
    return true; // Indicate save occurred
  } catch (error) {
    console.error("[Background Script] Error saving tweet:", tweet.id, error);
    return false;
  }
}

// Listen for messages from content scripts or the popup
browser.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  console.log(
    "[Background Script] Message received:",
    msg.type,
    "from:",
    sender.tab ? "Tab " + sender.tab.id : "Popup/Other"
  );

  if (msg.type === "SAVE_TWEET") {
    saveTweet(msg.tweet).then((saved) => {
      if (saved) {
        // Notify the popup (if open) that a new tweet was saved
        browser.runtime.sendMessage({ type: "TWEET_SAVED" }).catch((err) => {
          // Ignore error if popup isn't open (expected behavior)
          if (!err.message.includes("Could not establish connection")) {
            console.error("[Background Script] Error notifying popup:", err);
          }
        });
      }
      sendResponse({ ok: true, saved }); // Respond to the content script
    });
    return true; // Indicates async response
  } else if (msg.type === "GET_TWEETS") {
    console.log("[Background Script] Received GET_TWEETS request");
    browser.storage.local
      .get("tweets")
      .then((r) => {
        console.log(
          "[Background Script] Sending tweets to requester:",
          (r.tweets || []).length
        );
        sendResponse(r.tweets || []); // Send back the array of tweets
      })
      .catch((err) => {
        console.error("[Background Script] Error getting tweets:", err);
        sendResponse([]); // Send empty array on error
      });
    return true; // Indicates async response
  } else if (msg.type === "EXPORT_TWEETS") {
    console.log("[Background Script] Received EXPORT_TWEETS request.");

    // Use an immediately-invoked async function expression (IIAFE)
    // to handle the async export logic without blocking the listener
    (async () => {
      let objectUrl = null; // Keep track of the URL to revoke it later
      try {
        // 1. Fetch tweets from storage
        const { tweets = [] } = await browser.storage.local.get("tweets");
        console.log(`[Background Script] Exporting ${tweets.length} tweets.`);
        if (tweets.length === 0) {
          console.warn(
            "[Background Script] Export requested, but no tweets found."
          );
          return; // Exit if nothing to export
        }

        // 2. Create a Blob from the JSON data
        const blob = new Blob([JSON.stringify(tweets, null, 2)], {
          type: "application/json",
        });

        // 3. Create an Object URL for the Blob
        objectUrl = URL.createObjectURL(blob);
        console.log(
          "[Background Script] Created Blob URL for export:",
          objectUrl
        );

        // 4. Initiate the download using the browser.downloads API
        const downloadId = await browser.downloads.download({
          url: objectUrl, // Use the Blob URL created in the background context
          filename: `tweets-archive-${Date.now()}.json`, // Suggest a filename
          saveAs: true, // Prompt the user with a "Save As" dialog
        });
        console.log(
          "[Background Script] Download initiated with ID:",
          downloadId
        );

        // 5. Monitor the download to revoke the Object URL when done
        const handleDownloadChange = (delta) => {
          // Only act on the download we just started
          if (delta.id !== downloadId) return;

          // Check if the download state has changed and is no longer "in_progress"
          if (delta.state && delta.state.current !== "in_progress") {
            console.log(
              `[Background Script] Download ${downloadId} ended (state: ${delta.state.current}). Revoking Blob URL: ${objectUrl}`
            );
            URL.revokeObjectURL(objectUrl); // Clean up the Blob URL to free memory
            browser.downloads.onChanged.removeListener(handleDownloadChange); // Remove this listener
          }
        };
        browser.downloads.onChanged.addListener(handleDownloadChange);
      } catch (err) {
        console.error("[Background Script] Export failed:", err);
        // Ensure the Object URL is revoked even if an error occurs during download initiation
        if (objectUrl) {
          console.log(
            "[Background Script] Revoking Blob URL due to export error."
          );
          URL.revokeObjectURL(objectUrl);
        }
        // Consider sending an error notification back to the popup here if needed
      }
    })();

    // No 'return true' or sendResponse needed here, the export runs independently.
  }
});

console.log("[Background Script] Listener attached.");
```

**Key changes:**

- Added listener for `"EXPORT_TWEETS"` message.
- All Blob creation, `URL.createObjectURL`, and `browser.downloads.download` calls now happen within this background listener.
- Includes the crucial `browser.downloads.onChanged` listener (`handleDownloadChange`) to monitor the download status and call `URL.revokeObjectURL` when the download is complete (or fails), preventing memory leaks.

---

## 7. Building the Popup UI (`popup/`) (Revised for Export)

The popup now delegates the export process.

**`popup/popup.html`**

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Tweet Saver</title>
    <style>
      body {
        font-family: sans-serif;
        width: 300px;
        padding: 10px;
      }
      ul {
        max-height: 200px;
        overflow-y: auto;
        padding-left: 20px;
        margin-top: 10px;
      }
      li {
        margin-bottom: 5px;
        font-size: 0.9em;
      }
      button {
        margin-bottom: 10px;
      }
      details {
        border: 1px solid #ccc;
        padding: 5px;
        margin-bottom: 10px;
      }
      summary {
        cursor: pointer;
        font-weight: bold;
      }
      #stats-details p {
        margin: 2px 0;
        font-size: 0.9em;
      }
    </style>
  </head>
  <body>
    <button id="export">Export JSON</button>
    <details id="export-stats">
      <summary>Export Stats</summary>
      <div id="stats-details"><p>Loading stats...</p></div>
    </details>
    <ul id="tweet-list">
      <li>Loading tweets...</li>
    </ul>
    <script src="popup.js"></script>
  </body>
</html>
```

- Includes a `<details>` element for collapsible stats.
- Adds basic CSS for better presentation.

**`popup/popup.js`**

```js
const list = document.getElementById("tweet-list");
const statsDiv = document.getElementById("stats-details");

// Function to fetch tweets and update the UI (list + stats)
async function updateUI() {
  console.log("[Popup Script] Updating UI...");
  try {
    const tweets = await browser.runtime.sendMessage({ type: "GET_TWEETS" });
    console.log("[Popup Script] Received tweets for UI:", tweets.length);

    // Clear existing list items
    list.innerHTML = "";
    if (tweets.length === 0) {
      list.innerHTML = "<li>No tweets saved yet. Scroll on Twitter/X!</li>";
    } else {
      // Update tweet list (show newest first)
      tweets
        .slice()
        .reverse()
        .forEach((t) => {
          const li = document.createElement("li");
          li.textContent = `[${t.id}] ${t.author || "Unknown"}: ${t.text.slice(
            0,
            80
          )}...`;
          list.appendChild(li);
        });
    }

    // Prepare and display export stats
    const jsonStr = JSON.stringify(tweets, null, 2);
    const totalTweets = tweets.length;
    const sizeBytes = new Blob([jsonStr]).size; // More accurate size
    const sizeKB = (sizeBytes / 1024).toFixed(2);
    const sample =
      totalTweets > 0
        ? `[${tweets[0].id}] ${
            tweets[0].author || "Unknown"
          }: ${tweets[0].text.slice(0, 60)}...`
        : "N/A";

    statsDiv.innerHTML =
      `<p>Total tweets: ${totalTweets}</p>` +
      `<p>Estimated export size: ~${sizeKB} KB</p>` +
      `<p>Sample (oldest): ${sample}</p>`;

    console.log("[Popup Script] UI Updated successfully.");
  } catch (err) {
    console.error("[Popup Script] Failed to update UI:", err);
    list.innerHTML = "<li>Error loading tweets.</li>";
    statsDiv.innerHTML = `<p>Error loading stats</p>`;
  }
}

// Initial UI update when popup opens
updateUI();

// Listen for messages from the background script to auto-refresh
browser.runtime.onMessage.addListener((message, sender) => {
  // Ensure message is from background (no sender.tab) and is the right type
  if (!sender.tab && message.type === "TWEET_SAVED") {
    console.log(
      "[Popup Script] Received TWEET_SAVED notification, refreshing UI..."
    );
    updateUI(); // Re-run the UI update function
  }
});

// Export functionality
document.getElementById("export").addEventListener("click", () => {
  console.log(
    "[Popup Script] Export button clicked, sending EXPORT_TWEETS message."
  );
  // Simply send a message to the background script to handle the export
  browser.runtime
    .sendMessage({ type: "EXPORT_TWEETS" })
    .then(() => {
      console.log("[Popup Script] EXPORT_TWEETS message sent successfully.");
      // Optional: Briefly disable button or show feedback?
    })
    .catch((err) => {
      console.error(
        "[Popup Script] Failed to send EXPORT_TWEETS message:",
        err
      );
      // Inform the user that the export could not be initiated
      alert(`Failed to initiate export: ${err.message}`);
    });
});
```

**Key changes:**

- The `"export"` button's click listener now _only_ sends the `"EXPORT_TWEETS"` message.
- All Blob creation and download logic has been removed from the popup script.

---

## 8. Understanding the Export Fix

The primary reason the original download failed was due to the lifecycle of the popup page and a security restriction in Firefox:

1.  **Popup Lifecycle**: When the export button was clicked, the popup script created a `Blob` containing the JSON data and generated a temporary `blob:moz-extension://...` URL for it using `URL.createObjectURL()`. This URL is tied to the document that created it (the popup). As soon as the user interacted with the "Save As" dialog presented by `browser.downloads.download({ saveAs: true })`, the popup often closed or lost focus, causing its document and the associated Blob URL to become invalid. When the browser's download manager then tried to access the `blob:` URL to get the data, the source was gone, resulting in a failed download.

2.  **Firefox Security Restriction (Bug 1696174)**: Independently, Firefox blocks `browser.downloads.download()` calls that use `blob:` or `data:` URLs created in contexts _other than_ the extension's background script (or service worker in MV3). This is a security measure to prevent potentially malicious web pages or content scripts from easily initiating arbitrary downloads via an extension.

**The Solution: Delegate to the Background Script**

Moving the export logic to the background script (`background.js`) solves both problems:

- **Persistent Context**: The background script runs in a more persistent context than the popup. When it creates the `Blob` and the `ObjectURL`, that URL remains valid even after the popup closes, allowing the download manager to access the data.
- **Bypassing Restrictions**: Since the `Blob` and `ObjectURL` are now created _within_ the background script context, the download call `browser.downloads.download({ url: objectUrl, ... })` is permitted by Firefox, bypassing the restriction mentioned in Bug 1696174.
- **Memory Management**: Generating `ObjectURL`s consumes memory. It's crucial to release this memory once the URL is no longer needed. The fix uses `browser.downloads.onChanged.addListener` to monitor the download's progress. Once the download state changes to anything other than `"in_progress"` (e.g., `"complete"`, `"interrupted"`), the listener calls `URL.revokeObjectURL(objectUrl)`, freeing the memory associated with the Blob.

This delegation pattern (popup requests action, background performs action) is common and robust for handling tasks that require a stable context or privileged APIs like `browser.downloads`.

---

## 9. Debugging the Extension

Extension development often requires checking logs from different parts of the extension.

### Where to Find Logs:

1.  **Content Script (`content-script.js`) Logs:**

    - Appear in the standard **Browser Console** (Ctrl+Shift+J or Cmd+Opt+J).
    - **Important:** Make sure the console is open for the **tab** where the content script is running (e.g., the `twitter.com` or `x.com` tab).
    - Look for messages prefixed with `[Content Script]`.

2.  **Background Script (`background.js`) Logs:**

    - These have their **own dedicated console**.
    - Go to `about:debugging#/runtime/this-firefox` in your Firefox address bar.
    - Find your "Scroll Tweet Saver" extension.
    - Click the **Inspect** button next to it.
    - A new Developer Tools window will open. Check the **Console** tab there.
    - Look for messages prefixed with `[Background Script]`.

3.  **Popup Script (`popup.js`) Logs:**
    - These appear in the console for the popup itself.
    - **Click the extension icon** in the toolbar to open the popup.
    - **Right-click inside the popup** area.
    - Select **Inspect** (or "Inspect Element").
    - A new Developer Tools window will open. Check the **Console** tab there.
    - Look for messages prefixed with `[Popup Script]`.

### Common Issues:

- **"Firefox profile cannot be accessed" / "Profile in use":** Usually means another Firefox instance (often the regular one) is running and locking the profile `web-ext` wants to use.
  - **Solution:** Close ALL Firefox windows completely (you might need `pkill firefox` or Task Manager). Alternatively, use `web-ext run --firefox-profile /path/to/separate/profile` to specify a different profile directory.
- **Tweets Not Being Saved:**
  - Check the **Content Script** console for `[Content Script] Found and sending tweet...` messages. If none appear, the `MutationObserver` or `extractTweet` selectors might be broken by Twitter/X updates.
  - Check the **Background Script** console for `[Background Script] Received SAVE_TWEET...` and `Tweet saved successfully...` messages. If saves fail, check for storage errors.
- **Export Download Fails:**
  - Check the **Popup Script** console for `[Popup Script] Export failed:` errors when you click the button. The error message often indicates the cause (e.g., permission issue, invalid data).
  - Check the **Background Script** console for errors related to `browser.downloads.download`.
  - Ensure the `downloads` permission is in `manifest.json`. Firefox might sometimes block downloads initiated by extensions for security reasons; check browser settings.

---

## 10. Running the Extension

Use `web-ext` for easy development:

```bash
# Navigate to the parent directory containing 'tweet-saver/'
cd /path/to/your/workspace

# Run the extension (auto-reloads on file changes)
web-ext run --source-dir ./tweet-saver/ --firefox-developer-preview

# Or specify your Firefox Developer Edition path if needed:
# web-ext run --source-dir ./tweet-saver/ --firefox /path/to/firefox-dev-edition/firefox
```

Alternatively, load it manually via `about:debugging#/runtime/this-firefox` -> "Load Temporary Add-on" and select the `manifest.json` file inside your `tweet-saver` directory.

---

This updated guide should provide a solid foundation for understanding and extending the Scroll Tweet Saver.
