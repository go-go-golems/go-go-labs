package render

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/types"
	"github.com/pkg/errors"
)

// DatadogTemplateFuncs provides template helpers for Datadog search queries
var DatadogTemplateFuncs = template.FuncMap{
	"ddDateTime": func(t time.Time) string {
		return t.Format(time.RFC3339)
	},
	"ddStringIn": func(values []string) string {
		if len(values) == 0 {
			return ""
		}
		quoted := make([]string, len(values))
		for i, v := range values {
			// Escape quotes in values
			escaped := strings.ReplaceAll(v, `"`, `\"`)
			quoted[i] = `"` + escaped + `"`
		}
		return strings.Join(quoted, ",")
	},
	"ddLike": func(pattern string) string {
		// Escape special characters for Datadog search
		escaped := strings.ReplaceAll(pattern, `"`, `\"`)
		return `"` + escaped + `"`
	},
	"ddFacet": func(field string) string {
		// Ensure field name is properly formatted for faceting
		if !strings.HasPrefix(field, "@") && !strings.Contains(field, ".") {
			return "@" + field
		}
		return field
	},
}

// RenderDatadogQuery renders a query template with the given parameters
func RenderDatadogQuery(ctx context.Context, tmpl string, params map[string]interface{}) (types.DatadogQuery, error) {
	// Create template with Datadog helpers
	t, err := template.New("query").
		Funcs(DatadogTemplateFuncs).
		Option("missingkey=error"). // Strict mode - error on missing keys
		Parse(tmpl)
	if err != nil {
		return types.DatadogQuery{}, errors.Wrap(err, "failed to parse query template")
	}

	// Render the template
	var buf bytes.Buffer
	err = t.Execute(&buf, params)
	if err != nil {
		return types.DatadogQuery{}, errors.Wrap(err, "failed to execute query template")
	}

	// Clean up the rendered query
	renderedQuery := cleanupQuery(buf.String())

	// Extract time range from params
	var from, to time.Time
	if fromParam, ok := params["from"]; ok {
		if fromTime, ok := fromParam.(time.Time); ok {
			from = fromTime
		}
	}
	if toParam, ok := params["to"]; ok {
		if toTime, ok := toParam.(time.Time); ok {
			to = toTime
		}
	}

	// Extract other parameters
	var limit int
	if limitParam, ok := params["limit"]; ok {
		if limitInt, ok := limitParam.(int); ok {
			limit = limitInt
		}
	}

	var sort string
	if sortParam, ok := params["sort"]; ok {
		if sortStr, ok := sortParam.(string); ok {
			sort = sortStr
		}
	}

	return types.DatadogQuery{
		Query: renderedQuery,
		From:  from,
		To:    to,
		Limit: limit,
		Sort:  sort,
	}, nil
}

// cleanupQuery removes extra whitespace and normalizes the query string
func cleanupQuery(query string) string {
	// Remove leading/trailing whitespace
	query = strings.TrimSpace(query)

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	query = spaceRegex.ReplaceAllString(query, " ")

	// Remove spaces around colons for field:value patterns
	colonRegex := regexp.MustCompile(`\s*:\s*`)
	query = colonRegex.ReplaceAllString(query, ":")

	return query
}

// ValidateQuery performs basic validation on the rendered query
func ValidateQuery(query string) error {
	// Check for balanced quotes
	quoteCount := strings.Count(query, `"`) - strings.Count(query, `\"`)
	if quoteCount%2 != 0 {
		return errors.New("unbalanced quotes in query")
	}

	// Check for basic syntax issues
	if strings.Contains(query, "::") {
		return errors.New("invalid double colon in query")
	}

	return nil
}
