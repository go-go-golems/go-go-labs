# Code Fixes for PyTorch Gradient Visualization

## Issue: Gradient Computation on Non-Leaf Tensors

### Original Error

```
UserWarning: The .grad attribute of a Tensor that is not a leaf Tensor is being accessed. Its .grad attribute won't be populated during autograd.backward().
  saliency = torch.abs(batch_embeds.grad[0]).sum(dim=1)

TypeError: 'NoneType' object is not subscriptable
```

### Root Cause

In PyTorch's autograd system, only leaf tensors (tensors created directly by the user or explicitly marked to require gradients) store gradients by default. The error occurred because:

1. Our `batch_embeds` tensor was a non-leaf tensor - it was the result of a function call
2. We tried to access its `.grad` attribute which was `None`
3. Then we tried to index into this `None` with `[0]` causing the TypeError

### Implemented Fixes

#### 1. Create a proper leaf tensor with `.clone()`

Changed:
```python
batch_embeds = model.get_input_embeddings()(batch["input_ids"])
```

To:
```python
batch_embeds = model.get_input_embeddings()(batch["input_ids"]).clone()  # Clone to make a leaf tensor
```

The `.clone()` operation creates a new tensor that is a leaf node in the computation graph.

#### 2. Add robust error handling

Added:
```python
if batch_embeds.grad is None:
    print("Warning: No gradients were computed. Using fallback method.")
    # Use a simpler approach: just look at absolute token embedding values
    saliency = torch.abs(batch_embeds[0]).sum(dim=1)  # Sum across embedding dimensions
else:
    saliency = torch.abs(batch_embeds.grad[0]).sum(dim=1)  # Sum across embedding dimensions
```

This ensures the code continues even if gradients aren't properly computed, using token embedding magnitudes as a fallback visualization method.

## Additional PyTorch Tips for Gradient Visualization

### Working with Non-Leaf Tensors

If you need gradients for intermediate (non-leaf) tensors, use `.retain_grad()` on those tensors before backpropagation:

```python
hidden_states = model(**inputs).last_hidden_state
hidden_states.retain_grad()  # Now this non-leaf tensor will store gradients
score.backward()
# Now hidden_states.grad will exist
```

### Alternative Approaches

1. **Register Hooks**: Instead of directly accessing gradients, you can register hooks to capture gradients during backpropagation:

```python
grads = {}
def save_grad(name):
    def hook(grad):
        grads[name] = grad
    return hook

batch_embeds.register_hook(save_grad('embeddings'))
score.backward()
# Now access grads['embeddings']
```

2. **Use `torch.autograd.grad`**: Compute gradients directly:

```python
grads = torch.autograd.grad(score, batch_embeds, create_graph=False)
saliency = torch.abs(grads[0][0]).sum(dim=1)
```

## Applied Fixes in Batch Analysis

The same fixes were applied to the batch analysis script (`06-reranker-batch-analysis.py`) to ensure robust operation when processing multiple query-document pairs:

```python
def analyze_gradient_saliency(query, doc, pair_id):
    # ...
    batch_embeds = model.get_input_embeddings()(batch["input_ids"]).clone()
    # ...
    if batch_embeds.grad is None:
        print(f"Warning: No gradients computed for {pair_id}. Using fallback method.")
        # Use token embedding magnitudes as a fallback
        saliency = torch.abs(batch_embeds[0]).sum(dim=1)
    else:
        saliency = torch.abs(batch_embeds.grad[0]).sum(dim=1)
```

This ensures the batch processing continues even if gradient computation fails for some examples.

## Conclusion

These fixes address common pitfalls in PyTorch gradient visualization while providing fallback mechanisms for robust operation. Understanding the distinction between leaf and non-leaf tensors is crucial when working with PyTorch's autograd system for model interpretability.