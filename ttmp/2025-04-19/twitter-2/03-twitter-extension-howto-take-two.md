**Building a Firefox Extension to Collect Tweets While Scrolling**

---

### Introduction

This tutorial will guide you step by step through building a Firefox extension that detects tweets as you scroll through Twitter and lets you export them to a local file. It assumes no prior experience with browser extensions, and we will explain every major concept as we go.

---

### Step 1: What is a Firefox Extension?

A Firefox extension (also called a WebExtension) is a small program that can modify and enhance the functionality of Firefox. These extensions are written in JavaScript, HTML, and CSS. They use a well-defined set of APIs provided by the browser to interact with web pages, browser tabs, storage, and more.

An extension typically consists of the following pieces:

- A **manifest file** that describes the extension's permissions and behavior.
- A **background script** that runs independently of any web page and manages state or events.
- One or more **content scripts** that are injected into web pages to interact with their contents.
- Optional **popup UIs** or browser action buttons.

---

### Step 2: Setting Up the Project Structure

Let's create a folder for our extension. Call it `tweet-collector/` and inside it, create the following files:

```
tweet-collector/
├── manifest.json
├── background.js
├── content.js
└── popup/
    ├── popup.html
    └── popup.js
```

We will fill each file in upcoming steps.

---

### Step 3: Writing the Manifest File

The manifest file is the heart of any browser extension. It tells Firefox what your extension is, what it needs permission to do, and how it's structured. Here's a basic one:

```json
{
  "manifest_version": 3,
  "name": "Tweet Collector",
  "description": "Collect tweets while you scroll and export them.",
  "version": "1.0.0",
  "permissions": ["storage", "downloads"],
  "host_permissions": ["https://twitter.com/*", "https://x.com/*"],
  "background": {
    "service_worker": "background.js",
    "scripts": ["background.js"]
  },
  "content_scripts": [
    {
      "matches": ["https://twitter.com/*", "https://x.com/*"],
      "js": ["content.js"],
      "run_at": "document_idle"
    }
  ],
  "action": {
    "default_popup": "popup/popup.html",
    "default_title": "Export collected tweets"
  }
}
```

Let's break it down:

- `manifest_version`: We use version 3, which is the latest standard.
- `permissions`: These are capabilities the extension needs, such as saving data (`storage`) and downloading files (`downloads`).
- `host_permissions`: These declare which web pages (origins) the extension can interact with—in our case, Twitter URLs.
- `background`: This declares the background script. **Note:** We include both `service_worker` (used by Chrome) and `scripts` (used by Firefox) for cross-browser compatibility, as Firefox doesn't fully support `service_worker` in Manifest V3 yet.
- `content_scripts`: These scripts are injected into matching web pages. Ours will extract tweets.
- `action`: Defines a toolbar button that opens a small popup window for downloading the data.

---

### Step 4: The Content Script – Watching Twitter

Content scripts are the workhorses that interact directly with the web pages you visit. Our `content.js` script will be injected into Twitter pages, where it will keep an eye out for tweets appearing on the screen and extract their juicy details.

#### 4.1 Initial Setup

First, let's set up some basics. We need a way to remember which tweets we've already processed to avoid duplicates. A JavaScript `Set` is perfect for this. We also define a constant holding the CSS selector that helps us identify the main text element within a tweet.

```js
console.log("Content script loading...");

const SEEN = new Set();
const TEXT_SELECTOR = 'div[data-testid="tweetText"]';
```

We also add a log message to confirm when our script starts loading.

#### 4.2 Extracting Tweet Details (`extractTweet` function)

This is the core function responsible for pulling information out of a potential tweet element found on the page (`node`).

