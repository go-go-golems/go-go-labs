# Module 4: Document Summarization with DSPy - Interactive Exercises
# %%

# type: ignore # Ignore missing stubs for dspy
import dspy  # type: ignore

import os
import time
from typing import List, Dict, Any, Tuple, Optional
from functools import partial

# Setup - reuse configuration from previous modules
# Set cache directory (optional)
os.environ["DSPY_CACHEDIR"] = os.path.join(os.getcwd(), "cache")

# Configure LM with explicit cache settings
lm = dspy.LM(
    "openai/gpt-4o-mini", api_key=os.getenv("OPENAI_API_KEY"), cache=False
)  # Set to False to disable caching
dspy.configure(lm=lm, track_usage=True)  # Enable usage tracking

# Verify LM is loaded before proceeding
assert (
    dspy.settings.lm is not None
), "Please configure a language model before running predictions"


# %% Example 1: Define Signature Classes for Document Summarization Components


class ProduceGist(dspy.Signature):
    """Produce a one- or two-sentence gist of what this chunk is about, so we can assign it to a class."""

    toc_path: list[str] = dspy.InputField(
        desc="Path down which this chunk has traveled so far in the Table of Contents"
    )
    chunk: str = dspy.InputField()
    gist: str = dspy.OutputField()


class ProduceHeaders(dspy.Signature):
    """
    Produce a list of headers (top-level Table of Contents) for structuring a report on *all* chunk contents.
    Make sure every chunk would belong to exactly one section.
    """

    toc_path: list[str] = dspy.InputField()
    chunk_summaries: str = dspy.InputField()
    headers: list[str] = dspy.OutputField()


class WriteSection(dspy.Signature):
    """
    Craft a Markdown section, given a path down the table of contents, which ends with this section's specific heading.
    Start the section with that heading; use sub-headings of depth at least +1 relative to the ToC path.
    Your section's content is to be entirely derived from the given list of chunks. That content must be complete but
    very concise, with all necessary knowledge from the chunks reproduced and repetitions or irrelevant details omitted.
    """

    toc_path: list[str] = dspy.InputField()
    content_chunks: list[str] = dspy.InputField()
    section_content: str = dspy.OutputField()


# %% Example 2: Basic Document Summarization Function


def basic_summarize(
    toc_path: list[str],
    chunks: list[str],
) -> str:
    """A basic implementation of document summarization using DSPy."""

    # If we don't have enough chunks or the TOC path is sufficient, just use ChainOfThought
    if len(chunks) < 5 or len(toc_path) >= 3:
        content = dspy.ChainOfThought(WriteSection)(
            toc_path=toc_path, content_chunks=chunks
        ).section_content
        return f"{toc_path[-1]}\n\n{content}"

    # For larger documents, we need to break it down
    # First, generate gists for each chunk
    chunk_gists = []
    for chunk in chunks:
        gist_result = dspy.ChainOfThought(ProduceGist)(toc_path=toc_path, chunk=chunk)
        chunk_gists.append(gist_result.gist)

    chunk_summaries = "\n".join(chunk_gists)

    # Generate headers based on chunk summaries
    headers_result = dspy.ChainOfThought(ProduceHeaders)(
        toc_path=toc_path, chunk_summaries=chunk_summaries
    )
    headers = headers_result.headers

    # Group chunks by topic/header
    sections = {topic: [] for topic in headers}

    # For each chunk, classify which header it belongs to
    for chunk in chunks:
        # Simple classification by keyword matching (in practice use an LM)
        best_topic = headers[0]  # Default to first header
        for topic in headers:
            if topic.lower() in chunk.lower():
                best_topic = topic
                break
        sections[best_topic].append(chunk)

    # Generate section content for each header
    section_texts = []
    for topic, topic_chunks in sections.items():
        if topic_chunks:  # Skip empty sections
            section_path = toc_path + [topic]
            section_content = dspy.ChainOfThought(WriteSection)(
                toc_path=section_path, content_chunks=topic_chunks
            ).section_content
            section_texts.append(section_content)

    # Combine all sections into a final document
    return toc_path[-1] + "\n\n" + "\n\n".join(section_texts)


# %% Example 3: Parallel Processing with DSPy


