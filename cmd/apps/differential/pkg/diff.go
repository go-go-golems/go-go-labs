package pkg

import (
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/apps/differential/kmp"
	"strings"
	"text/template"
)

// Change represents a single change in the DSL.
type Change struct {
	Comment string `json:"comment" yaml:"comment"`
	Action  Action `json:"action" yaml:"action"`
	Old     string `json:"old,omitempty" yaml:"old,omitempty"`
	New     string `json:"new,omitempty" yaml:"new,omitempty"`
	Content string `json:"content,omitempty" yaml:"content,omitempty"`
	Above   string `json:"above,omitempty" yaml:"above,omitempty"`
}

const changeStringTemplate = `
{{ .Action }}: {{ .Comment }}

{{ if .Old }}Old code:
{{.Old}}
{{- end }}

{{ if .New }}New code:
{{.New}}
{{- end }}

{{ if .Content }}Content:
{{.Content}}
{{- end }}

{{ if .DestinationAbove }}Destination above:
{{.DestinationAbove}}
{{- end }}

{{ if .DestinationBelow }}Destination below:
{{.DestinationBelow}}
{{- end }}
`

func (c *Change) String() string {
	tpl := template.Must(template.New("change").Parse(changeStringTemplate))
	var sb strings.Builder
	err := tpl.Execute(&sb, c)
	if err != nil {
		return fmt.Sprintf("%s: %s", c.Action, c.Comment)
	}
	return sb.String()
}

// DSL represents the entire DSL document.
type DSL struct {
	Path    string   `json:"path" yaml:"path"`
	Changes []Change `json:"changes" yaml:"changes"`
}

type Action string

const (
	ActionReplace Action = "replace"
	ActionDelete  Action = "delete"
	ActionMove    Action = "move"
	ActionInsert  Action = "insert"
	ActionPrepend Action = "prepend"
	ActionAppend  Action = "append"
)

type ErrCodeBlock struct{}

func (e *ErrCodeBlock) Error() string {
	return "specified code block not found in the source"
}

type ErrInvalidChange struct {
	msg string
}

func (e *ErrInvalidChange) Error() string {
	return fmt.Sprintf("invalid change: %s", e.msg)
}

// FindLocation is a function that identifies the position of a specific block of
// code within a given source code. It takes two parameters: sourceLines and
// locationLines and uses KMPSearch to find the matching index.
//
// The function returns two values: the line number (or -1 if not found), and an error
// if the string was not found.
func FindLocation(sourceLines []string, locationLines []string) (int, error) {
	if len(locationLines) == 0 {
		return -1, &ErrCodeBlock{}
	}

	l := kmp.KMPSearch(sourceLines, locationLines)
	if l == -1 {
		return -1, &ErrCodeBlock{}
	}

	return l, nil
}

// ApplyChange applies a specified change to a given set of source lines.
//
// It takes two parameters:
// - sourceLines: A slice of strings representing the source lines to be modified.
// - change: A Change struct detailing the change to be applied.
//
// The function supports four types of actions specified in the Change struct:
// - ActionReplace: Replaces the old content with the new content in the source lines.
// - ActionDelete: Removes the old content from the source lines.
// - ActionMove: Moves the old content to a new location in the source lines.
// - ActionInsert: Inserts new content at a specified location in the source lines.
//
// The function returns a slice of strings representing the modified source lines,
// and an error if the action is unsupported or if there is an issue locating the
// content or destination in the source lines.
func ApplyChange(sourceLines []string, change Change) ([]string, error) {
	switch change.Action {
	case ActionReplace, ActionDelete, ActionMove:
		contentLines := strings.Split(change.Old, "\n")
		if change.Action != ActionReplace {
			contentLines = strings.Split(change.Content, "\n")
		}
		startIdx := kmp.KMPSearch(sourceLines, contentLines)
		if startIdx == -1 {
			return nil, &ErrCodeBlock{}
		}
		endIdx := startIdx + len(contentLines)

		if change.Action == ActionReplace {
			newLines := strings.Split(change.New, "\n")
			sourceLines = append(sourceLines[:startIdx], append(newLines, sourceLines[endIdx:]...)...)
		} else if change.Action == ActionDelete {
			sourceLines = append(sourceLines[:startIdx], sourceLines[endIdx:]...)
		} else if change.Action == ActionMove {
			destination := change.Above
			destLines := strings.Split(destination, "\n")
			segment := make([]string, endIdx-startIdx)
			copy(segment, sourceLines[startIdx:endIdx])

			sourceLines = append(sourceLines[:startIdx], sourceLines[endIdx:]...)

			moveIdx, err := FindLocation(sourceLines, destLines)
			if err != nil {
				return nil, err
			}

			if len(sourceLines) < moveIdx {
				sourceLines = append(sourceLines, segment...)
			} else {
				sourceLines = append(sourceLines[:moveIdx], append(segment, sourceLines[moveIdx:]...)...)
			}
		}

	case ActionInsert:
		contentLines := strings.Split(change.Content, "\n")
		if len(sourceLines) == 0 {
			sourceLines = append(sourceLines, contentLines...)
			break
		}
		destination := change.Above

		if destination == "" {
			sourceLines = append(sourceLines, contentLines...)
			break
		}
		destLines := strings.Split(destination, "\n")
		insertIdx, err := FindLocation(sourceLines, destLines)
		if err != nil {
			return nil, err
		}
		sourceLines = append(sourceLines[:insertIdx], append(contentLines, sourceLines[insertIdx:]...)...)

	case ActionPrepend:
		contentLines := strings.Split(change.Content, "\n")
		return append(contentLines, sourceLines...), nil

	case ActionAppend:
		contentLines := strings.Split(change.Content, "\n")
		return append(sourceLines, contentLines...), nil

	default:
		return nil, &ErrInvalidChange{"Unsupported action"}
	}

	return sourceLines, nil
}
