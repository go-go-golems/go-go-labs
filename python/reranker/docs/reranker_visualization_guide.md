# Reranker Visualization Guide

## Overview of Visualization Scripts

This document provides an overview of the four reranker visualization scripts created to help understand and interpret cross-encoder reranker models.

## 1. Head-Level Inspection (03-reranker-head-grid.py)

### Purpose
Inspect individual attention heads rather than just the average attention, revealing specialized patterns in different heads.

### Technique
```python
# Instead of using mean(0) to average across heads
avg_att = att_last.mean(0)  # This hides individual head behaviors

# We plot each head separately in a grid
for head in range(num_heads):
    head_att = att_last[head]
    query_to_doc_att = head_att[q_idx][:, d_idx]  # Q→D attention only
    
    plt.subplot(rows, cols, head + 1)
    plt.imshow(query_to_doc_att, aspect="auto")
```

### Insights Provided
- Reveals specialized heads (e.g., some focus on syntax, others on semantics)
- Identifies attention patterns that might be diluted in the average
- Shows which heads contribute most to model decisions

## 2. Gradient-Based Saliency (04-reranker-gradient-saliency.py)

### Purpose
Identify which tokens most influence the final relevance score by measuring how sensitive the score is to small changes in token embeddings.

### Technique
```python
# Enable gradient tracking
batch_embeds = model.get_input_embeddings()(batch["input_ids"]).clone()
batch_embeds.requires_grad_(True)

# Run model and backpropagate from relevance score
score = outputs.logits[0, 0]  # Single relevance score
score.backward()

# Measure gradient magnitudes
saliency = torch.abs(batch_embeds.grad[0]).sum(dim=1)
```

### Insights Provided
- Identifies the most influential tokens for model decisions
- Shows which terms in the query and document contribute most to relevance
- Can reveal attention to tokens not obvious from attention matrices alone

## 3. Token Trajectory Visualization (05-reranker-token-trajectory.py)

### Purpose
Track how representations of key tokens evolve through the model's layers using dimensionality reduction.

### Technique
```python
# Collect hidden states for tokens across all layers
token_states = [all_hidden_states[:, idx, :].cpu().numpy() for idx in interesting_tokens]

# Reduce dimensionality for visualization
pca = PCA(n_components=2)
all_vectors = np.vstack([states for states in token_states])
pca_fit = pca.fit_transform(all_vectors)

# Create animation showing token movement through layers
def update(frame):
    for i, (line, text) in enumerate(zip(lines, texts)):
        x_data = token_trajectories[i][:frame+1, 0]
        y_data = token_trajectories[i][:frame+1, 1]
        line.set_data(x_data, y_data)
```

### Insights Provided
- Visualizes how token meanings evolve as they pass through the model
- Reveals which layers perform the most significant transformations
- Shows when tokens' representations converge or diverge

## 4. Batch Analysis (06-reranker-batch-analysis.py)

### Purpose
Scale up the previous visualizations to analyze multiple query-document pairs, saving results for easy comparison.

### Technique
```python
def run_batch_analysis():
    # Process each query and its documents
    for q_idx, query in enumerate(queries):
        for d_idx, doc in enumerate(documents[q_idx]):
            pair_id = f"q{q_idx+1}_d{d_idx+1}"
            
            # Generate visualizations
            analyze_attention_heads(query, doc, pair_id)
            analyze_gradient_saliency(query, doc, pair_id)
```

### Insights Provided
- Enables comparison across different query-document pairs
- Facilitates quality assurance of reranker behavior at scale
- Helps identify patterns or biases in model attention across various inputs

## Implementation Tips

1. **Model Size Considerations**: For local development, use smaller models like `cross-encoder/ms-marco-MiniLM-L6-v2` instead of larger models like `BAAI/bge-reranker-large` if GPU resources are limited.

2. **Batch Size**: The examples use batch size 1 for clarity, but you can increase batch size for efficiency when visualizing multiple query-document pairs.

3. **Saving Visualizations**: Both static images (PNG) and interactive visualizations (HTML with plotly) can be useful for different scenarios.

4. **Error Handling**: Always implement fallback mechanisms for visualization in case of gradient computation issues or other unexpected model behaviors.

## Next Steps for Reranker Analysis

- **Compare Against Baselines**: Visualize both rerankers and regular retrieval models on the same inputs
- **Correlation Analysis**: Relate attention/gradients patterns to final score changes
- **Fine-tuning Impacts**: Compare visualizations before and after fine-tuning
- **Custom Dataset Analysis**: Apply these techniques to domain-specific queries and documents