```js
function extractTweet(node) {
  console.log("extractTweet called for node:", node);
  // 1. Find the parent <article> element
  const article = node.closest("article");
  if (!article) {
    console.log("No article found for node.");
    return null; // Not a tweet structure we recognize
  }
  console.log("Found article:", article);

  // 2. Find the tweet's unique ID from its timestamp link
  const idHref = article.querySelector("time")?.parentElement?.href;
  if (!idHref) {
    console.log("No ID href found in article.");
    return null; // Cannot find the ID link
  }
  const id = idHref.split("/").pop(); // Extract ID from URL

  // 3. Check if we've already seen this tweet
  if (SEEN.has(id)) {
    console.log(`Tweet ID ${id} already seen.`);
    return null; // Skip duplicates
  }
  console.log(`New tweet ID found: ${id}`);

  // 4. Extract the tweet text (joining spans within the text element)
  const text = [...article.querySelectorAll(TEXT_SELECTOR + " span")]
    .map((el) => el.textContent)
    .join("");

  // 5. Extract the author's name
  const author =
    article.querySelector('a[role="link"] span')?.textContent || "";
  console.log(`Extracted author: ${author}, text: ${text.substring(0, 50)}...`);

  // 6. Mark this tweet ID as seen and return the data
  SEEN.add(id);
  return { id, author, text };
}
```

Here's a breakdown of the steps within `extractTweet`:

1.  **Find the `<article>`:** Tweets on Twitter are typically contained within an `<article>` HTML tag. We search upwards from the found text element (`node`) to locate this parent article.
2.  **Get the ID:** We find the tweet's timestamp element (`<time>`). Its parent link (`<a>`) usually contains the tweet's unique URL, from which we extract the ID (the last part of the URL path).
3.  **Check for Duplicates:** We use our `SEEN` set to check if we've already processed this `id`. If so, we stop.
4.  **Extract Text:** We find all the `<span>` elements inside the main text container (`TEXT_SELECTOR`) and join their content together to form the full tweet text.
5.  **Extract Author:** We find the author's link and grab the text from the `<span>` inside it.
6.  **Record and Return:** We add the `id` to our `SEEN` set and return an object containing the extracted `id`, `author`, and `text`.

#### 4.3 Handling Potential Tweets (`handle` function)

This helper function takes a node, tries to extract tweet data from it using `extractTweet`, and if successful, sends the data to our background script.

```js
function handle(node) {
  console.log("Handling node:", node);
  const tweet = extractTweet(node);
  if (tweet) {
    // If extractTweet returned data, send it to the background script
    console.log("Sending tweet to background:", tweet);
    browser.runtime.sendMessage({ type: "SAVE_TWEET", tweet });
  } else {
    console.log("No tweet extracted from node or already seen.");
  }
}
```

We use `browser.runtime.sendMessage` to communicate with other parts of our extension. Here, we send an object with a `type` of `"SAVE_TWEET"` and the `tweet` data payload.

#### 4.4 Watching for Page Changes (`MutationObserver`)

Twitter loads new tweets dynamically as you scroll. How do we detect them? We use a `MutationObserver`, a powerful browser API that lets us watch for changes in the page's structure (the DOM).

```js
console.log("Setting up MutationObserver...");
const observer = new MutationObserver((muts) => {
  console.log(`MutationObserver triggered with ${muts.length} mutations.`);
  // Loop through all changes that occurred
  for (const m of muts) {
    // Loop through all nodes added to the page
    m.addedNodes.forEach((n) => {
      // Only process element nodes (not text nodes, etc.)
      if (n.nodeType === 1) {
        console.log("Processing added node:", n);
        // Check if the added node itself is a tweet text element
        if (n.matches?.(TEXT_SELECTOR)) {
          console.log("Node matches TEXT_SELECTOR, handling...");
          handle(n);
        }
        // Check if any children of the added node are tweet text elements
        const children = n.querySelectorAll?.(TEXT_SELECTOR);
        if (children && children.length > 0) {
          console.log(
            `Found ${children.length} matching children, handling each...`
          );
          children.forEach(handle);
        }
      }
    });
  }
});

// Start observing the entire body of the page for added child elements
observer.observe(document.body, { childList: true, subtree: true });
console.log("MutationObserver observing document.body.");
```

We create an observer that runs a function whenever mutations (changes) happen. Inside that function, we look specifically at `addedNodes` (new elements added to the page). For each added element, we check if it or any of its children match our `TEXT_SELECTOR`. If they do, we pass them to our `handle` function.
Finally, we tell the observer to start watching the `document.body` and all its descendants (`subtree: true`) for additions (`childList: true`).

#### 4.5 Handling Initially Loaded Tweets

The `MutationObserver` only catches tweets loaded _after_ our script runs. What about tweets already present when the page first loads? We need to scan for those too.

