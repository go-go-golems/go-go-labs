package pkg

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestApplyChange(t *testing.T) {
	tests := []struct {
		name        string
		sourceLines []string
		change      Change
		want        []string
		wantErr     error
	}{
		{
			name:        "replace action with valid parameters",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionReplace, Old: "line2", New: "newLine"},
			want:        []string{"line1", "newLine", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Empty source file",
			sourceLines: []string{},
			change:      Change{Action: ActionReplace, Old: "line1", New: "newLine"},
			want:        []string{},
			wantErr:     ErrCodeBlock{source: []string{"line1"}},
		},
		// NOTE(manuel, 2023-10-30) this will replace the first matching whitespace line, it is a bit of special case
		{
			name:        "Replacing whitespace line",
			sourceLines: []string{"\t", "    ", "line3"},
			change:      Change{Action: ActionReplace, Old: "    ", New: "newLine"},
			want:        []string{"newLine", "    ", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Empty target line",
			sourceLines: []string{"", "line2", "line3"},
			change:      Change{Action: ActionReplace, Old: "", New: "newLine"},
			want:        []string{"newLine", "line2", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Multiple line replacement",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionReplace, Old: "line1\nline2", New: "newLine1\nnewLine2"},
			want:        []string{"newLine1", "newLine2", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Beginning of file replacement",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionReplace, Old: "line1", New: "newLine"},
			want:        []string{"newLine", "line2", "line3"},
			wantErr:     nil,
		},
		{
			name:        "End of file replacement",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionReplace, Old: "line3", New: "newLine"},
			want:        []string{"line1", "line2", "newLine"},
			wantErr:     nil,
		},
		{
			name:        "Non-existent content",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionReplace, Old: "line4", New: "newLine"},
			wantErr:     ErrCodeBlock{source: []string{"line4"}},
		},
		{
			name:        "Mismatch with empty lines",
			sourceLines: []string{"", "", "line3"},
			change:      Change{Action: ActionReplace, Old: "", New: "newLine"},
			want:        []string{"newLine", "", "line3"},
			wantErr:     nil, // or an error if the behavior should be different
		},
		{
			name:        "Identical old and new content",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionReplace, Old: "line2", New: "line2"}, // no change in the content
			want:        []string{"line1", "line2", "line3"},                       // expected no change in the lines
			wantErr:     nil,                                                       // we expect no error here, as this is a valid operation, though it does nothing
		},
		{
			name:        "Replacing with nothing",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionReplace, Old: "line2", New: ""}, // the content of 'line2' is replaced with an empty string
			want:        []string{"line1", "", "line3"},                       // 'line2' is now empty, effectively deleting the content
			wantErr:     nil,                                                  // no error is expected here as this is a valid operation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDifferential(tt.sourceLines)
			err := d.ApplyChange(tt.change)
			// check if err are of the same type
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ApplyChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// check for nil == []string{}
			if tt.want == nil {
				tt.want = []string{}
			}
			if d.SourceLines == nil {
				d.SourceLines = []string{}
			}
			if err == nil {
				if !reflect.DeepEqual(d.SourceLines, tt.want) {
					t.Errorf("ApplyChange() = %v, want %v", d.SourceLines, tt.want)
				}
			}
		})
	}
}

