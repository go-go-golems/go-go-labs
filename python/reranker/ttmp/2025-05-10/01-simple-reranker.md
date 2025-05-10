link: https://chatgpt.com/c/681fb188-5f60-8012-9b5d-c94a4ebe5fa5
---

Below is a minimal end-to-end example of a **cross-encoder reranker** in Python that relies only on 🤗 Transformers + PyTorch. It uses BAAI’s **`bge-reranker-large`** (multilingual, strong accuracy, still light enough to run on a single GPU or CPU) but you can swap in any cross-encoder ranking model published on the Hub (e.g. `cross-encoder/ms-marco-MiniLM-L6-v2`) with no other code changes. ([Hugging Face][1], [Hugging Face][2])

---

### 1 . Install the two libraries you actually need

```bash
pip install --upgrade transformers torch  # accelerate is optional but speeds up GPU I/O
```

### 2 . Pure-Python “hello-world” reranker

```python
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
```

**How it works**

| Step         | What happens                                                                                                         |
| ------------ | -------------------------------------------------------------------------------------------------------------------- |
| Tokenisation | Each *(query, passage)* pair is joined with the appropriate separator token for the underlying language model.       |
| Forward pass | The cross-encoder reads both texts simultaneously and emits a single *relevance score* without producing embeddings. |
| Sorting      | Higher logits mean “more relevant”, so we just rank descending.                                                      |

---

### 3 . Why cross-encoders are the simplest strong baseline

* **One model, one pass** – no dual-encoder embedding look-ups and cosine maths.
* **Out-of-the-box quality** – models like `bge-reranker-v2-m3` or the MiniLM family were fine-tuned on MS MARCO / BEIR and already outperform many hand-rolled heuristics. ([Hugging Face][3])
* **No new indexing layer** – you can treat your existing BM25 / vector-store hits as *candidates* and just call `rerank()` on the top N.

---

### 4 . Next 30-second tweaks (optional)

```python
# switch model
MODEL_NAME = "cross-encoder/ms-marco-MiniLM-L6-v2"  # smaller, English-only

# ask only for the best 20 documents
top20 = rerank(query, docs_from_your_retriever, top_k=20)
```

That’s it—drop this function between your retriever and the end user and you have a robust, production-ready reranker with \~10 lines of real logic.

[1]: https://huggingface.co/BAAI/bge-reranker-large?utm_source=chatgpt.com "BAAI/bge-reranker-large - Hugging Face"
[2]: https://huggingface.co/cross-encoder/ms-marco-MiniLM-L6-v2?utm_source=chatgpt.com "cross-encoder/ms-marco-MiniLM-L6-v2 - Hugging Face"
[3]: https://huggingface.co/BAAI/bge-reranker-v2-m3?utm_source=chatgpt.com "BAAI/bge-reranker-v2-m3 - Hugging Face"
