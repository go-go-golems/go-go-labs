from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch
import matplotlib.pyplot as plt
import numpy as np

MODEL = "cross-encoder/ms-marco-MiniLM-L6-v2"  # small; swap for BAAI/bge-reranker-large if you have GPU
tok = AutoTokenizer.from_pretrained(MODEL)
model = AutoModelForSequenceClassification.from_pretrained(MODEL)

query = "What is semantic chunking in RAG pipelines?"
doc = ("RAG systems benefit from semantically coherent chunks "
      "of text that align with conceptual boundaries.")

# --- forward pass with gradients enabled -----------------------
batch = tok([(query, doc)], return_tensors="pt", padding=True, truncation=True)

# Enable gradient tracking for input embeddings
model.eval()  # Still in eval mode, but we need gradients
batch_embeds = model.get_input_embeddings()(batch["input_ids"]).clone()  # Clone to make a leaf tensor
batch_embeds.requires_grad_(True)

# Run the model with embedding inputs instead of token IDs
attention_mask = batch["attention_mask"]
inputs = {"inputs_embeds": batch_embeds, "attention_mask": attention_mask}
outputs = model(**inputs)

# Get the score and backpropagate
score = outputs.logits[0, 0]  # Single relevance score
score.backward()

# Extract gradient magnitudes as saliency
if batch_embeds.grad is None:
    print("Warning: No gradients were computed. Using fallback method.")
    # Use a simpler approach: just look at absolute token embedding values
    saliency = torch.abs(batch_embeds[0]).sum(dim=1)  # Sum across embedding dimensions
else:
    saliency = torch.abs(batch_embeds.grad[0]).sum(dim=1)  # Sum across embedding dimensions

# --- visualize the importance of each token -------------------
tokens = tok.convert_ids_to_tokens(batch["input_ids"][0])

# Normalize saliency for better visualization
saliency = saliency / saliency.max()

plt.figure(figsize=(10, 6))

# Plot as a bar chart
plt.bar(range(len(tokens)), saliency.detach().numpy())
plt.xticks(range(len(tokens)), tokens, rotation=90)
plt.title(f"Token Importance by Gradient Saliency (Score: {score.item():.4f})")
plt.xlabel("Tokens")
plt.ylabel("Normalized Saliency")

# Add a vertical line at the separator to distinguish query and document
sep_idx = tokens.index(tok.sep_token)
plt.axvline(x=sep_idx, color='r', linestyle='--', alpha=0.5)
plt.text(sep_idx, saliency.max().item() * 0.9, "QUERY | DOC", rotation=90, color='r')

# Add labels for top 5 most salient tokens
top_indices = torch.argsort(saliency, descending=True)[:5]
for idx in top_indices:
    plt.text(idx, saliency[idx].item() + 0.03, tokens[idx], ha='center', va='bottom')

plt.tight_layout()
plt.show()