Great. I’ll begin compiling a comprehensive book for senior developers with some ML background on building RAG (Retrieval-Augmented Generation) systems using cross-encoders. This will include theoretical foundations, implementation using Hugging Face in Python, system design and evaluation techniques, and experimentation workflows. I'll also cover deployment and performance considerations focused on cross-encoder-based architectures.

I’ll let you know when the initial content is ready for review.


# **Building Retrieval-Augmented Generation Systems with Cross-Encoders**

Retrieval-Augmented Generation (RAG) is a powerful approach that combines information retrieval with text generation to create more factual and context-aware AI systems. This book is a comprehensive guide for experienced developers (with some ML background) to build and experiment with RAG systems using **cross-encoder** models. We will cover the theory behind RAG, compare cross-encoders vs. bi-encoders, implement pipelines with Hugging Face and PyTorch, fine-tune models, evaluate their outputs, deploy them as APIs, and walk through real-world examples in question answering and document retrieval.

## **Chapter 1: Foundations of Retrieval-Augmented Generation and Cross-Encoders**

### **1.1 What is Retrieval-Augmented Generation (RAG)?**

RAG is an architecture that augments a generative model (like a language model) with a **retrieval component** to supply relevant context from an external knowledge base (documents, articles, etc.) during generation. Instead of relying solely on parametric knowledge in the model’s weights, a RAG system actively retrieves documents related to the user’s query and uses their content to generate more accurate and up-to-date answers. In other words, an input question or prompt is first used to fetch relevant textual documents, and then those documents are given as additional context for a generator (such as a Seq2Seq transformer) to produce the final answer. This approach can increase the factual accuracy and trustworthiness of the output, since the model can cite or draw from retrieved evidence rather than hallucinating facts.

At its core, a RAG system consists of two main learned components:

* **Question Encoder (Retriever)** – This transforms the input query into a representation used to find relevant documents. In original RAG implementations, this is often a *bi-encoder* model like DPR (Dense Passage Retriever) that encodes the query into a vector embedding.
* **Generator (Language Model)** – This is a sequence-to-sequence model (e.g. BART or T5) that takes the question along with retrieved documents (typically concatenated or prepended as context) and generates an answer. The generator uses the retrieved text to ground its output in external knowledge.

During a RAG forward pass, the system encodes the query, retrieves a set of top-\$k\$ documents from a corpus, and then feeds those documents plus the query into the generator to produce the answer. The retriever and generator can be trained jointly or separately. An example: the **original RAG paper (2020)** by Facebook AI used DPR (a dual-encoder bi-encoder) to retrieve Wikipedia passages and a BART-large model to generate answers, achieving strong results on knowledge-intensive QA tasks.

**Why RAG?** Traditional large language models have fixed knowledge and can produce outdated or incorrect information. RAG allows *injecting external knowledge at query time*, so the model can utilize up-to-date and specific information without retraining on new data. This makes systems more efficient (no need to fine-tune a giant model for each knowledge update) and often more interpretable, since you can trace which documents were used to generate the answer. In summary, RAG enhances generative AI with *information retrieval* to improve accuracy and relevancy of responses.

### **1.2 Bi-Encoders vs. Cross-Encoders for Retrieval**

A crucial aspect of RAG is the **retrieval step**: finding documents relevant to the user’s query. There are two major types of neural models for text retrieval:

* **Bi-Encoders (Dual Encoders)** – These encode the query and documents *separately* into vector embeddings. The similarity (e.g. cosine or dot-product) between the query embedding and a document embedding indicates relevance. Bi-encoders enable **efficient** retrieval: one can pre-compute embeddings for an entire document corpus and index them in a vector database. At query time, the query is encoded and the nearest neighbor search retrieves the most similar document vectors in milliseconds. This approach scales to very large corpora (millions of docs) because similarity search in embedding space can be accelerated with Approximate Nearest Neighbors (ANN) algorithms. However, because the query and document are encoded independently, bi-encoders may miss fine-grained interactions between query and text, resulting in somewhat lower accuracy compared to cross-encoders. They compress each text into a single fixed vector, which can lead to some **information loss** about the detailed query-document relationship.

* **Cross-Encoders** – These models take the query and a candidate document *together* as input and output a relevance score (for instance, a number between 0 and 1 indicating how well the document answers the query). A cross-encoder typically concatenates the query and document text, and feeds the pair through a transformer (like BERT) to directly predict a relevance or matching score. This allows the model to consider every interaction between query words and document words, making it **highly accurate** for relevance ranking. Cross-encoders have shown superior performance in identifying truly relevant documents because they examine the full context of the query-document pair. The trade-off is that they are **computationally expensive**: for each query-document pair, a full forward pass through a transformer is required. This does not scale to searching an entire corpus of millions of documents *per query* in real-time. Cross-encoders are thus typically used for re-ranking a smaller set of candidate documents rather than initial large-scale retrieval.

&#x20;*Figure 1: **Bi-Encoder vs. Cross-Encoder architectures.** A Bi-Encoder (left) independently encodes a query (Input A) and a document (Input B) into vectors (e.g., using BERT with pooling) and computes a similarity score (such as cosine similarity) between them. A Cross-Encoder (right) feeds the query and document together into a transformer model (e.g., BERT) and directly outputs a relevance score via a classifier layer. Cross-encoders can model interactions between the query and document tokens, generally yielding higher accuracy at the cost of speed.*

**Key Trade-offs – Efficiency vs. Accuracy:** Bi-encoders offer speed and scalability, while cross-encoders offer superior precision. A bi-encoder can index an entire corpus offline and handle queries quickly using vector similarity search, making it ideal for large-scale systems and real-time applications. In contrast, a cross-encoder must evaluate each candidate document on the fly and is thus much slower and harder to scale, but it can discern subtle relevance signals (phrase matches, context nuances) that a bi-encoder might ignore. In practice, cross-encoders often significantly improve ranking quality – for example, in information retrieval benchmarks, re-ranking with a cross-encoder can substantially boost metrics like MRR (Mean Reciprocal Rank) or precision\@1 because the model can precisely identify which document truly answers the query.

To leverage the strengths of both, modern RAG pipelines commonly use a **two-stage retrieval** approach: an initial fast retrieval using a bi-encoder, followed by a re-ranking of the top results using a cross-encoder. In this hybrid setup, the bi-encoder quickly narrows down the candidate set from millions of documents to, say, the top *k=50* or *100* that are roughly relevant. Then the cross-encoder takes those candidates and scores each one with respect to the query, producing a refined ranking where the top *n* (e.g. 5) are much more likely to be truly relevant. This combination achieves a balance between recall (not missing any good documents) and precision (choosing the best answers).

&#x20;*Figure 2: **Two-stage retrieval with a bi-encoder and cross-encoder.** The query is first used to retrieve *top\_k* candidates from a vector database (or index) using bi-encoder embeddings (fast but coarse retrieval). Then a cross-encoder (re-ranker) evaluates each of those candidates along with the query to produce a refined relevance score, selecting the top *n* results (here \$k=25\$ candidates narrowed to \$n=3\$). This approach maximizes retrieval recall while still only passing a small number of highly relevant documents to the generative model, which is crucial given the limited context window of models and the cost of processing long inputs.*

