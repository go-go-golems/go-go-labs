// Store filtered tweet list
let filteredTweets = [];

// Simple sentiment analysis
function analyzeSentiment(text) {
  const positiveWords = [
    "good",
    "great",
    "awesome",
    "excellent",
    "love",
    "happy",
    "best",
    "beautiful",
    "thanks",
    "thank",
    "amazing",
    "perfect",
    "wonderful",
    "nice",
    "cool",
    "win",
    "winning",
    "congrats",
    "congratulations",
    "excited",
    "exciting",
  ];
  const negativeWords = [
    "bad",
    "terrible",
    "hate",
    "awful",
    "worst",
    "sad",
    "disappointing",
    "disappointed",
    "horrible",
    "poor",
    "wrong",
    "sucks",
    "sorry",
    "problem",
    "fail",
    "failing",
    "failed",
    "annoying",
    "annoyed",
    "angry",
    "mad",
  ];

  text = text.toLowerCase();
  let positiveScore = 0;
  let negativeScore = 0;

  positiveWords.forEach((word) => {
    const regex = new RegExp(`\\b${word}\\b`, "gi");
    const matches = text.match(regex);
    if (matches) positiveScore += matches.length;
  });

  negativeWords.forEach((word) => {
    const regex = new RegExp(`\\b${word}\\b`, "gi");
    const matches = text.match(regex);
    if (matches) negativeScore += matches.length;
  });

  if (positiveScore > negativeScore) return "positive";
  if (negativeScore > positiveScore) return "negative";
  return "neutral";
}

