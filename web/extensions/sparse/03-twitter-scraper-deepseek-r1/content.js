const SEEN_IDS = new Set();
const SELECTORS = {
  container: 'article[data-testid="tweet"]',
  content: 'div[data-testid="tweetText"]',
  id: 'a[href*="/status/"]',
  author: 'div[data-testid="User-Name"] a[href^="/"]:last-child',
};

function handleNode(node) {
  const data = extractData(node);
  if (data) {
    browser.runtime
      .sendMessage({
        type: "SAVE_DATA",
        payload: data,
      })
      .catch(console.error);
  }
}

function extractData(node) {
  const container = node.closest(SELECTORS.container);
  if (!container) return null;

  // Extract tweet ID
  const idLink = container.querySelector(SELECTORS.id)?.href;
  const tweetId = idLink?.match(/\/status\/(\d+)/)?.[1];
  if (!tweetId || SEEN_IDS.has(tweetId)) return null;

  // Extract content
  const text = Array.from(
    container.querySelector(SELECTORS.content)?.querySelectorAll("span") || []
  )
    .map((span) => span.textContent)
    .join(" ")
    .trim();

  // Extract author handle
  const authorLink = container.querySelector(SELECTORS.author)?.href;
  const author = authorLink?.split("/").pop() || "unknown";

  SEEN_IDS.add(tweetId);
  return {
    id: tweetId,
    text,
    author,
    timestamp: new Date().toISOString(),
    url: window.location.href,
  };
}

// Mutation Observer Setup
const observer = new MutationObserver((mutations) => {
  for (const mutation of mutations) {
    if (mutation.type === "childList") {
      mutation.addedNodes.forEach((node) => {
        if (node.nodeType === Node.ELEMENT_NODE) {
          const containers = node.matches(SELECTORS.container)
            ? [node]
            : [...node.querySelectorAll(SELECTORS.container)];

          containers.forEach((container) => {
            const contentNode = container.querySelector(SELECTORS.content);
            if (contentNode) handleNode(contentNode);
          });
        }
      });
    }
  }
});

// Initialization
document.addEventListener("readystatechange", () => {
  if (document.readyState === "complete") {
    observer.observe(document.body, {
      childList: true,
      subtree: true,
    });

    // Process existing tweets
    document.querySelectorAll(SELECTORS.container).forEach((container) => {
      const content = container.querySelector(SELECTORS.content);
      if (content) handleNode(content);
    });
  }
});
