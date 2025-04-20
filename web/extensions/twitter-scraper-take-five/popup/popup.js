console.log("Popup script loading...");

const tweetList = document.getElementById("tweet-list");
const tweetCountEl = document.getElementById("tweet-count");
const authorCountEl = document.getElementById("author-count");
const emptyStateEl = document.getElementById("empty-state");
const downloadButton = document.getElementById("download");

// Function to display tweets in the popup UI
async function displayTweets() {
  console.log("Displaying tweets...");
  try {
    const { tweets = [] } = await browser.storage.local.get("tweets");
    console.log(`Loaded ${tweets.length} tweets from storage.`);

    // Update tweet count
    tweetCountEl.textContent = tweets.length;

    // Calculate unique authors
    const uniqueAuthors = new Set(tweets.map((tweet) => tweet.author)).size;
    authorCountEl.textContent = uniqueAuthors;

    // Clear the list first
    tweetList.innerHTML = "";

    if (tweets.length === 0) {
      // Show empty state message (clone it to ensure it's added back cleanly)
      tweetList.appendChild(emptyStateEl.cloneNode(true));
      downloadButton.disabled = true; // Disable download if no tweets
      console.log("No tweets to display, showing empty state.");
      return;
    }

    // Re-enable download button if there are tweets
    downloadButton.disabled = false;

    // Add each tweet to the list (most recent first)
    tweets
      .slice()
      .reverse()
      .forEach((tweet) => {
        const tweetElement = document.createElement("div");
        tweetElement.className = "tweet";

        const authorElement = document.createElement("div");
        authorElement.className = "author";
        authorElement.textContent = tweet.author || "[No Author]"; // Handle missing author

        const textElement = document.createElement("p");
        textElement.className = "text";
        textElement.textContent = tweet.text || "[No Text]"; // Handle missing text

        tweetElement.appendChild(authorElement);
        tweetElement.appendChild(textElement);
        tweetList.appendChild(tweetElement);
      });
    console.log("Finished displaying tweets.");
  } catch (error) {
    console.error("Error displaying tweets:", error);
    tweetList.innerHTML =
      '<div class="error-state">Error loading tweets.</div>';
  }
}

// Initial display when popup opens
document.addEventListener("DOMContentLoaded", () => {
  console.log("DOM content loaded, initial display.");
  displayTweets();
});

// Set up listener for changes in storage (live updates)
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.tweets) {
    console.log("Storage changed, updating display.");
    displayTweets();
  }
});

// Set up download button
downloadButton.addEventListener("click", () => {
  console.log("Download button clicked, sending message to background script.");
  browser.runtime.sendMessage({ type: "DOWNLOAD_TWEETS" });
});

console.log("Popup script loaded.");
