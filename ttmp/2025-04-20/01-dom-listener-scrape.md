Below is a minimal, self‑contained WebExtension that records every DOM mutation (added nodes, attribute changes, text changes) on any page you visit and shows them in a popup list. Drop the files into a folder such as `dom‑mutation‑logger/`, load the extension temporarily from `about:debugging`, and start browsing.

---

### Project structure

```
dom-mutation-logger/
├── manifest.json
├── background.js
├── content.js
└── popup/
    ├── popup.html
    └── popup.js
```

---

### `manifest.json`

```json
{
  "manifest_version": 3,
  "name": "DOM Mutation Logger",
  "description": "Watches DOM mutations and lists them in a popup.",
  "version": "1.0.0",
  "permissions": ["storage"],
  "host_permissions": ["<all_urls>"],
  "background": {
    "scripts": ["background.js"]       // Firefox uses persistent scripts
  },
  "content_scripts": [
    {
      "matches": ["<all_urls>"],
      "js": ["content.js"],
      "run_at": "document_start"
    }
  ],
  "action": {
    "default_popup": "popup/popup.html",
    "default_title": "Show DOM mutations"
  }
}
```

---

### `content.js`

```js
// Record a simple, serialisable snapshot for each mutation entry.
function serialiseMutation(m) {
  const ts = new Date().toISOString();
  switch (m.type) {
    case "childList":
      return [...m.addedNodes].map((n) => ({
        ts,
        type: "addedNode",
        nodeName: n.nodeName,
        snippet: n.outerHTML?.slice(0, 120) || n.textContent.slice(0, 120)
      }));
    case "attributes":
      return [{
        ts,
        type: "attributeChange",
        nodeName: m.target.nodeName,
        attr: m.attributeName,
        newValue: m.target.getAttribute(m.attributeName)
      }];
    case "characterData":
      return [{
        ts,
        type: "textChange",
        nodeName: m.target.parentNode?.nodeName || "#text",
        newValue: m.target.data.slice(0, 120)
      }];
    default:
      return [];
  }
}

const observer = new MutationObserver((mutations) => {
  const records = mutations.flatMap(serialiseMutation);
  if (records.length) {
    browser.runtime.sendMessage({ type: "STORE_MUTATIONS", records });
  }
});

observer.observe(document.documentElement, {
  childList: true,
  subtree: true,
  attributes: true,
  characterData: true
});
```

---

### `background.js`

```js
async function store(records) {
  const { mutations = [] } = await browser.storage.local.get("mutations");
  // Append new entries
  mutations.push(...records);
  await browser.storage.local.set({ mutations });
}

browser.runtime.onMessage.addListener((msg) => {
  if (msg.type === "STORE_MUTATIONS") {
    store(msg.records);
  }
});
```

---

