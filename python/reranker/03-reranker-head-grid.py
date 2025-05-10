from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch
import matplotlib.pyplot as plt
import numpy as np

MODEL = "cross-encoder/ms-marco-MiniLM-L6-v2"  # small; swap for BAAI/bge-reranker-large if you have GPU
tok = AutoTokenizer.from_pretrained(MODEL)
model = AutoModelForSequenceClassification.from_pretrained(MODEL, output_attentions=True).eval()

query = "What is semantic chunking in RAG pipelines?"
doc = ("RAG systems benefit from semantically coherent chunks "
      "of text that align with conceptual boundaries.")

# --- one forward pass with attentions --------------------------
batch = tok([(query, doc)], return_tensors="pt", padding=True, truncation=True)
with torch.no_grad():
    out = model(**batch, output_attentions=True)

att_last = out.attentions[-1][0]          # shape (heads, seq, seq)

tokens = tok.convert_ids_to_tokens(batch["input_ids"][0])
sep = tokens.index(tok.sep_token)     # first [SEP] = end of query
q_idx = range(1, sep)                 # skip [CLS]
d_idx = range(sep + 1, len(tokens) - 1)  # skip [SEP]s

# --- plot grid of attention heads ------------------------------
num_heads = att_last.shape[0]
rows = int(np.ceil(np.sqrt(num_heads)))
cols = int(np.ceil(num_heads / rows))

plt.figure(figsize=(15, 15))

for head in range(num_heads):
    head_att = att_last[head]
    query_to_doc_att = head_att[q_idx][:, d_idx]  # Q→D attention only
    
    plt.subplot(rows, cols, head + 1)
    plt.imshow(query_to_doc_att, aspect="auto")
    plt.title(f"Head {head}")
    
    # Only add labels for rightmost and bottom plots
    if head % cols == 0:  # leftmost column
        plt.yticks(range(len(q_idx)), [tokens[i] for i in q_idx])
    else:
        plt.yticks([])
        
    if head >= num_heads - cols:  # bottom row
        plt.xticks(range(len(d_idx)), [tokens[i] for i in d_idx], rotation=90)
    else:
        plt.xticks([])

plt.suptitle(f"Attention Heads - Last Layer (Query → Document)")
plt.tight_layout()
plt.subplots_adjust(top=0.95)  # Make room for suptitle
plt.show()