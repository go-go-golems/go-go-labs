// Store filtered tweet list
let filteredTweets = [];

// Apply search filter to tweets
function applySearch(tweets) {
  const searchInput = document.getElementById("search-input");
  const searchText = searchInput.value.toLowerCase();

  if (!searchText) {
    return tweets; // Return all tweets if search is empty
  }

  return tweets.filter((tweet) => {
    // Search in both text and author
    return (
      tweet.text.toLowerCase().includes(searchText) ||
      tweet.author.toLowerCase().includes(searchText)
    );
  });
}

// Set up search functionality
function setupSearch() {
  const searchInput = document.getElementById("search-input");

  searchInput.addEventListener("input", () => {
    displayTweets(); // Refresh the display when search input changes
  });
}

// Set up tab navigation
function setupTabs() {
  const tabs = document.querySelectorAll(".tab");
  const tabContents = document.querySelectorAll(".tab-content");

  tabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      // Remove active class from all tabs and contents
      tabs.forEach((t) => t.classList.remove("active"));
      tabContents.forEach((c) => c.classList.remove("active"));

      // Add active class to clicked tab and corresponding content
      tab.classList.add("active");
      const tabName = tab.getAttribute("data-tab");
      document.getElementById(`${tabName}-tab`).classList.add("active");
    });
  });
}

// Convert tweets to CSV format
function tweetsToCSV(tweets) {
  const headers = ["id", "author", "text", "timestamp"];
  let csv = headers.join(",") + "\n";

  tweets.forEach((tweet) => {
    const row = [
      tweet.id,
      `"${tweet.author.replace(/"/g, '""')}"`, // Handle quotes in author
      `"${tweet.text.replace(/"/g, '""')}"`, // Handle quotes in text
      tweet.timestamp || "", // Add timestamp
    ];
    csv += row.join(",") + "\n";
  });

  return csv;
}

// Convert tweets to plain text format
function tweetsToText(tweets) {
  let text = `Total Tweets: ${tweets.length}\n\n`;

  tweets.forEach((tweet) => {
    text += `Author: @${tweet.author}\n`;
    text += `Tweet: ${tweet.text}\n`;
    if (tweet.timestamp) {
      const date = new Date(tweet.timestamp);
      text += `Time: ${date.toLocaleString()}\n`; // Format timestamp
    }
    text += "--------------------\n\n";
  });

  return text;
}

// Set up export buttons
function setupExportButtons() {
  const jsonExport = document.getElementById("json-export");
  const csvExport = document.getElementById("csv-export");
  const txtExport = document.getElementById("txt-export");
  const clearData = document.getElementById("clear-data");

  jsonExport.addEventListener("click", () => {
    exportTweets("json");
  });

  csvExport.addEventListener("click", () => {
    exportTweets("csv");
  });

  txtExport.addEventListener("click", () => {
    exportTweets("txt");
  });

  clearData.addEventListener("click", () => {
    if (
      confirm(
        "Are you sure you want to clear all collected tweets? This cannot be undone."
      )
    ) {
      browser.storage.local
        .remove("tweets") // Use remove for clarity
        .then(() => {
          console.log("Tweet data cleared.");
          displayTweets(); // Refresh UI to show empty state
          alert("All tweet data has been cleared!");
        })
        .catch((err) => {
          console.error("Error clearing data:", err);
          alert("Failed to clear tweet data.");
        });
    }
  });
}

// Export tweets in selected format
async function exportTweets(format) {
  const exportScope = document.getElementById("export-scope").value;
  let { tweets = [] } = await browser.storage.local.get("tweets");

  // Use filtered tweets if in filtered mode and search is active
  const searchInput = document.getElementById("search-input");
  if (exportScope === "filtered" && searchInput.value) {
    tweets = filteredTweets; // Use the already filtered list
  } else if (exportScope === "filtered" && !searchInput.value) {
    // If 'filtered' is selected but search is empty, export all
    // Or alert the user? Let's export all for simplicity.
    console.log(
      "Exporting all tweets as search is empty despite 'filtered' scope."
    );
  }

  if (tweets.length === 0) {
    alert("No tweets to export!");
    return;
  }

  let data, mimeType, filename;
  const timestamp = new Date().toISOString().replace(/:/g, "-");

  switch (format) {
    case "csv":
      data = tweetsToCSV(tweets);
      mimeType = "text/csv";
      filename = `tweets_${timestamp}.csv`;
      break;
    case "txt":
      data = tweetsToText(tweets);
      mimeType = "text/plain";
      filename = `tweets_${timestamp}.txt`;
      break;
    case "json":
    default:
      data = JSON.stringify(tweets, null, 2);
      mimeType = "application/json";
      filename = `tweets_${timestamp}.json`;
  }

  // Send data to background script for download
  browser.runtime
    .sendMessage({
      type: "DOWNLOAD_DATA",
      data,
      mimeType,
      filename,
    })
    .catch((err) => {
      console.error("Error sending download message:", err);
      alert("Failed to initiate download.");
    });
}

