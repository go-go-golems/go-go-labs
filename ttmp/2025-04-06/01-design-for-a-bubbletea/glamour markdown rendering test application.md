# Design for Bubble Tea Glamour Markdown Rendering Test Application

**Goal:** Create a Bubble Tea application to interactively test the `glamour` library's markdown rendering capabilities, especially focusing on how it handles incomplete or partial markdown input. The application will feature a text input area at the bottom and a view panel above displaying the `glamour`-rendered output of the text input in real-time.

**Location:** `cmd/apps/bubbletea-markdown-test/main.go`
**Package:** `github.com/go-go-golems/go-go-labs/cmd/apps/bubbletea-markdown-test`

**Dependencies:**
- `github.com/charmbracelet/bubbles/textarea`
- `github.com/charmbracelet/bubbles/viewport`
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/glamour`
- `github.com/charmbracelet/lipgloss`

## Plan

1. **Define `main.go` Structure:**
   - Set up the `main` package.
   - Import necessary Charmbracelet libraries (`bubbletea`, `textarea`, `viewport`, `glamour`, `lipgloss`).

   ```go
   package main

   import (
       "fmt"
       "log"

       "github.com/charmbracelet/bubbles/textarea"
       "github.com/charmbracelet/bubbles/viewport"
       tea "github.com/charmbracelet/bubbletea"
       "github.com/charmbracelet/glamour"
       "github.com/charmbracelet/lipgloss"
   )

   func main() {
       // ... main function setup ...
   }

   // ... model definition ...
   // ... Init, Update, View methods ...
   ```

2. **Define the `model` Struct:**
   - This struct will hold the application's state.
   - `viewport viewport.Model`: To display the rendered markdown. Needs scrolling capability.
   - `textarea textarea.Model`: For user input.
   - `renderer *glamour.TermRenderer`: The glamour renderer instance.
   - `width int`, `height int`: To store terminal dimensions.
   - `err error`: To store any rendering errors.

   ```go
   type model struct {
       viewport    viewport.Model
       textarea    textarea.Model
       renderer    *glamour.TermRenderer
       width       int
       height      int
       err         error
   }
   ```

3. **Implement `Init()` Method:**
   - This method initializes the model's state when the application starts.
   - Initialize the `textarea` with placeholder text and focus.
   - Initialize the `viewport`.
   - Create a `glamour` renderer (e.g., with `glamour.WithAutoStyle()`).
   - Perform an initial render of the textarea's default value.
   - Return `textarea.Blink`.

   ```go
   func initialModel() model {
       ta := textarea.New()
       ta.Placeholder = "Enter markdown here..."
       ta.Focus()
       // ... set initial dimensions ...

       vp := viewport.New( /* initial width, height */ )
       // ... configure viewport options (e.g., style) ...

       renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle()) // Handle error appropriately

       m := model{
           textarea: ta,
           viewport: vp,
           renderer: renderer,
       }

       // Initial render
       renderedContent, err := m.renderer.Render(m.textarea.Value())
       if err != nil {
           m.err = err
           m.viewport.SetContent("Error rendering markdown.")
       } else {
            m.viewport.SetContent(renderedContent)
       }

       return m
   }

   func (m model) Init() tea.Cmd {
       return textarea.Blink // Start the cursor blinking
   }
   ```

4. **Implement `Update(msg tea.Msg)` Method:**
   - This method handles incoming messages (key presses, window size changes, etc.).
   - Use a `switch` statement on `msg.(type)`.
   - **`tea.KeyMsg`**:
       - Handle exit keys (Ctrl+C, Esc).
       - Forward other key messages to the `textarea` and `viewport` for their internal updates.
       - After updating the `textarea`, re-render its content using `glamour`.
       - Update the `viewport`'s content with the newly rendered markdown or an error message.
       - Store any rendering errors in `m.err`.
   - **`tea.WindowSizeMsg`**:
       - Update `m.width` and `m.height`.
       - Adjust the height/width of the `textarea` and `viewport`.
       - Update `viewport.Width` and `viewport.Height`.
       - Update `textarea.SetWidth` and `textarea.SetHeight`.
   - Return the updated model and any resulting commands (`tea.Cmd`).

   ```go
   func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       var (
           cmd  tea.Cmd
           cmds []tea.Cmd
       )

       switch msg := msg.(type) {
       case tea.KeyMsg:
           switch msg.String() {
           case "ctrl+c", "esc":
               return m, tea.Quit
           default:
               m.textarea, cmd = m.textarea.Update(msg)
               cmds = append(cmds, cmd)
               
               // Re-render on textarea change
               renderedContent, err := m.renderer.Render(m.textarea.Value())
               if err != nil {
                   m.err = err
                   m.viewport.SetContent(fmt.Sprintf("Render Error:\n%s\n\n%s", err.Error(), m.textarea.Value()))
               } else {
                   m.err = nil
                   m.viewport.SetContent(renderedContent)
               }
           }

       case tea.WindowSizeMsg:
           m.width = msg.Width
           m.height = msg.Height
           
           // Example: 3 lines for textarea, 1 for separator, rest for viewport
           textAreaHeight := 3
           viewportHeight := m.height - textAreaHeight - 1 // Adjust for potential borders/separators

           m.viewport.Width = m.width
           m.viewport.Height = viewportHeight
           
           // Re-render with new dimensions
           renderedContent, err := m.renderer.Render(m.textarea.Value())
           if err != nil {
               m.err = err
               m.viewport.SetContent(fmt.Sprintf("Resize Render Error:\n%s\n\n%s", err.Error(), m.textarea.Value()))
           } else {
               m.err = nil
               m.viewport.SetContent(renderedContent)
           }

           m.textarea.SetWidth(m.width)
           // Textarea height is managed internally by lines, but we set the visual box height
           m.textarea.SetHeight(textAreaHeight)
       }

       // Also update viewport for scrolling
       m.viewport, cmd = m.viewport.Update(msg)
       cmds = append(cmds, cmd)

       return m, tea.Batch(cmds...)
   }
   ```

5. **Implement `View()` Method:**
   - This method generates the string representation of the UI.
   - Use `lipgloss.JoinVertical` to stack the `viewport` and `textarea`.
   - Add styles (borders, margins) using `lipgloss`.
   - Potentially add a separator or status line between the viewport and textarea.
   - Display the rendering error if `m.err != nil`, perhaps in a footer or overlay.

   ```go
   func (m model) View() string {
       // Define styles maybe outside the method or in the model
       textAreaStyle := lipgloss.NewStyle().Height(3) // Example fixed height

       // Render error display
       errorLine := ""
       if m.err != nil {
            errorLine = fmt.Sprintf("\nError: %v", m.err) // Simple error line
       }

       return lipgloss.JoinVertical(lipgloss.Left,
           m.viewport.View(),
           textAreaStyle.Render(m.textarea.View()), // Apply style/height constraints
           errorLine, // Add error line at the bottom
       )
   }
   ```

6. **Implement `main()` Function:**
   - Create the initial model using `initialModel()`.
   - Create a new Bubble Tea program: `p := tea.NewProgram(initialModel())`.
   - Run the program: `if err := p.Start(); err != nil { ... }`.
   - Handle potential errors from `p.Start()`.
   - Use `log.Fatal` for fatal errors during setup or execution.

   ```go
   func main() {
       p := tea.NewProgram(initialModel(), tea.WithAltScreen()) // Use AltScreen
       if err := p.Start(); err != nil {
           log.Fatalf("Alas, there's been an error: %v", err)
       }
   }
   ```