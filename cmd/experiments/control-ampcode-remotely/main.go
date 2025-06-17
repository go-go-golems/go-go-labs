package main

import (
    "bufio"
    "context"
    "embed"
    "encoding/json"
    "fmt"
    "io/fs"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/creack/pty"
    "github.com/gorilla/websocket"
    "github.com/pkg/errors"
    "golang.org/x/sync/errgroup"
    "github.com/spf13/cobra"
)

//go:embed static/*
var staticFiles embed.FS

var verbose bool
var useTmux bool
var useStrace bool

// AmpState represents a high level state of the amp agent.
type AmpState string

const (
    StateAsking          AmpState = "asking"
    StateThinking        AmpState = "thinking"
    StateRunningCommand  AmpState = "running_command"
    StateOutput          AmpState = "output"
    StateIdle            AmpState = "idle"
    StateQuitting        AmpState = "quitting"
)

// Event sent to the frontend.
type Event struct {
    State AmpState `json:"state"`
    Line  string   `json:"line,omitempty"`
    TS    int64    `json:"ts"`
    From  string   `json:"from,omitempty"` // "amp" or "client"
}

// Regex patterns (compiled once)
var (
    stripAnsi      = regexp.MustCompile(`\x1b\[[0-9;?]*[0-9;]*[A-Za-z]`)
    reAsking       = regexp.MustCompile(`^>\s+.+`)
    reThinking     = regexp.MustCompile(`(◉|◎)\s+(Thinking|Preparing Task)\.\.\.`)
    reRunning      = regexp.MustCompile(`(◉|◎)\s+Running\s+(command|tool|commands)\.\.\.`)
    reOutputBlock  = regexp.MustCompile(`^╭.*`) // start of output block
    reIdle         = regexp.MustCompile(`^>\s*$`)
    reQuitting     = regexp.MustCompile(`^Shutting down\.\.\.$`)
)

// Hub maintains websocket clients and broadcasts events.
type Hub struct {
    mu    sync.Mutex
    conns map[*websocket.Conn]struct{}
}

func newHub() *Hub {
    return &Hub{conns: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) add(conn *websocket.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.conns[conn] = struct{}{}
}

func (h *Hub) remove(conn *websocket.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()
    delete(h.conns, conn)
}

func (h *Hub) broadcast(ev Event) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if buf, err := json.Marshal(ev); err == nil {
        log.Printf("WS send: %s\n", string(buf))
    }
    for c := range h.conns {
        if err := c.WriteJSON(ev); err != nil {
            c.Close()
            delete(h.conns, c)
        }
    }
}

