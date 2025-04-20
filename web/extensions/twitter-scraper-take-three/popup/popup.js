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

// Animate count number when it changes
function animateCountChange(element, oldValue, newValue) {
  if (oldValue !== newValue) {
    element.classList.add("pulse");
    setTimeout(() => {
      element.classList.remove("pulse");
    }, 500);
  }
}

// Add "NEW" badge to recently added tweets
function addNewBadge(tweetElement, isRecent) {
  if (isRecent) {
    const badge = document.createElement("span");
    badge.className = "funky-badge new-badge";
    badge.textContent = "NEW";
    tweetElement.appendChild(badge);

    // Remove the badge after 3 seconds
    setTimeout(() => {
      badge.style.opacity = "0";
      badge.style.transition = "opacity 0.5s";
      setTimeout(() => badge.remove(), 500);
    }, 3000);
  }
}

// Format timestamp
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
    return date.toLocaleDateString();
  }
}

// Convert tweets to CSV format
function tweetsToCSV(tweets) {
  const headers = [
    "id",
    "author",
    "text",
    "timestamp",
    "date",
    "url",
    "replies",
    "reposts",
    "likes",
    "views",
    "bookmarks",
  ];
  let csv = headers.join(",") + "\n";

  tweets.forEach((tweet) => {
    const row = [
      tweet.id,
      `"${tweet.author?.replace(/"/g, '""') || ""}"`,
      `"${tweet.text?.replace(/"/g, '""') || ""}"`,
      tweet.timestamp || "",
      `"${tweet.date || ""}"`,
      `"${tweet.url || ""}"`,
      tweet.replies || "0",
      tweet.reposts || "0",
      tweet.likes || "0",
      tweet.views || "0",
      tweet.bookmarks || "0",
    ];
    csv += row.join(",") + "\n";
  });

  return csv;
}

