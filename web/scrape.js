// Create and add control overlay
overlay = document.createElement('div');
overlay.style.cssText = 'position:fixed;top:10px;right:10px;background:rgba(0,0,0,0.8);color:white;padding:10px;z-index:9999;border-radius:5px;font-family:Arial;';
overlay.innerHTML = `
  <div style="margin-bottom:5px;font-weight:bold;text-align:center;">Tweet Scraper</div>
  <div style="display:flex;justify-content:space-between;margin-bottom:5px;">
    <button id="play-btn" style="padding:5px 10px;cursor:pointer;background:#4CAF50;border:none;border-radius:3px;color:white;">Play</button>
    <button id="pause-btn" style="padding:5px 10px;cursor:pointer;background:#FFC107;border:none;border-radius:3px;color:black;">Pause</button>
    <button id="stop-btn" style="padding:5px 10px;cursor:pointer;background:#F44336;border:none;border-radius:3px;color:white;">Stop</button>
  </div>
  <div id="status" style="font-size:12px;margin-top:5px;">Ready: 0 tweets</div>
`;
document.body.appendChild(overlay);

// Initialize variables
allTweets = {};
isRunning = false;
isPaused = false;
tweetCount = 0;
scrollInterval = null;

// Function to process visible tweets
function processTweets() {
  // Using the more specific data-testid="tweet" selector
  tweets = document.querySelectorAll('[data-testid="tweet"]');
  tweets.forEach(tweet => {
    try {
      // Generate a unique ID for the tweet (using a combination of content and author)
      tweetText = tweet.querySelector('[data-testid="tweetText"]')?.textContent.trim() || "";
      if (!tweetText) return;
      
      username = tweet.querySelector('[data-testid="User-Name"] div[dir="ltr"]:first-child')?.textContent.trim() || "";
      timestamp = tweet.querySelector('[data-testid="User-Name"] + div a time')?.textContent.trim() || "";
      
      // Create a simple hash for the tweet
      tweetId = `${username}_${tweetText.substring(0, 20)}_${timestamp}`;
      
      // Skip if we've already processed this tweet
      if (allTweets[tweetId]) return;
      
      // Extract remaining data
      handle = tweet.querySelector('[data-testid="User-Name"] + div a div[dir="ltr"]')?.textContent.trim() || "";
      verified = tweet.querySelector('[data-testid="icon-verified"]') ? "true" : "false";
      
      // Extract engagement metrics
      metrics = Array.from(tweet.querySelectorAll('[role="group"] [data-testid="app-text-transition-container"] span span')).map(el => el.textContent);
      replies = metrics[0] || "0";
      reposts = metrics[1] || "0";
      likes = metrics[2] || "0";
      views = metrics[3] || "0";
      
      // Extract bookmarks if available
      bookmarks = tweet.querySelector('[data-testid="bookmark"] [data-testid="app-text-transition-container"] span span')?.textContent || "0";
      
      // Store the tweet data
      allTweets[tweetId] = {
        username,
        handle,
        verified,
        timestamp,
        text: tweetText.replace(/\n/g, ' ').replace(/"/g, '""'),
        replies,
        reposts,
        likes,
        views,
        bookmarks
      };
      
      tweetCount++;
      document.getElementById('status').textContent = `Collected: ${tweetCount} tweets`;
    } catch (e) {
      console.error("Error processing tweet:", e);
    }
  });
}

// Function to scroll down
function scrollDown() {
  if (isRunning && !isPaused) {
    window.scrollBy(0, 500);
    processTweets();
  }
}

// Function to generate and download CSV
function generateCSV() {
  csvContent = "username,handle,verified,timestamp,text,replies,reposts,likes,views,bookmarks\n";
  
  Object.values(allTweets).forEach(tweet => {
    csvContent += `"${tweet.username}","${tweet.handle}","${tweet.verified}","${tweet.timestamp}","${tweet.text}","${tweet.replies}","${tweet.reposts}","${tweet.likes}","${tweet.views}","${tweet.bookmarks}"\n`;
  });
  
  // Create downloadable CSV file
  blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  url = URL.createObjectURL(blob);
  link = document.createElement('a');
  link.setAttribute('href', url);
  link.setAttribute('download', 'tweets_export.csv');
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  
  console.log(`Downloaded CSV with ${tweetCount} tweets`);
}

// Button event listeners
document.getElementById('play-btn').addEventListener('click', function() {
  if (!isRunning) {
    isRunning = true;
    isPaused = false;
    scrollInterval = setInterval(scrollDown, 2000);
    this.style.background = '#388E3C';
    document.getElementById('pause-btn').style.background = '#FFC107';
    document.getElementById('status').textContent = `Running: ${tweetCount} tweets`;
  } else if (isPaused) {
    isPaused = false;
    this.style.background = '#388E3C';
    document.getElementById('pause-btn').style.background = '#FFC107';
    document.getElementById('status').textContent = `Running: ${tweetCount} tweets`;
  }
});

document.getElementById('pause-btn').addEventListener('click', function() {
  if (isRunning && !isPaused) {
    isPaused = true;
    this.style.background = '#FF8F00';
    document.getElementById('play-btn').style.background = '#4CAF50';
    document.getElementById('status').textContent = `Paused: ${tweetCount} tweets`;
  }
});

document.getElementById('stop-btn').addEventListener('click', function() {
  isRunning = false;
  isPaused = false;
  clearInterval(scrollInterval);
  document.getElementById('play-btn').style.background = '#4CAF50';
  document.getElementById('pause-btn').style.background = '#FFC107';
  document.getElementById('status').textContent = `Stopped: ${tweetCount} tweets`;
  
  // Generate and download CSV
  generateCSV();
  
  // Ask if user wants to remove the overlay
  setTimeout(() => {
    if (confirm('Do you want to remove the scraper controls?')) {
      document.body.removeChild(overlay);
    }
  }, 1000);
});

// Initial processing of tweets already on screen
processTweets();
document.getElementById('status').textContent = `Ready: ${tweetCount} tweets`;
