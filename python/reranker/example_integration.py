import requests
import json
from typing import List, Dict, Any

class ArxivReranker:
    """Client class for the ArXiv paper reranker service"""
    
    def __init__(self, api_url="http://localhost:8000"):
        """Initialize the reranker client
        
        Args:
            api_url: URL of the reranker API service
        """
        self.api_url = api_url.rstrip('/')
    
    def rerank(self, query: str, arxiv_results: List[Dict[str, Any]], top_n: int = 10) -> List[Dict[str, Any]]:
        """Rerank ArXiv papers based on relevance to query
        
        Args:
            query: Search query or intent
            arxiv_results: List of ArXiv paper results
            top_n: Number of top results to return
            
        Returns:
            List of reranked papers with relevance scores
        """
        # Prepare the request data
        data = {
            "query": query,
            "results": arxiv_results,
            "top_n": top_n
        }
        
        # Send request to the rerank endpoint
        response = requests.post(f"{self.api_url}/rerank", json=data)
        
        # Check response
        if response.status_code != 200:
            raise Exception(f"Reranking failed: {response.text}")
        
        # Parse response
        result = response.json()
        return result["reranked_results"]
    
    def rerank_json(self, query: str, arxiv_json: Dict[str, Any], top_n: int = 10) -> Dict[str, Any]:
        """Rerank papers from standard ArXiv JSON format
        
        Args:
            query: Search query or intent
            arxiv_json: ArXiv search results JSON with 'results' key
            top_n: Number of top results to return
            
        Returns:
            Dict with query and reranked results
        """
        # Send request to the rerank_json endpoint
        response = requests.post(
            f"{self.api_url}/rerank_json",
            params={"query": query, "top_n": top_n},
            json=arxiv_json
        )
        
        # Check response
        if response.status_code != 200:
            raise Exception(f"JSON reranking failed: {response.text}")
        
        # Parse response
        return response.json()

# Example usage in an application
def main():
    # Sample ArXiv search results file
    sample_file = "ttmp/2025-05-10/03-format-of-mcp-arxiv-search-results.json"
    
    # Load the sample data
    with open(sample_file, 'r') as f:
        arxiv_data = json.load(f)
    
    # Create the reranker client
    reranker = ArxivReranker()
    
    # Different search queries to test
    queries = [
        "cross-encoder models for information retrieval",
        "multi-modal retrieval augmented generation",
        "graph-based knowledge retrieval"
    ]
    
    # Test each query
    for query in queries:
        print(f"\n\nQuery: {query}")
        print("-" * 80)
        
        # Get reranked results
        try:
            results = reranker.rerank_json(query, arxiv_data, top_n=3)
            
            # Display results
            for i, paper in enumerate(results["reranked_results"]):
                print(f"\n{i+1}. {paper['Title']} (Score: {paper['score']:.2f})")
                print(f"   Authors: {', '.join(paper['Authors'])}")
                print(f"   Published: {paper['Published'][:10]}")
                print(f"   Category: {paper['Metadata'].get('primary_category', 'N/A')}")
        
        except Exception as e:
            print(f"Error: {e}")
            print("Is the reranker server running? Start it with 'python arxiv_reranker_server.py'")
            break

if __name__ == "__main__":
    main()