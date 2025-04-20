console.log("Background script initialized");

let collectedData = [];

async function loadInitialData() {
  try {
    const result = await browser.storage.local.get("collectedDataStore");
    collectedData = result.collectedDataStore || [];
    console.log(`Loaded ${collectedData.length} items from storage`);
  } catch (error) {
    console.error("Storage load error:", error);
  }
}

async function saveData(item) {
  const exists = collectedData.some((d) => d.id === item.id);
  if (exists) return;

  collectedData.push(item);
  try {
    await browser.storage.local.set({ collectedDataStore: collectedData });
    console.log(`Saved item ${item.id}. Total: ${collectedData.length}`);
  } catch (error) {
    console.error("Storage save error:", error);
    collectedData.pop();
  }
}

async function handleDownload() {
  if (collectedData.length === 0) return;

  try {
    const json = JSON.stringify(collectedData, null, 2);
    const blob = new Blob([json], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const filename = `tweets-${Date.now()}.json`;

    await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true,
    });
    console.log("Download initiated");
  } catch (error) {
    console.error("Download failed:", error);
  }
}

browser.runtime.onMessage.addListener((msg) => {
  switch (msg.type) {
    case "SAVE_DATA":
      saveData(msg.payload);
      return true;
    case "DOWNLOAD_DATA":
      handleDownload();
      return true;
    case "GET_DATA":
      return Promise.resolve(collectedData);
  }
});

loadInitialData();
