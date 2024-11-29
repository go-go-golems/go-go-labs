# Document Interface
**Primary Documentation Files:**
- `02-text-detection.txt` - Core document structure and Block relationships
- `textract-api-docs/api-DocumentMetadata.txt` - Document metadata structure
- `textract-api-docs/api-Block.txt` - Block structure that makes up document

Key concepts from docs:
```
- Document is made up of Block objects
- Contains list of child IDs for lines of text, key-value pairs, tables, queries
- Metadata includes number of pages
- Document can be processed sync or async
```

# Page Interface
**Primary Documentation Files:**
- `03-pages.txt` - Page structure and content
- `textract-api-docs/api-Block.txt` - PAGE block type details
- `09-layout-response.txt` - Layout elements on a page

Key concepts from docs:
```
- Each page contains child blocks for detected items
- Can contain: lines, tables, forms, key-value pairs, queries
- Has geometry information (bounding box)
- Returns items in implied reading order (left to right, top to bottom)
```

# Block Interface
**Primary Documentation Files:**
- `02-text-detection.txt` - Block structure and relationships
- `textract-api-docs/api-Block.txt` - Complete Block API reference
- `textract-api-docs/api-Relationship.txt` - Relationship types

Key concepts from docs:
```
- Basic unit of all detected items
- Has unique identifier
- Contains confidence scores
- Has relationships (parent/child)
- Contains geometry information
```

# Line Interface
**Primary Documentation Files:**
- `04-lines-words.txt` - Line and word structure
- `textract-api-docs/api-Block.txt` - LINE block type specifics

Key concepts from docs:
```
- String of tab-delimited contiguous words
- Contains WORD blocks as children
- Has confidence scores
- Has geometry information
```

# Table Interface
**Primary Documentation Files:**
- `06-tables.txt` - Comprehensive table information
- `textract-api-docs/api-Block.txt` - TABLE block type details

Key concepts from docs:
```
- Contains cells, merged cells
- Has table titles and footers
- Can be structured or semi-structured
- Contains row/column information
- Has specific relationship types
```

# Form Interface
**Primary Documentation Files:**
- `05-form-data.txt` - Form data and key-value pairs
- `07-selection-elements.txt` - Selection elements in forms
- `textract-api-docs/api-Block.txt` - KEY_VALUE_SET block types

Key concepts from docs:
```
- Contains key-value pairs
- Can include selection elements
- Keys and values have relationships
- Has confidence scores
```

# KeyValue Interface
**Primary Documentation Files:**
- `05-form-data.txt` - Detailed key-value structure
- `textract-api-docs/api-Block.txt` - KEY_VALUE_SET block type

Key concepts from docs:
```
- KEY and VALUE blocks in KEY_VALUE_SET
- Contains EntityType (KEY or VALUE)
- Has relationships between key and value
- Contains child words
```

# SelectionElement Interface
**Primary Documentation Files:**
- `07-selection-elements.txt` - Selection element details
- `textract-api-docs/api-Block.txt` - SELECTION_ELEMENT block type

Key concepts from docs:
```
- Can be in forms or tables
- Has selection status (SELECTED/NOT_SELECTED)
- Has confidence score
- Contains geometry information
```

# Query Interface
**Primary Documentation Files:**
- `08-queries.txt` - Query structure and responses
- `textract-api-docs/api-Query.txt` - Query API details
- `textract-api-docs/api-Block.txt` - QUERY block types

Key concepts from docs:
```
- Contains question and alias
- Has relationship to answers
- Contains confidence scores
- Can specify page ranges
```

# Geometry Types
**Primary Documentation Files:**
- `textract-api-docs/api-Geometry.txt` - Geometry structure
- `textract-api-docs/api-BoundingBox.txt` - BoundingBox details
- `textract-api-docs/api-Point.txt` - Point structure

Key concepts from docs:
```
- BoundingBox uses ratios of page dimensions
- Points are coordinate pairs
- Polygon provides fine-grained boundary
- Coordinates relative to top-left origin
```

# Support Types (FilterOptions, DocumentOptions)
**Primary Documentation Files:**
- `02-text-detection.txt` - Block types and filtering concepts
- `textract-api-docs/api-Block.txt` - Entity types and block types
- Various docs showing confidence scores and options

Key concepts from docs:
```
- Block types define item categories
- Entity types specify roles
- Confidence scores are 0-100
- Options affect processing behavior
```

This mapping should help ensure the implementation aligns with AWS Textract's actual behavior and capabilities as documented.
