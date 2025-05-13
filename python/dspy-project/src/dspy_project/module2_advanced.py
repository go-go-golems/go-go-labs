# Module 2: Advanced DSPy Concepts - Interactive Exercises
# %%

# type: ignore # Ignore missing stubs for dspy
import dspy  # type: ignore

import time
import os
from typing import Literal

# Setup - reuse configuration from module1
# Set cache directory (optional)
os.environ["DSPY_CACHEDIR"] = os.path.join(os.getcwd(), "cache")

# Configure LM with explicit cache settings
lm = dspy.LM(
    "openai/gpt-4o-mini", api_key=os.getenv("OPENAI_API_KEY"), cache=False
)  # Set to False to disable caching
dspy.configure(lm=lm)

# Verify LM is loaded before proceeding
assert (
    dspy.settings.lm is not None
), "Please configure a language model before running predictions"


# %% Example 1: Advanced Signature Patterns
# Define a detailed QA signature with typed fields and validation
class DetailedQASig(dspy.Signature):
    """A signature for detailed question answering with metadata."""

    question: str = dspy.InputField(desc="The user's question")
    context: str = dspy.InputField(desc="Additional context or background information")
    max_length: int = dspy.InputField(
        desc="Maximum length of the answer in words", default=100
    )

    answer: str = dspy.OutputField(desc="The answer to the question")
    confidence: float = dspy.OutputField(desc="Confidence score between 0 and 1")
    sources: list[str] = dspy.OutputField(desc="List of sources used in the answer")


# Create and test the detailed QA module
qa_module = dspy.ChainOfThought(DetailedQASig)
result = qa_module(
    question="What is quantum computing?",
    context="Quantum computing is a type of computation that harnesses quantum mechanical phenomena like superposition and entanglement to perform calculations. It uses qubits instead of classical bits.",
    max_length=150,
)
print(f"Answer: {result.answer}")
print(f"Confidence: {result.confidence}")
print(f"Sources: {result.sources}")


# %% Example 2: Nested Signatures for Research
class ResearchStep(dspy.Signature):
    """A single research step."""

    query: str = dspy.InputField()
    findings: str = dspy.OutputField()
    next_questions: list[str] = dspy.OutputField()


class ResearchProcess(dspy.Signature):
    """A multi-step research process."""

    initial_question: str = dspy.InputField()
    max_steps: int = dspy.InputField(default=3)

    steps: list[ResearchStep] = dspy.OutputField()
    final_answer: str = dspy.OutputField()


# Create and test the research process
research_module = dspy.ChainOfThought(ResearchProcess)
result = research_module(
    initial_question="How does climate change affect biodiversity?", max_steps=2
)
print("Final Answer:", result.final_answer)
print("\nResearch Steps:")
for i, step in enumerate(result.steps, 1):
    print(f"\nStep {i}:")
    print(f"Query: {step.query}")
    print(f"Findings: {step.findings}")
    print(f"Next Questions: {', '.join(step.next_questions)}")


# %% Example 3: Pipeline Pattern Implementation
class ResearchPipeline:
    def __init__(self):
        self.retriever = dspy.Predict(dspy.Signature("query -> relevant_docs"))
        self.analyzer = dspy.ChainOfThought(
            dspy.Signature("docs, query -> key_points")
        )  # type: ignore
        self.synthesizer = dspy.ChainOfThought(
            dspy.Signature("key_points, query -> final_answer")
        )  # type: ignore

    def research(self, query: str) -> str:
        # Step 1: Retrieve relevant documents
        docs_result = self.retriever(query=query)

        # Step 2: Analyze documents for key points
        analysis = self.analyzer(docs=docs_result.relevant_docs, query=query)

        # Step 3: Synthesize final answer
        final = self.synthesizer(key_points=analysis.key_points, query=query)

        return final.final_answer


# Test the pipeline
pipeline = ResearchPipeline()
answer = pipeline.research("What are the main causes of ocean acidification?")
print("Pipeline Answer:", answer)


# %% Example 4: Smart Router with Branching Logic
def calculate(expression: str) -> float:
    """Simple calculator tool."""
    return eval(expression)


