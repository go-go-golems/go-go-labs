# Document Reranker Application: Developer Onboarding Guide

Welcome to the Document Reranker project! This guide is designed to give you, as a new developer, a comprehensive understanding of our web application. We'll cover its architecture, core components, and how you can contribute to its development and extension. Our goal is to get you up to speed quickly so you can start making meaningful contributions.

## Table of Contents

1.  [Introduction: What is the Document Reranker?](#introduction-what-is-the-document-reranker)
2.  [Getting Started: Setting Up Your Environment](#getting-started-setting-up-your-environment)
3.  [Project Architecture: A Bird's Eye View](#project-architecture-a-birds-eye-view)
4.  [Deep Dive: Core Components](#deep-dive-core-components)
    *   [The Brain: Reranker Module (`reranker.py`)](#the-brain-reranker-module-rerankerpy)
    *   [The Memory: Database Module (`database.py`)](#the-memory-database-module-databasepy)
    *   [The Conductor: Flask Application (`app.py`)](#the-conductor-flask-application-apppy)
    *   [The Face: HTML Templates (`templates/`)](#the-face-html-templates-templates)
5.  [Workflow Walkthroughs](#workflow-walkthroughs)
    *   [Submitting Data via YAML](#submitting-data-via-yaml)
    *   [Using the Manual Query Form](#using-the-manual-query-form)
    *   [How Data is Stored and Retrieved](#how-data-is-stored-and-retrieved)
6.  [Code Quality and Debugging](#code-quality-and-debugging)
    *   [Logging Strategy](#logging-strategy)
    *   [Error Handling Approach](#error-handling-approach)
    *   [Debugging Tips and Tricks](#debugging-tips-and-tricks)
7.  [How to Contribute: Extending the Application](#how-to-contribute-extending-the-application)
    *   [Adding a New Page/Feature](#adding-a-new-pagefeature)
    *   [Modifying Core Logic (e.g., Reranking, Database)](#modifying-core-logic-eg-reranking-database)
    *   [Improving the User Interface](#improving-the-user-interface)
8.  [Troubleshooting Common Issues](#troubleshooting-common-issues)
9.  [Roadmap: Future Enhancements](#roadmap-future-enhancements)
10. [Final Words of Encouragement](#final-words-of-encouragement)

---

## 1. Introduction: What is the Document Reranker?

The Document Reranker is a powerful yet user-friendly web application built with **Flask** (a Python web framework) for the backend and **HTMX** for dynamic frontend interactions without complex JavaScript. At its heart, it utilizes a sophisticated transformer model, specifically **BAAI/bge-reranker-large**, to intelligently re-rank a list of documents based on their semantic relevance to a given query.

**Key Objectives & Features:**

*   **Flexible Data Input:** Users aren't locked into one way of providing data. They can:
    *   Upload structured **YAML files** containing both the query and a list of documents.
    *   Directly **paste YAML content** into a text area.
    *   Construct queries **manually** by typing a query and selecting from a pre-existing, centrally stored list of documents.
*   **Persistent Storage:** All documents, queries, and the resulting reranked lists (including scores) are saved in an **SQLite database**. This creates a valuable history, allowing users to:
    *   Revisit past query results.
    *   Reuse previously uploaded documents for new queries without re-uploading.
*   **User-Friendly Interface:** The application provides a clean interface to:
    *   View reranking results clearly.
    *   Browse historical queries.
    *   Manage and inspect the document database.
    *   Access example usage and a cheatsheet for guidance.

This tool is particularly useful in scenarios where you have a set of potentially relevant documents for a query (e.g., from an initial search or retrieval step) and need a more refined, semantically-aware ranking.

---

## 2. Getting Started: Setting Up Your Environment

To start developing on the Document Reranker, you'll need to set up your local environment.

**Prerequisites:**

*   **Python:** Ensure you have Python 3.8 or newer installed. You can check with `python --version`.
*   **pip:** Python's package installer, usually comes with Python.
*   **Git:** For version control (you've likely used this to get the code).
*   **(Optional but Recommended) Virtual Environment:** To keep project dependencies isolated.
    ```bash
    python -m venv .venv
    source .venv/bin/activate  # On Windows: .venv\Scripts\activate
    ```

**Installation Steps:**

1.  **Clone the Repository:** If you haven't already, get the code from the project's repository.
    ```bash
    # git clone <repository-url>
    cd <project-directory>/python/reranker
    ```

2.  **Install Dependencies:** All Python package requirements are listed in `requirements.txt`.
    ```bash
    pip install -r requirements.txt
    ```
    This will install Flask, PyYAML, PyTorch, Transformers, and other necessary libraries.

**Running the Application Locally:**

1.  **Navigate to the Application Directory:**
    ```bash
    cd python/reranker
    ```

2.  **Run the Flask Development Server:**
    ```bash
    python app.py
    ```
    You should see output indicating the server is running, typically on `http://127.0.0.1:5000/` or `http://localhost:5000/`.

3.  **Access in Browser:** Open your web browser and go to the address shown in the terminal.

**Using the VS Code Debugger (Recommended for Development):**

Our project includes a pre-configured `launch.json` file for easy debugging in Visual Studio Code.

*   **File Location:** `.vscode/launch.json`
*   **Configuration Name:** `"Python: Reranker Flask App"`

This configuration automatically sets the correct working directory (`python/reranker`), Python path, and environment variables for Flask development mode.

**To use it:**
1.  Open the project folder in VS Code.
2.  Go to the "Run and Debug" panel (usually a play button with a bug icon on the sidebar, or `Ctrl+Shift+D`).
3.  Select `"Python: Reranker Flask App"` from the dropdown menu at the top.
4.  Click the green play button (or press `F5`) to start debugging. You can now set breakpoints, inspect variables, and step through the code.

---

## 3. Project Architecture: A Bird's Eye View

The application follows a modular structure, primarily centered around a Flask backend.

```
python/reranker/
│
├── app.py                # CORE: Main Flask application logic, routing, request handling.
├── reranker.py           # CORE: Handles ML model loading and the reranking process.
├── database.py           # CORE: Manages SQLite database interactions (CRUD operations).
│
├── app.log               # OUTPUT: Application activity and error logs are written here.
├── reranker.db           # DATA: The SQLite database file.
│
├── requirements.txt      # CONFIG: Lists all Python package dependencies.
├── README.md             # DOCS: General project overview and user instructions.
│
├── templates/            # UI: Contains all Jinja2 HTML templates.
│   ├── base.html         #   - Foundation template for consistent layout, navigation, CSS/JS links.
│   ├── index.html        #   - Home page, offering different ways to input data.
│   ├── results.html      #   - Displays the reranked documents and scores.
│   ├── history.html      #   - Shows a list of past queries and their details.
│   ├── documents.html    #   - Allows browsing of all unique documents in the database.
│   ├── examples.html     #   - Provides pre-filled YAML examples for users.
│   ├── cheatsheet.html   #   - Offers guidance on YAML format and reranking concepts.
│   └── error.html        #   - A user-friendly page for displaying application errors.
│
├── static/               # UI: (Currently empty) Intended for static files like custom CSS, JS, or images.
│
├── examples/             # DATA: Sample YAML files to demonstrate application usage.
│   ├── rag_pipeline.yaml
│   ├── llm_architecture.yaml
│   └── coffee_brewing.yaml
│
└── docs/                 # DOCS: Contains detailed developer documentation.
    └── how-to-develop-on-the-reranker-application.md  # (This file) Your guide to understanding and contributing.
```

**Key Interactions:**

*   A user interacts with the HTML pages rendered by **Flask** (`app.py`) from the `templates/` directory.
*   When a reranking request is made, `app.py` calls functions in `reranker.py` to perform the ML inference.
*   Both `app.py` and `reranker.py` (indirectly, for storing documents if new) interact with `database.py` to save and retrieve data.
*   Logging is configured in `app.py` and used across modules to record events to `app.log` and the console.

---

## 4. Deep Dive: Core Components

Let's explore the main Python modules in more detail.

### The Brain: Reranker Module (`reranker.py`)

This module is the heart of the reranking functionality. It encapsulates all logic related to loading the machine learning model and using it to score documents against a query.

*   **Model Used:** `BAAI/bge-reranker-large` from Hugging Face. This is a powerful cross-encoder model specifically trained for reranking tasks.
*   **Lazy Loading:** The tokenizer and model are loaded only when first needed (`_get_tokenizer()` and `_get_model()`). This improves application startup time.
*   **Device Agnostic:** Automatically uses a CUDA-enabled GPU if `torch.cuda.is_available()` is true, otherwise defaults to CPU. This is managed by the `DEVICE` variable.

**Key Functions & Their Roles:**

*   `_get_tokenizer()`:
    *   Checks if the global `_tokenizer` variable is already populated.
    *   If not, it loads `AutoTokenizer.from_pretrained(MODEL_NAME)`.
    *   Logs the loading process and any errors.
    *   Returns the tokenizer instance.
*   `_get_model()`:
    *   Similar to `_get_tokenizer()`, checks the global `_model`.
    *   If not loaded, it instantiates `AutoModelForSequenceClassification.from_pretrained(MODEL_NAME)`.
    *   Moves the model to the determined `DEVICE` (GPU or CPU).
    *   Sets the model to evaluation mode (`model.eval()`) as we don't train it here.
    *   Logs the process and potential errors.
    *   Returns the model instance.
*   `rerank(query: str, documents: list[str], top_k: int | None = None) -> list[tuple[str, float]]`:
    *   This is the main public function of the module.
    *   Takes a `query` string, a `documents` list of strings, and an optional `top_k` integer.
    *   Handles empty document lists gracefully.
    *   Calls `_get_tokenizer()` and `_get_model()` to ensure they are ready.
    *   **Core Logic:**
        1.  Creates `pairs` of `(query, document)` for every document.
        2.  Uses `tokenizer.batch_encode_plus(...)` to prepare these pairs for the model. This includes padding, truncation, and converting to PyTorch tensors.
        3.  Moves the encoded batch to the `DEVICE`.
        4.  Within a `torch.no_grad()` context (to disable gradient calculations, saving memory and computation during inference):
            *   Passes the batch through the `model`.
            *   Extracts the relevance `scores` from the model's output logits.
        5.  Zips the original `documents` with their `scores`.
        6.  Sorts these pairs in descending order of score.
        7.  If `top_k` is provided, slices the list to return only the top K results.
    *   Returns a list of tuples, where each tuple is `(document_string, score_float)`.
    *   Includes comprehensive logging of its operations and any errors.

**What to Pay Attention To:**

*   **Model Performance:** The `MODEL_NAME` can be changed to experiment with other rerankers available on Hugging Face. Be mindful of model size and computational requirements.
*   **Input Length:** Transformer models have maximum input sequence lengths (often 512 tokens). The `truncation=True` in `batch_encode_plus` handles longer inputs by cutting them off. This could impact results if critical information is at the end of very long documents.
*   **Error Handling:** The module includes `try-except` blocks around model/tokenizer loading and the reranking process, logging errors to aid debugging.

### The Memory: Database Module (`database.py`)

This module is responsible for all persistent storage using SQLite. It handles the creation of the database and tables, and provides functions for adding, retrieving, and managing data related to documents, queries, and their reranked results.

*   **Database File:** `reranker.db` (created in the same directory as `database.py`).
*   **Row Factory:** Uses `sqlite3.Row` so that database rows can be accessed by column name (like dictionaries), which is more readable.

**Key Functions & Their Roles:**

*   `init_db()`:
    *   Called automatically when the module is imported.
    *   Connects to the database (or creates it if it doesn't exist).
    *   Executes `CREATE TABLE IF NOT EXISTS` statements for `documents`, `queries`, and `results` tables. This makes the function idempotent – safe to run multiple times.
    *   Logs the initialization process.
*   `get_db_connection()`:
    *   A context manager (used with `with ... as conn:`) that handles opening and closing the database connection.
    *   Ensures the connection is always closed, even if errors occur.
    *   Sets `conn.row_factory = sqlite3.Row`.
*   `get_or_create_document(content: str) -> int`:
    *   Crucial for avoiding duplicate document entries.
    *   First, queries the `documents` table to see if `content` already exists.
    *   If yes, returns the `id` of the existing document.
    *   If no, inserts the new `content` into the `documents` table and returns the `id` of the newly inserted row (`cursor.lastrowid`).
*   `save_query_and_results(query: str, results: list[tuple[str, float]], top_k: int | None = None) -> int`:
    *   Orchestrates saving a complete reranking operation.
    *   Inserts the `query` and `top_k` into the `queries` table, getting a `query_id`.
    *   Iterates through the `results` (list of document-score tuples):
        *   For each document in the results, calls `get_or_create_document()` to get its `doc_id`.
        *   Inserts a row into the `results` table linking the `query_id`, `doc_id`, `score`, and its `rank` (1-based index).
    *   Returns the `query_id` of the saved query.
*   `get_recent_queries(limit: int = 10) -> list[sqlite3.Row]`:
    *   Retrieves the most recent queries from the `queries` table, ordered by `created_at` descending.
    *   Used by the `/history` page.
*   `get_query_results(query_id: int) -> list[sqlite3.Row]`:
    *   Fetches the detailed results for a specific `query_id`.
    *   Joins `results`, `queries`, and `documents` tables to reconstruct the query, document content, score, and rank.
    *   Used when a user clicks to view a specific past query from the history.
*   `get_all_documents() -> list[sqlite3.Row]`:
    *   Retrieves all unique documents stored in the `documents` table.
    *   Used to populate the document selection list on the "Manual Query" tab of the home page and the `/documents` page.

**Database Schema in Detail:**

*   **`documents` Table:**
    *   `id INTEGER PRIMARY KEY AUTOINCREMENT`: Unique identifier for each document.
    *   `content TEXT UNIQUE NOT NULL`: The actual text of the document. The `UNIQUE` constraint is vital for preventing duplicates. `NOT NULL` ensures every document has content.
    *   `created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`: Timestamp of when the document was first added.
*   **`queries` Table:**
    *   `id INTEGER PRIMARY KEY AUTOINCREMENT`: Unique identifier for each query run.
    *   `query TEXT NOT NULL`: The text of the query.
    *   `top_k INTEGER`: Optional. If the user specified a K value, it's stored here. Can be `NULL`.
    *   `created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`: Timestamp of when the query was executed.
*   **`results` Table:**
    *   `id INTEGER PRIMARY KEY AUTOINCREMENT`: Unique identifier for each result entry.
    *   `query_id INTEGER NOT NULL`: Foreign key referencing `queries.id`. Links this result to a specific query.
    *   `document_id INTEGER NOT NULL`: Foreign key referencing `documents.id`. Links this result to a specific document.
    *   `score REAL NOT NULL`: The relevance score assigned by the reranker model.
    *   `rank INTEGER NOT NULL`: The 1-based rank of this document for the given query.
    *   `created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`: Timestamp for when this specific result entry was saved.

**What to Pay Attention To:**

*   **Schema Migrations:** If you need to change the table structure (e.g., add a column), SQLite doesn't have built-in, advanced migration tools like some larger database systems. For simple changes, you might add logic to `init_db()` to check for existing columns (using `PRAGMA table_info(table_name)`) and `ALTER TABLE` if needed. For complex changes, manual SQL scripts or a separate migration utility might be considered.
*   **Data Integrity:** The `UNIQUE` constraint on `documents.content` is key. Foreign keys help maintain relationships but rely on the application logic to insert data correctly.
*   **Query Performance:** For the current scale, SQLite is fine. If the database grows very large (millions of rows), you might need to add database indexes to specific columns (e.g., `queries.created_at`, `results.query_id`) to speed up lookups.

### The Conductor: Flask Application (`app.py`)

This is the central nervous system of the web application. It uses the Flask microframework to define URL routes, handle incoming HTTP requests, process data, interact with the other modules (`reranker.py`, `database.py`), and render HTML templates to the user.

*   **Logging Configuration:** `app.py` sets up the application-wide logging.
    *   It configures `logging.basicConfig` to set the default logging level (e.g., `DEBUG`), format for log messages, and handlers.
    *   Currently, it uses two handlers:
        *   `logging.FileHandler`: Writes logs to `app.log`.
        *   `logging.StreamHandler`: Prints logs to the console (standard output).
    *   A `logger = logging.getLogger(__name__)` instance is created for use within this module. Other modules also create their own loggers.
*   **Flask App Initialization:** `app = Flask(__name__)` creates the Flask application instance.
*   `app.config["MAX_CONTENT_LENGTH"]`: Limits the maximum size of uploaded files (e.g., YAML files) to prevent excessively large uploads.

**Key Routes & Their Logic (`@app.route(...)` decorators):**

*   `@app.route("/") def index()`:
    *   Handles requests to the home page.
    *   Calls `db.get_all_documents()` to fetch all stored documents.
    *   Passes these documents to `templates/index.html` so they can be displayed in the "Select Documents" list for the manual query form.
    *   Renders `index.html`.
*   `@app.route("/rerank", methods=["POST"]) def handle_rerank()`:
    *   This is a critical route that handles the actual reranking requests. It only accepts `POST` requests.
    *   **Form Differentiation:** It first checks hidden input fields (`yaml_form` or `manual_form`) to determine how the data was submitted.
    *   **If YAML form (`yaml_form`):**
        *   Reads YAML content either from an uploaded file (`request.files["file"]`) or a textarea (`request.form.get("yaml_text")`).
        *   Parses the YAML using `yaml.safe_load()`.
        *   Validates the parsed data structure (must be a dict with `query` and `documents` keys, and `documents` must be a list).
        *   Extracts `query`, `documents`, and optional `top_k`.
    *   **If Manual form (`manual_form`):**
        *   Gets the `query` from a text input.
        *   Gets a list of `selected_documents` (these are document IDs) from checkboxes.
        *   Gets an optional `top_k` value.
        *   Validates that a query and at least one document ID were provided.
        *   Retrieves the actual content of the selected documents from the database using their IDs.
    *   **Common Reranking Logic (after data extraction):**
        1.  Calls `reranker.rerank(query, documents, top_k)` to get the ranked results.
        2.  Calls `db.save_query_and_results(query, results, top_k)` to store the operation in the database. This returns a `query_id`.
        3.  Renders `templates/results.html`, passing the `query`, the reranked `results`, and the `query_id`. This `results.html` is often injected into a part of the `index.html` page by HTMX.
    *   Includes extensive logging and `try-except` blocks for error handling throughout the process. Returns JSON errors for HTMX to handle or renders an error page.
*   `@app.route("/history") def history()`:
    *   Calls `db.get_recent_queries()` (e.g., last 20).
    *   Renders `templates/history.html`, passing the list of queries.
*   `@app.route("/history/<int:query_id>") def view_query_results(query_id)`:
    *   Called when a user clicks on a specific query in the history.
    *   Calls `db.get_query_results(query_id)` to fetch the saved results for that query.
    *   If results are found, it formats them and renders `templates/results.html` again, but this time indicating it's `from_history` (which might change some UI elements, like showing a "Back to History" button).
*   `@app.route("/documents") def documents()`:
    *   Calls `db.get_all_documents()`.
    *   Renders `templates/documents.html`, passing the list of all stored documents.
*   `@app.route("/examples") def examples()`:
    *   Reads the content of YAML files from the `examples/` directory.
    *   Renders `templates/examples.html`, passing the list of example names and their content.
*   `@app.route("/cheatsheet") def cheatsheet()`:
    *   Simply renders `templates/cheatsheet.html`.
*   `@app.template_filter('format_datetime') def format_datetime(...)`:
    *   A custom Jinja2 template filter.
    *   Takes a datetime string (as stored in the DB) and formats it into a more human-readable string (e.g., "YYYY-MM-DD HH:MM:SS").
    *   Used in templates like `history.html`.
*   `@app.errorhandler(404) def page_not_found(e)` and `@app.errorhandler(500) def server_error(e)`:
    *   Custom error handlers for HTTP 404 (Not Found) and 500 (Internal Server Error) responses.
    *   They log the error and render `templates/error.html` with an appropriate message.

**What to Pay Attention To:**

*   **Request Object:** Flask's `request` object is used extensively to access form data (`request.form`), uploaded files (`request.files`), query parameters, etc.
*   **HTMX Integration:** While not directly visible in `app.py` Python code, many routes are designed to be called by HTMX. For instance, the `/rerank` route, when successful, returns an HTML fragment (`results.html`) which HTMX then injects into the main page, rather than causing a full page reload. This makes the UI feel more responsive.
*   **Error Handling and Return Types:** Notice how some parts of `/rerank` return `jsonify({"error": ...}), 400` for client-side errors (expected by HTMX for form validation issues or bad input), while more severe server-side issues or navigation errors might lead to rendering `error.html` or redirecting.
*   **Security:** For this internal tool, security considerations like input sanitization against XSS (Cross-Site Scripting) or CSRF (Cross-Site Request Forgery) protection are basic. If this were to be exposed externally, more robust security measures would be needed (e.g., using Flask-WTF for forms, escaping all user-generated content in templates if not already handled by Jinja2's autoescaping).

### The Face: HTML Templates (`templates/`)

This directory contains all the HTML files that define the user interface of the application. They are written using Jinja2, Flask's default templating engine, which allows embedding Python-like expressions and control structures within HTML. Bootstrap 5 is used for styling, and HTMX attributes are embedded for dynamic client-side interactions.

*   **Jinja2 Templating:**
    *   `{% ... %}` for statements (e.g., `{% extends ... %}`, `{% for ... %}`, `{% if ... %}`).
    *   `{{ ... }}` for expressions to print to the template output (e.g., `{{ query }}`, `{{ doc.content }}`).
    *   Template inheritance (`{% extends "base.html" %}`, `{% block content %}`) is used to maintain a consistent layout.
*   **Bootstrap 5:** Provides pre-built CSS classes for styling components like cards, buttons, forms, navigation, and layout grids, making the UI look clean and professional with minimal custom CSS. CDN links are in `base.html`.
*   **HTMX:** Used to enhance HTML by allowing elements to make AJAX requests and update parts of the page without full reloads.
    *   Attributes like `hx-post`, `hx-get`, `hx-target`, `hx-swap`, `hx-indicator` are used.
    *   Example: In `index.html`, the forms use `hx-post="/rerank" hx-target="#results"` which means on submit, they POST to `/rerank` and the HTML response from that route will replace the content of the element with `id="results"`.

**Key Templates & Their Purpose:**

*   `base.html`:
    *   The master template. All other main pages `{% extends "base.html" %}`.
    *   Contains the overall HTML structure (`<head>`, `<body>`), CDN links for Bootstrap CSS & JS, and the HTMX library.
    *   Defines common elements like the header (application title) and the navigation bar.
    *   Includes named `{% block ... %}` sections (e.g., `home_active`, `content`) that child templates can override.
    *   Includes some basic custom CSS for consistent styling.
*   `index.html`:
    *   The main landing page.
    *   Features a tabbed interface for three ways to input data:
        1.  **Upload YAML:** A file input field.
        2.  **Paste YAML:** A textarea for direct YAML input.
        3.  **Manual Query:** A text input for the query, a number input for `top_k`, and a multi-select list of documents (populated from the database, passed from `app.py`). Includes "Select All" / "Deselect All" JavaScript enhancements.
    *   All forms on this page use HTMX to submit to the `/rerank` endpoint and display results in the `#results` div on the same page.
    *   Shows a loading indicator (`#loading`) during HTMX requests.
*   `results.html`:
    *   Not a full page on its own but an HTML *fragment* rendered by the `/rerank` and `/history/<query_id>` routes.
    *   Displays the query and a table of reranked documents with their scores and ranks.
    *   Includes a "View in History" link if the results are fresh (not already being viewed from history).
    *   Includes a "Back to History" link if being viewed from history.
*   `history.html`:
    *   Displays a list of past queries in card format.
    *   Each card shows the query text, timestamp (formatted using the custom `format_datetime` filter), and optional `top_k`.
    *   Provides a "View Results" button for each query, linking to `/history/<query_id>`.
*   `documents.html`:
    *   Lists all unique documents stored in the database.
    *   Includes a search bar (client-side JavaScript filtering) to quickly find documents.
    *   Shows document ID and content (with a copy button for the content).
*   `examples.html`:
    *   Presents example YAML files in an accordion structure.
    *   Each example shows its content in a `<pre>` tag.
    *   Provides a "Copy to Clipboard" button and a "Use This Example" button. The latter uses JavaScript and `localStorage` to temporarily store the example content and then, upon redirecting to the home page, pre-fills the "Paste YAML" textarea and switches to that tab.
*   `cheatsheet.html`:
    *   A static informational page explaining the expected YAML format, details about the reranking model, and general tips for getting good results.
*   `error.html`:
    *   A user-friendly page displayed when the application encounters an unrecoverable error (HTTP 500) or a page is not found (HTTP 404).
    *   Shows the error message passed from `app.py` and suggests common troubleshooting steps.

**What to Pay Attention To:**

*   **HTMX Workflow:** Understand how forms in `index.html` trigger POST requests to `/rerank` and how the `results.html` fragment updates the `#results` div. This is key to the application's interactive feel.
*   **Jinja2 Logic:** Pay attention to loops (`{% for %}`), conditionals (`{% if %}`), and template inheritance.
*   **Bootstrap Classes:** Familiarize yourself with common Bootstrap classes for layout (like `container`, `row`, `col-`), components (like `card`, `btn`, `nav`, `alert`), and utilities.
*   **Client-Side JavaScript:** Some templates (`index.html`, `examples.html`, `documents.html`) have small, embedded `<script>` tags for minor UI enhancements (e.g., copy-to-clipboard, select-all, client-side search). These are generally self-contained.

---

## 5. Workflow Walkthroughs

Understanding how data flows through the application is crucial.

### Submitting Data via YAML (Upload or Paste)

1.  **User Action:**
    *   **Upload:** The user selects the "Upload YAML" tab on `index.html`, chooses a `.yaml` or `.yml` file, and clicks "Rerank Documents."
    *   **Paste:** The user selects the "Paste YAML" tab, pastes YAML content into the textarea, and clicks "Rerank Documents."
2.  **Client-Side (HTMX):**
    *   The form submission is intercepted by HTMX.
    *   An AJAX `POST` request is made to the `/rerank` endpoint defined in `app.py`.
    *   A hidden input field `yaml_form=1` is included to tell the backend which type of form was submitted.
    *   The `#loading` indicator is displayed.
3.  **Server-Side (`app.py` - `handle_rerank()` function):**
    *   The `/rerank` route receives the request.
    *   It detects `yaml_form`.
    *   **File Upload:** If `request.files["file"]` exists, its content is read and decoded.
    *   **Pasted Text:** Otherwise, `request.form.get("yaml_text")` is used.
    *   **Validation:** The YAML content is parsed using `yaml.safe_load()`. The structure is validated (must be a dictionary containing `query` and `documents` list). If invalid, a JSON error response is returned (e.g., HTTP 400).
    *   **Data Extraction:** `query`, `documents` list, and optional `top_k` are extracted from the parsed YAML.
4.  **Reranking (`reranker.py`):**
    *   `app.py` calls `reranker.rerank(query, documents_list, top_k)`.
    *   `reranker.py` loads the model/tokenizer (if not already loaded), processes the query-document pairs, and returns a sorted list of `(document, score)` tuples.
5.  **Database Storage (`database.py`):**
    *   `app.py` calls `db.save_query_and_results(query, ranked_results, top_k)`.
    *   `database.py` performs the following:
        *   Inserts the query into the `queries` table.
        *   For each document in the `ranked_results`:
            *   Calls `db.get_or_create_document()` to add the document to the `documents` table if it's new, or get its existing ID.
            *   Inserts an entry into the `results` table, linking the query, document, score, and rank.
    *   The new `query_id` is returned to `app.py`.
6.  **Response Generation (`app.py`):**
    *   `app.py` renders the `templates/results.html` fragment, passing the original `query`, the `ranked_results`, and the `query_id`.
7.  **Client-Side (HTMX Update):**
    *   The HTML fragment received from the server replaces the content of the `<div id="results">` element on the `index.html` page.
    *   The `#loading` indicator is hidden.

### Using the Manual Query Form

1.  **User Action:**
    *   The user selects the "Manual Query" tab on `index.html`.
    *   Types a query into the "Query" input field.
    *   Optionally enters a number for "Top K".
    *   Selects one or more documents from the checklist (which is populated from the `documents` table in the database).
    *   Clicks "Rerank Documents."
2.  **Client-Side (HTMX):**
    *   Similar to the YAML workflow, HTMX intercepts the form submission.
    *   An AJAX `POST` request is made to `/rerank`.
    *   A hidden input field `manual_form=1` is included.
    *   The `#loading` indicator is displayed.
3.  **Server-Side (`app.py` - `handle_rerank()` function):**
    *   The `/rerank` route receives the request and detects `manual_form`.
    *   **Data Extraction:**
        *   `query` = `request.form.get("query")`.
        *   `document_ids` = `request.form.getlist("selected_documents")` (this gets a list of document IDs from the checkboxes).
        *   `top_k` = `request.form.get("top_k")`.
    *   **Validation:** Checks if the query is empty or if no documents were selected. If so, returns a JSON error.
    *   **Document Retrieval (`database.py`):**
        *   `app.py` connects to the database and executes a SQL query like `SELECT content FROM documents WHERE id IN (?, ?, ...)` using the `document_ids`.
        *   The content of the selected documents is fetched into a list.
4.  **Reranking (`reranker.py`):**
    *   `app.py` calls `reranker.rerank(query, fetched_document_contents, top_k)`.
    *   (Same reranking process as above).
5.  **Database Storage (`database.py`):**
    *   `app.py` calls `db.save_query_and_results(query, ranked_results, top_k)`.
    *   (Same storage process as above; the documents, if already in the DB from selection, will just have their IDs retrieved by `get_or_create_document`).
6.  **Response Generation (`app.py`):**
    *   `app.py` renders `templates/results.html` with the query, results, and query_id.
7.  **Client-Side (HTMX Update):**
    *   (Same as YAML workflow: `#results` div is updated).

### How Data is Stored and Retrieved

*   **Documents are Unique:** The `documents` table stores each unique piece of document content only once, thanks to the `UNIQUE` constraint on the `content` column and the `get_or_create_document()` logic. This saves space and ensures consistency.
*   **Queries are Logged:** Every query attempt (that passes initial validation) is stored in the `queries` table along with its timestamp and any `top_k` parameter.
*   **Results Link Queries and Documents:** The `results` table acts as a many-to-many link between queries and documents, additionally storing the specific `score` and `rank` for that document in the context of that query.
*   **History Retrieval:** The `/history` page queries the `queries` table for recent entries. Clicking "View Results" on an item fetches data from the `results` table (joined with `queries` and `documents`) for that specific `query_id`.
*   **Document Browsing:** The `/documents` page directly queries the `documents` table to list all stored content. The "Manual Query" tab also uses this data to populate its selection list.

---

## 6. Code Quality and Debugging

Maintaining a healthy codebase requires good logging, robust error handling, and effective debugging practices.

### Logging Strategy

We employ Python's standard `logging` module, configured in `app.py` at application startup.

*   **Configuration (`app.py`):**
    ```python
    logging.basicConfig(
        level=logging.DEBUG,  # Captures DEBUG, INFO, WARNING, ERROR, CRITICAL
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s', # Log message format
        handlers=[
            logging.FileHandler(os.path.join(os.path.dirname(__file__), 'app.log')), # Writes to app.log
            logging.StreamHandler()  # Writes to console
        ]
    )
    # Each module gets its own logger:
    logger = logging.getLogger(__name__)
    ```
*   **Usage:**
    *   `logger.debug("Detailed message for debugging")`
    *   `logger.info("General operational information")`
    *   `logger.warning("Something unexpected, or a potential issue")`
    *   `logger.error("An error occurred, exception caught")`
    *   `logger.critical("A very serious error, app might not recover")`
    *   When catching exceptions, use `logger.error(traceback.format_exc())` to log the full stack trace.
*   **Log File:** All logs are aggregated in `python/reranker/app.log`. This file is crucial for diagnosing issues after they occur, especially in a deployed environment.
*   **Purpose:**
    *   **Troubleshooting:** Logs provide a trail of events leading up to an error.
    *   **Monitoring:** Can be used to observe application activity and performance.
    *   **Auditing:** In some cases, can provide a record of operations.

### Error Handling Approach

The application aims to handle errors gracefully at different levels.

*   **Explicit `try...except` Blocks:**
    *   Most routes in `app.py` and critical functions in `database.py` and `reranker.py` are wrapped in `try...except Exception as e:` blocks.
    *   When an exception is caught:
        1.  The error is logged (often with `logger.error(str(e))` and `logger.error(traceback.format_exc())`).
        2.  A user-friendly response is generated.
*   **Flask Error Handlers (`@app.errorhandler`):**
    *   `app.py` defines handlers for common HTTP errors:
        *   `@app.errorhandler(404)`: For "Page Not Found" errors. Renders `templates/error.html`.
        *   `@app.errorhandler(500)`: For "Internal Server Error" (uncaught exceptions in request handling). Renders `templates/error.html`.
*   **User-Facing Error Page (`templates/error.html`):**
    *   Provides a generic error message to the user, avoiding exposure of technical details.
    *   Suggests checking application logs for more information.
*   **JSON Errors for HTMX:**
    *   In routes called by HTMX (like `/rerank`), if there's a validation error or a recoverable issue, the server often returns a JSON response with an error message and an appropriate HTTP status code (e.g., 400 for Bad Request).
    ```python
    return jsonify({"error": "Invalid YAML format. Must contain 'query' and 'documents' fields"}), 400
    ```
    HTMX can be configured to handle these JSON error responses on the client-side (though currently, we mainly rely on Flask rendering an error snippet or the user seeing a console error if HTMX fails).

### Debugging Tips and Tricks

1.  **Leverage the VS Code Debugger:** This is the most powerful tool.
    *   Set breakpoints in `app.py`, `reranker.py`, or `database.py`.
    *   Step through code execution (F10 for step over, F11 for step into).
    *   Inspect variable values at runtime.
    *   Examine the call stack.
2.  **Examine `app.log`:** This log file is your best friend for understanding what happened, especially for errors that are hard to reproduce or occur intermittently. Look for `ERROR` or `CRITICAL` messages and the accompanying stack traces.
3.  **Use `print()` Judiciously:** For quick, temporary checks of variable values or code flow, `print()` statements can be helpful, especially if the debugger feels too slow or cumbersome for a small check. Remember to remove them afterwards!
4.  **Flask Debug Mode:** The application runs with `debug=True` by default in `app.py` (`app.run(debug=True, ...)`). This provides:
    *   **Interactive Debugger in Browser:** If an unhandled exception occurs during a request, Flask can show an interactive traceback in the browser (Werkzeug debugger). *Be very careful with this in any environment that is not strictly local development, as it can expose sensitive information.*
    *   **Automatic Reloader:** The server will automatically restart when it detects code changes, speeding up the development cycle.
5.  **Browser Developer Tools (for HTMX/Frontend):**
    *   **Network Tab:** Inspect the AJAX requests made by HTMX. Check the request headers, payload, and the server's response (including HTML fragments or JSON errors).
    *   **Console Tab:** Look for JavaScript errors or messages logged by HTMX.
6.  **Isolate the Problem:**
    *   If an error occurs in `/rerank`, try to determine if it's in YAML parsing, manual form data extraction, document retrieval from DB, the reranking model itself, or saving results to DB. Add specific logging or breakpoints in each section.
    *   Test components individually if possible. For example, you could write a small standalone script to test functions in `reranker.py` or `database.py` with sample data.
7.  **Understand Expected Data Formats:** Many errors stem from data not being in the expected format (e.g., YAML structure, list vs. dict, string vs. int). Log or debug data structures at critical points.

---

## 7. How to Contribute: Extending the Application

We encourage contributions! Here's how you can extend the application.

### Adding a New Page/Feature

Let's say you want to add a new page, for example, an "Admin Dashboard" at `/admin`.

1.  **Define the Route in `app.py`:**
    ```python
    @app.route("/admin")
    def admin_dashboard():
        # Add logic here: fetch data, perform actions, etc.
        # For example, get some stats from the database
        try:
            num_queries = db.get_total_queries_count() # Assume you create this DB function
            num_documents = db.get_total_documents_count() # Assume you create this
            logger.info("Admin dashboard loaded.")
            return render_template("admin_dashboard.html", 
                                   query_count=num_queries, 
                                   doc_count=num_documents)
        except Exception as e:
            logger.error(f"Error loading admin dashboard: {str(e)}\n{traceback.format_exc()}")
            return render_template("error.html", error="Could not load admin dashboard.")
    ```

2.  **Create the HTML Template (`templates/admin_dashboard.html`):**
    ```html
    {% extends "base.html" %}

    {% block admin_active %}active{% endblock %} {# For highlighting in nav, if you add it #}

    {% block content %}
    <div class="card">
        <div class="card-header">
            <h5 class="card-title mb-0">Admin Dashboard</h5>
        </div>
        <div class="card-body">
            <p>Welcome to the Admin Dashboard!</p>
            <p>Total Queries Processed: {{ query_count }}</p>
            <p>Total Unique Documents: {{ doc_count }}</p>
            {# Add more admin-specific content here #}
        </div>
    </div>
    {% endblock %}
    ```

3.  **Add a Link in `templates/base.html` (Navigation Bar):**
    Find the `<nav>` section and add a new link:
    ```html
    <a class="nav-link {% block admin_active %}{% endblock %}" href="/admin">Admin</a>
    ```
    Remember to define the corresponding `{% block admin_active %}` in `admin_dashboard.html` if you want the link to appear "active" when on that page.

4.  **Add any necessary backend logic:**
    *   If your new page needs to interact with the database, add new functions to `database.py` (like `get_total_queries_count()` in the example).
    *   If it involves new processing, consider if it fits in an existing module or needs a new one.

### Modifying Core Logic (e.g., Reranking, Database)

*   **Changing the Reranker Model:**
    1.  In `reranker.py`, update `MODEL_NAME = "NewModel/name-here"`.
    2.  Test thoroughly. Different models might have different input/output expectations or performance characteristics.
    3.  Consider making `MODEL_NAME` configurable, perhaps via an environment variable or a setting in the UI if users should be able to switch.
*   **Altering Database Schema (e.g., Adding a "tags" column to `documents`):**
    1.  **Update `database.py` `init_db()`:**
        Modify the `CREATE TABLE documents` statement:
        ```sql
        CREATE TABLE IF NOT EXISTS documents (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            content TEXT UNIQUE NOT NULL,
            tags TEXT, -- New column for comma-separated tags, for example
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
        ```
    2.  **Handle Migration (Important!):** For existing databases, the new column won't exist. You need to add logic to apply this change. This can be done within `init_db()` or a separate migration script.
        ```python
        # In init_db(), after table creation attempts:
        try:
            cursor.execute("PRAGMA table_info(documents)")
            columns = [col[1] for col in cursor.fetchall()]
            if "tags" not in columns:
                logger.info("Migrating documents table: adding 'tags' column.")
                cursor.execute("ALTER TABLE documents ADD COLUMN tags TEXT")
                conn.commit()
        except Exception as e:
            logger.error(f"Error migrating documents table: {e}")
        ```
    3.  **Update Data Handling Functions:**
        *   Modify `get_or_create_document()` if tags need to be saved.
        *   Update `get_all_documents()` if tags need to be retrieved.
        *   Update any part of `app.py` or templates that would use or display these tags.
*   **Modifying Reranking Logic:**
    *   If you want to change how scores are calculated or how documents are preprocessed before reranking, the changes would primarily be in `reranker.py` within the `rerank()` function.
    *   Always log changes and test extensively.

### Improving the User Interface

*   **Styling Changes:**
    *   Primarily involve editing CSS. You can add custom styles to the `<style>` block in `templates/base.html` or create a new CSS file in `static/` and link it in `base.html`.
    *   Leverage Bootstrap 5 utility classes as much as possible for consistency.
*   **Adding Client-Side Interactivity (with JavaScript/HTMX):**
    *   **Simple JS:** For small enhancements, you can embed `<script>` tags directly in templates (as seen in `examples.html` for copy-pasting or `documents.html` for client-side search).
    *   **HTMX:** For more significant dynamic updates without full page reloads (like form submissions, partial page updates), continue using HTMX attributes on HTML elements. This often involves ensuring your Flask endpoints return HTML fragments rather than full pages when called by HTMX.
    *   **Example (Client-side filtering for a list):**
        Refer to the search functionality in `documents.html`. It uses an input field and JavaScript to show/hide list items based on the search term.

---

## 8. Troubleshooting Common Issues

Here are some common problems you might encounter and how to address them:

*   **Model Loading Issues (often on first run or if dependencies change):**
    *   **Symptom:** Slow startup, errors in logs related to Hugging Face, timeouts.
    *   **Causes:**
        *   No internet connection (models are downloaded on first use).
        *   Insufficient disk space (models can be large).
        *   Proxy/firewall issues blocking downloads from Hugging Face.
        *   Corrupted model cache.
    *   **Solutions:**
        *   Ensure stable internet and enough disk space.
        *   Try clearing the Hugging Face cache (usually in `~/.cache/huggingface/`).
        *   Manually download the model using `transformers-cli download BAAI/bge-reranker-large` to test connectivity and see more specific download errors.
        *   For development, consider using a smaller, faster-loading reranker model if the primary one is too slow for quick iterations.
*   **Database Errors (`sqlite3.OperationalError`, etc.):**
    *   **Symptom:** Errors like "table not found," "database is locked," "unable to open database file."
    *   **Causes:**
        *   `reranker.db` file permissions (app might not have write access).
        *   Path issues if `DB_PATH` in `database.py` is incorrect or the app is run from an unexpected directory.
        *   Concurrent access issues (less common with SQLite in this single-process app, but possible if multiple instances run against the same file without care).
        *   Schema mismatches if the code expects columns/tables that don't exist in an older DB file.
    *   **Solutions:**
        *   Verify file permissions for `reranker.db` and its directory.
        *   Ensure `DB_PATH` is correctly pointing to the desired location.
        *   Check logs for the specific SQL error.
        *   If you suspect schema issues with an old DB, you can try deleting `reranker.db` (it will be recreated by `init_db()`, but **you will lose all data**). Do this only as a last resort or in a dev environment.
*   **Memory Issues (RAM Exhaustion):**
    *   **Symptom:** Application becomes very slow, crashes, or the OS kills the process.
    *   **Causes:**
        *   Loading a very large model into memory (especially on systems with limited RAM, or if also running on CPU).
        *   Processing a very large number of documents or very long documents simultaneously in `reranker.py`.
        *   Inefficient data handling in Python (e.g., holding too many large objects in memory).
    *   **Solutions:**
        *   If on a low-RAM system, ensure you're not running other memory-heavy applications.
        *   Consider batching the documents sent to `reranker.rerank()` if you are dealing with thousands at once. The model itself processes in batches, but Python overhead for managing huge lists can be an issue.
        *   Profile memory usage if this becomes a persistent problem.
*   **HTMX Not Working as Expected:**
    *   **Symptom:** Full page reloads instead of partial updates, or nothing happens on click.
    *   **Causes:**
        *   HTMX JavaScript library not loaded (check `base.html` and browser network tools).
        *   Incorrect `hx-*` attributes (typos in attribute names, wrong target IDs).
        *   Flask endpoint not returning an HTML fragment (might be returning a full page, or JSON when HTML is expected, or an error).
        *   JavaScript errors on the page conflicting with HTMX.
    *   **Solutions:**
        *   Use browser developer tools:
            *   **Network Tab:** Check the request made by HTMX. Is it going to the right URL? What was the server's response code and content?
            *   **Console Tab:** Look for JavaScript errors.
        *   Verify your `hx-target` IDs match existing element IDs in the DOM.
        *   Ensure the Flask route called by HTMX is rendering a template fragment suitable for swapping.

---

## 9. Roadmap: Future Enhancements

This application has a solid foundation, but there are many exciting ways it could be improved:

*   **Performance & Scalability:**
    *   **Caching:** Implement caching for reranker model predictions, especially if the same query-document pairs are seen often. Flask-Caching could be used.
    *   **Background Tasks/Workers:** For very large reranking jobs, offload them to a background worker queue (e.g., Celery with Redis/RabbitMQ) to prevent blocking web requests and provide a better user experience.
    *   **Database Optimization:** For very large datasets, analyze query performance and add appropriate database indexes to `queries`, `documents`, and `results` tables (e.g., on foreign keys, frequently filtered columns).
*   **Advanced Reranking Features:**
    *   **Document Chunking:** Integrate a document chunking mechanism. Instead of reranking whole documents, chunk them into smaller, semantically coherent pieces, rerank the chunks, and then present results based on the best chunks.
    *   **Multiple Model Support:** Allow users to select from a list of different reranker models via the UI.
    *   **Hybrid Search/Reranking:** Combine keyword-based search scores with semantic reranker scores.
*   **User Interface & Experience (UI/UX):**
    *   **Dark Mode:** A popular feature for modern web apps.
    *   **Enhanced Mobile Responsiveness:** Further refine styles for smaller screens.
    *   **Document Preview:** Allow users to see a more detailed preview of a document directly from the results or document list.
    *   **Result Comparison:** UI to compare results from different queries or models side-by-side.
    *   **Pagination for History/Documents:** If these lists become very long, implement server-side pagination for better performance.
*   **Functionality & Integration:**
    *   **User Authentication & Accounts:** Allow users to have private histories and document sets. Flask-Login is a good starting point.
    *   **API Endpoints:** Expose the reranking functionality via a RESTful API for programmatic access by other services.
    *   **Vector Database Integration:** Store document embeddings (from a bi-encoder model) in a vector database (e.g., FAISS, Weaviate, Pinecone) for an initial semantic search step before reranking.
    *   **Direct Document Source Integration:** Allow connecting to external document sources (e.g., S3, shared drives, Confluence) instead of just manual uploads/DB.
    *   **Export Results:** Allow users to download reranked results in formats like CSV or JSON.
*   **Operational Improvements:**
    *   **Configuration Management:** Move settings like `MODEL_NAME` or `DB_PATH` to environment variables or a configuration file for easier deployment.
    *   **More Sophisticated Migrations:** For database schema changes, implement a more robust migration system (e.g., using Alembic if SQLAlchemy were introduced, or a simpler custom script-based approach).
    *   **Unit & Integration Tests:** Develop a comprehensive test suite to ensure code quality and prevent regressions.

---

## 10. Final Words of Encouragement

You've joined a project with a lot of potential! The Document Reranker is already a useful tool, and with your contributions, it can become even better.

**As you work with the codebase, remember these key principles:**

*   **Read the Logs:** They are your first line of defense when troubleshooting.
*   **Understand the Flow:** Trace how data moves from the UI, through `app.py`, to `reranker.py` and `database.py`, and back to the UI.
*   **Test Your Changes:** Whether it's a small UI tweak or a big backend modification, test it thoroughly. Consider writing automated tests for new backend logic.
*   **Write Clear Code:** Aim for readability and maintainability. Add comments where the logic isn't immediately obvious.
*   **Document Your Work:** If you add a significant new feature or change existing behavior in a non-trivial way, update this guide or other relevant documentation.
*   **Ask Questions:** If you're stuck, unsure about the best approach, or don't understand a part of the code, don't hesitate to ask other team members. Collaboration is key!

We're excited to have you on board. Happy coding, and let's make the Document Reranker even more awesome! 