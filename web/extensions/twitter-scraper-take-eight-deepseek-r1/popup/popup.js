async function updateUI() {
  const { tweets = [] } = await browser.storage.local.get("tweets");

  document.getElementById("tweet-count").textContent = tweets.length;
  document.getElementById("author-count").textContent = new Set(
    tweets.map((t) => t.author)
  ).size;

  const list = document.getElementById("tweet-list");
  list.innerHTML = tweets.length ? "" : "<div>No tweets collected yet</div>";

  tweets
    .slice()
    .reverse()
    .forEach((t) => {
      const div = document.createElement("div");
      div.className = "tweet";
      div.innerHTML = `
      <div class="author">${t.author}</div>
      <div class="text">${t.text.slice(0, 80)}${
        t.text.length > 80 ? "..." : ""
      }</div>
    `;
      list.appendChild(div);
    });
}

document.addEventListener("DOMContentLoaded", updateUI);
browser.storage.onChanged.addListener(updateUI);
document.getElementById("download").addEventListener("click", () => {
  browser.runtime.sendMessage({ type: "DOWNLOAD_TWEETS" });
});
