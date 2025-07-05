### Bubble Tea Development Playbook

#### 1. Project layout & file hygiene 📂

| File/Dir          | Responsibility                                                                                  | Tips                                                                       |
| ----------------- | ----------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `cmd/app/main.go` | Wire up `tea.NewProgram`, parse CLI flags, and launch your root model                           | Keep **main.go** under \~40 lines; everything else should live in packages |
| `pkg/ui/model/`   | Pure data structures and the `Init/Update/View` trio for each screen                            | One file per screen or domain object; avoid circular imports               |
| `pkg/ui/view/`    | Re-usable Lipgloss styles and helper functions (`renderHeader`, `renderFooter`, …)              | Centralised styles makes theming and tests easier                          |
| `pkg/ui/keys/`    | `key.Binding` definitions plus **ShortHelp/FullHelp** implementations                           | Makes key bindings discoverable and avoids scattering strings              |
| `pkg/ui/bubbles/` | Thin wrappers around standard Bubbles (list, table, textarea, …) when you need custom behaviour | Wrap, don’t fork                                                           |

> **Why split?** Fast compilation, simpler code reviews, and the freedom to unit-test models without spinning up a full TUI. Splitting also keeps **Update** and **View** small, which preserves UI responsiveness. ([leg100.github.io][1], [zackproser.com][2])

---

#### 2. Styling with Lip Gloss 🎨

*Lipgloss* is Bubble Tea’s sibling library for ANSI styling: borders, padding, alignment and colours. Create a central `var Styles struct{ … }` so every component shares the same palette and border language.

```go
var Styles = struct {
    Title, Pane, Selected lipgloss.Style
}{
    Title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")),
    Pane:    lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
    Selected: lipgloss.NewStyle().
        Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("#5af")),
}
```

Lipgloss’ border helpers (`NormalBorder`, `RoundedBorder`, custom overrides) keep ASCII art out of your code and are terminal-safe. ([github.com][3], [github.com][4])

---

#### 3. Debugging workflow with **tmux** + **vhs** 🛠️

| Tool     | Use-case                                                                                                                                                                                                                                    |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **tmux** | Run the app and interact/capture its output.                                                                                                                                                                                                |
| **vhs**  | Record deterministic “screenshots” (a Markdown-like script → animated terminal demo) to document bugs or reproduce a user flow in CI.<br><br>The installed version is able to take Screenshot XXX.txt or Screenshot XXX.ansi , not just png |
|          |                                                                                                                                                                                                                                             |

**Tip:** Keep a `scripts/demo.tape` in the repo; interns can run `vhs < tape` to watch or re-generate the demo GIF.

---

#### 4. Embrace the Bubble Tea architecture 🫖

Bubble Tea centres on a **model → Update() → View()** triplet:

```go
type model struct { …

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // mutate a copy of m, return next Cmd
    case errMsg:
        m.err = msg
    }
    return m, nil
}

func (m model) View() string { return renderUI(m) }
```

* Never mutate the model from a goroutine; instead return a `tea.Cmd` and handle its message in `Update`.
* Routinely handle `tea.WindowSizeMsg` at the root model and propagate it to child components so everyone can re-flow. ([github.com][5], [leg100.github.io][1])

---

#### 5. Prefer official **Bubbles** before rolling your own 🧩

The `github.com/charmbracelet/bubbles` repo ships high-quality widgets: `list`, `table`, `textinput`, `textarea`, `viewports`, `progress`, `spinner`, `paginator`, `stopwatch`, etc. Start there, embed or wrap them, and customise via their public `Styles` fields instead of copying code. It buys you free features like mouse support and built-in help. ([github.com][6])

---

#### 6. Unified key handling with **keymap + help** ⌨️

Create a dedicated `keymap` struct:

```go
type KeyMap struct {
    Quit   key.Binding
    Up     key.Binding
    Down   key.Binding
    Toggle key.Binding
}

func NewKeyMap() KeyMap {
    return KeyMap{
        Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
        Up:   key.NewBinding(key.WithKeys("k", "up"),   key.WithHelp("↑/k", "up")),
        Down: key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "down")),
        Toggle: key.NewBinding(key.WithKeys("space"),   key.WithHelp("␣", "select")),
    }
}
```

Satisfy the **bubbles/help.KeyMap** interface:

```go
func (k KeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Up, k.Down, k.Quit} }
func (k KeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{{k.Up, k.Down}, {k.Toggle, k.Quit}}
}
```

Then drop a `help.Model` in your root model and draw it inside `View()`. This approach gives you automatic **short** (one-liner) and **long** (two-row) help displays. ([github.com][7], [github.com][8], [pkg.go.dev][9], [alexho.dev][10])

Make sure keyboard KeyMsg are forwrad to the right model.

---

#### 7. Performance & polish checklist ✅

1. **Keep Update/View small**; heavy work goes in goroutines that return a `tea.Cmd`. ([leg100.github.io][1])
2. **Batch redraws**: when several messages arrive, process them first and call `tea.Batch(cmds…)` to avoid flicker.
3. **Respect resize events**: always brace for narrow terminals (`if w ≤ 40 { … }`).
4. **Unit-test models**: feed synthetic `tea.KeyMsg` and assert the next model.

---

### TL;DR for the intern

1. **Organise**: one concept per file, one folder per layer.
2. **Style**: define reusable Lipgloss styles up-front.
3. **Debug & demo**: live in tmux, record with vhs, validate screenshots.
4. **Think in M → U → V**: all state lives in the model, all IO via Cmds.
5. **Reuse Bubbles**: they ship tables, lists, progress bars, etc.
6. **Keymaps + help**: consistent bindings and auto-generated short/long help.


