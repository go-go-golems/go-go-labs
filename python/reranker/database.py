#!/usr/bin/env python3
import os
import sqlite3
import logging
import traceback
from contextlib import contextmanager
import json

# Configure logging
logger = logging.getLogger(__name__)

DB_PATH = os.path.join(os.path.dirname(__file__), "reranker.db")
TIMEOUT = 20.0  # Timeout in seconds for database operations
logger.info(f"Database path: {DB_PATH}")

def init_db():
    """Initialize the database with required tables if they don't exist."""
    logger.info("Initializing database")
    try:
        with get_db_connection() as conn:
            cursor = conn.cursor()
            
            # Create documents table
            logger.debug("Creating documents table if it doesn't exist")
            cursor.execute('''
            CREATE TABLE IF NOT EXISTS documents (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                content TEXT UNIQUE NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
            ''')
            
            # Create queries table
            logger.debug("Creating queries table if it doesn't exist")
            cursor.execute('''
            CREATE TABLE IF NOT EXISTS queries (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                query TEXT NOT NULL,
                top_k INTEGER,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
            ''')
            
            # Create results table
            logger.debug("Creating results table if it doesn't exist")
            cursor.execute('''
            CREATE TABLE IF NOT EXISTS results (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                query_id INTEGER NOT NULL,
                document_id INTEGER NOT NULL,
                score REAL NOT NULL,
                rank INTEGER NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (query_id) REFERENCES queries (id),
                FOREIGN KEY (document_id) REFERENCES documents (id)
            )
            ''')
            
            conn.commit()
            logger.info("Database initialized successfully")
    except Exception as e:
        logger.error(f"Error initializing database: {str(e)}")
        logger.error(traceback.format_exc())
        raise

@contextmanager
def get_db_connection():
    """Context manager for database connections with improved timeout and transaction handling."""
    logger.debug(f"Opening database connection to {DB_PATH}")
    conn = None
    try:
        conn = sqlite3.connect(DB_PATH, timeout=TIMEOUT)
        conn.row_factory = sqlite3.Row
        conn.isolation_level = None  # Enable autocommit mode
        conn.execute("BEGIN")  # Start transaction explicitly
        yield conn
        conn.execute("COMMIT")  # Commit transaction
        logger.debug("Database connection closed normally")
    except Exception as e:
        if conn:
            try:
                conn.execute("ROLLBACK")  # Rollback on error
                logger.debug("Transaction rolled back due to error")
            except Exception as rollback_error:
                logger.error(f"Error during rollback: {str(rollback_error)}")
        logger.error(f"Database connection error: {str(e)}")
        logger.error(traceback.format_exc())
        raise
    finally:
        if conn:
            conn.close()

def get_or_create_document(content, conn=None):
    """Get a document ID or create it if it doesn't exist. Optionally reuse an existing connection."""
    should_close_conn = conn is None
    try:
        if should_close_conn:
            conn = sqlite3.connect(DB_PATH, timeout=TIMEOUT)
            conn.row_factory = sqlite3.Row
        
        cursor = conn.cursor()
        cursor.execute("SELECT id FROM documents WHERE content = ?", (content,))
        result = cursor.fetchone()
        
        if result:
            logger.debug(f"Found existing document with ID: {result['id']}")
            return result["id"]
        
        logger.debug("Creating new document")
        cursor.execute("INSERT INTO documents (content) VALUES (?)", (content,))
        new_id = cursor.lastrowid
        logger.debug(f"Created new document with ID: {new_id}")
        return new_id
    except Exception as e:
        logger.error(f"Error in get_or_create_document: {str(e)}")
        logger.error(traceback.format_exc())
        raise
    finally:
        if should_close_conn and conn:
            conn.close()

def save_query_and_results(query, results, top_k=None):
    """Save a query and its results to the database with improved transaction handling."""
    logger.info(f"Saving query: '{query}' with {len(results)} results")
    try:
        with get_db_connection() as conn:
            cursor = conn.cursor()
            
            # Save the query
            logger.debug(f"Inserting query: '{query}' with top_k={top_k}")
            cursor.execute(
                "INSERT INTO queries (query, top_k) VALUES (?, ?)",
                (query, top_k)
            )
            query_id = cursor.lastrowid
            logger.debug(f"Query inserted with ID: {query_id}")
            
            # Save the results
            for rank, (doc, score) in enumerate(results, 1):
                logger.debug(f"Processing result {rank} with score {score}")
                # Reuse the connection for get_or_create_document
                doc_id = get_or_create_document(doc, conn)
                cursor.execute(
                    "INSERT INTO results (query_id, document_id, score, rank) VALUES (?, ?, ?, ?)",
                    (query_id, doc_id, score, rank)
                )
            
            logger.info(f"Successfully saved query ID {query_id} with {len(results)} results")
            return query_id
    except Exception as e:
        logger.error(f"Error in save_query_and_results: {str(e)}")
        logger.error(traceback.format_exc())
        raise

def get_recent_queries(limit=10):
    """Get recent queries with pagination."""
    logger.info(f"Getting {limit} recent queries")
    try:
        with get_db_connection() as conn:
            cursor = conn.cursor()
            cursor.execute(
                "SELECT id, query, top_k, created_at FROM queries ORDER BY created_at DESC LIMIT ?",
                (limit,)
            )
            results = cursor.fetchall()
            logger.debug(f"Retrieved {len(results)} recent queries")
            return results
    except Exception as e:
        logger.error(f"Error in get_recent_queries: {str(e)}")
        logger.error(traceback.format_exc())
        raise

def get_query_results(query_id):
    """Get the results for a specific query."""
    logger.info(f"Getting results for query ID: {query_id}")
    try:
        with get_db_connection() as conn:
            cursor = conn.cursor()
            cursor.execute(
                """
                SELECT q.query, d.content as document, r.score, r.rank
                FROM results r
                JOIN queries q ON r.query_id = q.id
                JOIN documents d ON r.document_id = d.id
                WHERE r.query_id = ?
                ORDER BY r.rank
                """,
                (query_id,)
            )
            results = cursor.fetchall()
            logger.debug(f"Retrieved {len(results)} results for query ID {query_id}")
            return results
    except Exception as e:
        logger.error(f"Error in get_query_results: {str(e)}")
        logger.error(traceback.format_exc())
        raise

def get_all_documents():
    """Get all unique documents in the database."""
    logger.info("Getting all documents")
    try:
        with get_db_connection() as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT id, content FROM documents ORDER BY id")
            results = cursor.fetchall()
            logger.debug(f"Retrieved {len(results)} documents")
            return results
    except Exception as e:
        logger.error(f"Error in get_all_documents: {str(e)}")
        logger.error(traceback.format_exc())
        raise

# Initialize the database on module import
try:
    init_db()
except Exception as e:
    logger.critical(f"Failed to initialize database: {str(e)}")
    logger.critical(traceback.format_exc()) 