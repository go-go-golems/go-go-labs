// Function to display tweets in the popup UI
async function displayTweets() {
  console.log("Displaying tweets in popup...");
  const { tweets = [] } = await browser.storage.local.get("tweets");
  const tweetList = document.getElementById("tweet-list");
  const tweetCount = document.getElementById("tweet-count");
  const authorCount = document.getElementById("author-count");

  // Update tweet count
  tweetCount.textContent = tweets.length;
  console.log(`Total tweets to display: ${tweets.length}`);

  // Calculate unique authors
  const uniqueAuthors = new Set(tweets.map((tweet) => tweet.author)).size;
  authorCount.textContent = uniqueAuthors;
  console.log(`Unique authors: ${uniqueAuthors}`);

  // Clear the list
  tweetList.innerHTML = "";

  if (tweets.length === 0) {
    // Show empty state message
    console.log("Tweet list is empty, showing empty state.");
    const emptyState = document.createElement("div");
    emptyState.className = "empty-state";
    emptyState.textContent =
      "No tweets collected yet. Scroll through Twitter to collect tweets.";
    tweetList.appendChild(emptyState);
    return;
  }

  // Add each tweet to the list (most recent first)
  tweets
    .slice() // Create a copy to avoid modifying the original array
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
  console.log("Tweet list populated.");
}

// Initial display when popup opens
document.addEventListener("DOMContentLoaded", () => {
  console.log("Popup DOM loaded.");
  displayTweets();
});

// Set up listener for changes in storage (live updates)
browser.storage.onChanged.addListener((changes, area) => {
  console.log("Storage change detected:", changes, area);
  if (area === "local" && changes.tweets) {
    console.log("Tweets storage changed, updating display.");
    displayTweets();
  }
});

// Set up download button
document.getElementById("download").addEventListener("click", () => {
  console.log("Download button clicked, sending message to background.");
  browser.runtime.sendMessage({ type: "DOWNLOAD_TWEETS" });
});

console.log("Popup script loaded.");
