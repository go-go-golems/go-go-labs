# Module 1: DSPy Fundamentals - Interactive Exercises
# type: ignore # Ignore missing stubs for dspy
import dspy
import os

# Setup
# Configure DSPy with your preferred LM
# Choose ONE of the following configuration options:

# %%
### Option 1: OpenAI Configuration ###
# Set cache directory (optional)
os.environ["DSPY_CACHEDIR"] = os.path.join(os.getcwd(), "cache")

# Configure LM with explicit cache settings
lm = dspy.LM(
    "openai/gpt-4o-mini", api_key=os.getenv("OPENAI_API_KEY"), cache=True
)  # Set to False to disable caching
dspy.configure(lm=lm)

### Option 2: Local Model Configuration ###
# dspy.configure(lm=dspy.LocalLM(model_path="path/to/local/model"))

# Verify LM is loaded before proceeding
assert (
    dspy.settings.lm is not None
), "Please configure a language model before running predictions"

# %% Example 1: Basic DSPy Signature and Module
# Define a simple signature for question answering
sig = dspy.Signature("question -> answer")

# Create a module that uses direct prediction strategy for this signature
answer_question = dspy.Predict(sig)

# When you have configured an LM above, uncomment to call the module
# result = answer_question(question="Where is the Eiffel Tower located?")
# print(result.answer)

# %% Exercise 1: Define Your Own DSPy Signature and Module
# Choose a task (e.g., sentiment analysis, translation, summarization)
# Define a signature using dspy.Signature
# my_sig = dspy.Signature("input -> output")
my_sig = dspy.Signature("conversation -> insights")
my_module = dspy.Predict(my_sig)
result = my_module(
    conversation="""
                Joe: I'm feeling sad today.
                Mary: I'm sorry to hear that.
                Joe: Thanks for listening.
                Mary: You're welcome.
                Joe: I'm feeling happy today.
                Mary: I'm glad to hear that.
                Joe: Thanks for listening.
                Mary: You're welcome."""
)
print(result.insights)

# Create a module for your signature
# my_module = dspy.Predict(my_sig)

# Test with an example input
# result = my_module(input="Your test input here")
# print(result.output)


# %% More Complex Signature with Multiple Fields
# Define a more complex signature using a class
class ComplexSig(dspy.Signature):
    """Classify text and explain the classification."""

    text = dspy.InputField(desc="The text to classify")
    category = dspy.OutputField(desc="The category: positive, negative, or neutral")
    explanation = dspy.OutputField(desc="Explanation of why this category was chosen")


# Create a module with this signature
complex_classifier = dspy.Predict(ComplexSig)

# Test it when ready
result = complex_classifier(text="I really enjoyed the movie despite its length.")
print(f"Category: {result.category}")
print(f"Explanation: {result.explanation}")

# %% Chain of Thought Example
# Creating a module that produces reasoning before answering
cot_module = dspy.ChainOfThought("question -> answer")

# When ready, test with:
result = cot_module(question="What is the capital of France and what is it known for?")
print("Reasoning:", result.reasoning)
print("Answer:", result.answer)

# %%
