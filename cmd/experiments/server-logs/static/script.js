document.addEventListener("DOMContentLoaded", function () {
  const sendRequestButton = document.getElementById("send-request");
  const endpointSelect = document.getElementById("endpoint");
  const responseData = document.getElementById("response-data");

  sendRequestButton.addEventListener("click", function () {
    const endpoint = endpointSelect.value;

    // Clear previous response
    responseData.textContent = "Loading...";

    // Send API request
    fetch(endpoint, {
      headers: {
        // Still need Authorization for the messy middleware
        Authorization: "Bearer fake-token-for-testing",
        // We don't set Content-Type here for GET requests
      },
    })
      .then((response) => {
        // Get the raw text response, regardless of Content-Type
        return response.text().then((text) => ({
          status: response.status,
          statusText: response.statusText,
          headers: response.headers,
          body: text,
        }));
      })
      .then((data) => {
        // Display the raw response text
        let output = `Status: ${data.status} ${data.statusText}\n`;
        output += "Headers:\n";
        data.headers.forEach((value, key) => {
          output += `  ${key}: ${value}\n`;
        });
        output += "\nBody:\n";
        output += data.body;
        responseData.textContent = output;
      })
      .catch((error) => {
        // Display network errors
        responseData.textContent = "Network Error: " + error.message;
      });
  });
});
