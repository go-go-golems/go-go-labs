#!/usr/bin/env python3

from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch

DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
MODEL_NAME = "BAAI/bge-reranker-large"

# Lazy loading of model and tokenizer
_tokenizer = None
_model = None

def _get_tokenizer():
    global _tokenizer
    if _tokenizer is None:
        _tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
    return _tokenizer

def _get_model():
    global _model
    if _model is None:
        _model = AutoModelForSequenceClassification.from_pretrained(MODEL_NAME).to(DEVICE)
        _model.eval()
    return _model

def rerank(query: str, documents: list[str], top_k: int | None = None):
    """Return (doc, score) sorted best-to-worst for a single query."""
    if not documents:
        return []
        
    # Get tokenizer and model
    tokenizer = _get_tokenizer()
    model = _get_model()
    
    # build parallel list of (query, doc) pairs, batch-encode, and run the model
    pairs = [(query, doc) for doc in documents]
    batch = tokenizer.batch_encode_plus(
        pairs, padding=True, truncation=True, return_tensors="pt"
    ).to(DEVICE)

    with torch.no_grad():
        # logits[:, 0] is a single relevance score for every pair
        scores = model(**batch).logits.squeeze(-1).cpu()

    ranked = sorted(zip(documents, scores.tolist()), key=lambda x: x[1], reverse=True)
    return ranked if top_k is None else ranked[:top_k] 