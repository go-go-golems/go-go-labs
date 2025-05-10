#!/usr/bin/env python3
import os
import yaml
from flask import Flask, render_template, request, jsonify

from reranker import rerank

app = Flask(__name__)
app.config["MAX_CONTENT_LENGTH"] = 16 * 1024 * 1024  # 16MB max upload size

@app.route("/")
def index():
    return render_template("index.html")

@app.route("/rerank", methods=["POST"])
def handle_rerank():
    if "file" not in request.files and not request.form.get("yaml_text"):
        return jsonify({"error": "No file or YAML text provided"}), 400
    
    try:
        if "file" in request.files and request.files["file"].filename:
            yaml_content = request.files["file"].read().decode("utf-8")
        else:
            yaml_content = request.form.get("yaml_text", "")
        
        data = yaml.safe_load(yaml_content)
        
        if not isinstance(data, dict) or "query" not in data or "documents" not in data:
            return jsonify({"error": "Invalid YAML format. Must contain 'query' and 'documents' fields"}), 400
        
        query = data["query"]
        documents = data["documents"]
        top_k = data.get("top_k", None)
        
        if not isinstance(documents, list):
            return jsonify({"error": "Documents must be a list"}), 400
        
        results = rerank(query, documents, top_k)
        
        return render_template("results.html", 
                              query=query, 
                              results=results)
    
    except yaml.YAMLError:
        return jsonify({"error": "Invalid YAML format"}), 400
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route("/examples")
def examples():
    examples_dir = os.path.join(os.path.dirname(__file__), "examples")
    examples = []
    
    for filename in os.listdir(examples_dir):
        if filename.endswith(".yaml") or filename.endswith(".yml"):
            filepath = os.path.join(examples_dir, filename)
            with open(filepath, "r") as f:
                content = f.read()
            examples.append({"name": filename, "content": content})
    
    return render_template("examples.html", examples=examples)

@app.route("/cheatsheet")
def cheatsheet():
    return render_template("cheatsheet.html")

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5000) 