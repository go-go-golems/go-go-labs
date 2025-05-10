#! /usr/bin/env python3

# https://chatgpt.com/c/681fb188-5f60-8012-9b5d-c94a4ebe5fa5
from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch

DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
MODEL_NAME = "BAAI/bge-reranker-large"      # swap to another HF model if you prefer

# one-time setup
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
model     = AutoModelForSequenceClassification.from_pretrained(MODEL_NAME).to(DEVICE)
model.eval()                                # we never train in a reranking pass

def rerank(query: str, documents: list[str], top_k: int | None = None):
    """Return (doc, score) sorted best-to-worst for a single query."""
    # build parallel list of (query, doc) pairs, batch-encode, and run the model
    pairs    = [(query, doc) for doc in documents]
    batch    = tokenizer.batch_encode_plus(
        pairs, padding=True, truncation=True, return_tensors="pt"
    ).to(DEVICE)

    with torch.no_grad():
        # logits[:, 0] is a single relevance score for every pair
        scores = model(**batch).logits.squeeze(-1).cpu()

    ranked = sorted(zip(documents, scores.tolist()), key=lambda x: x[1], reverse=True)
    return ranked if top_k is None else ranked[:top_k]

# quick smoke test
if __name__ == "__main__":
    query = "What is semantic chunking in RAG pipelines?"
    docs  = [
        "Chunking breaks text into fixed-length windows.",
        "RAG systems benefit from semantically coherent chunks.",
        "Chunking is unrelated to computer memory management."
    ]

    for doc, score in rerank(query, docs):
        print(f"{score: .4f}\t{doc}")
