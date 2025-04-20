console.log("Background script loading...");

let collectedData = [];

async function loadInitialData() {
  try {
    // Use browser.storage.local which is preferred for extensions
    const result = await browser.storage.local.get("collectedDataStore");
    if (result.collectedDataStore && Array.isArray(result.collectedDataStore)) {
      collectedData = result.collectedDataStore;
      console.log(`Loaded ${collectedData.length} items from storage.`);
    } else {
      console.log("No valid data found in storage or storage is empty.");
      // Ensure collectedData is an array if storage was empty or corrupt
      collectedData = [];
      // Optionally initialize storage if it was empty/invalid
      await browser.storage.local.set({ collectedDataStore: [] });
    }
  } catch (error) {
    console.error("Error loading data from storage:", error);
    collectedData = []; // Reset to empty array on error
  }
}

async function saveData(item) {
  console.log("Attempting to save item:", item);

  if (!item || typeof item.id === "undefined") {
    console.warn("Attempted to save invalid item:", item);
    return;
  }

  const alreadyExists = collectedData.some((d) => d.id === item.id);
  if (alreadyExists) {
    console.log(`Item ID ${item.id} already exists. Skipping.`);
    return;
  }

  collectedData.push(item);
  try {
    await browser.storage.local.set({ collectedDataStore: collectedData });
    console.log(
      `Item ID ${item.id} saved. Total items: ${collectedData.length}`
    );
  } catch (error) {
    console.error("Error saving data to storage:", error);
    // Attempt to revert the in-memory change if save failed
    collectedData = collectedData.filter((d) => d.id !== item.id);
  }
}

async function handleDownloadRequest() {
  console.log("Download request received.");
  if (!Array.isArray(collectedData) || collectedData.length === 0) {
    console.log("No data collected to download.");
    // Optional: Consider sending a message back to the popup to inform the user
    return;
  }

  console.log(`Preparing download for ${collectedData.length} items.`);

  try {
    const jsonString = JSON.stringify(collectedData, null, 2);
    const blob = new Blob([jsonString], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
    const filename = `collected_twitter_data_${timestamp}.json`;

    // Use the browser.downloads API
    const downloadId = await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true, // Prompt user for save location
    });
    console.log(`Download initiated with ID: ${downloadId} for ${filename}`);

    // Firefox automatically revokes the object URL after the download starts
    // No need for URL.revokeObjectURL(url) in this context for Firefox.
  } catch (err) {
    console.error("Download failed:", err);
    // Revoke URL manually if download API failed before starting
    if (typeof url !== "undefined") {
      URL.revokeObjectURL(url);
    }
  }
}

// Listener for messages
browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
  console.log("Background script received message:", message, "from", sender);

  switch (message.type) {
    case "SAVE_DATA":
      saveData(message.payload);
      // No need to send a response back for this simple case
      return true; // Indicate async work might happen (good practice)
    case "DOWNLOAD_DATA":
      handleDownloadRequest();
      return true; // Indicate async handling
    case "GET_DATA_COUNT":
      // Respond synchronously with the current count
      sendResponse({ count: collectedData.length });
      return false; // Indicate synchronous response
    default:
      console.warn("Unknown message type received:", message.type);
      // Indicate no response will be sent
      return false;
  }
});

// Load existing data when the background script initializes
loadInitialData()
  .then(() => {
    console.log("Initial data loaded successfully.");
  })
  .catch((error) => {
    console.error("Failed to load initial data:", error);
  });

console.log("Background script loaded and event listeners attached.");
