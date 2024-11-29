// Example 1: Basic document processing
func processDocument(response *TextractResponse) error {
    // Create new document with options
    doc, err := textract.NewDocument(response, 
        WithConfidenceThreshold(0.9),
        WithCustomProcessor("CUSTOM", myProcessor),
    )
    if err != nil {
        return fmt.Errorf("creating document: %w", err)
    }

    // Process pages
    for _, page := range doc.Pages() {
        // Process lines
        for _, line := range page.Lines() {
            fmt.Printf("Line: %s (%.2f%%)\n", line.Text(), line.Confidence())
            
            // Process words in line
            for _, word := range line.Words() {
                fmt.Printf("Word: %s\n", word.Text())
            }
        }

        // Process tables
        for _, table := range page.Tables() {
            // Check table title
            if title := table.Title(); title != nil {
                fmt.Printf("Table Title: %s\n", title.Text())
            }

            // Process cells
            for _, row := range table.Rows() {
                for _, cell := range row.Cells() {
                    fmt.Printf("Cell: %s\n", cell.Text())
                }
            }
        }

        // Process forms
        form := page.Forms()
        for _, field := range form.Fields() {
            fmt.Printf("Field: %s = %s\n", 
                field.KeyText(), 
                field.ValueText())
        }
    }

    return nil
}

// Example 2: Advanced filtering and processing
func processWithFiltering(doc textract.Document) error {
    // Filter blocks with specific criteria
    blocks := doc.FilterBlocks(FilterOptions{
        MinConfidence: 0.95,
        BlockTypes: []BlockType{
            BlockTypeKeyValueSet,
            BlockTypeTable,
        },
    })

    // Process filtered blocks
    for _, block := range blocks {
        switch block.BlockType() {
        case BlockTypeKeyValueSet:
            if block.EntityType() == EntityTypeKey {
                kv := block.AsKeyValue()
                fmt.Printf("KV: %s = %s\n", 
                    kv.KeyText(), 
                    kv.ValueText())
            }
        case BlockTypeTable:
            table := block.AsTable()
            fmt.Printf("Table with %d rows\n", 
                table.RowCount())
        }
    }

    return nil
}

// Example 3: Query processing
func processQueries(doc textract.Document) error {
    // Create queries
    queries := []textract.Query{
        {Text: "What is the invoice number?", Alias: "INVOICE_NUM"},
        {Text: "What is the total amount?", Alias: "TOTAL"},
    }

    // Process document with queries
    results, err := doc.ProcessQueries(queries)
    if err != nil {
        return err
    }

    // Handle results
    for _, result := range results {
        fmt.Printf("Query '%s' (%s): %s (%.2f%%)\n",
            result.Query().Text(),
            result.Query().Alias(),
            result.Text(),
            result.Confidence())
    }

    return nil
}
