# Managing n8n Workflows via the REST API (Self-Hosted Guide)

**Note:** This guide assumes you have a self-hosted n8n instance and an API key for authentication. We will cover how to use n8n’s REST API to create, modify, trigger, and monitor workflows programmatically. All examples use existing n8n nodes (no custom nodes needed) and are given in Python (using the `requests` library) – they can be adapted to JavaScript easily.

## 1. Overview of the n8n REST API and Authentication

**API Base URL & Version:** In a typical self-hosted setup, the REST endpoints are under your n8n server URL with path `/api/v1/...`. For example, if your n8n is at `http://localhost:5678`, an endpoint could be `http://localhost:5678/api/v1/workflows`. (Replace host/port with your domain and port. If n8n is served under a subdirectory, include that in the path.)

**Authentication:** n8n uses API keys to authenticate API calls. You must create an API key in the n8n UI (Settings → **n8n API** → "Create an API key") and copy it. Include this key in an HTTP header for every API request:

```
X-N8N-API-KEY: <your-api-key>
```

For example, to list all active workflows on a self-hosted instance, you would call:

```bash
curl -X GET 'http://<YOUR_N8N_HOST>:<PORT>/api/v1/workflows?active=true' \
     -H 'X-N8N-API-KEY: <your-api-key>' \
     -H 'accept: application/json'
```

.

In Python, this would be:

```python
import requests

api_url = "http://localhost:5678/api/v1/workflows?active=true"
headers = {"X-N8N-API-KEY": "<your-api-key>"}
resp = requests.get(api_url, headers=headers)
print(resp.json())  # list of active workflows
```

> **Tip:** On self-hosted n8n, you can use the built-in **API Playground** (Swagger UI) to explore available endpoints and test calls. This can be accessed via Settings → _API_ in the n8n UI, if enabled. It’s a safe way to try endpoints on a test workflow before automating against live data.

## 2. Creating a New Workflow via API

To create a workflow programmatically, use the `POST /workflows` endpoint. The request body should be JSON with at least a **name** for the workflow, and typically a list of **nodes** plus their **connections**.

- **Endpoint:** `POST /api/v1/workflows`
- **Headers:** `X-N8N-API-KEY: <your-key>`, `Content-Type: application/json`
- **Body:** JSON object including:

  - `name`: a string name for the workflow.
  - `nodes`: an array of node objects (each defining a node’s type and parameters).
  - `connections`: an object defining how nodes are connected.

**Example – Create a simple workflow:** The following example creates a workflow with a Webhook trigger and a Respond node, effectively setting up a basic API endpoint. The Webhook node will listen for HTTP requests, and the Respond node will reply with a fixed message.

```python
import requests

workflow = {
    "name": "Simple API Endpoint Workflow",
    "nodes": [
        {
            "name": "Webhook Trigger",
            "type": "n8n-nodes-base.webhook",
            "typeVersion": 1,
            "position": [200, 300],
            "parameters": {
                "path": "my-endpoint",   # URL path for the webhook
                "method": "POST"         # HTTP method to listen for
            }
        },
        {
            "name": "Respond",
            "type": "n8n-nodes-base.respondToWebhook",
            "typeVersion": 1,
            "position": [400, 300],
            "parameters": {
                "responseMode": "onReceived",
                "responseData": "{{\"message\": \"Workflow executed successfully\"}}",
                "responseCode": 200
            }
        }
    ],
    "connections": {
        "Webhook Trigger": {
            "main": [ [ { "node": "Respond", "type": "main", "index": 0 } ] ]
        }
    }
}

api_url = "http://localhost:5678/api/v1/workflows"
headers = {"X-N8N-API-KEY": "<your-api-key>", "Content-Type": "application/json"}
resp = requests.post(api_url, headers=headers, json=workflow)
print(resp.status_code, resp.json().get("id"))
```

This will create a new workflow (inactive by default) and return its `id` in the response. In the above JSON:

- We defined two nodes: a **Webhook Trigger** (listens on path `/webhook/<unique-id>/my-endpoint` once activated) and a **Respond to Webhook** node that sends a JSON response.
- The `connections` object connects the Webhook’s output to the Respond node’s input. The structure `connections: { "Webhook Trigger": { "main": [[ {node: "Respond", ...} ]] } }` means _Webhook Trigger → Respond node_ on the default (“main”) output/input.

