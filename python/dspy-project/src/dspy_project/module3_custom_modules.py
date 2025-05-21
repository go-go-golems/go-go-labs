# Module 3: Custom DSPy Modules - Interactive Exercises
# %%

# type: ignore # Ignore missing stubs for dspy
import dspy  # type: ignore

import os
import typing
import time
from typing import List, Dict, Any, Optional, Tuple, Literal, Union

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


# %% Example 1: Understanding the Module Base Class
# All DSPy modules inherit from dspy.Module
# Let's look at a simple custom module that inherits directly from Module


class SimpleQA(dspy.Module):
    """A simple question answering module.

    This module takes a question and returns an answer. It uses a simple
    Predict module internally.
    """

    def __init__(self):
        super().__init__()
        # Initialize sub-modules in __init__
        self.predictor = dspy.Predict("question -> answer")

    def forward(self, question: str) -> dspy.Prediction:
        """Forward method defines the module's behavior."""
        # Call the internal predictor module
        return self.predictor(question=question)


# Test the simple module
simple_qa = SimpleQA()
result = simple_qa(question="What is the capital of France?")
print(f"Simple QA Answer: {result.answer}")

# You can also check the usage statistics
print("\nLM Usage:")
print(result.get_lm_usage())


# %% Example 2: Custom Module with Multiple Stages
# Let's create a more complex module that breaks a task into multiple stages


class TwoStageQA(dspy.Module):
    """A two-stage QA module that first extracts key points from the question,
    then answers based on those key points.
    """

    def __init__(self):
        super().__init__()
        # Define the two sub-modules we'll use
        self.key_point_extractor = dspy.ChainOfThought(
            "question -> key_points: list[str]"
        )
        self.final_answerer = dspy.ChainOfThought(
            "question, key_points: list[str] -> answer"
        )

    def forward(self, question: str) -> dspy.Prediction:
        """This method implements the two-stage process."""
        # Stage 1: Extract key points from the question
        key_points_result = self.key_point_extractor(question=question)

        # Stage 2: Generate final answer using key points
        final_result = self.final_answerer(
            question=question, key_points=key_points_result.key_points
        )

        # Return a Prediction combining results from both stages
        return dspy.Prediction(
            key_points=key_points_result.key_points, answer=final_result.answer
        )


# Test the two-stage module
two_stage_qa = TwoStageQA()
result = two_stage_qa(
    question="What are the main factors contributing to climate change?"
)
print(f"Key Points: {result.key_points}")
print(f"Two-Stage Answer: {result.answer}")


# %% Example 3: Custom Module with Conditional Logic
# This module demonstrates how to incorporate conditional logic within a module


class ConditionalQA(dspy.Module):
    """A module that uses different answering strategies depending on question complexity."""

    def __init__(self):
        super().__init__()
        # Define classifier to determine question complexity
        self.classifier = dspy.Predict("question -> complexity: str")

        # Define different responders for different complexity levels
        self.simple_responder = dspy.Predict("question -> answer")
        self.complex_responder = dspy.ChainOfThought("question -> answer")

    def forward(self, question: str) -> dspy.Prediction:
        """Forward method with conditional branching based on question complexity."""
        # First, classify the question
        classification = self.classifier(question=question)
        complexity = classification.complexity

        # Choose responder based on complexity
        if complexity.lower() in ["simple", "basic", "easy"]:
            result = self.simple_responder(question=question)
            method_used = "simple"
        else:  # complex, difficult, etc.
            result = self.complex_responder(question=question)
            method_used = "complex"

        # Return the answer along with metadata about how it was processed
        return dspy.Prediction(
            answer=result.answer, complexity=complexity, method_used=method_used
        )


# Test the conditional module
conditional_qa = ConditionalQA()

# Try with simple and complex questions
simple_q = "What is the color of the sky?"
complex_q = "What are the implications of quantum entanglement for information theory?"

simple_result = conditional_qa(question=simple_q)
complex_result = conditional_qa(question=complex_q)

print(f"\nSimple Question: {simple_q}")
print(f"Complexity: {simple_result.complexity}")
print(f"Method Used: {simple_result.method_used}")
print(f"Answer: {simple_result.answer}")

print(f"\nComplex Question: {complex_q}")
print(f"Complexity: {complex_result.complexity}")
print(f"Method Used: {complex_result.method_used}")
print(f"Answer: {complex_result.answer}")


