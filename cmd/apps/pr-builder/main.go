package main

import (
	"bytes"
	"fmt"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func main() {
	// Open the current repository, assuming we're in the root of it.
	r, err := git.PlainOpen(".")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Find the latest commit for the "origin/main" branch (the "remote" and "branch" might need adjusting depending on your settings).
	ref, err := r.Reference(plumbing.ReferenceName("refs/remotes/origin/main"), true)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the commit object.
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the Tree from the commit.
	tree, err := commit.Tree()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Diff the working directory and the given tree.
	head, _ := r.Head()
	headCommit, _ := r.CommitObject(head.Hash())
	headTree, _ := headCommit.Tree()
	changes, err := tree.Diff(headTree)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Prepare a map to hold the diff output for each file.
	fileDiffs := make(map[string]string)

	// Iterate through changes and populate the fileDiffs map.
	for _, change := range changes {
		// Generate patch for each change.
		patch, err := change.Patch()
		if err != nil {
			fmt.Println(err)
			return
		}

		var diffBuffer bytes.Buffer
		err = patch.Encode(&diffBuffer)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Add the diff output to the map.
		fileDiffs[change.From.Name] = diffBuffer.String()
	}

	// Print the diff result.
	for file, diff := range fileDiffs {
		_, _ = fmt.Fprintf(os.Stdout, "Changes to %s:\n%s\n", file, diff)
	}
}
