# Advanced Course: Building Agents with DSPy (Declarative Self-Improving Python)

## Module 1: DSPy Fundamentals Refresher

DSPy (Declarative Self-improving Python) is a framework for *programming* language models through modular code, rather than ad-hoc prompt strings. It provides high-level abstractions that let you define what an AI component should do (in natural language) and then handles how to prompt the model for you. In this module, we review DSPy’s core concepts: **Signatures**, **Modules**, **Compilation**, and **Teleprompters** (optimizers). These abstractions make it easier to build reliable AI systems by separating the *specification* of a task from the *prompt engineering details*. We will briefly revisit each concept with examples before diving into advanced topics.

* **Signatures:** A signature in DSPy is a concise, declarative specification of a task’s input and output interface. It’s like a function signature in natural language – describing *what* transformation we want (e.g. “question -> answer”), without detailing *how* to prompt the model. Signatures define input fields and output fields (with types) for an AI module. For example, a signature `"sentence -> sentiment"` tells DSPy we will input a sentence and expect a sentiment label in return. You can define signatures inline as strings or as Python classes with `InputField` and `OutputField` attributes for more complex tasks (allowing docstrings, field descriptions, and type hints).

* **Modules:** A module is the basic building block in DSPy that encapsulates a particular prompting or reasoning strategy. Each module is associated with a signature and knows how to use a language model (LM) to fulfill that signature’s task. DSPy includes standard modules like `dspy.Predict` for direct prompt-and-answer, `dspy.ChainOfThought` for chain-of-thought reasoning, `dspy.ReAct` for reasoning and tool use, and others. A module essentially replaces manual prompts with a reusable, parameterizable component. For example, `dspy.Predict("sentence -> sentiment")` creates a sentiment classification module. When called with an input, the module will construct the appropriate prompt (based on the signature) and parse the model’s response into the structured output. Modules can be composed together, allowing complex agent behaviors to be built from simpler pieces. In fact, `dspy.Predict` is the fundamental module that other modules build upon.

* **Compilation:** In DSPy, *compilation* refers to the process of transforming your high-level DSPy program into an effective prompt pipeline and (optionally) optimized model parameters. Initially, given your code and signatures, DSPy will compile it into a prompt template or sequence of prompts that the LM can execute. More importantly, you can *compile with optimization* – meaning using example-based techniques to improve the program’s prompts or the model’s weights. Think of DSPy’s compiler as analogous to a software compiler: it takes your declarative code and produces a low-level execution plan (in this case, prompts and model adjustments) that achieves your specified behavior. We will explore basic compilation now (which just uses built-in prompt patterns), and advanced compilation/optimization later in the course.

* **Teleprompters (Optimizers):** Teleprompters – now officially called *optimizers* – are algorithms that *automatically improve* your DSPy programs. Instead of hand-tuning prompts or gathering large training sets for fine-tuning, you can provide a few examples and a performance metric, and an optimizer will adjust the program. Optimizers can synthesize helpful few-shot examples, refine the instructions in prompts, or even fine-tune model weights to better align with the task. In essence, a DSPy optimizer “tunes the prompts and/or the LM weights” of your modules to maximize a given metric (accuracy, F1, etc.). We will see how to use optimizers in a later module for *self-improving* our agents.

**Example:** Let’s illustrate these fundamentals with a simple DSPy program. We’ll create a signature and module for a basic task, then call it:

```python
import dspy

# Define a simple signature for question answering
sig = dspy.Signature("question -> answer")

# Create a module that uses a direct prediction strategy for this signature
answer_question = dspy.Predict(sig)

# Call the module with an input
result = answer_question(question="Where is the Eiffel Tower located?")
print(result.answer)
```

When you run this, DSPy will construct a prompt behind the scenes to ask the model the question. The output `result` is a `dspy.Prediction` object with an `answer` field. For instance, the model might output **“Paris, France”** as the answer. This simple example demonstrates how *Signatures* (here `"question -> answer"`) and a *Module* (`Predict`) let us query the model without writing any prompt manually.

**Exercise:** *Hands-on practice with DSPy basics* – Define your own DSPy signature and module for a simple task, and test it with a language model.

1. **Define a Signature:** Choose a task such as translating text or classifying sentiment. Use `dspy.Signature("input -> output")` notation to define it. For example, define a signature `"text -> summary"` for summarization, or `"sentence -> sentiment"` for sentiment analysis.
2. **Create a Module:** Using your signature, initialize an appropriate module. For instance, `dspy.Predict("sentence -> sentiment")` for classification, or `dspy.Predict("text -> summary")` for a straightforward summarization module. (We use `Predict` for now as it’s the simplest module that just asks the model to produce the output directly).
3. **Call the Module:** Provide an example input and call the module, e.g. `result = module(sentence="I love this movie!")`. Print out the structured output (e.g. `result.sentiment`). If you have access to an LM (ensure you configured `dspy.configure(lm=...)` with an API key or local model), observe the model’s output.
4. **Verify and Experiment:** Check if the output makes sense for your input. Try a couple of different inputs to see the responses. This exercise helps you become familiar with DSPy’s syntax and the request-response cycle. By using signatures and modules, you’ve avoided writing any prompt – DSPy handled it for you, giving a taste of how we’ll build more complex agents.

## Module 2: Building Modular DSPy Programs

One of the strengths of DSPy is the ability to **compose modules** into larger programs. Just as we structure software into functions or classes, we can structure AI behaviors into modular components. In this module, we explore how to build multi-step, *modular DSPy programs* that incorporate control flow and reusability. The goal is to learn how to break down complex tasks into manageable sub-tasks implemented by individual modules, which can then be linked together in a pipeline or decision logic.

**Composition of Modules:** DSPy modules can be called from regular Python code, so you can orchestrate them however you need. For example, you might have one module generate an intermediate result which another module uses. This can be done sequentially (call module A, then use its output to call module B) or even conditionally (choose a module based on some input). Because each module encapsulates a prompting strategy, you can **reuse** modules across different programs or inputs easily. This encourages a *modular design*: for instance, you might create a `Summarize` module and a `Translate` module, and reuse them in different pipelines or agent tasks.

**Control Flow:** Standard Python control flow (if/else, loops, function calls) can wrap around DSPy module calls. For example, you can implement a simple decision: *if* the user’s query looks like a math problem, call a math-solving module; *else* call a general Q\&A module. While DSPy doesn’t enforce a particular pattern for multi-module programs, it provides flexibility to integrate LLM calls wherever needed in your code. Additionally, DSPy has higher-level composite modules (like `dspy.MultiChainComparison` or `ProgramOfThought`) that internally run multiple reasoning paths and compare results, but you can also manage multi-step logic manually.

**Reusability and Abstraction:** By defining modules for sub-tasks, you create building blocks that can be tested and optimized in isolation, then combined. For example, suppose you need an agent to *write an article*. You might have one module that **outlines** the article and another that **drafts** a section given an outline point. Each can be developed and optimized separately (perhaps by different team members or with different techniques). Then you can write a higher-level function or class that calls the outline module, then calls the draft module for each section of the outline. This separation of concerns makes the program easier to maintain and improves clarity.

**Example:** Let’s build a simple two-step DSPy pipeline. Imagine we want to answer a question using a two-stage approach: (1) first find relevant information about the question, (2) then formulate a direct answer. We’ll simulate step (1) by a dummy search (or it could be a RAG retrieval, which we’ll cover later), and step (2) by an LLM answer module:

```python
import dspy

# Module 1: A simple retrieval module (for demo, we hardcode or simulate retrieval)
def retrieve_info(question: str) -> str:
    # In a real scenario, this might query a database or search engine.
    # Here we simulate by returning a canned context for demonstration.
    knowledge_base = {
        "Eiffel Tower": "The Eiffel Tower is located in Paris, France.",
    }
    for key, info in knowledge_base.items():
        if key.lower() in question.lower():
            return info
    return ""  # return empty if no info found

# Module 2: An LLM module that given context and question, produces an answer.
qa_module = dspy.Predict("context, question -> answer")

# Composition: use the retrieve_info result as context for the QA module.
question = "Where is the Eiffel Tower located?"
context = retrieve_info(question)
response = qa_module(context=context, question=question)
print("Answer:", response.answer)
```

