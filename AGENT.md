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
</goGuidelines>

<webGuidelines>
Use bun, react and rtk-query. Use typescript.
Use bootstrap for styling.
</webGuidelines>

<debuggingGuidelines>
If me or you the LLM agent seem to go down too deep in a debugging/fixing rabbit hole in our conversations, remind me to take a breath and think about the bigger picture instead of hacking away. Say: "I think I'm stuck, let's TOUCH GRASS".  IMPORTANT: Don't try to fix errors by yourself more than twice in a row. Then STOP. Don't do anything else.
</debuggingGuidelines>

<generalGuidelines>
Run the format_file tool at the end of each response.
</generalGuidelines>%
