package main

import (
	"context"
	"fmt"
	"log"

	"go.lsp.dev/protocol"
)

// MockClient implements protocol.Client interface and logs all requests
type MockClient struct {
	log *log.Logger
}

func NewMockClient() *MockClient {
	return &MockClient{
		log: log.Default(),
	}
}

func (c *MockClient) Progress(ctx context.Context, params *protocol.ProgressParams) error {
	c.log.Printf("Progress: %+v\n", params)
	return nil
}

func (c *MockClient) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) error {
	c.log.Printf("WorkDoneProgressCreate: %+v\n", params)
	return nil
}

func (c *MockClient) LogMessage(ctx context.Context, params *protocol.LogMessageParams) error {
	c.log.Printf("LogMessage: %+v\n", params)
	return nil
}

func (c *MockClient) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) error {
	c.log.Printf("PublishDiagnostics: %+v\n", params)
	return nil
}

func (c *MockClient) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) error {
	c.log.Printf("ShowMessage: %+v\n", params)
	return nil
}

func (c *MockClient) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	c.log.Printf("ShowMessageRequest: %+v\n", params)
	// Return first action if available, otherwise nil
	if len(params.Actions) > 0 {
		return &params.Actions[0], nil
	}
	return nil, nil
}

func (c *MockClient) Telemetry(ctx context.Context, params interface{}) error {
	c.log.Printf("Telemetry: %+v\n", params)
	return nil
}

func (c *MockClient) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) error {
	c.log.Printf("RegisterCapability: %+v\n", params)
	return nil
}

func (c *MockClient) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) error {
	c.log.Printf("UnregisterCapability: %+v\n", params)
	return nil
}

func (c *MockClient) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (bool, error) {
	c.log.Printf("ApplyEdit: %+v\n", params)
	// Always return true to indicate success
	return true, nil
}

func (c *MockClient) Configuration(ctx context.Context, params *protocol.ConfigurationParams) ([]interface{}, error) {
	c.log.Printf("Configuration: %+v\n", params)
	// Return empty configs for each requested item
	configs := make([]interface{}, len(params.Items))
	return configs, nil
}

func (c *MockClient) WorkspaceFolders(ctx context.Context) ([]protocol.WorkspaceFolder, error) {
	c.log.Printf("WorkspaceFolders requested\n")
	// Return the same workspace folder we used in initialization
	return []protocol.WorkspaceFolder{
		{
			URI:  "file:///Users/manuel/code/wesen/corporate-headquarters/go-go-labs/",
			Name: "go-go-labs",
		},
	}, nil
}

// Request handles any custom requests that might be sent to the client
func (c *MockClient) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	c.log.Printf("Unknown request method '%s' with params: %+v\n", method, params)
	return nil, fmt.Errorf("method '%s' not implemented", method)
}
