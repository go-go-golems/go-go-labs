package debug

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg/textract/parser"
	"github.com/spf13/cobra"
)

func newDumpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dumps the contents of a Textract JSON file",
		Long:  "Parses a Textract JSON file and dumps all available information using the Document interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, file := range args {
				docs, err := parser.LoadFromJSON(file)
				if err != nil {
					return fmt.Errorf("loading JSON: %w", err)
				}

				fmt.Printf("File: %s\n", filepath.Base(file))
				fmt.Printf("Documents: %d\n", len(docs))
				fmt.Printf("===================%s\n", strings.Repeat("=", len(filepath.Base(file))))

				for docIdx, doc := range docs {
					fmt.Printf("\nDocument %d:\n", docIdx+1)
					fmt.Printf("  Pages: %d\n", doc.PageCount())

					// Process each page
					for _, page := range doc.Pages() {
						fmt.Printf("\nPage %d:\n", page.Number())
						fmt.Printf("  Lines: %d\n", len(page.Lines()))
						fmt.Printf("  Tables: %d\n", len(page.Tables()))
						fmt.Printf("  Forms: %d\n", len(page.Forms()))
						fmt.Printf("  Words: %d\n", len(page.Words()))

						// Print lines
						fmt.Printf("\n  Line contents:\n")
						for i, line := range page.Lines() {
							fmt.Printf("    %d: %q (confidence: %.2f%%)\n", i+1, line.Text(), line.Confidence())
						}

						// Print tables
						for i, table := range page.Tables() {
							fmt.Printf("\n  Table %d:\n", i+1)
							fmt.Printf("    Rows: %d\n", table.RowCount())
							fmt.Printf("    Columns: %d\n", table.ColumnCount())

							// Print table contents
							for r, row := range table.Rows() {
								fmt.Printf("    Row %d:\n", r+1)
								for c, cell := range row.Cells() {
									fmt.Printf("      [%d,%d]: %q\n", r+1, c+1, cell.Text())
								}
							}
						}

						// Print forms
						for i, form := range page.Forms() {
							fmt.Printf("\n  Form %d:\n", i+1)
							fmt.Printf("    Fields: %d\n", len(form.Fields()))
							fmt.Printf("    Selection Elements: %d\n", len(form.SelectionElements()))

							// Print form fields
							for _, field := range form.Fields() {
								fmt.Printf("    Field: %q -> %q (confidence: %.2f%%)\n",
									field.KeyText(),
									field.ValueText(),
									field.Confidence())
							}
						}
					}
				}
			}

			return nil
		},
	}

	return cmd
}
