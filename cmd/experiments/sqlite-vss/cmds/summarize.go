package cmds

import (
	"embed"
	_ "embed"
	geppetto_cmds "github.com/go-go-golems/geppetto/pkg/cmds"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/pkg/errors"
)

//go:embed prompts/answer-question.yaml
var answerQuestionYaml string

//go:embed prompts
var promptsFS embed.FS

type AnswerQuestionCommand struct {
}

func NewAnswerQuestionCommand() (*geppetto_cmds.GeppettoCommand, error) {
	g := &geppetto_cmds.GeppettoCommandLoader{}
	cmds_, err := g.LoadCommands(
		promptsFS,
		"prompts/answer-question.yaml",
		[]glazed_cmds.CommandDescriptionOption{},
		[]alias.Option{},
	)

	if err != nil {
		return nil, err
	}

	if len(cmds_) != 1 {
		return nil, errors.New("expected exactly one command")
	}

	return cmds_[0].(*geppetto_cmds.GeppettoCommand), nil

}
