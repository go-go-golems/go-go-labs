**Building a Firefox Extension: A Step-by-Step Guide**

---

### Introduction

This guide walks you through creating a basic Firefox extension. We'll build an example extension that interacts with a specific website (like Twitter/X) to extract information as you browse and allows you to download that information. This guide focuses on the concepts and requires you to fill in some of the implementation details.

---

### Step 1: Understanding Browser Extensions

Browser extensions (WebExtensions) are small programs enhancing browser functionality. They use JavaScript, HTML, and CSS, interacting with the browser and web pages via specific APIs.

Key components:

- **Manifest (`manifest.json`):** The blueprint defining permissions, scripts, and structure.
- **Background Script:** Runs independently, managing long-term tasks or state.
- **Content Scripts:** Injected into web pages to interact with their content (DOM).
- **UI Elements:** Popups, sidebars, or option pages for user interaction.

---

### Step 2: Project Setup

Create a directory for your extension (e.g., `my-first-extension/`). Inside, create the basic files:

```
my-first-extension/
├── manifest.json
├── background.js
├── content.js
└── popup/
    ├── popup.html
    └── popup.js
```

---

### Step 3: The Manifest File (`manifest.json`)

This JSON file describes your extension. Create `manifest.json` with the following structure. Think about what each field means:

```json
{
  "manifest_version": 3, // Specifies the manifest version (use the latest)
  "name": "My Data Collector", // Your extension's name
  "description": "Extracts data while browsing and allows download.", // A brief description
  "version": "1.0.0", // Your extension's version

  // What permissions does your extension need?
  // - To store data? (Hint: "storage")
  // - To download files? (Hint: "downloads")
  "permissions": [
    /* Add necessary permissions here */
  ],

  // Which websites should your content script run on?
  // Use match patterns like "https://*.example.com/*"
  "host_permissions": [
    /* Add website patterns here, e.g., "https://twitter.com/*", "https://x.com/*" */
  ],

  // Define the background script
  "background": {
    // For Manifest V3, Chrome uses service_worker, Firefox still often needs scripts
    "service_worker": "background.js", // Chrome compatibility
    "scripts": ["background.js"] // Firefox compatibility
  },

  // Define the content script(s)
  "content_scripts": [
    {
      // Match patterns defining where this script runs (should match host_permissions)
      "matches": [
        /* Add website patterns here */
      ],
      // The script file(s) to inject
      "js": ["content.js"],
      // When should the script run? 'document_idle' is often a good choice.
      "run_at": "document_idle"
    }
  ],

  // Define the browser action (the toolbar button)
  "action": {
    // The HTML file for the popup window
    "default_popup": "popup/popup.html",
    // Tooltip text when hovering over the icon
    "default_title": "Show Collected Data"
  }
}
```

**Task:** Fill in the `permissions`, `host_permissions`, and `matches` arrays based on the goal of collecting data (like tweets) and downloading it.

---

### Step 4: The Content Script (`content.js`) - Interacting with the Page

This script runs on the target website(s) defined in the manifest. Its job is to find relevant data (e.g., tweets) and send it to the background script for storage.

#### 4.1 Initial Setup

Start `content.js`. You'll need a way to avoid processing the same item multiple times. A JavaScript `Set` is suitable for storing unique IDs. You also need a way to identify the data elements you want to extract (e.g., the main text of a tweet). Use your browser's developer tools (Inspect Element) to find a reliable CSS selector for these elements.

```js
console.log("Content script starting...");

// Keep track of items already processed to avoid duplicates
const SEEN_IDS = new Set();

// Define the CSS selector for the main content element you want to extract
// Example for Twitter: 'div[data-testid="tweetText"]'
const CONTENT_SELECTOR = "YOUR_CSS_SELECTOR_HERE";

// Define selectors for other parts you might want, like author or timestamp
// const AUTHOR_SELECTOR = 'SELECTOR_FOR_AUTHOR';
// const TIMESTAMP_SELECTOR = 'SELECTOR_FOR_TIMESTAMP'; // Or a link containing the ID
```

**Task:** Replace `YOUR_CSS_SELECTOR_HERE` with the actual CSS selector for the content you want to capture from the target website. Add other selectors as needed.

