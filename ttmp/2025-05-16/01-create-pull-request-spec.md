# create pull request bash

get_git_diff() {
prompto get git-diff.sh -- -e sorbet/ -b origin/main --no-package
}

create_pull_request_from_diff() {
local description="$1"
local diff_output="$2"

if [ -z "$description" ]; then
echo "Error: Description is required."
return 1
fi

echo "$diff_output" | pinocchio tokens count -

while true; do
read -k1 "REPLY?Proceed with creating the pull request? (y/n/v/e) "
echo
case $REPLY in
      [Yy]) break ;;
      [Nn]) 
        echo "Pull request creation aborted."
        return 1 ;;
      [Vv])
        echo "$diff_output" | ${PAGER:-less}
        continue ;;
      [Ee])
        temp_file=$(mktemp)
echo "$diff_output" > "$temp_file"
${EDITOR:-vim} "$temp_file"
diff_output=$(cat "$temp_file")
rm "$temp_file"
        echo "$diff_output" | pinocchio tokens count -
continue ;;
\*)
echo "Invalid option. Please choose y, n, v (view) or e (edit)"
continue ;;
esac
done

# Create the pull request

echo "$diff_output" | \
  pinocchio code create-pull-request --non-interactive --diff - --description "$description" | \
 md-extract | tee /tmp/pr.yaml

read -k1 "REPLY?Proceed with creating the pull request? (y/n/e) "
echo
echo "reply: $REPLY"
  if [[ $REPLY =~ ^[Ee]$ ]]; then
${EDITOR:-vim} /tmp/pr.yaml
  elif [[ ! $REPLY =~ ^[Yy]$ ]]; then
echo "Pull request creation aborted."
return 1
fi

gh pr create --title "$(yq .title < /tmp/pr.yaml)" --body "$(yq .body < /tmp/pr.yaml)"
}

create_pull_request_from_yaml() {
cat /tmp/pr.yaml
read -k1 "REPLY?Proceed with creating the pull request? (y/n/e) "
echo
echo "reply: $REPLY"
  if [[ $REPLY =~ ^[Ee]$ ]]; then
${EDITOR:-vim} /tmp/pr.yaml
  elif [[ ! $REPLY =~ ^[Yy]$ ]]; then
echo "Pull request creation aborted."
return 1
fi

gh pr create --title "$(yq .title < /tmp/pr.yaml)" --body "$(yq .body < /tmp/pr.yaml)"
}

create_pull_request() {
local description="$1"

git fetch origin

local diff_output
diff_output=$(get_git_diff)

create_pull_request_from_diff "$description" "$diff_output"
}

create_pull_request_from_clipboard() {
local description="$1"

local diff_output
diff_output=$(xsel -b)

create_pull_request_from_diff "$description" "$diff_output"
}

# prompt

name: create-pull-request
short: Generate comprehensive pull request descriptions
flags:

- name: commits
  type: stringFromFile
  help: File containing the commits history
  default: ""
- name: issue
  type: string
  help: File containing the issue description corresponding to this pull request
- name: description
  type: string
  help: Description of the pull request
  required: true
- name: title
  type: string
  help: Title of the pull request
  default: ""
- name: diff
  type: stringFromFile
  help: File containing the diff of the changes
- name: code
  type: fileList
  help: List of code files
  default: []
- name: additional_system
  type: string
  help: Additional system prompt
  default: ""
- name: additional
  type: stringList
  help: Additional prompt
  default: []
- name: context
  type: fileList
  help: Additional context from files
- name: concise
  type: bool
  help: Give concise answers
  default: false
- name: use_bullets
  type: bool
  help: Use bullet points in the answer
  default: false
- name: use_keywords
  type: bool
  help: Use keywords in the answer
  default: false
- name: bracket
  type: bool
  help: Use bracketed text in the answer
  default: true
- name: without_files
  type: bool
  help: Do not include files in the answer
  default: true
  system-prompt: |
  You are an experienced software engineer and technical leader.
  You are skilled at understanding and describing code changes, generating concise and informative titles,
  and crafting detailed pull request descriptions. You are adept at prompting for additional information when necessary.
  If not enough information is provided to create a good pull request,
  ask the user for additional clarifying information.
  Your ultimate goal is to create pull request descriptions that are clear, concise, and informative,
  facilitating the team's ability to review and merge the changes effectively.
  {{ .additional_system }}
  prompt: |
  {{ define "context" -}}
  {{ if .commits }}Begin by understanding and describing the commits as provided by the user to ensure you have accurately captured the changes. The commits are:
  --- BEGIN COMMITS
  {{ .commits }}
  --- END COMMITS{{end}}

{{ if .issue }}The issue corresponding to this pull request is: {{ .issue }}.{{ end }}

The description of the pull request is: {{ .description }}.

{{ if .title}}Now, generate a concise and informative title that accurately represents the changes and title. The title is: {{ .title }}.{{end}}

