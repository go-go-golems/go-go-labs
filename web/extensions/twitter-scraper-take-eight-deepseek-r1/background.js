async function saveTweet(tweet) {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  if (!tweets.some((t) => t.id === tweet.id)) {
    await browser.storage.local.set({ tweets: [...tweets, tweet] });
  }
}

async function handleDownload() {
  const { tweets = [] } = await browser.storage.local.get("tweets");
  const filename = `tweets_${Date.now()}.json`;
  const blob = new Blob([JSON.stringify(tweets, null, 2)], {
    type: "application/json",
  });
  const url = URL.createObjectURL(blob);

  await browser.downloads.download({
    url,
    filename,
    saveAs: true,
  });
}

browser.runtime.onMessage.addListener((msg) => {
  switch (msg.type) {
    case "SAVE_TWEET":
      return saveTweet(msg.tweet);
    case "DOWNLOAD_TWEETS":
      return handleDownload();
  }
});