func main() {
    rootCmd := &cobra.Command{
        Use:   "control-amp",
        Short: "Run amp in a PTY and expose a web UI for control",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx, cancel := context.WithCancel(cmd.Context())
            defer cancel()
            return run(ctx)
        },
    }
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print raw amp output to stdout")
    rootCmd.PersistentFlags().BoolVar(&useTmux, "tmux", false, "launch amp inside a tmux session instead of a local PTY")
    rootCmd.PersistentFlags().BoolVar(&useStrace, "strace", false, "wrap amp in strace (-ff -e read,write)")
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run(ctx context.Context) error {
    var reader io.ReadCloser
    var sendInput func(string) error
    var ptyFile *os.File

    log.Println("Locating 'amp' executable...")
    ampPath, errLook := exec.LookPath("amp")
    if errLook != nil {
        return errors.Wrap(errLook, "cannot find 'amp' in PATH")
    }

    hub := newHub()

    if useTmux {
        log.Println("tmux mode enabled - launching amp inside tmux session")
        sess := fmt.Sprintf("ampcode-%d", time.Now().UnixNano())
        fifo := fmt.Sprintf("/tmp/%s.fifo", sess)

        // create fifo
        if err := exec.Command("mkfifo", fifo).Run(); err != nil {
            return errors.Wrap(err, "creating fifo")
        }

        // build command to run inside tmux (with optional strace)
        tmuxCmd := ampPath
        if useStrace {
            straceBin, err := exec.LookPath("strace")
            if err != nil {
                return errors.Wrap(err, "strace requested but not found")
            }
            tracePrefix := fmt.Sprintf("/tmp/amp-strace-%d", time.Now().UnixNano())
            log.Printf("strace enabled inside tmux, prefix: %s\n", tracePrefix)
            tmuxCmd = fmt.Sprintf("%s -ff -tt -s 256 -e read,write -o %s %s", straceBin, tracePrefix, ampPath)
        }

        // start tmux session detached running command
        if err := exec.Command("tmux", "new-session", "-d", "-s", sess, tmuxCmd).Run(); err != nil {
            return errors.Wrap(err, "starting tmux session")
        }

        // pipe pane output into fifo
        if err := exec.Command("tmux", "pipe-pane", "-t", sess+":0.0", "-o", "stdbuf -o0 cat -u > "+fifo).Run(); err != nil {
            return errors.Wrap(err, "pipe-pane setup")
        }

        // open fifo for reading
        f, err := os.Open(fifo)
        if err != nil {
            return errors.Wrap(err, "opening fifo")
        }
        reader = f

        sendInput = func(cmd string) error {
            argsLit := []string{"send-keys", "-t", sess, "-l", cmd}
            log.Printf("tmux> tmux %s\n", strings.Join(argsLit, " "))
            if err := exec.Command("tmux", argsLit...).Run(); err != nil {
                return err
            }
            argsCR := []string{"send-keys", "-t", sess, "C-m"}
            log.Printf("tmux> tmux %s\n", strings.Join(argsCR, " "))
            return exec.Command("tmux", argsCR...).Run()
        }

        // Ensure session killed on exit
        go func() {
            <-ctx.Done()
            _ = exec.Command("tmux", "kill-session", "-t", sess).Run()
            _ = os.Remove(fifo)
        }()

        log.Printf("AMP running in tmux session '%s'. Attach with: tmux attach -t %s\n", sess, sess)
    } else {
        log.Printf("Starting amp process (%s)...\n", ampPath)
        var cmdArgs []string
        if useStrace {
            straceBin, err := exec.LookPath("strace")
            if err != nil {
                return errors.Wrap(err, "strace requested but not found")
            }
            tracePrefix := fmt.Sprintf("/tmp/amp-strace-%d", time.Now().UnixNano())
            log.Printf("strace enabled, output prefix: %s\n", tracePrefix)
            cmdArgs = append(cmdArgs, "-ff", "-tt", "-s", "256", "-e", "read,write", "-o", tracePrefix, ampPath)
            ampCmd := exec.CommandContext(ctx, straceBin, cmdArgs...)
            file, err := pty.Start(ampCmd)
            if err != nil {
                return errors.Wrap(err, "starting amp in pty")
            }
            ptyFile = file
            _ = pty.Setsize(ptyFile, &pty.Winsize{Cols: 120, Rows: 40})
            log.Println("AMP started within PTY; resizing to 120x40")
        } else {
            ampCmd := exec.CommandContext(ctx, ampPath)
            file, err := pty.Start(ampCmd)
            if err != nil {
                return errors.Wrap(err, "starting amp in pty")
            }
            ptyFile = file
            _ = pty.Setsize(ptyFile, &pty.Winsize{Cols: 120, Rows: 40})
            log.Println("AMP started within PTY; resizing to 120x40")
        }

        reader = ptyFile

        sendInput = func(cmd string) error {
            log.Printf("pty> write %q (no CR)\n", cmd)
            if _, err := ptyFile.Write([]byte(cmd)); err != nil {
                return err
            }
            // short pause so slave sees data before CR
            time.Sleep(30 * time.Millisecond)
            log.Printf("pty> write CR\\r separately\n")
            _, err := ptyFile.Write([]byte{'\r'})
            return err
        }
    }

    eg, ctx := errgroup.WithContext(ctx)

    // Reader goroutine: parse amp output and broadcast state
    eg.Go(func() error {
        scanner := bufio.NewScanner(reader)
        scanner.Split(crlfSplitFunc)
        var lastState AmpState
        for scanner.Scan() {
            rawLine := strings.Trim(scanner.Text(), "\r")
            if verbose {
                log.Printf("amp> %s %s\n", time.Now().Format("15:04:05.000"), strconv.Quote(rawLine))
            }
            line := stripAnsi.ReplaceAllString(rawLine, "")
            state := detectState(line)
            // If still in thinking mode and we encounter a non-empty line that isn't classified,
            // treat it as the beginning of output. This allows us to recognise when the agent
            // has finished "◉ Thinking…" and started replying (e.g. the "Hello!" line).
            if state == "" && strings.TrimSpace(line) != "" && lastState == StateThinking {
                state = StateOutput
            }
            if state != "" && state != lastState {
                lastState = state
                hub.broadcast(Event{State: state, Line: line, TS: time.Now().UnixMilli(), From: "amp"})
                log.Printf("State change: %s -- %s\n", state, line)
            }
            // Always broadcast raw output for interested clients
            if state == StateOutput {
                hub.broadcast(Event{State: StateOutput, Line: line, TS: time.Now().UnixMilli(), From: "amp"})
            }
        }
        if scanner.Err() != nil {
            return errors.Wrap(scanner.Err(), "scanner error")
        }
        return nil
    })

    // HTTP server
    server := &http.Server{Addr: ":8080"}
    log.Println("HTTP server listening on http://localhost:8080")

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            _ = conn.Close()
            return
        }
        hub.add(conn)
        // Initial state message
        hub.broadcast(Event{State: StateIdle, TS: time.Now().UnixMilli()})

        go func() {
            defer hub.remove(conn)
            for {
                _, msg, err := conn.ReadMessage()
                if err != nil {
                    return
                }
                log.Printf("WS recv: %s\n", string(msg))
                // Expect JSON {"type":"input","data":"..."}
                var payload struct {
                    Type string `json:"type"`
                    Data string `json:"data"`
                }
                if err := json.Unmarshal(msg, &payload); err != nil {
                    continue
                }
                if payload.Type == "input" {
                    input := strings.TrimSpace(payload.Data)
                    if input != "" {
                        _ = sendInput(input)
                        // broadcast what user typed back to everyone (including self)
                        hub.broadcast(Event{State: StateAsking, Line: input, TS: time.Now().UnixMilli(), From: "client"})
                    }
                }
            }
        }()
    })

    // Serve embedded static files. We want "/" to map to "static/index.html".
    subFS, err := fs.Sub(staticFiles, "static")
    if err != nil {
        return errors.Wrap(err, "sub fs")
    }
    http.Handle("/", http.FileServer(http.FS(subFS)))

    eg.Go(func() error {
        <-ctx.Done()
        // shutdown http server gracefully
        ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        return server.Shutdown(ctxTimeout)
    })

    eg.Go(func() error {
        if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            return errors.Wrap(err, "http server")
        }
        return nil
    })

    // Wait for all goroutines to finish
    return eg.Wait()
}

func detectState(line string) AmpState {
    switch {
    case reQuitting.MatchString(line):
        return StateQuitting
    case reRunning.MatchString(line):
        return StateRunningCommand
    case reThinking.MatchString(line):
        return StateThinking
    case reAsking.MatchString(line):
        return StateAsking
    case reOutputBlock.MatchString(line) || strings.HasPrefix(line, "│ "):
        return StateOutput
    case reIdle.MatchString(line):
        return StateIdle
    default:
        return ""
    }
}

// crlfSplitFunc splits on either \n or \r to capture lines that use carriage returns for animations
func crlfSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
    if atEOF && len(data) == 0 {
        return 0, nil, nil
    }
    for i, b := range data {
        if b == '\n' || b == '\r' {
            return i + 1, data[:i], nil
        }
    }
    if atEOF {
        return len(data), data, nil
    }
    return 0, nil, nil
} 