**Benefits of Cross-Encoders in RAG:** By using a cross-encoder re-ranker, the RAG system can significantly improve the quality of the retrieved context before generation. The cross-encoder ensures that the final documents fed into the generator have a deeper semantic match to the query (not just superficial keyword overlap or embedding proximity). This often leads to more accurate answers in question answering tasks and more relevant documents in search tasks. Cross-encoders shine in **high-precision scenarios** like FAQ answering, technical Q\&A, or precise document retrieval where the nuances of language matter. For example, if a query is *“What did the user report about error code 503 in the log?”*, a bi-encoder might retrieve many documents about error codes, but the cross-encoder can home in on the one document where a user specifically mentioned “error code 503” in a log context, because it sees the whole query-document pair.

**Drawbacks and Mitigations:** The downside is the runtime cost. Each additional document passed to a cross-encoder increases latency linearly. Thus, one should keep \$k\$ (initial retrieved docs) and \$n\$ (re-ranked docs used) reasonably small, based on latency requirements. Typical systems might retrieve 50-100 with the bi-encoder, re-rank all of them, and then use the top 5 for generation. This yields a good balance where the cross-encoder’s expense is manageable. If the corpus is small (hundreds or a few thousands of documents), one could even use the cross-encoder as the primary retriever (scoring all documents or a large portion of them) – but for large corpora this is infeasible for real-time use. It’s also worth noting that cross-encoder models can be large (BERT-base or larger); using a smaller cross-encoder (such as a MiniLM or DistilBERT based model) can reduce cost at some accuracy loss. We will discuss performance optimizations in a later chapter.

### **1.3 Summary of Cross-Encoder vs Bi-Encoder Trade-offs**

To recap the differences and use-cases, below is a quick comparative summary:

| **Aspect**            | **Bi-Encoder (Dual)**                                                                                                                                                                      | **Cross-Encoder**                                                                                                                                                                                                  |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Encoding Approach** | Encode query and document *independently* into vectors.                                                                                                                                    | Encode query + document *jointly* in one forward pass.                                                                                                                                                             |
| **Output**            | Fixed-size embedding per text; similarity computed via dot product or cosine.                                                                                                              | Direct relevance score or classification (e.g. 0 to 1) for the pair.                                                                                                                                               |
| **Strengths**         | **Scalable & Fast**: Pre-compute embeddings; use ANN for million-scale retrieval in milliseconds. Captures general semantic similarity well.                                               | **High Accuracy**: Considers full context interactions; excels at nuanced relevance and ranking. Often improves precision in top results.                                                                          |
| **Weaknesses**        | **Lower Precision**: May miss context nuances (separate encoding loses some info). Quality of results depends on embedding quality.                                                        | **Slow & Costly**: Not scalable to large corpus by itself (must score each doc). Higher computational load per query (can’t pre-index).                                                                            |
| **Ideal Use Cases**   | **First-stage retrieval** for large collections (web search, corporate DB); any scenario requiring quick broad recall. Real-time applications where some accuracy trade-off is acceptable. | **Re-ranking** a small set of candidates for improved precision. Also suitable alone for small or static corpora where top accuracy is needed (e.g., a curated FAQ list). High-precision tasks like detailed Q\&A. |

In RAG pipelines, combining both methods is common: use the bi-encoder for efficient recall of candidates, then apply the cross-encoder for precise ranking. This way, you leverage the bi-encoder’s speed and the cross-encoder’s accuracy to achieve a balanced solution. The following chapters will explore how to implement such systems using Hugging Face tools and modern open-source libraries.

## **Chapter 2: Building a RAG Pipeline with Hugging Face Ecosystem**

Now that we understand the theory and motivation, let’s build a retrieval-augmented generation pipeline step by step using Python. We will use the **Hugging Face Transformers** library and other open-source tools such as **Sentence Transformers**, **Faiss**, and **PyTorch** to implement a working RAG system. The pipeline will incorporate a bi-encoder for initial retrieval, a cross-encoder for reranking, and a generator model for final answer generation.

### **2.1 Preparing the Data and Environment**

Before coding, ensure you have the required libraries installed. At minimum, you will need:

* **Transformers** (🤗 Hugging Face) – for using pre-trained models (both bi-encoder and cross-encoder, as well as generative models).
* **Datasets** (🤗 Hugging Face) – for handling the document corpus conveniently (optional but useful).
* **SentenceTransformers** – a library that wraps Transformers models for easier use in embedding and cross-encoder scenarios.
* **Faiss** (Facebook AI Similarity Search) – for efficient vector indexing and similarity search on the embeddings.
* **PyTorch** – the deep learning framework used by the above libraries.

Install these via pip if needed:

```bash
pip install transformers datasets sentence-transformers faiss-cpu torch
```

We assume you have a corpus of documents for retrieval. This could be a list of texts (e.g., paragraphs from articles or a company knowledge base). For demonstration, let’s assume `documents` is a Python list of text strings. In a real project, you might load documents from a database or files and perhaps chunk them into paragraphs if they are long (since retrieval works better on reasonably sized passages). Many RAG implementations split long documents (like full articles) into smaller chunks (\~100-300 words each) for more granular retrieval.

**Example Data Setup:** Suppose we are building a QA system on a set of technical articles. We might have:

```python
documents = [
    "Article1: ... (text of article 1) ...",
    "Article2: ... (text of article 2) ...",
    # ... more documents
]
```

For efficiency, pre-process and store these with their embeddings:

1. **Choose a Bi-Encoder Model:** Select a pre-trained bi-encoder to generate embeddings. Good choices include models from the SentenceTransformers library fine-tuned for semantic search or QA (e.g., `'sentence-transformers/multi-qa-MiniLM-L6-cos-v1'` which is a MiniLM model optimized for question<->answer similarity). Hugging Face also provides DPR encoders like `'facebook/dpr-ctx_encoder-single-nq-base'` (for Wikipedia passages) which can be used for dense retrieval. For our general example, we’ll use a MiniLM model for speed.

2. **Choose a Cross-Encoder Model:** Select a cross-encoder for reranking. A popular option is `'cross-encoder/ms-marco-MiniLM-L-6-v2'`, a MiniLM-based cross-encoder trained on the MS MARCO passage ranking dataset. This model inputs a question and passage and outputs a relevance score (roughly calibrated such that higher is more relevant). Other cross-encoders include larger models like `'cross-encoder/ms-marco-electra-base'` or even ones fine-tuned on specific domains if available. We prioritize smaller, faster cross-encoders for experimentation, with the possibility to swap in a larger model for better accuracy if needed.

3. **Choose a Generator Model:** Decide on a sequence-to-sequence model for answer generation. If using a Hugging Face RAG model, the generator is typically **BART** or **T5**. You can use `'facebook/bart-large-cnn'` (a BART model good for summarization) as a starting point or a T5 model like `'t5-base'`. There are also RAG-specific models like `'facebook/rag-token-base'` which bundle a question encoder, retriever, and generator; however, those come pre-configured with DPR (bi-encoder) and are less flexible for integrating a custom cross-encoder. Instead, we will assemble our own pipeline. For our example, we might use a pre-trained BART or T5 and assume it’s been fine-tuned for QA style generation (if not, it may still work, but fine-tuning on a QA dataset would help—covered in Chapter 3).

With this setup, let's instantiate the models in code and build the retrieval components:

```python
from sentence_transformers import SentenceTransformer, CrossEncoder, util
from transformers import AutoTokenizer, AutoModelForSeq2SeqLM
import faiss
import torch

# 1. Bi-encoder model for embeddings (using SentenceTransformers for convenience)
bi_encoder = SentenceTransformer('sentence-transformers/multi-qa-MiniLM-L6-cos-v1')
bi_encoder.max_seq_length = 256  # Limit length to 256 tokens per doc (for efficiency)

# Compute embeddings for all documents
doc_embeddings = bi_encoder.encode(documents, convert_to_tensor=True)
# (Optionally, move embeddings to FAISS index for fast similarity search)
dimension = doc_embeddings.shape[1]
index = faiss.IndexFlatIP(dimension)  # Inner product (cosine similarity if vectors are normalized)
index.add(doc_embeddings.cpu().numpy())

# 2. Cross-encoder model for reranking
cross_encoder = CrossEncoder('cross-encoder/ms-marco-MiniLM-L-6-v2')  # outputs relevance scores

# 3. Generator model (seq2seq LM) and tokenizer
gen_model_name = 'facebook/bart-large-cnn'
gen_tokenizer = AutoTokenizer.from_pretrained(gen_model_name)
gen_model = AutoModelForSeq2SeqLM.from_pretrained(gen_model_name)
```

Let’s break down what we did:

* We loaded a **bi-encoder** and encoded all `documents` into vectors. We also set up a FAISS index (`IndexFlatIP`) for these vectors, which will allow similarity search by inner product. (Our chosen model outputs normalized embeddings suited for cosine similarity; using inner product on normalized vectors is equivalent to cosine similarity.)
* We loaded a **cross-encoder** using `CrossEncoder` from SentenceTransformers, which wraps a Hugging Face model under the hood and provides a convenient `.predict()` method to score pairs.
* We loaded a **generator** (BART) with its tokenizer. We will use this to generate answers given a query and retrieved text.

### **2.2 Retrieval and Re-Ranking Workflow**

With models in place, the RAG pipeline proceeds in stages for each input query. The typical flow is:

1. **Encode Query & Initial Retrieval:** Encode the incoming query using the bi-encoder to get a query embedding. Use the FAISS index to find the top-\$k\$ most similar document vectors (this yields indices of the top documents). This step gives us a set of candidate documents quickly.
2. **Re-Rank with Cross-Encoder:** For each candidate document, form a pair (query, document text) and feed it to the cross-encoder to get a relevance score. Sort the candidates by this score to get a re-ranked list.
3. **Select Top-\$n\$ Documents:** Take the top few documents after re-ranking (e.g., top 3 or 5) that will serve as the context for generation.
4. **Generate Answer:** Concatenate the query with the content of the top documents (or otherwise structure the input) and pass this into the generative model to produce the final answer sentence/paragraph.

We can implement these steps in code as a function:

```python
def answer_query(query, top_k=20, top_n=5):
    # 1. Encode query and retrieve top_k candidates via vector search
    query_embedding = bi_encoder.encode(query, convert_to_tensor=True)
    # Use FAISS index search
    _, retrieved_indices = index.search(query_embedding.cpu().numpy(), top_k)
    retrieved_indices = retrieved_indices.flatten().tolist()
    candidate_docs = [documents[i] for i in retrieved_indices]
    
    # 2. Re-rank candidates with cross-encoder
    cross_inputs = [[query, doc] for doc in candidate_docs]
    scores = cross_encoder.predict(cross_inputs)  # higher score = more relevant
    # Pair each doc with its score, then sort by score descending
    ranked_pairs = sorted(zip(candidate_docs, scores), key=lambda x: x[1], reverse=True)
    
    # 3. Select top_n after re-ranking
    top_docs = [doc for doc, score in ranked_pairs[:top_n]]
    
    # 4. Generate answer using the top_n docs as context
    # Here we simply concatenate the top docs with the query. In practice, you might format this better.
    context = " ".join(top_docs)
    input_text = query + " </s> " + context  # using </s> as a separator for BART
    inputs = gen_tokenizer([input_text], max_length=512, return_tensors='pt', truncation=True)
    output_ids = gen_model.generate(**inputs, max_new_tokens=64)
    answer = gen_tokenizer.decode(output_ids[0], skip_special_tokens=True)
    return answer, top_docs
```

In this function:

* We retrieve `top_k` documents using the **FAISS** index. Typically, you set `top_k` larger than the number of documents you ultimately want to use, to give the cross-encoder a broad set to evaluate (e.g., retrieve 20, then later use the best 5).
* We then score each candidate with the cross-encoder. The `CrossEncoder.predict` method handles encoding the pairs and outputting a score for each.
* We sort and take `top_n` documents. Those will be provided to the generator.
* We build an input for the generator by concatenating the query and the retrieved text. Depending on the format the generator expects, you might include special separators or a prompt template. Here, we simply separate the query and context with a special end-of-sentence token for BART. We limit the input length to avoid exceeding the model’s context size (512 tokens for BART-large).
* Finally, we run the generator’s `generate` to produce an answer. We limited it to `max_new_tokens=64` for the answer length (adjust as needed). We decode the result to get a string.

Let’s test this pipeline with a dummy example (assuming `documents` were populated):

```python
query = "What is retrieval-augmented generation?"
answer, supporting_docs = answer_query(query, top_k=10, top_n=3)
print("Q:", query)
print("A:", answer)
print("Supporting docs used:", supporting_docs)
```

The output might be something like:

```
Q: What is retrieval-augmented generation?
A: Retrieval-Augmented Generation (RAG) is a technique that combines a generative model with an external knowledge source, allowing the model to retrieve relevant documents and use them to produce more informed and accurate responses.
Supporting docs used: ['... definition of RAG ...', '... another doc mentioning RAG ...', '...']
```

*(The actual answer will depend on the content of the documents. This is an illustrative example.)*

This pipeline shows the basic implementation using Hugging Face tools. Notably, we used `SentenceTransformer` and `CrossEncoder` from the SentenceTransformers library for convenience, which internally uses Hugging Face models (you could alternatively use the lower-level Hugging Face `AutoModel` classes directly, but SentenceTransformers simplifies the encoding and similarity search steps).

### **2.3 Integration with Hugging Face RAG and Other Tools**

Hugging Face’s Transformers library actually provides an integrated RAG model and retriever, which can simplify some parts of the above. For example, `RagRetriever` can be used to load or build a Faiss index and `RagTokenForGeneration` or `RagSequenceForGeneration` can handle the retrieve-and-generate loop internally. However, those classes are geared towards using DPR (bi-encoder) and don’t natively support a cross-encoder in the loop. In our approach, we manually inserted a cross-encoder reranker.

It’s worth noting that in an end-to-end RAG model like `'facebook/rag-token-nq'`, the sequence generator (BART) marginalizes over multiple retrieved documents to produce the answer (scoring tokens across docs) as described in the original RAG paper. Our simple pipeline doesn’t marginalize probabilities over documents; instead, we concatenated top docs and let the generator treat it as one input. An alternative approach is to generate an answer from each top document separately and then choose the best answer, but that gets complex and usually requires a trained model. For simplicity, providing all top contexts to a single forward pass of a generator works well for many scenarios (assuming the generator can handle the combined context length).

**Using Hugging Face Datasets and Faiss:** If your document corpus is large, consider using the `datasets` library to store and index it. You can create a `Dataset` with your documents and use `dataset.add_faiss_index` to build a Faiss index of embeddings for efficient retrieval. This can then be queried via `dataset.get_nearest_examples` when a query embedding is given. Using a dataset also makes it easy to store additional fields like document titles or IDs alongside the text.