class QueryClassificationType(dspy.Signature):
    query_type: Literal["factual", "opinion", "calculation"] = dspy.OutputField(
        desc="The type of query"
    )
    query: str = dspy.InputField(desc="The query to classify")


class SmartRouter:
    def __init__(self):
        self.classifier = dspy.Predict(QueryClassificationType)
        self.factual_qa = dspy.ChainOfThought("question -> answer")
        self.opinion_qa = dspy.Predict("question -> opinion")
        self.calculator = dspy.ReAct("question -> result", tools=[calculate])

    def answer(self, query: str) -> str:
        print(f"Answer Query: {query}\n")

        # Determine query type
        query_type = self.classifier(query=query).query_type
        print(f"Query Type: {query_type}\n")

        # Route to appropriate module
        if query_type == "factual":
            return self.factual_qa(question=query).answer
        elif query_type == "opinion":
            return self.opinion_qa(question=query).opinion
        elif query_type == "calculation":
            return self.calculator(question=query).result
        else:
            return "I'm not sure how to handle this type of question."


# Test the router with different types of questions
router = SmartRouter()
questions = [
    "What is the capital of France?",  # factual
    "Do you think AI will benefit humanity?",  # opinion
    "What is 123 * 456?",  # calculation
]

for q in questions:
    print(f"\nQuestion: {q}")
    print(f"Answer: {router.answer(q)}")


# %% Example 5: Conversational Agent with Memory
class SmartChatbot:
    def __init__(self):
        self.memory = dspy.History(messages=[])
        self.summarizer = dspy.Predict("conversation -> summary")
        self.responder = dspy.ChainOfThought("history, question -> response")

    def chat(self, user_message: str) -> str:
        # If conversation is getting long, summarize older messages
        if len(self.memory.messages) > 5:
            old_messages = self.memory.messages[:-3]  # Keep last 3 messages fresh
            summary = self.summarizer(
                conversation="\n".join(str(m) for m in old_messages)
            ).summary
            print(f"Summary: {summary}\n")

            # Replace old messages with summary
            new_messages = [
                {
                    "role": "system",
                    "content": f"Previous conversation summary: {summary}",
                }
            ] + self.memory.messages[-3:]
            self.memory = dspy.History(messages=new_messages)

        # Get response using history
        response = self.responder(history=self.memory, question=user_message)

        # Update memory
        self.memory.messages.append({"role": "user", "content": user_message})
        self.memory.messages.append({"role": "assistant", "content": response.response})

        return response.response


# Test the chatbot with a conversation
chatbot = SmartChatbot()
conversation = [
    "Hi, my name is Alice!",
    "What's the weather like today?",
    "I live in New York.",
    "What city did I say I live in?",
    "What is the capital of France?",
    "How do you feel about AI?",
    "I like computers",
]

for msg in conversation:
    print(f"\nUser: {msg}")
    response = chatbot.chat(msg)
    print(f"Bot: {response}")

print("\nConversation History:", chatbot.memory.messages)


# %% Example 6: Debugging and Testing
class DebuggedAgent:
    def __init__(self):
        self.module = dspy.ChainOfThought("question -> answer")
        self.log_history = []

    def process(self, question: str) -> str:
        # Log input
        self.log_history.append(
            {"timestamp": time.time(), "input": question, "type": "input"}
        )

        try:
            # Run prediction
            result = self.module(question=question)

            # Log successful output
            self.log_history.append(
                {
                    "timestamp": time.time(),
                    "output": result.answer,
                    "reasoning": getattr(result, "reasoning", None),
                    "type": "output",
                }
            )

            return result.answer

        except Exception as e:
            # Log errors
            self.log_history.append(
                {"timestamp": time.time(), "error": str(e), "type": "error"}
            )
            raise

    def get_debug_info(self) -> dict:
        """Get debugging information about recent interactions."""
        return {
            "total_calls": len([x for x in self.log_history if x["type"] == "input"]),
            "success_rate": len([x for x in self.log_history if x["type"] == "output"])
            / max(len([x for x in self.log_history if x["type"] == "input"]), 1),
            "recent_errors": [x for x in self.log_history if x["type"] == "error"][-5:],
            "last_successful": next(
                (x for x in reversed(self.log_history) if x["type"] == "output"), None
            ),
        }


