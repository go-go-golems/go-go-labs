async function displayTweets() {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  const tweetList = document.getElementById("tweet-list");
  const tweetCount = document.getElementById("tweet-count");
  const authorCount = document.getElementById("author-count");

  tweetCount.textContent = tweets.length;
  authorCount.textContent = new Set(tweets.map((t) => t.author)).size;

  tweetList.innerHTML = "";
  if (!tweets.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "No tweets collected yet. Scroll Twitter to start.";
    tweetList.appendChild(empty);
    return;
  }

  tweets
    .slice()
    .reverse()
    .forEach((tweet) => {
      const el = document.createElement("div");
      el.className = "tweet";
      const auth = document.createElement("div");
      auth.className = "author";
      auth.textContent = tweet.author;
      const txt = document.createElement("p");
      txt.className = "text";
      txt.textContent = tweet.text;
      el.append(auth, txt);
      tweetList.appendChild(el);
    });
}

document.addEventListener("DOMContentLoaded", displayTweets);

browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.tweets) {
    displayTweets();
  }
});

document.getElementById("download").addEventListener("click", () => {
  browser.runtime.sendMessage({ type: "DOWNLOAD_TWEETS" });
});
