from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch
import matplotlib.pyplot as plt
import numpy as np
import os
from tqdm import tqdm

MODEL = "cross-encoder/ms-marco-MiniLM-L6-v2"  # small; swap for BAAI/bge-reranker-large if you have GPU
tok = AutoTokenizer.from_pretrained(MODEL)
model = AutoModelForSequenceClassification.from_pretrained(MODEL, output_attentions=True, output_hidden_states=True).eval()

# Create output directory
OUTPUT_DIR = "reranker_analysis_outputs"
os.makedirs(OUTPUT_DIR, exist_ok=True)

# Sample query-document pairs for batch analysis
queries = [
    "What is semantic chunking in RAG pipelines?",
    "How do transformer models handle long context?",
    "What are the best practices for fine-tuning LLMs?"
]

documents = [
    [
        "Chunking breaks text into fixed-length windows.",
        "RAG systems benefit from semantically coherent chunks.",
        "Chunking is unrelated to computer memory management."
    ],
    [
        "Transformers use attention to process sequences in parallel.",
        "Long context handling requires efficient memory management.",
        "Position embeddings are crucial for sequence understanding."
    ],
    [
        "Fine-tuning LLMs requires careful learning rate scheduling.",
        "Small datasets often lead to overfitting in LLM fine-tuning.",
        "PEFT methods reduce parameter count for efficient training."
    ]
]

# --- ANALYSIS FUNCTIONS ---

def analyze_attention_heads(query, doc, pair_id):
    """Generate attention head visualization for a query-doc pair."""
    batch = tok([(query, doc)], return_tensors="pt", padding=True, truncation=True)
    
    with torch.no_grad():
        out = model(**batch, output_attentions=True)
    
    att_last = out.attentions[-1][0]  # shape (heads, seq, seq)
    tokens = tok.convert_ids_to_tokens(batch["input_ids"][0])
    sep = tokens.index(tok.sep_token)  # first [SEP] = end of query
    q_idx = range(1, sep)  # skip [CLS]
    d_idx = range(sep + 1, len(tokens) - 1)  # skip [SEP]s
    
    # Average attention across heads
    avg_att = att_last.mean(0)
    heat = avg_att[q_idx][:, d_idx]  # Q→D attention only
    
    plt.figure(figsize=(10, 6))
    plt.imshow(heat, aspect="auto")
    plt.colorbar()
    plt.xticks(range(len(d_idx)), [tokens[i] for i in d_idx], rotation=90)
    plt.yticks(range(len(q_idx)), [tokens[i] for i in q_idx])
    plt.title(f"Average Attention: Query → Document\nScore: {out.logits[0][0].item():.4f}")
    plt.tight_layout()
    
    # Save figure
    plt.savefig(f"{OUTPUT_DIR}/attention_{pair_id}.png")
    plt.close()

def analyze_gradient_saliency(query, doc, pair_id):
    """Generate gradient saliency visualization for a query-doc pair."""
    batch = tok([(query, doc)], return_tensors="pt", padding=True, truncation=True)
    
    # Enable gradient tracking for input embeddings
    model.eval()
    batch_embeds = model.get_input_embeddings()(batch["input_ids"]).clone()  # Clone to make a leaf tensor
    batch_embeds.requires_grad_(True)
    
    # Run the model with embedding inputs
    attention_mask = batch["attention_mask"]
    inputs = {"inputs_embeds": batch_embeds, "attention_mask": attention_mask}
    outputs = model(**inputs)
    
    # Get the score and backpropagate
    score = outputs.logits[0, 0]  # Single relevance score
    score.backward()
    
    # Extract gradient magnitudes as saliency
    if batch_embeds.grad is None:
        print(f"Warning: No gradients computed for {pair_id}. Using fallback method.")
        # Use token embedding magnitudes as a fallback
        saliency = torch.abs(batch_embeds[0]).sum(dim=1)
    else:
        saliency = torch.abs(batch_embeds.grad[0]).sum(dim=1)
    
    # Prepare for visualization
    tokens = tok.convert_ids_to_tokens(batch["input_ids"][0])
    saliency = saliency / saliency.max()  # Normalize
    
    plt.figure(figsize=(12, 5))
    plt.bar(range(len(tokens)), saliency.detach().numpy())
    plt.xticks(range(len(tokens)), tokens, rotation=90)
    plt.title(f"Token Importance by Gradient Saliency (Score: {score.item():.4f})")
    
    # Add a vertical line at the separator
    sep_idx = tokens.index(tok.sep_token)
    plt.axvline(x=sep_idx, color='r', linestyle='--')
    
    plt.tight_layout()
    plt.savefig(f"{OUTPUT_DIR}/saliency_{pair_id}.png")
    plt.close()

# --- MAIN BATCH ANALYSIS ---

def run_batch_analysis():
    """Run analysis on all query-document pairs and save visualizations."""
    # Track all scores for summary
    all_scores = []
    
    # Process each query and its documents
    for q_idx, query in enumerate(queries):
        print(f"Processing query {q_idx+1}/{len(queries)}: {query}")
        
        # Process each document for this query
        query_scores = []
        for d_idx, doc in enumerate(documents[q_idx]):
            pair_id = f"q{q_idx+1}_d{d_idx+1}"
            print(f"  Analyzing pair {pair_id}: {doc[:30]}...")
            
            # Run the model to get the score
            batch = tok([(query, doc)], return_tensors="pt", padding=True, truncation=True)
            with torch.no_grad():
                score = model(**batch).logits[0][0].item()
            
            query_scores.append((doc, score))
            all_scores.append((query, doc, score))
            
            # Generate visualizations
            analyze_attention_heads(query, doc, pair_id)
            analyze_gradient_saliency(query, doc, pair_id)
        
        # Rank documents for this query
        ranked_docs = sorted(query_scores, key=lambda x: x[1], reverse=True)
        print(f"  Ranked results for query {q_idx+1}:")
        for rank, (doc, score) in enumerate(ranked_docs):
            print(f"    {rank+1}. [{score:.4f}] {doc[:50]}...")
        print()
    
    # Generate summary report
    plt.figure(figsize=(10, 6))
    plt.bar(range(len(all_scores)), [s[2] for s in all_scores])
    plt.xticks(range(len(all_scores)), [f"Q{i//3+1}D{i%3+1}" for i in range(len(all_scores))], rotation=45)
    plt.title("All Query-Document Scores")
    plt.tight_layout()
    plt.savefig(f"{OUTPUT_DIR}/score_summary.png")
    plt.close()
    
    print(f"\nAnalysis complete! Visualizations saved to {OUTPUT_DIR}/")

if __name__ == "__main__":
    run_batch_analysis()