**Using existing nodes:** The `type` field of each node uses the internal name of the node (for example, `n8n-nodes-base.webhook` for the Webhook Trigger, `n8n-nodes-base.httpRequest` for an HTTP Request node, etc.). A good way to get these node definitions and parameters is to first create a workflow in the n8n GUI, then retrieve it via API to see the JSON structure. For instance, if you build a workflow in the editor, you can GET `/workflows/<id>` to see its `nodes` and `connections` JSON, and use that as a template for your API calls.

## 3. Adding and Connecting Nodes Programmatically

When creating or updating workflows via API, you define **nodes** and their **connections** in the JSON payload. Each node object typically contains:

- `name`: The node’s label (must be unique in the workflow).
- `type`: The internal type (e.g., `n8n-nodes-base.function` for a Function node).
- `typeVersion`: Version of the node (usually 1 for core nodes, or a specific version if the node has multiple versions).
- `position`: X, Y coordinates for the editor (not functionally important, but required for UI layout).
- `parameters`: An object of the node’s specific settings (e.g., for an HTTP Request node, URL and other request options).
- `credentials`: _(Optional)_ credentials data if the node uses an auth credential (usually just an ID reference to a stored credential).

**Connections structure:** The `connections` object defines how nodes link together. It is a mapping from a **source node name** to an array of outputs, each containing a list of connections. For most nodes with one output, it looks like:

```json
"connections": {
  "SourceNode": {
    "main": [
      [ { "node": "TargetNode", "type": "main", "index": 0 } ]
    ]
  }
}
```

This means _SourceNode’s main output -> TargetNode’s main input (at index 0)_. The `index` is used when a node has multiple inputs or outputs (0 for first input/output, 1 for second, etc.). For example, an IF node has two outputs (“true” at index 0, “false” at index 1), and a Merge node has multiple inputs – the `index` helps route connections properly.

**Adding a node to an existing workflow:** To add a node programmatically, you would retrieve the workflow JSON, modify the `nodes` array and `connections`, then send an update (via PUT or PATCH). For instance, suppose we want to add an **Email node** to the above “Simple API Endpoint Workflow” to send an email whenever the webhook is received:

1. **Fetch the current workflow:** `GET /api/v1/workflows/<workflowId>` to get the JSON (including existing nodes and connections).

2. **Append the new node:** e.g., add a Gmail node object to the `nodes` list.

3. **Connect it:** Modify `connections` so that the output of Webhook (or another node) connects to the new Gmail node, or perhaps connect the output of Respond to the Gmail (depending on desired order). Typically, you’d connect the Webhook Trigger to both the Respond node and the new Email node (using multiple connections from one output) or chain Respond -> Email. In n8n, you can have multiple connections from one node’s output (they will run in parallel). The connections JSON would then have two entries under `"Webhook Trigger"`’s `main` array. For example:

   ```json
   "Webhook Trigger": {
     "main": [
       [
         { "node": "Respond", "type": "main", "index": 0 },
         { "node": "Send Email", "type": "main", "index": 0 }
       ]
     ]
   }
   ```

   This connects **Webhook Trigger → Respond** and **Webhook Trigger → Send Email** in parallel. Alternatively, to chain sequentially, connect Webhook → Email, then Email → Respond.

4. **Update via API:** Send a `PUT /api/v1/workflows/<id>` with the modified JSON (or `PATCH` if supported for partial updates). Ensure you include the full `nodes` and `connections` if using PUT. For example:

```python
workflow_data = requests.get(f"{api_url}/{workflow_id}", headers=headers).json()
# modify workflow_data: add node and connections...
requests.put(f"{api_url}/{workflow_id}", headers=headers, json=workflow_data)
```

After updating, you can GET the workflow again to verify the new node appears.

> **Important:** The workflow `id` is assigned by n8n. You cannot specify a custom `id` when creating a workflow (it will be ignored). Each created workflow gets a new unique ID. If you need to reference workflows (e.g., for Execute Workflow nodes or webhooks), use the new IDs or update references accordingly.

## 4. Querying Workflows (Listing and Retrieving)

You can query existing workflows via the API to get their details or list multiple workflows:

- **List Workflows:** `GET /api/v1/workflows` – returns a list of workflow summaries. You can filter by active status or other criteria. For example, `GET /workflows?active=true` returns only active (enabled) workflows. The API supports pagination parameters as well (e.g., `limit` and `offset`), and possibly filtering by tags or other fields if needed.

- **Get Workflow by ID:** `GET /api/v1/workflows/<id>` – returns the full JSON definition of the workflow with the given ID, including all its nodes and connections. This is useful for retrieving a workflow to inspect or copy. For example, if you created a workflow via the GUI and want to duplicate it via API, you could fetch it and reuse its `nodes` and `connections` in a create call.

**Example – List and Get:**

```python
# List all workflows (with optional filters)
resp = requests.get("http://localhost:5678/api/v1/workflows?limit=50", headers=headers)
workflows = resp.json().get('data') or resp.json()  # depending on n8n version, may be in .data
for wf in workflows:
    print(wf['id'], wf['name'], "ACTIVE" if wf['active'] else "INACTIVE")

# Get a specific workflow
workflow_id = workflows[0]['id']
resp = requests.get(f"http://localhost:5678/api/v1/workflows/{workflow_id}", headers=headers)
workflow_details = resp.json()
print(workflow_details['name'], "has", len(workflow_details['nodes']), "nodes")
```

The list endpoint returns basic info like `id`, `name`, `active` state, etc., for each workflow. The single-workflow GET returns the full definition (which can be quite large if the workflow has many nodes).

## 5. Updating and Deleting Workflows

**Updating a workflow:** Use `PUT /api/v1/workflows/<id>` to replace a workflow’s definition with a new one, or `PATCH /api/v1/workflows/<id>` for partial updates (if supported by your n8n version). Common updates include renaming a workflow, toggling its active state, or altering nodes:

- _Rename or change settings:_ You can send a JSON with a new `"name"` or change the `"active"` flag (true/false). For instance, to activate a workflow, you could do a PATCH with `{"active": true}`. (Alternatively, n8n may provide a convenience endpoint to activate, but setting the field via update is straightforward.)
- _Add/remove nodes:_ As discussed in Section 3, you can modify the nodes array and connections then PUT the entire workflow JSON.
- _Update nodes:_ You can also update parameters of existing nodes by editing the JSON (for example, change an HTTP Request node’s URL or an Email node’s recipient) and pushing the update.

**Example – Toggle workflow active state (activate/deactivate):**

```python
workflow_id = "123"  # use your workflow's ID
# Activate the workflow
requests.patch(f"{api_url}/{workflow_id}", headers=headers,
               json={"active": True})
# Deactivate the workflow
requests.patch(f"{api_url}/{workflow_id}", headers=headers,
               json={"active": False})
```

_(If PATCH is not supported, you would GET the workflow, set `"active": true/false` in the JSON, and PUT it back.)_ Activating a workflow enables triggers (like Webhook, Cron, etc.) to start listening or scheduling. Deactivating stops those triggers.

**Deleting a workflow:** Use `DELETE /api/v1/workflows/<id>`. This will permanently remove the workflow from n8n. Example:

```python
resp = requests.delete(f"{api_url}/{workflow_id}", headers=headers)
if resp.status_code == 200:
    print("Workflow deleted.")
```

Be cautious with deletions – they cannot be undone via API. (It’s wise to backup workflows using the export feature or via GET calls before bulk deletions.)

## 6. Triggering Workflows via API and Monitoring Executions

Unlike some automation tools, n8n’s public API **does not have a direct “run workflow” endpoint** that you can call with a workflow ID to execute it on demand. Instead, workflows are typically triggered by their trigger nodes or schedules. To externally trigger a workflow through an API call, the recommended approach is to use a **Webhook node** in the workflow:

- **Webhook Trigger:** Add a **Webhook** node as the starting node of your workflow (with a unique path and method). Activate the workflow. n8n will then expose a URL (on your n8n server) that you can call to trigger that workflow. The URL format is usually:

  ```
  http(s)://<your-n8n-host>/webhook/<workflow-id>/<path>
  ```

  - For example, if your Webhook node has path `my-endpoint` and the workflow ID is `20`, an active webhook URL might be: `http://localhost:5678/webhook/20/my-endpoint` (for immediate execution) and a test URL for manual execution in editor (with an extra test segment). **Always use the production URL (no `test` segment) for external calls on active workflows.**

