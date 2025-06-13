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
	"github.com/rs/zerolog/log"
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
	log.Debug().
		Str("template", tmpl).
		Msg("Starting query template rendering")

	// Log parameters (without values to avoid exposing sensitive data)
	paramKeys := make([]string, 0, len(params))
	for key := range params {
		paramKeys = append(paramKeys, key)
	}
	log.Debug().
		Strs("param_keys", paramKeys).
		Msg("Template parameters available")

	// Create template with Datadog helpers
	log.Debug().Msg("Creating template with Datadog helper functions")
	t, err := template.New("query").
		Funcs(DatadogTemplateFuncs).
		Option("missingkey=error"). // Strict mode - error on missing keys
		Parse(tmpl)
	if err != nil {
		log.Error().
			Err(err).
			Str("template", tmpl).
			Msg("Failed to parse query template")
		return types.DatadogQuery{}, errors.Wrap(err, "failed to parse query template")
	}
	log.Debug().Msg("Template parsed successfully")

	// Render the template
	log.Debug().Msg("Executing template with parameters")
	var buf bytes.Buffer
	err = t.Execute(&buf, params)
	if err != nil {
		log.Error().
			Err(err).
			Str("template", tmpl).
			Strs("param_keys", paramKeys).
			Msg("Failed to execute query template")
		return types.DatadogQuery{}, errors.Wrap(err, "failed to execute query template")
	}

	rawRendered := buf.String()
	log.Debug().
		Str("raw_rendered", rawRendered).
		Msg("Template executed successfully")

	// Clean up the rendered query
	renderedQuery := cleanupQuery(rawRendered)
	log.Debug().
		Str("raw_query", rawRendered).
		Str("cleaned_query", renderedQuery).
		Msg("Query cleaned up")

	// Extract time range from params
	log.Debug().Msg("Extracting time range from parameters")
	var from, to time.Time
	if fromParam, ok := params["from"]; ok {
		if fromTime, ok := fromParam.(time.Time); ok {
			from = fromTime
			log.Debug().Time("from", from).Msg("From time extracted from parameters")
		} else {
			log.Debug().
				Interface("from_param", fromParam).
				Msg("From parameter is not a time.Time")
		}
	}
	if toParam, ok := params["to"]; ok {
		if toTime, ok := toParam.(time.Time); ok {
			to = toTime
			log.Debug().Time("to", to).Msg("To time extracted from parameters")
		} else {
			log.Debug().
				Interface("to_param", toParam).
				Msg("To parameter is not a time.Time")
		}
	}

	// Extract other parameters
	log.Debug().Msg("Extracting other query parameters")
	var limit int
	if limitParam, ok := params["limit"]; ok {
		if limitInt, ok := limitParam.(int); ok {
			limit = limitInt
			log.Debug().Int("limit", limit).Msg("Limit extracted from parameters")
		} else {
			log.Debug().
				Interface("limit_param", limitParam).
				Msg("Limit parameter is not an int")
		}
	}

	var sort string
	if sortParam, ok := params["sort"]; ok {
		if sortStr, ok := sortParam.(string); ok {
			sort = sortStr
			log.Debug().Str("sort", sort).Msg("Sort extracted from parameters")
		} else {
			log.Debug().
				Interface("sort_param", sortParam).
				Msg("Sort parameter is not a string")
		}
	}

	query := types.DatadogQuery{
		Query: renderedQuery,
		From:  from,
		To:    to,
		Limit: limit,
		Sort:  sort,
	}

	log.Debug().
		Str("final_query", query.Query).
		Time("from", query.From).
		Time("to", query.To).
		Int("limit", query.Limit).
		Str("sort", query.Sort).
		Msg("Query rendering completed successfully")

	return query, nil
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
	log.Debug().
		Str("query", query).
		Msg("Validating rendered query")

	// Check for balanced quotes
	quoteCount := strings.Count(query, `"`) - strings.Count(query, `\"`)
	if quoteCount%2 != 0 {
		log.Error().
			Str("query", query).
			Int("quote_count", quoteCount).
			Msg("Query validation failed: unbalanced quotes")
		return errors.New("unbalanced quotes in query")
	}
	log.Debug().Int("quote_count", quoteCount).Msg("Quote balance check passed")

	// Check for basic syntax issues
	if strings.Contains(query, "::") {
		log.Error().
			Str("query", query).
			Msg("Query validation failed: invalid double colon")
		return errors.New("invalid double colon in query")
	}
	log.Debug().Msg("Double colon check passed")

	log.Debug().
		Str("query", query).
		Msg("Query validation completed successfully")
	return nil
}
