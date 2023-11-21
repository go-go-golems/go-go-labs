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

type ErrCodeBlock struct {
	source []string
}

func (e ErrCodeBlock) Is(e_ error) bool {
	// equal if ErroCodeBlock and lines are equal
	if e2, ok := e_.(ErrCodeBlock); ok {
		return strings.Join(e.source, "\n") == strings.Join(e2.source, "\n")
	}

	return false
}

func (e ErrCodeBlock) Error() string {
	return fmt.Sprintf("specified code block not found in the source: %s", strings.Join(e.source, "\\n"))
}

type ErrInvalidChange struct {
	msg string
}

func (e *ErrInvalidChange) Error() string {
	return fmt.Sprintf("invalid change: %s", e.msg)
}

type Differential struct {
	SourceLines         []string
	sourceLinesStripped []string
}

func NewDifferential(sourceLines []string) *Differential {
	// Strip the source lines of any leading or trailing whitespace
	sourceLinesStripped := make([]string, len(sourceLines))
	for i, line := range sourceLines {
		sourceLinesStripped[i] = strings.TrimSpace(line)
	}
	return &Differential{
		SourceLines:         sourceLines,
		sourceLinesStripped: sourceLinesStripped,
	}
}

// FindLocation is a function that identifies the position of a specific block of
// code within a given source code. It takes two parameters: SourceLines and
// locationLines and uses KMPSearch to find the matching index.
//
// The function returns two values: the line number (or -1 if not found), and an error
// if the string was not found.
func (d *Differential) FindLocation(locationLines []string) (int, error) {

	if len(locationLines) == 0 {
		return -1, ErrCodeBlock{
			source: locationLines,
		}
	}

	strippedLocationLines := make([]string, len(locationLines))
	for i, line := range locationLines {
		strippedLocationLines[i] = strings.TrimSpace(line)
	}
	l := kmp.KMPSearch(d.sourceLinesStripped, strippedLocationLines)
	if l == -1 {
		return -1, ErrCodeBlock{
			source: locationLines,
		}
	}

	return l, nil
}

func (d *Differential) SetSourceLines(sourceLines []string) {
	d.SourceLines = sourceLines
	d.sourceLinesStripped = make([]string, len(sourceLines))
	for i, line := range sourceLines {
		d.sourceLinesStripped[i] = strings.TrimSpace(line)
	}
}

// ApplyChange applies a specified change to a given set of source lines.
//
// It takes two parameters:
// - SourceLines: A slice of strings representing the source lines to be modified.
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
func (d *Differential) ApplyChange(change Change) error {
	switch change.Action {
	case ActionReplace, ActionDelete, ActionMove:
		contentLines := strings.Split(change.Old, "\n")
		if change.Action != ActionReplace {
			contentLines = strings.Split(change.Content, "\n")
		}
		startIdx, err := d.FindLocation(contentLines)
		if err != nil {
			return err
		}
		if startIdx == -1 {
			return ErrCodeBlock{
				source: contentLines,
			}
		}
		endIdx := startIdx + len(contentLines)

		if change.Action == ActionReplace {
			newLines := strings.Split(change.New, "\n")
			d.SetSourceLines(append(d.SourceLines[:startIdx], append(newLines, d.SourceLines[endIdx:]...)...))
		} else if change.Action == ActionDelete {
			d.SetSourceLines(append(d.SourceLines[:startIdx], d.SourceLines[endIdx:]...))
		} else if change.Action == ActionMove {
			destination := change.Above
			destLines := strings.Split(destination, "\n")
			segment := make([]string, endIdx-startIdx)
			copy(segment, d.SourceLines[startIdx:endIdx])

			d.SetSourceLines(append(d.SourceLines[:startIdx], d.SourceLines[endIdx:]...))

			moveIdx, err := d.FindLocation(destLines)
			if err != nil {
				return err
			}

			if len(d.SourceLines) < moveIdx {
				d.SetSourceLines(append(d.SourceLines, segment...))
			} else {
				d.SetSourceLines(append(d.SourceLines[:moveIdx], append(segment, d.SourceLines[moveIdx:]...)...))
			}
		}

	case ActionInsert:
		contentLines := strings.Split(change.Content, "\n")
		// remove last empty line
		if len(contentLines) > 0 && contentLines[len(contentLines)-1] == "" {
			contentLines = contentLines[:len(contentLines)-1]
		}
		if len(d.SourceLines) == 0 {
			d.SetSourceLines(append(d.SourceLines, contentLines...))
			break
		}
		destination := strings.TrimSuffix(change.Above, "\n")

		if destination == "" {
			d.SetSourceLines(append(d.SourceLines, contentLines...))
			break
		}
		destLines := strings.Split(destination, "\n")
		insertIdx, err := d.FindLocation(destLines)
		if err != nil {
			return err
		}
		d.SetSourceLines(append(d.SourceLines[:insertIdx], append(contentLines, d.SourceLines[insertIdx:]...)...))

	case ActionPrepend:
		contentLines := strings.Split(change.Content, "\n")
		d.SetSourceLines(append(contentLines, d.SourceLines...))

	case ActionAppend:
		contentLines := strings.Split(change.Content, "\n")
		d.SetSourceLines(append(d.SourceLines, contentLines...))

	default:
		return &ErrInvalidChange{"Unsupported action"}
	}

	return nil
}
