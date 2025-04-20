console.log("Background script loading...");

async function saveTweet(tweet) {
  console.log("Save function called with tweet:", tweet);
  try {
    const { tweets = [] } = await browser.storage.local.get("tweets");
    if (tweets.some((t) => t.id === tweet.id)) {
      console.log(`Tweet ID ${tweet.id} already exists in storage. Skipping.`);
      return;
    }
    tweets.push(tweet);
    await browser.storage.local.set({ tweets });
    console.log(`Tweet ID ${tweet.id} saved. Total tweets: ${tweets.length}`);
  } catch (error) {
    console.error("Error saving tweet:", error);
  }
}

async function handleDownloadRequest() {
  console.log("Download request received.");
  try {
    const { tweets = [] } = await browser.storage.local.get("tweets");
    if (tweets.length === 0) {
      console.log("No tweets to download.");
      // Optionally, send a message back to the popup to inform the user?
      return;
    }

    console.log(`Preparing to download ${tweets.length} tweets.`);
    const blob = new Blob([JSON.stringify(tweets, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    // Create a filename with a timestamp. Replace colons and other potential invalid chars.
    const filename = `tweets_${new Date()
      .toISOString()
      .replace(/[:.]/g, "-")}.json`;

    try {
      await browser.downloads.download({
        url: url,
        filename: filename,
        saveAs: true, // Prompt user for save location
      });
      console.log(`Download initiated for ${filename}`);
    } catch (err) {
      console.error("Download failed:", err);
      // Revoke URL if download fails
      URL.revokeObjectURL(url);
    }
    // Note: Firefox manages revoking the object URL automatically for successful downloads,
    // but we revoke manually on error just in case.
  } catch (error) {
    console.error("Error handling download request:", error);
  }
}

browser.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  console.log("Received message in background script:", msg);
  if (msg.type === "SAVE_TWEET") {
    if (msg.tweet) {
      saveTweet(msg.tweet);
    } else {
      console.error("Received SAVE_TWEET message without tweet data");
    }
  } else if (msg.type === "DOWNLOAD_TWEETS") {
    handleDownloadRequest();
  } else {
    console.log("Received unknown message type:", msg.type);
  }

  // Note: Returning true would indicate an async response, but we don't need it here.
  // return true;
});

console.log("Background script loaded and message listener added.");