# %% Example 4: Custom Module with Retry Logic and Error Handling
# This example shows how to build in resilience


class RobustQA(dspy.Module):
    """A module that implements retry logic and validation for QA."""

    def __init__(self, max_retries=2):
        super().__init__()
        self.max_retries = max_retries
        self.qa_model = dspy.ChainOfThought("question -> answer")
        self.validator = dspy.Predict(
            "question, answer -> is_valid: bool, feedback: str"
        )

    def forward(self, question: str) -> dspy.Prediction:
        """Forward method with built-in validation and retry logic."""
        attempts = 0
        validation_history = []

        while attempts <= self.max_retries:
            # Get an answer
            result = self.qa_model(question=question)
            answer = result.answer

            # Validate the answer
            validation = self.validator(question=question, answer=answer)
            is_valid = validation.is_valid
            feedback = validation.feedback

            # Store validation result
            validation_history.append(
                {
                    "attempt": attempts + 1,
                    "answer": answer,
                    "is_valid": is_valid,
                    "feedback": feedback,
                }
            )

            # If valid, return the result
            if is_valid:
                return dspy.Prediction(
                    answer=answer,
                    validation_history=validation_history,
                    attempts=attempts + 1,
                    valid=True,
                )

            # If we've reached max retries, return the last result
            if attempts == self.max_retries:
                break

            # Increment attempts counter
            attempts += 1

        # If we get here, we've run out of retries
        return dspy.Prediction(
            answer=answer,  # Return the last answer even though invalid
            validation_history=validation_history,
            attempts=attempts + 1,
            valid=False,
        )


# Test the robust module
robust_qa = RobustQA(max_retries=2)

# Try with a question that might need validation
result = robust_qa(question="What is the population of the United States?")
print("\nRobust QA Result:")
print(f"Final Answer: {result.answer}")
print(f"Valid: {result.valid}")
print(f"Attempts: {result.attempts}")
print("Validation History:")
for entry in result.validation_history:
    print(
        f"  Attempt {entry['attempt']}: {'Valid' if entry['is_valid'] else 'Invalid'}"
    )
    print(f"  Feedback: {entry['feedback']}")


# %% Example 5: Custom Module with Memory
# This example shows how to maintain state between calls


class StatefulQA(dspy.Module):
    """A QA module that maintains context from previous interactions."""

    def __init__(self):
        super().__init__()
        self.qa_with_context = dspy.ChainOfThought("question, context -> answer")
        self.context_updater = dspy.Predict(
            "previous_context, question, answer -> new_context"
        )
        # Initialize empty context
        self.current_context = "No previous context available."

    def forward(self, question: str) -> dspy.Prediction:
        """Forward method that maintains and updates context."""
        # Use current context to answer the question
        result = self.qa_with_context(question=question, context=self.current_context)

        # Update the context based on this interaction
        context_update = self.context_updater(
            previous_context=self.current_context,
            question=question,
            answer=result.answer,
        )

        # Store the updated context for next time
        self.current_context = context_update.new_context

        # Return the result with the current context
        return dspy.Prediction(answer=result.answer, context=self.current_context)


# Test the stateful module with a sequence of related questions
stateful_qa = StatefulQA()

# Ask a sequence of related questions
questions = [
    "My name is Alice. What's your name?",
    "What's my name?",
    "I live in New York. Do you know where that is?",
    "Where do I live?",
]

for i, question in enumerate(questions, 1):
    print(f"\nQuestion {i}: {question}")
    result = stateful_qa(question=question)
    print(f"Answer: {result.answer}")
    print(f"Current Context: {result.context}")


# %% Example 6: Multi-Modal Module (Text to Structure)
# This example shows how to create a module that converts unstructured text to structured data


class EntityExtractor(dspy.Module):
    """A module that extracts structured entities from text."""

    def __init__(self, entity_types=None):
        super().__init__()
        self.entity_types = entity_types or [
            "person",
            "organization",
            "location",
            "date",
        ]

        # Define the signature with dynamic entity types
        signature_str = f"text -> entities: list[dict[str, str]]"
        self.extractor = dspy.ChainOfThought(signature_str)

    def forward(self, text: str) -> dspy.Prediction:
        """Extract entities from text."""
        # Prepare prompt with entity type instructions
        entity_instruction = (
            f"Extract entities of types: {', '.join(self.entity_types)}"
        )

        # Extract entities
        result = self.extractor(text=text)

        # Process and validate entities
        processed_entities = []
        for entity in result.entities:
            # Ensure each entity has a name and type
            if isinstance(entity, dict) and "name" in entity and "type" in entity:
                # Only keep entities of requested types
                if entity["type"].lower() in [t.lower() for t in self.entity_types]:
                    processed_entities.append(entity)

        # Return the processed entities
        return dspy.Prediction(
            entities=processed_entities, entity_count=len(processed_entities)
        )


