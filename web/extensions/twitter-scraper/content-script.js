/*  ╭─────────────  Configure what we want to pull  ─────────────╮ */
function extractTweet(article) {
  // 1️⃣ canonical tweet URL & ID
  const statusLink = [...article.querySelectorAll("a")].find((a) =>
    /\/status\/\d+/.test(a.href)
  );
  if (!statusLink) return null; // safety
  const url = statusLink.href;
  const id = url.match(/status\/(\d+)/)[1];

  // 2️⃣ author (username + display name if present)
  const authorLink = [...article.querySelectorAll("a")].find((a) =>
    /^https?:\/\/(?:twitter|x)\.com\/[^\/?]+$/.test(a.href)
  );
  const username = authorLink ? authorLink.href.split("/").pop() : null;
  const author = authorLink?.innerText || username;

  // 3️⃣ full tweet text (handles multiple blocks & emojis)
  const textEls = article.querySelectorAll('[data-testid="tweetText"]');
  const text = [...textEls].map((el) => el.innerText).join("\n");

  // 4️⃣ ISO timestamp
  const ts = article.querySelector("time")?.getAttribute("datetime") || null;

  return { id, url, author, username, text, timestamp: ts };
}

/*  ╭─────────────  Live capture as you keep scrolling  ─────────╮ */
window.__seenTweets = new Set();

function handleNewArticle(article) {
  const t = extractTweet(article);
  if (!t || window.__seenTweets.has(t.id)) return;
  window.__seenTweets.add(t.id);
  console.log("[Content Script] Found and sending tweet:", t.id, t.author);
  // Send the tweet to the background script instead of just logging
  browser.runtime
    .sendMessage({ type: "SAVE_TWEET", tweet: t })
    .catch((err) =>
      console.error("[Content Script] Error sending tweet:", err)
    );
}

const observer = new MutationObserver((muts) => {
  muts.forEach((m) => {
    m.addedNodes.forEach((node) => {
      if (node.nodeType !== 1) return;
      // Check if the node itself is a tweet
      if (node.matches?.('article[data-testid="tweet"]'))
        handleNewArticle(node);
      // Check if the node contains tweets (e.g., when a container is added)
      node
        .querySelectorAll?.('article[data-testid="tweet"]')
        .forEach(handleNewArticle);
    });
  });
});

console.log("[Content Script] Initializing observer...");
observer.observe(document, { childList: true, subtree: true });
console.log("[Content Script] Observer attached.");
