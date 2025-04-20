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
  console.log("Handling download request...");
  try {
    console.log("Retrieving tweets from storage...");
    const { tweets = [] } = await browser.storage.local.get("tweets");
    console.log(`Retrieved ${tweets.length} tweets.`);

    if (tweets.length === 0) {
      console.warn("No tweets found in storage to download.");
      // Consider sending a message back to popup or showing a notification
      return;
    }

    console.log("Creating Blob...");
    const blob = new Blob([JSON.stringify(tweets, null, 2)], {
      type: "application/json",
    });
    console.log("Blob created.");

    console.log("Creating Object URL...");
    const url = URL.createObjectURL(blob);
    console.log("Object URL created:", url);

    const filename = "tweets.json";
    console.log(`Initiating download for ${filename}`);

    const downloadId = await browser.downloads.download({
      url,
      filename: filename,
      saveAs: true,
    });

    console.log(`Download initiated successfully. Download ID: ${downloadId}`);

    // Revoking the URL here is important, but we need to be careful.
    // Revoking too soon can break the download.
    // A common pattern is to listen for the download to complete or fail.
    // For simplicity here, we revoke after a delay, but listening to download
    // events (`browser.downloads.onChanged`) is more robust.
    setTimeout(() => {
      console.log("Revoking Object URL (after delay):", url);
      URL.revokeObjectURL(url);
      console.log("Object URL revoked.");
    }, 10000); // Revoke after 10 seconds
  } catch (error) {
    console.error("Error during background download process:", error);
    if (typeof url !== "undefined" && url) {
      console.log("Revoking Object URL due to error:", url);
      URL.revokeObjectURL(url);
    }
  }
}

browser.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  console.log("Received message in background script:", msg);
  if (msg.type === "SAVE_TWEET") {
    console.log("Message type is SAVE_TWEET, calling save function...");
    save(msg.tweet);
  } else if (msg.type === "DOWNLOAD_TWEETS") {
    console.log(
      "Message type is DOWNLOAD_TWEETS, calling handleDownloadRequest..."
    );
    handleDownloadRequest();
    // Optional: Send a response back to the popup if needed
    // sendResponse({ status: "Download initiated" });
  } else {
    console.log(`Received unknown message type: ${msg.type}`);
  }
  // Return true to indicate you wish to send a response asynchronously
  // (important if using sendResponse in async functions like handleDownloadRequest)
  // return true;
});

console.log("Background script loaded and message listener added.");