### `popup/popup.html`

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <style>
    body { font-family: Arial, sans-serif; min-width: 320px; padding: 12px; }
    h1   { margin: 0 0 12px; font-size: 16px; }
    #list { max-height: 400px; overflow-y: auto; border: 1px solid #ccc; padding: 6px; }
    .item { margin-bottom: 8px; border-bottom: 1px solid #eee; padding-bottom: 6px; }
    .ts   { color: #666; font-size: 11px; }
    .type { font-weight: bold; }
    pre   { white-space: pre-wrap; word-break: break-word; margin: 4px 0 0; }
  </style>
</head>
<body>
  <h1>DOM Mutations</h1>
  <div id="list"></div>
  <script src="popup.js"></script>
</body>
</html>
```

---

### `popup/popup.js`

```js
async function render() {
  const { mutations = [] } = await browser.storage.local.get("mutations");
  const list = document.getElementById("list");
  list.innerHTML = "";

  if (mutations.length === 0) {
    list.textContent = "No mutations recorded yet.";
    return;
  }

  // Newest first
  mutations.slice().reverse().forEach((m) => {
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
```

Load the extension, refresh any page, open the toolbar icon, and you’ll see an ever‑growing log of added nodes, attribute updates, and text changes.

---

Below are **surgical edits** to turn the “DOM Mutation Logger” into a **logged‑to‑console** tool that can also **export the history as a YAML file**.

---

## 1 — `manifest.json`

Add the `downloads` permission (needed for programmatic file saves):

```diff
@@
-  "permissions": ["storage"],
+  "permissions": ["storage", "downloads"],
```

---

## 2 — `content.js`  (logging added)

```diff
-const observer = new MutationObserver((mutations) => {
-  const records = mutations.flatMap(serialiseMutation);
+const observer = new MutationObserver((mutations) => {
+  console.log(`[DOM‑MUTATION‑LOGGER] observed ${mutations.length} raw mutations`);
+  const records = mutations.flatMap(serialiseMutation);
   if (records.length) {
+    console.table(records);
     browser.runtime.sendMessage({ type: "STORE_MUTATIONS", records });
   }
 });
```

Everything else stays the same.  
You now get a nice `console.table` in every page you’re inspecting.

---

## 3 — `background.js`  (download handlers & verbose logs)

```diff
 console.log("[DOM‑MUTATION‑LOGGER] background script loaded");
 
 browser.runtime.onMessage.addListener((msg) => {
   if (msg.type === "STORE_MUTATIONS") {
+    console.log(`[DOM‑MUTATION‑LOGGER] Storing ${msg.records.length} records`);
     store(msg.records);
+  } else if (msg.type === "DOWNLOAD_YAML") {
+    downloadAsYAML();
   }
 });
 
 async function store(records) {
@@
   await browser.storage.local.set({ mutations });
+  console.log(`[DOM‑MUTATION‑LOGGER] total stored: ${mutations.length}`);
 }
+
+// --- YAML download ----------------------------------------------------------
+function toYAML(objArray) {
+  // ultra‑small, 80 % solution encoder (doesn’t handle complex nesting)
+  const esc = (s) =>
+    String(s)
+      .replace(/\\/g, "\\\\")
+      .replace(/"/g, '\\"');
+  return objArray
+    .map((o) =>
+      Object.entries(o)
+        .map(([k, v]) => `  ${k}: "${esc(v)}"`)
+        .join("\n")
+    )
+    .map((block) => `- ${block.trimStart()}`)
+    .join("\n");
+}
+
+async function downloadAsYAML() {
+  const { mutations = [] } = await browser.storage.local.get("mutations");
+  if (!mutations.length) return;
+
+  const yaml = toYAML(mutations);
+  const blob = new Blob([yaml], { type: "text/yaml" });
+  const url = URL.createObjectURL(blob);
+  const filename = `dom-mutations_${new Date()
+    .toISOString()
+    .replace(/:/g, "-")}.yaml`;
+
+  try {
+    await browser.downloads.download({ url, filename, saveAs: true });
+    console.log(`[DOM‑MUTATION‑LOGGER] YAML download triggered → ${filename}`);
+  } catch (e) {
+    console.error("[DOM‑MUTATION‑LOGGER] YAML download failed:", e);
+  }
+}
```

---

## 4 — `popup/popup.html`  (+ new button)

```diff
   <div id="list"></div>
+
+  <button id="download-yaml" style="margin-top:8px">
+    Download YAML
+  </button>
   <script src="popup.js"></script>
```

---

## 5 — `popup/popup.js`  (download‑button listener)

```diff
 browser.storage.onChanged.addListener((ch, area) => {
   if (area === "local" && ch.mutations) render();
 });
 
+document
+  .getElementById("download-yaml")
+  .addEventListener("click", () =>
+    browser.runtime.sendMessage({ type: "DOWNLOAD_YAML" })
+  );
```

---

### That’s it — reload the temporary add‑on

1.  Go to `about:debugging`, click **Reload** on the extension.  
2.  Watch mutations stream in the page console (`console.table`).  
3.  Open the toolbar popup → **Download YAML**.

You now have **console logging** plus a clean **YAML export** of every stored DOM change.