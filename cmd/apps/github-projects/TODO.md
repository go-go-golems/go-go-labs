- [ ] Fix updating project items: 

(from /tmp/github-projects.log, written by the MCP)

2025-06-21T18:58:17.632645056-04:00 ERR workspaces/2025-06-20/github-graphql-cli/go-go-labs/cmd/apps/github-projects/pkg/github/client.go:128 > GraphQL query failed error="graphql: The single select option Id does not belong to the field" exec_duration=136.547723 query="\n\t\tmutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $value: ProjectV2FieldValue!) {\n\t\t\tupdateProjectV2ItemFieldValue(input: {\n\t\t\t\tprojectId: $projectId\n\t\t\t\titemId: $itemId\n\t\t\t\tfieldId: $fieldId\n\t\t\t\tvalue: $value\n\t\t\t}) {\n\t\t\t\tprojectV2Item { id }\n\t\t\t}\n\t\t}\n\t" total_duration=136.688326 variables={"fieldId":"PVTSSF_lADOB23p8s4ALtcXzgHd8GY","itemId":"PVTI_lADOB23p8s4ALtcXzgbwFps","projectId":"PVT_kwDOB23p8s4ALtcX","value":{"singleSelectOptionId":"status_in-progress"}}
2025-06-21T18:58:17.632724921-04:00 DBG workspaces/2025-06-20/github-graphql-cli/go-go-labs/cmd/apps/github-projects/pkg/github/client.go:133 > Attempting error recovery error_type=graphql_execution
2025-06-21T18:58:17.63275349-04:00 ERR workspaces/2025-06-20/github-graphql-cli/go-go-labs/cmd/apps/github-projects/pkg/github/projects.go:685 > mutation execution failed error="GraphQL query failed: graphql: The single select option Id does not belong to the field" duration=136.911926 fieldID=PVTSSF_lADOB23p8s4ALtcXzgHd8GY function=UpdateFieldValue itemID=PVTI_lADOB23p8s4ALtcXzgbwFps projectID=PVT_kwDOB23p8s4ALtcX value={"singleSelectOptionId":"status_in-progress"}
2025-06-21T18:58:17.632780229-04:00 ERR workspaces/2025-06-20/github-graphql-cli/go-go-labs/cmd/apps/github-projects/mcp.go:465 > failed to update project item error="failed to update field Status: failed to update field value: GraphQL query failed: graphql: The single select option Id does not belong to the field" taskID=PVTI_lADOB23p8s4ALtcXzgbwFps
2025-06-21T18:58:17.632806391-04:00 DBG workspaces/2025-06-20/github-graphql-cli/go-go-mcp/pkg/transport/stdio/transport.go:150 > Sending response component=stdio_transport pid=399339 response={"id":61,"jsonrpc":"2.0","result":{"content":[{"text":"Failed to update project item: failed to update field Status: failed to update field value: GraphQL query failed: graphql: The single select option Id does not belong to the field","type":"text"}],"isError":true}} session_id=64e97b5b-4c0a-4f4b-bbd4-39c776016ddd

- [ ] Add searching by status/label for getting project items (in CLI and MCP)
- [ ] Add sorting + limit + date range for getting project items as well
- [ ] Updating allows setting labels
- [ ] Add verb and tool to add item comment
