console.log("Background script loading...");

// Store collected tweets here
let collectedData = [];

// Function to load existing data from storage when the script starts
async function loadInitialData() {
  try {
    const result = await browser.storage.local.get("collectedDataStore");
    if (result.collectedDataStore) {
      collectedData = result.collectedDataStore;
      console.log(`Loaded ${collectedData.length} tweets from storage.`);
    } else {
      console.log("No tweets found in storage.");
    }
  } catch (error) {
    console.error("Error loading data from storage:", error);
  }
}

// Function to save tweet received from content script
async function saveData(item) {
  console.log("Attempting to save tweet:", item);

  // Check if a tweet with the same ID already exists
  const alreadyExists = collectedData.some((d) => d.id === item.id);
  if (alreadyExists) {
    console.log(`Tweet ID ${item.id} already exists. Skipping.`);
    return;
  }

  // Add the tweet to our collection
  collectedData.push(item);

  // Save to local storage
  try {
    await browser.storage.local.set({ collectedDataStore: collectedData });
    console.log(
      `Tweet ID ${item.id} saved. Total tweets: ${collectedData.length}`
    );
  } catch (error) {
    console.error("Error saving data to storage:", error);
    // Remove the item if saving failed to keep array and storage in sync
    collectedData.pop();
  }
}

// Function to handle download requests (triggered by the popup)
async function handleDownloadRequest() {
  console.log("Download request received.");
  if (collectedData.length === 0) {
    console.log("No tweets collected to download.");
    return;
  }

  console.log(`Preparing download for ${collectedData.length} tweets.`);

  try {
    // Convert the collected tweets array into a pretty-printed JSON string
    const jsonString = JSON.stringify(collectedData, null, 2);

    // Create a Blob object from the JSON string
    const blob = new Blob([jsonString], { type: "application/json" });

    // Create an object URL for the Blob
    const url = URL.createObjectURL(blob);

    // Generate a filename with current timestamp
    const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
    const filename = `twitter_data_${timestamp}.json`;

    // Use the browser.downloads API to save the file
    await browser.downloads.download({
      url: url,
      filename: filename,
      saveAs: true, // Ask user where to save
    });

    console.log(`Download initiated for ${filename}`);
  } catch (err) {
    console.error("Download failed:", err);
  }
}

// Function to clear all stored data
async function handleClearRequest() {
  console.log("Clear data request received.");
  collectedData = [];
  try {
    await browser.storage.local.remove("collectedDataStore");
    console.log("All tweet data cleared from storage.");
    return { success: true };
  } catch (error) {
    console.error("Error clearing data:", error);
    return { success: false, error: error.message };
  }
}

// Listener for messages from other parts of the extension
browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
  console.log("Background script received message:", message);

  if (message.type === "SAVE_DATA") {
    // Received data from content script
    saveData(message.payload);
    return true; // Indicate async handling
  } else if (message.type === "DOWNLOAD_DATA") {
    // Received request from popup
    handleDownloadRequest();
    return true; // Indicate async handling
  } else if (message.type === "CLEAR_DATA") {
    // Received request to clear data
    handleClearRequest().then(sendResponse);
    return true; // Indicate async handling
  } else if (message.type === "GET_DATA_COUNT") {
    // Popup requesting count
    sendResponse({ count: collectedData.length });
  } else if (message.type === "GET_RECENT_DATA") {
    // Popup requesting recent items for display
    const recentItems = collectedData.slice(-10).reverse(); // Last 10 items, newest first
    sendResponse({ items: recentItems });
  } else {
    console.warn("Unknown message type received:", message.type);
  }
});

// Load existing data when the background script starts
loadInitialData();

console.log("Background script loaded and ready.");
