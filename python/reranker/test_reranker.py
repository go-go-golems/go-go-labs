import json
import logging
import requests

# Configure a simple logger for the test script
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("reranker_test")

# Path to the sample ArXiv results file
SAMPLE_FILE = "ttmp/2025-05-10/03-format-of-mcp-arxiv-search-results.json"

def test_reranker_local():
    """Test the reranker by making a request to the local server"""
    # Load the sample data
    with open(SAMPLE_FILE, 'r') as f:
        arxiv_data = json.load(f)
    
    # The search query
    query = "cross-encoder models for information retrieval reranking"
    
    # Make a request to the rerank_json endpoint
    response = requests.post(
        "http://localhost:8000/rerank_json",
        params={"query": query, "top_n": 5},
        json=arxiv_data
    )
    
    # Check if the request was successful
    if response.status_code == 200:
        results = response.json()
        
        # Print the reranked results
        print(f"Query: {results['query']}")
        print("\nReranked Results:")
        
        for i, paper in enumerate(results['reranked_results']):
            print(f"\n{i+1}. {paper['Title']} (Score: {paper['score']:.2f})")
            print(f"   Authors: {', '.join(paper['Authors'])}")
            print(f"   Abstract: {paper['Abstract'][:200]}...")
            print(f"   URL: {paper['SourceURL']}")
    else:
        print(f"Error: {response.status_code}")
        print(response.text)

def test_reranker_direct():
    """Test the reranker directly by importing the module"""
    from arxiv_reranker_server import rerank_json
    import asyncio
    
    logger.info("Starting direct reranker test without server")
    
    # Load the sample data
    try:
        with open(SAMPLE_FILE, 'r') as f:
            arxiv_data = json.load(f)
        logger.info(f"Loaded sample data from {SAMPLE_FILE}: {len(arxiv_data.get('results', []))} papers")
    except Exception as e:
        logger.error(f"Failed to load sample data: {e}")
        return
    
    # The search query
    query = "cross-encoder models for information retrieval reranking"
    logger.info(f"Using query: '{query}'")
    
    try:
        # Call the rerank_json function directly
        logger.info("Running rerank_json function...")
        response = asyncio.run(rerank_json(query=query, arxiv_json=arxiv_data, top_n=5))
        logger.info("Reranking completed successfully")
        
        # Convert Pydantic model to dict for easier handling
        # Support both Pydantic v1 and v2
        results = response.model_dump() if hasattr(response, 'model_dump') else response.dict()
        logger.info(f"Converted response to dictionary with {len(results.get('reranked_results', []))} results")
    except Exception as e:
        logger.error(f"Reranking failed: {e}", exc_info=True)
        return
    
    # Print the reranked results
    print(f"Query: {results['query']}")
    print("\nReranked Results:")
    
    for i, paper in enumerate(results['reranked_results']):
        # Convert each paper to dict as well (support both Pydantic versions)
        if isinstance(paper, dict):
            paper_dict = paper
        else:
            # Support both Pydantic v1 and v2
            paper_dict = paper.model_dump() if hasattr(paper, 'model_dump') else paper.dict()
        print(f"\n{i+1}. {paper_dict['Title']} (Score: {paper_dict['score']:.2f})")
        print(f"   Authors: {', '.join(paper_dict['Authors'])}")
        print(f"   Abstract: {paper_dict['Abstract'][:200]}...")
        print(f"   URL: {paper_dict['SourceURL']}")

if __name__ == "__main__":
    print("Running direct test (no server needed)")
    test_reranker_direct()
    
    # Uncomment to test with running server
    # print("\n\nRunning test with local server")
    # test_reranker_local()