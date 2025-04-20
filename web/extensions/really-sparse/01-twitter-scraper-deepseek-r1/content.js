console.log("Twitter/X content script loaded");
const SEEN_IDS = new Set();
const CONTENT_SELECTOR = 'div[data-testid="tweetText"]';

function extractText(element) {
  return Array.from(element.childNodes)
    .map((node) => {
      if (node.nodeType === Node.TEXT_NODE) return node.textContent;
      if (node.tagName === "IMG") return node.alt;
      return node.textContent;
    })
    .join(" ")
    .trim();
}

function extractData(node) {
  const container = node.closest("article");
  if (!container) return null;

  const link = container.querySelector('a[href*="/status/"]');
  if (!link) return null;

  const tweetId = link.href.split("/").pop();
  if (SEEN_IDS.has(tweetId)) return null;

  const text = extractText(node);
  if (!text) return null;

  const timestamp = container.querySelector("time")?.dateTime;
  const author = container.querySelector(
    '[data-testid="User-Name"]'
  )?.textContent;

  SEEN_IDS.add(tweetId);
  return {
    id: tweetId,
    text,
    author,
    timestamp,
    url: link.href,
  };
}

function handleNode(node) {
  const data = extractData(node);
  if (data) {
    browser.runtime
      .sendMessage({ type: "SAVE_DATA", payload: data })
      .catch((err) => console.error("Send error:", err));
  }
}

// Mutation Observer setup
const observer = new MutationObserver((mutations) => {
  for (const mutation of mutations) {
    mutation.addedNodes.forEach((node) => {
      if (node.nodeType === Node.ELEMENT_NODE) {
        if (node.matches(CONTENT_SELECTOR)) {
          handleNode(node);
        } else {
          const matches = node.querySelectorAll(CONTENT_SELECTOR);
          matches.forEach(handleNode);
        }
      }
    });
  }
});

observer.observe(document.body, {
  childList: true,
  subtree: true,
});

// Initial scan
document.querySelectorAll(CONTENT_SELECTOR).forEach(handleNode);
