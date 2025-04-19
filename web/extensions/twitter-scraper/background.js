const TWEET_LIMIT = 5000;

async function saveTweet(tweet) {
  console.log("[Background Script] Received SAVE_TWEET for:", tweet.id);
  try {
    const { tweets = [] } = await browser.storage.local.get("tweets");
    if (tweets.some((t) => t.id === tweet.id)) {
      console.log(
        "[Background Script] Tweet already exists, skipping:",
        tweet.id
      );
      return false; // Indicate no save occurred
    }
    tweets.push(tweet);
    if (tweets.length > TWEET_LIMIT) tweets.shift();
    await browser.storage.local.set({ tweets });
    console.log("[Background Script] Tweet saved successfully:", tweet.id);
    return true; // Indicate save occurred
  } catch (error) {
    console.error("[Background Script] Error saving tweet:", tweet.id, error);
    return false;
  }
}

browser.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  console.log(
    "[Background Script] Message received:",
    msg.type,
    "from:",
    sender.tab ? "Tab " + sender.tab.id : "Popup/Other"
  );
  if (msg.type === "SAVE_TWEET") {
    saveTweet(msg.tweet).then((saved) => {
      if (saved) {
        // Notify popup if open
        browser.runtime.sendMessage({ type: "TWEET_SAVED" }).catch((err) => {
          // Ignore error if popup isn't open
          if (!err.message.includes("Could not establish connection")) {
            console.error("[Background Script] Error notifying popup:", err);
          }
        });
      }
      sendResponse({ ok: true, saved });
    });
    return true; // Indicates async response
  }
  if (msg.type === "GET_TWEETS") {
    console.log("[Background Script] Received GET_TWEETS request");
    browser.storage.local
      .get("tweets")
      .then((r) => {
        console.log(
          "[Background Script] Sending tweets to requester:",
          (r.tweets || []).length
        );
        sendResponse(r.tweets || []);
      })
      .catch((err) => {
        console.error("[Background Script] Error getting tweets:", err);
        sendResponse([]); // Send empty array on error
      });
    return true; // Indicates async response
  }
  if (msg.type === "EXPORT_TWEETS") {
    console.log(
      "[Background Script] Received EXPORT_TWEETS request from:",
      sender.tab ? sender.tab.id : "Popup"
    );

    // Use an IIFE async function to handle the export logic
    (async () => {
      let url = null; // Keep track of the URL to revoke it
      try {
        const { tweets = [] } = await browser.storage.local.get("tweets");
        console.log(`[Background Script] Exporting ${tweets.length} tweets.`);
        if (tweets.length === 0) {
          console.warn(
            "[Background Script] Export requested, but no tweets found."
          );
          // Optional: Notify popup? For now, just log.
          return; // Nothing to export
        }

        const blob = new Blob([JSON.stringify(tweets, null, 2)], {
          type: "application/json",
        });
        url = URL.createObjectURL(blob);
        console.log("[Background Script] Created Blob URL for export:", url);

        const downloadId = await browser.downloads.download({
          url: url,
          filename: `tweets-archive-${Date.now()}.json`, // Use timestamp in filename
          saveAs: true, // Prompt user for save location
        });
        console.log(
          "[Background Script] Download initiated with ID:",
          downloadId
        );

        // --- Keep the Blob alive until download finishes ---
        const holdBlob = (delta) => {
          // We only care about the download we just started
          if (delta.id !== downloadId) return;

          // Check if the download state has changed and is no longer in progress
          if (delta.state && delta.state.current !== "in_progress") {
            console.log(
              `[Background Script] Download ${downloadId} finished with state: ${delta.state.current}. Revoking Blob URL: ${url}`
            );
            URL.revokeObjectURL(url); // Clean up the Blob URL
            browser.downloads.onChanged.removeListener(holdBlob); // Remove this listener
          }
        };
        browser.downloads.onChanged.addListener(holdBlob);
      } catch (err) {
        console.error("[Background Script] Export failed:", err);
        // Revoke URL if it was created before the error occurred
        if (url) {
          console.log(
            "[Background Script] Revoking Blob URL due to export error."
          );
          URL.revokeObjectURL(url);
        }
        // Optional: Send error message back to popup?
      }
    })();

    // Note: We don't 'return true' here because we are not using sendResponse for this message type.
    // The export process runs independently after the message is received.
  }
});

console.log("[Background Script] Listener attached.");
