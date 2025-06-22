package shared

// VarChangedMsg is emitted when a template variable value is changed in the form.
type VarChangedMsg struct {
	Name  string
	Value string
}

// ToggleChangedMsg is emitted when a toggle or bullet selection is changed in the form.
type ToggleChangedMsg struct {
	SectionID   string
	VariantID   string
	// If BulletIndex is non-nil, toggle the bullet at that index; if nil, toggle the variant state.
	BulletIndex *int
}

// PreviewUpdatedMsg is emitted when the prompt preview has been updated in the root model.
type PreviewUpdatedMsg struct {
	Text string
}

// SectionVariantMsg is emitted when a section variant is cycled in the form.
type SectionVariantMsg struct {
	SectionID string
} 