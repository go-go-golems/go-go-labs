package parser

import (
	"fmt"
	"slices"
	"strings"
)

// SelectionElement represents a checkbox or radio button in a form
type SelectionElement interface {
	// Status
	IsSelected() bool
	SelectionStatus() SelectionStatus
	Confidence() float64

	// Navigation
	Block() Block
	Form() Form

	// Geometry
	BoundingBox() BoundingBox
	Polygon() []Point
}

// SelectionElementType represents the type of selection element
type SelectionElementType string

const (
	SelectionElementTypeCheckbox    SelectionElementType = "CHECKBOX"
	SelectionElementTypeRadioButton SelectionElementType = "RADIO_BUTTON"
)

// formImpl implements the Form interface
type formImpl struct {
	page              Page
	fields            []KeyValue
	selectionElements []SelectionElement
	keyIndex          map[string][]KeyValue // Maps key text to KeyValue pairs
}

func newForm(page Page) Form {
	return &formImpl{
		page:     page,
		keyIndex: make(map[string][]KeyValue),
	}
}

// Fields returns all key-value pairs in the form
func (f *formImpl) Fields() []KeyValue {
	return f.fields
}

// SelectionElements returns all selection elements in the form
func (f *formImpl) SelectionElements() []SelectionElement {
	return f.selectionElements
}

// GetFieldByKey returns the first key-value pair with matching key text
func (f *formImpl) GetFieldByKey(key string) KeyValue {
	matches := f.SearchFieldsByKey(key)
	if len(matches) > 0 {
		return matches[0]
	}
	return nil
}

// SearchFieldsByKey returns all key-value pairs with matching key text
func (f *formImpl) SearchFieldsByKey(key string) []KeyValue {
	// Try exact match first
	if fields, ok := f.keyIndex[key]; ok {
		return fields
	}

	// Try case-insensitive match
	lowerKey := strings.ToLower(key)
	for k, fields := range f.keyIndex {
		if strings.ToLower(k) == lowerKey {
			return fields
		}
	}

	return nil
}

// Page returns the parent page
func (f *formImpl) Page() Page {
	return f.page
}

// Internal methods

func (f *formImpl) addField(field KeyValue) {
	f.fields = append(f.fields, field)
	key := field.KeyText()
	f.keyIndex[key] = append(f.keyIndex[key], field)
}

func (f *formImpl) addSelectionElement(element SelectionElement) {
	f.selectionElements = append(f.selectionElements, element)
}

// keyValueImpl implements the KeyValue interface
type keyValueImpl struct {
	keyBlock   Block
	valueBlock Block
	form       Form
}

func newKeyValue(keyBlock, valueBlock Block, form Form) (KeyValue, error) {
	if keyBlock.BlockType() != BlockTypeKeyValueSet || !slices.Contains(keyBlock.EntityTypes(), EntityTypeKey) {
		return nil, fmt.Errorf("invalid key block type: %s/%s", keyBlock.BlockType(), keyBlock.EntityTypes())
	}

	if valueBlock.BlockType() != BlockTypeKeyValueSet || !slices.Contains(valueBlock.EntityTypes(), EntityTypeValue) {
		return nil, fmt.Errorf("invalid value block type: %s/%s", valueBlock.BlockType(), valueBlock.EntityTypes())
	}

	return &keyValueImpl{
		keyBlock:   keyBlock,
		valueBlock: valueBlock,
		form:       form,
	}, nil
}

// Key returns the key block
func (kv *keyValueImpl) Key() Block {
	return kv.keyBlock
}

// Value returns the value block
func (kv *keyValueImpl) Value() Block {
	return kv.valueBlock
}

// KeyText returns the key text
func (kv *keyValueImpl) KeyText() string {
	return kv.keyBlock.Text()
}

// ValueText returns the value text
func (kv *keyValueImpl) ValueText() string {
	return kv.valueBlock.Text()
}

// Confidence returns the minimum confidence between key and value
func (kv *keyValueImpl) Confidence() float64 {
	return min(kv.keyBlock.Confidence(), kv.valueBlock.Confidence())
}

// Form returns the parent form
func (kv *keyValueImpl) Form() Form {
	return kv.form
}

// Helper function
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