Here, we manually composed a pipeline: `retrieve_info` (a Python function, acting as a “module” for retrieval) followed by `qa_module` (a DSPy module using the LM). The first part finds a relevant sentence about the Eiffel Tower, and the second part uses the LM to generate an answer given that context. In this simple case, the LM might simply return *“Paris, France.”* as the answer, possibly by extracting it from the context. While this example is rudimentary (and doesn’t showcase DSPy’s full power yet), it illustrates how you can integrate *external logic or tools* with DSPy modules to form a multi-step solution.

**Exercise:** *Compose a multi-module agent.*

1. **Design a Pipeline:** Pick a task that could naturally be broken into two or more steps. For example, *document question answering* (first retrieve relevant document passages, then answer) or *multi-format response* (generate an answer, then translate it). Write down the steps your program should take.
2. **Implement Sub-Modules:** For each step, decide if it will be a DSPy module or a regular function/tool. For instance, you could have:

   * Module A: Use `dspy.Predict` or `dspy.ChainOfThought` to perform step 1 (e.g. extract key facts from text).
   * Module B: Use another module for step 2 (e.g. take those facts and form an answer or perform a calculation).
     Each module should have a clear signature (you can use inline signatures like `"input -> output"` for simplicity).
3. **Compose the Workflow:** Write a function or script that calls the modules in sequence, passing outputs from one to the next. This can involve some Python logic. For example, *if* Module A’s output meets some condition, call Module B; otherwise maybe call Module C (you can create a dummy Module C if needed for an alternate path).
4. **Test as a Whole:** Run your composed program on a test input. Observe the outputs at each stage (you might print intermediate results for clarity). Verify that the final result makes sense. Debug the flow if the outcome isn’t as expected, adjusting how modules are used or adding conditional handling as needed.
5. **Reusability Check:** Consider how you could reuse one of the modules in a different context. For example, could Module A (say it’s a summarizer) be used on its own for other summarization tasks? This reflection will reinforce the benefit of modular design.

Through this exercise, you practice breaking a problem into DSPy components and orchestrating them. In later modules, we will enrich these pipelines with retrieval, tool use, and memory, but the principles of composition and control flow remain the same.

## Module 3: Retrieval-Augmented Generation (RAG)

Large Language Models have inherent knowledge limits – they can’t know about information not in their training data or that changes over time. **Retrieval-Augmented Generation (RAG)** addresses this by equipping an LLM with the ability to fetch relevant information from an external source (like a document database or knowledge base) before answering. In a RAG pipeline, when a query comes in, the system first **retrieves** relevant text (e.g. documents, snippets) related to the query, and then the **generation** stage uses both the query and the retrieved context to produce a final answer. This method allows the model to provide up-to-date, specific answers and often improves factual accuracy.

**Implementing RAG in DSPy:** There are a couple of approaches to build RAG with DSPy:

* *Pipeline approach:* Use an external retriever (embedding-based search, vector database, etc.) in your code, then pass the retrieved text as part of the input to a DSPy module (such as a chain-of-thought prompt that reads “context + question -> answer”). This approach treats retrieval as a separate step before calling the LM. For example, you might use DSPy’s built-in `dspy.ColBERTv2` tool or a custom function to fetch top relevant passages, and then call a `dspy.ChainOfThought("context, question -> answer")` module with those passages.

* *Agent approach:* Use a DSPy agent (like a `ReAct` module) that has a search tool integrated. In this setup, the language model can decide when to invoke the search tool to retrieve information mid-prompt. For example, a `dspy.ReAct("question -> answer", tools=[search_tool])` could allow the model to issue a search query as an action, get results, and then continue reasoning to formulate the answer. This is a more dynamic approach where the model controls retrieval as needed, which can be especially powerful for multi-hop questions or when the number of retrieval steps isn’t fixed.

**High-Performance RAG:** DSPy’s advantage is that RAG pipelines can be optimized via its compilation algorithms. For instance, DSPy can learn good example demonstrations for how to incorporate retrieved context into answers, or even refine how the query is transformed into a search query. By compiling a RAG program with a few Q\&A examples, you can significantly improve its accuracy (we’ll touch more on optimization in Module 6). A well-structured RAG system might retrieve multiple snippets and chain reasoning over them. DSPy’s modules like `ChainOfThought` can handle multi-part context input and produce an answer with a rationalization (reasoning trace), helping ensure the model actually *uses* the retrieved info.

**Example:** Suppose we want to build a Q\&A system that answers questions about a knowledge base (e.g., a set of Wikipedia abstracts). We can do this by writing a retrieval function and then using a DSPy module for generation. Let’s outline a simple example using DSPy’s ColBERTv2 tool for document search:

```python
import dspy

# Tool: ColBERTv2 semantic search (assuming a server or index is available)
search = dspy.ColBERTv2(url="http://your-colbert-server/wiki_index")

# Define a retrieval function using the tool
def search_wikipedia(query: str, k: int = 3) -> list[str]:
    results = search(query, k=k)  # returns top k documents (as dicts)
    return [r['text'] for r in results]

# DSPy module: Chain-of-thought QA that takes context and question
qa = dspy.ChainOfThought("context, question -> answer")

# Use the retrieval + generation pipeline
question = "What's the name of the castle that David Gregory inherited?"
context_docs = search_wikipedia(question, k=3)
answer_pred = qa(context=context_docs, question=question)
print(answer_pred.answer)
```

In this code, `search_wikipedia` queries a vector store (or search index) to get relevant texts for the question. We then pass those texts as the `context` into a `ChainOfThought` module along with the question. The model will produce a reasoning trace and a final answer. For example, it might reason: *“David Gregory... inherited **Kinnairdy Castle** in 1664, so the castle’s name is Kinnairdy Castle.”* and output **“Kinnairdy Castle”** as the answer. By examining `answer_pred.reasoning` (if available), you could see how it used the context, which is useful for understanding and debugging the RAG process.

**Exercise:** *Build a retrieval-augmented QA agent.*

1. **Prepare a Knowledge Source:** For practice, create or identify a small knowledge base. This could be a list of text passages in a Python list or a simple dataset (e.g. a few Wikipedia paragraphs or product FAQs). If you have a vector database (like Qdrant, FAISS, or an embedding model), you can use it, but it’s fine to simulate retrieval by scanning a list of strings for keywords.
2. **Implement a Retrieval Function:** Write a Python function that takes a query and returns a few relevant pieces of text. For example, you can do a simple keyword match over your list of passages (find passages that contain words from the query). Or use an embedding-based approach if available. This function stands in for the “retriever” component of RAG.
3. **Define the QA Module:** Create a DSPy module that can take the retrieved context plus the question and generate an answer. You might use `dspy.Predict("context, question -> answer")` for a direct approach, or `dspy.ChainOfThought("context, question -> answer")` to allow the model to reason with the context. The choice may depend on how complex the questions are – chain-of-thought can be useful if multi-step reasoning over context is needed.
4. **Integrate and Test:** Write code to use your retriever for a given question, then feed the result into the QA module. Test it on a question you know the answer to from your knowledge base. Check if the answer is correct and if it actually uses the provided context. You can print intermediate steps (like which passages were retrieved) to verify it’s working as expected.
5. **Iterate:** Try a few different questions, including ones where no relevant info is in the knowledge base (the agent should ideally say it doesn’t know or give a best guess). Observe how the system behaves. This is an opportunity to refine your retrieval function (maybe adjust it to return more or fewer results, or filter them) or your prompt (module choice) to get better answers.

By completing this exercise, you will gain experience in linking external data with an LLM – a crucial technique for real-world agents like document assistants, customer support bots, or research aides that need up-to-date information. RAG is a powerful pattern for grounding LLMs with facts, and now you’ve seen how DSPy facilitates building such pipelines.

## Module 4: Tool Use and Actuation

One of the exciting capabilities of agentic language models is their ability to use **tools** – calling external APIs, running computations, or performing actions in the world. DSPy supports tool use via modules like `ReAct`, which implements the ReAct (Reason + Act) paradigm: the model **reasons** about a problem, decides on an **action** (tool invocation), executes the tool, and then uses the tool’s result to continue reasoning. This module will teach you how to create agents that can act beyond pure text generation, effectively turning an LLM into a decision-making system that can query databases, do math, call web services, etc.

