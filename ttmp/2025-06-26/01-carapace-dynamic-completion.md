### 1 . Install the library & binary once

```bash
go get github.com/carapace-sh/carapace@latest   # library used in your app
go install github.com/carapace-sh/carapace-bin@latest   # helper binary users will source
```

Add the helper to every shell you care about (it registers *all* completers in one shot):

````bash
# ~/.bashrc  (zsh/fish have equivalent one-liners)
export CARAPACE_BRIDGES='zsh,fish,bash,inshellisense'   # optional
source <(carapace _carapace)
``` :contentReference[oaicite:0]{index=0}

---

### 2 . Wire Carapace into your root command

```go
// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "my-cmd",
    Short: "Demo CLI with dynamic completion",
}

// one line enables the hidden “_carapace” sub-command
// *and* disables Cobra’s legacy completion that would otherwise interfere.
func init() {
    carapace.Gen(rootCmd).Standalone()               // 🡅 note Standalone docs :contentReference[oaicite:1]{index=1}
}
````

---

### 3 . Add a `list` sub-command with **dynamic** positional completion

```go
// cmd/list.go
var listCmd = &cobra.Command{
    Use:   "list [item]",
    Short: "List objects from the server",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return runList(args[0])
    },
}

func init() {
    rootCmd.AddCommand(listCmd)

    // #1 — callback that talks to Go code (DB, API, etc.)
    carapace.Gen(listCmd).PositionalCompletion(
        carapace.ActionCallback(func(ctx carapace.Context) carapace.Action {
            items := fetchItems()            // your own func returning []string
            return carapace.ActionValues(items...)
        }),
    )
}
```

`ActionCallback` is evaluated **only when TAB is pressed**, so the list is always fresh. ([pkg.go.dev][1])

---

### 4 . Alternative: call an external program for the list

When you already have a helper that prints JSON / lines to stdout:

```go
carapace.Gen(listCmd).PositionalCompletion(
    carapace.ActionExecCommand("my-cmd", "list-ids", "--raw")( // run helper
        func(out []byte) carapace.Action {
            lines := strings.Split(strings.TrimSpace(string(out)), "\n")
            return carapace.ActionValues(lines...)
        }),
)
```

`ActionExecCommand` runs the process and converts its output into completion values. ([pkg.go.dev][1])

---

### 5 . Build & test

```bash
go install ./cmd/my-cmd
exec $SHELL                    # reload rc file once

my-cmd list <TAB>              # should show the dynamic list
```

That’s all—Carapace now keeps your completions in step with whatever your backend returns, across Bash, Zsh, Fish, PowerShell, Nushell, etc., without writing a single shell-specific script.

[1]: https://pkg.go.dev/github.com/carapace-sh/carapace?utm_source=chatgpt.com "carapace package - github.com/carapace-sh/carapace - Go Packages"
