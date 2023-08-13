package ui

import "github.com/go-go-golems/go-go-labs/cmd/bandcamp/pkg"

type UpdateSearchResultsMsg struct {
	Results []*pkg.Result
}

type SelectEntryMsg struct {
	Result *pkg.Result
}

type ErrMsg struct {
	Err error
}

type ClearErrorMsg struct{}

type InsertPlaylistEntryMsg struct {
	Track *pkg.Track
}