- **Call the Webhook URL:** You can use any HTTP client (curl, requests, etc.) to send a request to this URL. The workflow will start execution when n8n receives that request. For instance, triggering via Python:

  ```python
  import requests
  data = {"name": "Alice", "email": "alice@example.com"}  # sample payload
  resp = requests.post("http://localhost:5678/webhook/20/my-endpoint", json=data)
  print(resp.status_code, resp.text)
  ```

  If your workflow has a Respond node, the `resp.text` will contain whatever response your workflow sent. If not, n8n’s default webhook response is typically a generic confirmation or empty 200 response.

- **Alternative triggers:** You can also trigger workflows on schedules (Cron node) or other trigger nodes (like polling triggers, IMAP email triggers, etc.), but those run automatically and not via an external call. For manual ad-hoc triggering, Webhook is the most straightforward approach.

**Monitoring Executions:** n8n logs workflow executions, and the API provides endpoints to query them:

- `GET /api/v1/executions` – retrieves past execution records. You can filter by workflow ID or status. For example:

  - `GET /executions?workflowId=20` – all executions of workflow ID 20.
  - `GET /executions?status=success` – executions that finished successfully (status can be `success`, `error`, or `waiting`).
  - You can combine filters and use pagination (`limit`, `offset`). The response will include each execution’s ID, status, start/end times, workflow ID, etc..

- `GET /api/v1/executions/<execId>` – get details of a specific execution. **Note:** By default, n8n might save full execution data only for failed executions or for a short time (to conserve database space). If full data is available, this endpoint will include the node-by-node data (inputs/outputs). If not (e.g., if you disabled saving successful execution data), you may only get a summary. Ensure your n8n config (`EXECUTIONS_DATA_SAVE_ON_SUCCESS` etc.) is set to retain data if you need detailed info.

**Example – Query executions:**

```python
# List last 5 executions of a workflow
wfid = 20
resp = requests.get(f"http://localhost:5678/api/v1/executions?workflowId={wfid}&limit=5", headers=headers)
for exec in resp.json():
    print(exec["id"], exec["status"], exec["finished"], exec["startedAt"])
# Get details of one execution
exec_id = resp.json()[0]["id"]
details = requests.get(f"http://localhost:5678/api/v1/executions/{exec_id}", headers=headers).json()
print(details.keys())  # might include 'data' with nodes if saved, plus meta info
```

This can be used to build monitoring dashboards or to programmatically check if a workflow run succeeded or failed. For example, you could trigger a workflow via webhook and then poll `/executions?workflowId=<id>&lastId=<previousId>` to detect the new run’s status.

If you need real-time notifications of execution results, consider building that logic into your workflow (e.g., using the Respond node to send an HTTP callback or a message) because the API itself is more for querying after the fact.

## 7. Examples of Workflow Templates from the Community

The n8n community has a large repository of pre-built workflows (templates) – over 2,000 automation workflows are available for inspiration. You can use these as templates by obtaining their JSON and creating similar workflows via the API. Below are a couple of realistic examples and how they would be represented:

**Example A: Form Submission to Google Sheets and Email** – Imagine a workflow that handles form submissions by saving data to a Google Sheet and then emailing a confirmation. In n8n, this could be composed of: a Webhook trigger → Google Sheets node (Append Row) → Gmail node (Send Email). A simplified JSON snippet for the core nodes might look like:

```json
"nodes": [
  {
    "name": "Webhook Trigger",
    "type": "n8n-nodes-base.webhook",
    "typeVersion": 1,
    "position": [300, 200],
    "parameters": { "path": "form_submit", "method": "POST" }
  },
  {
    "name": "Google Sheets",
    "type": "n8n-nodes-base.googleSheets",
    "typeVersion": 1,
    "position": [500, 200],
    "parameters": {
      "operation": "append",
      "sheetId": "<GOOGLE_SHEET_ID>",
      "range": "Sheet1!A:D",
      "columns": [
        { "columnName": "Name",  "value": "={{$json[\"name\"]}}" },
        { "columnName": "Email", "value": "={{$json[\"email\"]}}" }
      ]
    },
    "credentials": {
      "googleApi": { "id": "<CREDENTIAL_ID>" }
    }
  },
  {
    "name": "Send Email",
    "type": "n8n-nodes-base.gmailSendEmail",  // or n8n-nodes-base.emailSend if using SMTP
    "typeVersion": 1,
    "position": [700, 200],
    "parameters": {
      "to": "={{$node[\"Webhook Trigger\"].json[\"email\"]}}",
      "subject": "Thanks for your submission",
      "text": "Hello {{$node[\"Webhook Trigger\"].json[\"name\"]}}, thanks for submitting the form."
    },
    "credentials": {
      "googleApi": { "id": "<CREDENTIAL_ID_FOR_GMAIL>" }
    }
  }
],
"connections": {
  "Webhook Trigger": {
    "main": [ [ { "node": "Google Sheets", "type": "main", "index": 0 } ] ]
  },
  "Google Sheets": {
    "main": [ [ { "node": "Send Email", "type": "main", "index": 0 } ] ]
  }
}
```

