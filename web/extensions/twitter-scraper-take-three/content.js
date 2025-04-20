console.log("Content script loading...");

const SEEN = new Set();
// Selector for the div containing the main text content of a tweet
const TEXT_SELECTOR = 'div[data-testid="tweetText"]';
// Selector for the main tweet container element
const ARTICLE_SELECTOR = 'article[data-testid="tweet"]';

// Extracts tweet data from a given article element
function processArticle(articleElement) {
  console.log("processArticle called for:", articleElement);

  // 1. Find the tweet's unique ID from its timestamp link within the article
  const timeElement = articleElement.querySelector("time");
  const idHref = timeElement?.parentElement?.href;
  if (!idHref) {
    console.log("No ID href found in article.", articleElement);
    return; // Cannot find the ID link
  }
  const id = idHref.split("/").pop();
  if (!id) {
    console.log("Could not parse ID from href:", idHref);
    return;
  }

  // 2. Check if we've already seen this tweet
  if (SEEN.has(id)) {
    // console.log(`Tweet ID ${id} already seen.`); // Reduce noise, uncomment if needed
    return; // Skip duplicates
  }
  console.log(`New tweet ID found: ${id}`);

  // 3. Extract the tweet text from the dedicated div
  // Use textContent which gets text from all descendants, including spans
  const textElement = articleElement.querySelector(TEXT_SELECTOR);
  const text = textElement?.textContent?.trim() || "";
  if (!text) {
    console.log(
      `No text found for tweet ID ${id} using selector ${TEXT_SELECTOR}`
    );
  }

  // 4. Extract the author's name (find the first link in the header section)
  // Look for the link within the div containing the user's name and handle
  const authorLink = articleElement.querySelector(
    'div[data-testid="User-Name"] a[role="link"]'
  );
  const author = authorLink?.textContent?.trim() || "Unknown Author";
  console.log(`Extracted author: ${author}, text: ${text.substring(0, 50)}...`);

  // 5. Mark this tweet ID as seen and prepare data
  SEEN.add(id);
  const tweetData = { id, author, text, timestamp: Date.now() };

  // 6. Send data to the background script
  console.log("Sending tweet to background:", tweetData);
  try {
    browser.runtime.sendMessage({ type: "SAVE_TWEET", tweet: tweetData });
  } catch (error) {
    console.error("Failed to send message to background script:", error);
    // Attempt to remove from SEEN if send failed? Might cause duplicates later.
    // SEEN.delete(id);
  }
}

console.log("Setting up MutationObserver...");
const observer = new MutationObserver((mutationsList) => {
  // console.log(`MutationObserver triggered with ${mutationsList.length} mutations.`); // Reduce noise
  for (const mutation of mutationsList) {
    if (mutation.type === "childList") {
      mutation.addedNodes.forEach((node) => {
        // Check if the added node is an element node (nodeType 1)
        if (node.nodeType === 1) {
          // console.log("Processing added node:", node); // Reduce noise
          // Check if the added node itself is a tweet article
          if (node.matches?.(ARTICLE_SELECTOR)) {
            console.log("Added node is an article, processing...");
            processArticle(node);
          }
          // Check if the added node contains any tweet articles
          const articles = node.querySelectorAll?.(ARTICLE_SELECTOR);
          if (articles && articles.length > 0) {
            console.log(
              `Found ${articles.length} articles within added node, processing each...`
            );
            articles.forEach(processArticle);
          }
        }
      });
    }
  }
});

// Start observing the entire body for added child elements and subtree modifications
observer.observe(document.body, { childList: true, subtree: true });
console.log("MutationObserver observing document.body.");

// --- Initial Scan ---
// Use setTimeout to allow the page structure to potentially settle slightly after load
console.log("Scheduling initial scan for existing tweets...");
setTimeout(() => {
  console.log("Running initial scan for existing tweets...");
  document.querySelectorAll(ARTICLE_SELECTOR).forEach(processArticle);
  console.log("Initial scan complete. Content script setup finished.");
}, 1000); // Wait 1 second after script load before scanning
