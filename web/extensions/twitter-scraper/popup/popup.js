const list = document.getElementById("tweet-list");
const statsDiv = document.getElementById("stats-details");

async function updateUI() {
  console.log("[Popup Script] Updating UI...");
  try {
    const tweets = await browser.runtime.sendMessage({ type: "GET_TWEETS" });
    console.log("[Popup Script] Received tweets for UI:", tweets.length);

    // Update tweet list
    list.innerHTML = "";
    tweets
      .slice()
      .reverse()
      .forEach((t) => {
        const li = document.createElement("li");
        li.textContent = `${t.author}: ${t.text.slice(0, 80)}`;
        list.appendChild(li);
      });

    // Prepare export stats
    const jsonStr = JSON.stringify(tweets, null, 2);
    const totalTweets = tweets.length;
    const sizeKB = (jsonStr.length / 1024).toFixed(2);
    const sample =
      totalTweets > 0
        ? `${tweets[0].author}: ${tweets[0].text.slice(0, 80)}...`
        : "No tweets available";

    statsDiv.innerHTML =
      `<p>Total tweets: ${totalTweets}</p>` +
      `<p>Estimated export size: ${jsonStr.length} characters (~${sizeKB} KB)</p>` +
      `<p>Sample tweet: ${sample}</p>`;
    console.log("[Popup Script] UI Updated successfully.");
  } catch (err) {
    console.error("[Popup Script] Failed to update UI with tweets:", err);
    statsDiv.innerHTML = `<p>Error loading stats</p>`;
  }
}

// Initial UI update
updateUI();

// Listen for updates from the background script
browser.runtime.onMessage.addListener((message, sender) => {
  // Make sure the message is from the background script (no sender.tab)
  if (!sender.tab && message.type === "TWEET_SAVED") {
    console.log(
      "[Popup Script] Received TWEET_SAVED notification, refreshing UI..."
    );
    updateUI(); // Re-run the UI update function
  }
});

// Export functionality - NOW JUST SENDS A MESSAGE
document.getElementById("export").addEventListener("click", () => {
  console.log(
    "[Popup Script] Export button clicked, sending EXPORT_TWEETS message."
  );
  browser.runtime
    .sendMessage({ type: "EXPORT_TWEETS" })
    .then(() => {
      console.log("[Popup Script] EXPORT_TWEETS message sent successfully.");
    })
    .catch((err) => {
      console.error(
        "[Popup Script] Failed to send EXPORT_TWEETS message:",
        err
      );
      alert(`Failed to initiate export: ${err.message}`); // Inform user if message fails
    });
});
