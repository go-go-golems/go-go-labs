-- SQL Examples for SQLite Vector Search with get_embedding Function

-- Example 1: Insert documents with computed embeddings
INSERT INTO documents (content, embedding) 
VALUES 
  ('Machine learning is awesome', get_embedding('Machine learning is awesome')),
  ('SQLite is a great database', get_embedding('SQLite is a great database')),
  ('Vector search enables semantic search', get_embedding('Vector search enables semantic search'));

-- Example 2: Search for similar documents using computed query embedding
SELECT 
    content,
    cosine_similarity(embedding, get_embedding('artificial intelligence')) as similarity
FROM documents
WHERE similarity > 0.3
ORDER BY similarity DESC
LIMIT 5;

-- Example 3: Find the most similar document to a query
SELECT 
    content,
    cosine_similarity(embedding, get_embedding('database technology')) as similarity
FROM documents
ORDER BY similarity DESC
LIMIT 1;

-- Example 4: Update existing documents with new embeddings
UPDATE documents 
SET embedding = get_embedding(content)
WHERE embedding IS NULL OR embedding = '';

-- Example 5: Create a view that always computes fresh embeddings
CREATE VIEW fresh_embeddings AS
SELECT 
    id,
    content,
    get_embedding(content) as embedding
FROM documents;

-- Example 6: Compare similarity between two arbitrary texts
SELECT cosine_similarity(
    get_embedding('I love programming'),
    get_embedding('Coding is fun')
) as similarity;

-- Example 7: Find duplicates or near-duplicates
SELECT 
    a.content as doc1,
    b.content as doc2,
    cosine_similarity(a.embedding, get_embedding(b.content)) as similarity
FROM documents a
CROSS JOIN documents b
WHERE a.id < b.id
  AND cosine_similarity(a.embedding, get_embedding(b.content)) > 0.9
ORDER BY similarity DESC;

-- Example 8: Batch processing with computed embeddings
CREATE TEMPORARY TABLE new_docs AS
SELECT 
    content,
    get_embedding(content) as embedding
FROM (
    SELECT 'Natural language processing' as content
    UNION SELECT 'Deep learning neural networks'
    UNION SELECT 'Computer vision algorithms'
    UNION SELECT 'Reinforcement learning agents'
);

-- Example 9: Similarity search with multiple query terms
WITH query_terms AS (
    SELECT get_embedding('machine learning') as ml_embedding,
           get_embedding('data science') as ds_embedding,
           get_embedding('artificial intelligence') as ai_embedding
)
SELECT 
    content,
    GREATEST(
        cosine_similarity(embedding, (SELECT ml_embedding FROM query_terms)),
        cosine_similarity(embedding, (SELECT ds_embedding FROM query_terms)),
        cosine_similarity(embedding, (SELECT ai_embedding FROM query_terms))
    ) as max_similarity
FROM documents
ORDER BY max_similarity DESC
LIMIT 10;

-- Example 10: Create a function-based index simulation
-- (SQLite doesn't support function-based indexes, but we can precompute)
CREATE TABLE document_embeddings AS
SELECT 
    id,
    content,
    get_embedding(content) as embedding,
    length(get_embedding(content)) as embedding_dimensions
FROM documents;

-- Example 11: Semantic clustering by similarity threshold
WITH similarity_matrix AS (
    SELECT 
        a.id as doc_a,
        b.id as doc_b,
        a.content as content_a,
        b.content as content_b,
        cosine_similarity(a.embedding, get_embedding(b.content)) as similarity
    FROM documents a
    CROSS JOIN documents b
    WHERE a.id != b.id
)
SELECT 
    doc_a,
    content_a,
    COUNT(*) as similar_docs
FROM similarity_matrix
WHERE similarity > 0.7
GROUP BY doc_a, content_a
ORDER BY similar_docs DESC;

-- Example 12: Real-time similarity search with user input
-- This would be called from application code with user query
SELECT 
    id,
    content,
    cosine_similarity(embedding, get_embedding(?)) as relevance_score
FROM documents
WHERE relevance_score > 0.2
ORDER BY relevance_score DESC
LIMIT 20;

-- Example 13: Multi-language support (if model supports it)
SELECT 
    'English' as language,
    cosine_similarity(
        get_embedding('Hello world'),
        get_embedding('Greetings universe')
    ) as similarity
UNION
SELECT 
    'Cross-language',
    cosine_similarity(
        get_embedding('Hello world'),
        get_embedding('Bonjour le monde')
    ) as similarity;

-- Example 14: Embedding quality analysis
SELECT 
    content,
    length(get_embedding(content)) as embedding_size,
    CASE 
        WHEN length(get_embedding(content)) = '[]' THEN 'Empty'
        WHEN json_array_length(get_embedding(content)) > 300 THEN 'Large'
        WHEN json_array_length(get_embedding(content)) > 100 THEN 'Medium'
        ELSE 'Small'
    END as embedding_category
FROM documents
LIMIT 10;

-- Example 15: Incremental similarity search
-- Find documents similar to a growing collection
WITH target_collection AS (
    SELECT get_embedding('machine learning data science') as target_embedding
)
SELECT 
    content,
    cosine_similarity(
        embedding, 
        (SELECT target_embedding FROM target_collection)
    ) as similarity
FROM documents
WHERE cosine_similarity(
    embedding, 
    (SELECT target_embedding FROM target_collection)
) > (
    SELECT AVG(cosine_similarity(
        embedding, 
        (SELECT target_embedding FROM target_collection)
    )) + 0.1
    FROM documents
)
ORDER BY similarity DESC;

-- Example 16: Performance testing query
-- Test embedding generation speed
SELECT 
    COUNT(*) as processed_docs,
    'Embeddings generated' as status
FROM (
    SELECT get_embedding(content) as emb
    FROM documents
    LIMIT 100
) 
WHERE emb != '[]';

-- Example 17: Error handling in SQL
SELECT 
    content,
    CASE 
        WHEN get_embedding(content) = '[]' THEN 'ERROR: Failed to generate embedding'
        WHEN length(get_embedding(content)) < 10 THEN 'WARNING: Suspiciously short embedding'
        ELSE 'OK: Valid embedding'
    END as embedding_status
FROM documents;