def parallelize(module_class, batch_examples=None):
    """Create a parallelized version of a module to process multiple examples.

    This is a simplified version of parallelization that doesn't use actual
    parallel computation but demonstrates the concept.
    """
    module = module_class()

    def process_batch(examples):
        results = []
        for example in examples:
            result = module(**example)
            results.append(result)
        return results

    return process_batch


# %% Example 4: Advanced Document Summarization with Parallelization


def massively_summarize(
    toc_path: list[str],
    chunks: list[str],
) -> str:
    """A more advanced implementation of document summarization using parallelization."""

    # If we have few chunks or deep TOC path, use simple method
    if len(chunks) < 5 or len(toc_path) >= 3:
        content = dspy.ChainOfThought(WriteSection)(
            toc_path=toc_path, content_chunks=chunks
        ).section_content
        return f"{toc_path[-1]}\n\n{content}"

    # Create parallelized modules
    produce_gist = parallelize(lambda: dspy.ChainOfThought(ProduceGist))

    # Process chunks in parallel to extract gists
    chunk_summaries = produce_gist(
        [{"toc_path": toc_path, "chunk": chunk} for chunk in chunks]
    )
    # Extract just the gist text from each result
    chunk_summaries = [summary.gist for summary in chunk_summaries]

    # Generate headers based on chunk summaries
    produce_headers = dspy.ChainOfThought(ProduceHeaders)
    headers = produce_headers(
        toc_path=toc_path, chunk_summaries="\n".join(chunk_summaries)
    ).headers

    # Classify chunks into topics
    classify = parallelize(
        lambda: dspy.ChainOfThought(f"toc_path: list[str], chunk -> topic: str")
    )
    topics = classify([{"toc_path": toc_path, "chunk": chunk} for chunk in chunks])
    topics = [result.topic for result in topics]

    # Group chunks by topic
    sections = {topic: [] for topic in headers}
    for topic, chunk in zip(topics, chunks):
        # Find the best matching header if the topic doesn't exactly match
        best_header = None
        for header in headers:
            if topic.lower() == header.lower() or topic.lower() in header.lower():
                best_header = header
                break

        # Use first header as fallback if no match found
        if best_header is None:
            best_header = headers[0]

        sections[best_header].append(chunk)

    # Create a parallel version of massively_summarize for recursively processing each section
    parallel_massively_summarize = parallelize(lambda: massively_summarize)

    # Process each section in parallel
    summarized_sections = []
    section_inputs = []

    for topic, section_chunks in sections.items():
        if section_chunks:  # Skip empty sections
            section_inputs.append(
                {"toc_path": toc_path + [topic], "chunks": section_chunks}
            )

    if section_inputs:
        summarized_sections = parallel_massively_summarize(section_inputs)

    # Join the sections and return
    return toc_path[-1] + "\n\n" + "\n".join(summarized_sections)


# %% Example 5: Implementation with Real Parallelization using DSPy's parallel utilities


def massively_summarize_with_dspy_parallel(
    toc_path: list[str],
    chunks: list[str],
) -> str:
    """
    Implementation of document summarization using DSPy's actual parallelization features.
    This is closer to the code shown in the screenshots.
    """

    # If we have few chunks or deep TOC path, use simple method
    if len(chunks) < 5 or len(toc_path) >= 3:
        content = dspy.ChainOfThought(WriteSection)(
            toc_path=toc_path, content_chunks=chunks
        ).section_content
        return f"{toc_path[-1]}\n\n{content}"

    # Create parallelized modules using DSPy's parallelization
    produce_gist = dspy.ChainOfThought(ProduceGist)

    # Generate gists for each chunk
    chunk_summaries = []
    for chunk in chunks:
        chunk_summaries.append(dspy.Example(toc_path=toc_path, chunk=chunk))

    # Apply parallelization
    chunk_summaries = produce_gist(chunk_summaries)
    chunk_summaries = [summary.gist for summary in chunk_summaries]

    # Generate headers based on chunk summaries
    produce_headers = dspy.ChainOfThought(ProduceHeaders)
    headers = produce_headers(
        toc_path=toc_path, chunk_summaries=chunk_summaries
    ).headers

    # Classify chunks by topic
    classify = dspy.ChainOfThought(f"toc_path: list[str], chunk -> topic: str")

    topics_examples = []
    for chunk in chunks:
        topics_examples.append(dspy.Example(toc_path=toc_path, chunk=chunk))

    topics = classify(topics_examples)
    topics = [result.topic for result in topics]

    # Group chunks by topic
    sections = {topic: [] for topic in headers}
    for topic, chunk in zip(topics, chunks):
        if topic in sections:
            sections[topic].append(chunk)
        else:
            # Fallback to first header if topic doesn't match any header
            sections[headers[0]].append(chunk)

    # Process each section recursively
    summarized_sections = []
    for topic, section_chunks in sections.items():
        if section_chunks:  # Skip empty sections
            section_result = massively_summarize_with_dspy_parallel(
                toc_path=toc_path + [topic], chunks=section_chunks
            )
            summarized_sections.append(section_result)

    # Combine results
    return toc_path[-1] + "\n\n" + "\n".join(summarized_sections)


