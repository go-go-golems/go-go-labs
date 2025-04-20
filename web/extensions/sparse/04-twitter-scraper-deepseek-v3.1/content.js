const SEEN = new Set();
const TEXT_SELECTOR = 'div[data-testid="tweetText"]';

function extractTweet(node) {
  const article = node.closest("article");
  if (!article) return null;

  const idHref = article.querySelector("time")?.parentElement?.href;
  if (!idHref) return null;
  const id = idHref.split("/").pop();

  if (SEEN.has(id)) return null;

  const text = [...article.querySelectorAll(TEXT_SELECTOR + " span")]
    .map((el) => el.textContent)
    .join("");

  const author = article.querySelector('a[role="link"] span')?.textContent || "";

  SEEN.add(id);
  return { id, author, text };
}

function handle(node) {
  const tweet = extractTweet(node);
  if (tweet) {
    browser.runtime.sendMessage({ type: "SAVE_TWEET", tweet });
  }
}

const observer = new MutationObserver((muts) => {
  for (const m of muts) {
    m.addedNodes.forEach((n) => {
      if (n.nodeType === 1) {
        if (n.matches?.(TEXT_SELECTOR)) {
          handle(n);
        }
        const children = n.querySelectorAll?.(TEXT_SELECTOR);
        if (children && children.length > 0) {
          children.forEach(handle);
        }
      }
    });
  }
});

observer.observe(document.body, { childList: true, subtree: true });
document.querySelectorAll(TEXT_SELECTOR).forEach(handle); 