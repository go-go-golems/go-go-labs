document.addEventListener("DOMContentLoaded", initPopup);

async function initPopup() {
  const downloadBtn = document.getElementById("downloadBtn");
  const statsDiv = document.getElementById("stats");
  const tweetList = document.getElementById("tweet-list");

  async function updateUI() {
    try {
      const { count } = await browser.runtime.sendMessage({
        type: "GET_STATS",
      });
      statsDiv.textContent = `${count} tweets captured`;

      const { tweetStore } = await browser.storage.local.get("tweetStore");
      tweetList.innerHTML = tweetStore.length
        ? tweetStore
            .slice(-10)
            .reverse()
            .map(
              (t) => `
          <div class="tweet">
            <div>${t.author}</div>
            <div>${t.text.slice(0, 80)}${t.text.length > 80 ? "..." : ""}</div>
          </div>
        `
            )
            .join("")
        : "<div>No tweets captured yet</div>";
    } catch (err) {
      console.error("Popup error:", err);
    }
  }

  downloadBtn.addEventListener("click", () => {
    browser.runtime.sendMessage({ type: "DOWNLOAD_DATA" });
  });

  browser.storage.onChanged.addListener(updateUI);
  await updateUI();
}
