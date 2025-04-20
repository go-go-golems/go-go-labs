// ... existing code ...
console.log("Content script starting...");

const SEEN_IDS = new Set();
const CONTENT_SELECTOR = 'div[data-testid="tweetText"]';

function extractData(node) {
  console.log("Attempting to extract data from node:", node);
  const container = node.closest('article');
  if (!container) return null;

  const link = container.querySelector('a[href*="/status/"]');
  const idLink = link?.href;
  if (!idLink) return null;
  const id = idLink.split('/').pop();

  if (SEEN_IDS.has(id)) {
    console.log(`ID ${id} already seen.`);
    return null;
  }

  const textElement = container.querySelector(CONTENT_SELECTOR);
  const text = textElement?.textContent.trim() || "";

  console.log(`Extracted new data for ID ${id}`);
  SEEN_IDS.add(id);
  return { id, text };
}

function handleNode(node) {
  console.log("Handling potential data node:", node);
  const data = extractData(node);
  if (data) {
    console.log("Sending data to background script:", data);
    browser.runtime
      .sendMessage({ type: "SAVE_DATA", payload: data })
      .catch((error) => console.error("Error sending message:", error));
  }
}

console.log("Setting up MutationObserver...");
const observer = new MutationObserver((mutationsList) => {
  for (const mutation of mutationsList) {
    if (mutation.type === "childList") {
      mutation.addedNodes.forEach((node) => {
        if (node.nodeType === Node.ELEMENT_NODE) {
          if (node.matches && node.matches(CONTENT_SELECTOR)) {
            handleNode(node);
          } else if (node.querySelectorAll) {
            node.querySelectorAll(CONTENT_SELECTOR).forEach(handleNode);
          }
        }
      });
    }
  }
});

observer.observe(document.body, { childList: true, subtree: true });
console.log("MutationObserver is now observing the document body.");

console.log("Processing existing content on page load...");
document.querySelectorAll(CONTENT_SELECTOR).forEach(handleNode);

console.log("Content script setup complete.");
// ... existing code ...