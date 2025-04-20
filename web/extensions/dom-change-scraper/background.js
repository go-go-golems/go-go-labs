console.log("[DOM\u2011MUTATION\u2011LOGGER] background script loaded");

browser.runtime.onMessage.addListener((msg) => {
  if (msg.type === "STORE_MUTATIONS") {
    console.log(
      `[DOM\u2011MUTATION\u2011LOGGER] Storing ${msg.records.length} records`
    );
    store(msg.records);
  } else if (msg.type === "DOWNLOAD_YAML") {
    downloadAsYAML();
  }
});

async function store(records) {
  const { mutations = [] } = await browser.storage.local.get("mutations");
  // Append new entries
  mutations.push(...records);
  await browser.storage.local.set({ mutations });
  console.log(
    `[DOM\u2011MUTATION\u2011LOGGER] total stored: ${mutations.length}`
  );
}

// --- YAML download ----------------------------------------------------------
function toYAML(objArray) {
  // ultra\u2011small, 80 % solution encoder (doesn't handle complex nesting)
  const esc = (s) => String(s).replace(/\\/g, "\\\\").replace(/"/g, '\\"');
  return objArray
    .map((o) =>
      Object.entries(o)
        .map(([k, v]) => `  ${k}: "${esc(v)}"`)
        .join("\n")
    )
    .map((block) => `- ${block.trimStart()}`)
    .join("\n");
}

async function downloadAsYAML() {
  const { mutations = [] } = await browser.storage.local.get("mutations");
  if (!mutations.length) return;

  const yaml = toYAML(mutations);
  const blob = new Blob([yaml], { type: "text/yaml" });
  const url = URL.createObjectURL(blob);
  const filename = `dom-mutations_${new Date()
    .toISOString()
    .replace(/:/g, "-")}.yaml`;

  try {
    await browser.downloads.download({ url, filename, saveAs: true });
    console.log(
      `[DOM\u2011MUTATION\u2011LOGGER] YAML download triggered → ${filename}`
    );
  } catch (e) {
    console.error("[DOM\u2011MUTATION\u2011LOGGER] YAML download failed:", e);
  }
}
