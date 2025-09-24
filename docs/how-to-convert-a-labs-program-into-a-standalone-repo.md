---
Title: How to Convert a Labs Program into a Standalone Repo
Slug: how-to-convert-a-labs-program-into-a-standalone-repo
Short: Step‑by‑step guide to extract an app from go-go-labs into its own repo/module with working build, lint, and release.
Topics:
- labs
- packaging
- modules
- tooling
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# How to Convert a Labs Program into a Standalone Repo

This tutorial walks you through extracting a single application from the `go-go-labs` monorepo into its own clean repository/module, complete with updated imports, documentation, examples, linting, and release automation.

We use the recent extraction of the zine layout tool (formerly `cmd/apps/zine-layouter`) as a concrete example. The same steps apply to any other app in the monorepo.

For documentation style and structure guidance, see:

```
glaze help how-to-write-good-documentation-pages
```

## Prerequisites

This guide assumes:
- You have the go-go-labs repository checked out locally.
- You have a target folder for the new project (e.g., `./zine-layout/`).
- You can edit `go.work` at the repository root.
- You have a recent Go toolchain installed (matching `go.work`/`toolchain`).

Optional but helpful:
- `golangci-lint` for linting
- `goreleaser` for snapshot releases

## 1) Identify What to Extract

Goal: Find the app entrypoint and all related packages/files to move.

- App entrypoint: `go-go-labs/cmd/apps/<app-name>/main.go`
- Related library packages are typically under `go-go-labs/pkg/<domain>/...`
- Example/spec files, tests, and DSL docs may live alongside the app (e.g., `layouts/`, `tests/`, or `*-dsl.md`).

Example (zine layouter):
- Entrypoint: `go-go-labs/cmd/apps/zine-layouter/main.go`
- Library: `go-go-labs/pkg/zinelayout/`
- Examples & DSL: `go-go-labs/cmd/apps/zine-layouter/{layouts,tests,zine-layout-dsl.md}`

## 2) Create the Target Module Skeleton

Goal: Prepare a home for the extracted code with a standard layout.

- Create the new project folder: `./zine-layout/`
- Add typical structure:
  - `cmd/<binary-name>/` for the CLI entrypoint
  - `pkg/` for library code
  - `doc/` and `examples/` for documentation and specs
  - Copy template files if available (.gitignore, .goreleaser.yaml, .golangci.yml, Makefile)

Example layout:

```
zine-layout/
  cmd/
    zine-layout/
      main.go
  pkg/
    zinelayout/
      ...
  examples/
    layouts/
    tests/
  doc/
    dsl.md
  .goreleaser.yaml
  .golangci.yml
  Makefile
  README.md
```

## 3) Move Code Using mv

Goal: Physically move the app and its library code into the new module while preserving history in your working tree.

Commands:

```bash
mkdir -p zine-layout/cmd/zine-layout
mv go-go-labs/cmd/apps/zine-layouter/main.go zine-layout/cmd/zine-layout/main.go

mkdir -p zine-layout/pkg
mv go-go-labs/pkg/zinelayout zine-layout/pkg/zinelayout
```

Copy related examples and docs:

```bash
mkdir -p zine-layout/examples/layouts zine-layout/examples/tests zine-layout/doc
cp go-go-labs/cmd/apps/zine-layouter/layouts/*.yaml zine-layout/examples/layouts/
cp go-go-labs/cmd/apps/zine-layouter/tests/*.yaml zine-layout/examples/tests/
cp go-go-labs/cmd/apps/zine-layouter/zine-layout-dsl.md zine-layout/doc/dsl.md
```

Tip: Start by moving only what you need to build the CLI and its immediate dependencies. You can copy additional material later.

## 4) Create go.mod and Add the Module to go.work

Goal: Turn the new folder into a proper module and let the workspace see it.

In `zine-layout/go.mod`:

```go
module github.com/go-go-golems/zine-layout

go 1.24.3
```

Then add the module to the workspace in `go.work` at the repo root:

```diff
 use (
   ./glazed
   ./go-go-labs
+  ./zine-layout
 )
```

Run `go mod tidy` inside the new module:

```bash
cd zine-layout
go mod tidy
```

## 5) Update Imports and Package Paths

Goal: Point all imports from the old monorepo paths to the new module path.

- In the CLI entrypoint and library code, replace:
  - `github.com/go-go-golems/go-go-labs/pkg/...` → `github.com/go-go-golems/zine-layout/pkg/...`

Examples to search and fix:

```bash
rg -n "go-go-labs/pkg/zinelayout" zine-layout
```

Update imports accordingly in both the CLI and library packages. Don’t forget any docs with inline code snippets (e.g., `units_doc.md`).

## 6) Replace Template Placeholders (XXX)

Goal: Clean up any template boilerplate carried into the new module (names, paths, binaries).

