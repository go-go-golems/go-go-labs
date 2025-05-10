#!/usr/bin/env python3
import os
import yaml
import logging
import traceback
from typing import List, Dict, Any, Optional, Tuple, Union
from flask import Flask, render_template, request, jsonify, redirect, url_for, Response
from datetime import datetime

import database as db
from reranker import rerank

# Configure logging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(os.path.join(os.path.dirname(__file__), 'app.log')),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
app.config["MAX_CONTENT_LENGTH"] = 16 * 1024 * 1024  # 16MB max upload size

def foobar(s: str) -> str:
    return "23"

@app.route("/")
def index() -> Union[str, Tuple[str, int]]:
    # Get all documents from the database for the document selector
    try:
        documents_data: List[Dict[str, Any]] = db.get_all_documents()
        logger.info(f"Index page loaded with {len(documents_data)} documents")
        return render_template("index.html", documents=documents_data)
    except Exception as e:
        logger.error(f"Error loading index page: {str(e)}")
        logger.error(traceback.format_exc())
        return render_template("error.html", error=str(e))

@app.route("/rerank", methods=["POST"])
def handle_rerank() -> Union[Response, str, Tuple[Response, int]]:
    logger.info("Rerank request received")
    query: Optional[str] = None
    documents_content: Optional[List[str]] = None
    top_k_value: Optional[int] = None
    try:
        # Check which form was submitted
        if "yaml_form" in request.form:
            logger.info("Processing YAML form submission")
            # Process YAML input (file or text)
            yaml_content: str = ""
            if "file" in request.files and request.files["file"].filename:
                yaml_content = request.files["file"].read().decode("utf-8")
                logger.info(f"YAML file uploaded: {request.files['file'].filename}")
            else:
                yaml_content = request.form.get("yaml_text", "")
                logger.info("YAML text pasted")
            
            if not yaml_content.strip():
                logger.warning("Empty YAML content provided")
                return jsonify({"error": "No YAML content provided"}), 400
                
            try:
                data: Dict[str, Any] = yaml.safe_load(yaml_content)
                logger.debug(f"Parsed YAML: {data}")
                
                if not isinstance(data, dict) or "query" not in data or "documents" not in data:
                    logger.warning("Invalid YAML format (missing required fields)")
                    return jsonify({"error": "Invalid YAML format. Must contain 'query' and 'documents' fields"}), 400
                
                query = data["query"]
                documents_content = data["documents"]
                top_k_value = data.get("top_k", None)
                logger.debug(f"YAML form - top_k raw value: {top_k_value!r}, type: {type(top_k_value)}")
                
                # Convert top_k to int if provided
                if top_k_value is not None:
                    try:
                        top_k_value = int(top_k_value)
                        logger.debug(f"YAML form - top_k converted value: {top_k_value!r}, type: {type(top_k_value)}")
                    except (ValueError, TypeError):
                        logger.warning(f"Invalid top_k value in YAML: {top_k_value!r}, type: {type(top_k_value)}")
                        return jsonify({"error": "top_k must be a valid integer"}), 400
                
                if not isinstance(documents_content, list):
                    logger.warning("Documents is not a list")
                    return jsonify({"error": "Documents must be a list"}), 400
            except yaml.YAMLError as e:
                logger.error(f"YAML parsing error: {str(e)}")
                return jsonify({"error": "Invalid YAML format"}), 400
        
        elif "manual_form" in request.form:
            logger.info("Processing manual form submission")
            # Process manual query input
            query = request.form.get("query", "").strip()
            document_ids: List[str] = request.form.getlist("selected_documents")
            raw_top_k: Optional[str] = request.form.get("top_k")
            logger.debug(f"Manual form - top_k raw value: {raw_top_k!r}, type: {type(raw_top_k)}")
            
            logger.debug(f"Manual query: '{query}', document_ids: {document_ids}, raw_top_k: {raw_top_k!r}")
            
            if not query:
                logger.warning("Empty query provided")
                return jsonify({"error": "Query cannot be empty"}), 400
                
            if not document_ids:
                logger.warning("No documents selected")
                return jsonify({"error": "No documents selected"}), 400
                
            # Convert top_k to int if provided
            if raw_top_k:
                try:
                    top_k_value = int(raw_top_k)
                    logger.debug(f"Manual form - top_k converted value: {top_k_value!r}, type: {type(top_k_value)}")
                except ValueError:
                    logger.warning(f"Invalid top_k value: {raw_top_k!r}, type: {type(raw_top_k)}")
                    return jsonify({"error": "top_k must be a valid integer"}), 400
                    
            # Get document content from database
            documents_from_db: List[str] = []
            try:
                with db.get_db_connection() as conn:
                    cursor = conn.cursor()
                    placeholders: str = ",".join("?" for _ in document_ids)
                    query_sql: str = f"SELECT content FROM documents WHERE id IN ({placeholders})"
                    logger.debug(f"SQL query: {query_sql} with params {document_ids}")
                    cursor.execute(query_sql, document_ids)
                    documents_from_db = [row["content"] for row in cursor.fetchall()]
                    logger.info(f"Retrieved {len(documents_from_db)} documents from database")
            except Exception as e:
                logger.error(f"Database error retrieving documents: {str(e)}")
                logger.error(traceback.format_exc())
                return jsonify({"error": f"Database error: {str(e)}"}), 500
            documents_content = documents_from_db
        
        else:
            logger.warning("Invalid form submission (missing form identifier)")
            return jsonify({"error": "Invalid form submission"}), 400
        
        if query is None or documents_content is None:
             logger.error("Query or documents not processed correctly")
             return jsonify({"error": "Internal server error: query or documents missing."}), 500

        # Perform reranking
        logger.info(f"Performing reranking for query: '{query}' with {len(documents_content)} documents and top_k={top_k_value!r} (type: {type(top_k_value)})")
        try:
            results: List[Tuple[str, float]] = rerank(query, documents_content, top_k_value)
            logger.info(f"Reranking complete, got {len(results)} results")
            logger.debug(f"First few results: {results[:3] if results else []}")
        except Exception as e:
            logger.error(f"Reranking error: {str(e)}")
            logger.error(traceback.format_exc())
            return jsonify({"error": f"Reranking error: {str(e)}"}), 500
        
        # Save query and results to database
        try:
            query_id: int = db.save_query_and_results(query, results, top_k_value)
            logger.info(f"Saved query (ID: {query_id}) and results to database")
        except Exception as e:
            logger.error(f"Database error saving results: {str(e)}")
            logger.error(traceback.format_exc())
            return jsonify({"error": f"Database error: {str(e)}"}), 500
        
        return render_template("results.html", 
                              query=query, 
                              results=results,
                              query_id=query_id)
    
    except Exception as e:
        logger.error(f"Unexpected error in handle_rerank: {str(e)}")
        logger.error(traceback.format_exc())
        return jsonify({"error": str(e)}), 500

