package main

import (
    "context"
    "fmt"
    "io"
    "os"
    "time"

    "github.com/pkg/errors"
    "github.com/spf13/cobra"
)

var (
    dbPath   string
    yamlFile string
    verbose  bool
)

func main() {
    root := &cobra.Command{
        Use:   "doodle-clone",
        Short: "Apply YAML scheduling actions to a local SQLite doodle-like store",
    }

    root.PersistentFlags().StringVar(&dbPath, "db", "./doodle.db", "Path to sqlite database file")
    root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

    applyCmd := &cobra.Command{
        Use:   "apply",
        Short: "Apply a YAML document with scheduling actions",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runApply(cmd.Context())
        },
    }
    applyCmd.Flags().StringVarP(&yamlFile, "file", "f", "", "YAML file path (use '-' for stdin)")
    _ = applyCmd.MarkFlagRequired("file")

    root.AddCommand(applyCmd)

    if err := root.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func runApply(ctx context.Context) error {
    store, err := OpenStore(ctx, dbPath)
    if err != nil {
        return errors.Wrap(err, "open store")
    }
    defer func() { _ = store.Close() }()

    if err := store.Init(ctx); err != nil {
        return errors.Wrap(err, "init db schema")
    }

    var r io.Reader
    if yamlFile == "-" {
        r = os.Stdin
    } else {
        f, err := os.Open(yamlFile)
        if err != nil {
            return errors.Wrap(err, "open yaml file")
        }
        defer func() { _ = f.Close() }()
        r = f
    }

    doc, rawBytes, err := ParseDocument(r)
    if err != nil {
        return errors.Wrap(err, "parse yaml")
    }
    if verbose {
        fmt.Printf("Loaded YAML (%d bytes) version=%q actions=%d\n", len(rawBytes), doc.Version, len(doc.Actions))
    }

    exec := NewExecutor(store, WithVerbose(verbose))
    results, err := exec.Run(ctx, doc)
    if err != nil {
        return errors.Wrap(err, "execute actions")
    }

    // Minimal summary
    now := time.Now().Format(time.RFC3339)
    fmt.Printf("\n== Applied at %s ==\n", now)
    for i, r := range results {
        fmt.Printf("- action[%d] id=%q op=%s -> %s\n", i, r.InputID, r.Operation, r.Summary)
    }
    return nil
}


