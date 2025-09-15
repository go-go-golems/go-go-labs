package action

import "strings"

// ShouldTrigger evaluates trigger inputs against the PR context.
func ShouldTrigger(in *Inputs, pr *PRContext) bool {
	if in == nil || pr == nil {
		return false
	}

	if phrase := strings.TrimSpace(in.TriggerPhrase); phrase != "" {
		phrase = strings.ToLower(phrase)
		haystack := []string{pr.TriggerText, pr.Body, pr.Title}
		matched := false
		for _, candidate := range haystack {
			if strings.Contains(strings.ToLower(candidate), phrase) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if label := strings.TrimSpace(in.LabelTrigger); label != "" {
		matched := false
		for _, l := range pr.Labels {
			if strings.EqualFold(l, label) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if assignee := strings.TrimSpace(in.AssigneeTrigger); assignee != "" {
		matched := false
		for _, a := range pr.Assignees {
			if strings.EqualFold(a, assignee) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}
