const SEEN = new Set();
const TEXT_SELECTOR = 'div[data-testid="tweetText"]';

function extractTweet(node) {
  const article = node.closest("article");
  if (!article) return null;

  const idHref = article.querySelector("time")?.parentElement?.href;
  if (!idHref) return null;
  const id = idHref.split("/").pop();

  if (SEEN.has(id)) return null;

  const text = [...article.querySelectorAll(`${TEXT_SELECTOR} span`)]
    .map((el) => el.textContent)
    .join("");
  const author =
    article.querySelector('a[role="link"] span')?.textContent || "";

  SEEN.add(id);
  return { id, author, text, timestamp: new Date().toISOString() };
}

function handle(node) {
  const tweet = extractTweet(node);
  if (tweet) {
    browser.runtime.sendMessage({ type: "SAVE_TWEET", tweet });
  }
}

// MutationObserver setup
const observer = new MutationObserver((mutations) => {
  mutations.forEach((m) => {
    m.addedNodes.forEach((n) => {
      if (n.nodeType === 1) {
        if (n.matches?.(TEXT_SELECTOR)) handle(n);
        n.querySelectorAll?.(TEXT_SELECTOR).forEach(handle);
      }
    });
  });
});

observer.observe(document.body, { childList: true, subtree: true });

// Initial scan
document.querySelectorAll(TEXT_SELECTOR).forEach(handle);
