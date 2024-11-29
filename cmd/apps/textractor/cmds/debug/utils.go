package debug

import (
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

// runAWSCommand executes an AWS CLI command with the given arguments
func runAWSCommand(args ...string) error {
	cmd := exec.Command("aws", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

// LoadResources loads the Textractor resources using the state loader
func LoadResources(cmd *cobra.Command) (*pkg.TextractorResources, error) {
	stateLoader := pkg.NewStateLoader()
	return stateLoader.LoadStateFromCommand(cmd)
}