// Format timestamp for display
function formatTime(timestamp) {
  if (!timestamp) return "";

  const now = Date.now();
  const diff = now - timestamp;

  // Less than a minute
  if (diff < 60000) {
    return "just now";
  }
  // Less than an hour
  else if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000);
    return `${minutes}m ago`;
  }
  // Less than a day
  else if (diff < 86400000) {
    const hours = Math.floor(diff / 3600000);
    return `${hours}h ago`;
  }
  // Otherwise show date
  else {
    const date = new Date(timestamp);
    // Use locale-specific date format
    return date.toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }
}

// Function to display tweets in the popup UI (with optimizations)
async function displayTweets() {
  try {
    const { tweets = [] } = await browser.storage.local.get("tweets");
    const tweetList = document.getElementById("tweet-list");
    const tweetCount = document.getElementById("tweet-count");
    const authorCount = document.getElementById("author-count");

    // Apply search filter
    // We store the filtered list globally for export purposes
    filteredTweets = applySearch(tweets);

    // Update counts (only if they've changed to reduce minor DOM updates)
    if (tweetCount.textContent !== tweets.length.toString()) {
      tweetCount.textContent = tweets.length;
    }

    // Calculate unique authors from the original list
    const uniqueAuthors = new Set(tweets.map((tweet) => tweet.author)).size;
    if (authorCount.textContent !== uniqueAuthors.toString()) {
      authorCount.textContent = uniqueAuthors;
    }

    // Create a document fragment to build the new list (improves performance)
    const fragment = document.createDocumentFragment();

    if (filteredTweets.length === 0) {
      // Show appropriate empty state message
      const emptyState = document.createElement("div");
      emptyState.className = "empty-state";
      const searchInput = document.getElementById("search-input");
      if (tweets.length === 0) {
        emptyState.textContent =
          "No tweets collected yet. Scroll through Twitter to collect tweets.";
      } else if (searchInput.value) {
        emptyState.textContent = "No tweets match your search.";
      } else {
        // Should not happen if tweets.length > 0 and search is empty, but good fallback
        emptyState.textContent = "No tweets to display.";
      }
      fragment.appendChild(emptyState);
    } else {
      // Add each filtered tweet to the fragment (most recent first)
      filteredTweets
        .slice()
        .reverse()
        .forEach((tweet) => {
          const tweetElement = document.createElement("div");
          tweetElement.className = "tweet";

          // Create tweet header with author and time
          const tweetHeader = document.createElement("div");
          tweetHeader.className = "tweet-header";

          const authorElement = document.createElement("div");
          authorElement.className = "author";
          authorElement.textContent = tweet.author || "[unknown author]"; // Handle missing author

          // Add time if available
          const timeElement = document.createElement("div");
          timeElement.className = "tweet-time";
          timeElement.textContent = formatTime(tweet.timestamp);

          tweetHeader.appendChild(authorElement);
          tweetHeader.appendChild(timeElement);

          const textElement = document.createElement("p");
          textElement.className = "text";
          textElement.textContent = tweet.text || "[no text]"; // Handle missing text

          tweetElement.appendChild(tweetHeader);
          tweetElement.appendChild(textElement);
          fragment.appendChild(tweetElement);
        });
    }

    // Only clear and update the DOM once to prevent flickering
    tweetList.innerHTML = ""; // Clear existing content efficiently
    tweetList.appendChild(fragment);
  } catch (error) {
    console.error("Error displaying tweets:", error);
    // Optionally display an error message in the UI
    const tweetList = document.getElementById("tweet-list");
    if (tweetList) {
      tweetList.innerHTML =
        '<div class="empty-state">Error loading tweets.</div>';
    }
  }
}

// Initial setup when popup opens
document.addEventListener("DOMContentLoaded", () => {
  console.log("Popup DOM loaded.");
  setupTabs();
  setupSearch();
  setupExportButtons();
  displayTweets(); // Initial display
});

// Set up listener for changes in storage (live updates)
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.tweets) {
    console.log("Storage changed, updating display...");
    displayTweets(); // Refresh display on data change
  }
});

console.log("Popup script loaded.");
