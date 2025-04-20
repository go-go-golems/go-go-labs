async function save(tweet) {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (tweets.some((t) => t.id === tweet.id)) return;
  tweets.push(tweet);
  await browser.storage.local.set({ tweets });
}

async function handleDownloadRequest() {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (tweets.length === 0) return;

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
  } catch (err) {
    console.error("Download failed:", err);
  }
}

browser.runtime.onMessage.addListener((msg) => {
  if (msg.type === "SAVE_TWEET") {
    save(msg.tweet);
  } else if (msg.type === "DOWNLOAD_TWEETS") {
    handleDownloadRequest();
  }
});
