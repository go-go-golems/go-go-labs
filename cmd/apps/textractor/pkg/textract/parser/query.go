package parser

import "fmt"

// queryImpl implements the Query interface
type queryImpl struct {
	block   Block
	page    Page
	text    string
	alias   string
	results []QueryResult
}

var _ Query = &queryImpl{}

func newQuery(block Block, page Page) (Query, error) {
	if block.BlockType() != BlockTypeQuery {
		return nil, fmt.Errorf("block type must be QUERY, got %s", block.BlockType())
	}

	query := &queryImpl{
		block: block,
		page:  page,
		text:  block.Text(),
	}

	// Extract alias and results
	if err := query.processQueryData(); err != nil {
		return nil, fmt.Errorf("processing query data: %w", err)
	}

	return query, nil
}

// Text returns the query text
func (q *queryImpl) Text() string {
	return q.text
}

// Alias returns the query alias
func (q *queryImpl) Alias() string {
	return q.alias
}

// Results returns the query results
func (q *queryImpl) Results() []QueryResult {
	return q.results
}

// Page returns the parent page
func (q *queryImpl) Page() Page {
	return q.page
}

// Internal methods

func (q *queryImpl) processQueryData() error {
	// Find alias from block properties
	if raw, ok := q.block.(*blockImpl); ok && raw.rawBlock != nil {
		// Extract alias from AWS block data
		// Implementation depends on AWS SDK structure
	}

	// Find results through relationships
	results := findChildBlocks(q.block, "ANSWER")
	for _, result := range results {
		qr, err := newQueryResult(result, q)
		if err != nil {
			return fmt.Errorf("creating query result: %w", err)
		}
		q.results = append(q.results, qr)
	}

	return nil
}

// queryResultImpl implements the QueryResult interface
type queryResultImpl struct {
	block      Block
	query      Query
	text       string
	confidence float64
}

func newQueryResult(block Block, query Query) (QueryResult, error) {
	if block.BlockType() != BlockTypeQueryResult {
		return nil, fmt.Errorf("block type must be QUERY_RESULT, got %s", block.BlockType())
	}

	return &queryResultImpl{
		block:      block,
		query:      query,
		text:       block.Text(),
		confidence: block.Confidence(),
	}, nil
}

// Text returns the result text
func (qr *queryResultImpl) Text() string {
	return qr.text
}

// Confidence returns the confidence score
func (qr *queryResultImpl) Confidence() float64 {
	return qr.confidence
}

// Query returns the parent query
func (qr *queryResultImpl) Query() Query {
	return qr.query
}

// Block returns the underlying block
func (qr *queryResultImpl) Block() Block {
	return qr.block
}
