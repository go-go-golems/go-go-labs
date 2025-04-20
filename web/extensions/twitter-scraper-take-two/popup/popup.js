console.log("Popup script loaded.");

document.getElementById("download").addEventListener("click", () => {
  console.log("Download button clicked. Sending message to background script.");
  browser.runtime
    .sendMessage({ type: "DOWNLOAD_TWEETS" })
    .then(() => {
      console.log("Message sent successfully.");
      // Optionally close the popup
      // window.close();
    })
    .catch((error) => {
      console.error("Error sending message to background script:", error);
      alert(`Could not request download: ${error.message}`);
    });
});
