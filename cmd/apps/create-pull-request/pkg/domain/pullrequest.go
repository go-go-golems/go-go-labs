package domain

// PullRequestSpec represents the main entity for a pull request
type PullRequestSpec struct {
	Title        string       `yaml:"title"`
	Body         string       `yaml:"body"`
	Changelog    string       `yaml:"changelog"`
	ReleaseNotes ReleaseNotes `yaml:"release_notes"`
}

// ReleaseNotes represents notes for a release
type ReleaseNotes struct {
	Title string `yaml:"title"`
	Body  string `yaml:"body"`
}