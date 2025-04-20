console.log("Background script loading...");

async function saveTweet(tweet) {
  console.log("Save function called with tweet:", tweet);
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (tweets.some((t) => t.id === tweet.id)) {
    console.log(`Tweet ID ${tweet.id} already in storage. Skipping.`);
    return;
  }
  tweets.push(tweet);
  await browser.storage.local.set({ tweets });
  console.log(`Tweet ${tweet.id} saved. Total tweets now: ${tweets.length}`);
}

async function handleDownloadRequest() {
  console.log("Download request received.");
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (!tweets.length) {
    console.log("No tweets to download.");
    return;
  }
  const blob = new Blob([JSON.stringify(tweets, null, 2)], {
    type: "application/json",
  });
  const url = URL.createObjectURL(blob);
  const filename = `tweets_${new Date().toISOString().replace(/:/g, "-")}.json`;
  try {
    await browser.downloads.download({
      url,
      filename,
      saveAs: true,
    });
    console.log(`Download initiated: ${filename}`);
  } catch (err) {
    console.error("Download failed:", err);
  }
}

browser.runtime.onMessage.addListener((msg, sender) => {
  console.log("Received message in background:", msg);
  if (msg.type === "SAVE_TWEET") {
    saveTweet(msg.tweet);
  } else if (msg.type === "DOWNLOAD_TWEETS") {
    handleDownloadRequest();
  } else {
    console.log("Unknown message type:", msg.type);
  }
});

console.log("Background script loaded and listener active.");
