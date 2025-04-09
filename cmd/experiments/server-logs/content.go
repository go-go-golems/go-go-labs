package main

// Hard-coded template
var indexTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <header>
        <h1>Log Server Dashboard</h1>
        <p>Server Time: {{.ServerTime}}</p>
    </header>
    <main>
        <section class="stats">
            <h2>System Statistics</h2>
            <ul>
                <li>Total Logs: {{.LogCount}}</li>
                <li>Users: {{.UserCount}}</li>
                <li>API Requests: {{index .Metrics "requests"}}</li>
                <li>Errors: {{index .Metrics "errors"}}</li>
                <li>Active Users: {{index .Metrics "active_users"}}</li>
            </ul>
        </section>
        
        <section class="recent-logs">
            <h2>Recent Logs</h2>
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Timestamp</th>
                        <th>Level</th>
                        <th>Message</th>
                        <th>Source</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .RecentLogs}}
                    <tr class="log-level-{{.Level}}">
                        <td>{{.Id}}</td>
                        <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
                        <td>{{.Level}}</td>
                        <td>{{.Message}}</td>
                        <td>{{.Source}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>
        
        <section class="api-test">
            <h2>API Test Console</h2>
            <div class="api-form">
                <div class="form-group">
                    <label for="endpoint">Endpoint:</label>
                    <select id="endpoint">
                        <option value="/api/logs">GET /api/logs</option>
                        <option value="/api/logs/1">GET /api/logs/1</option>
                        <option value="/api/users">GET /api/users</option>
                        <option value="/api/metrics">GET /api/metrics</option>
                    </select>
                </div>
                <button id="send-request">Send Request</button>
            </div>
            <div class="response">
                <h3>Response:</h3>
                <pre id="response-data"></pre>
            </div>
        </section>
    </main>
    <footer>
        <p>Messy Log Server Example © 2023</p>
    </footer>
    <script src="/static/script.js"></script>
</body>
</html>
`

// CSS content
var cssContent = `
* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: Arial, sans-serif;
    line-height: 1.6;
    color: #333;
    padding: 20px;
    max-width: 1200px;
    margin: 0 auto;
}

header {
    background-color: #f4f4f4;
    padding: 20px;
    margin-bottom: 20px;
    border-radius: 5px;
}

h1, h2, h3 {
    margin-bottom: 10px;
}

section {
    background-color: #fff;
    padding: 20px;
    margin-bottom: 20px;
    border-radius: 5px;
    box-shadow: 0 2px 5px rgba(0,0,0,0.1);
}

.stats ul {
    list-style: none;
}

.stats li {
    padding: 5px 0;
    border-bottom: 1px solid #eee;
}

table {
    width: 100%;
    border-collapse: collapse;
}

table th, table td {
    padding: 10px;
    text-align: left;
    border-bottom: 1px solid #ddd;
}

table th {
    background-color: #f4f4f4;
}

.log-level-error {
    background-color: #ffebee;
}

.log-level-warning {
    background-color: #fff8e1;
}

.log-level-info {
    background-color: #e8f5e9;
}

.api-form {
    margin-bottom: 20px;
}

.form-group {
    margin-bottom: 15px;
}

label {
    display: block;
    margin-bottom: 5px;
}

select, input {
    width: 100%;
    padding: 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
}

button {
    padding: 10px 15px;
    background-color: #4CAF50;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

button:hover {
    background-color: #45a049;
}

.response {
    background-color: #f8f8f8;
    padding: 15px;
    border-radius: 4px;
    height: 300px;
    overflow: auto;
}

pre {
    white-space: pre-wrap;
    font-family: monospace;
}

footer {
    text-align: center;
    padding: 20px;
    margin-top: 20px;
    background-color: #f4f4f4;
    border-radius: 5px;
}
`

// JavaScript content
var jsContent = `
document.addEventListener('DOMContentLoaded', function() {
    const sendRequestButton = document.getElementById('send-request');
    const endpointSelect = document.getElementById('endpoint');
    const responseData = document.getElementById('response-data');
    
    sendRequestButton.addEventListener('click', function() {
        const endpoint = endpointSelect.value;
        
        // Clear previous response
        responseData.textContent = 'Loading...';
        
        // Send API request
        fetch(endpoint, {
            headers: {
                'Authorization': 'Bearer fake-token-for-testing',
                'Content-Type': 'application/json'
            }
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok: ' + response.status);
            }
            return response.json();
        })
        .then(data => {
            // Format and display the response
            responseData.textContent = JSON.stringify(data, null, 2);
        })
        .catch(error => {
            responseData.textContent = 'Error: ' + error.message;
        });
    });
});
`
