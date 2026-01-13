---
Title: Build a React + RTK Query app with Vite and a Dagger go:generate Builder
Slug: how-to-create-a-web-app-with-react-rtk-vite-dagger-gen
Short: Step-by-step guide to scaffold a React/RTK Query web app with Vite and build it using a Dagger-powered go:generate hook, served by a Go binary.
Topics:
- web
- react
- dagger
- go-generate
- build
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Build a React + RTK Query app with Vite and a Dagger go:generate Builder

## Overview

This guide shows how to create a small, production-friendly web frontend (React + RTK Query + Vite) and bundle it with a Go backend using a self-contained Dagger build. The Go backend exposes a `go:generate` hook to produce the static `dist/` assets without requiring Node to be installed locally. You will finish with a single Go command that serves the built SPA and a few API endpoints.

For writing style and structure, see:

```
glaze help how-to-write-good-documentation-pages
```

## Prerequisites

The approach assumes:

- Go 1.21+ (1.22+ recommended)
- Docker or another container runtime available to Dagger
- Dagger SDK for Go (`go get dagger.io/dagger@latest`)
- Optional local Node never required; the build runs in a container via Dagger

Basic familiarity with React and Go web servers is helpful.

## Project Layout

The following layout keeps frontend, builder, and server cohesive:

```
your-repo/
  cmd/
    app/
      main.go           # Go entrypoint (serves API + static)
      gen.go            # go:generate hook → runs ../build-web
      dist/             # output of the Vite build (generated)
    build-web/
      main.go           # Dagger builder for web/
  web/
    index.html
    vite.config.ts
    package.json
    tsconfig.json
    src/
      main.tsx
      store.ts
      api.ts            # RTK Query base slice
      views/
        Home.tsx
        Health.tsx
```

You can adapt names (e.g., `cmd/app`) to your project.

## Step 1 — Scaffold Vite + React

Create a minimal Vite React app in `web/`.

web/package.json:

```json
{
  "name": "my-web",
  "private": true,
  "version": "0.0.1",
  "type": "module",
  "packageManager": "pnpm@10.15.0",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview --port 5173"
  },
  "dependencies": {
    "@reduxjs/toolkit": "^2.2.3",
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-redux": "^9.0.0",
    "react-router-dom": "^6.22.3"
  },
  "devDependencies": {
    "@types/react": "^18.2.66",
    "@types/react-dom": "^18.2.22",
    "@vitejs/plugin-react": "^4.3.1",
    "typescript": "^5.5.3",
    "vite": "^5.4.0"
  }
}
```

web/vite.config.ts:

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: { outDir: 'dist', sourcemap: false },
  server: { port: 5173 }
})
```

web/index.html:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>My App</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

web/src/main.tsx:

```tsx
import React from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import { store } from './store'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Home } from './views/Home'
import { Health } from './views/Health'

const root = createRoot(document.getElementById('root')!)
root.render(
  <React.StrictMode>
    <Provider store={store}>
      <BrowserRouter>
        <div style={{ padding: 16 }}>
          <div style={{ float: 'right' }}><Health /></div>
          <Routes>
            <Route path="/" element={<Home />} />
          </Routes>
        </div>
      </BrowserRouter>
    </Provider>
  </React.StrictMode>
)
```

web/src/store.ts:

```ts
import { configureStore } from '@reduxjs/toolkit'
import { api } from './api'

export const store = configureStore({
  reducer: { [api.reducerPath]: api.reducer },
  middleware: (gDM) => gDM().concat(api.middleware)
})
```

web/src/api.ts:

```ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  endpoints: (b) => ({
    health: b.query<{ ok: boolean }, void>({ query: () => '/health' })
  })
})

export const { useHealthQuery } = api
```

web/src/views/Health.tsx:

```tsx
import React from 'react'
import { useHealthQuery } from '../api'

export const Health: React.FC = () => {
  const { data, isLoading, isError } = useHealthQuery()
  if (isLoading) return <span>…</span>
  if (isError) return <span style={{ color: 'red' }}>Server DOWN</span>
  return <span style={{ color: data?.ok ? 'green' : 'red' }}>{data?.ok ? 'Server OK' : 'Server DOWN'}</span>
}
```

web/src/views/Home.tsx:

```tsx
import React from 'react'

export const Home: React.FC = () => (
  <main>
    <h1>My App</h1>
    <p>Starter UI using React + RTK Query + Vite.</p>
  </main>
)
```

## Step 2 — Dagger Builder (Go)

Create a Go program to build `web/` inside a container and export the `dist/` output to your server directory. Place it at `cmd/build-web/main.go`.

```go
package main