**Defining Tools:** In DSPy, a tool is typically a Python function that performs some operation (search, calculation, database lookup, etc.). You can register these with certain DSPy modules (like `dspy.ReAct`) so that the language model knows they exist and can request to use them. Tools should have clear input/output behavior (e.g., a function that takes a string query and returns a string answer, or takes some data and returns a computed result). DSPy provides a few built-in tools: for example, `dspy.PythonInterpreter` (which can execute Python code, useful for math or logical operations) and the `ColBERTv2` search tool we saw earlier. You can also easily create custom tools by writing your own Python functions and including them in an agent module’s tool list.

**ReAct Agent Mechanics:** When you use `dspy.ReAct`, you specify a signature for the overall task and provide a list of tools. Under the hood, DSPy will use a special prompt format that lists the available tools and expects the model to output *actions* (like “use tool X with these arguments”) and *thoughts* (the reasoning). The agent alternates between the model thinking and the system executing tools, until a final answer is produced. For example, imagine a question: *“What is 9362158 divided by the year of birth of David Gregory of Kinnairdy Castle?”* This requires both retrieving David Gregory’s birth year (a search task) and doing a division (a math task). A ReAct agent with a search tool and a calculator tool can handle it:

1. It will think about the question and likely decide to **search** for "David Gregory Kinnairdy Castle birth year".
2. After the search tool returns some text (say, a biography snippet), the agent extracts the birth year (e.g. 1625) from it.
3. Next, it will realize it needs to divide 9362158 by 1625, so it invokes the **math tool** (which could be a Python interpreter) with the expression `9362158/1625`.
4. The tool executes and returns the result (5769.27...), and the agent then formulates the final answer.
5. The conversation ends with the agent outputting the answer, possibly with a brief explanation or just the number, depending on the prompt style.

Throughout this process, the agent is essentially writing a chain-of-thought that includes tool calls, and DSPy is managing the execution of those calls. The flow described might look like this in an actual trace: first a **thought** like “I should search for David Gregory’s birth year”, then an **action** calling the search tool, then receiving the result, then a **thought** “He was born in 1625, now divide 9362158 by 1625”, then an **action** calling the calculator, and finally the **answer**.

**Implementing Tools in DSPy:** Let’s write a quick example agent with a tool. We’ll create a simple arithmetic tool and an agent that can use it:

```python
import dspy

# Define a simple tool function for arithmetic evaluation
def evaluate_math(expression: str) -> float:
    """Evaluate a math expression and return the result."""
    return dspy.PythonInterpreter({}).execute(expression)

# Suppose we also have a search tool (for illustration, can reuse search_wikipedia from before)
def search_wikipedia(query: str) -> list[str]:
    # ... (use an API or stub as before) ...
    return [ "David Gregory was born in 1625 and inherited Kinnairdy Castle in 1664." ]

# Create a ReAct agent that can use these tools
agent = dspy.ReAct("question -> answer: float", tools=[evaluate_math, search_wikipedia])

# Ask a complex question that needs both tools
query = "What is 9362158 divided by the year of birth of David Gregory of Kinnairdy Castle?"
result = agent(question=query)
print("Final Answer:", result.answer)
```

Here we defined two tools: `evaluate_math` uses DSPy’s Python interpreter to do calculations, and `search_wikipedia` (stubbed for demonstration) returns a relevant piece of text. We then initialize a `ReAct` agent with the signature `"question -> answer: float"` (meaning the answer expected is a number) and provide the tools. When we call `agent(...)`, the LLM will go through reasoning steps and use the tools. If you inspect the internal trace (e.g., via `dspy.inspect_history()` or the `result` object if it contains a reasoning log), you would see the sequence of thoughts and tool actions taken by the agent. In the end, it should print the correct numeric answer (in this case, **5769.27** approximately) and possibly how it got it. This demonstrates how tool use allows the agent to **actuate** external logic (searching data, performing math) as part of its reasoning loop.

**Exercise:** *Develop an agent with custom tool use.*

1. **Choose a Use Case:** Think of a scenario where an LLM alone isn’t enough – it needs to perform some action. Examples: a *calculator* agent, a *wiki lookup* agent, a *weather bot* that calls a weather API, or a *database query* agent. Define what tools would be helpful for that scenario.
2. **Implement the Tool Function:** Write one (or more) Python functions as your tools. Keep them simple and focused: e.g., `def get_current_time(): return datetime.now().isoformat()`, or `def find_in_knowledge_base(term: str) -> str: ...` to return an answer from a mini database. Test your function independently to ensure it works.
3. **Create a DSPy Agent Module:** Use `dspy.ReAct` (or you could use `dspy.Tool` in simpler cases) with an appropriate signature. For instance, `"question -> answer"` or `"input -> output"` if it’s not a Q\&A format. Provide the list of your tool functions to the module via the `tools=[...]` argument.
4. **Run a Query Through the Agent:** Call your agent on a prompt that should trigger tool usage. For example, if you made a calculator agent, ask: *“What is (15^2 + 10) / 3?”* or if a weather bot with a dummy API, ask: *“What’s the weather today in New York?”*.
5. **Observe and Refine:** Check the agent’s answer. If possible (and if DSPy outputs a reasoning trace or you use `inspect_history`), review the steps it took: did it call the tool appropriately? Did it use the result correctly? If the agent didn’t use the tool when it should, you might need to adjust the prompt or signature (for example, ensure the agent knows it *can* use the tool by how the task is described in the signature or adding an instruction in the system prompt).
6. **Iterate with Another Tool (Optional):** For more practice, add a second tool to the agent and handle a query that might require both. For example, combine a math tool and a search tool as we did above, and try a query needing both (like a real-world scenario: “Using today’s exchange rate, how much is 100 USD in EUR?” which might need an API call for the rate, then a calculation).

This exercise gives you hands-on experience with extending LLM capabilities through tools. In real-world applications, tool use is how we get AI agents to interact with up-to-date data and services – from databases and web APIs to calculators and beyond. After this, you should understand how DSPy facilitates tool integration and how an agent decides to use tools during its reasoning.

## Module 5: Stateful Multi-turn Agents

So far, our agents have handled one query at a time. However, many practical agents operate in a **conversation** or multi-turn interaction, where maintaining *state* or *memory* across turns is crucial. In this module, we learn how to design agents that can remember past dialogue, maintain context over multiple turns, and chain reasoning across a conversation. We’ll leverage DSPy’s capabilities (like the `dspy.History` primitive for chat history) to build stateful agents, such as chatbots that recall what the user said earlier, or assistants that carry information forward through a session.

**Maintaining Conversation Context:** Large language models can take a sequence of messages (history) as input (like how ChatGPT works with system/user/assistant messages). DSPy provides a structured way to include conversation history using an `InputField` of type `dspy.History` in your signature. Essentially, you can design your signature to have a field (say `history`) that accumulates previous question-answer pairs. Each time the agent is called, you pass in the updated history along with the new user query, and the model will be prompted with that history in context. The `dspy.History` object is simply a list of message dictionaries, where each entry could have keys like "question" and "answer" (or whatever fields your signature defines for a turn).

There are a couple of patterns for multi-turn memory:

* **Explicit History Field:** As mentioned, include a `history: dspy.History` in your signature’s inputs. After each turn, update this history with the latest question and answer. This way, the next call to the agent gets the entire conversation so far. DSPy’s chat adapter will format this appropriately for the model (usually as a series of user/assistant messages).
* **Implicit through external memory:** Alternatively, you could manage memory outside the DSPy call (e.g., keep a list of past messages and manually prepend them in a prompt for a `dspy.Predict`). However, using DSPy’s built-in `History` type is cleaner and less error-prone, as it automatically handles things like formatting and ensuring the model sees the history in a structured way.

**Chaining Reasoning Across Turns:** Multi-turn interactions sometimes involve multi-step problem solving spanning turns (for example, an autonomous research assistant might carry over partial results from one query to the next). With DSPy, you can carry state in variables or in the history between turns. For instance, you might store an intermediate answer or plan in the history so that the next question can refer to it. Another strategy is to use the model’s output in one turn to formulate the next question (this bleeds into planning agents territory). In essence, designing multi-turn agents might involve both *chat memory* (remembering what has been said) and *state memory* (remembering what has been concluded or decided).

**Example:** Let’s create a simple conversational agent that remembers a user’s name. The agent will greet the user and then be able to recall the name if asked later:

```python
import dspy

# Define a signature for a simple chatbot with memory
class ChatSig(dspy.Signature):
    user_input: str = dspy.InputField()
    history: dspy.History = dspy.InputField()    # to store past interactions
    answer: str = dspy.OutputField()

# Create a chat module (we use Predict for simplicity; it will handle multi-turn format internally)
chatbot = dspy.Predict(ChatSig)

# Simulate a conversation
history = dspy.History(messages=[])

# Turn 1: User greets and introduces name
user_msg1 = "Hello, my name is Alice."
result1 = chatbot(user_input=user_msg1, history=history)
print("Bot:", result1.answer)

# Update history with turn 1
history.messages.append({"user_input": user_msg1, "answer": result1.answer})

# Turn 2: User asks a question that requires remembering the name
user_msg2 = "Nice to meet you. What is my name?"
result2 = chatbot(user_input=user_msg2, history=history)
print("Bot:", result2.answer)
```

In this hypothetical code, the first user input is *“Hello, my name is Alice.”* The bot might respond with something like *“Hello Alice! How can I help you today?”* (the exact answer depends on the model and prompt). We then append the first exchange to the history. On the second turn, the user asks *“What is my name?”*. Because we provided the history which includes the user’s statement of their name, the model can answer *“Your name is Alice.”* correctly. This simple example demonstrates how including `history` allows the agent to carry information forward. The key is that after each `chatbot()` call, we updated the `history` object with the latest user\_input and bot answer. In a real application, you would loop this process for each new message.

**Memory Management Considerations:** As conversations grow, the history can become long. You might need to truncate or summarize history if it gets near model context length limits. This could be done by keeping only the last N turns or using a DSPy module to summarize old interactions when the history is too large (an advanced technique). Also, ensure that the history format keys match your signature exactly (in our case "user\_input" and "answer"). The DSPy framework will inject the conversation into the prompt format so the model sees something like earlier user messages and its own responses.

**Exercise:** *Create a stateful conversational agent.*

1. **Define a Chat Signature:** Design a DSPy `Signature` class for a chat. It should include at least the user message (`InputField`) and a history (`InputField` of type `dspy.History`), plus an `OutputField` for the assistant’s reply. For example, `class MyChatSig(dspy.Signature): ...` as in the example above.
2. **Initialize a Chat Module:** Use `dspy.Predict(MyChatSig)` or `dspy.ChainOfThought(MyChatSig)` to create your chat agent. (`Predict` may suffice for friendly chat; `ChainOfThought` could be used if you want the bot to internally reason each turn, though that’s usually not needed for casual dialogue).
3. **Simulate a Dialogue:** Create a `dspy.History(messages=[])` object to start with an empty conversation. Then call your module with a first user message, e.g. a greeting or a question. Capture the output, print it, and then update the history with the user message and the bot’s answer (as a dict like `{"user_input": ..., "answer": ...}` matching your field names).
4. **Continue the Conversation:** Provide a second user input, this time something that tests memory. It could be *“What did I just say?”* or referencing something from the first turn (*“Can you summarize what I asked?”* or *“You remember my name, right?”* if you told it your name). Call the module again with the new input and the updated history. Check if the agent’s answer correctly uses the context from the history.
5. **Iterate More Turns:** If you like, go for a longer dialogue. Try to have the agent learn a fact in one turn and use it later. For example, turn 1: *“I live in Canada.”* Bot responds. Turn 2: *“Where do I live?”* – the bot should answer *“Canada”*.
6. **Error Handling:** If the agent forgets or answers incorrectly, consider why. Is the history being passed correctly? You might inspect the `history.messages` content or even print it to debug. Ensure each turn you append the new Q\&A properly. If the model still fails, you might need to enforce memory by adjusting the prompt (like instructing the model “The user’s name is X” in the system prompt), but ideally DSPy’s conversation handling should suffice for simple cases.

This exercise helps you practice maintaining state. Building a stateful agent is essential for realistic applications like customer service chatbots (which must remember what issue the customer described) or personal assistants (that should recall user preferences mentioned earlier). By using DSPy’s history mechanism, you keep the implementation clean and let the framework handle how the past dialogue is injected into the model prompt.

## Module 6: Self-Improving Programs (Compilation & Optimization)

A standout feature of DSPy is the ability to **optimize your AI programs** using example-based supervision – essentially making them *self-improve* over time. In this module, we focus on how to compile and refine DSPy programs using small datasets of examples and evaluation metrics. The goal is to achieve better performance (accuracy, relevance, etc.) without manual prompt tuning, by leveraging DSPy’s *optimizers* (formerly teleprompters) to adjust prompts or model weights.

**How DSPy Optimization Works:** You provide three things to an optimizer: (1) your DSPy program (one module or a whole pipeline), (2) a **metric function** that scores the program’s output against a target (or measures some objective), and (3) a **training set** of inputs (optionally with expected outputs). The optimizer will then *compile* the program, meaning it will iteratively modify the program’s prompts or few-shot examples, possibly generate synthetic data or fine-tune a smaller model, all to improve the metric. Importantly, you often don’t need a large dataset – a handful of examples (even 5-10) can be enough to bootstrap improvements.

DSPy comes with various optimizers implementing different strategies:

* **Automatic Few-Shot Learning:** e.g., `dspy.BootstrapRS` (Bootstrap with Random Search) or `dspy.BootstrapGPT`. These try to find good demonstration examples to prepend to prompts by generating and evaluating candidate examples.
* **Instruction Optimization:** e.g., `dspy.MIPROv2` (a prompt instruction optimizer) which tweaks the wording of the instructions in your prompts to improve performance. This can discover better ways to ask the model to do the task.
* **Automatic Finetuning:** e.g., `dspy.BootstrapFinetune`, which uses a larger set of examples (possibly generated or user-provided) to actually fine-tune the language model’s weights. This is useful if you can gather more data or want to specialize a smaller model.
* **Program Transformations:** Other optimizers might restructure the program (e.g., ensemble multiple modules or insert additional reasoning steps) to improve results.

Using these optimizers is quite straightforward in code. You instantiate the optimizer with any needed settings (like which metric to use), then call `optimized_program = optimizer.compile(original_program, trainset=..., **options)`. The result is a new program (often of the same class as original, but now with optimized prompt or added examples) that you can use just like before, but it will perform better on the kind of inputs you trained on. A simple example: if you have a `classify` module that isn’t very accurate initially, you can provide a few labeled examples and use `dspy.BootstrapRS` to improve it. After compilation, the `classify` module might now include a few effective exemplars in its prompt and yield higher accuracy on similar inputs.

**Example:** Let’s say we built a sentiment classifier earlier, but we find it sometimes mislabels certain phrases. We’ll create a small training set and optimize it:

```python
import random, dspy

# Original program: sentiment classifier
classifier = dspy.Predict("sentence -> sentiment")

# Tiny training set of examples (sentence and correct sentiment)
trainset = [
    {"sentence": "I absolutely love this!", "sentiment": "positive"},
    {"sentence": "This is the worst thing ever.", "sentiment": "negative"},
    {"sentence": "I'm not sure how I feel about this.", "sentiment": "neutral"}
]

# Define a metric function: accuracy (1 if correct, 0 if not)
def accuracy_metric(output, example, trace=None):
    # output: a Prediction object from the classifier
    # example: the trainset dict with expected sentiment
    return 1.0 if hasattr(output, 'sentiment') and output.sentiment == example['sentiment'] else 0.0

# Choose an optimizer, e.g., MIPROv2 for instruction optimization
optimizer = dspy.MIPROv2(metric=accuracy_metric)

# Compile/optimize the classifier using the trainset
optimized_classifier = optimizer.compile(classifier, trainset=trainset)

# Test the improved classifier
test_sentence = "It wasn't great, not terrible either."
pred_before = classifier(sentence=test_sentence)
pred_after = optimized_classifier(sentence=test_sentence)
print("Before Optimization:", pred_before.sentiment)
print("After Optimization:", pred_after.sentiment)
```

In this snippet, we created a small `trainset` with a few labeled examples and defined an `accuracy_metric` to compare the model’s output to the expected sentiment. We used `MIPROv2` as the optimizer; upon calling `compile`, it will adjust the classifier. For instance, it might find that adding an instruction like *"Respond with positive, negative, or neutral sentiment."* or adding one of the examples to the prompt helps. If the original classifier was shaky on certain inputs, the `optimized_classifier` should now be better. We can compare outputs on a test sentence: ideally, the optimized version is more likely to get it correct. Indeed, DSPy’s documentation reports significant score improvements with just a few examples – e.g., more than doubling accuracy in some tasks after compilation.

**Exercise:** *Optimize a DSPy program with example-based supervision.*