func TestApplyChange_ActionDelete(t *testing.T) {
	tests := []struct {
		name        string
		sourceLines []string
		change      Change
		want        []string
		wantErr     error
	}{
		{
			name:        "Valid deletion",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionDelete, Content: "line2"},
			want:        []string{"line1", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Non-existent content",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionDelete, Content: "line4"},
			wantErr:     ErrCodeBlock{source: []string{"line4"}},
		},
		{
			name:        "Empty target content",
			sourceLines: []string{"line1", "", "line3"},
			change:      Change{Action: ActionDelete, Content: ""},
			want:        []string{"line1", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Deleting multiple lines",
			sourceLines: []string{"line1", "line2", "line3", "line4"},
			change:      Change{Action: ActionDelete, Content: "line2\nline3"},
			want:        []string{"line1", "line4"},
			wantErr:     nil,
		},
		{
			name:        "Deleting at the beginning",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionDelete, Content: "line1"},
			want:        []string{"line2", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Deleting at the end",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionDelete, Content: "line3"},
			want:        []string{"line1", "line2"},
			wantErr:     nil,
		},
		{
			name:        "Empty source lines",
			sourceLines: []string{},
			change:      Change{Action: ActionDelete, Content: "line1"},
			wantErr:     ErrCodeBlock{source: []string{"line1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDifferential(tt.sourceLines)
			err := d.ApplyChange(tt.change)

			// Error handling: Check if the expected error matches the actual error
			if (err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error()) || (err == nil && tt.wantErr != nil) || (err != nil && tt.wantErr == nil) {
				t.Errorf("ApplyChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// In case of no error, check if the output matches the expected result
			if err == nil {
				if !reflect.DeepEqual(d.SourceLines, tt.want) {
					t.Errorf("ApplyChange() = %v, want %v", d.SourceLines, tt.want)
				}
			}
		})
	}
}

func TestApplyChange_ActionMove(t *testing.T) {
	tests := []struct {
		name        string
		sourceLines []string
		change      Change
		want        []string
		wantErr     error
	}{
		{
			name:        "Valid move to a new location",
			sourceLines: []string{"line1", "line2", "line3", "line4"},
			change:      Change{Action: ActionMove, Content: "line3", Above: "line1"},
			want:        []string{"line3", "line1", "line2", "line4"},
			wantErr:     nil,
		},
		{
			name:        "Move to the beginning",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionMove, Content: "line2", Above: "line1"},
			want:        []string{"line2", "line1", "line3"},
			wantErr:     nil,
		},
		//{
		//	name:        "Move to the end",
		//	SourceLines: []string{"line1", "line2", "line3"},
		//	change:      Change{Action: ActionAppend, Content: "line1"},
		//	want:        []string{"line2", "line3", "line1"},
		//	wantErr:     nil,
		//},
		// TODO(manuel, 2023-10-28) We don't support moving to the end rn
		{
			name:        "Non-existent content",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionMove, Content: "line4", Above: "line2"},
			want:        nil, // or []string{"line1", "line2", "line3"} if the function does not modify the input on error
			wantErr:     ErrCodeBlock{source: []string{"line4"}},
		},
		{
			name:        "Non-existent destination",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionMove, Content: "line2", Above: "line4"},
			want:        nil, // or []string{"line1", "line2", "line3"} if the function does not modify the input on error
			wantErr:     ErrCodeBlock{source: []string{"line4"}},
		},
		{
			name:        "Move multiple lines",
			sourceLines: []string{"line1", "line2", "line3", "line4"},
			change:      Change{Action: ActionMove, Content: "line1\nline2", Above: "line4"},
			want:        []string{"line3", "line1", "line2", "line4"},
			wantErr:     nil,
		},
		{
			name:        "Move with empty content",
			sourceLines: []string{"line1", "", "line3"},
			change:      Change{Action: ActionMove, Content: "", Above: "line3"},
			want:        []string{"line1", "", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Move within empty source lines",
			sourceLines: []string{},
			change:      Change{Action: ActionMove, Content: "line1", Above: "line3"},
			want:        nil, // or []string{} if the function does not modify the input on error
			wantErr:     ErrCodeBlock{source: []string{"line1"}},
		},
		{
			name:        "Moving content to its current location",
			sourceLines: []string{"line1", "line2", "line3"},
			change:      Change{Action: ActionMove, Content: "line2", Above: "line3"},
			want:        []string{"line1", "line2", "line3"}, // no change expected
			wantErr:     nil,                                 // no error is expected here, as this is a valid operation, though it does nothing
		},
		{
			name:        "Moving content to a position indicated by content above",
			sourceLines: []string{"line1", "line2", "line3", "line4"},
			change:      Change{Action: ActionMove, Content: "line4", Above: "line2"},
			want:        []string{"line1", "line4", "line2", "line3"}, // 'line4' should now be above 'line2'
			wantErr:     nil,
		},
		{
			name: "Move content lower in the file",
			sourceLines: []string{
				"line1",
				"line2",
				"line3",
				"line4",
				"line5",
				"line6",
			},
			change: Change{
				Action:  ActionMove,
				Content: "line2\nline3",
				Above:   "line6",
			},
			want:    []string{"line1", "line4", "line5", "line2", "line3", "line6"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDifferential(tt.sourceLines)
			err := d.ApplyChange(tt.change)

			// Error handling: Check if the expected error matches the actual error
			if (err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error()) || (err == nil && tt.wantErr != nil) || (err != nil && tt.wantErr == nil) {
				t.Errorf("ApplyChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// In case of no error, check if the output matches the expected result
			if err == nil {
				if !reflect.DeepEqual(d.SourceLines, tt.want) {
					t.Errorf("ApplyChange() = %v, want %v", d.SourceLines, tt.want)
				}
			}
		})
	}
}

func TestApplyChange_ActionInsert(t *testing.T) {
	tests := []struct {
		name        string
		sourceLines []string
		change      Change
		want        []string
		wantErr     error
	}{
		{
			name:        "Insert at the beginning",
			sourceLines: []string{"line2", "line3"},
			change:      Change{Action: ActionPrepend, Content: "line1"},
			want:        []string{"line1", "line2", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Insert at the end",
			sourceLines: []string{"line1", "line2"},
			change:      Change{Action: ActionAppend, Content: "line3"},
			want:        []string{"line1", "line2", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Insert with non-existent destination",
			sourceLines: []string{"line1", "line2"},
			change:      Change{Action: ActionInsert, Content: "line3", Above: "line4"},
			want:        nil,                                     // Depending on the behavior, this could return the original lines unmodified.
			wantErr:     ErrCodeBlock{source: []string{"line4"}}, // Or any other error indicating the destination does not exist.
		},
		{
			name:        "Insert in the middle",
			sourceLines: []string{"line1", "line3"},
			change:      Change{Action: ActionInsert, Content: "line2", Above: "line3"},
			want:        []string{"line1", "line2", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Insert in the middle (multiline)",
			sourceLines: []string{"line1", "line3"},
			change:      Change{Action: ActionInsert, Content: "line2\nfoobar\n", Above: "line3"},
			want:        []string{"line1", "line2", "foobar", "line3"},
			wantErr:     nil,
		},
		{
			name:        "Insert multiple lines",
			sourceLines: []string{"line1", "line4"},
			change:      Change{Action: ActionInsert, Content: "line2\nline3", Above: "line4"},
			want:        []string{"line1", "line2", "line3", "line4"},
			wantErr:     nil,
		},
		{
			name:        "Insert empty line",
			sourceLines: []string{"line1", "line2"},
			change:      Change{Action: ActionInsert, Content: "\n", Above: "line2"},
			want:        []string{"line1", "", "line2"},
			wantErr:     nil,
		},
		{
			name:        "Insert into empty source",
			sourceLines: []string{},
			change:      Change{Action: ActionPrepend, Content: "line1"},
			want:        []string{"line1"},
			wantErr:     nil,
		},
		{
			name:        "Insert with whitespace content",
			sourceLines: []string{"line1", "line2"},
			change:      Change{Action: ActionInsert, Content: "   ", Above: "line2"},
			want:        []string{"line1", "   ", "line2"},
			wantErr:     nil,
		},
		{
			name:        "Insert special characters",
			sourceLines: []string{"line1", "line2"},
			change:      Change{Action: ActionInsert, Content: "@#$%^", Above: "line2"},
			want:        []string{"line1", "@#$%^", "line2"},
			wantErr:     nil,
		},
		{
			name:        "Insert without specifying destination",
			sourceLines: []string{"line1", "line2"},
			change:      Change{Action: ActionInsert, Content: "line3"}, // No destination specified implies appending.
			want:        []string{"line1", "line2", "line3"},
			wantErr:     nil, // No error expected if the function appends by default.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDifferential(tt.sourceLines)
			err := d.ApplyChange(tt.change)

			// Error handling: Check if the expected error matches the actual error
			if (err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error()) || (err == nil && tt.wantErr != nil) || (err != nil && tt.wantErr == nil) {
				t.Errorf("ApplyChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// In case of no error, check if the output matches the expected result
			if err == nil {
				if !reflect.DeepEqual(d.SourceLines, tt.want) {
					t.Errorf("ApplyChange() = %v, want %v", d.SourceLines, tt.want)
				}
			}
		})
	}
}

// CorpusData represents the structure of data used in our fuzzing corpus.
type CorpusData struct {
	SourceLines []string
	Change      Change
}

func LoadCorpus(f *testing.F) {
	examples := []CorpusData{
		{
			SourceLines: []string{"line1", "line2", "line3"},
			Change: Change{
				Action:  ActionInsert,
				Content: "new content",
			},
		},
		{
			SourceLines: []string{"function test() {}", "let a = 1;"},
			Change: Change{
				Action: ActionReplace,
				Old:    "let a = 1;",
				New:    "let a = 2;",
			},
		},
		{
			SourceLines: []string{"function test() {}", "let a = 1;"},
			Change: Change{
				Action:  ActionDelete,
				Content: "let a = 1;",
			},
		},
		{
			SourceLines: []string{"function test() {}", "let a = 1;"},
			Change: Change{
				Action:  ActionMove,
				Content: "let a = 1;",
				Above:   "function test() {}",
			},
		},
		{
			SourceLines: []string{"function test() {}", "let a = 1;"},
			Change: Change{
				Action:  ActionAppend,
				Content: "new content\nnew content",
			},
		},
		{
			SourceLines: []string{"function test() {}", "let a = 1;"},
			Change: Change{
				Action:  ActionPrepend,
				Content: "new content\nnew content",
			},
		},
		{
			SourceLines: []string{"function test() {}", "let a = 1;"},
			Change: Change{
				Action:  ActionInsert,
				Content: "new content\nnew content",
			},
		},
		{
			SourceLines: []string{"function test() {}", "let a = 1;"},
			Change: Change{
				Action: ActionReplace,
				Old:    "let a = 1;",
				New:    "let a = 2;\nlet a = 3;",
			},
		},
		{
			SourceLines: []string{"function test() {}", "let a = 1;", "let a = 2;"},
			Change: Change{
				Action:  ActionDelete,
				Content: "let a = 1;\nlet a = 2;",
			},
		},
	}

	// Serialize and save the examples to the corpus directory.
	for _, example := range examples {
		f.Add(
			strings.Join(example.SourceLines, "\n"),
			string(example.Change.Action),
			example.Change.Old,
			example.Change.New,
			example.Change.Content,
			example.Change.Above)
	}
}

// FuzzApplyChange is the function that the fuzzer will execute.
func FuzzApplyChange(f *testing.F) {
	LoadCorpus(f)

	f.Fuzz(func(t *testing.T, source string, action string, old string, new_ string, content string, destinationAbove string) {
		change := Change{
			Comment: "Test change",
			Action:  Action(action),
			Old:     old,
			New:     new_,
			Content: content,
			Above:   destinationAbove,
		}

		sourceLines := strings.Split(source, "\n")
		d := NewDifferential(sourceLines)

		// We're not checking the result here, as we're not testing correctness.
		// We're testing that the function can handle a variety of inputs without crashing.
		err := d.ApplyChange(change)

		// If you expect certain kinds of inputs to produce errors (and know what those errors are),
		// you can handle them here.
		if err != nil {
			t.Skip() // We acknowledge the error but skip to allow the fuzzer to continue.
		}
	})
}