This corresponds to a workflow where a webhook receives a JSON payload with `name` and `email`, then the Google Sheets node appends that data to a spreadsheet, and finally the Gmail node sends an email to the provided address. Such a workflow was described as a use-case in the community. When creating via API, you’d include real Google Sheet ID, proper credentials IDs (which you can obtain by first configuring credentials in the n8n UI or via API), etc. This example uses only built-in nodes.

**Example B: Scheduled Data Fetch and Notification** – A workflow that runs every day to fetch data from an API and send a report. This could use a Cron trigger → HTTP Request node → Email or Slack node. For instance:

- Cron node (trigger every day at 9am)
- HTTP Request node (GET some API endpoint, e.g., weather data or stock prices)
- Function or Set node (prepare a summary message)
- Slack node (post the message to a channel)

Using the API, you’d construct nodes for each of these and connect Cron → HTTP → Function → Slack. (For brevity, the JSON is omitted here, but you can find similar templates in the community library.)

**Finding and using templates:** You can browse the [n8n workflow template library](https://n8n.io/workflows/) for pre-built examples. Each template page often has an “Use for free” or “Download JSON” option. You can import that JSON directly via the n8n UI or copy its structure for API use. For example, the “Creating an API endpoint” template (Webhook + Respond) we mentioned is available in the templates library and can serve as a reference. Always make sure to update any credentials or IDs when reusing a template JSON in your environment.

## 8. Additional Tips and References

- **Official Documentation:** The [n8n REST API docs](https://docs.n8n.io/api/) provide reference for all endpoints and object schemas. Key sections include Authentication, the endpoint reference, and examples of requests.

- **n8n API Node:** n8n itself provides an internal "**n8n API**" node that you can use _within_ workflows to call the API. This is useful if you want one workflow to manage others (e.g., a master workflow that updates workflows via API). It’s essentially a wrapper around the same REST calls discussed.

- **Testing in UI first:** As suggested by the n8n team, when creating complex workflows via API, it can be helpful to first build them in the GUI to ensure everything (nodes, params, credentials) works, then fetch the JSON to use as a template. This can save a lot of trial and error in figuring out the exact JSON structure for certain nodes.

- **No direct execution endpoint:** Remember that to trigger workflows on demand you must use trigger nodes like Webhook. There is currently _no public API endpoint to arbitrarily start a workflow_ by ID. Plan your workflow designs accordingly (e.g., include a Webhook trigger if you need external on-demand starts).

- **Monitoring and logs:** Use the Executions endpoints to retrieve logs or build external monitoring. The execution data includes statuses and timestamps which you can analyze or store elsewhere. Note that by default, successful execution data might not be stored in full (only metadata) unless you change n8n’s settings.

By leveraging n8n’s REST API, you can fully manage your automation workflows from outside the n8n UI – creating new workflows, updating them, and monitoring their runs as part of a larger system. This enables use-cases like dynamically generating workflows (for example, based on user input or other events) and integrating n8n with your own app’s backend. For more details and advanced uses, refer to the official docs and the n8n community forums. Happy automating!

**Sources:**

- n8n Official Documentation – _REST API (Authentication & Usage)_
- n8n Community Forum – _Programmatically Creating Workflows (tips on nodes & connections)_
- n8n Community Forum – _Triggering Workflows via API (webhook workaround)_
- n8n Community Forum – _API Workflow Example (Generating workflow via API)_
- n8n Documentation – _Export/Import Workflows as JSON_
- n8n Template Library – _Community workflow templates (pre-built examples)_