#### 4.2 Extracting Data (`extractData` function)

Create a function that takes a potential data element (a DOM node) and tries to extract the relevant information.

```js
function extractData(node) {
  console.log("Attempting to extract data from node:", node);

  // Pseudocode:
  // 1. Find the main container element (e.g., the <article> for a tweet).
  //    - Use `node.closest('article_selector_here')`. If not found, return null.
  // 2. Find a unique ID for this item.
  //    - Often found in a link's `href` or a timestamp. Inspect the page structure!
  //    - Extract the ID (e.g., from the URL). If no ID found, return null.
  // 3. Check if this ID is already in SEEN_IDS.
  //    - If yes, log it and return null (it's a duplicate).
  // 4. Extract the main content text.
  //    - Find the element using CONTENT_SELECTOR within the container.
  //    - Get its text content. Handle cases where text might be split across multiple spans.
  // 5. Extract other details (e.g., author).
  //    - Use the selectors defined earlier.
  // 6. Add the extracted ID to SEEN_IDS.
  // 7. Return an object containing the extracted data (e.g., { id, text, author }).

  // Example structure (you need to implement the logic):
  const container = node.closest(/* ... */);
  if (!container) return null;

  const idLink = container.querySelector(/* ... */)?.href;
  if (!idLink) return null;
  const id = idLink.split('/').pop(); // Adapt this logic based on the URL structure

  if (SEEN_IDS.has(id)) {
    console.log(`ID ${id} already seen.`);
    return null;
  }

  const textElement = container.querySelector(CONTENT_SELECTOR);
  const text = /* ... logic to get full text from textElement ... */;

  // const authorElement = container.querySelector(AUTHOR_SELECTOR);
  // const author = /* ... logic to get author name ... */;

  console.log(`Extracted new data for ID ${id}`);
  SEEN_IDS.add(id);
  return {
    id: id,
    text: text,
    // author: author, // Uncomment if extracting author
    // Add other extracted fields here
  };
}
```

**Task:** Implement the logic inside `extractData` based on the pseudocode and the specific structure of the website you are targeting. Use DOM methods like `querySelector`, `querySelectorAll`, `closest`, `textContent`, etc.

#### 4.3 Handling Found Data (`handleNode` function)

Create a helper function that calls `extractData` and, if successful, sends the data to the background script using the `browser.runtime.sendMessage` API.

```js
function handleNode(node) {
  console.log("Handling potential data node:", node);
  const data = extractData(node);

  if (data) {
    // If data was extracted successfully...
    console.log("Sending data to background script:", data);
    // Send a message object with a type and the data payload
    browser.runtime
      .sendMessage({ type: "SAVE_DATA", payload: data })
      .catch((error) => console.error("Error sending message:", error));
  } else {
    // Optional: Log why data wasn't extracted (e.g., not found, duplicate)
    // console.log("No data extracted or already seen for this node.");
  }
}
```

**Task:** Ensure the message type (`"SAVE_DATA"`) is consistent with what the background script will listen for.

#### 4.4 Watching for Dynamic Content (`MutationObserver`)

Websites like Twitter load content dynamically as you scroll. A `MutationObserver` lets your script react to changes in the page's structure (DOM).

```js
console.log("Setting up MutationObserver...");

const observer = new MutationObserver((mutationsList) => {
  // This function runs whenever the observed DOM changes
  console.log(`DOM mutations detected: ${mutationsList.length}`);

  for (const mutation of mutationsList) {
    if (mutation.type === "childList") {
      // We are interested in nodes being added to the page
      mutation.addedNodes.forEach((node) => {
        // Check if the added node itself is the content element or contains it
        if (node.nodeType === Node.ELEMENT_NODE) {
          // Process only element nodes
          console.log("Processing added node:", node);

          // Pseudocode:
          // 1. Check if the added node 'node' matches the CONTENT_SELECTOR.
          //    - If yes, call handleNode(node).
          // 2. Check if the added node 'node' *contains* any elements matching CONTENT_SELECTOR.
          //    - Use `node.querySelectorAll(CONTENT_SELECTOR)`.
          //    - If found, loop through the results and call handleNode for each one.

          // Example implementation structure:
          if (node.matches && node.matches(CONTENT_SELECTOR)) {
            handleNode(node);
          } else if (node.querySelectorAll) {
            const children = node.querySelectorAll(CONTENT_SELECTOR);
            if (children.length > 0) {
              console.log(
                `Found ${children.length} matching children within added node.`
              );
              children.forEach(handleNode);
            }
          }
        }
      });
    }
  }
});

// Configuration for the observer:
const config = {
  childList: true, // Observe additions/removals of child nodes
  subtree: true, // Observe the target node and all its descendants
};

// Start observing the document body for changes
observer.observe(document.body, config);
console.log("MutationObserver is now observing the document body.");
```

