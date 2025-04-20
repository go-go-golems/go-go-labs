console.log("Content script loading...");

const SEEN = new Set();
const TEXT_SELECTOR = 'div[data-testid="tweetText"]';

function extractTweet(node) {
  console.log("extractTweet called for node:", node);
  const article = node.closest("article");
  if (!article) {
    console.log("No article found for node.");
    return null;
  }
  console.log("Found article:", article);

  const idHref = article.querySelector("time")?.parentElement?.href;
  if (!idHref) {
    console.log("No ID href found in article.");
    return null;
  }
  const id = idHref.split("/").pop();

  if (SEEN.has(id)) {
    console.log(`Tweet ID ${id} already seen.`);
    return null;
  }
  console.log(`New tweet ID found: ${id}`);

  const text = [...article.querySelectorAll(TEXT_SELECTOR + " span")]
    .map((el) => el.textContent)
    .join("");

  const author =
    article.querySelector('a[role="link"] span')?.textContent || "";
  console.log(`Extracted author: ${author}, text: ${text.substring(0, 50)}...`);

  SEEN.add(id);
  return { id, author, text };
}

function handle(node) {
  console.log("Handling node:", node);
  const tweet = extractTweet(node);
  if (tweet) {
    console.log("Sending tweet to background:", tweet);
    browser.runtime.sendMessage({ type: "SAVE_TWEET", tweet });
  } else {
    console.log("No tweet extracted from node or already seen.");
  }
}

console.log("Setting up MutationObserver...");
const observer = new MutationObserver((muts) => {
  console.log(`MutationObserver triggered with ${muts.length} mutations.`);
  for (const m of muts) {
    m.addedNodes.forEach((n) => {
      if (n.nodeType === 1) {
        console.log("Processing added node:", n);
        if (n.matches?.(TEXT_SELECTOR)) {
          console.log("Node matches TEXT_SELECTOR, handling...");
          handle(n);
        }
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

observer.observe(document.body, { childList: true, subtree: true });
console.log("MutationObserver observing document.body.");

console.log("Handling existing tweets on page load...");
document.querySelectorAll(TEXT_SELECTOR).forEach(handle);
console.log("Content script loaded and initial handling complete.");
