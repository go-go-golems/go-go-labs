async function render() {
  const { mutations = [] } = await browser.storage.local.get("mutations");
  const list = document.getElementById("list");
  list.innerHTML = "";

  if (mutations.length === 0) {
    list.textContent = "No mutations recorded yet.";
    return;
  }

  // Newest first
  mutations
    .slice()
    .reverse()
    .forEach((m) => {
      const div = document.createElement("div");
      div.className = "item";
      div.innerHTML = `
      <span class="ts">${m.ts}</span><br>
      <span class="type">${m.type}</span> — <code>${m.nodeName}</code>
      <pre>${(m.snippet || m.attr + " = " + m.newValue || m.newValue || "")
        .toString()
        .replace(/</g, "&lt;")}</pre>`;
      list.appendChild(div);
    });
}

document.addEventListener("DOMContentLoaded", render);

// Live update when content script stores new mutations
browser.storage.onChanged.addListener((ch, area) => {
  if (area === "local" && ch.mutations) render();
});

document
  .getElementById("download-yaml")
  .addEventListener("click", () =>
    browser.runtime.sendMessage({ type: "DOWNLOAD_YAML" })
  );
