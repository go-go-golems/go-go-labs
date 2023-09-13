package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "promptgen",
		Short: "Promptgen generates prompts from a list of repositories",
		Long: `This program loads a list of repositories from a yaml config file
and looks for a file that matches the prompt.`,
	}

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(listCmd)

	viper.SetConfigName("config")
	viper.AddConfigPath(os.ExpandEnv("$HOME/.promptgen"))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var getCmd = &cobra.Command{
	Use:   "get prompt [args]",
	Short: "Get and execute or print a file",
	Args:  cobra.MinimumNArgs(1),
	RunE:  get,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all prompts in the repositories",
	RunE:  list,
}

func get(cmd *cobra.Command, args []string) error {
	prompt := args[0]
	restArgs := args[1:]

	repositories := viper.GetStringSlice("repositories")

	for _, repo := range repositories {
		if _, err := os.Stat(filepath.Join(repo, "promptgen")); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(filepath.Join(repo, "promptgen"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// strip the repo path from the file name
			relpath := path[len(repo):]
			name := strings.TrimPrefix(relpath, "/promptgen/")

			// compare prompt with the stripped filename
			if !info.IsDir() && name == prompt {
				if (info.Mode() & 0111) != 0 {
					// The file is executable, execute it
					c := exec.Command(path, restArgs...)
					c.Dir = repo
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					return c.Run()
				} else {
					// The file is not executable, print it
					fmt.Println(path)
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func list(cmd *cobra.Command, args []string) error {
	repositories := viper.GetStringSlice("repositories")

	for _, repo := range repositories {
		// skip if directory doesn't exist
		if _, err := os.Stat(filepath.Join(repo, "promptgen")); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(filepath.Join(repo, "promptgen"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// strip the repo path from the file name
			relpath := path[len(repo):]
			name := strings.TrimPrefix(relpath, "/promptgen/")

			if !info.IsDir() {
				fmt.Println(repo, name)
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
