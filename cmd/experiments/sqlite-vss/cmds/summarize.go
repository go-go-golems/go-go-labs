package cmds

import (
	"embed"
	_ "embed"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	pinocchio_cmds "github.com/go-go-golems/pinocchio/pkg/cmds"
	"github.com/pkg/errors"
)

//go:embed prompts
var promptsFS embed.FS

type AnswerQuestionCommand struct {
}

func NewAnswerQuestionCommand() (*pinocchio_cmds.PinocchioCommand, error) {
	// TODO(manuel, 2024-04-26) We could add an easier way to load just one file
	g := &pinocchio_cmds.PinocchioCommandLoader{}
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

	return cmds_[0].(*pinocchio_cmds.PinocchioCommand), nil
}
