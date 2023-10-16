package ui

import (
	pkg2 "github.com/go-go-golems/go-go-labs/cmd/apps/bandcamp/pkg"
)

type UpdateSearchResultsMsg struct {
	Results []*pkg2.Result
}

type SelectEntryMsg struct {
	Result *pkg2.Result
}

type ErrMsg struct {
	Err error
}

type ClearErrorMsg struct{}

type InsertPlaylistEntryMsg struct {
	Track *pkg2.Track
}
