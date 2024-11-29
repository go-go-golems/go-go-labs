package parser

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/textract"
)

// getRowIndex extracts row index from a cell block
func getRowIndex(block Block) int {
	if raw, ok := block.(*blockImpl); ok && raw.rawBlock != nil {
		return int(aws.Int64Value(raw.rawBlock.RowIndex)) - 1
	}
	return -1
}

// getColumnIndex extracts column index from a cell block
func getColumnIndex(block Block) int {
	if raw, ok := block.(*blockImpl); ok && raw.rawBlock != nil {
		return int(aws.Int64Value(raw.rawBlock.ColumnIndex)) - 1
	}
	return -1
}

// getRowSpan extracts row span from a cell block
func getRowSpan(block Block) int {
	if raw, ok := block.(*blockImpl); ok && raw.rawBlock != nil {
		if raw.rawBlock.RowSpan != nil {
			return int(aws.Int64Value(raw.rawBlock.RowSpan))
		}
	}
	return 1
}

// getColumnSpan extracts column span from a cell block
func getColumnSpan(block Block) int {
	if raw, ok := block.(*blockImpl); ok && raw.rawBlock != nil {
		if raw.rawBlock.ColumnSpan != nil {
			return int(aws.Int64Value(raw.rawBlock.ColumnSpan))
		}
	}
	return 1
}

// getSelectionStatus extracts selection status from a selection element block
func getSelectionStatus(block Block) SelectionStatus {
	if raw, ok := block.(*blockImpl); ok && raw.rawBlock != nil {
		if raw.rawBlock.SelectionStatus != nil {
			switch aws.StringValue(raw.rawBlock.SelectionStatus) {
			case "SELECTED":
				return SelectionStatusSelected
			case "NOT_SELECTED":
				return SelectionStatusNotSelected
			}
		}
	}
	return SelectionStatusNotSelected
}

// getRelationshipIDs extracts IDs from a relationship
func getRelationshipIDs(rel *textract.Relationship) []string {
	if rel == nil || rel.Ids == nil {
		return nil
	}

	ids := make([]string, len(rel.Ids))
	for i, id := range rel.Ids {
		ids[i] = aws.StringValue(id)
	}
	return ids
}

// findRelationshipsByType finds relationships of a specific type
func findRelationshipsByType(block *textract.Block, relType string) []*textract.Relationship {
	if block == nil || block.Relationships == nil {
		return nil
	}

	var matches []*textract.Relationship
	for _, rel := range block.Relationships {
		if aws.StringValue(rel.Type) == relType {
			matches = append(matches, rel)
		}
	}
	return matches
}

// findChildBlocks finds child blocks by relationship type
func findChildBlocks(block Block, relType string) []Block {
	if raw, ok := block.(*blockImpl); ok && raw.rawBlock != nil {
		rels := findRelationshipsByType(raw.rawBlock, relType)
		if len(rels) == 0 {
			return nil
		}

		var children []Block
		for _, rel := range rels {
			for _, id := range getRelationshipIDs(rel) {
				if child := raw.document.(*documentImpl).blockIndex[id]; child != nil {
					children = append(children, child)
				}
			}
		}
		return children
	}
	return nil
}