```js
console.log("Handling existing tweets on page load...");
document.querySelectorAll(TEXT_SELECTOR).forEach(handle);
console.log("Content script loaded and initial handling complete.");
```

This simple line uses `querySelectorAll` to find all elements matching `TEXT_SELECTOR` that exist _right now_ and sends each one to our `handle` function. This ensures we capture tweets visible on initial load.

---

### Step 5: The Background Script – Saving Tweets

The background script runs in a service worker and can persist data or respond to events. We'll use `browser.storage.local`, a simple key-value store, to keep tweet data across pages.

Create `background.js`:

```js
console.log("Background script loading...");

async function save(tweet) {
  console.log("Save function called with tweet:", tweet);
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (tweets.some((t) => t.id === tweet.id)) {
    console.log(`Tweet ID ${tweet.id} already exists in storage. Skipping.`);
    return;
  }
  tweets.push(tweet);
  await browser.storage.local.set({ tweets });
  console.log(`Tweet ID ${tweet.id} saved. Total tweets: ${tweets.length}`);
}

// Message listener will be added in Step 7
```

When a content script sends a message, we check if the tweet is new and then store it.

---

### Step 6: Building the Popup UI

We'll create a modern, visually appealing popup with:

- Statistics showing the number of tweets collected and unique authors
- A real-time list of collected tweets that updates as you scroll through Twitter
- A download button styled to match Twitter's design aesthetic

Create `popup/popup.html`:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <style>
      body {
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
          Helvetica, Arial, sans-serif;
        width: 350px;
        min-height: 300px;
        padding: 16px;
        background-color: #f7f9fa;
        color: #2f3336;
      }
      h1 {
        font-size: 18px;
        font-weight: 600;
        margin-top: 0;
        margin-bottom: 16px;
        color: #1da1f2;
      }
      .stats {
        display: flex;
        justify-content: space-between;
        margin-bottom: 16px;
        padding: 12px;
        background-color: white;
        border-radius: 8px;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
      }
      .count {
        font-size: 14px;
        font-weight: 500;
      }
      .count-number {
        font-size: 24px;
        font-weight: 700;
        display: block;
        color: #1da1f2;
      }
      .tweet-list {
        max-height: 350px;
        overflow-y: auto;
        margin-bottom: 16px;
        border-radius: 8px;
        background-color: white;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
      }
      .tweet {
        padding: 12px;
        border-bottom: 1px solid #e6ecf0;
      }
      .tweet:last-child {
        border-bottom: none;
      }
      .author {
        font-weight: 700;
        margin-bottom: 4px;
      }
      .text {
        font-size: 14px;
        margin: 0;
        color: #14171a;
        line-height: 1.4;
      }
      button {
        display: block;
        width: 100%;
        padding: 12px;
        background-color: #1da1f2;
        color: white;
        border: none;
        border-radius: 25px;
        font-weight: bold;
        cursor: pointer;
        transition: background-color 0.2s;
      }
      button:hover {
        background-color: #1a91da;
      }
      .empty-state {
        text-align: center;
        padding: 24px 16px;
        color: #657786;
        font-style: italic;
      }
    </style>
  </head>
  <body>
    <h1>Tweet Collector</h1>

    <div class="stats">
      <div class="count">
        <span class="count-number" id="tweet-count">0</span>
        Tweets collected
      </div>
      <div class="count">
        <span class="count-number" id="author-count">0</span>
        Unique authors
      </div>
    </div>

    <div class="tweet-list" id="tweet-list">
      <div class="empty-state">
        No tweets collected yet. Scroll through Twitter to collect tweets.
      </div>
    </div>

    <button id="download">Download tweets JSON</button>
    <script src="popup.js"></script>
  </body>
