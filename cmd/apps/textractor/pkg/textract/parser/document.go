package parser

import (
	"fmt"
	"slices"
	"sort"

	"github.com/rs/zerolog/log"
)

// documentImpl implements the Document interface
type documentImpl struct {
	raw           *TextractResponse
	pages         []Page
	metadata      DocumentMetadata
	blockIndex    map[string]Block
	pageIndex     map[int]Page
	keyValuePairs []KeyValue
}

// NewDocument creates a new Document from a TextractResponse
func NewDocument(response *TextractResponse, opts ...DocumentOption) (Document, error) {
	if response == nil {
		return nil, fmt.Errorf("textract response cannot be nil")
	}

	// Apply options
	options := &DocumentOptions{
		ConfidenceThreshold: 0.0,
		EnableMergedCells:   true,
	}
	for _, opt := range opts {
		opt(options)
	}

	doc := &documentImpl{
		raw:        response,
		blockIndex: make(map[string]Block),
		pageIndex:  make(map[int]Page),
		metadata: DocumentMetadata{
			Pages: int(*response.Metadata.Pages),
		},
	}

	// Process blocks and build indexes
	if err := doc.processBlocks(options); err != nil {
		return nil, fmt.Errorf("processing blocks: %w", err)
	}

	return doc, nil
}

// Pages returns all pages in the document
func (d *documentImpl) Pages() []Page {
	return d.pages
}

// Raw returns the underlying TextractResponse
func (d *documentImpl) Raw() *TextractResponse {
	return d.raw
}

// GetPageByIndex returns a specific page by its index
func (d *documentImpl) GetPageByIndex(idx int) (Page, error) {
	if page, ok := d.pageIndex[idx]; ok {
		return page, nil
	}
	return nil, fmt.Errorf("page index %d not found", idx)
}

// FindKeyValuePairs finds all key-value pairs with matching key text
func (d *documentImpl) FindKeyValuePairs(key string) []KeyValue {
	var results []KeyValue
	for _, page := range d.pages {
		for _, form := range page.Forms() {
			kvs := form.SearchFieldsByKey(key)
			results = append(results, kvs...)
		}
	}
	return results
}

// FilterBlocks returns blocks matching the given criteria
func (d *documentImpl) FilterBlocks(opts FilterOptions) []Block {
	var results []Block
	for _, block := range d.blockIndex {
		if d.matchesFilter(block, opts) {
			results = append(results, block)
		}
	}
	return results
}

// PageCount returns the total number of pages
func (d *documentImpl) PageCount() int {
	return d.metadata.Pages
}

// DocumentMetadata returns the document metadata
func (d *documentImpl) DocumentMetadata() DocumentMetadata {
	return d.metadata
}

// Internal helper methods

func (d *documentImpl) processBlocks(options *DocumentOptions) error {
	// First pass: create all blocks and build index
	for _, rawBlock := range d.raw.Blocks {
		block, err := newBlock(rawBlock, d)
		if err != nil {
			return fmt.Errorf("creating block: %w", err)
		}
		if block.BlockType() == BlockTypeTable {
			log.Info().Str("block.blockID", block.ID()).
				Str("block.type", string(block.BlockType())).
				Msg("created table block")
		}
		d.blockIndex[block.ID()] = block
	}

	// Second pass: establish relationships between blocks
	for _, rawBlock := range d.raw.Blocks {
		block := d.blockIndex[*rawBlock.Id]
		block_ := block.(*blockImpl)

		// Process relationships from raw block
		if rawBlock.Relationships != nil {
			for _, rel := range rawBlock.Relationships {
				relType := stringValue(rel.Type)
				switch relType {
				case "CHILD":
					// Add children to current block
					log.Info().Str("block.blockID", block.ID()).
						Str("block.type", string(block.BlockType())).
						Int("block.page", block.Page()).
						Interface("rel.Ids", rel.Ids).
						Msg("processing children")
					for _, childID := range rel.Ids {
						if childBlock, ok := d.blockIndex[*childID]; ok {
							if block.BlockType() == BlockTypeTable {
								log.Info().Str("child.blockID", childBlock.ID()).Str("parent.blockID", block.ID()).Msg("adding child")
							}

							block_.children = append(block_.children, childBlock)
							// Add parent reference to child
							childImpl := childBlock.(*blockImpl)
							childImpl.parents = append(childImpl.parents, block)
						} else {
							// Print all known block index values and their types
							log.Info().Msg("Known block index values:")
							blockIDs := make([]string, 0, len(d.blockIndex))
							for blockID := range d.blockIndex {
								blockIDs = append(blockIDs, blockID)
							}
							sort.Strings(blockIDs)

							for _, blockID := range blockIDs {
								block := d.blockIndex[blockID]
								log.Info().
									Str("blockID", blockID).
									Str("blockType", string(block.BlockType())).
									Msg("Block in index")
							}
							log.Fatal().Str("child.blockID", *childID).Msg("child block not found")
						}
					}
				case "MERGED_CELL":
					// Handle merged cell relationships if needed
					// This is used for table cells that span multiple rows/columns
					continue

				case "VALUE":
					// Handle key-value relationships
					// These are processed later in processBlockRelationships
					// I think these must be added as parents...
					continue
				}
			}
		}
	}

	// Third pass: process higher-level relationships and validate structure
	for _, block := range d.blockIndex {
		if err := d.processBlockRelationships(block); err != nil {
			return fmt.Errorf("processing relationships: %w", err)
		}
	}

	// Build page hierarchy
	if err := d.buildPages(); err != nil {
		return fmt.Errorf("building pages: %w", err)
	}

	return nil
}

