console.log("Popup script running.");

const itemCountElement = document.getElementById("item-count");
const dataListElement = document.getElementById("data-list");
const downloadButton = document.getElementById("download-button");
const clearButton = document.getElementById("clear-button");

// Function to update the popup display
async function updatePopup() {
  console.log("Updating popup display...");

  try {
    // Get the data count from the background script
    const countResponse = await browser.runtime.sendMessage({
      type: "GET_DATA_COUNT",
    });
    const count = countResponse.count;

    // Update the count display
    itemCountElement.textContent = count;

    // Request recent data to display
    const dataResponse = await browser.runtime.sendMessage({
      type: "GET_RECENT_DATA",
    });
    const recentData = dataResponse.items || [];

    // Clear previous content
    dataListElement.innerHTML = "";

    if (count === 0) {
      // Display a message if no data
      dataListElement.innerHTML =
        "<p class='empty-message'>No tweets collected yet. Browse Twitter/X to start collecting!</p>";
    } else {
      // Display the recent items
      recentData.forEach((item) => {
        const itemDiv = document.createElement("div");
        itemDiv.className = "data-item";

        // Create author element
        const authorDiv = document.createElement("div");
        authorDiv.className = "author";
        authorDiv.textContent = item.author || "Unknown";
        itemDiv.appendChild(authorDiv);

        // Create text element
        const textDiv = document.createElement("div");
        textDiv.className = "text";
        textDiv.textContent = item.text;
        itemDiv.appendChild(textDiv);

        // Create meta info element with link to tweet
        const metaDiv = document.createElement("div");
        metaDiv.className = "meta";

        // Add timestamp if available
        if (item.timestamp) {
          const date = new Date(item.timestamp);
          const formattedDate = date.toLocaleString();
          metaDiv.textContent = `Posted: ${formattedDate} • `;
        }

        // Add link to original tweet
        const link = document.createElement("a");
        link.href = item.url;
        link.textContent = "View Tweet";
        link.target = "_blank"; // Open in new tab
        metaDiv.appendChild(link);

        itemDiv.appendChild(metaDiv);

        // Add the item to the list
        dataListElement.appendChild(itemDiv);
      });

      // If showing a limited number, add a note
      if (recentData.length < count) {
        const noteDiv = document.createElement("div");
        noteDiv.className = "empty-message";
        noteDiv.textContent = `Showing ${recentData.length} of ${count} tweets. Download JSON for complete data.`;
        dataListElement.appendChild(noteDiv);
      }
    }
  } catch (error) {
    console.error("Error updating popup:", error);
    dataListElement.innerHTML =
      "<p class='empty-message'>Error loading data.</p>";
  }
}

// Add event listener for the download button
downloadButton.addEventListener("click", () => {
  console.log("Download button clicked.");
  // Send a message to the background script to trigger the download
  browser.runtime
    .sendMessage({ type: "DOWNLOAD_DATA" })
    .catch((error) => console.error("Error sending download message:", error));
});

// Add event listener for the clear button
clearButton.addEventListener("click", async () => {
  console.log("Clear button clicked.");
  if (confirm("Are you sure you want to clear all collected data?")) {
    try {
      const response = await browser.runtime.sendMessage({
        type: "CLEAR_DATA",
      });
      if (response.success) {
        // Update the popup to reflect the cleared data
        updatePopup();
      } else {
        console.error("Error clearing data:", response.error);
      }
    } catch (error) {
      console.error("Error sending clear message:", error);
    }
  }
});

// Listen for changes in storage
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.collectedDataStore) {
    console.log("Storage changed, updating popup...");
    updatePopup(); // Re-render the popup content
  }
});

// Initial update when the popup is opened
document.addEventListener("DOMContentLoaded", updatePopup);