Common files to check:
- `Makefile` (release target proxy line, install target binary name)
- `.goreleaser.yaml` (project_name, main path, binary name, homepage)
- `AGENT.md` or similar developer guide references
- `.github` workflows if present

Search and replace:

```bash
rg -n --hidden "\bXXX\b" zine-layout
```

Update to use the new binary name (e.g., `zine-layout`) and proper GitHub repo path (`github.com/go-go-golems/zine-layout`).

## 7) Write or Update README and Docs

Goal: Provide a complete README for the standalone project and include any DSL or examples documentation.

Include:
- Overview and feature list
- Install and quick-start commands
- CLI flags (brief table or bullets)
- A minimal spec example
- Pointers to `examples/` and `doc/` (DSL, units, etc.)

Follow the documentation style guide to keep the structure clear, focused, and scannable:

```
glaze help how-to-write-good-documentation-pages
```

## 8) Build and Tidy

Goal: Ensure the new module compiles cleanly under the workspace.

```bash
cd zine-layout
go mod tidy
go build ./...
```

If the build fails:
- Re-check imported module paths.
- Ensure `go.work` includes `./zine-layout`.
- Run `rg` to find any lingering old imports.

## 9) Lint and Fix Common Issues

Goal: Run `golangci-lint` and address formatting and static analysis issues.

Install locally (avoids system-wide changes):

```bash
GOBIN=$(pwd)/.bin GO111MODULE=on go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
PATH=$(pwd)/.bin:$PATH golangci-lint run -v
```

Typical fixes you may encounter:
- gofmt: run `go fmt ./...`
- printf format: e.g., printing a float with `%.0f` instead of `%d`
- predeclared identifiers: rename helpers like `max` → `intMax`
- exhaustive switches: use tagged switches over enums (e.g., YAML node kinds)
- unused constants/vars: remove or use

Repeat format+lint until clean:

```bash
go fmt ./...
PATH=$(pwd)/.bin:$PATH golangci-lint run -v
```

## 10) Validate the Release Pipeline (Snapshot)

Goal: Ensure `.goreleaser.yaml` works and produces artifacts.

Install and run snapshot release:

```bash
GOBIN=$(pwd)/.bin GO111MODULE=on go install github.com/goreleaser/goreleaser/v2@latest
PATH=$(pwd)/.bin:$PATH goreleaser release --skip=sign --snapshot --clean
ls -la dist/
```

You should see:
- Archived binaries (tar.gz) for configured platforms
- `.deb` and `.rpm` packages if enabled
- `checksums.txt`, Homebrew formula under `dist/homebrew/`

If it fails:
- Check paths: `main: ./cmd/<binary>`, `binary: <binary-name>`
- Update `project_name`, `homepage`, and descriptions
- Remove or update deprecated sections as needed

## 11) Bring Over Examples and DSL Docs

Goal: Package everything developers need to understand and test the tool.

- Copy example specs (`examples/layouts/`, `examples/tests/`).
- Copy DSL or format guides into `doc/` (e.g., `dsl.md`).
- Link to units and expression docs if they live under `pkg/.../parser/`.

Verify the CLI with an example:

```bash
./dist/zine-layout --spec examples/layouts/two_pages_two_inputs.yaml \
  --output-dir out/ img1.png img2.png
```

## 12) Final Checks and Polish

Before you publish:
- Confirm `README.md` is accurate and complete.
- Ensure `go.mod` module path matches the intended GitHub repo.
- Search for any remaining old import paths or `XXX` placeholders.
- Consider adding a `doc/README.md` that indexes docs and examples.

## Troubleshooting

- Build reports: "directory prefix . does not contain modules listed in go.work"
  - Ensure `./zine-layout` is added to `go.work`.
- Lint fails on YAML kind switches
  - Use a tagged `switch` and handle all needed cases; ensure function returns a value on all paths.
- GoReleaser fails with missing tags
  - Use `--snapshot` for local validation; real releases need tags.
- Broken imports after move
  - `rg -n "go-go-labs" zine-layout` to find stale paths and fix.

## Recap of the Concrete Steps Used

1. Create new project structure under `zine-layout/`.
2. `mv` the CLI entrypoint and related library packages.
3. Copy examples and DSL docs into `examples/` and `doc/`.
4. Add `zine-layout/go.mod` and reference it from `go.work`.
5. Update imports from `go-go-labs/...` → `zine-layout/...`.
6. Replace `XXX` placeholders in Makefile, `.goreleaser.yaml`, `AGENT.md`.
7. Write a full `README.md` with usage, flags, and example.
8. `go mod tidy` and `go build ./...`.
9. Install and run `golangci-lint`; fix issues; `go fmt` as needed.
10. Run `goreleaser release --skip=sign --snapshot --clean` and verify `dist/`.
11. Sanity-run the CLI with an example spec.

By following these steps, you can reliably extract any app from the monorepo and ship it as a tidy, standalone project with a working build, lint, and release pipeline.

