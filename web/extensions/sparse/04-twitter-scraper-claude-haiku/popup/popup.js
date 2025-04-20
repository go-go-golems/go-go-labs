// Function to display tweets in the popup UI
async function displayTweets() {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  const tweetList = document.getElementById("tweet-list");
  const tweetCount = document.getElementById("tweet-count");
  const authorCount = document.getElementById("author-count");

  // Update tweet count
  tweetCount.textContent = tweets.length;

  // Calculate unique authors
  const uniqueAuthors = new Set(tweets.map((tweet) => tweet.author)).size;
  authorCount.textContent = uniqueAuthors;

  // Clear the list
  tweetList.innerHTML = "";

  if (tweets.length === 0) {
    // Show empty state message
    const emptyState = document.createElement("div");
    emptyState.className = "empty-state";
    emptyState.textContent =
      "No tweets collected yet. Scroll through Twitter to collect tweets.";
    tweetList.appendChild(emptyState);
    return;
  }

  // Add each tweet to the list (most recent first)
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

// Initial display when popup opens
document.addEventListener("DOMContentLoaded", displayTweets);

// Set up listener for changes in storage (live updates)
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.tweets) {
    displayTweets();
  }
});

// Set up download button
document.getElementById("download").addEventListener("click", () => {
  browser.runtime.sendMessage({ type: "DOWNLOAD_TWEETS" });
});
