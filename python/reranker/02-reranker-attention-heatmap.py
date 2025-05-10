from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch, matplotlib.pyplot as plt

MODEL = "cross-encoder/ms-marco-MiniLM-L6-v2"     # small; swap for BAAI/bge-reranker-large if you have GPU
tok   = AutoTokenizer.from_pretrained(MODEL)
model = AutoModelForSequenceClassification.from_pretrained(MODEL, output_attentions=True).eval()

query = "What is semantic chunking in RAG pipelines?"
doc   = ("RAG systems benefit from semantically coherent chunks "
         "of text that align with conceptual boundaries.")

# --- one forward pass with attentions --------------------------
batch = tok([(query, doc)], return_tensors="pt", padding=True, truncation=True)
with torch.no_grad():
    out = model(**batch, output_attentions=True)

att_last = out.attentions[-1][0]          # shape (heads, seq, seq)
avg_att  = att_last.mean(0)               # average over heads

tokens = tok.convert_ids_to_tokens(batch["input_ids"][0])
sep     = tokens.index(tok.sep_token)     # first [SEP] = end of query
q_idx   = range(1, sep)                   # skip [CLS]
d_idx   = range(sep + 1, len(tokens) - 1) # skip [SEP]s

heat = avg_att[q_idx][:, d_idx]           # Q→D attention only

# --- simple visual ---------------------------------------------
plt.figure(figsize=(8, 5))
plt.imshow(heat, aspect="auto")           # default colormap, one plot
plt.xticks(range(len(d_idx)), [tokens[i] for i in d_idx], rotation=90)
plt.yticks(range(len(q_idx)),  [tokens[i] for i in q_idx])
plt.title("Average attention (last layer, all heads)\nQuery ➜ Document tokens")
plt.tight_layout(); plt.show()