func (d *documentImpl) matchesFilter(block Block, opts FilterOptions) bool {
	if block.Confidence() < opts.MinConfidence {
		return false
	}

	if len(opts.BlockTypes) > 0 {
		matched := false
		for _, bt := range opts.BlockTypes {
			if block.BlockType() == bt {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(opts.EntityTypes) > 0 {
		matched := false
		for _, et := range opts.EntityTypes {
			if slices.Contains(block.EntityTypes(), et) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (d *documentImpl) processBlockRelationships(block Block) error {
	// Process children
	children := block.Children()
	if len(children) == 0 {
		return nil
	}

	switch block.BlockType() {
	case BlockTypePage:
		for _, child := range children {
			if err := d.processPageChild(block, child); err != nil {
				return fmt.Errorf("processing page child: %w", err)
			}
		}

	case BlockTypeLine:
		for _, child := range children {
			if err := d.processLineChild(block, child); err != nil {
				return fmt.Errorf("processing line child: %w", err)
			}
		}

	case BlockTypeTable:
		for _, child := range children {
			if err := d.processTableChild(block, child); err != nil {
				return fmt.Errorf("processing table child: %w", err)
			}
		}

	case BlockTypeKeyValueSet:
		// Process key-value relationships
		if slices.Contains(block.EntityTypes(), EntityTypeKey) {
			// Find corresponding value block through parent relationships
			for _, parent := range block.Parents() {
				if parent.BlockType() == BlockTypeKeyValueSet &&
					slices.Contains(parent.EntityTypes(), EntityTypeValue) {
					kv, err := newKeyValue(block, parent, nil) // Form will be set later
					if err != nil {
						return fmt.Errorf("creating key-value pair: %w", err)
					}
					d.keyValuePairs = append(d.keyValuePairs, kv)
				}
			}
		}

		// Process children of the key-value set
		for _, child := range children {
			if err := d.processKeyValueChild(block, child); err != nil {
				return fmt.Errorf("processing key-value child: %w", err)
			}
		}
	}

	return nil
}

func (d *documentImpl) processPageChild(page Block, child Block) error {
	switch child.BlockType() {
	case BlockTypeLine, BlockTypeTable, BlockTypeKeyValueSet,
		BlockTypeSelectionElement, BlockTypeQuery, BlockTypeQueryResult:
		return nil // These are valid child types for a page
	default:
		return fmt.Errorf("invalid page child type: %s", child.BlockType())
	}
}

func (d *documentImpl) processLineChild(line Block, child Block) error {
	if child.BlockType() != BlockTypeWord {
		return fmt.Errorf("invalid line child type: %s", child.BlockType())
	}
	return nil
}

func (d *documentImpl) processTableChild(table Block, child Block) error {
	if child.BlockType() != BlockTypeCell {
		return fmt.Errorf("invalid table child type: %s", child.BlockType())
	}
	return nil
}

func (d *documentImpl) processKeyValueChild(kvSet Block, child Block) error {
	switch child.BlockType() {
	case BlockTypeWord, BlockTypeSelectionElement:
		return nil
	default:
		return fmt.Errorf("invalid key-value child type: %s", child.BlockType())
	}
}

func (d *documentImpl) buildPages() error {
	// Find all PAGE blocks
	var pageBlocks []Block
	for _, block := range d.blockIndex {
		if block.BlockType() == BlockTypePage {
			pageBlocks = append(pageBlocks, block)
		}
	}

	// Sort pages by page number
	sort.Slice(pageBlocks, func(i, j int) bool {
		return pageBlocks[i].Page() < pageBlocks[j].Page()
	})

	// Create page objects
	for _, pageBlock := range pageBlocks {
		page, err := newPage(d, pageBlock, pageBlock.Page())
		if err != nil {
			return fmt.Errorf("creating page %d: %w", pageBlock.Page(), err)
		}

		// Store in both slices and map for different access patterns
		d.pages = append(d.pages, page)
		d.pageIndex[page.Number()] = page

		// Organize page elements in reading order
		if err := d.organizePageElements(page); err != nil {
			return fmt.Errorf("organizing page %d elements: %w", page.Number(), err)
		}
	}

	// Associate key-value pairs with forms
	for _, kv := range d.keyValuePairs {
		// Find the page containing this key-value pair
		pageNum := kv.Key().Page()
		page, err := d.GetPageByIndex(pageNum)
		if err != nil {
			return fmt.Errorf("finding page for key-value pair: %w", err)
		}

		// Find or create a form for this page
		var form Form
		forms := page.Forms()
		if len(forms) == 0 {
			form = newForm(page)
			forms = append(forms, form)
		} else {
			form = forms[0] // Use first form on page
		}

		// Update the key-value pair's form reference
		kvImpl := kv.(*keyValueImpl)
		kvImpl.form = form

		// Add to form
		formImpl := form.(*formImpl)
		formImpl.addField(kv)
	}

	return nil
}

// Helper function for buildPages

func (d *documentImpl) organizePageElements(page Page) error {
	pageImpl := page.(*pageImpl)

	// Sort elements by reading order (top-to-bottom, left-to-right)
	sort.SliceStable(pageImpl.lines, func(i, j int) bool {
		bi := pageImpl.lines[i].BoundingBox()
		bj := pageImpl.lines[j].BoundingBox()

		// If lines are roughly on the same vertical position (within threshold)
		if abs(bi.Top-bj.Top) < 0.02 { // 2% threshold
			return bi.Left < bj.Left // Sort left-to-right
		}
		return bi.Top < bj.Top // Sort top-to-bottom
	})

	// Sort tables by position
	sort.SliceStable(pageImpl.tables, func(i, j int) bool {
		bi := pageImpl.tables[i].BoundingBox()
		bj := pageImpl.tables[j].BoundingBox()
		return bi.Top < bj.Top
	})

	return nil
}

// Helper function for float comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