# %% Example 6: Final Implementation Matching the Screenshot
# This closely mimics the implementation shown in the screenshot


def massively_summarize_final(
    toc_path: list[str],
    chunks: list[str],
) -> str:
    """Document summarization implementation similar to the screenshot code."""

    # If we have few chunks or deep TOC path, use simple method
    if len(chunks) < 5 or len(toc_path) >= 3:
        content = dspy.ChainOfThought(WriteSection)(
            toc_path=toc_path, content_chunks=chunks
        ).section_content
        return f"{toc_path[-1]}\n\n{content}"

    # Process using parallel modules
    produce_gist = parallelize(lambda: dspy.ChainOfThought(ProduceGist))
    chunk_summaries = produce_gist(
        [{"toc_path": toc_path, "chunk": chunk} for chunk in chunks]
    )
    chunk_summaries = [summary.gist for summary in chunk_summaries]

    # Generate headers based on chunk summaries
    produce_headers = dspy.ChainOfThought(ProduceHeaders)
    headers = produce_headers(
        toc_path=toc_path, chunk_summaries=chunk_summaries
    ).headers

    # Classify chunks into topics
    classify = parallelize(
        lambda: dspy.ChainOfThought(
            f"toc_path: list[str], chunk -> topic: Literal{headers}"
        )
    )
    topics = classify([{"toc_path": toc_path, "chunk": chunk} for chunk in chunks])
    topics = [result.topic for result in topics]

    # Create sections dictionary
    sections = {topic: [] for topic in headers}

    # Group chunks by topic
    for topic, chunk in zip(topics, chunks):
        sections[topic].append(chunk)

    # Recursively process each section
    parallel_massively_summarize = parallelize(lambda: massively_summarize_final)

    # Process each section in parallel
    section_inputs = [
        {"toc_path": toc_path + [topic], "chunks": section_chunks}
        for topic, section_chunks in sections.items()
        if section_chunks
    ]

    summarized_sections = parallel_massively_summarize(section_inputs)

    # Join the results
    return toc_path[-1] + "\n\n" + "\n".join(summarized_sections)


# %% Test the summarization functions with example data

# Example document chunks
document_chunks = [
    "DSPy is a framework for programming with foundation models. It allows you to build modular AI systems and optimize them automatically.",
    "One key concept in DSPy is the Signature, which defines the input/output interface for modules. Signatures can be defined using either strings or classes.",
    "DSPy modules like Predict and ChainOfThought implement different prompting strategies. They can be composed to create complex pipelines.",
    "Compiling a DSPy program optimizes its prompts or weights based on examples. This is done using optimizers (previously called teleprompters).",
    "DSPy supports tool use through modules like ReAct, which implements the Reason+Act paradigm, allowing models to use external tools.",
    "RAG (Retrieval-Augmented Generation) is easy to implement in DSPy by combining retrieval modules with generation modules.",
    "DSPy programs can be optimized using small datasets of examples. Even just 5-10 examples can lead to significant improvements.",
    "Multi-turn conversation can be implemented in DSPy using the History primitive to maintain context across turns.",
    "DSPy can integrate with other frameworks like LangChain for data prep and external utilities.",
    "Testing and evaluating DSPy agents is important to ensure reliability. DSPy provides tools for this like tracing and inspection.",
]

# Test our implementation
print("Basic Summarization:\n")
basic_result = basic_summarize(["DSPy Documentation"], document_chunks)
print(basic_result)

print("\n\nAdvanced Summarization:\n")
advanced_result = massively_summarize_final(["DSPy Documentation"], document_chunks)
print(advanced_result)

# %% Exercise Area
# Experiment with your own document summarization implementations
