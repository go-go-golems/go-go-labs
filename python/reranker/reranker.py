#!/usr/bin/env python3

import logging
import traceback
from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch

# Configure logging
logger = logging.getLogger(__name__)

DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
MODEL_NAME = "BAAI/bge-reranker-large"

logger.info(f"Using device: {DEVICE}")
logger.info(f"Reranker model: {MODEL_NAME}")

# Lazy loading of model and tokenizer
_tokenizer = None
_model = None

def _get_tokenizer():
    global _tokenizer
    if _tokenizer is None:
        logger.info(f"Loading tokenizer: {MODEL_NAME}")
        try:
            _tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
            logger.info("Tokenizer loaded successfully")
        except Exception as e:
            logger.error(f"Error loading tokenizer: {str(e)}")
            logger.error(traceback.format_exc())
            raise
    return _tokenizer

def _get_model():
    global _model
    if _model is None:
        logger.info(f"Loading model: {MODEL_NAME} on {DEVICE}")
        try:
            _model = AutoModelForSequenceClassification.from_pretrained(MODEL_NAME).to(DEVICE)
            _model.eval()
            logger.info("Model loaded successfully and set to eval mode")
        except Exception as e:
            logger.error(f"Error loading model: {str(e)}")
            logger.error(traceback.format_exc())
            raise
    return _model

def rerank(query: str, documents: list[str], top_k: int | None = None):
    """Return (doc, score) sorted best-to-worst for a single query."""
    if not documents:
        logger.warning("Empty documents list provided to rerank")
        return []
    
    logger.info(f"Reranking {len(documents)} documents for query: '{query}'")
    logger.debug(f"top_k value: {top_k!r}, type: {type(top_k)}")
    
    try:
        # Get tokenizer and model
        tokenizer = _get_tokenizer()
        model = _get_model()
        
        # build parallel list of (query, doc) pairs, batch-encode, and run the model
        pairs = [(query, doc) for doc in documents]
        logger.debug(f"Created {len(pairs)} query-document pairs")
        
        logger.debug("Encoding query-document pairs")
        batch = tokenizer.batch_encode_plus(
            pairs, padding=True, truncation=True, return_tensors="pt"
        ).to(DEVICE)
        
        logger.debug("Running inference")
        with torch.no_grad():
            # logits[:, 0] is a single relevance score for every pair
            scores = model(**batch).logits.squeeze(-1).cpu()
        
        logger.debug(f"Inference complete, got {len(scores)} scores")
        
        ranked = sorted(zip(documents, scores.tolist()), key=lambda x: x[1], reverse=True)
        
        if top_k is not None:
            logger.debug(f"Slicing ranked results with top_k={top_k!r} (type: {type(top_k)})")
            try:
                ranked = ranked[:top_k]
                logger.info(f"Returning top {top_k} results")
            except TypeError as e:
                logger.error(f"Failed to slice ranked results with top_k={top_k!r} (type: {type(top_k)})")
                raise
        else:
            logger.info(f"Returning all {len(ranked)} results")
        
        # Log a few top results
        for i, (doc, score) in enumerate(ranked[:3] if ranked else []):
            logger.debug(f"Result {i+1}: Score={score:.4f}, Doc='{doc[:50]}...'")
            
        return ranked
        
    except Exception as e:
        logger.error(f"Error in rerank function: {str(e)}")
        logger.error(traceback.format_exc())
        raise 