1. **Select a Target Task:** Use one of the agents or modules you built in earlier exercises that could improve with learning. Good candidates are classification tasks (e.g., label an email as spam/ham), extraction tasks (pull out a field from text), or even a simplified RAG Q\&A where known questions have known answers.
2. **Gather a Few Examples:** Create a small training dataset (`trainset`). For a classifier or QA, you might have a list of dictionaries like `{"input": X, "expected": Y}`. If your program takes multiple inputs, include them all (e.g., for a signature `A, B -> C`, each example needs A and B and the expected C). You don’t need many examples – even 5-10 can be enough to see improvement. If you don’t have real data, you can fabricate some plausible examples.
3. **Define a Metric:** Write a Python function that takes the program’s output and the example, and returns a score. For straightforward tasks, accuracy works (return 1 or 0 for correct/incorrect). You could also use a more nuanced metric if needed (e.g., partial credit), but keep it simple initially.
4. **Choose an Optimizer:** Depending on the task and data size, pick an optimizer:

   * If only a few examples and you suspect prompt wording matters, try `dspy.MIPROv2` (optimizes instructions).
   * If you want it to generate a few-shot prompt automatically from examples, try `dspy.BootstrapRS` or `dspy.BootstrapGPT`.
   * If you actually have a larger dataset (say dozens or hundreds of examples) and a smaller model that can be fine-tuned, `dspy.BootstrapFinetune` could be used (requires more setup, maybe skip for now unless you’re ambitious).
5. **Compile the Program:** Call `optimizer.compile(your_program, trainset=trainset)` and wait for it to run. (Under the hood it may run the model many times on your examples, possibly generating variants, which can take some time.) It will return an optimized version of your program.
6. **Compare Performance:** Test the original program vs the optimized program on a few inputs (ideally including ones from the trainset and some new ones). Measure by your metric or simply observe the difference. Did the optimized program produce more accurate or consistent results? For example, you might see that an answer now includes a detail it missed before, or a classification that was wrong is now corrected. If you see improvement, congratulations – your DSPy agent just *learned* from examples! If not, analyze: maybe the metric wasn’t aligned or the examples weren’t sufficient or representative. Consider adding another example or tweaking the metric and try again.

By doing this exercise, you practice the **data-centric iteration** that DSPy encourages: instead of tweaking prompts by guesswork, you provided examples and let the system optimize. This is a powerful method to adapt agents to specialized tasks or domain-specific behavior. Throughout this process, always keep in mind the size of examples (few-shot means you don’t need big data) and that the optimizers are not magic but can significantly speed up the prompt/weight tuning that you’d otherwise do manually.

## Module 7: Advanced Compilation Strategies

Now that you know how to optimize a single DSPy program, let's explore **advanced strategies** to push your agents’ performance further. This module delves into customizing the compilation process, leveraging feedback loops, and employing heuristics to guide program improvement. It also touches on extending DSPy with your own optimizers or combining multiple optimization techniques. The key theme is going beyond one-off compilation: using iterative refinement and creative strategies to build highly effective agents.

**Custom Compilers/Optimizers:** While DSPy provides a suite of built-in optimizers, advanced users can implement custom optimization logic. For example, you might want a specialized optimizer that focuses on a particular type of error your agent makes. Since DSPy optimizers are essentially algorithms that iteratively run your program and adjust it, you can write your own by using the DSPy API to manipulate modules. A simple custom approach might be: generate a variety of prompt variants for a module, test them on a small set of examples (by calling the module and scoring outputs), then pick the best variant. This is essentially what some optimizers do internally (like random search over prompt wording), and you can tailor it if needed. You could even create an optimizer that enforces certain constraints – e.g., always include a specific phrase or adhere to a format (this overlaps with the idea of **DSPy Assertions**, an advanced feature for enforcing output structure or constraints, based on recent research).

**Feedback Loops and Self-Refinement:** Consider enabling an agent to critique or refine its own output. A feedback loop strategy might involve the agent generating an answer, then evaluating that answer with either a heuristic or a secondary model, and then using that feedback to improve. For instance, you might compile a program that first attempts an answer, then checks (via another module or a tool) if the answer is correct or high-quality, and if not, revises the answer. This can be done by structuring your DSPy program in multiple stages:

* Stage 1: Generate initial answer (and reasoning trace).
* Stage 2: Evaluate the answer (maybe a module that compares the answer to known facts or runs test cases).
* Stage 3: If evaluation is negative, modify the prompt or ask the LM to "try again with adjustments".
  This essentially can turn into an **iterative agent** that keeps improving until a criterion is met. In DSPy, you could implement this logic in Python (looping until the evaluation metric is high) or potentially as a recursive module call.

A concrete example of feedback loop is using a second LLM to act as a *critic*. Your agent can produce an answer and a rationale, and then you ask a critic model (maybe via a `dspy.Predict("answer, question -> critique")`) to analyze if the answer fully addressed the question or if there were mistakes. If the critique identifies issues, your agent can incorporate that and try again. While implementing such a loop requires careful prompt design to avoid going in circles, it can significantly enhance reliability for complex tasks.

**Heuristics in Optimization:** Not all improvements need heavy ML optimization; sometimes simple heuristics can guide the compilation. For example:

* If you know that certain keywords or styles yield better responses from the model (perhaps from your manual prompt experiments), you can constrain the search space of an optimizer to prompts containing those keywords.
* You might use a heuristic to generate candidate few-shot examples: e.g., for a translation task, automatically create a couple of example pairs using Google Translate as a reference, then give those to DSPy’s optimizer.
* Another heuristic approach is **ensemble and majority voting**: use multiple variants of an agent and then pick the answer that most agents agree on. There is actually a DSPy concept of an *Ensemble teleprompter* to merge outputs. In advanced usage, you could compile multiple versions of a module (e.g., one optimized for precision, one for creativity) and then have a meta-module choose between them (or even combine their outputs). This can improve robustness by not relying on a single prompt/model instance.

**Example – Ensemble Strategy:** Suppose you have two different prompt strategies for a task (say, two styles of chain-of-thought). Rather than pick one, you can run both and then decide:

```python
# Assume module1 and module2 are two DSPy modules solving the same task differently
module1 = dspy.ChainOfThought("question -> answer")  # strategy A
module2 = dspy.ReAct("question -> answer", tools=[...])  # strategy B with tools

def ensemble_answer(question):
    pred1 = module1(question=question)
    pred2 = module2(question=question)
    # Heuristic: if both modules agree on the answer, use it; otherwise, choose one with more reasoning length or confidence
    if pred1.answer == pred2.answer:
        return pred1.answer
    else:
        # As a simple heuristic, pick the answer from the module that gave a longer reasoning (assuming maybe more thorough)
        return pred1.answer if len(getattr(pred1, "reasoning", "")) > len(getattr(pred2, "reasoning", "")) else pred2.answer

# Using the ensemble function
q = "Some complex question here?"
print("Ensemble answer:", ensemble_answer(q))
```

This snippet sketches an ensemble where two modules produce answers and we choose between them. We used a heuristic of reasoning length as a proxy for confidence (this is just an example – in practice, you might examine a confidence score if available, or have another module adjudicate). While this ensemble logic is written outside DSPy (in plain Python), you could also imagine integrating it as part of a DSPy program or even compiling it (DSPy’s `teleprompt.Ensemble` optimizer might allow combining modules more systematically). The advanced point here is: by combining multiple approaches, we can often get better results than any single approach.

**Exercise:** *Experiment with advanced improvement techniques.*

Choose one of the following mini-projects (or both, if ambitious) to deepen your understanding of advanced strategies:

* **A. Custom Optimizer/Heuristic Tuning:** Identify a particular weakness in one of your agents. For example, maybe your document QA agent sometimes includes irrelevant info in the answer, or your chatbot sometimes gives too brief responses. Design a custom improvement approach for this. Perhaps write a loop that adjusts a prompt parameter and test, or filter the agent’s output through a heuristic. Concretely, you could do something like: *automatically prepend a specific instruction* (like "Give the answer in two sentences") and see if it improves quality – essentially treating prompt editing as a search problem where you manually apply a heuristic (the heuristic being your guess that two sentences are better). Measure before-and-after on a few test inputs.

