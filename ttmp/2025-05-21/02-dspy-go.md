## Porting **DSPy** to Go – Architecture & Design Document

_(v0.9 – May 2025)_

---

### 0 · Executive summary

DSPy is a declarative, self-optimising framework for building modular AI programs on top of large-language-model (LLM) calls ([GitHub][1]).
The goal of this project is to create **dspy-go**, a faithful, idiomatic Go 1.23 implementation that preserves the mental model of “programming—not prompting” while exploiting Go’s strengths: static typing, strong tooling, excellent concurrency primitives, and first-class observability.

---

### 1 · Goals & non-goals

| **Goals**                                                                                                                         | **Non-goals**                                                                                                                      |
| --------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| 1. Feature parity with DSPy ≥ v2.6 for: Modules/Signatures, LLM abstraction, Retrieval, Assertions, Optimisers, and basic Agents. | Binary compatibility with Python code.                                                                                             |
| 2. Idiomatic Go API (`go vet` + `golangci-lint` clean; generics-friendly).                                                        | Re-implementing every experimental optimiser in the Python repo. We will port the most used ones (Bootstrap, MIPRO, Oracle) first. |
| 3. Pluggable back-ends (interface + mock only for now) & vector stores (interface + mock only).                                   | Building our own vector DB.                                                                                                        |
| 4. Built-in structured logging, streaming & cancellation (`context.Context`).                                                     | A GUI/IDE.                                                                                                                         |

---

### 2 · High-level architecture

```text
┌─────────────────────────────────────────────────────┐
│        cmd/dspy-go (CLI, examples, REPL)            │
└─────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────┐
│  dspy/                                            │
│  ├─ core/          - Signatures, Module graph      │
│  ├─ lm/            - LLM interface impls           │
│  ├─ retrieve/      - Retriever interface + stores  │
│  ├─ assert/        - Assertion DSL                 │
│  ├─ optimize/      - Optimisers & schedulers       │
│  ├─ compile/       - Prompt compiler & cache       │
│  ├─ agent/         - Simple agent loops            │
│  └─ otel/          - Tracing helpers               │
└─────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────┐
│  adapters/ (optional)                              │
│  ├─ python/ (cgo+gopy) – thin bridge               │
│  ├─ rust/   (FFI) – future                         │
│  └─ grpc/   – network boundary                     │
└─────────────────────────────────────────────────────┘
```

---

### 3 · Key design decisions

| #   | Decision                                                                                                                                 | Rationale                                                                             |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| D1  | **Generics for Signatures**: `type Signature[T any] interface { Args() T; Returns() any }`                                               | Keeps compile-time safety while allowing arbitrary structs for arguments/outputs.     |
| D2  | **Functional‐options pattern** for configuration (instead of global `DSPy.configure`).                                                   | Plays well with Go’s preference for explicit configuration, promotes testability.     |
| D3  | **Prompt compilation** is a pure function cached by a SHA256 of (module-graph, optimiser params, backend ID).                            | Enables transparent memoisation and reproducible builds.                              |
| D4  | **Streaming** everywhere: `Predict(ctx context.Context, in any) (<-chan Chunk, error)` plus helper to collect full string.               | Enables server-side streaming (OpenAI SSE, Anthropic) and chunk-wise post-processing. |
| D5  | **Context cancellation is authoritative**; all long-running calls must honour it.                                                        | Aligns with Go’s concurrency conventions.                                             |
| D6  | **Observability first**: every public operation emits structured `slog` logs and spans. Build tags allow users to compile without OTEL.  | Easier production debugging; negligible overhead when disabled.                       |
| D7  | **No runtime reflection in hot paths**. Compile-time code-gen (`go generate`) turns user-defined struct tags into efficient marshallers. | Avoids the cost of `encoding/json` reflection when assembling prompts.                |

---

### 4 · Domain-object mapping

| **DSPy (Python)**                                      | **dspy-go construct**                                                         | Notes                                                                                   |
| ------------------------------------------------------ | ----------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `Signature` class with `input_fields`, `output_fields` | Generic struct implementing `dspy.Signature`                                  | Uses struct tags to mark fields (`dspy:"input"`, `dspy:"output"`).                      |
| `Module` base class, `__call__` overriding             | `type Module interface { Forward(ctx context.Context, in any) (any, error) }` | Concrete modules are plain structs; method names chosen to avoid clash with Go builtin. |
| `Compiler` that walks a module graph                   | `compile.Builder` + `compile.Engine`                                          | Produces `PromptPlan` objects cached in BoltDB/Filecache.                               |
| `LM` object (`OpenAI`, `LlamaCpp`, …)                  | `lm.Client` interface with `Chat`, `Complete`, `Embed` methods                | Concrete back-ends in sub-packages (`lm/openai`, `lm/anthropic`, …).                    |
| Optimisers (`MIPRO`, `Bootstrap`, …)                   | `optimize.Algorithm` interface                                                | Each optimiser executes concurrently via worker-pools, returns updated `PromptPlan`.    |
| Assertions (`dspy.Assert("in result", ...)`)           | `assert.Checker` interface + fluent helpers                                   | Errors bubble up and are collected by optimiser.                                        |

---

### 5 · Detailed package sketches

#### 5.1 `core`

```go
package core

type Signature interface {
    Args() any
    Returns() any
}

type Module interface {
    // Forward executes the module or its compiled prompt.
    Forward(ctx context.Context, in any) (any, error)
}

type PlanID [32]byte // sha256
```

#### 5.2 `lm`

```go
type Client interface {
    Chat(ctx context.Context, msgs []Message, opts ...Option) ([]Chunk, error)
    Embed(ctx context.Context, texts []string, opts ...Option) ([][]float32, error)
}
```

Back-ends implement `Client`. Rate limiting & retries live in decorator layers (`middleware` pattern). Streaming uses `Chunk` struct with role/content/index fields.

#### 5.3 `retrieve`

Defines a `Retriever` interface:

```go
type Retriever interface {
    Search(ctx context.Context, query string, k int) ([]Doc, error)
}
```

Implementations:

- `retrieve.Ephemeral` – naïve in-mem cosine for tests
- `retrieve.Pinecone`, `retrieve.Qdrant`, `retrieve.Lance` – production

#### 5.4 `optimize`

Optimiser contract:

```go
type Algorithm interface {
    Tune(ctx context.Context, plan *compile.PromptPlan, data []Example) error
}
```

Schedulers coordinate parallel tuning jobs via `errgroup.Group`.

---

### 6 · Concurrency model

- Module graphs are executed **breadth-first**.

  - Each node can run in its own goroutine; dependencies resolved with channels.

- The optimiser uses **worker pools** sized by `GOMAXPROCS`.
- Vector searches use context-aware clients; we do _not_ introduce our own `sync.Pool` unless profiling shows benefit.

---

### 9 · Testing plan

| Layer       | Strategy                                                                             |
| ----------- | ------------------------------------------------------------------------------------ |
| Core types  | Pure unit tests, generics instantiation checks.                                      |
| LLM clients | Stub transport returning canned SSE streams; golden tests.                           |
| Compiler    | Snapshot tests: module graph → plan SHA.                                             |
| Optimisers  | Deterministic random seeds; assert monotonic eval gain.                              |
| End-to-end  | Docker Compose with Qdrant + OpenAI stub; run notebooks from DSPy docs ported to Go. |

---

### 10 · Migration roadmap

| Phase | Milestone               | Deliverable                             | ETA   |
| ----- | ----------------------- | --------------------------------------- | ----- |
| 0     | Bootstrap               | repo scaffold, CI, lint, Makefile       | +2 w  |
| 1     | Core & LLM              | `core`, `lm/openai`, `cmd/dspy-go chat` | +6 w  |
| 2     | Compiler                | `compile` + caching                     | +10 w |
| 3     | Retrieval               | `retrieve/*`, demo RAG pipeline         | +14 w |
| 4     | Assertions & Optimisers | `assert`, `optimize/bootstrap`          | +18 w |
| 5     | Agents, Examples        | CLI REPL, ported tutorials              | +22 w |
| 6     | Beta release            | v0.1.0-beta, docs site                  | +24 w |

---

### 11 · Open questions

1. **Macro language**: Should we introduce a small code-gen DSL (à la `sqlc`) to reduce boilerplate for signatures, or rely solely on struct tags?
2. **Fine-tuning**: Do we expose weights fine-tuning via LoRA adapters (needs CGO for `ggml`), or treat that as out-of-scope?
3. **Python interop**: Which direction is required most: calling Go modules from Python notebooks, or vice-versa? Determines priority of adapters.

---

### 12 · Conclusion

This document lays out an opinionated yet flexible blueprint for bringing DSPy’s declarative, self-improving paradigm to the Go ecosystem. By leaning on Go’s type system, concurrency story, and industry-standard observability stack, **dspy-go** aims to offer production engineers a first-class alternative without sacrificing the rapid iteration cycle that made DSPy popular in Python.

[1]: https://github.com/stanfordnlp/dspy "GitHub - stanfordnlp/dspy: DSPy: The framework for programming—not prompting—language models"