# Test the debugged agent
debug_agent = DebuggedAgent()

# Try some successful queries
for question in ["What is the sun?", "How do birds fly?"]:
    try:
        answer = debug_agent.process(question)
        print(f"\nQuestion: {question}")
        print(f"Answer: {answer}")
    except Exception as e:
        print(f"Error: {e}")

# Try an invalid query to test error handling
try:
    debug_agent.process("")
except Exception as e:
    print("\nCaught expected error for empty input")

# Print debug info
print("\nDebug Info:", debug_agent.get_debug_info())

# %% Exercise Area
# Use this cell to work on the chapter exercises:
# 1. Advanced Signature Design
# 2. Module Composition
# 3. State Management
# 4. Debugging and Testing

# Your exercise implementations here:


# %% Exercise: Advanced Router Agent with Multi-Step Task Handling
class WikiData:
    """Mock Wikipedia data for demonstration purposes."""

    data = {
        "quantum computing": "Quantum computing is a type of computing that harnesses quantum mechanical phenomena like superposition and entanglement to perform calculations.",
        "eiffel tower": "The Eiffel Tower is a wrought-iron lattice tower located in Paris, France. It was constructed from 1887 to 1889 as the entrance to the 1889 World's Fair.",
        "albert einstein": "Albert Einstein (14 March 1879 – 18 April 1955) was a German-born theoretical physicist who is widely held to be one of the greatest and most influential scientists of all time.",
        "python programming": "Python is a high-level, general-purpose programming language. Its design philosophy emphasizes code readability with the use of significant indentation.",
    }

    @classmethod
    def search(cls, query: str) -> str:
        """Search the mock Wikipedia database."""
        query = query.lower()
        for term, content in cls.data.items():
            if term in query or query in term:
                return content
        return "No information found for this query."


class WeatherData:
    """Mock weather data for demonstration purposes."""

    data = {
        "new york": {"temperature": 22, "condition": "Sunny", "humidity": 60},
        "london": {"temperature": 15, "condition": "Rainy", "humidity": 75},
        "tokyo": {"temperature": 28, "condition": "Cloudy", "humidity": 65},
        "paris": {"temperature": 20, "condition": "Clear", "humidity": 55},
        "sydney": {"temperature": 26, "condition": "Windy", "humidity": 50},
    }

    @classmethod
    def get_weather(cls, location: str) -> str:
        """Get weather information for a location."""
        location = location.lower()
        if location in cls.data:
            weather = cls.data[location]
            return f"Weather in {location.title()}: {weather['temperature']}°C, {weather['condition']}, Humidity: {weather['humidity']}%"
        return f"Weather information for {location} is not available."


class TranslationTool:
    """Mock translation tool for demonstration purposes."""

    translations = {
        ("hello", "french"): "Bonjour",
        ("hello", "spanish"): "Hola",
        ("hello", "german"): "Hallo",
        ("thank you", "french"): "Merci",
        ("thank you", "spanish"): "Gracias",
        ("thank you", "german"): "Danke",
        ("goodbye", "french"): "Au revoir",
        ("goodbye", "spanish"): "Adiós",
        ("goodbye", "german"): "Auf Wiedersehen",
    }

    @classmethod
    def translate(cls, text: str, language: str) -> str:
        """Translate text to the specified language."""
        text = text.lower()
        language = language.lower()

        if (text, language) in cls.translations:
            return (
                f"'{text}' in {language.title()}: {cls.translations[(text, language)]}"
            )
        return f"Translation of '{text}' to {language} is not available."


