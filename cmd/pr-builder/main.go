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

	// Get the worktree.
	w, err := r.Worktree()
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
	changes, err := w.Diff(&git.DiffOptions{PathFilter: func(path string) bool {
		// Optional: Use this filter to include/exclude files based on path.
		return true // Returning true includes all files.
	}})
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
		if _, err := patch.WriteTo(&diffBuffer); err != nil {
			fmt.Println(err)
			return
		}

		// Add the diff output to the map.
		fileDiffs[change.From.Name] = diffBuffer.String()

		// Free up the patch resources (recommended by the library).
		patch.Free()
	}

	// Print the diff result.
	for file, diff := range fileDiffs {
		fmt.Fprintf(os.Stdout, "Changes to %s:\n%s\n", file, diff)
	}
}