**Hybrid Retrieval (Lexical + Semantic):** In some cases you might combine **lexical search** (keyword BM25) with **semantic search** (bi-encoder). Tools like **Elasticsearch** or **Whoosh** can retrieve by keywords, which sometimes finds exact matches that embedding search might miss (especially for rare proper nouns or codes). A hybrid approach is beyond our scope here, but keep in mind it’s possible: e.g., retrieve some docs by BM25, some by embeddings, merge the sets, then apply the cross-encoder to all. This can increase recall for certain queries.

By the end of this chapter, you should have a functional pipeline that given a query can retrieve relevant info and generate an answer. Next, we will look at how to improve and customize this system via fine-tuning the models on domain-specific data.

## **Chapter 3: Fine-Tuning Bi-Encoders and Cross-Encoders for RAG**

While pre-trained models give a decent starting point, the best performance often comes from fine-tuning the retriever and reranker on data specific to your task or domain. In this chapter, we discuss strategies to fine-tune the bi-encoder and cross-encoder, as well as the generator, to improve RAG system performance. We’ll also cover how to create training data for these models, even if you don’t have it initially (through *synthetic data generation*).

### **3.1 Fine-Tuning the Bi-Encoder (Retriever)**

**Objective:** The bi-encoder retriever is typically fine-tuned to produce embeddings such that relevant question-document pairs have higher similarity than irrelevant pairs. A common training approach is **contrastive learning**: given a query and a set of documents (one relevant, others not), train the encoder such that the query’s embedding is closer to the relevant doc’s embedding than to the others. Datasets like MS MARCO (for passage retrieval) or Natural Questions are commonly used for such training, where queries are user questions and positives are passages containing the answer.

If you have a dataset of query-document pairs (or question-answer pairs with the answer text as the document), you can fine-tune using the SentenceTransformers library or Hugging Face’s `Trainer`. SentenceTransformers offers losses like `MultipleNegativesRankingLoss` which are very effective for this scenario.

**Example:** Fine-tuning with SentenceTransformers:

```python
from sentence_transformers import losses, InputExample, SentenceTransformer

# Suppose train_data is a list of (query, positive_doc) pairs for training
train_examples = [InputExample(texts=[q, d]) for q, d in train_data]

train_dataloader = DataLoader(train_examples, shuffle=True, batch_size=16)
train_loss = losses.MultipleNegativesRankingLoss(bi_encoder)

# Fine-tune bi_encoder model
bi_encoder.fit(
    train_objectives=[(train_dataloader, train_loss)],
    epochs=1,
    optimizer_params={'lr': 2e-5}
)
```

This will adjust the bi-encoder’s embedding space for your specific data. After fine-tuning, you’d rebuild the document embeddings/index with the updated model. Fine-tuning can significantly improve retrieval recall – the model learns terminology and content style of your corpus (for example, if your domain is medical text, fine-tuning on medical Q\&A will make the embeddings much more sensitive to medical terms). If labeled data is scarce, consider using pre-trained models as is, or use weak supervision (e.g., assume that documents written by the same author on a topic are related, etc.) to create positive pairs.

Another approach: **Synthetic Data**. If you lack training pairs, you can generate them. For instance, given each document, use an LLM to generate a plausible question that the document answers. This gives a (question, document) pair for training. This method was demonstrated in a Hugging Face blog where they generated synthetic QA pairs from documents to fine-tune RAG retrievers. It’s a clever way to bootstrap training data: the generator “invents” questions for which the document would be useful, and then the retriever is fine-tuned on those.

After fine-tuning, always validate the retriever’s performance on some dev set if possible (e.g., check Recall\@K: the percentage of queries for which a relevant doc is in the top K results). A strong retriever is the backbone of a RAG system – if it fails to retrieve the necessary information, the generator cannot possibly produce a correct answer (garbage in, garbage out!). Studies have shown that improvements in retrieval quality directly improve final answer accuracy.

### **3.2 Fine-Tuning the Cross-Encoder (Re-Ranker)**

The cross-encoder can also be fine-tuned, typically formulated as a **pairwise ranking** or **regression** problem. If you have labeled data indicating which document is more relevant to a query, you can fine-tune the cross-encoder to predict higher scores for relevant pairs. Commonly, cross-encoders for retrieval are trained on datasets like MS MARCO where queries have human-annotated relevant passages. The model (often a RoBERTa or MiniLM) is trained to output a score (sometimes via a sigmoid layer to produce a probability) that correlates with relevance. This can be done as a binary classification (relevant vs not relevant) or as a regression (predicting a relevance grade).

Using SentenceTransformers’ `CrossEncoder`, you can fine-tune with a simple regression objective. For example:

```python
from sentence_transformers import CrossEncoder

# Assume cross_train_data is a list of (query, doc, label) tuples, label=1 for relevant, 0 for irrelevant
cross_encoder_model = CrossEncoder('cross-encoder/ms-marco-MiniLM-L-6-v2', num_labels=1)  # 1 output (regression)
train_dataset = list(zip([q for q, d, l in cross_train_data],
                          [d for q, d, l in cross_train_data],
                          [float(l) for q, d, l in cross_train_data]))
# Note: CrossEncoder expects a list of (sentence1, sentence2, label) tuples for training
cross_encoder_model.fit(train_dataset,
                        epochs=2,
                        batch_size=16,
                        optimizer_params={'lr': 2e-5})
```

This will fine-tune the cross-encoder so that it learns to output higher scores for query-doc pairs labeled as relevant. If you have pairwise data (like query, doc1, doc2 where doc1 is more relevant than doc2), you could also train with a ranking loss (CrossEncoder can accept such data by duplicating queries with different docs and appropriate labels).

Hugging Face’s Trainer can also be used by formulating the task as a sequence classification problem: feed the query and doc through a model like `BERT` with a classification head to output a single logit, and train with a regression or binary cross entropy loss. The SentenceTransformers approach is essentially doing that under the hood.

After fine-tuning, your cross-encoder will be specialized to your specific documents and query style, which can boost re-ranking effectiveness. For instance, if your system is for legal documents Q\&A, fine-tuning on some example legal QA pairs will teach the cross-encoder to pay attention to legal terms and domain-specific language.

**Training Time Considerations:** Cross-encoders are slower to train than bi-encoders per sample (since the model processes both texts at once). However, you typically need fewer pairs to achieve gains – because even a small number of labeled relevant vs irrelevant examples can help the model adjust to your domain’s vocabulary and question style. Monitor the model’s performance on a validation set of query-doc pairs (if available) to avoid overfitting.

A fine-tuned cross-encoder can substantially improve ranking. For example, with a good cross-encoder reranker, it’s common to achieve large jumps in precision\@1 or success rate in retrieving the exact correct passage for a query. This directly translates to better answers from the generator.

### **3.3 Fine-Tuning the Generator (Optional)**

The generative model (like BART or T5) can also be fine-tuned on your QA task, especially if you have examples of questions and ideal answers (with context). If you have a dataset of query and answer pairs (with documents), you can fine-tune the Seq2Seq model to generate the answer given the concatenated retrieved documents plus question as input. This is essentially a supervised training of the RAG pipeline end-to-end. For instance, using the `Seq2SeqTrainer` in Transformers with a dataset that provides input texts (question + retrieved context) and target text (answer) can refine the generator to better utilize retrieved info.

