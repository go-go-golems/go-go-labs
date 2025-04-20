console.log("Content script loading...");

const SEEN = new Set();
const TEXT_SELECTOR = 'div[data-testid="tweetText"]';

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

// Handle initially loaded tweets
console.log("Handling existing tweets on page load...");
document.querySelectorAll(TEXT_SELECTOR).forEach(handle);
console.log("Content script loaded and initial handling complete.");
