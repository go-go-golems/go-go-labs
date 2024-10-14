package snakemake

import (
	"strings"
	"time"
)

// parseTime parses a time string into a time.Time object.
func parseTime(timeStr string) (time.Time, error) {
	const layout = "Mon Jan 2 15:04:05 2006"
	return time.Parse(layout, timeStr)
}

// parseResources parses a resource string and returns a slice of Resources.
func parseResources(resourcesStr string) []Resource {
	var resources []Resource
	parts := strings.Split(resourcesStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		name := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		resources = append(resources, Resource{
			Name:  name,
			Value: value,
		})
	}
	return resources
}

// mergeResources merges two slices of Resources, with new resources overriding existing ones.
func mergeResources(existing, new []Resource) []Resource {
	resourceMap := make(map[string]string)
	for _, r := range existing {
		resourceMap[r.Name] = r.Value
	}
	for _, r := range new {
		resourceMap[r.Name] = r.Value
	}
	var merged []Resource
	for name, value := range resourceMap {
		merged = append(merged, Resource{Name: name, Value: value})
	}
	return merged
}