</html>
```

The HTML includes:

- A title with the extension name
- A statistics section showing counts of tweets and authors
- A scrollable container for displaying collected tweets
- A styled download button

Now create `popup/popup.js`:

```js
// Function to display tweets in the popup UI
async function displayTweets() {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  const tweetList = document.getElementById("tweet-list");
  const tweetCount = document.getElementById("tweet-count");
  const authorCount = document.getElementById("author-count");

  // Update tweet count
  tweetCount.textContent = tweets.length;

  // Calculate unique authors
  const uniqueAuthors = new Set(tweets.map((tweet) => tweet.author)).size;
  authorCount.textContent = uniqueAuthors;

  // Clear the list
  tweetList.innerHTML = "";

  if (tweets.length === 0) {
    // Show empty state message
    const emptyState = document.createElement("div");
    emptyState.className = "empty-state";
    emptyState.textContent =
      "No tweets collected yet. Scroll through Twitter to collect tweets.";
    tweetList.appendChild(emptyState);
    return;
  }

  // Add each tweet to the list (most recent first)
  tweets
    .slice()
    .reverse()
    .forEach((tweet) => {
      const tweetElement = document.createElement("div");
      tweetElement.className = "tweet";

      const authorElement = document.createElement("div");
      authorElement.className = "author";
      authorElement.textContent = tweet.author;

      const textElement = document.createElement("p");
      textElement.className = "text";
      textElement.textContent = tweet.text;

      tweetElement.appendChild(authorElement);
      tweetElement.appendChild(textElement);
      tweetList.appendChild(tweetElement);
    });
}

// Initial display when popup opens
document.addEventListener("DOMContentLoaded", displayTweets);

// Set up listener for changes in storage (live updates)
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.tweets) {
    displayTweets();
  }
});

// Set up download button
document.getElementById("download").addEventListener("click", () => {
  browser.runtime.sendMessage({ type: "DOWNLOAD_TWEETS" });
});
```

The JavaScript:

1. Creates a function to display tweets in the popup UI
2. Updates statistics (tweet count and unique author count)
3. Populates the tweet list with the most recently collected tweets at the top
4. Sets up a listener to update the UI in real-time when new tweets are collected
5. Initializes the download button

This provides users with live feedback as tweets are collected while scrolling through Twitter.

---

### Step 7: Handling the Download in the Background

The background script will listen for the `DOWNLOAD_TWEETS` message and perform the actual download.

**Why handle downloads in the background script?**

Popup windows are temporary. If the user clicks the download button and then clicks elsewhere, the popup might close before the asynchronous download operation fully completes. This can lead to failed downloads. The background script, however, runs persistently as long as the browser is open (or until it becomes idle in Manifest V3), providing a stable environment to manage operations like downloads reliably.

Modify `background.js` to add the message listener and the download logic:

```js
// Existing save function from Step 5...
console.log("Background script loading...");

// ... (save function code as above)

async function handleDownloadRequest() {
  console.log("Download request received.");
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (tweets.length === 0) {
    console.log("No tweets to download.");
    return;
  }

  console.log(`Preparing to download ${tweets.length} tweets.`);
  const blob = new Blob([JSON.stringify(tweets, null, 2)], {
    type: "application/json",
  });
  const url = URL.createObjectURL(blob);
  // Create a filename with a timestamp. We replace colons (:) from toISOString()
  // because they are illegal characters in filenames on some operating systems.
  const filename = `tweets_${new Date().toISOString().replace(/:/g, "-")}.json`;

  try {
    await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true, // Prompt user for save location
    });
    console.log(`Download initiated for ${filename}`);
  } catch (err) {
    console.error("Download failed:", err);
  }
  // Note: Firefox manages revoking the object URL automatically for downloads
}

browser.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  console.log("Received message in background script:", msg);
  if (msg.type === "SAVE_TWEET") {
    // ... (call save)
  } else if (msg.type === "DOWNLOAD_TWEETS") {
    // ... (call handleDownloadRequest)
  } else {
    // ... (log unknown message)
  }
});

console.log("Background script loaded and message listener added.");
```

This code retrieves tweets from storage, creates a downloadable JSON file (Blob), generates a temporary URL for it (handling potential illegal characters in the default timestamp), and then uses the `browser.downloads.download` API to trigger the download, prompting the user to choose a save location.

---

### Step 8: Testing the Extension

1. Open Firefox and go to `about:debugging#/runtime/this-firefox`
2. Click "Load Temporary Add-on"
3. Select your `manifest.json` file
4. Open twitter.com and scroll around
5. Click the toolbar icon and use the popup button to download the collected tweets

---

### Step 9: Packaging for Distribution

Once you're done developing:

1. Zip the folder
2. Upload to [addons.mozilla.org](https://addons.mozilla.org/)

You can also install `web-ext` (a Mozilla CLI tool) to automate this.

---
