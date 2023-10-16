package pkg

import "strings"

type Parameter struct {
	Type  string       `json:"type"`
	Key   string       `json:"key"`
	Value string       `json:"value"`
	List  []*Parameter `json:"list"`
	Map   []*Parameter `json:"map"`
}

type Filter struct {
	Type      string       `json:"type"`
	Parameter []*Parameter `json:"parameter"`
}

type Variable struct {
	AccountID   string       `json:"accountId"`
	ContainerID string       `json:"containerId"`
	VariableID  string       `json:"variableId"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Parameter   []*Parameter `json:"parameter"`
	Fingerprint string       `json:"fingerprint"`
}

type Tag struct {
	AccountID       string       `json:"accountId"`
	ContainerID     string       `json:"containerId"`
	TagID           string       `json:"tagId"`
	Name            string       `json:"name"`
	Type            string       `json:"type"`
	Fingerprint     string       `json:"fingerprint"`
	TagFiringOption string       `json:"tagFiringOption"`
	Parameter       []*Parameter `json:"parameter"`
}

type Trigger struct {
	AccountID         string    `json:"accountId"`
	ContainerID       string    `json:"containerId"`
	TriggerID         string    `json:"triggerId"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	Fingerprint       string    `json:"fingerprint"`
	CustomEventFilter []*Filter `json:"customEventFilter"`
	Filter            []*Filter `json:"filter"`
}

type GTMExport struct {
	ContainerVersion struct {
		Variable []*Variable `json:"variable"`
		Tag      []*Tag      `json:"tag"`
		Trigger  []*Trigger  `json:"trigger"`
	} `json:"containerVersion"`
}

func FilterToString(filter *Filter) string {
	var sb strings.Builder

	// Check for negate parameter first
	for _, param := range filter.Parameter {
		if param.Key == "negate" && param.Value == "true" {
			sb.WriteString("!")
			break
		}
	}

	if filter.Type == "EQUALS" {
		for _, param := range filter.Parameter {
			if param.Key == "negate" {
				continue
			}
			if !strings.HasPrefix(param.Key, "arg") {
				sb.WriteString(param.Key)
				sb.WriteString(": ")
				sb.WriteString(param.Value)
			} else if strings.HasPrefix(param.Key, "arg") {
				sb.WriteString(param.Value)
				if param != filter.Parameter[len(filter.Parameter)-1] {
					sb.WriteString(" = ")
				}
			}
		}
	} else {
		sb.WriteString(filter.Type)
		sb.WriteString("(")
		for i, param := range filter.Parameter {
			if param.Key == "negate" {
				continue
			}
			if i > 0 {
				sb.WriteString(", ")
			}
			if !strings.HasPrefix(param.Key, "arg") {
				sb.WriteString(param.Key)
				sb.WriteString(": ")
			}
			sb.WriteString(param.Value)
		}
		sb.WriteString(")")
	}

	return sb.String()
}