# Test the entity extractor
entity_extractor = EntityExtractor(
    entity_types=["person", "organization", "location", "date"]
)

sample_text = """
Apple Inc. announced today that CEO Tim Cook will be visiting their new campus in Cupertino 
on January 15, 2023. The event will also be attended by former CEO Steve Jobs' widow, 
Laurene Powell Jobs, and representatives from Microsoft.
"""

result = entity_extractor(text=sample_text)
print("\nExtracted Entities:")
for entity in result.entities:
    print(f"  {entity['name']} - {entity['type']}")
print(f"Total entities: {result.entity_count}")


# %% Example 7: Custom Module with Parallel Processing
# This example shows how to implement a module that runs multiple sub-modules in parallel


class EnsembleQA(dspy.Module):
    """A module that uses multiple QA strategies and ensembles their results."""

    def __init__(self):
        super().__init__()
        # Define multiple answering strategies
        self.basic_qa = dspy.Predict("question -> answer")
        self.cot_qa = dspy.ChainOfThought("question -> answer")
        self.pot_qa = dspy.ProgramOfThought("question -> answer")

        # Define a judge to select or combine answers
        self.judge = dspy.ChainOfThought(
            "question, answers: list[str] -> final_answer, reasoning"
        )

    def forward(self, question: str) -> dspy.Prediction:
        """Run multiple QA strategies and ensemble the results."""
        # Get answers from each strategy
        basic_result = self.basic_qa(question=question)
        cot_result = self.cot_qa(question=question)
        pot_result = self.pot_qa(question=question)

        # Collect all answers
        all_answers = [basic_result.answer, cot_result.answer, pot_result.answer]

        # Let the judge choose the best answer
        judge_result = self.judge(question=question, answers=all_answers)

        # Return the final answer with metadata
        return dspy.Prediction(
            answer=judge_result.final_answer,
            reasoning=judge_result.reasoning,
            all_candidate_answers=all_answers,
        )


# Test the ensemble QA module
ensemble_qa = EnsembleQA()
result = ensemble_qa(question="How many planets are in our solar system?")
print("\nEnsemble QA Result:")
print(f"Final Answer: {result.answer}")
print(f"Reasoning: {result.reasoning}")
print("All Candidate Answers:")
for i, answer in enumerate(result.all_candidate_answers, 1):
    print(f"  {i}. {answer}")


# %% Example 8: Custom Module with External Tool Integration
# This example shows how to create a module that uses external tools


def get_weather(location: str) -> str:
    """Mock weather API - would be a real API call in production."""
    # In real usage, this would call an external weather API
    weather_data = {
        "new york": "72°F, Sunny",
        "london": "61°F, Rainy",
        "tokyo": "78°F, Cloudy",
        "paris": "65°F, Clear",
        "sydney": "82°F, Partly Cloudy",
    }
    location = location.lower()
    return weather_data.get(location, f"Weather data not available for {location}")


def calculate(expression: str) -> str:
    """Evaluate a math expression."""
    try:
        result = eval(expression)
        return f"The result of {expression} is {result}"
    except Exception as e:
        return f"Error evaluating expression: {e}"


class ToolAugmentedQA(dspy.Module):
    """A module that can use external tools based on the question type."""

    def __init__(self):
        super().__init__()
        # Define a classifier to determine question type
        self.classifier = dspy.Predict("question -> question_type: str")

        # Define a QA model that can use tools
        self.tool_qa = dspy.ReAct("question -> answer", tools=[get_weather, calculate])

        # Define a standard QA for questions that don't need tools
        self.standard_qa = dspy.ChainOfThought("question -> answer")

    def forward(self, question: str) -> dspy.Prediction:
        """Answer questions, using tools when appropriate."""
        # Classify the question
        classification = self.classifier(question=question)
        question_type = classification.question_type.lower()

        # Determine if tools are needed
        if any(
            keyword in question_type
            for keyword in ["weather", "temperature", "math", "calculate"]
        ):
            # Use tool-augmented QA
            result = self.tool_qa(question=question)
            used_tools = True
        else:
            # Use standard QA
            result = self.standard_qa(question=question)
            used_tools = False

        # Return the result with metadata
        return dspy.Prediction(
            answer=result.answer, question_type=question_type, used_tools=used_tools
        )


