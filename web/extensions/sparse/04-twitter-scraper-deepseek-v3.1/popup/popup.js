async function displayTweets() {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  const tweetList = document.getElementById("tweet-list");
  const tweetCount = document.getElementById("tweet-count");
  const authorCount = document.getElementById("author-count");

  tweetCount.textContent = tweets.length;
  const uniqueAuthors = new Set(tweets.map((tweet) => tweet.author)).size;
  authorCount.textContent = uniqueAuthors;

  tweetList.innerHTML = "";

  if (tweets.length === 0) {
    const emptyState = document.createElement("div");
    emptyState.className = "empty-state";
    emptyState.textContent =
      "No tweets collected yet. Scroll through Twitter to collect tweets.";
    tweetList.appendChild(emptyState);
    return;
  }

  tweets
    .slice()
    .reverse()
    .forEach((tweet) => {
      const tweetElement = document.createElement("div");
      tweetElement.className = "tweet";

      const authorElement = document.createElement("div");
      authorElement.className = "author";
      authorElement.textContent = tweet.author;

      const textElement = document.createElement("p");
      textElement.className = "text";
      textElement.textContent = tweet.text;

      tweetElement.appendChild(authorElement);
      tweetElement.appendChild(textElement);
      tweetList.appendChild(tweetElement);
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