class Tool:
    """Represents a tool that can be called by the agent."""

    def __init__(self, name, function, description, input_schema):
        """
        Initialize a tool.

        Args:
            name: The name of the tool
            function: The function to call
            description: A description of what the tool does
            input_schema: A dictionary mapping parameter names to descriptions
        """
        self.name = name
        self.function = function
        self.description = description
        self.input_schema = input_schema

    def __call__(self, **kwargs):
        """Call the tool function with the provided parameters."""
        return self.function(**kwargs)

    def get_description(self):
        """Get a detailed description of the tool including its parameters."""
        param_desc = "\n".join(
            [f"  - {name}: {desc}" for name, desc in self.input_schema.items()]
        )
        return f"{self.name}: {self.description}\nParameters:\n{param_desc}"


# Define functions for tools
def calculate(expression: str) -> str:
    """Evaluate a mathematical expression."""
    try:
        result = eval(expression)
        return f"The result of {expression} is {result}"
    except Exception as e:
        return f"Error evaluating expression: {e}"


def search_wiki(query: str) -> str:
    """Search Wikipedia for information."""
    return WikiData.search(query)


def get_weather(location: str) -> str:
    """Get weather information for a location."""
    return WeatherData.get_weather(location)


def translate_text(text: str, target_language: str) -> str:
    """Translate text to the specified language."""
    return TranslationTool.translate(text, target_language)


def task_finished(reason: str) -> str:
    """Mark the task as complete with a reason."""
    return f"Task completed: {reason}"


# Create Tool instances
calculator_tool = Tool(
    name="calculator",
    function=calculate,
    description="Evaluates mathematical expressions",
    input_schema={
        "expression": "The mathematical expression to evaluate (e.g., '2 + 2', '10 * 5', 'math.sqrt(144)')"
    },
)

wiki_tool = Tool(
    name="wiki",
    function=search_wiki,
    description="Searches for information in a Wikipedia-like database",
    input_schema={"query": "The topic or question to search for information about"},
)

weather_tool = Tool(
    name="weather",
    function=get_weather,
    description="Gets current weather information for a location",
    input_schema={"location": "The name of the city or location to get weather for"},
)

translator_tool = Tool(
    name="translator",
    function=translate_text,
    description="Translates simple phrases from English to another language",
    input_schema={
        "text": "The English text to translate",
        "target_language": "The target language (currently supports 'french', 'spanish', 'german')",
    },
)

finished_tool = Tool(
    name="finished",
    function=task_finished,
    description="Marks the task as complete when all required information has been gathered",
    input_schema={"reason": "Explanation of why the task is considered complete"},
)

# %%


class RouterSig(dspy.Signature):
    query: str = dspy.InputField(desc="The user's query or current subtask")
    task_history: str = dspy.InputField(desc="The history of actions taken so far")
    tools_available: str = dspy.InputField(
        desc="Descriptions of available tools and their parameters"
    )
    summary_of_previous_tool_result: str = dspy.OutputField(
        desc="A brief summary of the previous tool result (optional)", required=False
    )
    tool_name: Literal["calculator", "wiki", "weather", "translator", "finished"] = (
        dspy.OutputField(desc="The name of the tool to use")
    )
    tool_params: str = dspy.OutputField(
        desc="Parameters to pass to the tool as a JSON object matching the tool's required parameters"
    )


