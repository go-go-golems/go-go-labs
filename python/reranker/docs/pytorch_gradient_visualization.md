# PyTorch Gradient Visualization for Rerankers: Concepts and Fixes

This document explains the PyTorch concepts used in our reranker visualization scripts, particularly focusing on gradient-based visualizations and the fixes implemented to handle common issues.

## Core PyTorch Concepts

### Leaf vs Non-Leaf Tensors

**Key Concept**: In PyTorch's autograd system, tensors are classified as either "leaf" or "non-leaf" nodes in the computation graph.

- **Leaf Tensors**: Created directly by the user (via `torch.tensor()`, etc.) or marked with `requires_grad=True`. These tensors can accumulate gradients.
- **Non-Leaf Tensors**: Result from operations on other tensors. By default, they don't retain gradients after backpropagation.

In our code, this line created a non-leaf tensor:
```python
batch_embeds = model.get_input_embeddings()(batch["input_ids"])
```

Which was fixed by adding `.clone()` to create a leaf tensor:
```python
batch_embeds = model.get_input_embeddings()(batch["input_ids"]).clone()
```

### Gradient Computation and Access

**Issue**: When we call `backward()` on a tensor, PyTorch computes gradients for all leaf tensors in the graph that require gradients. Non-leaf tensors don't store gradients by default.

The error we encountered:
```
UserWarning: The .grad attribute of a Tensor that is not a leaf Tensor is being accessed. Its .grad attribute won't be populated during autograd.backward().
```

Occurred because we tried to access `.grad` on a non-leaf tensor. The solution was twofold:

1. Make `batch_embeds` a leaf tensor with `.clone()`
2. Add a fallback for cases where gradients might still be `None`:
   ```python
   if batch_embeds.grad is None:
       # Fallback visualization method
       saliency = torch.abs(batch_embeds[0]).sum(dim=1)
   else:
       saliency = torch.abs(batch_embeds.grad[0]).sum(dim=1)
   ```

## Techniques for Reranker Visualization

### 1. Gradient-Based Saliency

We calculate token importance by:
1. Forward pass through the model
2. Backpropagation of the relevance score
3. Measuring the gradient magnitude of the input embeddings

This reveals which tokens most influence the final score when changed slightly.

```python
# Get the score and backpropagate
score = outputs.logits[0, 0]  # Single relevance score
score.backward()

# Extract gradient magnitudes as saliency
saliency = torch.abs(batch_embeds.grad[0]).sum(dim=1)  # Sum across embedding dimensions
```

### 2. Alternative: Embedding Magnitude Fallback

When gradients aren't available, we can approximate token importance by the magnitude of their embeddings:

```python
# Alternative method when gradients are unavailable
saliency = torch.abs(batch_embeds[0]).sum(dim=1)  # Sum across embedding dimensions
```

This is less accurate but serves as a useful fallback.

## Additional PyTorch Tips for Visualization

### 1. Ensuring Gradients with `retain_grad()`

If you need gradients for non-leaf tensors, use `.retain_grad()` on those tensors before backpropagation:

```python
hidden_states.retain_grad()  # Now this non-leaf tensor will store gradients
```

### 2. Creating Leaf Tensors

Methods to create leaf tensors from existing ones:
- `.clone()` - Creates a copy that's a leaf tensor
- `.detach().requires_grad_(True)` - Detaches from graph and makes a new leaf
- `torch.tensor(tensor.numpy())` - Recreates from scratch (less efficient)

### 3. Accessing Internal Model States

To access hidden states or attentions, initialize models with the appropriate flags:

```python
model = AutoModelForSequenceClassification.from_pretrained(
    MODEL_NAME, 
    output_attentions=True,           # Get attention matrices
    output_hidden_states=True         # Get hidden states for all layers
).eval()
```

## Conclusion

Understanding PyTorch's autograd system is crucial for creating visualization tools that leverage gradients for interpretation. By properly handling leaf vs. non-leaf tensors and implementing robust fallback mechanisms, we can create reliable visualization tools that help understand how reranker models make decisions.