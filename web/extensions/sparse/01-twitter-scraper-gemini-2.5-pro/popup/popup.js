console.log("Popup script running.");

const itemCountElement = document.getElementById("item-count");
const dataListElement = document.getElementById("data-list");
const downloadButton = document.getElementById("download-button");
const clearButton = document.getElementById("clear-button"); // Get clear button

// Function to render the list of items
function renderList(data) {
  dataListElement.innerHTML = ""; // Clear previous content

  if (!Array.isArray(data) || data.length === 0) {
    dataListElement.innerHTML =
      "<p>No data collected yet. Browse Twitter/X!</p>";
    return;
  }

  // Display items (e.g., newest first)
  data
    .slice()
    .reverse()
    .forEach((item) => {
      const itemDiv = document.createElement("div");
      itemDiv.className = "data-item";

      const idSpan = document.createElement("span");
      idSpan.className = "item-id";
      idSpan.textContent = `[${item.id}]`; // Show ID clearly

      const textSpan = document.createElement("span");
      textSpan.className = "item-text";
      // Display a snippet of the text
      textSpan.textContent = item.text
        ? item.text.substring(0, 100) + (item.text.length > 100 ? "..." : "")
        : "(No text)";
      textSpan.title = item.text; // Show full text on hover

      itemDiv.appendChild(idSpan);
      itemDiv.appendChild(textSpan);
      dataListElement.appendChild(itemDiv);
    });
}

// Function to update the popup display from storage
async function updatePopup() {
  console.log("Updating popup display...");
  try {
    const result = await browser.storage.local.get("collectedDataStore");
    const data = result.collectedDataStore || [];

    itemCountElement.textContent = data.length;
    renderList(data);
  } catch (error) {
    console.error("Error fetching data from storage for popup:", error);
    dataListElement.innerHTML = "<p>Error loading data.</p>";
    itemCountElement.textContent = "Err";
  }
}

// Add event listener for the download button
downloadButton.addEventListener("click", () => {
  console.log("Download button clicked.");
  browser.runtime
    .sendMessage({ type: "DOWNLOAD_DATA" })
    .catch((error) => console.error("Error sending download message:", error));
});

// Add event listener for the clear button
clearButton.addEventListener("click", async () => {
  console.log("Clear button clicked.");
  if (confirm("Are you sure you want to clear all collected data?")) {
    try {
      await browser.storage.local.set({ collectedDataStore: [] });
      console.log("Cleared data from storage.");
      // Manually update the popup immediately after clearing
      itemCountElement.textContent = "0";
      renderList([]);
    } catch (error) {
      console.error("Error clearing data:", error);
      alert("Failed to clear data. See console for details.");
    }
  }
});

// Listen for changes in storage and update the popup live
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.collectedDataStore) {
    console.log("Storage changed, updating popup...");
    const newData = changes.collectedDataStore.newValue || [];
    itemCountElement.textContent = newData.length;
    renderList(newData); // Re-render the list with new data
  }
});

// Initial update when the popup is opened
document.addEventListener("DOMContentLoaded", updatePopup);
