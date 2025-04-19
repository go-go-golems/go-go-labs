const list = document.getElementById("tweet-list");
const statsDiv = document.getElementById("stats-details");

async function updateUI() {
  console.log("[Popup Script] Updating UI...");
  try {
    const tweets = await browser.runtime.sendMessage({ type: "GET_TWEETS" });
    console.log("[Popup Script] Received tweets for UI:", tweets.length);

    // Update tweet list
    list.innerHTML = "";
    if (tweets.length === 0) {
      list.innerHTML = "<li>No tweets saved yet. Scroll on Twitter/X!</li>";
    } else {
      tweets
        .slice()
        .reverse()
        .forEach((t) => {
          const li = document.createElement("li");

          // Create author span with proper styling
          const authorSpan = document.createElement("span");
          authorSpan.className = "tweet-author";
          authorSpan.textContent = t.author || "Unknown";

          // Create text span with proper styling
          const textSpan = document.createElement("span");
          textSpan.className = "tweet-text";
          textSpan.textContent = `: ${t.text.slice(0, 120)}${
            t.text.length > 120 ? "..." : ""
          }`;

          // Append both spans to the list item
          li.appendChild(authorSpan);
          li.appendChild(textSpan);

          list.appendChild(li);
        });
    }

    // Prepare export stats
    const jsonStr = JSON.stringify(tweets, null, 2);
    const totalTweets = tweets.length;
    const sizeKB = (jsonStr.length / 1024).toFixed(2);
    const sizeMB = (jsonStr.length / (1024 * 1024)).toFixed(2);

    const sizeDisplay = sizeKB > 1000 ? `${sizeMB} MB` : `${sizeKB} KB`;

    const sample =
      totalTweets > 0
        ? `${tweets[0].author || "Unknown"}: ${tweets[0].text.slice(0, 80)}...`
        : "No tweets available";

    statsDiv.innerHTML =
      `<p><strong>Total tweets:</strong> ${totalTweets}</p>` +
      `<p><strong>Export size:</strong> ~${sizeDisplay}</p>` +
      `<p><strong>Oldest tweet:</strong> ${sample}</p>`;

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

// Reset functionality
document.getElementById("reset").addEventListener("click", () => {
  if (confirm("Are you sure you want to delete all saved tweets?")) {
    console.log(
      "[Popup Script] Reset button clicked, sending RESET_TWEETS message."
    );
    browser.runtime
      .sendMessage({ type: "RESET_TWEETS" })
      .then((response) => {
        console.log(
          "[Popup Script] RESET_TWEETS message sent successfully. Response:",
          response
        );
        if (response && response.success) {
          alert("All tweets have been reset successfully!");
          updateUI(); // Refresh UI to show empty state
        }
      })
      .catch((err) => {
        console.error(
          "[Popup Script] Failed to send RESET_TWEETS message:",
          err
        );
        alert(`Failed to reset tweets: ${err.message}`);
      });
  }
});