# Test the tool-augmented QA module
tool_qa = ToolAugmentedQA()

# Try with different types of questions
weather_q = "What's the weather in London?"
math_q = "What is 24 * 7 divided by 3?"
general_q = "Who wrote Romeo and Juliet?"

weather_result = tool_qa(question=weather_q)
math_result = tool_qa(question=math_q)
general_result = tool_qa(question=general_q)

print("\nTool-Augmented QA Results:")
print(f"\nWeather Question: {weather_q}")
print(f"Question Type: {weather_result.question_type}")
print(f"Used Tools: {weather_result.used_tools}")
print(f"Answer: {weather_result.answer}")

print(f"\nMath Question: {math_q}")
print(f"Question Type: {math_result.question_type}")
print(f"Used Tools: {math_result.used_tools}")
print(f"Answer: {math_result.answer}")

print(f"\nGeneral Question: {general_q}")
print(f"Question Type: {general_result.question_type}")
print(f"Used Tools: {general_result.used_tools}")
print(f"Answer: {general_result.answer}")


# %% Exercise Area
# Use this cell to experiment with your own custom modules.
# Some ideas:
# 1. Create a summarizer module with different levels of detail
# 2. Build a module that can do multi-hop reasoning
# 3. Create a module that generates and validates structured data
# 4. Build a module that can handle multi-modal inputs (e.g., text + images)


# %% Exercise: Custom Multi-hop Research Agent
# Let's create a research agent that can break a complex question
# into sub-questions, research each one, and synthesize a final answer


class ResearchAgent(dspy.Module):
    """A multi-hop research agent that can decompose and investigate complex questions."""

    def __init__(self, max_hops=3):
        super().__init__()
        self.max_hops = max_hops

        # Define sub-modules for different steps of the research process
        self.question_decomposer = dspy.ChainOfThought(
            "main_question -> sub_questions: list[str]"
        )

        self.sub_question_answerer = dspy.ChainOfThought(
            "question -> answer, confidence: float"
        )

        self.synthesizer = dspy.ChainOfThought(
            "main_question, research_findings: list[dict[str, str]] -> final_answer"
        )

    def forward(self, question: str) -> dspy.Prediction:
        """Conduct multi-hop research on a complex question."""
        # Step 1: Decompose the main question into sub-questions
        decomposition = self.question_decomposer(main_question=question)
        sub_questions = decomposition.sub_questions

        # Limit the number of sub-questions to max_hops
        sub_questions = sub_questions[: self.max_hops]

        # Step 2: Research each sub-question
        research_findings = []
        for i, sub_q in enumerate(sub_questions):
            # Answer the sub-question
            sub_result = self.sub_question_answerer(question=sub_q)

            # Store the finding
            finding = {
                "sub_question": sub_q,
                "answer": sub_result.answer,
                "confidence": sub_result.confidence,
            }
            research_findings.append(finding)

        # Step 3: Synthesize findings into a final answer
        synthesis = self.synthesizer(
            main_question=question, research_findings=research_findings
        )

        # Return the complete research package
        return dspy.Prediction(
            final_answer=synthesis.final_answer,
            sub_questions=sub_questions,
            research_findings=research_findings,
        )


# Test the research agent
research_agent = ResearchAgent(max_hops=3)
complex_question = "How has climate change affected marine ecosystems, and what are the potential long-term impacts?"

result = research_agent(question=complex_question)

print("\nResearch Agent Results:")
print(f"Main Question: {complex_question}")
print("\nSub-Questions:")
for i, sq in enumerate(result.sub_questions, 1):
    print(f"  {i}. {sq}")

print("\nResearch Findings:")
for i, finding in enumerate(result.research_findings, 1):
    print(f"  Finding {i}:")
    print(f"    Sub-Question: {finding['sub_question']}")
    print(f"    Answer: {finding['answer']}")
    print(f"    Confidence: {finding['confidence']}")

print(f"\nFinal Answer: {result.final_answer}")

# %%
