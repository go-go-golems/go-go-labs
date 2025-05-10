from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch
import matplotlib.pyplot as plt
import numpy as np
from sklearn.decomposition import PCA
from matplotlib.animation import FuncAnimation

MODEL = "cross-encoder/ms-marco-MiniLM-L6-v2"  # small; swap for BAAI/bge-reranker-large if you have GPU
tok = AutoTokenizer.from_pretrained(MODEL)
model = AutoModelForSequenceClassification.from_pretrained(MODEL, output_hidden_states=True).eval()

query = "What is semantic chunking in RAG pipelines?"
doc = ("RAG systems benefit from semantically coherent chunks "
      "of text that align with conceptual boundaries.")

# --- one forward pass capturing hidden states for all layers ---
batch = tok([(query, doc)], return_tensors="pt", padding=True, truncation=True)
with torch.no_grad():
    outputs = model(**batch, output_hidden_states=True)

# Get all hidden states (layer by layer)
hidden_states = outputs.hidden_states  # tuple of tensors (layer, batch, seq_len, hidden_dim)

# Convert to a single tensor for easier processing
all_hidden_states = torch.stack(hidden_states)  # shape: (layers, batch, seq_len, hidden_dim)
all_hidden_states = all_hidden_states.squeeze(1)  # remove batch dimension, now (layers, seq_len, hidden_dim)

# Get tokens
tokens = tok.convert_ids_to_tokens(batch["input_ids"][0])

# Select interesting tokens to track
interesting_tokens = [
    tokens.index("semantic") if "semantic" in tokens else tokens.index("##semantic") if "##semantic" in tokens else 1,
    tokens.index("chunk") if "chunk" in tokens else tokens.index("##chunk") if "##chunk" in tokens else 2,
    tokens.index("RAG") if "RAG" in tokens else 3,
]

token_names = [tokens[idx] for idx in interesting_tokens]

# Collect hidden states for these tokens across all layers
token_states = [all_hidden_states[:, idx, :].cpu().numpy() for idx in interesting_tokens]  # list of (layers, hidden_dim)

# Apply PCA to visualize in 2D
pca = PCA(n_components=2)
# Flatten all token states from all layers into one big matrix
all_vectors = np.vstack([states for states in token_states])
pca_fit = pca.fit_transform(all_vectors)

# Reshape back to get PCA'd trajectories for each token
num_layers = all_hidden_states.shape[0]
token_trajectories = [pca_fit[i*num_layers:(i+1)*num_layers] for i in range(len(interesting_tokens))]

# --- create an animated plot of token trajectories ------------
fig, ax = plt.subplots(figsize=(10, 8))

# Set axis limits based on all data points
x_min, y_min = pca_fit.min(axis=0) - 0.5
x_max, y_max = pca_fit.max(axis=0) + 0.5
ax.set_xlim(x_min, x_max)
ax.set_ylim(y_min, y_max)

# Create line objects for each token
lines = [ax.plot([], [], 'o-', lw=2, label=name)[0] for name in token_names]

# Create text objects to show layer numbers
texts = [ax.text(0, 0, '', fontsize=10) for _ in range(len(interesting_tokens))]

# Function to update the animation at each frame
def update(frame):
    for i, (line, text) in enumerate(zip(lines, texts)):
        # Up to the current frame, add points to the line
        x_data = token_trajectories[i][:frame+1, 0]
        y_data = token_trajectories[i][:frame+1, 1]
        line.set_data(x_data, y_data)
        
        # Update position of layer number text
        if frame > 0:
            text.set_position((x_data[-1], y_data[-1]))
            text.set_text(f"L{frame}")
    
    ax.set_title(f"Token Trajectory Through Layers (Layer {frame}/{num_layers-1})")
    return lines + texts

# Create the animation
ani = FuncAnimation(fig, update, frames=num_layers, interval=500, blit=True)

# Add legend, title and labels
ax.legend(loc='upper right')
ax.set_xlabel('PCA Component 1')
ax.set_ylabel('PCA Component 2')
ax.set_title('Token Trajectory Through Layers')

plt.tight_layout()
plt.show()

# Uncomment to save animation
# ani.save('token_trajectory.gif', writer='pillow', fps=2)