**Task:** Implement the pseudocode logic within the `MutationObserver` callback to correctly identify and handle relevant nodes added to the page.

#### 4.5 Handling Initial Content

The `MutationObserver` only catches content loaded _after_ your script runs. You also need to process any relevant content already present when the page loads.

```js
console.log("Processing existing content on page load...");

// Pseudocode:
// 1. Find all elements currently in the document that match CONTENT_SELECTOR.
//    - Use `document.querySelectorAll(CONTENT_SELECTOR)`.
// 2. Loop through the found elements.
// 3. For each element, call `handleNode`.

// Example implementation:
document.querySelectorAll(CONTENT_SELECTOR).forEach(handleNode);

console.log("Content script setup complete.");
```

**Task:** Ensure this initial scan correctly uses your `CONTENT_SELECTOR` and `handleNode` function.

---

### Step 5: The Background Script (`background.js`) - Storing Data

This script runs in the background and listens for messages from the content script. We'll use it to store the collected data using `browser.storage.local`.

```js
console.log("Background script loading...");

// Store collected data here. Using an array is common.
let collectedData = [];

// Function to load existing data from storage when the script starts
async function loadInitialData() {
  try {
    const result = await browser.storage.local.get("collectedDataStore");
    if (result.collectedDataStore) {
      collectedData = result.collectedDataStore;
      console.log(`Loaded ${collectedData.length} items from storage.`);
    } else {
      console.log("No data found in storage.");
    }
  } catch (error) {
    console.error("Error loading data from storage:", error);
  }
}

// Function to save data item received from content script
async function saveData(item) {
  console.log("Attempting to save item:", item);

  // Pseudocode:
  // 1. Check if an item with the same ID already exists in the `collectedData` array.
  //    - Use `collectedData.some(d => d.id === item.id)`
  // 2. If it's a duplicate, log it and return (do nothing).
  // 3. If it's new, add the `item` to the `collectedData` array.
  // 4. Save the updated `collectedData` array to `browser.storage.local`.
  //    - Use `browser.storage.local.set({ collectedDataStore: collectedData })`.
  //    - Remember storage operations are asynchronous (use await).
  // 5. Log success and the new total count.

  // Example structure:
  const alreadyExists = collectedData.some((d) => d.id === item.id);
  if (alreadyExists) {
    console.log(`Item ID ${item.id} already exists. Skipping.`);
    return;
  }

  collectedData.push(item);
  try {
    await browser.storage.local.set({ collectedDataStore: collectedData });
    console.log(
      `Item ID ${item.id} saved. Total items: ${collectedData.length}`
    );
  } catch (error) {
    console.error("Error saving data to storage:", error);
    // Optional: Remove the item if saving failed to keep array and storage in sync
    collectedData.pop();
  }
}

// Function to handle download requests (will be triggered by the popup)
async function handleDownloadRequest() {
  console.log("Download request received.");
  if (collectedData.length === 0) {
    console.log("No data collected to download.");
    // Optional: Notify the user via the popup?
    return;
  }

  console.log(`Preparing download for ${collectedData.length} items.`);

  // Pseudocode:
  // 1. Convert the `collectedData` array into a JSON string.
  //    - Use `JSON.stringify(collectedData, null, 2)` for pretty printing.
  // 2. Create a Blob object from the JSON string.
  //    - Set the type to 'application/json'.
  // 3. Create an object URL for the Blob.
  //    - Use `URL.createObjectURL(blob)`.
  // 4. Generate a filename (e.g., include a timestamp).
  //    - Make sure the filename is valid (e.g., replace colons from timestamps).
  // 5. Use the `browser.downloads.download` API.
  //    - Pass the object URL, filename, and `saveAs: true` to prompt the user.
  // 6. Wrap the download call in a try/catch block to handle errors.
  // 7. Important: In Firefox, the object URL is revoked automatically after download.
  //    In Chrome, you might need `URL.revokeObjectURL(url)` in some contexts,
  //    but for downloads, it's usually handled.

  // Example structure:
  try {
    const jsonString = JSON.stringify(collectedData, null, 2);
    const blob = new Blob([jsonString], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
    const filename = `collected_data_${timestamp}.json`;

    await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true, // Ask user where to save
    });
    console.log(`Download initiated for ${filename}`);
    // URL.revokeObjectURL(url); // Usually not needed for browser.downloads in MV3
  } catch (err) {
    console.error("Download failed:", err);
  }
}

// Listener for messages from other parts of the extension (content script, popup)
browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
  console.log("Background script received message:", message);

  if (message.type === "SAVE_DATA") {
    // Received data from content script
    saveData(message.payload);
    // Note: sendResponse is not used here, but could be to send ack back
    return true; // Indicate you might send a response asynchronously (even if you don't)
  } else if (message.type === "DOWNLOAD_DATA") {
    // Received request from popup
    handleDownloadRequest();
    return true; // Indicate async handling
  } else if (message.type === "GET_DATA_COUNT") {
    // Example: Popup might ask for count on opening
    sendResponse({ count: collectedData.length });
    // Return false or nothing for synchronous response
  } else {
    console.warn("Unknown message type received:", message.type);
  }
  // Return true if you intend to use sendResponse asynchronously, otherwise omit or return false.
});

// Load existing data when the background script starts
loadInitialData();

console.log("Background script loaded and ready.");
```