// Extract topics from tweet text
function extractTopics(text) {
  const hashtags = text.match(/#[a-zA-Z0-9_]+/g) || [];
  const topics = hashtags.map((tag) => tag.substring(1));

  // If no hashtags, try to extract common topics
  if (topics.length === 0) {
    const commonTopics = [
      "technology",
      "politics",
      "sports",
      "entertainment",
      "business",
      "health",
      "science",
      "gaming",
      "music",
      "food",
      "travel",
      "fashion",
    ];

    commonTopics.forEach((topic) => {
      if (text.toLowerCase().includes(topic)) {
        topics.push(topic);
      }
    });
  }

  return topics.length > 0 ? topics : ["general"];
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
  const headers = ["id", "author", "text", "sentiment", "topics", "timestamp"];
  let csv = headers.join(",") + "\n";

  tweets.forEach((tweet) => {
    const row = [
      tweet.id,
      `"${tweet.author.replace(/"/g, '""')}"`,
      `"${tweet.text.replace(/"/g, '""')}"`,
      tweet.sentiment || "neutral",
      `"${(tweet.topics || []).join(", ")}"`,
      tweet.timestamp || "",
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
    if (tweet.sentiment) text += `Sentiment: ${tweet.sentiment}\n`;
    if (tweet.topics && tweet.topics.length)
      text += `Topics: ${tweet.topics.join(", ")}\n`;
    text += `Time: ${formatTime(tweet.timestamp)}\n`;
    text += "--------------------\n\n";
  });

  return text;
}

// Theme toggle functionality
function setupThemeToggle() {
  const themeToggle = document.getElementById("theme-toggle");
  const storedTheme = localStorage.getItem("theme");

  // Apply stored theme or default to light
  if (storedTheme === "dark") {
    document.body.setAttribute("data-theme", "dark");
    themeToggle.textContent = "☀️";
  }

  themeToggle.addEventListener("click", () => {
    if (document.body.getAttribute("data-theme") === "dark") {
      // Switch to light
      document.body.removeAttribute("data-theme");
      localStorage.setItem("theme", "light");
      themeToggle.textContent = "🌙";
    } else {
      // Switch to dark
      document.body.setAttribute("data-theme", "dark");
      localStorage.setItem("theme", "dark");
      themeToggle.textContent = "☀️";
    }
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

// Set up filter toggle
function setupFilters() {
  const filterToggle = document.getElementById("filter-toggle");
  const filtersPanel = document.getElementById("filters-panel");

  filterToggle.addEventListener("click", () => {
    filtersPanel.classList.toggle("show");
    filterToggle.textContent = filtersPanel.classList.contains("show")
      ? "✖"
      : "🔍";
  });
}

// Initialize author and topic filter dropdowns
function updateFilterDropdowns(tweets) {
  const authorFilter = document.getElementById("author-filter");
  const topicFilter = document.getElementById("topic-filter");

  // Clear existing options (keeping the first "All" option)
  while (authorFilter.options.length > 1) {
    authorFilter.remove(1);
  }

  while (topicFilter.options.length > 1) {
    topicFilter.remove(1);
  }

  // Get unique authors
  const authors = [...new Set(tweets.map((tweet) => tweet.author))].sort();

  // Get unique topics
  const allTopics = tweets.flatMap((tweet) => tweet.topics || []);
  const topics = [...new Set(allTopics)].sort();

  // Add options
  authors.forEach((author) => {
    const option = document.createElement("option");
    option.value = author;
    option.textContent = author;
    authorFilter.appendChild(option);
  });

  topics.forEach((topic) => {
    const option = document.createElement("option");
    option.value = topic;
    option.textContent = topic;
    topicFilter.appendChild(option);
  });
}

// Apply filters to tweets
function applyFilters(tweets) {
  const searchInput = document.getElementById("search-input");
  const authorFilter = document.getElementById("author-filter");
  const sentimentFilter = document.getElementById("sentiment-filter");
  const topicFilter = document.getElementById("topic-filter");

  const searchText = searchInput.value.toLowerCase();
  const author = authorFilter.value;
  const sentiment = sentimentFilter.value;
  const topic = topicFilter.value;

  return tweets.filter((tweet) => {
    // Apply search text filter
    if (
      searchText &&
      !tweet.text.toLowerCase().includes(searchText) &&
      !tweet.author.toLowerCase().includes(searchText)
    ) {
      return false;
    }

    // Apply author filter
    if (author && tweet.author !== author) {
      return false;
    }

    // Apply sentiment filter
    if (sentiment && tweet.sentiment !== sentiment) {
      return false;
    }

    // Apply topic filter
    if (topic && (!tweet.topics || !tweet.topics.includes(topic))) {
      return false;
    }

    return true;
  });
}

// Set up search and filter functionality
function setupSearch() {
  const searchInput = document.getElementById("search-input");
  const authorFilter = document.getElementById("author-filter");
  const sentimentFilter = document.getElementById("sentiment-filter");
  const topicFilter = document.getElementById("topic-filter");

  const filterElements = [
    searchInput,
    authorFilter,
    sentimentFilter,
    topicFilter,
  ];

  filterElements.forEach((element) => {
    element.addEventListener("input", () => {
      displayTweets();
    });
  });
}

// Update sentiment charts
function updateSentimentCharts(tweets) {
  // Count sentiments
  const sentiments = {
    positive: 0,
    neutral: 0,
    negative: 0,
  };

  tweets.forEach((tweet) => {
    sentiments[tweet.sentiment || "neutral"]++;
  });

  // Calculate percentages and heights (minimum 10% for visibility)
  const total = Math.max(tweets.length, 1); // Avoid division by zero
  const positiveHeight = Math.max(10, (sentiments.positive / total) * 100);
  const neutralHeight = Math.max(10, (sentiments.neutral / total) * 100);
  const negativeHeight = Math.max(10, (sentiments.negative / total) * 100);

  // Update bars
  const positiveBar = document.getElementById("positive-bar");
  const neutralBar = document.getElementById("neutral-bar");
  const negativeBar = document.getElementById("negative-bar");

  positiveBar.style.height = `${positiveHeight}%`;
  neutralBar.style.height = `${neutralHeight}%`;
  negativeBar.style.height = `${negativeHeight}%`;

  // Update values
  positiveBar.querySelector(".bar-value").textContent = sentiments.positive;
  neutralBar.querySelector(".bar-value").textContent = sentiments.neutral;
  negativeBar.querySelector(".bar-value").textContent = sentiments.negative;
}

// Update topics chart
function updateTopicsChart(tweets) {
  const topicsChart = document.getElementById("topics-chart");

  // If no tweets, show empty state
  if (tweets.length === 0) {
    topicsChart.innerHTML =
      '<div class="empty-state">Collect tweets to see topic insights</div>';
    return;
  }

  // Count topics
  const topicCounts = {};
  tweets.forEach((tweet) => {
    (tweet.topics || ["general"]).forEach((topic) => {
      topicCounts[topic] = (topicCounts[topic] || 0) + 1;
    });
  });

  // Sort topics by count (descending)
  const sortedTopics = Object.entries(topicCounts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5); // Show top 5 topics

  // Create topic bars
  topicsChart.innerHTML = "";
  sortedTopics.forEach(([topic, count]) => {
    const percentage = (count / tweets.length) * 100;

    const topicRow = document.createElement("div");
    topicRow.className = "filter-row";

    const topicName = document.createElement("div");
    topicName.className = "filter-label";
    topicName.textContent = topic;

    const topicCountLabel = document.createElement("div");
    topicCountLabel.className = "filter-label";
    topicCountLabel.textContent = count;

    const progressContainer = document.createElement("div");
    progressContainer.style.display = "flex";
    progressContainer.style.height = "10px";
    progressContainer.style.width = "100%";
    progressContainer.style.borderRadius = "5px";
    progressContainer.style.overflow = "hidden";
    progressContainer.style.backgroundColor = "var(--border-color)";
    progressContainer.style.marginTop = "5px";

    const progress = document.createElement("div");
    progress.style.width = `${percentage}%`;
    progress.style.backgroundColor = "var(--primary-color)";

    progressContainer.appendChild(progress);
    topicRow.appendChild(topicName);
    topicRow.appendChild(topicCountLabel);
    topicsChart.appendChild(topicRow);
    topicsChart.appendChild(progressContainer);
  });
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
      createConfetti();
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

// Create fun confetti effect when new tweets are added
function createConfetti() {
  const confettiContainer = document.createElement("div");
  confettiContainer.style.position = "absolute";
  confettiContainer.style.top = "0";
  confettiContainer.style.left = "0";
  confettiContainer.style.width = "100%";
  confettiContainer.style.height = "100%";
  confettiContainer.style.pointerEvents = "none";
  confettiContainer.style.zIndex = "1000";
  document.body.appendChild(confettiContainer);

  const colors = ["#1da1f2", "#ffad1f", "#e0245e", "#17bf63"];

  // Create 30 confetti particles
  for (let i = 0; i < 30; i++) {
    const confetti = document.createElement("div");
    confetti.style.position = "absolute";
    confetti.style.width = Math.random() * 10 + 5 + "px";
    confetti.style.height = Math.random() * 6 + 3 + "px";
    confetti.style.backgroundColor =
      colors[Math.floor(Math.random() * colors.length)];
    confetti.style.borderRadius = Math.random() > 0.5 ? "50%" : "0";
    confetti.style.top = "-10px";
    confetti.style.left = Math.random() * 100 + "%";
    confetti.style.transform = `rotate(${Math.random() * 360}deg)`;
    confetti.style.opacity = "1";
    confetti.style.transition = `top ${Math.random() * 2 + 1}s, left ${
      Math.random() * 2 + 1
    }s, opacity 0.5s`;

    confettiContainer.appendChild(confetti);

    // Animate confetti falling
    setTimeout(() => {
      confetti.style.top = 100 + Math.random() * 20 + "%";
      confetti.style.left =
        parseFloat(confetti.style.left) + (Math.random() * 40 - 20) + "%";
      confetti.style.transform = `rotate(${Math.random() * 360}deg)`;

      // Remove after animation
      setTimeout(() => {
        confetti.style.opacity = "0";
        setTimeout(() => confetti.remove(), 500);
      }, 1000);
    }, 10);
  }

  // Remove container after all confetti are gone
  setTimeout(() => confettiContainer.remove(), 3000);
}

// Function to display tweets in the popup UI
async function displayTweets() {
  let { tweets = [] } = await browser.storage.local.get("tweets");
  const tweetList = document.getElementById("tweet-list");
  const tweetCount = document.getElementById("tweet-count");
  const authorCount = document.getElementById("author-count");
  const topicsCount = document.getElementById("topics-count");

  // Process tweets by adding sentiment and topics if not already present
  tweets = tweets.map((tweet) => {
    if (!tweet.sentiment) {
      tweet.sentiment = analyzeSentiment(tweet.text);
    }
    if (!tweet.topics) {
      tweet.topics = extractTopics(tweet.text);
    }
    return tweet;
  });

  // Save processed tweets back if we added new properties
  browser.storage.local.set({ tweets });

  // Update filter dropdowns
  updateFilterDropdowns(tweets);

  // Apply filters
  filteredTweets = applyFilters(tweets);

  // Store previous values for animation
  const prevTweetCount = parseInt(tweetCount.textContent);
  const prevAuthorCount = parseInt(authorCount.textContent);

  // Get all topics
  const allTopics = tweets.flatMap((tweet) => tweet.topics || []);
  const uniqueTopics = new Set(allTopics);

  // Update counts with animation
  tweetCount.textContent = tweets.length;
  animateCountChange(tweetCount, prevTweetCount, tweets.length);

  // Calculate unique authors
  const uniqueAuthors = new Set(tweets.map((tweet) => tweet.author)).size;
  authorCount.textContent = uniqueAuthors;
  animateCountChange(authorCount, prevAuthorCount, uniqueAuthors);

  // Update topics count
  topicsCount.textContent = uniqueTopics.size;

  // Update charts
  updateSentimentCharts(tweets);
  updateTopicsChart(tweets);

  // If new tweets were added, show confetti
  if (tweets.length > prevTweetCount && prevTweetCount > 0) {
    createConfetti();
  }

  // Clear the list
  tweetList.innerHTML = "";

  if (filteredTweets.length === 0) {
    // Show empty state message
    const emptyState = document.createElement("div");
    emptyState.className = "empty-state";

    if (tweets.length === 0) {
      emptyState.textContent =
        "No tweets collected yet. Scroll through Twitter to collect tweets.";
    } else {
      emptyState.textContent = "No tweets match your current filters.";
    }

    tweetList.appendChild(emptyState);
    return;
  }

  // Get timestamp from 30 seconds ago to mark recent tweets
  const recentTimestamp = Date.now() - 30000;
  // Keep track of tweet IDs we've already seen in this popup view
  const viewedTweetIds = new Set();

  // Add each tweet to the list (most recent first)
  filteredTweets
    .slice()
    .reverse()
    .forEach((tweet, index) => {
      const tweetElement = document.createElement("div");
      tweetElement.className = "tweet";

      // Create tweet header with author and time
      const tweetHeader = document.createElement("div");
      tweetHeader.className = "tweet-header";

      const authorElement = document.createElement("div");
      authorElement.className = "author";
      authorElement.textContent = tweet.author;

      // Add sentiment indicator to author
      const sentimentDot = document.createElement("span");
      sentimentDot.className = `sentiment ${tweet.sentiment || "neutral"}`;
      authorElement.appendChild(sentimentDot);

      // Add time if available
      const timeElement = document.createElement("div");
      timeElement.className = "tweet-time";
      timeElement.textContent = formatTime(tweet.timestamp);

      tweetHeader.appendChild(authorElement);
      tweetHeader.appendChild(timeElement);

      // Create text element
      const textElement = document.createElement("p");
      textElement.className = "text";
      textElement.textContent = tweet.text;

      // Add topic badges if available
      if (tweet.topics && tweet.topics.length > 0) {
        const topicsContainer = document.createElement("div");
        topicsContainer.style.marginTop = "8px";

        tweet.topics.forEach((topic) => {
          const topicBadge = document.createElement("span");
          topicBadge.className = "topic-badge";
          topicBadge.textContent = topic;
          topicsContainer.appendChild(topicBadge);
        });

        tweetElement.appendChild(tweetHeader);
        tweetElement.appendChild(textElement);
        tweetElement.appendChild(topicsContainer);
      } else {
        tweetElement.appendChild(tweetHeader);
        tweetElement.appendChild(textElement);
      }

      // Check if this is a new tweet (recently seen)
      const isNewTweet =
        !viewedTweetIds.has(tweet.id) &&
        tweet.timestamp &&
        tweet.timestamp > recentTimestamp;

      // Add badge for recent tweets
      addNewBadge(tweetElement, isNewTweet || (!tweet.timestamp && index < 3));

      // Mark as viewed
      viewedTweetIds.add(tweet.id);

      tweetList.appendChild(tweetElement);
    });
}

// Initial setup when popup opens
document.addEventListener("DOMContentLoaded", () => {
  setupThemeToggle();
  setupTabs();
  setupFilters();
  setupSearch();
  setupExportButtons();
  displayTweets();
});

// Set up listener for changes in storage (live updates)
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.tweets) {
    displayTweets();
  }
});