// Convert tweets to plain text format
function tweetsToText(tweets) {
  let text = `Total Tweets: ${tweets.length}\n\n`;

  tweets.forEach((tweet) => {
    text += `Author: @${tweet.author || "Unknown"}\n`;
    text += `Tweet: ${tweet.text || ""}\n`;
    if (tweet.date) {
      text += `Date: ${tweet.date}\n`;
    }
    if (tweet.url) {
      text += `URL: ${tweet.url}\n`;
    }
    if (tweet.replies || tweet.reposts || tweet.likes) {
      text += `Stats: ${tweet.replies || 0} replies, ${
        tweet.reposts || 0
      } reposts, ${tweet.likes || 0} likes`;
      if (tweet.views) {
        text += `, ${tweet.views} views`;
      }
      if (tweet.bookmarks) {
        text += `, ${tweet.bookmarks} bookmarks`;
      }
      text += "\n";
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
    if (confirm("Are you sure you want to clear all collected tweets?")) {
      browser.storage.local.set({ tweets: [] });
      alert("All tweet data has been cleared!");
    }
  });
}

// Export tweets in selected format
async function exportTweets(format) {
  const exportScope = document.getElementById("export-scope").value;
  let { tweets = [] } = await browser.storage.local.get("tweets");

  // Use filtered tweets if in filtered mode
  if (exportScope === "filtered") {
    tweets = filteredTweets;
  }

  if (tweets.length === 0) {
    alert("No tweets to export!");
    return;
  }

  let data, mimeType, filename;

  switch (format) {
    case "csv":
      data = tweetsToCSV(tweets);
      mimeType = "text/csv";
      filename = `tweets_${new Date().toISOString().replace(/:/g, "-")}.csv`;
      break;
    case "txt":
      data = tweetsToText(tweets);
      mimeType = "text/plain";
      filename = `tweets_${new Date().toISOString().replace(/:/g, "-")}.txt`;
      break;
    case "json":
    default:
      data = JSON.stringify(tweets, null, 2);
      mimeType = "application/json";
      filename = `tweets_${new Date().toISOString().replace(/:/g, "-")}.json`;
  }

  browser.runtime.sendMessage({
    type: "DOWNLOAD_DATA",
    data,
    mimeType,
    filename,
  });
}

// Function to display tweets in the popup UI - optimized to reduce flickering
async function displayTweets() {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  const tweetList = document.getElementById("tweet-list");
  const tweetCount = document.getElementById("tweet-count");
  const authorCount = document.getElementById("author-count");

  // Store previous values for animation
  const prevTweetCount = parseInt(tweetCount.textContent) || 0;
  const prevAuthorCount = parseInt(authorCount.textContent) || 0;

  // Apply search filter
  filteredTweets = applySearch(tweets);

  // Update counts (only if they've changed)
  if (tweetCount.textContent !== tweets.length.toString()) {
    tweetCount.textContent = tweets.length;
    animateCountChange(tweetCount, prevTweetCount, tweets.length);
  }

  // Calculate unique authors
  const uniqueAuthors = new Set(tweets.map((tweet) => tweet.author)).size;
  if (authorCount.textContent !== uniqueAuthors.toString()) {
    authorCount.textContent = uniqueAuthors;
    animateCountChange(authorCount, prevAuthorCount, uniqueAuthors);
  }

  // Create a document fragment to build the new list (improves performance)
  const fragment = document.createDocumentFragment();

  // Check if we're in desktop mode - show more tweets in that case
  const isDesktopView =
    document.title.includes("Desktop") ||
    document.querySelector(".desktop-layout") !== null;

  if (filteredTweets.length === 0) {
    // Show empty state message
    const emptyState = document.createElement("div");
    emptyState.className = "empty-state";

    if (tweets.length === 0) {
      emptyState.textContent =
        "No tweets collected yet. Scroll through Twitter to collect tweets.";
    } else {
      emptyState.textContent = "No tweets match your search.";
    }

    fragment.appendChild(emptyState);
  } else {
    // Get timestamp from 30 seconds ago to mark recent tweets
    const recentTimestamp = Date.now() - 30000;
    // Keep track of tweet IDs we've already seen in this popup view
    const viewedTweetIds = new Set();

    // Add each tweet to the fragment (most recent first)
    // In desktop mode, show more tweets
    const tweetsToShow = filteredTweets.slice().reverse();
    // For desktop view, we show all tweets, for popup we limit to latest 50
    const limitedTweets = isDesktopView
      ? tweetsToShow
      : tweetsToShow.slice(0, 50);

    limitedTweets.forEach((tweet, index) => {
      const tweetElement = document.createElement("div");
      tweetElement.className = "tweet";

      // Create tweet header with author and time
      const tweetHeader = document.createElement("div");
      tweetHeader.className = "tweet-header";

      const authorElement = document.createElement("div");
      authorElement.className = "author";
      authorElement.textContent = tweet.author;

      // Add time if available
      const timeElement = document.createElement("div");
      timeElement.className = "tweet-time";
      timeElement.textContent = formatTime(tweet.timestamp);
      if (tweet.date) {
        timeElement.title = tweet.date; // Show full date on hover
      }

      tweetHeader.appendChild(authorElement);
      tweetHeader.appendChild(timeElement);

      const textElement = document.createElement("p");
      textElement.className = "text";
      textElement.textContent = tweet.text;

      // Add source link if available
      if (tweet.url) {
        const linkElement = document.createElement("a");
        linkElement.href = tweet.url;
        linkElement.className = "tweet-link";
        linkElement.textContent = "View on Twitter";
        linkElement.target = "_blank";
        linkElement.rel = "noopener noreferrer";
        textElement.appendChild(document.createElement("br"));
        textElement.appendChild(linkElement);
      }

      // Add stats if available
      const statsElement = document.createElement("div");
      statsElement.className = "tweet-stats";
      const stats = [];
      if (tweet.replies) stats.push(`💬 ${tweet.replies}`);
      if (tweet.reposts) stats.push(`🔁 ${tweet.reposts}`);
      if (tweet.likes) stats.push(`❤️ ${tweet.likes}`);
      if (tweet.views) stats.push(`👁️ ${tweet.views}`);
      if (tweet.bookmarks) stats.push(`🔖 ${tweet.bookmarks}`);
      if (stats.length > 0) {
        statsElement.textContent = stats.join("  ");
        textElement.appendChild(statsElement);
      }

      tweetElement.appendChild(tweetHeader);
      tweetElement.appendChild(textElement);

      // Check if this is a new tweet (recently seen)
      const isNewTweet =
        !viewedTweetIds.has(tweet.id) &&
        tweet.timestamp &&
        tweet.timestamp > recentTimestamp;

      // Add badge for recent tweets
      addNewBadge(tweetElement, isNewTweet || (!tweet.timestamp && index < 3));

      // Mark as viewed
      viewedTweetIds.add(tweet.id);

      fragment.appendChild(tweetElement);
    });
  }

  // Only clear and update the DOM once
  tweetList.innerHTML = "";
  tweetList.appendChild(fragment);
}

// Add a pop out button to the popup
function setupPopoutButton() {
  const popoutBtn = document.getElementById("popout-btn");
  if (popoutBtn) {
    popoutBtn.addEventListener("click", () => {
      browser.tabs.create({
        url: browser.runtime.getURL("popup/desktop.html"),
      });
    });
  }
}

// Initial setup when popup opens
document.addEventListener("DOMContentLoaded", () => {
  // Remove the detection logic since we'll use separate HTML files
  setupTabs();
  setupSearch();
  setupExportButtons();
  setupPopoutButton();
  displayTweets();
});

// Set up listener for changes in storage (live updates)
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.tweets) {
    displayTweets();
  }
});