{{if .diff }}The diff of the changes is:
--- BEGIN DIFF
{{ .diff }}
--- END DIFF. {{ end }}

{{ if .code }}The code files are:
{{ range .code }}Path: {{ .Path }}
Content: {{ .Content }}
{{ end }}.{{end}}

Finally, craft a detailed pull request description that provides all the necessary information for reviewing the changes, using clear and understandable language.
If not enough information is provided to create a good pull request, ask the user for additional clarifying information.

{{ if .without_files }}Do not mention filenames unless it is very important.{{ end }}
Do not mention trivial changes like changed imports.

Be concise and use bullet point lists and keyword sentences.
No need to write much about how useful the feature will be, stay pragmatic.

Remember: use bullet points and keyword like sentences.
Don't use capitalized title case for the title.

Output the results as a YAML file with the following structure, wrapping the body at 80 characters.

```yaml
title: ...
body: |
  ...
changelog: |
  ... # A concise, single-line description of the main changes for the changelog
release_notes:
  title: ... # A user-friendly title for the release notes
  body: |
    ... # A more detailed description focusing on user-facing changes and benefits
```

For the changelog entry:

- Keep it short and focused on the main changes
- Use present tense (e.g., "Add feature X" not "Added feature X")
- Focus on technical changes

For the release notes:

- Title should be user-friendly and descriptive
- Body should explain the changes from a user's perspective
- Include any new features, improvements, or breaking changes
- Explain benefits and use cases where relevant

Capitalize the first letter of all titles.

{{ if .additional }}
Additional instructions:
{{ .additional | join "\n- " }}
{{ end }}

{{ if .concise -}} Give a concise answer, answer in a single sentence if possible, skip unnecessary explanations. {{- end }}
{{ if .use_bullets -}} Use bullet points in the answer. {{- end }}
{{ if .use_keywords -}} Use keywords in the answer, not full sentences. {{- end }}
{{- end }}

{{ template "context" . }}

{{ if .context}}Additional Context:
{{ range .context }}
Path: {{ .Path }}

---

{{ .Content }}

---

{{- end }}
{{ end }}

{{ if .bracket }}
{{ template "context" . }}
{{ end }}

---

# git gather script

#!/bin/bash

# Default values

branch="origin/main"
exclude_files=()
context_size="-U3"
include_paths=""
exclude_paths=""
exclude_package=false

# Function to display usage information

usage() {
echo "Usage: $0 [options]"
echo "Options:"
echo " -b, --branch BRANCH Specify a branch (default: origin/main)"
echo " -e, --exclude FILES Exclude specific files (comma-separated list)"
echo " -s, --short Reduce diff context size to 5 lines"
echo " -o, --only PATHS Include specific paths only (comma-separated list)"
echo " --no-package Exclude common package manager files (go.mod, go.sum, package.json, package-lock.json, etc.)"
exit 1
}

# Parse command-line arguments

while [[$# -gt 0]]; do
key="$1"
  case $key in
    -b|--branch)
      branch="$2"
      shift
      shift
      ;;
    -e|--exclude)
      exclude_files+=($(echo "$2" | tr ',' ' '))
shift
shift
;;
-l|--long)
context_size="-U10"
shift
;;
-s|--short)
context_size="-U1"
shift
;;
-o|--only)
include_paths="$2"
shift
shift
;;
--no-tests)
exclude_files+=("_.test.js" "_.test.ts" "_.spec.js" "_.spec.ts" "_\_test.go")
shift
;;
--no-package)
exclude_package=true
shift
;;
-h|--help)
usage
;;
_)
echo "Unknown option: $1"
usage
;;
esac
done

# Exclude common package manager files

if [ "$exclude_package" = true ]; then
exclude_files+=("go.sum" "go.work.sum" "package-lock.json" "yarn.lock" "composer.lock" "yarn.lock" "sorbet" "_.min.js" "_\_templ.go" "\*.rbi")
fi

# Construct the exclusion patterns

exclude_patterns=""
for file in "${exclude_files[@]}"; do
  exclude_patterns+=" :!**/$file"

# while we're printing, add quotes

# exclude_patterns+=" ':!$file'"

done

# Construct the inclusion patterns

include_patterns=""
if [ -n "$include_paths" ]; then
IFS=',' read -ra paths <<< "$include_paths"
  for path in "${paths[@]}"; do
include_patterns+=" :$path"
done
fi

cd "$PROMPTO_PARENT_PWD"

# Print git diff summary with --stat

echo git diff --stat "$branch" -- . $exclude_patterns $include_patterns
git diff --stat "$branch" -- . $exclude_patterns $include_patterns

# Run git diff command with diff-filter=d to exclude deleted files

echo git diff -w "$context_size" "$branch" --diff-filter=d -- . $exclude_patterns $include_patterns
cd "$PROMPTO_PARENT_PWD" && pwd && git diff -w "$context_size" "$branch" --diff-filter=d -- . $exclude_patterns $include_patterns
