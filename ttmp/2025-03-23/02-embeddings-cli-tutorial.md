# Geppetto Embeddings CLI Tutorial

This tutorial demonstrates how to create a simple command-line tool that computes embeddings and similarity scores using the Geppetto embeddings framework with Glazed command parameters.

## 1. Introduction

The Geppetto embeddings package provides a unified interface for generating text embeddings from various providers such as OpenAI and Ollama. These embeddings can be used to compute similarity between texts, find relevant information, and build semantic search applications.

This tutorial will cover:
1. Setting up a basic CLI tool using Glazed and Cobra
2. Implementing commands to generate embeddings
3. Computing similarity between texts
4. Using parameter layers to configure embedding providers

## 2. Project Setup

Let's start by creating the project structure. We'll use the Go Modules system and organize our code as follows:

```
cmd/
  embeddings-cli/
    main.go
    commands/
      root.go
      embed.go
      similarity.go
```

### Initialize the Project

```bash
mkdir -p cmd/embeddings-cli/commands
touch cmd/embeddings-cli/main.go
touch cmd/embeddings-cli/commands/root.go
touch cmd/embeddings-cli/commands/embed.go
touch cmd/embeddings-cli/commands/similarity.go
```

## 3. Implementing the CLI

### Main Entry Point

Let's start with the main entry point of our application:

```go
// cmd/embeddings-cli/main.go
package main

import (
	"fmt"
	"os"

	"github.com/your-username/embeddings-cli/cmd/embeddings-cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

### Root Command

Next, let's implement the root command that will serve as the entry point for our CLI:

```go
// cmd/embeddings-cli/commands/root.go
package commands

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "embeddings-cli",
	Short: "A CLI tool for computing embeddings and similarity",
	Long: `embeddings-cli provides commands to generate embeddings for text
and compute similarity between texts using various embedding models.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands to root command
	// This will be called from the other command files
}
```

## 4. Implementing the Embed Command

Now, let's implement a command to generate embeddings for a text input:

```go
// cmd/embeddings-cli/commands/embed.go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/geppetto/pkg/embeddings"
	"github.com/go-go-golems/geppetto/pkg/embeddings/config"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type EmbedCommand struct {
	*cmds.CommandDescription
}

type EmbedSettings struct {
	Text        string `glazed.parameter:"text"`
	OutputFile  string `glazed.parameter:"output-file"`
	FormatJSON  bool   `glazed.parameter:"format-json"`
	FormatArray bool   `glazed.parameter:"format-array"`
}

func NewEmbedCommand() (*EmbedCommand, error) {
	// Create the embeddings parameter layer
	embeddingsLayer, err := config.NewEmbeddingsParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings parameter layer")
	}

	// Create the API key parameter layer
	embeddingsApiKey, err := config.NewEmbeddingsApiKeyParameter()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings API key parameter layer")
	}

	return &EmbedCommand{
		CommandDescription: cmds.NewCommandDescription(
			"embed",
			cmds.WithShort("Generate embeddings for text"),
			cmds.WithLong("Generate embeddings for the provided text using the configured embeddings provider."),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"text",
					parameters.ParameterTypeString,
					parameters.WithHelp("Text to generate embeddings for"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"output-file",
					parameters.ParameterTypeString,
					parameters.WithHelp("File to write the embeddings to (optional)"),
				),
				parameters.NewParameterDefinition(
					"format-json",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Output the embeddings as JSON"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"format-array",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Output the embeddings as a flat array"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(
				embeddingsLayer,
				embeddingsApiKey,
			),
		),
	}, nil
}

func (c *EmbedCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	// Parse command settings
	s := &EmbedSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "could not initialize settings")
	}

	// Create embeddings provider from parsed layers
	factory, err := embeddings.NewSettingsFactoryFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "could not create embeddings factory")
	}

	provider, err := factory.NewProvider()
	if err != nil {
		return errors.Wrap(err, "could not create embeddings provider")
	}

	// Generate embeddings
	embedding, err := provider.GenerateEmbedding(ctx, s.Text)
	if err != nil {
		return errors.Wrap(err, "could not generate embeddings")
	}

	// Determine output writer
	var outputWriter io.Writer = w
	if s.OutputFile != "" {
		file, err := os.Create(s.OutputFile)
		if err != nil {
			return errors.Wrap(err, "could not create output file")
		}
		defer file.Close()
		outputWriter = file
	}

	// Format and write the embeddings
	switch {
	case s.FormatJSON:
		data := map[string]interface{}{
			"text":       s.Text,
			"embeddings": embedding,
			"model":      provider.GetModel(),
		}
		encoder := json.NewEncoder(outputWriter)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			return errors.Wrap(err, "could not encode embeddings as JSON")
		}
	case s.FormatArray:
		for i, val := range embedding {
			if i > 0 {
				fmt.Fprint(outputWriter, " ")
			}
			fmt.Fprintf(outputWriter, "%f", val)
		}
		fmt.Fprintln(outputWriter)
	default:
		for i, val := range embedding {
			fmt.Fprintf(outputWriter, "[%d] %f\n", i, val)
		}
	}

	return nil
}

func init() {
	embedCmd, err := NewEmbedCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating embed command: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromWriterCommand(embedCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building cobra command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)
}
```

## 5. Implementing the Similarity Command

Now, let's implement a command to compute similarity between two texts:

```go
// cmd/embeddings-cli/commands/similarity.go
package commands

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/go-go-golems/geppetto/pkg/embeddings"
	"github.com/go-go-golems/geppetto/pkg/embeddings/config"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type SimilarityCommand struct {
	*cmds.CommandDescription
}

type SimilaritySettings struct {
	Text1 string `glazed.parameter:"text1"`
	Text2 string `glazed.parameter:"text2"`
}

func NewSimilarityCommand() (*SimilarityCommand, error) {
	// Create the embeddings parameter layer
	embeddingsLayer, err := config.NewEmbeddingsParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings parameter layer")
	}

	// Create the API key parameter layer
	embeddingsApiKey, err := config.NewEmbeddingsApiKeyParameter()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings API key parameter layer")
	}

	return &SimilarityCommand{
		CommandDescription: cmds.NewCommandDescription(
			"similarity",
			cmds.WithShort("Compute similarity between two texts"),
			cmds.WithLong("Compute the cosine similarity between two texts using the configured embeddings provider."),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"text1",
					parameters.ParameterTypeString,
					parameters.WithHelp("First text to compare"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"text2",
					parameters.ParameterTypeString,
					parameters.WithHelp("Second text to compare"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				embeddingsLayer,
				embeddingsApiKey,
			),
		),
	}, nil
}

// computeCosineSimilarity calculates the cosine similarity between two embedding vectors
func computeCosineSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct float64
	var norm1 float64
	var norm2 float64

	for i := 0; i < len(vec1); i++ {
		dotProduct += float64(vec1[i] * vec2[i])
		norm1 += float64(vec1[i] * vec1[i])
		norm2 += float64(vec2[i] * vec2[i])
	}

	// Avoid division by zero
	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

func (c *SimilarityCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	// Parse command settings
	s := &SimilaritySettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "could not initialize settings")
	}

	// Create embeddings provider from parsed layers
	factory, err := embeddings.NewSettingsFactoryFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "could not create embeddings factory")
	}

	provider, err := factory.NewProvider()
	if err != nil {
		return errors.Wrap(err, "could not create embeddings provider")
	}

	// Generate embeddings for both texts
	embedding1, err := provider.GenerateEmbedding(ctx, s.Text1)
	if err != nil {
		return errors.Wrap(err, "could not generate embeddings for text1")
	}

	embedding2, err := provider.GenerateEmbedding(ctx, s.Text2)
	if err != nil {
		return errors.Wrap(err, "could not generate embeddings for text2")
	}

	// Compute and print similarity
	similarity := computeCosineSimilarity(embedding1, embedding2)
	fmt.Fprintf(w, "Similarity: %.4f (%.2f%%)\n", similarity, similarity*100)

	return nil
}

func init() {
	similarityCmd, err := NewSimilarityCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating similarity command: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromWriterCommand(similarityCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building cobra command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)
}
```

## 6. Usage Examples

Once you've compiled the CLI tool, you can use it in the following ways:

### Generate Embeddings

```bash
# Using OpenAI
./embeddings-cli embed --text "Hello, world!" \
  --embeddings-type openai \
  --embeddings-engine text-embedding-3-small \
  --openai-api-key "sk-yourapikey"

# Using Ollama (local)
./embeddings-cli embed --text "Hello, world!" \
  --embeddings-type ollama \
  --embeddings-engine all-minilm \
  --embeddings-dimensions 384 \
  --ollama-base-url "http://localhost:11434"

# Save to file
./embeddings-cli embed --text "Hello, world!" \
  --embeddings-type openai \
  --openai-api-key "sk-yourapikey" \
  --output-file embeddings.json
```

### Compute Similarity

```bash
# Using OpenAI
./embeddings-cli similarity \
  --text1 "Hello, world!" \
  --text2 "Hi there, globe!" \
  --embeddings-type openai \
  --embeddings-engine text-embedding-3-small \
  --openai-api-key "sk-yourapikey"

# Using Ollama (local)
./embeddings-cli similarity \
  --text1 "Hello, world!" \
  --text2 "Hi there, globe!" \
  --embeddings-type ollama \
  --embeddings-engine all-minilm \
  --embeddings-dimensions 384 \
  --ollama-base-url "http://localhost:11434"
```

## 7. Using Configuration Files

Instead of passing all parameters on the command line every time, you can create a configuration file. The Glazed framework can load parameters from YAML files.

Create a file named `config.yaml`:

```yaml
embeddings-type: openai
embeddings-engine: text-embedding-3-small
openai-api-key: sk-yourapikey
embeddings-cache-type: file
embeddings-cache-directory: ~/.embeddings-cache
```

Then use it with the `--config` flag:

```bash
./embeddings-cli embed --text "Hello, world!" --config config.yaml
```

## 8. Adding Caching Support

The Geppetto embeddings framework supports caching embeddings to avoid repeated API calls. You can enable this with the appropriate flags:

```bash
# Memory cache
./embeddings-cli embed --text "Hello, world!" \
  --embeddings-type openai \
  --openai-api-key "sk-yourapikey" \
  --embeddings-cache-type memory \
  --embeddings-cache-max-entries 1000

# File cache
./embeddings-cli embed --text "Hello, world!" \
  --embeddings-type openai \
  --openai-api-key "sk-yourapikey" \
  --embeddings-cache-type file \
  --embeddings-cache-directory "~/.embeddings-cache" \
  --embeddings-cache-max-size 1073741824  # 1GB
```

## 9. Conclusion

You've now created a fully functional CLI tool for generating embeddings and computing similarity using different providers. The Geppetto embeddings framework provides a unified interface that makes it easy to switch between providers and add caching.

Key concepts covered:
1. Creating CLI commands using Glazed and Cobra
2. Using parameter layers to configure embedding providers
3. Generating embeddings and computing similarity between texts
4. Adding caching to improve performance

This CLI tool can be extended with additional commands for more advanced embedding operations, such as semantic search, document clustering, or visualizing embeddings. 