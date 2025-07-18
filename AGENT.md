# AGENT.md

## Project Structure

- only a single go.mod file at the root
- package name is github.com/go-go-golems/go-go-labs at the root
- new go programs go into cmd/XXX per default (package name github.com/go-go-golems/go-go-labs/cmd/XXX)

## Code Style Guidelines

- Go: Uses gofmt, go 1.23+, github.com/pkg/errors for error wrapping
- Go: Uses zerolog for logging, cobra for CLI, viper for config
- Go: Follow standard naming (CamelCase for exported, camelCase for unexported)
- Python: PEP 8 formatting, uses logging module for structured logging
- Python: Try/except blocks with specific exceptions and error logging
- Use interfaces to define behavior, prefer structured concurrency
- Pre-commit hooks use lefthook (configured in lefthook.yml)

<goGuidelines>
- When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
- When building web applications, use htmx, bootstrap and the templ templating language.
- Always use a context argument when appropriate.
- Use cobra for command-line applications.
- Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
- Use github.com/pkg/errors for wrapping errors.
- When starting goroutines, use errgroup.
- Don't create new go.mod in the subdirectories, instead rely on the top level one
- Create apps in self-contained folders, usually in cmd/apps or in cmd/experiments
- Only use the toplevel go.mod, don't create new ones.
- When writing a new experiment / app, add zerolog logging to help debug and figure out how it works, add --log-level flag to set the log level.
</goGuidelines>

<webGuidelines>
- Use bun, react and rtk-query. Use typescript.
- Use bootstrap for styling.
- Store css, html and js in different files in a static directory.
- Use go:embed to serve static files.
- Use templ for go templates, assume I'm running templ generate -watch in the background.
- Always serve static files under /static/ URL paths, never directly under functional paths like /admin/
</webGuidelines>

<terminalUIGuidelines>
Use VHS by charmbracelet to create gif animations. The version installed can take .txt/.ansi screenshots, so use that to validate things working correctly.
Use this to debug TUI applications and create demo gifs. Use txt screenshots to validate the UI working correctly by looking at the screenshot.
</terminalUIGuidelines>

<debuggingGuidelines>
If me or you the LLM agent seem to go down too deep in a debugging/fixing rabbit hole in our conversations, remind me to take a breath and think about the bigger picture instead of hacking away. Say: "I think I'm stuck, let's TOUCH GRASS".  IMPORTANT: Don't try to fix errors by yourself more than twice in a row. Then STOP. Don't do anything else.

</debuggingGuidelines>

<generalGuidelines>
Don't add backwards compatibility layers unless explicitly asked.

If it looks like your edits aren't applied, stop immediately and say "STOPPING BECAUSE EDITING ISN'T WORKING".

Run the format_file tool at the end of each response.
</generalGuidelines>

<concurrencyGuidelines>
When building caches or collections with per-item locking:

1. **Strict Lock Hierarchy**: Always acquire global mutex before per-item mutex (`globalMu → itemMu`)
2. **Single Ownership**: Each data structure should have one clear owner of its mutex
3. **Minimize Critical Sections**: Release locks before calling functions that may re-acquire them
4. **Prefer RWMutex**: Use read locks for readers, write locks for writers
5. **Test with Race Detector**: Always run `go test -race` to catch data races

**Common Pitfalls**:
- Calling methods that re-acquire locks while holding locks
- Updating shared state without proper locking
- Mixed read/write lock usage on same data
</concurrencyGuidelines>

<gitGuidelines>
After each successful run, output a git commit message as yaml, using gitmoji for th title.
If new APIs, endpoints or CLI commands were added, document them in the commit message.
Make a list of the tests that were run and their results.

Avoid general statements like "created extensible foundation" or "this allows for...".

title: <title>
description: <description>

and store in directory root as .git-commit-message.yaml
</gitGuidelines>