However, fine-tuning the generator requires ground-truth answers and is a bit more involved (and computationally heavier if using a large model). Many times, using a pre-trained generator (especially if it’s something like GPT-3 or other large LMs via API) is sufficient and focus is placed on the retrieval components.

In summary, fine-tuning is a powerful lever to adapt RAG components:

* Fine-tuned **bi-encoders** improve recall of relevant info.
* Fine-tuned **cross-encoders** improve precision of ranking.
* Fine-tuned **generators** improve the quality and correctness of the final answer (ensuring the model properly uses the context).

If data is limited, prioritize fine-tuning the retriever and reranker, as they directly affect whether the necessary information is available to the generator. There’s an adage in RAG: *“75% retrieval, 25% generation”*, emphasizing that a lot of the heavy lifting in answer quality comes from how well you retrieve. With fine-tuned retrieval components, even a generic generator can perform well by having the right facts at hand.

## **Chapter 4: Experimentation, Evaluation, and Best Practices**

Building a RAG system is an iterative process. In this chapter, we cover best practices for examining and interpreting the outputs of your cross-encoder based RAG system, as well as methodologies to experiment and evaluate performance. We assume by now you have a pipeline (from Chapter 2) and possibly fine-tuned components (from Chapter 3). Now you need to **validate and refine** it.

### **4.1 Interpreting Cross-Encoder Scores and Outputs**

The cross-encoder provides a relevance score for each query-document pair. Often, these scores are unbounded or on an arbitrary scale (depending on the model’s final layer). For example, a cross-encoder like `ms-marco-MiniLM-L-6-v2` outputs a single logit which isn’t a probability but can be any real number; higher means more relevant. If it was trained with a sigmoid, it might output roughly 0 to 1. It’s useful to understand these scores in relative terms:

* They are primarily for **ranking**. The absolute value isn’t as important as the ordering. A score of 8.0 vs 2.0 indicates the first document is considered more relevant to the query than the second. Some cross-encoders (like ones trained with softmax classification across multiple docs) might output something akin to a logit or probability of relevance.
* You might observe that scores for different queries aren’t directly comparable. That is fine – you compare scores only among candidates for the *same query* to rank them. If needed, you can calibrate or normalize scores (for instance, some use a softmax across scores of candidates to interpret them as relative probabilities).

When analyzing outputs, it’s helpful to print out not just the final answer but also the intermediate steps:

* **Check retrieved documents:** Make sure the bi-encoder is retrieving reasonably relevant documents. If many retrieved docs are obviously irrelevant, your bi-encoder (or the index) might need improvement (or you might need to increase `top_k` so that some good ones are included).
* **Check cross-encoder ranking:** See which documents got the highest scores. Often a cross-encoder can surface a document that was, say, the bi-encoder’s 10th candidate to now be the top 1 because of a subtle match (this is a good sign, showing the reranker is adding value). If the cross-encoder’s top choice seems odd or not actually relevant, examine why. It could be an indication of an issue (maybe the cross-encoder model isn’t well-tuned for your data yet).
* **Score distribution:** Sometimes cross-encoder scores for the top few docs are very close, which means they were similarly relevant. If scores drop off sharply after the first few, it indicates a clear distinction between relevant and irrelevant for that query. This can inform how many docs you ultimately want to feed to the generator. For instance, if only the top 1-2 have high scores and the rest are low, using top 5 might be overkill (the bottom 3 might add noise).

**Thresholding and Filtering:** In some applications, you may decide not to produce an answer if no document achieves a certain relevance score (i.e., the system says “I don’t know” if it’s not confident). A cross-encoder can facilitate this by providing a score that correlates with relevance. You could set a threshold (learned from data or manual inspection) and if no retrieved passage exceeds it, choose to return a fallback response (“Sorry, I couldn’t find an answer.”). This kind of confidence estimation is important in production if you want to avoid giving incorrect answers when the system is unsure.

### **4.2 Evaluation Metrics for RAG Systems**

To quantitatively evaluate your RAG system, consider both the retrieval component and the final QA output:

* **Retrieval Evaluation:** Use information retrieval metrics. If you have a test set of queries with known relevant documents (or answers with known source), measure **Recall\@K** – e.g., does the correct document appear in the top 5 or 10 retrieved? Also, **Mean Reciprocal Rank (MRR)** or **NDCG** (Normalized Discounted Cumulative Gain) are used to evaluate ranking quality, especially if you have graded relevance. A high recall at a modest \$K\$ (like Recall\@5 or @10) is critical since the generator can only see so many docs. If recall is low, the system will often miss answering correctly. Evaluate the bi-encoder retrieval and cross-encoder re-ranked retrieval separately: for example, check Recall\@100 of the bi-encoder to ensure the relevant doc is initially fetched, and Recall\@5 after re-ranking to see if cross-encoder improved the placement.

* **Generation Evaluation:** If your task is question answering with ground-truth answers, you can use metrics like **Exact Match** (does the output match the gold answer exactly) or **F1 score** (overlap of answer tokens) for factual QA. For more open-ended generation or if answers can be phrased differently, metrics like BLEU or ROUGE could be used, though they have limitations. Human evaluation is often needed to fully assess answer quality. Another angle is to evaluate how often the generated answer contains a correct reference: if your system is supposed to provide a source citation from the retrieved docs, check if the cited doc truly supports the answer (this is more advanced and often done manually or with additional models).

* **Latency and Throughput:** This is a different kind of evaluation but important for deployment. Measure how long a query takes end-to-end. Cross-encoder re-ranking will typically be the slowest part (aside from the generator if you use a very large model). If each query is taking too long, you might need to adjust \$k\$ or use a smaller model to meet requirements. We’ll discuss optimization in Chapter 5.

**Experimentation Tips:** When experimenting with your RAG system, try varying some parameters to see their effect:

* The number of documents fed to the generator (`top_n`). Sometimes using 3 vs 5 context documents can change the answer or its quality. Too many may introduce irrelevant info that confuses the generator.
* The cross-encoder model itself: a larger cross-encoder (like a cross-encoder based on BERT-large) might yield slightly better ranking, at the cost of speed. If accuracy is paramount, you could test a bigger model offline to see potential gains.
* Use ablations: try turning off the cross-encoder (just use bi-encoder retrieval) and see how the final answers suffer. This can quantify how much the reranker is helping. Typically, you’ll observe the cross-encoder significantly improves precision so the answers are more often correct or precise rather than slightly off-topic.
* If possible, compare to a baseline like BM25 retrieval. This might be relevant if your corpus has a lot of specific jargon; sometimes a tuned BM25 can perform surprisingly well on certain datasets, and comparing helps validate that your dense retriever is indeed adding value.

**Analyzing Failures:** For queries where the system gives a wrong or unsatisfactory answer, trace through:

1. Did the retriever fetch the needed document? If not, that’s a retrieval miss (consider fine-tuning or adding data).
2. If it fetched it but cross-encoder didn’t rank it high, why? Maybe the cross-encoder wasn’t trained to recognize that relevance. This could point to needing more training data for such cases.
3. If the right doc was given to the generator but the answer is still wrong, perhaps the generator couldn’t parse the info correctly (maybe the context was too verbose, or it’s a tricky question requiring reasoning). In such cases, you might improve the prompt format or consider an alternate strategy like splitting the question.
4. Check if the generator hallucinated. Sometimes the generator might include facts not present in retrieved docs. This is a known issue: even with relevant docs, a language model might inject prior knowledge or guesses. To mitigate, you can try making the prompt more direct (“Using the following document, answer...”) or fine-tune the generator on QA tasks so it learns to copy from context.