**Task:** Implement the pseudocode within `saveData` and `handleDownloadRequest`. Ensure the message types match those used in `content.js` and `popup.js`. Consider what other messages might be useful (e.g., getting the current count for the popup).

---

### Step 6: The Popup UI (`popup/popup.html` and `popup/popup.js`)

This is the small window that appears when the user clicks the extension's toolbar icon.

#### 6.1 Popup HTML (`popup/popup.html`)

Create a simple HTML structure.

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Data Collector</title>
    <link rel="stylesheet" href="popup.css" />
    <!-- Optional: Link to CSS -->
    <style>
      /* Or include some basic CSS here */
      body {
        font-family: sans-serif;
        padding: 10px;
        min-width: 250px;
      }
      .stats {
        margin-bottom: 10px;
      }
      #data-list {
        max-height: 200px;
        overflow-y: auto;
        border: 1px solid #ccc;
        margin-bottom: 10px;
        padding: 5px;
      }
      .data-item {
        border-bottom: 1px solid #eee;
        padding: 3px 0;
      }
      .data-item:last-child {
        border-bottom: none;
      }
      button {
        padding: 8px 15px;
        cursor: pointer;
      }
    </style>
  </head>
  <body>
    <h1>Collected Data</h1>

    <div class="stats">
      Items collected: <span id="item-count">0</span>
      <!-- Add more stats if needed, e.g., unique authors -->
    </div>

    <div id="data-list">
      <p>Loading data...</p>
    </div>

    <button id="download-button">Download as JSON</button>

    <script src="popup.js"></script>
  </body>
</html>
```

**Task:** Adjust the HTML structure and basic styling as needed. Add placeholders for any other statistics you want to display.

#### 6.2 Popup JavaScript (`popup/popup.js`)

This script runs when the popup is opened. It needs to fetch the latest data/stats from the background script (or storage) and display them. It also handles the download button click.

```js
console.log("Popup script running.");

const itemCountElement = document.getElementById("item-count");
const dataListElement = document.getElementById("data-list");
const downloadButton = document.getElementById("download-button");