* **B. Self-Critique Loop:** Implement a simple self-critique mechanism for an agent. Use one of your DSPy modules as the main answer producer. Then create a second module (or even reuse the same model in a different way) to evaluate that answer. For example, for a math problem solver, second module could double-check the calculation; for a summarizer, second module could verify if certain key facts from the original text appear in the summary. If the evaluation fails, have the pipeline adjust. This could mean prompting the model again with a note like "The previous answer missed X, please correct that." You might only do one round of feedback due to complexity, but even that is instructive. Test your loop on an input where the first attempt is known to be flawed (maybe remove a detail on first attempt intentionally) and see if the second attempt improves it.

* **C. Ensemble Voting (Optional):** If you have two different versions of an agent (say, one optimized with few-shot examples and another optimized with instruction tuning), try combining their outputs. For a classification task, you could have them vote on the label (if disagree, perhaps use a third strategy or default to one). For a generation task, perhaps compare their answers and pick the one that better satisfies some criteria (maybe length or presence of a keyword). Run this ensemble on a small set of test cases to see if it fixes some mistakes that individual ones made.

These advanced exercises are open-ended and intended to stimulate creative problem-solving. There may not be a single “right” answer; the objective is to experiment with techniques that go beyond the default. In real-world deployments, these kinds of strategies (like adding a safety checker after a generative model, or doing multi-step self-refinement) are common to ensure the AI system meets quality and reliability targets. By attempting one of these, you’ll get a taste of what it’s like to engineer an AI system with belts-and-suspenders for performance – something that distinguishes advanced practitioners.

## Module 8: Integration with LangChain, OpenAI APIs, and Vector Stores

Modern AI applications often use a combination of frameworks and services. In this module, we discuss how DSPy can integrate with other popular tools like LangChain, connect to external model APIs (OpenAI, etc.), and utilize vector databases for advanced retrieval. The emphasis is on interoperability and leveraging the broader ecosystem: you can use DSPy as the “brains” of your agent while still taking advantage of data pipelines, memory stores, or UI frameworks that exist elsewhere.

**Using External LLM Providers:** DSPy is model-agnostic; it can work with local models or cloud APIs. You’ve seen that we configure `dspy.LM(...)` with model identifiers like `'openai/gpt-4'` or custom endpoints. This means you can seamlessly use OpenAI’s GPT-4, Anthropics’s Claude, or others as the underlying engine for DSPy modules. Integration is often as simple as installing the provider’s SDK or setting an API key. For example, to use OpenAI’s API, you ensure `OPENAI_API_KEY` is set and do `dspy.LM("openai/gpt-4", api_key=...)`. For local models, you might use Hugging Face transformers or specialized servers (like Ollama as shown in the earlier blog example). The takeaway is that switching the model in a DSPy agent is straightforward, which is great for deployment flexibility (e.g., using a smaller local model in offline settings, and a powerful cloud model when high accuracy is needed).

