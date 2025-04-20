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
  // Create a filename with a timestamp. We replace colons (:) from toISOString()
  // because they are illegal characters in filenames on some operating systems.
  const filename = `tweets_${new Date().toISOString().replace(/:/g, "-")}.json`;

  try {
    await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true, // Prompt user for save location
    });
    console.log(`Download initiated for ${filename}`);
  } catch (err) {
    console.error("Download failed:", err);
  }
  // Note: Firefox manages revoking the object URL automatically for downloads
}

// Function to handle custom download requests with specific data, mime type, and filename
async function handleCustomDownload(data, mimeType, filename) {
  console.log(`Preparing custom download: ${filename} (${mimeType})`);
  const blob = new Blob([data], { type: mimeType });
  const url = URL.createObjectURL(blob);

  try {
    await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true, // Prompt user for save location
    });
    console.log(`Download initiated for ${filename}`);
  } catch (err) {
    console.error("Download failed:", err);
  }
  // Note: Firefox manages revoking the object URL automatically for downloads, but
  // it's good practice to revoke it explicitly if needed in other contexts.
  // URL.revokeObjectURL(url); // Usually not needed for browser.downloads
}

browser.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  console.log("Received message in background script:", msg);
  if (msg.type === "SAVE_TWEET") {
    // Using .catch for basic error handling on the promise returned by save
    save(msg.tweet).catch((err) => console.error("Error saving tweet:", err));
  } else if (msg.type === "DOWNLOAD_TWEETS") {
    // Kept for potential backward compatibility or specific JSON download needs
    // Using .catch for basic error handling
    handleDownloadRequest().catch((err) =>
      console.error("Error handling JSON download request:", err)
    );
  } else if (msg.type === "DOWNLOAD_DATA") {
    // Handle the new message type for custom downloads
    // Using .catch for basic error handling
    handleCustomDownload(msg.data, msg.mimeType, msg.filename).catch((err) =>
      console.error("Error handling custom download:", err)
    );
  } else {
    console.log("Unknown message type received:", msg.type);
  }

  // If you needed to send an asynchronous response, you would return true.
  // return true; // Uncomment if sendResponse will be called asynchronously
});

console.log("Background script loaded and message listener added.");