// Function to update the popup display
async function updatePopup() {
  console.log("Updating popup display...");

  // Pseudocode:
  // 1. Get the latest data from storage.
  //    - Use `browser.storage.local.get("collectedDataStore")`.
  // 2. Update the statistics (e.g., item count).
  // 3. Clear the current list display in `dataListElement`.
  // 4. If no data, show a "No data collected yet" message.
  // 5. If data exists:
  //    - Loop through the data items (maybe show newest first - reverse the array).
  //    - For each item, create HTML elements (e.g., divs) to display its details (text, author).
  //    - Append these elements to `dataListElement`.

  // Example structure:
  try {
    const result = await browser.storage.local.get("collectedDataStore");
    const data = result.collectedDataStore || [];

    itemCountElement.textContent = data.length;
    dataListElement.innerHTML = ""; // Clear previous content

    if (data.length === 0) {
      dataListElement.innerHTML =
        "<p>No data collected yet. Browse the target site!</p>";
    } else {
      // Display items (e.g., newest first)
      data
        .slice()
        .reverse()
        .forEach((item) => {
          const itemDiv = document.createElement("div");
          itemDiv.className = "data-item";
          // You'll need to adapt this based on your data structure
          itemDiv.textContent = `${item.id}: ${item.text.substring(0, 50)}...`;
          // Add more details if needed
          dataListElement.appendChild(itemDiv);
        });
    }
  } catch (error) {
    console.error("Error updating popup:", error);
    dataListElement.innerHTML = "<p>Error loading data.</p>";
  }
}

// Add event listener for the download button
downloadButton.addEventListener("click", () => {
  console.log("Download button clicked.");
  // Send a message to the background script to trigger the download
  browser.runtime
    .sendMessage({ type: "DOWNLOAD_DATA" })
    .catch((error) => console.error("Error sending download message:", error));
});

// Listen for changes in storage - this updates the popup LIVE if it's open
// when new data is saved by the background script.
browser.storage.onChanged.addListener((changes, area) => {
  // Check if the change happened in 'local' storage and affected our data key
  if (area === "local" && changes.collectedDataStore) {
    console.log("Storage changed, updating popup...");
    updatePopup(); // Re-render the popup content
  }
});

// Initial update when the popup is opened
document.addEventListener("DOMContentLoaded", updatePopup);
```

**Task:** Implement the pseudocode in `updatePopup` to correctly fetch data and render it in the list. Ensure the message type sent by the download button matches the one the background script listens for (`DOWNLOAD_DATA`).

---

### Step 7: Loading and Testing

1.  Open Firefox.
2.  Navigate to `about:debugging#/runtime/this-firefox`.
3.  Click "Load Temporary Add-on...".
4.  Select the `manifest.json` file from your extension's directory.
5.  Look for errors in the debugging console (both for the extension itself and in the regular browser console when on the target website).
6.  Navigate to the website you specified in `host_permissions`.
7.  Scroll around or interact with the page to trigger your content script. Check the browser console for logs from `content.js`.
8.  Check the extension's background script console (via `about:debugging`) for logs from `background.js`.
9.  Click the extension's icon in the toolbar to open the popup. Check if it displays correctly.
10. Click the download button in the popup and verify the JSON file is downloaded.

**Debugging Tips:**

- Use `console.log` extensively in all scripts.
- Check the Browser Console (Ctrl+Shift+J or Cmd+Opt+J) for content script errors and logs.
- Check the Extension Console (click "Inspect" next to your extension in `about:debugging`) for background and popup script errors/logs.
- Reload the extension in `about:debugging` after making changes to the code. You might also need to refresh the target website page.

---

### Step 8: Next Steps (Optional)

- **Error Handling:** Add more robust error checking (e.g., what if selectors change on the target website?).
- **User Options:** Create an options page (`options_ui` in manifest) to let users configure selectors or target sites.
- **Clear Data:** Add a button in the popup to clear the collected data from storage.
- **Refinement:** Improve CSS selectors, data extraction logic, and UI presentation.
- **Packaging:** Once happy, zip the directory and submit it to [addons.mozilla.org](https://addons.mozilla.org/) or use `web-ext` to lint and build.

---

This guide provides the framework. The core task is to implement the specific logic for finding, extracting, and storing the data relevant to your chosen website and use case. Good luck!
