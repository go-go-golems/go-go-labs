package main

import (
	"fmt"
	"strings"
)

// WorkflowToMermaidResult holds the result of the workflow to mermaid conversion
type WorkflowToMermaidResult struct {
	Notes       []string // Sticky notes as markdown
	MermaidCode string   // The mermaid diagram code
}

// WorkflowToMermaid converts a workflow to a mermaid diagram and extracts sticky notes
func WorkflowToMermaid(workflow map[string]interface{}) WorkflowToMermaidResult {
	var sb strings.Builder

	sb.WriteString("graph TD\n")

	// Get nodes and connections
	nodes, nodesOk := workflow["nodes"].([]interface{})
	connections, connectionsOk := workflow["connections"].(map[string]interface{})

	var stickyNotes []string
	if !nodesOk || !connectionsOk {
		return WorkflowToMermaidResult{
			MermaidCode: "graph TD\n  [No nodes or connections found in workflow]\n",
			Notes:       []string{},
		}
	}

	// Build node map for quick lookups and collect sticky notes
	nodeMap := make(map[string]map[string]interface{})
	for _, node := range nodes {
		nodeObj, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		nodeName, nameOk := nodeObj["name"].(string)
		_, idOk := nodeObj["id"].(string)
		nodeType, typeOk := nodeObj["type"].(string)

		// Check if this is a sticky note
		if typeOk && nodeType == "n8n-nodes-base.stickyNote" {
			// Extract sticky note content
			if parameters, ok := nodeObj["parameters"].(map[string]interface{}); ok {
				if content, ok := parameters["content"].(string); ok {
					// Add to sticky notes collection
					stickyNotes = append(stickyNotes, content)
				}
			}
			// Skip adding sticky notes to the node map
			continue
		}

		if nameOk && idOk {
			nodeMap[nodeName] = nodeObj
		}
	}

	// Generate connections
	for sourceName, conn := range connections {
		connObj, ok := conn.(map[string]interface{})
		if !ok {
			continue
		}

		main, ok := connObj["main"].([]interface{})
		if !ok {
			continue
		}

		// Process each output
		for _, outputConnections := range main {
			conns, ok := outputConnections.([]interface{})
			if !ok {
				continue
			}

			// Process each connection from this output
			for _, conn := range conns {
				connDetail, ok := conn.(map[string]interface{})
				if !ok {
					continue
				}

				targetName, ok := connDetail["node"].(string)
				if !ok {
					continue
				}

				// Format node names for mermaid (escape special chars and add descriptive text)
				sourceId := sanitizeNodeName(sourceName)
				targetId := sanitizeNodeName(targetName)

				// Add the connection
				sb.WriteString(fmt.Sprintf("  %s[\"%s\"] --> %s[\"%s\"]\n",
					sourceId, escapeQuotes(sourceName), targetId, escapeQuotes(targetName)))
			}
		}

		// Handle other connection types (like ai_document, ai_languageModel, etc.)
		for connType, outputs := range connObj {
			// Skip 'main' as we already processed it
			if connType == "main" {
				continue
			}

			outputsList, ok := outputs.([]interface{})
			if !ok {
				continue
			}

			for _, outputConnections := range outputsList {
				conns, ok := outputConnections.([]interface{})
				if !ok {
					continue
				}

				for _, conn := range conns {
					connDetail, ok := conn.(map[string]interface{})
					if !ok {
						continue
					}

					targetName, ok := connDetail["node"].(string)
					if !ok {
						continue
					}

					// Format node names for mermaid
					sourceId := sanitizeNodeName(sourceName)
					targetId := sanitizeNodeName(targetName)

					// Add the connection with connection type
					sb.WriteString(fmt.Sprintf("  %s[\"%s\"] -.%s.-> %s[\"%s\"]\n",
						sourceId, escapeQuotes(sourceName), connType, targetId, escapeQuotes(targetName)))
				}
			}
		}
	}

	// Add any nodes that have no connections
	for _, node := range nodes {
		nodeObj, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		nodeName, ok := nodeObj["name"].(string)
		if !ok {
			continue
		}

		// Check if this node appears in connections
		found := false
		for sourceName := range connections {
			if sourceName == nodeName {
				found = true
				break
			}
		}

		// If not found as a source, check if it's a target
		if !found {
			for _, conn := range connections {
				connObj, ok := conn.(map[string]interface{})
				if !ok {
					continue
				}

				main, ok := connObj["main"].([]interface{})
				if !ok {
					continue
				}

				for _, outputConnections := range main {
					conns, ok := outputConnections.([]interface{})
					if !ok {
						continue
					}

					for _, conn := range conns {
						connDetail, ok := conn.(map[string]interface{})
						if !ok {
							continue
						}

						targetName, ok := connDetail["node"].(string)
						if !ok {
							continue
						}

						if targetName == nodeName {
							found = true
							break
						}
					}

					if found {
						break
					}
				}

				if found {
					break
				}
			}
		}

		// If still not found in any connection, add it as standalone node
		if !found {
			nodeId := sanitizeNodeName(nodeName)
			sb.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", nodeId, escapeQuotes(nodeName)))
		}
	}

	return WorkflowToMermaidResult{
		MermaidCode: sb.String(),
		Notes:       stickyNotes,
	}
}

// escapeQuotes escapes quotes in strings for mermaid labels
func escapeQuotes(s string) string {
	// Replace quotes with escaped quotes
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// sanitizeNodeName ensures the node name is valid for mermaid
func sanitizeNodeName(name string) string {
	// Replace spaces and special characters with underscores
	sanitized := strings.ReplaceAll(name, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, "\"", "")
	sanitized = strings.ReplaceAll(sanitized, "'", "")
	sanitized = strings.ReplaceAll(sanitized, "(", "")
	sanitized = strings.ReplaceAll(sanitized, ")", "")
	sanitized = strings.ReplaceAll(sanitized, "[", "")
	sanitized = strings.ReplaceAll(sanitized, "]", "")
	sanitized = strings.ReplaceAll(sanitized, "{", "")
	sanitized = strings.ReplaceAll(sanitized, "}", "")
	sanitized = strings.ReplaceAll(sanitized, "<", "")
	sanitized = strings.ReplaceAll(sanitized, ">", "")
	sanitized = strings.ReplaceAll(sanitized, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, ";", "_")
	sanitized = strings.ReplaceAll(sanitized, ",", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")

	// Ensure it starts with a letter (mermaid requirement)
	if len(sanitized) > 0 && !strings.ContainsAny(sanitized[:1], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		sanitized = "n_" + sanitized
	}

	return sanitized
}