**Logging and Iteration:** It’s good practice to log queries, retrieved doc IDs, cross-encoder scores, and final answers during testing. This log can be analyzed to find patterns, e.g., certain types of questions always fail or certain documents frequently appear as top but are actually not useful (could indicate a semantic embedding issue). Use this to iteratively refine your models or pipeline. For example, if you notice the cross-encoder is giving high scores to very long documents that just happen to share one keyword, you might implement a length penalty or ensure the cross-encoder sees a snippet rather than full text.

In summary, treat your RAG system as an evolving piece of software: test it with sample questions, measure retrieval quality and answer quality, identify weaknesses, and address them via better models or preprocessing. RAG systems involve multiple components, so careful evaluation of each part and the whole is key to building a reliable solution.

## **Chapter 5: Deployment and Scaling Strategies**

After developing a working RAG pipeline and refining it through experiments, the next step is deploying it in a production environment. This chapter discusses how to build a robust **REST API** for your RAG system, strategies for optimizing performance, and using vector stores in production. We assume by now your pipeline code is functioning and you want to expose it as a service for real users or integrate it into an application.

### **5.1 Building a RAG API Service**

A common way to serve an ML model (or pipeline of models) is to wrap it in a REST API. Users can send requests (with a query) to the API and get back a response (the generated answer, plus maybe the retrieved passages or sources).

**Framework:** You can use a lightweight web framework like **FastAPI** (in Python) to build the service. FastAPI is convenient for serving ML models because it’s fast and allows asynchronous endpoints, which can be useful if each request involves waiting for model processing (you can serve multiple requests concurrently with async).

**Example with FastAPI:**

```python
from fastapi import FastAPI
import uvicorn

app = FastAPI()

# Load models globally when the app starts
bi_encoder = SentenceTransformer('sentence-transformers/multi-qa-MiniLM-L6-cos-v1')
doc_embeddings = bi_encoder.encode(documents, convert_to_tensor=True)
dimension = doc_embeddings.shape[1]
index = faiss.IndexFlatIP(dimension)
index.add(doc_embeddings.cpu().numpy())
cross_encoder = CrossEncoder('cross-encoder/ms-marco-MiniLM-L-6-v2')
gen_tokenizer = AutoTokenizer.from_pretrained(gen_model_name)
gen_model = AutoModelForSeq2SeqLM.from_pretrained(gen_model_name)

@app.get("/answer")
def answer_query_api(q: str, top_k: int = 20, top_n: int = 5):
    answer, supporting_docs = answer_query(q, top_k=top_k, top_n=top_n)
    return {"query": q, "answer": answer, "supporting_docs": supporting_docs}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

In this simplified example:

* We initialize the models and index outside of the request function so that they are loaded into memory once (at startup) and reused for every query. This is crucial for performance – you do *not* want to reload models on each request.
* The `/answer` endpoint takes a query `q` and optional parameters for top\_k and top\_n. It uses our earlier `answer_query` function (from Chapter 2) to get the answer and supporting docs.
* The result is returned as JSON. We include the original query, the answer, and possibly the supporting docs (which could be full text or maybe just titles or IDs, depending on what you want clients to see).

In a real deployment, you might not return full document texts due to size; you could return just document IDs or snippet excerpts as evidence.

**Concurrency:** If using FastAPI with `uvicorn`, you can configure it to use multiple workers (processes) to handle multiple requests in parallel. Each worker would load the models, so be mindful of memory (loading a large model multiple times can be heavy – ensure your server has enough RAM or consider using a single process with async if the library supports it; however, heavy compute in Python often benefits from multi-process to bypass the GIL and utilize multiple CPU cores).

### **5.2 Performance Optimization**

Performance for a RAG system has a few bottlenecks: the vector search, the cross-encoder, and the generator. For each, consider:

* **Vector Search (Retriever):** FAISS is very fast (sub-millisecond for thousands of docs, and can handle millions with optimized indexes). If your corpus is huge, use Faiss indexes like IVF (inverted file) or HNSW for sub-linear search. These require training the index (for IVF) but can drastically speed up search with minimal loss in recall. Also, use **GPU Faiss** if available for even faster searches on large collections. If using an external vector DB (like Pinecone, Weaviate, Milvus), those services are optimized for scale – but there is a network overhead to calling them. For very low latency, having the index in-memory (as in our example) is fastest, but that might not be feasible if the dataset is gigantic or needs frequent updates.

* **Cross-Encoder Reranking:** This is often the slowest step, as it involves running a BERT-like model on *k* pairs. If each query re-ranks 100 candidates and your cross-encoder is a 6-layer MiniLM, it might be okay; but if it’s a 12-layer or larger, that’s 100 \* 12-layer inferences, which could be hundreds of milliseconds on CPU. **Optimize this by**:

  * Running the cross-encoder on a **GPU**. Transformers models speed up significantly on GPU. A single GPU can handle multiple queries’ re-ranking in parallel if you batch them. You might process candidates in batches (e.g., 25 at a time) to utilize vectorized compute.
  * Reducing *k*: maybe retrieving 50 instead of 100 if it doesn’t hurt recall too much.
  * Using a smaller cross-encoder or distilling a larger one. For instance, MiniLM or distilled BERT models can be 2-3x faster than BERT-base.
  * **Quantization**: Using 8-bit or 4-bit quantization for the cross-encoder model can speed it up with minor accuracy loss. Libraries like `transformers` support `model.quantize` or you can use ONNX Runtime with quantized kernels.
  * **Caching**: If certain queries repeat or are similar, caching might help, but for an open QA system cache hits are not very common unless queries are often identical.
  * **Multi-threading**: PyTorch by default uses multiple threads for CPU inference. You can experiment with `torch.set_num_threads()` if needed. On GPU, ensure to feed enough data to saturate it (i.e., batch the cross-encoder inputs).

* **Generator:** If using a large generator model (like a big T5 or a custom transformer), generation can be slow, especially if the output is long because it’s an auto-regressive process. Options:

  * Use smaller or optimized models. For instance, if you fine-tuned `t5-base` and it works, that’s faster than `t5-3B`. There are also distilled generators.
  * Limit the maximum answer length to what’s necessary.
  * If you have an option to use an API (like OpenAI GPT-3) vs local, weigh the latency and cost – a local model avoids network calls but might be slower if it’s not as well optimized. However, since we focus on open-source, assume local deployment.
  * Batch generation is possible if you get multiple queries at once, but if this is an interactive system, that’s usually not applicable.

* **System Architecture:** In a high-throughput setting, you might break the pipeline into microservices: one service for vector search, one for reranking, one for generation. This allows scaling each independently (e.g., multiple reranker instances on GPU). They would pass data via RPC. However, this adds complexity and only makes sense if each component itself is heavy and you want to scale horizontally. In many cases, a single service with sufficient hardware (GPU for cross-encoder and generator, CPU for Faiss or also GPU Faiss) can serve quite a lot of queries per second.

**Monitoring:** Deploy with logging enabled. Monitor response times and memory usage. If using GPU, monitor GPU utilization – ensure it’s not underutilized (which could indicate you should batch or increase concurrency) or overutilized (indicating potential saturation and need for additional GPU or smaller models).

### **5.3 Using Vector Stores in Production**

Vector stores (or vector databases) are specialized systems to handle vector data and similarity search, often with features like persistence, replication, and APIs. Examples include **Pinecone**, **Weaviate**, **Milvus**, **Vespa**, and even Elasticsearch’s vector capabilities.

Using such a store can simplify certain aspects:

* They manage the indexing and search for you, often scaling to billions of vectors.
* They allow dynamic updates (insert/delete vectors) which pure Faiss in-memory might not easily support without rebuilding.
* Some offer hybrid search (combine vector similarity with keyword filters).

In a RAG deployment, you can use a vector store as follows:

* During ingestion, send each document’s embedding (computed by your bi-encoder) to the store, with an ID and maybe metadata (like document title or other fields).
* At query time, send the query embedding to the store via its client (many have Python client libraries). Get back the top k IDs (and sometimes the store can also return the stored text or a reference to it).
* Fetch those documents (if the store doesn’t return full text, you might store text in a separate DB keyed by ID, or in the vector store metadata).
* Then perform cross-encoder reranking as usual.

One must consider the network overhead: calling an external service (even if self-hosted) adds latency. If the vector DB is on the same local network or machine, it can be quite fast (some vector DBs like Vespa you might run co-located with the app). If using a cloud service like Pinecone, the round-trip might add, say, 20-50ms. This might be acceptable for many apps, given the benefits.

**Compatibility with Cross-Encoder:** The vector store doesn’t directly interact with the cross-encoder; it’s simply providing the initial candidates. So any vector store is “compatible” – you just use it instead of Faiss in our pipeline. The cross-encoder step remains the same: you’ll take the retrieved documents (via IDs) and run the model locally to score them.

**Security and Privacy:** If your application data is sensitive, ensure whatever solution (self-hosted or cloud) meets your privacy needs. Self-hosting something like Milvus or Weaviate gives you full control over data. These open-source stores can be integrated similarly to Faiss (they have Python APIs or REST).

**Batching Requests:** In a multi-user scenario, you might get many queries per second. You can batch multiple queries to the vector store (some support searching multiple query vectors in one call). You can also batch cross-encoder inference across queries if you have leftover GPU capacity. This introduces some complexity (needs asynchronous request handling to group queries), but can improve throughput.

Finally, ensure you have a strategy for keeping the document index updated. If new documents are added frequently, you’ll want to periodically update the index or send the new vectors to the store (some stores do this in real-time). If documents change or get deleted, handle that in the store as well, to avoid retrieving outdated info.

### **5.4 Additional Deployment Tips**

* **Scaling horizontally:** If one instance of your service isn’t enough, you can run multiple instances (each with its own copy of models loaded) behind a load balancer. This is straightforward but memory intensive. If models are large, this may not scale linearly (you might prefer a multi-threaded single instance on a beefy machine versus many smaller instances).
* **Model serving solutions:** Consider using optimized model servers like **TorchServe** or **Triton Inference Server** for the models. These allow deployment of PyTorch or ONNX models with high performance and even auto-batching. However, integrating a pipeline across multiple models might require some custom logic.
* **Latency vs Throughput:** Decide what’s more critical. If low latency for single queries is the goal (e.g., an interactive chatbot), optimize for that (maybe sacrificing some throughput by using a bigger GPU just for one process). If throughput (queries per second) is more important (e.g., batch processing many queries or a high-traffic web app), you might tolerate slightly higher p95 latency to achieve more parallelism.
* **API Design:** The API could be extended to return not just the answer but also some confidence or source. For example, returning the titles or IDs of the documents used, so that a frontend could display “Answer (sourced from Document XYZ)”. This can improve user trust. It’s common in RAG-based apps (like Bing’s AI chat or others) to cite sources.
* **Error handling:** Make sure to handle cases like no results found (vector store returns nothing), or generation failure. Also, put timeouts – if the generator is taking too long, you might abort and return a message rather than make the user wait indefinitely.

By following these deployment strategies, you can bring your RAG system from a research environment into a production-grade service that is robust and efficient.

## **Chapter 6: Practical Examples and Case Studies**

In this final chapter, we walk through two end-to-end examples of RAG systems using cross-encoders: one for **open-domain question answering** and one for **document retrieval**. These examples will illustrate how all the pieces come together in real scenarios, with code snippets and explanations for each.

### **6.1 Case Study: Question Answering Assistant with RAG**

Imagine we want to build a QA assistant that answers questions about a collection of company policy documents. Users will ask in natural language, and the system should return a clear answer along with reference to the policy document.

**Steps Outline:**

1. **Data Preparation:** We have a set of policy documents (say in PDF or text). We split them into reasonably sized paragraphs and store them (with an ID and perhaps the document name). For example, after splitting, we might have `documents = [(id1, "paragraph text..."), (id2, "paragraph text..."), ...]`.
2. **Indexing:** Use a bi-encoder to embed all paragraphs and build a vector index (Faiss or a vector DB).
3. **Query Handling:** For each user question:

   * Retrieve top-k paragraphs via the bi-encoder+index.
   * Re-rank those paragraphs with the cross-encoder to get the most relevant ones.
   * Feed the top-n paragraphs to a generative model which will compose an answer.
4. **Answer Construction:** The model generates an answer which we may post-process. We also can retrieve the document name of the top paragraph to cite the source.

We’ll demonstrate a simplified version with code:

```python
# Assume documents is a list of (doc_id, text) pairs
doc_texts = [text for _, text in documents]

