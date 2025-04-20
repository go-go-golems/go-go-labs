console.log("Background script initialized");
let collectedData = [];

async function loadInitialData() {
  try {
    const result = await browser.storage.local.get("tweetStore");
    collectedData = result.tweetStore || [];
    console.log(`Loaded ${collectedData.length} tweets`);
  } catch (err) {
    console.error("Storage load error:", err);
  }
}

async function saveData(tweet) {
  try {
    const exists = collectedData.some((t) => t.id === tweet.id);
    if (!exists) {
      collectedData.push(tweet);
      await browser.storage.local.set({ tweetStore: collectedData });
      console.log(`Saved tweet ${tweet.id} (Total: ${collectedData.length})`);
    }
  } catch (err) {
    console.error("Save error:", err);
  }
}

async function handleDownload() {
  if (!collectedData.length) return;

  try {
    const blob = new Blob([JSON.stringify(collectedData, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const filename = `tweets-${Date.now()}.json`;

    await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true,
    });
    console.log(`Downloaded ${collectedData.length} tweets`);
  } catch (err) {
    console.error("Download failed:", err);
  }
}

browser.runtime.onMessage.addListener((msg) => {
  switch (msg.type) {
    case "SAVE_DATA":
      saveData(msg.payload);
      break;
    case "DOWNLOAD_DATA":
      handleDownload();
      break;
    case "GET_STATS":
      return Promise.resolve({ count: collectedData.length });
  }
});

loadInitialData();
