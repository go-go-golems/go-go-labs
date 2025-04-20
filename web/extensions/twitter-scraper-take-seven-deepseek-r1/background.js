console.log("Background script loading...");

async function save(tweet) {
  console.log("Save function called with tweet:", tweet);
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (tweets.some((t) => t.id === tweet.id)) {
    console.log(`Tweet ID ${tweet.id} already exists in storage. Skipping.`);
    return;
  }
  tweets.push(tweet);
  await browser.storage.local.set({ tweets });
  console.log(`Tweet ID ${tweet.id} saved. Total tweets: ${tweets.length}`);
}

async function handleDownloadRequest() {
  console.log("Download request received.");
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (tweets.length === 0) {
    console.log("No tweets to download.");
    return;
  }

  console.log(`Preparing to download ${tweets.length} tweets.`);
  const blob = new Blob([JSON.stringify(tweets, null, 2)], {
    type: "application/json",
  });
  const url = URL.createObjectURL(blob);
  const filename = `tweets_${new Date().toISOString().replace(/:/g, "-")}.json`;

  try {
    await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true,
    });
    console.log(`Download initiated for ${filename}`);
  } catch (err) {
    console.error("Download failed:", err);
  }
}

browser.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  console.log("Received message in background script:", msg);
  if (msg.type === "SAVE_TWEET") {
    save(msg.tweet);
  } else if (msg.type === "DOWNLOAD_TWEETS") {
    handleDownloadRequest();
  } else {
    console.log("Unknown message type:", msg.type);
  }
});

console.log("Background script loaded and message listener added."); 