**Integrating with LangChain:** [LangChain](https://python.langchain.com/) is a framework that many developers use to build chains and pipelines with LLMs, offering lots of integrations (vector stores, tools, memory, etc.). You can think of DSPy and LangChain as complementary: LangChain provides a toolkit for connecting data sources and orchestrating calls, while DSPy provides a powerful way to program the LLM’s prompt and reasoning behavior with optimization. In practice, you might use LangChain to do heavy data prep (like loading and splitting documents, as we saw in the RAG example with LangChain loaders) and then call a DSPy agent to actually answer questions based on that data. Or you might use DSPy to **compile** a prompt for a complex task and then plug that into a LangChain `LLMChain` if you prefer certain LangChain abstractions.

There is an emerging integration where LangChain can call DSPy as a prompt optimizer for its chains. In a prototype, one can wrap a LangChain chain and let DSPy’s compiler trace it and learn an optimized prompt for it. This is advanced, but it shows the principle: you could develop a chain with LangChain’s utilities, then use DSPy to automatically improve how that chain prompts the LLM. If you come from a LangChain background, adopting DSPy in parts of your system might gradually improve maintainability (prompts become declarative) and performance (via optimization). Conversely, if you have a mostly DSPy system but want a component that LangChain provides (say, a specific vector store integration or an agent executor with a nice interface), you can call LangChain from within DSPy (since you can call any Python code, a LangChain `RetrievalQA` for instance can be invoked inside a tool function or pipeline).

**Vector Stores and External Data:** Integration with vector databases (Pinecone, Qdrant, Milvus, etc.) is crucial for larger RAG applications. DSPy doesn’t reinvent vector storage; instead, you either use its provided tools (like ColBERTv2) or hook into outside libraries. For example, you can use the Qdrant Python client to perform similarity search as part of a DSPy tool. One approach:

```python
from qdrant_client import QdrantClient
from sentence_transformers import SentenceTransformer

# Initialize vector store and embedding model (this is outside DSPy)
qdrant = QdrantClient(...) 
embed_model = SentenceTransformer('all-MiniLM-L6-v2')

# Define a DSPy tool for vector search
def vector_search(query: str, top_k: int = 5) -> list[str]:
    vec = embed_model.encode(query).tolist()
    results = qdrant.search(collection_name="docs", query_vector=vec, limit=top_k)
    return [hit.payload.get("text", "") for hit in results]
```

Now `vector_search` can be used as a tool in a ReAct agent or to supply context to a chain-of-thought module. This way, any vector store of your choice is one function call away from your DSPy program. In fact, DSPy’s design encourages using external libraries when needed – for document loaders, for specialized models (image models, etc.). It’s as simple as calling those libraries within your DSPy module’s logic.

**Memory Integration:** If you are using something like LangChain’s conversation memory or a memory component of another system, you could similarly feed that into DSPy. For example, you could sync a LangChain `ConversationBufferMemory` with a `dspy.History` – they’re conceptually similar (a list of past messages). As long as you can extract plain text or a list of dicts from another system’s memory, you can construct a `dspy.History` from it and pass to your DSPy agent. This means you could drop a DSPy agent into an existing LangChain chatbot without losing the conversation history, by just converting formats appropriately.

**Example:** Here’s a conceptual example showing how you might wrap a DSPy agent in a LangChain interface (hypothetical code for illustration):

```python
# Suppose we have a DSPy agent defined (with a .predict() or __call__ interface)
dspy_agent = optimized_classifier  # from previous example, or any DSPy module

# We can create a LangChain LLMChain that calls this agent via a custom LLM class.
from langchain.llms.base import LLM

class DSPyAgentLLM(LLM):
    """Wrap a DSPy agent to use as an LLM in LangChain."""
    def _call(self, prompt: str, stop=None):
        # Here we assume the DSPy agent takes a raw prompt or has a simple signature we can use.
        result = dspy_agent(input=prompt)
        # If result is a DSPy Prediction object, extract text
        return str(result)
    @property
    def _identifying_params(self):
        return {"name": "dspy_agent_llm"}

# Now use it in a LangChain chain
llm_chain = LLMChain(llm=DSPyAgentLLM(), prompt=some_langchain_prompt_template)
output = llm_chain.run(some_input)
```

This pseudo-code demonstrates that you can integrate at various levels: either calling DSPy from LangChain or vice versa. The specifics will depend on what you’re trying to achieve, but the flexibility is there.

**Exercise:** *Bridge DSPy with external tools and frameworks.*

1. **Vector Store Connection:** If you have access to a vector database (like Qdrant, Pinecone, Weaviate, etc.) or even a local FAISS index, try connecting it to your DSPy RAG agent from Module 3. You can use that service’s Python client to add data (e.g., index some documents) and then query it in a tool function. Replace your earlier retrieval stub with a real vector search call. Test your QA agent again on a question that requires retrieving from that store. Does it return the correct info? This exercise will show you how to incorporate scalable retrieval.
2. **LangChain Utility Usage:** Use LangChain’s utilities to assist your DSPy program. For instance, use `langchain.text_splitter` to split a long document into chunks, then feed those chunks to your DSPy pipeline for summarization or Q\&A. Or use `langchain.document_loaders` to read a PDF or webpage, and then let a DSPy agent answer questions about it. You don’t have to integrate deeply; just see how LangChain can handle data ingestion and you pass the processed data into DSPy.
3. **Swap Models Easily:** Experiment with changing the underlying model for your agent. If you used OpenAI’s API so far, try running your agent on a local LLM (maybe a smaller model via HuggingFace or an open source API). Configure `dspy.LM` accordingly (for example, `dspy.LM("openlm-research/open_llama_7b")` if you have it installed, or using the Ollama example to run a local model). Evaluate the difference in responses or speed. Conversely, if you started with a local model, try an API model. This will teach you about latency and result differences, and how DSPy abstracts the model layer.
4. **(Optional) LangChain Agent with DSPy Tool:** For a fun challenge, create a LangChain Agent (using their `initialize_agent` with tools) and make one of the tools a call to a DSPy module. For example, a LangChain tool could be defined that calls a DSPy summarizer. Then you can ask the LangChain agent to do something that involves summarization, and it will internally use the DSPy-powered tool. This shows that you can insert DSPy’s optimized behaviors into a LangChain orchestrated workflow.

By completing these integration exercises, you become comfortable using DSPy in real-world contexts, where it rarely lives in isolation. In production, one might use LangChain for connectivity and DSPy for optimized reasoning, or use DSPy to develop a robust agent and then deploy it behind an API endpoint that other systems call. The ability to mix and match gives you the best of both worlds.

## Module 9: Testing and Evaluation of DSPy Agents

Building an agent is not just about getting it to work – it’s also about verifying it works correctly and consistently. In this module, we focus on **testing and evaluation** techniques specific to agent-based systems, especially those built with DSPy. We will discuss how to design test cases for AI behaviors, choose appropriate evaluation metrics, and utilize DSPy’s features (like tracing and history inspection) for debugging. Ensuring your agent performs as expected across various scenarios is crucial before deployment.

**Designing Test Cases:** Just like traditional software, AI agents benefit from unit tests and integration tests. However, testing AI outputs can be trickier because of variability. Here are some approaches:

* **Deterministic Mode for Testing:** For LLM-based components, you can run them with a fixed random seed or with `temperature=0` to make them deterministic. This way, given the same input, the output should be the same each time, which is important for test reproducibility.
* **Unit Testing Modules:** If you have submodules (e.g., a date parser module, a math tool), test them in isolation. For example, if you have a `calculate_tip` module that uses chain-of-thought to compute a tip, feed it a known input and assert that the output matches expected (with a tolerance if needed for floats). Since DSPy modules return structured outputs, you can directly compare fields.
* **Functional/Integration Tests:** Test the whole agent on representative queries. For instance, if you built a customer support bot, prepare a few example dialogues or questions it should handle and run them through the agent. Check not only if it gives a correct answer, but also if the format and tone are as desired.
* **Edge Case Testing:** Think of unusual or tricky inputs: If a tool is supposed to be used, what if the query is such that the agent might bypass the tool? E.g., test a question that *requires* a calculation and ensure the agent does call the calculator tool (you might check the reasoning trace to confirm this). Or test the memory: ask two unrelated questions in a conversation and verify the second answer doesn’t carry over info incorrectly (ensuring it resets or distinguishes contexts properly).

**Evaluation Metrics:** Depending on your agent’s role, you’ll choose metrics:

* For factual Q\&A or classification, metrics like **accuracy**, **F1 score**, or **exact match** might apply. You can create a small labeled set and compute these. In DSPy, you might already have a metric function from your optimization stage (Module 6) – reuse that for evaluation on a test set.
* For summarization or generative tasks, automated metrics like **ROUGE** or **BLEU** can give a sense of performance, but human evaluation is often needed for quality aspects (fluency, coherence). You can at least check constraints: e.g., if summary length is supposed to be under 100 words, have a test that counts words.
* For conversational agents, you might define success criteria per turn (did the agent eventually solve the user’s problem? did it avoid banned phrases?).
* **Robustness metrics:** It can be useful to test how changes in input affect output. E.g., if two phrasings of the same question yield wildly different answers, that’s a consistency issue. You can write tests that feed paraphrases of a query and assert that the answers are similar (perhaps by semantic similarity or by both containing some key fact).

**Trace-Based Debugging:** One advantage of DSPy’s structured approach is that you often have access to the internal *trace* of what happened during the agent’s execution. For example, if using `ChainOfThought` or `ReAct`, the returned `Prediction` object often includes a field (commonly `.reasoning` or `.trajectory`) that contains the chain-of-thought the model produced, including tool actions. You can use `dspy.inspect_history()` to print the last interactions with the model, or examine logs. When a test fails (say the agent gave a wrong answer or did something unexpected), inspecting this trace is incredibly useful:

* Did the agent misunderstand the question? (Check the reasoning text).
* Did it choose the wrong tool or not use a tool when it should? (Check the sequence of actions).
* Was there an error in a tool execution? (Maybe the trace or console log shows an exception or an empty result from a tool).

By analyzing traces, you can often pinpoint if the issue is with the prompt (maybe the instructions weren’t clear enough), with the tool (maybe the tool returned something in an unexpected format that confused the agent), or with the model (maybe the model lacks knowledge or made a reasoning error).

**Error Analysis and Iteration:** Testing will reveal some failures or suboptimal outputs. For each issue, consider potential fixes:

* Could this be fixed by providing another example and recompiling (self-improvement)?
* Does it need a code change (like handling a tool exception, or adding an if-condition for a known corner case)?
* Or is it something to document as a limitation?

For instance, if your agent sometimes says "I don't know" when the answer is actually in the context, you might add a heuristic that if the context is non-empty, discourage an "I don't know" response (maybe by tweaking the prompt or adding a post-processing check). Make those changes and re-run the tests.

**Continuous Evaluation:** If possible, integrate these tests into your development cycle. If you retrain or recompile the agent with new data, run the test suite again to ensure nothing that used to work got broken (regressions). Keep some traces of performance metrics over time if the project is long-running.

**Exercise:** *Test and evaluate your agent thoroughly.*

1. **Identify Key Scenarios:** List out important scenarios your agent should handle. For each, craft an example input (or sequence for multi-turn) and the expected outcome. For a chatbot, this could be a short dialogue. For a tool-using agent, it could be a question that definitely requires the tool.
2. **Write Automated Tests:** If you're in a coding environment, implement a few test cases (using `assert` statements or a testing framework) for these scenarios. For example, `assert "Paris" in answer_text` for a geography question about Eiffel Tower, or `assert agent_state.tool_used == "Calculator"` for a math query (you might need to capture that from the agent’s output or logs). If not coding, you can do this mentally or on paper: enumerate what you expect and then run the agent and compare.
3. **Use Deterministic Settings:** Configure your model to reduce randomness for tests – e.g., `dspy.configure(lm=my_lm, default_args={'temperature': 0})` so that by default it uses temperature 0 for generation. Or pass `temperature=0` in your module calls during tests.
4. **Run and Record Results:** Execute each test case and note whether the agent’s output matches expectations. For any failures, inspect why. Use `inspect_history` or print the `Prediction` details to see the reasoning or any error messages. For example, if the output was wrong, was the reasoning wrong or did the agent stop early, etc.
5. **Improve Based on Findings:** Take one failing test and try to fix it. For example, if the agent gave a British spelling but you expected American (maybe your test is too strict), either adjust the expectation or enforce the style via prompt. If the agent failed to use the tool, consider if the prompt to use tools is strong enough. You might adjust the system message like “If the question involves math, **always** use the calculator tool.” Re-run the test to see if it passes now.
6. **Evaluate Metric on Multiple Samples:** If you have, say, 10 sample Q\&A pairs, and your agent gets 8 correct, you have 80% accuracy on that sample – a useful quantitative measure. Compute whatever metric applies (even if informally). This will give you a baseline to compare with future improvements.
7. **Edge Case (Optional):** Throw something odd at your agent to see how it behaves – e.g., an empty question, or a very long input, or a question outside its knowledge. Ensure it fails gracefully (maybe by saying "I don't know" or some safe response rather than crashing or hallucinating). If it doesn’t handle it well, you might add handling (for example, if input is empty, return an error message directly in code).

This testing exercise will bolster your confidence in the agent. By simulating user interactions and verifying outcomes, you are essentially doing quality assurance for your AI. It also reinforces a mindset: treat your prompt-based programs as testable artifacts, not black boxes. With DSPy’s structured outputs and traceability, you have a lot of tools to make AI behavior more transparent and verifiable. In complex projects, a robust test suite can catch regressions when you tweak prompts or upgrade models, saving you from unpleasant surprises in production.

## Module 10: Deploying DSPy Agents

After developing and refining your DSPy agent, the final step is **deployment** – making it available in a production environment where it can interact with users or systems in real-time. In this module, we cover strategies for packaging and deploying DSPy agents, including considerations for performance, scalability, and maintainability. We also discuss runtime considerations like model hosting and monitoring.

**Packaging Your Agent:** A DSPy agent (or program) is essentially Python code, possibly with an optimized state (prompts, few-shot examples, or fine-tuned weights) learned during compilation. You should ensure you can reproduce that state in deployment. DSPy likely provides ways to **save and load** compiled programs or optimizers’ results. For example, if you used `optimizer.compile(...)` and got `optimized_program`, you might serialize it to disk (perhaps via `dspy.save(optimized_program, "agent.dspy")`) and later load it back with `dspy.load("agent.dspy")` – check the documentation for exact APIs. This allows you to avoid re-running the optimizer in production; you deploy the already optimized prompt/model.

You can package the agent into a Python module or package. If it’s part of a larger application (say a web app), include the DSPy setup in the app code. Ensure that your environment in production has all the dependencies: `dspy` library, any model backends (if using local models, the model files need to be present; if using OpenAI API, the API key and network access are needed, etc.), and any tools’ requirements (for instance, if you used `qdrant-client` for retrieval, that needs to be installed and the service endpoint configured).

**Deployment Options:**

* **Command-line or Batch:** For internal use or scheduled jobs, you might run the agent as a script. This is simple: just load the agent and feed inputs from a file or command-line arguments.
* **API Service:** A common approach is to wrap the agent in a web service (REST API or gRPC). For example, using FastAPI or Flask in Python, you can create an endpoint `/ask` that takes a query and returns the agent’s answer. Inside the endpoint, you call your DSPy program. Because DSPy programs are just Python callables once set up, this integration is straightforward. This allows other applications to HTTP-request answers from your agent.
* **Chatbot UI or Integration:** If it’s a user-facing chatbot, you might embed it in a chat interface. You could use a WebSocket or a polling mechanism to send user messages to the server running DSPy and stream back the responses (DSPy’s `StreamListener` or similar can help stream tokens if needed). Integration with platforms like Slack, Discord, or others is possible by writing a bot that calls your agent code.
* **Serverless or Cloud Functions:** For lighter workloads, you could deploy the agent logic as a serverless function (like AWS Lambda or Google Cloud Functions). However, be mindful of cold start times and package sizes, especially if using large models – serverless might be tricky for heavy ML. More commonly, you’d use a persistent service for an AI agent.
* **Edge or On-premise:** If deploying within an organization or on a device, ensure the hardware is up to the task. A smaller model might be needed for edge deployment. You might use quantized models or GPU acceleration as appropriate. DSPy doesn’t prevent you from using optimized model runtimes – for instance, running the model via TensorRT or using an API like Azure OpenAI if company policy requires it.

**Performance and Scalability:**

* **Concurrency:** If multiple users or processes will query the agent simultaneously, you need to handle concurrent execution. Language models calls are typically CPU/GPU-bound operations. You might allow a certain number of threads or processes. If using an external API (OpenAI), concurrency is limited by API rate limits – you might queue requests or scale out by running multiple instances of your service.
* **Latency:** Identify where latency comes from – model inference is usually the majority. If you need faster responses, consider using smaller or more efficient models in production (perhaps the one you fine-tuned). Also, if you compiled with heavy few-shot examples, your prompt might be long – see if you can shorten it (maybe via a concise instruction or a fine-tuned model to avoid few-shot). Using `temperature=0` in production can reduce variability (good for consistent behavior) and sometimes slightly reduce token consumption.
* **Monitoring:** Implement logging to monitor the agent’s activity in production. Log inputs and outputs (with care for sensitive data anonymization if needed), any tool usage and failures, and timing. This will help detect if the agent starts giving wrong answers or if some queries cause slowdowns or errors. If your agent is critical, you might even set up alerts for certain conditions (e.g., if tool calls fail repeatedly or if response time exceeds a threshold).

**Runtime Considerations:**

* **Model Hosting:** If using a local model, ensure the machine has enough memory (RAM/VRAM) for it. Consider using model serving solutions (like HuggingFace’s Text Generation Inference server, or FasterTransformer) to maximize throughput. If using OpenAI, ensure network connectivity and handle exceptions (e.g., retry on rate limit errors).
* **Resource Cleanup:** If your agent uses any external resources (files, network connections, database clients), make sure to handle cleanup or reuse connections between calls to improve performance.
* **Security:** If the agent will act (use tools that can affect systems, like running code or altering databases), be very careful. In production, you’d sandbox or restrict such capabilities. For example, if you allowed a Python tool, you’d want to limit its operations. Or if it can hit an API, ensure it can’t be misused to call arbitrary endpoints.

**Example – Deploying as Web API (pseudo-code):**

```python
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()
agent = dspy.load("my_optimized_agent.dspy")  # Load compiled agent

class Query(BaseModel):
    question: str
    history: list = []  # optionally handle multi-turn

@app.post("/ask")
def ask_agent(query: Query):
    # If using history, convert to dspy.History
    history_obj = dspy.History(messages=query.history) if query.history else None
    try:
        if history_obj:
            result = agent(question=query.question, history=history_obj)
        else:
            result = agent(question=query.question)
        return {"answer": str(result.answer)}
    except Exception as e:
        # Handle errors gracefully
        return {"error": str(e)}
```

This outlines a basic service where clients send a JSON with a question (and optionally past history), and the server returns the answer. You’d run this with Uvicorn or Gunicorn for production. You might also add an endpoint for health checks (return OK) and maybe an admin endpoint to reload the model (if you want to hot-swap a new version with zero downtime by loading a new agent object).

**Exercise:** *Plan and execute a deployment of your DSPy agent.*

1. **Choose a Deployment Mode:** Decide how you would deploy your agent. For many, a web API is a good exercise. If you can, implement a small web service around the agent (like the FastAPI example). If not, conceptually outline how you’d integrate it into an existing product or pipeline.
2. **Resource Planning:** Figure out what resources you need. Does your agent require a GPU at runtime? How much memory? If using an API, what are the expected costs and limits? Document these. For example, *"Using GPT-4 via OpenAI, we estimate \$0.05 per query on average length – need to monitor cost"*, or *"Will use a 7B parameter model on CPU, which may take \~5 seconds per response"*.
3. **Testing in a Staging Environment:** Before full deployment, test the agent in an environment that mimics production. If you wrote an API, run it locally and send some requests (you can use `curl` or a tool like Postman) to ensure it responds correctly. Check error handling by sending something malformed intentionally.
4. **Performance Tuning:** If responses are too slow, consider steps like lowering the model complexity, or adjusting prompt length. For memory-heavy agents, maybe limit the conversation history that’s sent (e.g., only last 5 messages). Apply these tweaks and measure again. Even if you can't fully simulate production load, try a quick loop of 10 calls to see if any issues like memory leaks or slowdowns occur.
5. **Deployment Execution:** Deploy the agent by your chosen method. This might mean pushing the container to a cloud service, or installing on a server, etc. Ensure the DSPy library and necessary model files are included. If using Docker, you’d create a Dockerfile that `pip install dspy` and other deps, copies your code, sets the entrypoint to run the API.
6. **Monitoring Setup:** Plan how you will monitor the agent. If possible, implement simple logging in the deployed version. Run it for a while and inspect logs. For example, log the questions asked (maybe hashed or truncated if privacy is needed) and the time taken and maybe token usage if accessible. This data will help you iterate on the agent even after deployment, by showing real user queries and any failure modes.

By deploying your agent, you complete the journey from concept to a usable AI service. You also encounter practical considerations that require both engineering and ML thinking: keeping the system efficient, reliable, and safe. Many promising prototypes falter in production due to lack of attention to these details – but with the structured approach and testing you’ve applied through this course, your agent is well-positioned to succeed in the real world.

---

Congratulations on completing the course! You’ve learned the theory and practice of building advanced agents with DSPy, from fundamental abstractions to real-world deployment. Along the way, we explored modular programming, retrieval augmentation, tool use, multi-turn interactions, self-improvement via optimization, and integration with other ecosystems. By working through examples like chatbots, planning agents, data extraction pipelines, and more, you’ve gained hands-on experience in constructing AI systems that are **modular**, **reasoned**, and **improvable**.

As next steps, continue to refine your agents and perhaps dive into the research papers behind DSPy for deeper insights. Keep experimenting with new modules and optimizers (the field is evolving rapidly, and new techniques emerge frequently). With the solid foundation you now have, you can build sophisticated, reliable AI agents that tackle real-world problems effectively. Happy coding, and may your agents keep improving themselves!
