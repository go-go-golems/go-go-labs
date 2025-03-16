package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

func main() {
	ctx := context.Background()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Connect to the running gopls server
	// make a io.ReadWriteCloser for the tcp connection
	tcpConn, err := net.Dial("tcp", "localhost:4389")
	if err != nil {
		log.Fatalf("Failed to connect to language server: %v", err)
	}
	defer tcpConn.Close()

	stream := jsonrpc2.NewStream(tcpConn)
	conn := jsonrpc2.NewConn(stream)

	ctx, cancel := context.WithCancel(ctx)

	server := protocol.ServerDispatcher(conn, logger)
	mockClient := NewMockClient()
	// client := protocol.ClientDispatcher(conn, logger)

	conn.Go(ctx, protocol.ClientHandler(mockClient, jsonrpc2.MethodNotFoundHandler))

	eg := errgroup.Group{}
	eg.Go(func() error {
		defer cancel()
		<-conn.Done()
		return nil
	})

	eg.Go(func() error {
		defer cancel()
		defer conn.Close()
		// Initialize the LSP session
		initParams := &protocol.InitializeParams{
			ClientInfo: &protocol.ClientInfo{
				Name:    "hover-client",
				Version: "1.0",
			},
			RootURI: protocol.DocumentURI("file:///Users/manuel/code/wesen/corporate-headquarters/go-go-labs/"),
			WorkspaceFolders: []protocol.WorkspaceFolder{
				{
					URI:  "file:///Users/manuel/code/wesen/corporate-headquarters/go-go-labs/",
					Name: "go-go-labs",
				},
			},
			Capabilities: protocol.ClientCapabilities{
				TextDocument: &protocol.TextDocumentClientCapabilities{
					// Add hover capabilities with preferred content format
					Hover: &protocol.HoverTextDocumentClientCapabilities{
						ContentFormat: []protocol.MarkupKind{protocol.Markdown},
					},
					// Add document symbol capabilities
					DocumentSymbol: &protocol.DocumentSymbolClientCapabilities{
						HierarchicalDocumentSymbolSupport: true,
					},
					// Add semantic tokens capabilities
					SemanticTokens: &protocol.SemanticTokensClientCapabilities{
						Formats: []protocol.TokenFormat{"relative"},
						Requests: protocol.SemanticTokensWorkspaceClientCapabilitiesRequests{
							Range: true,
							Full:  true,
						},
						TokenTypes:     []string{"keyword", "function", "variable", "string", "number", "comment"},
						TokenModifiers: []string{"declaration", "definition", "readonly", "static", "deprecated", "abstract", "async", "modification", "documentation", "defaultLibrary"},
					},
					// Add code action capabilities
					CodeAction: &protocol.CodeActionClientCapabilities{
						CodeActionLiteralSupport: &protocol.CodeActionClientCapabilitiesLiteralSupport{
							CodeActionKind: &protocol.CodeActionClientCapabilitiesKind{
								ValueSet: []protocol.CodeActionKind{},
							},
						},
					},
				},
				Workspace: &protocol.WorkspaceClientCapabilities{
					Configuration: true,
				},
			},
			InitializationOptions: map[string]interface{}{
				"symbolMatcher": "fuzzy", // Using fuzzy as default symbol matcher
			},
		}

		resp, err := server.Initialize(ctx, initParams)
		if err != nil {
			log.Fatalf("Failed to initialize LSP: %v", err)
		}

		fmt.Printf("Initialize response: %v\n", resp)

		err = server.Initialized(ctx, &protocol.InitializedParams{})
		if err != nil {
			log.Fatalf("Failed to initialize LSP: %v", err)
		}

		// didOpenParams := &protocol.DidOpenTextDocumentParams{
		// 	TextDocument: protocol.TextDocumentItem{
		// 		URI:        protocol.DocumentURI("file:///Users/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/lsp/main.go"),
		// 		LanguageID: "go",
		// 		Version:    1,
		// 		Text:       "",
		// 	},
		// }

		// err = server.DidOpen(ctx, didOpenParams)
		// if err != nil {
		// 	log.Fatalf("Failed to open text document: %v", err)
		// }

		// Request hover information for a specific position
		// hoverParams := &protocol.HoverParams{
		// 	TextDocumentPositionParams: protocol.TextDocumentPositionParams{
		// 		TextDocument: protocol.TextDocumentIdentifier{
		// 			URI: protocol.DocumentURI("file:///Users/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/lsp/main.go"),
		// 		},
		// 		Position: protocol.Position{
		// 			Line:      16, // 0-based line number
		// 			Character: 17, // 0-based character offset
		// 		},
		// 	},
		// }

		// hover, err := server.Hover(ctx, hoverParams)
		// if err != nil {
		// 	log.Fatalf("Failed to get hover information: %v", err)
		// }

		// // Print hover information
		// if hover != nil {
		// 	fmt.Printf("Hover content: %s\n", hover.Contents.Value)
		// }

		documentSymbolParams := &protocol.DocumentSymbolParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI("file:///Users/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/lsp/main.go"),
			},
		}

		documentSymbols, err := server.DocumentSymbol(ctx, documentSymbolParams)
		if err != nil {
			log.Fatalf("Failed to get document symbols: %v", err)
		}

		fmt.Printf("Document symbols: %v\n", documentSymbols)
		// serialize to yaml and then print
		yaml, err := yaml.Marshal(documentSymbols)
		if err != nil {
			log.Fatalf("Failed to marshal document symbols: %v", err)
		}
		fmt.Printf("Document symbols: %s\n", string(yaml))

		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatalf("Failed to wait for server: %v", err)
	}
}