class AdvancedRouterAgent:
    """Agent that uses a router to select between specialized tools for solving multi-step tasks."""

    def __init__(self, tools=None):
        # Define the set of available tools
        self.tools = tools or {
            "calculator": calculator_tool,
            "wiki": wiki_tool,
            "weather": weather_tool,
            "translator": translator_tool,
            "finished": finished_tool,
        }

        # Create a description of all available tools for the router
        tools_descriptions = "\n\n".join(
            [tool.get_description() for tool in self.tools.values()]
        )

        # Initialize the router module
        self.router = dspy.ChainOfThought(RouterSig)

        # Store tools description for router
        self.tools_descriptions = tools_descriptions

        # Keep a history of actions
        self.task_history = []

        # Store previous results for summarization
        self.previous_results = []

    def solve(self, query: str, max_steps: int = 4) -> str:
        """Solve a complex task that may require multiple tool calls."""
        print(f"Starting task: {query}\n")

        # Initialize task state
        current_query = query

        # Execute up to max_steps iterations
        for step in range(max_steps):
            print(f"Step {step+1}: Routing query: {current_query}")

            # Determine which tool to use
            task_history_str = "\n".join(
                [f"- {action}" for action in self.task_history]
            )

            # Get the most recent result to summarize if available
            previous_result = (
                self.previous_results[-1] if self.previous_results else None
            )

            route = self.router(
                query=current_query,
                task_history=task_history_str,
                tools_available=self.tools_descriptions,
            )

            print(f"Route: {route}")
            tool_name = route.tool_name
            tool_params = route.tool_params
            summary = getattr(route, "summary_of_previous_tool_result", None)

            # Log the summary if available
            if summary:
                print(f"Summary of previous result: {summary}")

            print(f"Tool selected: {tool_name}")
            print(f"Tool parameters: {tool_params}\n")

            # Check if the selected tool exists
            if tool_name in self.tools:
                tool = self.tools[tool_name]

                # Parse the tool parameters
                try:
                    params = eval(tool_params) if tool_params.strip() else {}
                except Exception as e:
                    error_msg = f"Error parsing tool parameters: {e}"
                    self.task_history.append(f"ERROR: {error_msg}")
                    print(f"Error: {error_msg}\n")
                    current_query = f"There was an error parsing parameters: {error_msg}. Please provide valid parameters for tool '{tool_name}' to complete the task: '{query}'"
                    continue

                # Handle task completion
                if tool_name == "finished":
                    result = tool(reason=params.get("reason", "Task complete"))
                    self.task_history.append(f"FINISHED: {result}")
                    self.previous_results.append(result)
                    print(f"Result: {result}")
                    return result

                # Execute the selected tool
                try:
                    if tool_name == "calculator":
                        result = tool(expression=params.get("expression", ""))
                    elif tool_name == "wiki":
                        result = tool(query=params.get("query", ""))
                    elif tool_name == "weather":
                        result = tool(location=params.get("location", ""))
                    elif tool_name == "translator":
                        result = tool(
                            text=params.get("text", ""),
                            target_language=params.get("target_language", ""),
                        )
                    else:
                        # Generic call for future extensibility
                        result = tool(**params)

                    # Record the action in history
                    self.task_history.append(f"{tool_name.upper()}: {result}")
                    self.previous_results.append(result)
                    print(f"Result: {result}\n")

                    # Update the query for the next step
                    current_query = f"Based on the previous result '{result}', what should I do next to complete the original task: '{query}'?"

                except Exception as e:
                    error_msg = f"Error executing tool {tool_name}: {e}"
                    self.task_history.append(f"ERROR: {error_msg}")
                    print(f"Error: {error_msg}\n")
                    current_query = f"There was an error: {error_msg}. Please try a different approach to complete the original task: '{query}'"
            else:
                error_msg = f"Unknown tool: {tool_name}. Available tools are: {', '.join(self.tools.keys())}"
                self.task_history.append(f"ERROR: {error_msg}")
                print(f"Error: {error_msg}\n")
                current_query = f"There was an error: {error_msg}. Please select a valid tool to complete the original task: '{query}'"

            # Check if we should continue or exit
            if tool_name == "finished":
                break

        # If we've reached max_steps without finishing
        return f"Task not completed within {max_steps} steps. Progress so far: {'; '.join(self.task_history)}"


# Test the advanced router agent with a multi-step task
router_agent = AdvancedRouterAgent()

# Example multi-step task that requires multiple tools
complex_task = "What is the weather in Paris, and how do you say 'thank you' in French? Also, calculate 25 * 4."

result = router_agent.solve(complex_task)
print(f"\nFinal result: {result}")

# Test with a multi-step task that requires history
multi_step_task = (
    "First, check the weather in London. Then, translate 'hello' to Spanish."
)
result = router_agent.solve(multi_step_task)
print(f"\nFinal result for multi-step task: {result}")
print(f"\nHistory of actions: {router_agent.task_history}")
print(f"\nPrevious results: {router_agent.previous_results}")

# # Another example requiring wiki lookup and calculation
# research_task = "Find information about quantum computing and then calculate the square root of 144."

# result = router_agent.solve(research_task)
# print(f"\nFinal result: {result}")

# %%
