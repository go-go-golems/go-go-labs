# Module 3: DSPy Optimization and Evaluation
# %%

import dspy  # type: ignore
import os
from typing import List, Optional
from dataclasses import dataclass

# Setup - reuse configuration from previous modules
os.environ["DSPY_CACHEDIR"] = os.path.join(os.getcwd(), "cache")

# Configure LM with explicit cache settings
lm = dspy.LM("openai/gpt-4o-mini", api_key=os.getenv("OPENAI_API_KEY"), cache=True)
dspy.configure(lm=lm)

# Verify LM is loaded before proceeding
assert (
    dspy.settings.lm is not None
), "Please configure a language model before running predictions"


# %% Example 1: Basic Evaluation
# Define a simple QA signature and module
class SimpleQA(dspy.Signature):
    """Simple question answering signature."""

    question: str = dspy.InputField()
    answer: str = dspy.OutputField()


qa_module = dspy.ChainOfThought(SimpleQA)


# Create evaluation dataset
@dataclass
class QAExample:
    question: str
    answer: str


eval_data = [
    QAExample("What is the capital of France?", "Paris"),
    QAExample("Who wrote Romeo and Juliet?", "William Shakespeare"),
    QAExample("What is the chemical symbol for gold?", "Au"),
]


# Evaluate using exact match metric
def evaluate_qa():
    correct = 0
    total = len(eval_data)

    for example in eval_data:
        result = qa_module(question=example.question)
        if result.answer.lower() == example.answer.lower():
            correct += 1

    accuracy = correct / total
    print(f"Accuracy: {accuracy:.2%}")


# Run evaluation
evaluate_qa()


# %% Example 2: Bootstrap Few-Shot Learning
# Define a more complex task: Sentiment Analysis
class SentimentAnalysis(dspy.Signature):
    """Classify text sentiment with explanation."""

    text: str = dspy.InputField()
    sentiment: str = dspy.OutputField(desc="One of: positive, negative, neutral")
    explanation: str = dspy.OutputField()


# Create training examples
train_data = [
    {
        "text": "This movie was absolutely fantastic!",
        "sentiment": "positive",
        "explanation": "Uses enthusiastic language ('absolutely fantastic')",
    },
    {
        "text": "The service was terrible and the food was cold.",
        "sentiment": "negative",
        "explanation": "Criticizes both service and food quality",
    },
    {
        "text": "The weather is quite normal today.",
        "sentiment": "neutral",
        "explanation": "Statement is purely factual without emotional content",
    },
]

# Create the predictor
sentiment_predictor = dspy.ChainOfThought(SentimentAnalysis)

# Create and configure the optimizer
optimizer = dspy.BootstrapFewShot(
    metric=dspy.Exact("sentiment"),
    max_bootstrapped_demos=2,
    max_labeled_demos=2,
)

# Compile the predictor
compiled_predictor = optimizer.compile(sentiment_predictor, trainset=train_data)

# Test the optimized predictor
test_texts = [
    "I love this product so much!",
    "This is just okay, nothing special.",
    "Worst experience ever, don't waste your money.",
]

for text in test_texts:
    result = compiled_predictor(text=text)
    print(f"\nText: {text}")
    print(f"Sentiment: {result.sentiment}")
    print(f"Explanation: {result.explanation}")


# %% Example 3: Advanced Optimization with Multiple Metrics
class TextClassification(dspy.Signature):
    """Classify text into categories with confidence."""

    text: str = dspy.InputField()
    category: str = dspy.OutputField()
    confidence: float = dspy.OutputField()
    reasoning: str = dspy.OutputField()


# Define multiple metrics
class MultiMetric:
    def __call__(self, gold, pred, trace=None):
        category_correct = gold["category"].lower() == pred.category.lower()
        confidence_reasonable = 0 <= float(pred.confidence) <= 1
        has_reasoning = len(pred.reasoning.split()) >= 10

        return all([category_correct, confidence_reasonable, has_reasoning])


# Create training data
classification_data = [
    {
        "text": "The patient shows signs of high fever and cough",
        "category": "medical",
        "confidence": 0.9,
        "reasoning": "Contains medical symptoms and diagnostic language",
    },
    {
        "text": "The stock market showed significant gains today",
        "category": "finance",
        "confidence": 0.85,
        "reasoning": "Discusses market performance and financial indicators",
    },
]

# Create and compile the classifier
classifier = dspy.ChainOfThought(TextClassification)
multi_optimizer = dspy.BootstrapFewShot(metric=MultiMetric(), max_bootstrapped_demos=2)

compiled_classifier = multi_optimizer.compile(classifier, trainset=classification_data)

# Test the classifier
test_cases = [
    "The company reported Q3 earnings below expectations",
    "The new treatment shows promising results in clinical trials",
]

for text in test_cases:
    result = compiled_classifier(text=text)
    print(f"\nText: {text}")
    print(f"Category: {result.category}")
    print(f"Confidence: {result.confidence}")
    print(f"Reasoning: {result.reasoning}")


# %% Example 4: Teleprompter Pattern
class Teleprompter(dspy.Module):
    """A module that generates and refines responses through multiple stages."""

    def __init__(self):
        super().__init__()

        # Define signature for initial draft
        self.drafter = dspy.ChainOfThought(
            dspy.Signature(
                """
                input_text -> 
                draft: A first attempt at the response
            """
            )
        )

        # Define signature for refinement
        self.refiner = dspy.ChainOfThought(
            dspy.Signature(
                """
                input_text, previous_draft -> 
                improved_draft: An improved version of the response,
                changes_made: Description of improvements
            """
            )
        )

        # Define signature for final polish
        self.polisher = dspy.ChainOfThought(
            dspy.Signature(
                """
                input_text, refined_draft -> 
                final_response: The polished final response,
                confidence: A score between 0 and 1
            """
            )
        )

    def forward(self, input_text: str) -> dict:
        # Generate initial draft
        draft = self.drafter(input_text=input_text)

        # Refine the draft
        refined = self.refiner(input_text=input_text, previous_draft=draft.draft)

        # Polish the response
        final = self.polisher(
            input_text=input_text, refined_draft=refined.improved_draft
        )

        return {
            "response": final.final_response,
            "confidence": final.confidence,
            "refinement_notes": refined.changes_made,
        }


# Create and test the teleprompter
teleprompter = Teleprompter()
complex_input = (
    "Explain the relationship between quantum entanglement and quantum computing"
)

result = teleprompter(complex_input)
print(f"\nInput: {complex_input}")
print(f"Final Response: {result['response']}")
print(f"Confidence: {result['confidence']}")
print(f"Refinement Notes: {result['refinement_notes']}")

# %% Exercise Area
# Exercises to try:
# 1. Create a custom metric for evaluation
# 2. Implement a new optimization strategy
# 3. Build a multi-stage pipeline with bootstrapping
# 4. Design a teleprompter for a specific use case

# Your exercise implementations here:
