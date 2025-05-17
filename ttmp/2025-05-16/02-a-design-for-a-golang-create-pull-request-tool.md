# Design for a Go-based Pull Request Creation Tool

This document outlines the architecture and design for a Go CLI tool that replaces the existing bash script for generating pull request descriptions using an LLM. The tool will support both direct CLI interaction and a TUI mode.

## 1. Goals

- Provide a robust and maintainable Go application for generating PR descriptions.
- Replicate and enhance the functionality of the existing bash script.
- Offer both a command-line interface (CLI) and a terminal user interface (TUI).
- Employ a clean architecture for better separation of concerns and testability.
- Allow configuration for different LLM providers and settings.

## 2. Architecture

We will adopt a Clean Architecture approach, dividing the application into the following layers:

- **Domain:** Contains the core business logic and entities. This layer is independent of any frameworks or external services.
  - Entities: `PullRequestSpec` (title, body, changelog, release notes), `Diff`, `CommitInfo`.
  - Use Cases: `GeneratePullRequestDescription`, `ValidatePullRequestSpec`.
- **Application:** Orchestrates the use cases from the Domain layer. It handles application-specific logic and acts as an intermediary between the Presentation and Domain layers.
  - Services: `PullRequestService` (coordinates fetching diffs, interacting with LLM, formatting output).
- **Infrastructure:** Implements external concerns like interacting with Git, file system, LLM APIs, and the GitHub CLI (`gh`).
  - Adapters: `GitAdapter` (for `git diff`, `git fetch`), `LlmAdapter` (to call `pinocchio code create-pull-request`), `GitHubCLIAdapter` (to call `gh pr create`), `FileSystemAdapter`.
- **Presentation:** Handles user interaction. This will include the CLI (using Cobra) and the TUI (e.g., using BubbleTea).
  - CLI: Cobra commands and flags.
  - TUI: Views for displaying diffs, editing PR descriptions, confirming actions.

```
+---------------------+     +-----------------------+     +---------------------+     +------------------------+
|    Presentation     |<--->|      Application      |<--->|       Domain        |<--->|     Infrastructure     |
| (CLI - Cobra, TUI)  |     | (PullRequestService)  |     | (Entities, UseCases)|     | (Git, LLM, GH CLI, FS) |
+---------------------+     +-----------------------+     +---------------------+     +------------------------+
```

## 3. CLI Design (Cobra)

The main command will be `gopr` (Go Pull Request). The design prioritizes automation and user-friendliness, gathering common information by default.

### 3.1. `gopr create`

This will be the primary command to create a pull request description.

**Usage:**

```bash
gopr create [flags] <description>
```

**Alias:** `gopr c`

**Arguments:**

- `description` (string, required): The initial high-level description of the pull request.

**Key Flags:**

- `--branch <string>` (default: "origin/main"): Target branch to diff against. This is used for generating both the diff and the commit history.
- `--issue <string>`: Issue reference (e.g., number, URL). If a number is provided, the tool can optionally attempt to fetch issue details using `gh issue view <id>` to enrich the LLM prompt.
- `--title <string>`: A suggested title for the pull request. If not provided, the LLM will generate it.
- `--output-file <file_path>` (default: "/tmp/pr.yaml"): File to store the generated PR YAML.
- `--non-interactive` (bool, default: false): Skip all interactive prompts for confirmation and editing.
- `--tui` (bool, default: false): Launch the Terminal User Interface for a more guided experience.

**Diff Customization Flags:**

- `--diff-file <file_path>`: **Override** automatic diff generation by providing a specific diff file.
- `--from-clipboard` (bool, default: false): Use diff from clipboard instead of generating it. This takes precedence over `--diff-file` and automatic generation.
- `--exclude <comma_separated_list>`: Files/patterns to exclude from the auto-generated diff.
- `--diff-context-size <int>` (default: 3): Set context size for `git diff -U<int>`. Use 1 for short, 10 for long, etc.
- `--only <comma_separated_paths>`: Include specific paths only in the auto-generated diff.
- `--no-tests` (bool): Exclude common test file patterns from the auto-generated diff.
- `--no-package` (bool): Exclude common package manager files (e.g., `go.sum`, `package-lock.json`) from the auto-generated diff.

**Commit History Flags:**

- `--commits-file <file_path>`: **Override** automatic commit history gathering by providing a file containing commit messages.
- `--no-commits` (bool, default: false): Do not include commit history in the context for the LLM.

**LLM Customization Flags:**

- `--llm-command <string>` (default: "pinocchio code create-pull-request"): Command to use for LLM interaction.
- `--llm-style <style_name>`: Predefined style for the LLM output (e.g., "default", "concise-bullets", "narrative"). This would map to a combination of internal prompt adjustments (like requesting bullets, conciseness).
  - _Alternative to individual flags like `--concise`, `--use-bullets` which are now considered advanced/override flags._
- `--llm-param <key=value_list>`: Pass additional raw parameters directly to the LLM prompt template (e.g. `--llm-param concise=true --llm-param use_bullets=true`). This provides fine-grained control.
- `--code-context <file_path_list>`: Provide specific code files as additional context to the LLM.
- `--additional-system-prompt <string>`: Additional system prompt for the LLM.
- `--additional-user-prompt <string_list>`: Additional user prompt content for the LLM.
- `--context-files <file_path_list>`: Additional arbitrary files to provide as context to the LLM.

**Workflow (Revised for Automation):**

1.  Parse flags and arguments. If `--tui` is present, launch TUI mode (see section 4).
2.  **Gather Git Diff:**
    - If `--from-clipboard`, use clipboard content.
    - Else if `--diff-file` is provided, use content from this file.
    - Else, automatically generate `git diff` using `GitAdapter` against `--branch`, applying `--exclude`, `--diff-context-size`, `--only`, `--no-tests`, `--no-package` flags.
3.  **Gather Commit History:**
    - If `--no-commits` is false and `--commits-file` is not provided, automatically fetch commit history (e.g., `git log <branch>..HEAD`) using `GitAdapter`.
    - If `--commits-file` is provided, use its content.
4.  **Gather Issue Details (Optional):**
    - If `--issue` is provided and looks like an ID, optionally use `GitHubCLIAdapter` to fetch issue title/body to add to LLM context.
5.  Display token count of the diff and commit messages (if gathered).
6.  **Interactive Diff Review (if not `--non-interactive`):**
    - Prompt user to proceed, view diff, or edit diff.
    - If view: display diff using pager.
    - If edit: open diff content (from file or in-memory temp file) in `$EDITOR`, read back, update token count.
7.  **Construct LLM Prompt:**
    - Use the provided `description`, processed `diff`, `commit_history`, and potentially `issue_details`.
    - Apply `--llm-style` and pass through `--llm-param` flags, as well as context from `--code-context`, `--additional-system-prompt`, `--additional-user-prompt`, and `--context-files`.
8.  Call `LlmAdapter` (e.g., `pinocchio code create-pull-request ...`) to generate the PR description YAML.
9.  Store the output YAML in `--output-file` (e.g., `/tmp/pr.yaml`).
10. **Interactive YAML Review (if not `--non-interactive`):**
    - Show the path to the YAML file.
    - Prompt user to proceed with `gh pr create`, edit the YAML, or abort.
    - If edit: open YAML in `$EDITOR`.
11. Extract title and body from the (potentially edited) YAML.
12. Call `GitHubCLIAdapter` to execute `gh pr create --title "<title>" --body "<body>"`.

### 3.2. `gopr create-from-yaml`

(Largely unchanged, focuses on a specific file)

**Usage:**

```bash
gopr create-from-yaml [flags] [<yaml_file_path>]
```

**Alias:** `gopr cfy`

**Arguments:**

- `yaml_file_path` (string, optional): Path to the PR YAML file. Defaults to `/tmp/pr.yaml` if not provided.

**Flags:**

- `--non-interactive` (bool, default: false): Skip interactive prompts for confirmation and editing.

**Workflow:**

1.  Determine YAML file path (argument or default).
2.  Read the YAML file using `FileSystemAdapter`.
3.  Display the content of the YAML file (or relevant parts).
4.  **Interactive YAML Review (if not `--non-interactive`):**
    - Prompt user to proceed, edit the YAML, or abort.
    - If edit: open YAML in `$EDITOR`.
5.  Extract title and body from the (potentially edited) YAML.
6.  Call `GitHubCLIAdapter` to execute `gh pr create --title "<title>" --body "<body>"`.

### 3.3. `gopr get-diff`

(Largely unchanged, but ensures consistency with `create` command's diff flags)

**Usage:**

```bash
gopr get-diff [flags]
```

**Flags:**

- `--branch <string>` (default: "origin/main")
- `--exclude <comma_separated_list>`
- `--diff-context-size <int>` (default: 3)
- `--only <comma_separated_paths>`
- `--no-tests` (bool)
- `--no-package` (bool)
- `--stat` (bool, default: false): Show `git diff --stat` output.
- `--output <file_path>`: Optionally write diff to a file instead of stdout.

**Workflow:**

1.  Use `GitAdapter` to generate `git diff` based on flags.
2.  If `--stat`, also get and print stat output.
3.  Print diff to stdout or write to `--output` file.

## 4. TUI Design (Conceptual - e.g., BubbleTea)

The TUI mode, invoked by `gopr create --tui <description>`, aims for a highly guided and interactive experience.

**Key Views/Components (Revised):**

- **Welcome / Initial Input View:**
  - Input for the PR `description`.
  - Simple options for common overrides: `target branch`, `issue ID`.
  - Button to "Start Analysis".
- **Context Gathering View (Dynamic Progress):**
  - Shows steps being performed:
    - "Fetching diff from `<branch>`..." (with spinner)
    - "Analyzing N excluded files..." (if any)
    - "Fetching commit history (M commits)..." (with spinner)
    - "Fetching issue details for #<id>..." (if applicable, with spinner)
  - Option to "Skip" steps or "Provide Manually" (e.g., skip auto-diff, point to a diff file).
- **Diff Review & Edit View:**
  - Displays the generated (or provided) `git diff` (scrollable).
  - Key stats: token count, lines changed.
  - Actions:
    - `Looks Good, Proceed to LLM`: Continue.
    - `Edit Diff`: Opens `$EDITOR`. After saving, diff is re-analyzed.
    - `Refine Diff Parameters`: (Advanced) Go back to change exclusion rules, branch, etc.
    - `Copy Diff`: To clipboard.
    - `Cancel`: Abort.
- **LLM Configuration View (Optional / Collapsible):**
  - Select `--llm-style`.
  - Advanced section for: `--llm-param`, `--code-context`, `--additional-system-prompt`, etc.
  - Defaults are usually sufficient.
- **LLM Interaction View:**
  - Shows a spinner: "Generating PR description with LLM..."
  - Displays LLM command being used (if verbose mode is on).
- **PR YAML Review & Edit View:**
  - Cleanly displays: `Title`, `Body`, `Changelog`, `Release Notes` (all editable).
  - Actions:
    - `Create PR on GitHub`: Proceeds to call `gh pr create`.
    - `Edit Raw YAML`: Opens the full YAML file in `$EDITOR`.
    - `Refine with LLM`: (Advanced) Go back to LLM config / re-prompt.
    - `Save YAML & Exit`: Saves to `/tmp/pr.yaml` and exits.
    - `Cancel`: Abort.
- **Final Status View:**
  - Result of `gh pr create` (link to PR or error message).

**Workflow (TUI - Revised):**

1.  **Welcome View:** User provides initial PR description, optionally branch/issue.
2.  **Context Gathering:** Tool automatically fetches diff, commits, issue details, showing progress. User can intervene if needed.
3.  **Diff Review:** User reviews, optionally edits/refines, and confirms the diff.
4.  **(Optional) LLM Config:** User can tweak LLM style or advanced parameters.
5.  **LLM Interaction:** Tool calls LLM.
6.  **PR YAML Review:** User reviews and edits (Title, Body, etc.) in a structured way or raw YAML.
7.  User confirms to create the PR on GitHub.
8.  **Status View:** Displays success or failure.

## 5. Core Logic (API/Pseudocode)

### 5.1. `GitAdapter` (Revised)

```go
package infrastructure

import "context" // Added import

type GitAdapter interface {
    // GetDiff fetches the git diff based on parameters.
    GetDiff(ctx context.Context, branch string, exclusions []string, contextSize int, includePaths []string, noTests bool, noPackage bool) (string, error)
    // GetDiffStat fetches the git diff --stat.
    GetDiffStat(ctx context.Context, branch string, exclusions []string, includePaths []string, noTests bool, noPackage bool) (string, error)
    // FetchOrigin runs `git fetch origin`.
    FetchOrigin(ctx context.Context) error
    // GetDiffFromClipboard reads diff from system clipboard.
    GetDiffFromClipboard(ctx context.Context) (string, error)
    // GetCommitHistory fetches commit messages from a specified range (e.g., origin/main..HEAD).
    GetCommitHistory(ctx context.Context, baseBranch string, head string) ([]string, error) // head typically "HEAD"
}

// Implementation would use os/exec to call git commands.
```

### 5.2. `LlmAdapter` (Minor change for flexibility)

```go
package infrastructure

import "context" // Added import

type LlmAdapter interface {
    // GeneratePullRequestSpec calls the LLM (e.g., pinocchio) to generate PR details.
    // promptParams now more generic to accommodate various LLM inputs.
    GeneratePullRequestSpec(ctx context.Context, llmCommand string, promptParams map[string]interface{}, diffContent string, commitHistoryContent string, otherContext map[string]string) (rawYamlOutput string, err error)
    // CountTokens calls the LLM's token counter for given text.
    CountTokens(ctx context.Context, llmCommand string, textContent string) (int, error)
}

// Implementation would use os/exec to call the specified llmCommand.
// It needs to be robust in constructing the actual prompt/command for the LLM
// based on llmCommand and promptParams.
```

### 5.3. `GitHubCLIAdapter` (Revised for issue fetching)

```go
package infrastructure

import "context" // Added import

type GitHubCLIAdapter interface {
    // CreatePullRequest runs `gh pr create`.
    CreatePullRequest(ctx context.Context, title string, body string) (string /* PR URL */, error)
    // GetIssueDetails fetches basic details for a given issue ID using `gh issue view <id>`.
    GetIssueDetails(ctx context.Context, issueID string) (title string, body string, err error)
}

// Implementation would use os/exec to call `gh` command.
```

### 5.4. `PullRequestService` (Application Layer - Revised)

```go
package application

import (
	"context"
	"gopr/internal/domain"       // Assuming domain types are here
	"gopr/internal/infrastructure" // Assuming adapter interfaces are here
)


type CreatePullRequestConfig struct {
    Description           string
    Branch                string
    IssueID               string
    Title                 string // User-suggested title
    OutputFile            string
    NonInteractive        bool
    TUI                   bool

    DiffFile              string
    FromClipboard         bool
    DiffExclusions        []string
    DiffContextSize       int
    DiffOnlyPaths         []string
    DiffNoTests           bool
    DiffNoPackage         bool

    CommitsFile           string
    NoCommits             bool

    LlmCommand            string
    LlmStyle              string
    LlmParams             map[string]string // For raw key=value params
    CodeContextFiles      []string
    AdditionalSystemPrompt string
    AdditionalUserPrompts []string
    ContextFiles          []string
}

type PullRequestService struct {
    gitAdapter       infrastructure.GitAdapter
    llmAdapter       infrastructure.LlmAdapter
    githubCliAdapter infrastructure.GitHubCLIAdapter
    fsAdapter        infrastructure.FileSystemAdapter // For reading/writing files, editor interaction
    // userInput        infrastructure.UserInputAdapter // For interactive prompts (CLI/TUI agnostic)
}

func NewPullRequestService(
    git infrastructure.GitAdapter,
    llm infrastructure.LlmAdapter,
    gh infrastructure.GitHubCLIAdapter,
    fs infrastructure.FileSystemAdapter,
) *PullRequestService {
    return &PullRequestService{
        gitAdapter:       git,
        llmAdapter:       llm,
        githubCliAdapter: gh,
        fsAdapter:        fs,
    }
}


// CreatePullRequest orchestrates the PR creation process.
func (s *PullRequestService) CreatePullRequest(
    ctx context.Context,
    config CreatePullRequestConfig,
) (*domain.PullRequestSpec, string /* prURL */, error) {
    var diffContent string
    var commitMessages []string
    var issueTitle, issueBody string
    var err error

    // 1. Get Diff
    if config.FromClipboard {
        diffContent, err = s.gitAdapter.GetDiffFromClipboard(ctx)
    } else if config.DiffFile != "" {
        diffContentBytes, ferr := s.fsAdapter.ReadFile(config.DiffFile)
        if ferr != nil { /* handle error */ }
        diffContent = string(diffContentBytes)
    } else {
        // Auto-generate diff
        err = s.gitAdapter.FetchOrigin(ctx) // Ensure remote refs are up-to-date
        if err != nil { /* handle error */ }
        diffContent, err = s.gitAdapter.GetDiff(ctx, config.Branch, config.DiffExclusions, config.DiffContextSize, config.DiffOnlyPaths, config.DiffNoTests, config.DiffNoPackage)
    }
    if err != nil { /* handle error: failed to get diff */ return nil, "", err }

    // 2. Get Commit History
    if !config.NoCommits {
        if config.CommitsFile != "" {
            commitsContentBytes, ferr := s.fsAdapter.ReadFile(config.CommitsFile)
            if ferr != nil { /* handle error */ }
            // Assuming one commit message per line, or parse as needed
            commitMessages = parseCommitMessages(string(commitsContentBytes))
        } else {
            commitMessages, err = s.gitAdapter.GetCommitHistory(ctx, config.Branch, "HEAD")
            if err != nil { /* handle error: failed to get commits */ }
        }
    }

    // 3. Get Issue Details (if issue ID provided)
    if config.IssueID != "" {
        // Basic check if it's a numeric ID or full URL.
        // If gh can handle URL as well, then simpler.
        issueTitle, issueBody, err = s.githubCliAdapter.GetIssueDetails(ctx, config.IssueID)
        if err != nil { /* log warning, but proceed */ }
    }

    // TODO: Convert commitMessages ([]string) to a single string for LLM
    commitHistoryStr := joinCommitMessages(commitMessages)


    // 4. User interaction for diff (if not non-interactive and not TUI)
    //    - View diff (s.fsAdapter.ViewWithPager(diffContent))
    //    - Edit diff (editedDiff, err := s.fsAdapter.EditTempFile(diffContent, ".diff"); diffContent = editedDiff)
    //    - Token count (s.llmAdapter.CountTokens for diffContent and commitHistoryStr)

    // 5. Prepare LLM prompt parameters
    llmPromptParams := map[string]interface{}{
        "description": config.Description,
        "user_title_suggestion": config.Title, // LLM can use or ignore this
        "llm_style": config.LlmStyle, // Internal mapping to specific instructions
        "raw_params": config.LlmParams, // For direct passthrough
        "additional_system_prompt": config.AdditionalSystemPrompt,
        "additional_user_prompts": config.AdditionalUserPrompts,
        // "issue_title": issueTitle, // if fetched
        // "issue_body": issueBody,   // if fetched
    }

    otherContext := make(map[string]string)
    if issueTitle != "" { otherContext["issue_title"] = issueTitle }
    if issueBody != "" { otherContext["issue_body"] = issueBody }
    // Add content from config.CodeContextFiles and config.ContextFiles to otherContext
    // for file := range config.CodeContextFiles { content, _ := s.fsAdapter.ReadFile(file); otherContext[file] = content }
    // for file := range config.ContextFiles { content, _ := s.fsAdapter.ReadFile(file); otherContext[file] = content }


    // 6. Call LLM
    rawYaml, err := s.llmAdapter.GeneratePullRequestSpec(ctx, config.LlmCommand, llmPromptParams, diffContent, commitHistoryStr, otherContext)
    if err != nil { /* handle LLM error */ return nil, "", err }

    err = s.fsAdapter.WriteFile(config.OutputFile, []byte(rawYaml))
    if err != nil { /* handle error writing output file */ return nil, "", err }

    // 7. Parse YAML to domain.PullRequestSpec
    var prSpec domain.PullRequestSpec
    // err = yaml.Unmarshal([]byte(rawYaml), &prSpec) // Use a Go YAML library
    if err != nil { /* handle YAML parsing error */ return nil, "", err }

    currentPRTitle := prSpec.Title
    currentPRBody := prSpec.Body

    // 8. User interaction for YAML (if not non-interactive and not TUI)
    //    - Display path to config.OutputFile
    //    - Prompt to proceed, edit, or abort
    //    - If edit: newYamlContent, err := s.fsAdapter.EditFile(config.OutputFile)
    //             re-parse YAML: yaml.Unmarshal(newYamlContent, &prSpec)
    //             currentPRTitle = prSpec.Title; currentPRBody = prSpec.Body


    // 9. Create PR on GitHub
    prURL, err := s.githubCliAdapter.CreatePullRequest(ctx, currentPRTitle, currentPRBody)
    if err != nil { /* handle GitHub PR creation error */ return nil, "", err }

    return &prSpec, prURL, nil
}

// Helper (conceptual)
func parseCommitMessages(content string) []string { /* ... */ return nil }
func joinCommitMessages(messages []string) string { /* ... */ return "" }

```

### 5.5. `Domain.PullRequestSpec`

(Unchanged from previous version, still relevant)

```go
package domain

type PullRequestSpec struct {
    Title        string        `yaml:"title"`
    Body         string        `yaml:"body"`
    Changelog    string        `yaml:"changelog"`
    ReleaseNotes ReleaseNotes `yaml:"release_notes"`
}

type ReleaseNotes struct {
    Title string `yaml:"title"`
    Body  string `yaml:"body"`
}
```

## 6. Configuration

- **LLM Command:** Configurable via flag (`--llm-command`) or an environment variable (e.g., `GOPR_LLM_COMMAND`).
- **Default Branch:** Could be configurable globally (e.g., in a `~/.config/gopr/config.yaml`) or per-project (`.git/gopr/config.yaml`).
- **Editor:** Uses `$EDITOR` environment variable.
- **Pager:** Uses `$PAGER` environment variable.
- **LLM Styles:** Could be defined in the global config file (`~/.config/gopr/config.yaml`) allowing users to create their own named styles mapping to sets of LLM parameters.

## 7. File Structure (Conceptual)

```
gopr/
├── cmd/
│   ├── gopr/ # Main application
│   │   ├── main.go
│   │   ├── cli/      # Cobra command definitions
│   │   │   ├── root.go
│   │   │   ├── create.go
│   │   │   ├── create_from_yaml.go
│   │   │   └── get_diff.go
│   │   └── tui/      # TUI components (BubbleTea)
│   │       ├── views/
│   │       └── models.go
├── internal/
│   ├── domain/
│   │   ├── pullrequest.go  // Entities (PullRequestSpec)
│   │   └── service.go      // Use cases (interfaces)
│   ├── application/
│   │   └── pullrequest_service.go // Service implementation
│   ├── infrastructure/
│   │   ├── git/
│   │   │   └── adapter.go
│   │   ├── llm/
│   │   │   └── adapter.go
│   │   ├── github/
│   │   │   └── adapter.go
│   │   ├── filesystem/
│   │   │   └── adapter.go
│   │   └── shell/          // Utility for running external commands
│   │       └── exec.go
│   └── config/
│       └── config.go       // Configuration loading
├── go.mod
├── go.sum
└── README.md
```

## 8. Key Dependencies (Go Modules)

- `github.com/spf13/cobra` (for CLI)
- `github.com/charmbracelet/bubbletea` (for TUI - optional)
- `github.com/charmbracelet/bubbles` (for TUI components - optional)
- `gopkg.in/yaml.v3` (for parsing LLM output)
- Potentially `github.com/cli/go-gh` for a more robust way to interact with `gh` if direct command execution is not preferred.
- `github.com/atotto/clipboard` (for clipboard operations)
- `github.com/spf13/viper` (for managing configuration, including LLM styles from files)

## 9. Next Steps / Open Questions

- **Error Handling:** Define a consistent error handling strategy (e.g., using `github.com/pkg/errors`).
- **Logging:** Implement logging for debugging and auditing.
- **Testing:** Plan for unit tests (especially for domain and application layers) and integration tests (for infrastructure adapters).
- **Interactivity Details:** Refine the exact prompts and interaction flows, especially for editing diffs and YAML.
- **LLM Abstraction:** Consider if a more generic LLM interface is needed if `pinocchio` is just one of many potential backends. The current design assumes `pinocchio`'s specific command structure.
- **`yq` dependency:** The Go application should parse the YAML itself rather than shelling out to `yq`. The `GitHubCLIAdapter` should take title and body as arguments.
- **LLM Style Definition:** Flesh out how `--llm-style` maps to concrete prompt changes or parameters. This could involve a configuration system (e.g., Viper) to load style definitions.
- **User Input Abstraction:** For CLI interactive prompts (confirmations, choices), consider a simple `UserInputAdapter` to make the `PullRequestService` testable without direct terminal interaction. The TUI would have its own way of handling these interactions.
- **Robust `os/exec` Handling:** Ensure proper error capturing, context cancellation, and input/output streaming for all external commands (`git`, `gh`, LLM command).

This design provides a foundational structure. Implementation will involve filling in the details for each component.