@app.route("/history")
def history() -> Union[str, Tuple[str, int]]:
    try:
        queries_data: List[Dict[str, Any]] = db.get_recent_queries(20)
        logger.info(f"History page loaded with {len(queries_data)} queries")
        return render_template("history.html", queries=queries_data)
    except Exception as e:
        logger.error(f"Error loading history page: {str(e)}")
        logger.error(traceback.format_exc())
        return render_template("error.html", error=str(e))

@app.route("/history/<int:query_id>")
def view_query_results(query_id: int) -> Union[str, Tuple[str, int]]:
    logger.info(f"Viewing results for query ID: {query_id}")
    try:
        query_results_data: List[Dict[str, Any]] = db.get_query_results(query_id)
        if not query_results_data:
            logger.warning(f"No results found for query ID: {query_id}. Returning empty results snippet.")
            # Return an HTML snippet indicating no results, suitable for HTMX target replacement
            return "<div class=\"alert alert-warning\" role=\"alert\">No results found for this query.</div>"
        
        # Restructure results for template
        current_query: str = query_results_data[0]["query"]
        formatted_results: List[Tuple[str, float]] = [(row["document"], row["score"]) for row in query_results_data]
        logger.info(f"Retrieved query '{current_query}' with {len(formatted_results)} results")
        
        return render_template("results.html", 
                              query=current_query, 
                              results=formatted_results,
                              query_id=query_id,
                              from_history=True)
    except Exception as e:
        logger.error(f"Error viewing query results: {str(e)}")
        logger.error(traceback.format_exc())
        return render_template("error.html", error=str(e))

@app.route("/examples")
def examples() -> Union[str, Tuple[str, int]]:
    try:
        examples_dir: str = os.path.join(os.path.dirname(__file__), "examples")
        examples_data: List[Dict[str, str]] = []
        
        for filename in os.listdir(examples_dir):
            if filename.endswith(".yaml") or filename.endswith(".yml"):
                filepath: str = os.path.join(examples_dir, filename)
                with open(filepath, "r") as f:
                    content: str = f.read()
                examples_data.append({"name": filename, "content": content})
        
        logger.info(f"Examples page loaded with {len(examples_data)} examples")
        return render_template("examples.html", examples=examples_data)
    except Exception as e:
        logger.error(f"Error loading examples page: {str(e)}")
        logger.error(traceback.format_exc())
        return render_template("error.html", error=str(e))

@app.route("/cheatsheet")
def cheatsheet() -> str:
    logger.info("Cheatsheet page loaded")
    return render_template("cheatsheet.html")

@app.route("/documents")
def documents() -> Union[str, Tuple[str, int]]:
    try:
        docs_data: List[Dict[str, Any]] = db.get_all_documents()
        logger.info(f"Documents page loaded with {len(docs_data)} documents")
        return render_template("documents.html", documents=docs_data)
    except Exception as e:
        logger.error(f"Error loading documents page: {str(e)}")
        logger.error(traceback.format_exc())
        return render_template("error.html", error=str(e))

@app.template_filter('format_datetime')
def format_datetime(value: Union[str, datetime], format: str='%Y-%m-%d %H:%M:%S') -> str:
    try:
        if isinstance(value, str):
            dt: datetime = datetime.fromisoformat(value.replace('Z', '+00:00'))
        else:
            dt = value
        return dt.strftime(format)
    except Exception as e:
        logger.error(f"Error formatting datetime {value}: {str(e)}")
        return str(value)

@app.errorhandler(404)
def page_not_found(e: Exception) -> Tuple[str, int]:
    logger.warning(f"404 error: {request.path}")
    return render_template('error.html', error="Page not found"), 404

@app.errorhandler(500)
def server_error(e: Exception) -> Tuple[str, int]:
    logger.error(f"500 error: {str(e)}")
    return render_template('error.html', error="Internal server error"), 500

if __name__ == "__main__":
    logger.info("Starting reranker application")
    app.run(debug=True, host="0.0.0.0", port=5000) 