import (
  "context"
  "fmt"
  "log"
  "os"
  "path/filepath"
  "strings"
  "dagger.io/dagger"
)

func main() {
  pnpmVersion := os.Getenv("WEB_PNPM_VERSION")
  if pnpmVersion == "" { pnpmVersion = "10.15.0" }

  ctx := context.Background()
  client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
  if err != nil { log.Fatalf("connect dagger: %v", err) }
  defer client.Close()

  // repo root assumed two levels up from here: cmd/build-web → repo/
  wd, _ := os.Getwd()
  repoRoot := filepath.Dir(filepath.Dir(wd))
  webPath := filepath.Join(repoRoot, "web")
  outPath := filepath.Join(filepath.Dir(wd), "app", "dist") // cmd/app/dist

  base := client.Container().From("node:22")
  if bi := os.Getenv("WEB_BUILDER_IMAGE"); bi != "" { base = client.Container().From(bi) }

  webDir := client.Host().Directory(webPath)
  ctr := base.
    WithWorkdir("/src").
    WithMountedDirectory("/src", webDir).
    WithEnvVariable("PNPM_HOME", "/pnpm")

  // Use Corepack to pin pnpm
  if os.Getenv("WEB_BUILDER_IMAGE") == "" || !strings.Contains(os.Getenv("WEB_BUILDER_IMAGE"), ":") {
    ctr = ctr.WithExec([]string{"sh", "-lc", fmt.Sprintf("corepack enable && corepack prepare pnpm@%s --activate", pnpmVersion)})
  }

  ctr = ctr.
    WithExec([]string{"sh", "-lc", "pnpm --version"}).
    WithExec([]string{"sh", "-lc", "pnpm install --reporter=append-only"}).
    WithExec([]string{"sh", "-lc", "pnpm build"})

  dist := ctr.Directory("/src/dist")
  if _, err := dist.Export(ctx, outPath); err != nil {
    log.Fatalf("export dist: %v", err)
  }
  log.Printf("exported web dist to %s", outPath)
}
```

Environment variables supported:

- `WEB_PNPM_VERSION` (default `10.15.0`)
- `WEB_BUILDER_IMAGE` (e.g., `node:22` or a pinned digest)
- `PNPM_CACHE_DIR` (optional: mount a host dir as pnpm store)
- `REGISTRY_USER`/`REGISTRY_TOKEN` (optional for authenticated registries)

## Step 3 — Hook the Builder with go:generate

Add `cmd/app/gen.go` to integrate the builder into your Go build flow.

```go
//go:generate go run ../build-web
package main
```

Running `go generate ./cmd/app` will execute the Dagger builder and write to `cmd/app/dist/`.

## Step 4 — Serve Static Files and a Health API (Go)

Add a minimal Go HTTP server in `cmd/app/main.go`. It serves the built SPA and exposes a health endpoint. You can integrate with any CLI framework (or Glazed/Cobra if you already use it).

```go
package main

import (
  "flag"
  "log"
  "net/http"
  "os"
  "path/filepath"
)

func main() {
  root := flag.String("root", "./dist", "path to built web assets")
  addr := flag.String("addr", ":8088", "listen address")
  flag.Parse()

  mux := http.NewServeMux()
  mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"ok":true}`))
  })

  abs, err := filepath.Abs(*root)
  if err != nil { log.Fatalf("resolve root: %v", err) }
  if _, err := os.Stat(abs); err != nil {
    log.Printf("warning: web dist not found at %s", abs)
  }
  mux.Handle("/", http.FileServer(http.Dir(abs)))

  log.Printf("serving on %s (web from %s)", *addr, abs)
  log.Fatal(http.ListenAndServe(*addr, mux))
}
```

## Step 5 — Build and Run Locally

- Build frontend assets via Dagger:

```bash
cd cmd/app
go generate
```

- Run the server:

```bash
go run . serve --addr :8088 --root ./dist
# or if using the simple flag-based server above:
go run . --addr :8088 --root ./dist
```

- Test:

```bash
curl -s localhost:8088/api/health | jq
open http://localhost:8088/
```

## Step 6 — Optional: Manage the Server with tmux

Using tmux lets you keep the server running while you iterate:

```bash
tmux kill-session -t web || true
tmux new-session -d -s web 'cd cmd/app && go run . --addr :8088 --root ./dist'
tmux attach -t web  # Ctrl-b d to detach
tmux kill-session -t web
```

