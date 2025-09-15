package action

// RuntimeEnv captures the GitHub runner environment values we rely on.
type RuntimeEnv struct {
	EventName       string
	EventPath       string
	Repository      string
	RunID           string
	Actor           string
	Workspace       string
	StepSummaryPath string
}

// LoadRuntimeEnv reads known GitHub environment keys through lookup.
func LoadRuntimeEnv(lookup func(string) string) RuntimeEnv {
	env := RuntimeEnv{
		EventName:       lookup("GITHUB_EVENT_NAME"),
		EventPath:       lookup("GITHUB_EVENT_PATH"),
		Repository:      lookup("GITHUB_REPOSITORY"),
		RunID:           lookup("GITHUB_RUN_ID"),
		Actor:           lookup("GITHUB_ACTOR"),
		Workspace:       lookup("GITHUB_WORKSPACE"),
		StepSummaryPath: lookup("GITHUB_STEP_SUMMARY"),
	}
	if env.Workspace == "" {
		env.Workspace = "/github/workspace"
	}
	return env
}
