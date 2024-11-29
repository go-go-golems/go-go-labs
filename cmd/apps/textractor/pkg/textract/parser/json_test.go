package parser

import (
	"strings"
	"testing"
)

func TestLoadFromJSONReader(t *testing.T) {
	// Test single document
	t.Run("single document", func(t *testing.T) {
		jsonData := `{
			"DocumentMetadata": {
				"Pages": 1
			},
			"Blocks": [
				{
					"Id": "page1",
					"BlockType": "PAGE",
					"Confidence": 100,
					"Geometry": {
						"BoundingBox": {
							"Width": 1.0,
							"Height": 1.0,
							"Left": 0.0,
							"Top": 0.0
						},
						"Polygon": [
							{"X": 0.0, "Y": 0.0},
							{"X": 1.0, "Y": 0.0},
							{"X": 1.0, "Y": 1.0},
							{"X": 0.0, "Y": 1.0}
						]
					},
					"Page": 1,
					"Relationships": [
						{
							"Type": "CHILD",
							"Ids": ["line1"]
						}
					]
				},
				{
					"Id": "line1",
					"BlockType": "LINE",
					"Text": "Hello World",
					"Confidence": 99.5,
					"Geometry": {
						"BoundingBox": {
							"Width": 0.5,
							"Height": 0.1,
							"Left": 0.1,
							"Top": 0.1
						}
					},
					"Page": 1,
					"Relationships": [
						{
							"Type": "CHILD",
							"Ids": ["word1", "word2"]
						}
					]
				}
			]
		}`

		docs, err := LoadFromJSONReader(strings.NewReader(jsonData))
		if err != nil {
			t.Fatalf("Failed to load JSON: %v", err)
		}

		if len(docs) != 1 {
			t.Fatalf("Expected 1 document, got %d", len(docs))
		}

		doc := docs[0]
		if doc.PageCount() != 1 {
			t.Errorf("Expected 1 page, got %d", doc.PageCount())
		}

		pages := doc.Pages()
		if len(pages) != 1 {
			t.Fatalf("Expected 1 page, got %d", len(pages))
		}

		lines := pages[0].Lines()
		if len(lines) != 1 {
			t.Fatalf("Expected 1 line, got %d", len(lines))
		}

		if lines[0].Text() != "Hello World" {
			t.Errorf("Expected 'Hello World', got '%s'", lines[0].Text())
		}
	})

	// Test multiple documents
	t.Run("multiple documents", func(t *testing.T) {
		jsonData := `[
			{
				"DocumentMetadata": {"Pages": 1},
				"Blocks": [
					{
						"Id": "page1",
						"BlockType": "PAGE",
						"Confidence": 100,
						"Page": 1,
						"Geometry": {
							"BoundingBox": {
								"Width": 1.0,
								"Height": 1.0,
								"Left": 0.0,
								"Top": 0.0
							}
						}
					}
				]
			},
			{
				"DocumentMetadata": {"Pages": 1},
				"Blocks": [
					{
						"Id": "page1",
						"BlockType": "PAGE",
						"Confidence": 100,
						"Page": 1,
						"Geometry": {
							"BoundingBox": {
								"Width": 1.0,
								"Height": 1.0,
								"Left": 0.0,
								"Top": 0.0
							}
						}
					}
				]
			}
		]`

		docs, err := LoadFromJSONReader(strings.NewReader(jsonData))
		if err != nil {
			t.Fatalf("Failed to load JSON: %v", err)
		}

		if len(docs) != 2 {
			t.Fatalf("Expected 2 documents, got %d", len(docs))
		}
	})
}