## Step 7 — CI Considerations

- Run `go generate ./cmd/app` in CI to produce the `dist/` artifacts before building your binary.
- Cache Dagger layers and pnpm store for faster builds (e.g., map `PNPM_CACHE_DIR` to a CI cache path).
- Ship the `dist/` directory in release artifacts or bake it into your container image.

## Playbook — GoReleaser with Dagger UI Prebuild (Split Jobs)

This playbook covers the case where you run split GoReleaser jobs (for example, Linux + macOS) and you want Dagger to build the UI once on Linux, then pass the built assets to the other jobs. This avoids relying on Docker/Colima in macOS CI while still embedding the UI into every binary.

### Goal

- Run the Dagger UI build exactly once in CI.
- Reuse the generated assets for all GoReleaser split jobs.
- Keep the GoReleaser hooks free of `go generate` for UI packaging.

### Step 1 — Makefile: explicit UI build target

Expose the Dagger build as a Makefile target, and make your binary build depend on it. Prefer `go run` so the Dagger SDK is available without requiring a prebuilt builder binary.

```makefile
ui-build:
	GOWORK=off go run ./internal/web/generate_build.go

build: ui-build
	go build -tags "sqlite_fts5,embed" ./cmd/app
```

If you embed assets from a different path (for example `cmd/app/dist`), update your Dagger builder to export to that directory and ensure your `go:embed` directive points there.

### Step 2 — Keep the Dagger SDK pinned during `go mod tidy`

If your build uses `go mod tidy`, add a tools file to retain the Dagger SDK dependency:

```go
//go:build tools
// +build tools

package tools

import (
	_ "dagger.io/dagger"
)
```

Then run tidy with `GOFLAGS=-tags=tools` in CI:

```yaml
before:
  hooks:
    - sh -c 'GOFLAGS=-tags=tools go mod tidy'
```

### Step 3 — CI: prebuild UI assets and upload as an artifact

Add a Linux job that runs the Dagger UI build and uploads the output directory. The artifact path must match the directory your embed build expects.

```yaml
jobs:
  ui-prebuild:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v6
        with:
          go-version: '>=1.19.5'
          cache: true
      - name: Build UI assets
        run: |
          GOWORK=off make ui-build
      - uses: actions/upload-artifact@v4
        with:
          name: ui-embed
          path: internal/web/embed/public
```

### Step 4 — GoReleaser split jobs: download the artifact

Each split job (Linux + macOS) downloads the UI assets to the exact same path before running GoReleaser.

```yaml
jobs:
  goreleaser-linux:
    runs-on: ubuntu-latest
    needs: [ui-prebuild]
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: '>=1.19.5'
          cache: true
      - uses: actions/download-artifact@v4
        with:
          name: ui-embed
          path: internal/web/embed/public
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser-pro
          version: "~> v2"
          args: release --clean --split

  goreleaser-darwin:
    runs-on: macos-latest
    needs: [ui-prebuild]
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: '>=1.19.5'
          cache: true
      - uses: actions/download-artifact@v4
        with:
          name: ui-embed
          path: internal/web/embed/public
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser-pro
          version: "~> v2"
          args: release --clean --split
```

### Step 5 — Remove UI go:generate from GoReleaser hooks

Once the prebuild job is in place, remove any `go generate ./internal/web` hook from `.goreleaser.yaml`. The generated assets are already present via the artifact.

### Validation checklist

- The prebuild job output directory matches your `go:embed` path.
- The artifact download path matches the embed path exactly.
- The GoReleaser hooks no longer invoke `go generate` for UI packaging.
- `make ui-build` succeeds on Linux and uses Dagger end-to-end.

## Troubleshooting

- Dagger cannot connect to engine:
  - Ensure Docker is running, or configure an alternate runtime.
  - If running in a restricted environment, request elevated permissions for network/socket ops.
- pnpm/Corepack issues:
  - The builder uses Corepack to pin pnpm. Override with `WEB_PNPM_VERSION` or pre-bake pnpm in a custom `WEB_BUILDER_IMAGE`.
- Static files 404:
  - Confirm `cmd/app/dist/` exists and contains `index.html` and an `assets/` folder.
  - Check you’re passing `--root ./dist` to the server.

## Next Steps

- Expand your API surface and wire RTK Query endpoints in `web/src/api.ts`.
- Add state slices to `store.ts` for UI features.
- Introduce environment-specific base URLs if you split frontend/backend origins in production.
