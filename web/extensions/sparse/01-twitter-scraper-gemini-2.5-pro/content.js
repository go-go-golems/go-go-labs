console.log("Content script starting...");

const SEEN_IDS = new Set();

// Selectors (these might change based on Twitter/X updates)
const TWEET_CONTAINER_SELECTOR = 'article[data-testid="tweet"]'; // The main container for a tweet
const TWEET_TEXT_SELECTOR = 'div[data-testid="tweetText"]'; // The element containing the tweet text
const TWEET_ID_LINK_SELECTOR = 'a[href*="/status/"]'; // Link containing the status ID

function extractData(node) {
  // console.log("Attempting to extract data from node:", node);

  // 1. Find the main tweet container element relative to the text node.
  const container = node.closest(TWEET_CONTAINER_SELECTOR);
  if (!container) {
    // console.log("Node is not inside a tweet container.");
    return null;
  }

  // 2. Find a unique ID for this item.
  // Look for the timestamp link which usually contains the ID
  let tweetId = null;
  const links = container.querySelectorAll(TWEET_ID_LINK_SELECTOR);
  for (const link of links) {
    // Find the link that looks like a status link (ends with digits)
    const match = link.href.match(/\/status\/(\d+)$/);
    if (match && match[1]) {
      tweetId = match[1];
      break;
    }
  }

  if (!tweetId) {
    console.log("Could not find tweet ID link within container:", container);
    return null;
  }

  // 3. Check if this ID is already in SEEN_IDS.
  if (SEEN_IDS.has(tweetId)) {
    // console.log(`ID ${tweetId} already seen.`);
    return null;
  }

  // 4. Extract the main content text.
  const textElement = container.querySelector(TWEET_TEXT_SELECTOR);
  if (!textElement) {
    console.warn(
      `Could not find tweet text for ID ${tweetId} in container:`,
      container
    );
    return null;
  }
  // Get all text content, including potential nested elements (like links, hashtags)
  const text = textElement.textContent || "";

  if (!text.trim()) {
    console.log(`Found empty text for ID ${tweetId}, skipping.`);
    return null;
  }

  // 5. No other details needed for this simple version.

  // 6. Add the extracted ID to SEEN_IDS.
  console.log(`Extracted new data for ID ${tweetId}`);
  SEEN_IDS.add(tweetId);

  // 7. Return an object containing the extracted data.
  return {
    id: tweetId,
    text: text.trim(),
    // Add other fields here if needed later (e.g., timestamp, author)
    // timestamp: new Date().toISOString() // Example timestamp
  };
}

function handleNode(node) {
  // console.log("Handling potential data node:", node);
  if (!node || node.nodeType !== Node.ELEMENT_NODE) {
    return; // Only process element nodes
  }

  // Use the text selector as the primary target for extraction triggering
  const data = extractData(node);

  if (data) {
    // If data was extracted successfully...
    console.log("Sending data to background script:", data);
    browser.runtime
      .sendMessage({ type: "SAVE_DATA", payload: data })
      .catch((error) => {
        if (error.message.includes("Could not establish connection")) {
          console.warn(
            "Connection to background script lost. May happen during extension reload."
          );
        } else {
          console.error("Error sending message:", error);
        }
      });
  } else {
    // console.log("No data extracted or already seen for this node.");
  }
}

// --- MutationObserver Setup ---
console.log("Setting up MutationObserver...");

const observer = new MutationObserver((mutationsList) => {
  // console.log(`DOM mutations detected: ${mutationsList.length}`);

  for (const mutation of mutationsList) {
    if (mutation.type === "childList") {
      mutation.addedNodes.forEach((node) => {
        if (node.nodeType === Node.ELEMENT_NODE) {
          // console.log("Processing added node:", node);

          // 1. Check if the added node itself matches the text selector.
          if (node.matches && node.matches(TWEET_TEXT_SELECTOR)) {
            handleNode(node);
          } else if (node.querySelectorAll) {
            // 2. Check if the added node *contains* elements matching the text selector.
            const children = node.querySelectorAll(TWEET_TEXT_SELECTOR);
            if (children.length > 0) {
              // console.log(`Found ${children.length} matching children within added node.`);
              children.forEach(handleNode);
            }
          }
        }
      });
    }
  }
});

const config = {
  childList: true,
  subtree: true,
};

// Start observing the document body. Using document.body is generally sufficient.
// Make sure body exists before observing
if (document.body) {
  observer.observe(document.body, config);
  console.log("MutationObserver is now observing the document body.");
} else {
  // If script runs before body exists, wait for DOMContentLoaded
  document.addEventListener("DOMContentLoaded", () => {
    if (document.body) {
      observer.observe(document.body, config);
      console.log(
        "MutationObserver is now observing the document body (after DOMContentLoaded)."
      );
    } else {
      console.error("Document body not found even after DOMContentLoaded.");
    }
  });
}

// --- Initial Content Scan ---
function scanInitialContent() {
  console.log("Processing existing content on page load...");
  // Find all elements currently in the document that match the TWEET_TEXT_SELECTOR.
  document.querySelectorAll(TWEET_TEXT_SELECTOR).forEach(handleNode);
  console.log("Initial content scan complete.");
}

// Run initial scan. Since run_at is document_idle, DOM should be ready.
// However, dynamic loads might still occur, hence the check.
if (
  document.readyState === "complete" ||
  document.readyState === "interactive"
) {
  scanInitialContent();
} else {
  document.addEventListener("DOMContentLoaded", scanInitialContent);
}

console.log("Content script setup complete.");
