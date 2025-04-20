document.addEventListener('DOMContentLoaded', initPopup);

async function initPopup() {
  const countEl = document.getElementById('count');
  const listEl = document.getElementById('tweet-list');
  const downloadBtn = document.getElementById('download');

  // Load initial data
  const data = await browser.runtime.sendMessage({ type: "GET_DATA" });
  updateDisplay(data, countEl, listEl);

  // Setup storage listener
  browser.storage.onChanged.addListener(() => {
    browser.runtime.sendMessage({ type: "GET_DATA" })
      .then(data => updateDisplay(data, countEl, listEl));
  });

  // Download handler
  downloadBtn.addEventListener('click', () => {
    browser.runtime.sendMessage({ type: "DOWNLOAD_DATA" });
  });
}

function updateDisplay(data, countEl, listEl) {
  countEl.textContent = data.length;
  listEl.innerHTML = data.length ? 
    data.slice(-10).reverse().map(tweet => `
      <div class="tweet">
        <strong>@${tweet.author}</strong>
        <p>${tweet.text.slice(0, 80)}${tweet.text.length > 80 ? '...' : ''}</p>
      </div>
    `).join('') : 
    '<div class="tweet">No tweets collected yet</div>';
} 