# Build index (as done before)
doc_embeddings = bi_encoder.encode(doc_texts, convert_to_tensor=True)
index.add(doc_embeddings.cpu().numpy())

def answer_question(question):
    # Retrieve
    q_embed = bi_encoder.encode(question, convert_to_tensor=True)
    _, indices = index.search(q_embed.cpu().numpy(), 20)
    indices = indices.flatten().tolist()
    candidates = [documents[i] for i in indices]  # list of (doc_id, text)
    # Re-rank
    cross_inputs = [[question, text] for (_, text) in candidates]
    scores = cross_encoder.predict(cross_inputs)
    ranked = sorted(zip(candidates, scores), key=lambda x: x[1], reverse=True)
    top_docs = ranked[:3]
    # Prepare generation input: include document names to help model possibly cite
    context = ""
    for (doc_id, text), score in top_docs:
        context += f"[{doc_id}] {text}\n"  # prepend an identifier for each doc
    input_text = question + "\n" + context
    inputs = gen_tokenizer([input_text], max_length=512, return_tensors='pt', truncation=True)
    summary_ids = gen_model.generate(**inputs, max_new_tokens=100)
    answer = gen_tokenizer.decode(summary_ids[0], skip_special_tokens=True)
    return answer, top_docs
```

This function will output an answer and the top supporting documents with their IDs. We format each document’s text with an ID like `[doc123]` so the model sees some source indicator; in practice, you might instruct the model via the prompt to use them. For instance, if using an instruction-tuned model, you might prompt: *"Use the following documents to answer. Document references are in brackets. Question: ... Documents: \[Doc1] ... \[Doc2] ... Answer:"*. This can encourage the model to include the references in the answer.

**Example run:**

```python
user_q = "What is the policy on parental leave for fathers?"
answer, top_docs = answer_question(user_q)
print("Q:", user_q)
print("A:", answer)
print("Used docs:", [doc_id for (doc_id, _), _score in top_docs])
```

Sample output:

```
Q: What is the policy on parental leave for fathers?
A: According to the company policy, fathers are entitled to parental leave of up to 12 weeks. This leave can be taken within the first year of the child's birth or adoption. [PolicyDoc1]
Used docs: ['PolicyDoc1', 'PolicyDoc3', 'PolicyDoc1']
```

*(The answer cites \[PolicyDoc1] as the source. We listed top\_docs; PolicyDoc1 appeared twice which might indicate two different paragraphs from the same document were in top 3.)*

This demonstrates an open-domain QA style RAG: the user gets a direct answer. The answer includes a reference in this case (we assumed the model learned to include the \[PolicyDoc1] reference; if not, one could add a post-processing step to append "Source: PolicyDoc1" manually using the known top doc).

Key points in this example:

* The cross-encoder helped pick the right paragraph about parental leave specifically for fathers (maybe a generic “parental leave” doc plus one that specifically mentions fathers). Without it, the top result might have been a more generic HR policy doc that doesn’t have the details.
* We ensured the generative input wasn’t too large by truncating. In practice, make sure important parts of documents aren’t truncated out (you might prefer feeding each doc separately to the generator if context window is an issue, or use a model with larger context like LongT5 if needed).
* We structured the prompt in a way to help attribution. In production, ensuring the answer is traceable to sources is often important.

For evaluation, we would test this with known Q\&A pairs or have HR verify the answers.

### **6.2 Case Study: Document Retrieval (Semantic Search Engine)**

Now consider a different scenario: a search application where the goal is not to generate an answer, but to find the most relevant documents or passages given a query. This is akin to a semantic search engine. Cross-encoders can greatly enhance the ranking of search results, making sure the best matches appear first.

**Scenario:** A legal document search engine where a user types a query (in plain language) and the system returns the top relevant clauses or documents from a law database.

**Approach:** We will use the bi-encoder to fetch candidates (for recall) and the cross-encoder to re-rank them for precision. Instead of a generative model, we directly present the top passages to the user (maybe highlighting query terms, etc., but here we focus on retrieval).

**Steps:**

1. Index the legal documents (could be broken into clauses or paragraphs).
2. For a given query, retrieve & re-rank as before.
3. Return the top results as the answer.

**Code Example:**

```python
def search_documents(query, top_k=100, top_n=5):
    # Retrieve with bi-encoder
    q_vec = bi_encoder.encode(query, convert_to_tensor=True)
    _, idxs = index.search(q_vec.cpu().numpy(), top_k)
    idxs = idxs.flatten().tolist()
    candidates = [documents[i] for i in idxs]  # each document is say (id, text)
    # Re-rank with cross-encoder
    cross_inps = [[query, text] for (_, text) in candidates]
    scores = cross_encoder.predict(cross_inps)
    ranked = sorted(zip(candidates, scores), key=lambda x: x[1], reverse=True)
    top_results = ranked[:top_n]
    return top_results

