console.log("Content script starting...");

// Keep track of tweets already processed to avoid duplicates
const SEEN_IDS = new Set();

// Define the CSS selector for the tweet text
const CONTENT_SELECTOR = "div[data-testid='tweetText']";
// Selector for tweet author
const AUTHOR_SELECTOR = "div[data-testid='User-Name'] a[tabindex='-1']";
// Selector for timestamp containing tweet ID
const TIMESTAMP_SELECTOR = "a[href*='/status/']";

function extractData(node) {
  console.log("Attempting to extract data from node:", node);

  // Find the main tweet container element
  const container = node.closest("article[data-testid='tweet']");
  if (!container) {
    console.log("No containing tweet article found.");
    return null;
  }

  // Find the link containing the tweet ID
  const timestampLink = container.querySelector(TIMESTAMP_SELECTOR);
  if (!timestampLink || !timestampLink.href) {
    console.log("No timestamp link with ID found.");
    return null;
  }

  // Extract the ID from the status URL
  const urlParts = timestampLink.href.split("/");
  const statusIndex = urlParts.indexOf("status");
  if (statusIndex === -1 || statusIndex + 1 >= urlParts.length) {
    console.log("Could not parse tweet ID from URL:", timestampLink.href);
    return null;
  }

  const id = urlParts[statusIndex + 1];
  if (!id) {
    console.log("Empty ID extracted.");
    return null;
  }

  // Check if we've already seen this tweet
  if (SEEN_IDS.has(id)) {
    console.log(`Tweet ID ${id} already seen.`);
    return null;
  }

  // Extract the tweet text
  const textElement = container.querySelector(CONTENT_SELECTOR);
  if (!textElement) {
    console.log("No tweet text element found.");
    return null;
  }
  const text = textElement.textContent.trim();

  // Extract the author name
  const authorElement = container.querySelector(AUTHOR_SELECTOR);
  const author = authorElement ? authorElement.textContent.trim() : "Unknown";

  // Extract timestamp, if available
  const timestamp =
    timestampLink.querySelector("time")?.getAttribute("datetime") || "";

  console.log(`Extracted new tweet for ID ${id}`);
  SEEN_IDS.add(id);
  return {
    id: id,
    text: text,
    author: author,
    timestamp: timestamp,
    url: timestampLink.href,
  };
}

function handleNode(node) {
  console.log("Handling potential tweet node:", node);
  const data = extractData(node);

  if (data) {
    // If data was extracted successfully...
    console.log("Sending tweet data to background script:", data);
    // Send a message object with a type and the data payload
    browser.runtime
      .sendMessage({ type: "SAVE_DATA", payload: data })
      .catch((error) => console.error("Error sending message:", error));
  }
}

console.log("Setting up MutationObserver...");

const observer = new MutationObserver((mutationsList) => {
  // This function runs whenever the observed DOM changes
  console.log(`DOM mutations detected: ${mutationsList.length}`);

  for (const mutation of mutationsList) {
    if (mutation.type === "childList") {
      // We are interested in nodes being added to the page
      mutation.addedNodes.forEach((node) => {
        // Check if the added node is an element node
        if (node.nodeType === Node.ELEMENT_NODE) {
          console.log("Processing added node:", node);

          // Check if the node itself matches our selector
          if (node.matches && node.matches(CONTENT_SELECTOR)) {
            handleNode(node);
          }
          // Check if the node contains any tweet text elements
          else if (node.querySelectorAll) {
            const children = node.querySelectorAll(CONTENT_SELECTOR);
            if (children.length > 0) {
              console.log(
                `Found ${children.length} matching tweets within added node.`
              );
              children.forEach((child) => {
                // For each tweet text element, find its containing article
                const tweetArticle = child.closest(
                  "article[data-testid='tweet']"
                );
                if (tweetArticle) {
                  handleNode(tweetArticle);
                }
              });
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

// Process any tweets already on the page when the script loads
console.log("Processing existing tweets on page load...");
document.querySelectorAll("article[data-testid='tweet']").forEach((article) => {
  if (article.querySelector(CONTENT_SELECTOR)) {
    handleNode(article);
  }
});

console.log("Content script setup complete.");
