#!/usr/bin/env python

# https://chatgpt.com/c/681e9a2e-e2a0-8012-9603-c8c759f26383

"""
Quick-n-dirty demo: run BGE-Reranker (large) on a mock query + 5 documents
and print scores in descending order.

Requires:
  pip install torch transformers sentencepiece
  # optional for speed: pip install accelerate
"""

from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch

# --- load model once ---------------------------------------------------------
MODEL_NAME = "BAAI/bge-reranker-large"
tokenizer   = AutoTokenizer.from_pretrained(MODEL_NAME)
model       = AutoModelForSequenceClassification.from_pretrained(MODEL_NAME)
model.eval()           # inference mode
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
model.to(device)

# --- mock query + candidate docs --------------------------------------------
query = "semantic chunking for retrieval-augmented generation in large language models"

docs = [
    {
        "id": "A1",
        "title": "Hierarchical semantic chunking improves RAG precision",
        "abstract": "We propose a hierarchical tokenizer that splits text into semantically coherent chunks..."
    },
    {
        "id": "B2",
        "title": "Adjustable rotor blades for offshore wind turbines",
        "abstract": "This engineering study explores variable-pitch rotor blades designed for high-altitude wind..."
    },
    {
        "id": "C3",
        "title": "Fast embedding pooling strategies",
        "abstract": "Pooling techniques such as mean, CLS-token and attention pooling are compared across BERT-based models..."
    },
    {
        "id": "D4",
        "title": "Automatic keyphrase extraction with token-level vectors",
        "abstract": "We evaluate token-level embeddings as features for unsupervised keyphrase extraction..."
    },
    {
        "id": "E5",
        "title": "Long-context transformers for academic search",
        "abstract": "A 32k-token transformer is trained to rerank academic papers retrieved from Crossref and arXiv..."
    },
]

# --- build [query] <SEP> [doc] pairs ----------------------------------------
pairs = [f"{query} [SEP] {d['title']}. {d['abstract']}" for d in docs]

enc = tokenizer(
    pairs,
    padding=True,
    truncation=True,
    max_length=512,
    return_tensors="pt"
).to(device)

with torch.no_grad():
    scores = model(**enc).logits.squeeze(-1)   # higher = more relevant

# --- sort + show ------------------------------------------------------------
ranked = sorted(zip(docs, scores.tolist()), key=lambda x: x[1], reverse=True)

for doc, score in ranked:
    print(f"{score: .4f} | {doc['id']} | {doc['title']}")

"""
Expected console output (numbers will vary):
  10.7321 | A1 | Hierarchical semantic chunking improves RAG precision
   9.8875 | E5 | Long-context transformers for academic search
   8.4032 | D4 | Automatic keyphrase extraction with token-level vectors
   7.9210 | C3 | Fast embedding pooling strategies
  -1.0023 | B2 | Adjustable rotor blades for offshore wind turbines
"""