# Example usage:
query = "tenant eviction notice period law"
results = search_documents(query, top_k=50, top_n=3)
print(f"Search results for: '{query}'")
for (doc_id, text), score in results:
    snippet = text[:200].replace('\n',' ')  # first 200 chars
    print(f"Document {doc_id} (score {score:.2f}): {snippet}...")
```

Expected output:

```
Search results for: 'tenant eviction notice period law'
Document LawClause17 (score 9.87): ...the landlord must provide a written eviction notice at least 30 days prior to termination of tenancy...
Document LawClause5 (score 8.45): ...tenancy can only be terminated with a 30-day notice (60 days if the tenant has occupied the property for more than a year) according to state law...
Document LawSummary1 (score 7.80): ...summary of eviction laws: generally require a minimum notice period which varies by jurisdiction, commonly 30 days...
```

Here, the cross-encoder (likely fine-tuned on QA or legal texts) identified that Document LawClause17 explicitly mentions a 30-day notice, which directly matches “notice period” in context of eviction, so it got the highest score. Perhaps LawClause5 also mentions it but slightly different context, etc. The important part is the **ordering**: thanks to re-ranking, the most on-point result is first.

This shows a pure retrieval use-case. We didn’t generate an answer because typically in search, the user wants to read the document. However, you could combine it: show the top paragraph and also generate a brief answer summary of it — a hybrid of search and QA.

**Implementational details:**

* We returned a snippet of the text. You might store documents with titles and return those with highlights of the snippet.
* The score could be used or not shown to user, but it’s used internally for ordering. If the score is above some threshold, you might label it as “highly relevant”.
* If no result has a decent score, perhaps inform the user no good match was found (or fall back to a keyword search).

The cross-encoder ensures precision: In semantic search, sometimes the first results from a pure vector search can be a bit off (especially if the query has multiple facets). The cross-encoder tends to reward documents that cover all aspects of the query. E.g., “tenant eviction notice period” – a bi-encoder might retrieve some documents about eviction in general (which cover “tenant eviction”) but maybe not specifically notice period. The cross-encoder will favor those that mention notice period requirement because it catches the interaction of "eviction" with "notice period".

**Evaluating search:** One can measure MRR or Precision\@3 by having a set of example queries with known relevant documents. If user click data is available, that’s also valuable for evaluation (which result did they click?). Over time, this could be used to further fine-tune the ranking model.

### **6.3 Lessons from the Examples**

Across these examples, a few general best practices emerge:

* **Document Chunking:** Both examples used splitting of documents (policies into paragraphs, laws into clauses). Appropriate chunk size is key – too large and you dilute relevance (and risk truncation in cross-encoder), too small and you may need many pieces to cover an answer. Experiment with what works for your domain (often 1-3 paragraphs per chunk is a starting point).
* **Model Choice:** We used relatively small models for speed. In a real system, you might try a larger bi-encoder (like `multi-qa-mpnet-base-dot-v1`) for possibly better recall, and a larger cross-encoder (like one based on Electra or even a mini version of MPNet cross-encoder). Always consider the trade-off. You can also ensemble signals (e.g., combine bi-encoder score and cross-encoder score, or cross-encoder plus some keyword overlap count) if needed for extra fine control.
* **Domain Adaptation:** If the domain is technical (like law or medicine), using models already fine-tuned on that domain (or fine-tuning them yourself as in Chapter 3) yields best results. For instance, a LegalBERT as cross-encoder could outperform a general MS MARCO model on legal text ranking.
* **Pipelines and Tools:** In the second example, we essentially built a search engine. There are existing frameworks (like **Haystack** by deepset) that can manage pipelines of retriever + ranker + reader. In fact, the Haystack framework was mentioned in a Hugging Face blog for building RAG with a cross-encoder reranker and an LLM generator. Such frameworks allow you to define components and connect them, and they handle a lot of boilerplate (caching, parallelism, etc.). As an experienced developer, you might appreciate the flexibility of coding it yourself, but for production it might be worth considering those tools for easier maintenance.
* **Iterate with Users:** If this is a user-facing system, gather feedback. If users can mark an answer as helpful or not, or choose the correct document, that data is gold for improving the models (you can feed it into fine-tuning later).

By now, you should have a solid understanding and practical know-how of building RAG systems with cross-encoders. We covered the theory of why cross-encoders are beneficial, how to implement the retrieval and generation pipeline with Hugging Face models, how to fine-tune models for better performance, how to evaluate and refine the system, and how to deploy it effectively. With this knowledge, you can apply RAG to a wide range of real-world problems – from enterprise Q\&A assistants to advanced search engines – delivering AI solutions that are both knowledgeable and accurate, backed by the best of retrieval and generation.
