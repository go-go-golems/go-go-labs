# Building a CLI Application with Glazed and Geppetto Embeddings

Welcome to this hands-on tutorial where we'll build a powerful, well-structured command-line tool that integrates text embeddings capabilities. Think of this as building your own "Swiss Army knife" for natural language processing, right from your terminal!

## Table of Contents

1. [Introduction](#introduction)
2. [Understanding the Architecture](#understanding-the-architecture)
3. [Project Setup](#project-setup)
4. [Building the Foundation](#building-the-foundation)
5. [Adding Parameter Layers](#adding-parameter-layers)
6. [Creating Commands](#creating-commands)
7. [Configuring Your Application](#configuring-your-application)
8. [Running and Testing](#running-and-testing)
9. [Advanced Customization](#advanced-customization)
10. [Conclusion](#conclusion)

## Introduction

Imagine you need to build a tool that can analyze how similar different pieces of text are—perhaps you want to compare product descriptions, find similar support tickets, or analyze document similarity. Such a tool needs to be flexible (supporting different embedding providers), configurable (allowing users to customize behavior), and well-structured (making it easy to maintain and extend).

This is where the **Glazed framework** and **Geppetto embeddings** come in:

- **Glazed** is a framework for building command-line applications with a focus on structured parameter handling and configuration management
- **Geppetto** provides embeddings capabilities—turning text into numerical vectors that capture semantic meaning

By the end of this tutorial, you'll have built a CLI tool that can:

- Generate embeddings for text
- Compute similarity between texts
- Load configuration from various sources
- Provide rich help documentation

**Who is this for?** This tutorial is designed for Go developers who want to build more sophisticated CLI applications, especially those working with AI and natural language processing.

Let's start by understanding the architecture we'll be building.

## Understanding the Architecture

Before diving into code, let's understand the key architectural components:

### The Big Picture

Think of our application like a well-organized kitchen:

- **The Main Program (`main.go`)**: The head chef, coordinating everything
- **Parameter Layers**: Recipe cards, organized by category
- **Commands**: Specific cooking techniques for different dishes
- **Middleware Chain**: The preparation sequence—wash, chop, season, etc.
- **Configuration System**: Ways to adjust recipes based on available ingredients

Let's explore each component:

#### 1. Commands

Commands are the core of our application. Each command:
- Has a specific purpose (like "generate embeddings" or "compute similarity")
- Takes specific parameters
- Produces specific output

#### 2. Parameter Layers

Parameters in Glazed are organized into "layers." This is a powerful concept that allows you to:
- Group related parameters together
- Reuse parameter groups across commands
- Set up layered configuration precedence

For example, all embedding-related parameters (model type, dimensions, etc.) form one layer, while API authentication parameters form another.

#### 3. Middleware Chain

The middleware chain processes parameters from different sources in a specific order. Think of it as a pipeline where:
1. Command-line arguments are checked first
2. If a parameter isn't specified there, check config files
3. If not there, check environment variables
4. If all else fails, use default values

#### 4. Configuration System

The configuration system allows users to store settings in:
- Configuration files
- Environment variables
- Profiles (named sets of parameters)

This makes the application flexible and adaptable to different environments.

**🧠 Thinking Point:** How would this architecture benefit a team where different members need different default settings? How about when deploying the same tool across development and production environments?

Now that we have the big picture, let's start building!

## Project Setup

Let's begin by setting up our project structure and dependencies.

### Step 1: Create the Project Directory

```bash
mkdir -p embeddings-cli/cmd/embeddings-cli
cd embeddings-cli
```

### Step 2: Initialize Go Module

```bash
go mod init github.com/yourusername/embeddings-cli
```

### Step 3: Install Dependencies

```bash
go get github.com/go-go-golems/glazed
go get github.com/go-go-golems/clay
go get github.com/go-go-golems/geppetto
go get github.com/spf13/cobra
go get github.com/rs/zerolog
go get github.com/pkg/errors
```

Let's look at what each dependency provides:

- **glazed**: The core framework for command structure and parameter handling
- **clay**: Helpers for configuration and logging
- **geppetto**: Embeddings generation and processing
- **cobra**: Command-line interface framework (used by Glazed)
- **zerolog**: Structured logging
- **errors**: Better error handling with wrapping

### Step 4: Create the Application Structure

We'll organize our code with this structure:

```
embeddings-cli/
├── cmd/
│   └── embeddings-cli/
│       ├── main.go             # Main entry point
│       ├── commands/           # Command implementations
│       │   ├── commands.go     # Command registration
│       │   ├── embed.go        # Embed command
│       │   └── similarity.go   # Similarity command
│       └── layers/             # Custom parameter layers
│           └── layers.go       # Layer definitions
```

```bash
mkdir -p cmd/embeddings-cli/commands
mkdir -p cmd/embeddings-cli/layers
touch cmd/embeddings-cli/main.go
touch cmd/embeddings-cli/commands/commands.go
touch cmd/embeddings-cli/commands/embed.go
touch cmd/embeddings-cli/commands/similarity.go
touch cmd/embeddings-cli/layers/layers.go
```

**✅ Checkpoint:** At this point, you should have a basic project structure with all the necessary dependencies installed.

## Building the Foundation

Now, let's implement the main entry point of our application, which will:
1. Set up logging
2. Create the root command
3. Initialize configuration
4. Register our commands
5. Execute the command

### Step 1: Implement main.go

Create `cmd/embeddings-cli/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	embeddings_config "github.com/go-go-golems/geppetto/pkg/embeddings/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/yourusername/embeddings-cli/cmd/embeddings-cli/commands"
	"github.com/yourusername/embeddings-cli/cmd/embeddings-cli/layers"
)

// Global commands registry
var appCommands []cmds.Command

func main() {
	// Initialize logging with a nice console format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create and configure root command
	rootCmd := &cobra.Command{
		Use:   "embeddings-cli",
		Short: "Text embeddings and similarity tool",
		Long: `A CLI tool for generating embeddings and computing similarity between texts.
It supports multiple embedding providers and configuration methods.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Reinitialize logger to apply log level from command line
			err := clay.InitLogger()
			cobra.CheckErr(err)
		},
	}

	// Initialize Viper for config file support
	err := clay.InitViper("embeddings-cli", rootCmd)
	cobra.CheckErr(err)
	
	// Initialize logger
	err = clay.InitLogger()
	cobra.CheckErr(err)

	// Set up help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Register commands from the commands package
	err = commands.RegisterCommands(&appCommands)
	cobra.CheckErr(err)

	// Register all commands using our middleware stack
	err = cli.AddCommandsToRootCommand(rootCmd, appCommands, []*alias.CommandAlias{},
		cli.WithCobraMiddlewaresFunc(GetEmbeddingsMiddlewares),
		cli.WithProfileSettingsLayer())
	cobra.CheckErr(err)

	// Execute the root command
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// GetEmbeddingsMiddlewares returns the middleware stack for processing commands
func GetEmbeddingsMiddlewares(
	parsedLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {
	// Parse command settings
	commandSettings := &cli.CommandSettings{}
	err := parsedLayers.InitializeStruct(cli.CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}

	// Parse profile settings
	profileSettings := &cli.ProfileSettings{}
	err = parsedLayers.InitializeStruct(cli.ProfileSettingsSlug, profileSettings)
	if err != nil {
		return nil, err
	}

	// Create middleware chain
	middlewareStack := []middlewares.Middleware{
		// Parse command-line flags
		middlewares.ParseFromCobraCommand(cmd, 
			parameters.WithParseStepSource("cobra"),
		),
		// Gather command-line arguments
		middlewares.GatherArguments(args, 
			parameters.WithParseStepSource("arguments"),
		),
	}

	// Add config file loading if specified
	if commandSettings.LoadParametersFromFile != "" {
		middlewareStack = append(middlewareStack,
			middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
	}

	// Add profile support
	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	defaultProfileFile := fmt.Sprintf("%s/embeddings-cli/profiles.yaml", xdgConfigPath)
	if profileSettings.ProfileFile == "" {
		profileSettings.ProfileFile = defaultProfileFile
	}
	if profileSettings.Profile == "" {
		profileSettings.Profile = "default"
	}

	// Add profile middleware
	middlewareStack = append(middlewareStack,
		middlewares.GatherFlagsFromProfiles(
			defaultProfileFile,
			profileSettings.ProfileFile,
			profileSettings.Profile,
			parameters.WithParseStepSource("profiles"),
			parameters.WithParseStepMetadata(map[string]interface{}{
				"profileFile": profileSettings.ProfileFile,
				"profile":     profileSettings.Profile,
			}),
		),
	)

	// Add viper and defaults
	middlewareStack = append(middlewareStack,
		middlewares.WrapWithWhitelistedLayers(
			[]string{
				embeddings_config.EmbeddingsSlug,
				embeddings_config.EmbeddingsApiKeySlug,
				cli.CommandSettingsSlug,
				cli.ProfileSettingsSlug,
			},
			middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("viper")),
		),
		middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	)

	return middlewareStack, nil
}
```

Let's examine the key components of this file:

1. **Logging Setup**: We use zerolog to create structured logs with a nice console format.

2. **Root Command Configuration**: We create the main command (`embeddings-cli`) with descriptions that show up in help text.

3. **Help System Setup**: Glazed enhances Cobra's built-in help with more comprehensive documentation.

4. **Middleware Configuration**: 
   - This is where the magic happens for parameter resolution
   - We define a clear order of precedence for parameter sources
   - Each middleware in the chain processes parameters from a different source

5. **Command Registration**: We register all commands with the root command.

**🔍 Deep Dive: The Middleware Chain**

The middleware chain is crucial for parameter resolution. Let's understand the order:

1. **Parse command-line flags**: Highest priority—user-specified flags override all other sources
2. **Gather arguments**: Positional arguments from the command line
3. **Load from config file**: If `--config` is specified, load parameters from that file
4. **Apply profile settings**: If a profile is specified, apply those parameters
5. **Load from environment variables**: Via Viper
6. **Apply defaults**: Lowest priority—only used if no other source specifies the parameter

This layered approach gives users flexibility in how they configure the application.

**✅ Checkpoint:** Make sure you understand how the main.go file sets up the application and how the middleware chain works. This understanding is fundamental to how Glazed applications process parameters.

## Adding Parameter Layers

Parameter layers organize related parameters together, making commands more maintainable and reusable. Now, let's implement the layers for our application.

### Step 1: Implement the Layers Package

Create `cmd/embeddings-cli/layers/layers.go`:

```go
package layers

import (
	"github.com/go-go-golems/geppetto/pkg/embeddings/config"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/pkg/errors"
)

// GetEmbeddingsLayers returns all parameter layers used by the embeddings commands
func GetEmbeddingsLayers() ([]layers.ParameterLayer, error) {
	// Create embeddings parameter layers
	embeddingsLayer, err := config.NewEmbeddingsParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings parameter layer")
	}

	embeddingsApiKey, err := config.NewEmbeddingsApiKeyParameter()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings API key parameter layer")
	}

	// Command settings layer (for --config flag)
	commandSettingsLayer, err := cli.NewCommandSettingsParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create command settings parameter layer")
	}

	return []layers.ParameterLayer{
		embeddingsLayer,
		embeddingsApiKey,
		commandSettingsLayer,
	}, nil
}
```

**What's happening here?**

We're creating three parameter layers:

1. **Embeddings Layer**: Contains parameters for controlling embeddings behavior:
   - `embeddings-type`: Provider to use (OpenAI, Ollama, etc.)
   - `embeddings-engine`: Model to use for embeddings
   - `embeddings-dimensions`: Vector dimensions
   - And more...

2. **API Key Layer**: Contains sensitive authentication parameters:
   - `openai-api-key`: API key for OpenAI

3. **Command Settings Layer**: Contains parameters for command behavior:
   - `config`: Path to configuration file
   - And more...

**🧩 How It Fits Together: Parameter Layers**

Parameter layers serve multiple purposes:
- **Modularity**: Each layer can evolve independently
- **Reusability**: The same layer can be used across multiple commands
- **Organization**: Related parameters are grouped together
- **Documentation**: Each layer has a clear purpose and description

Think of parameter layers like ingredients in a meal kit—pre-packaged groups of related items that can be combined in different ways depending on the recipe (command).

**✅ Checkpoint:** You should now understand how parameter layers organize related parameters. Make sure you can explain why grouping parameters into layers is beneficial.

## Creating Commands

Now let's implement the commands that will use these parameter layers. We'll start with a command registry.

### Step 1: Create the Command Registry

In `cmd/embeddings-cli/commands/commands.go`:

```go
package commands

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/pkg/errors"
)

// RegisterCommands initializes and registers all commands
func RegisterCommands(appCommands *[]cmds.Command) error {
	// Register embed command
	embedCmd, err := NewEmbedCommand()
	if err != nil {
		return errors.Wrap(err, "could not create embed command")
	}
	*appCommands = append(*appCommands, embedCmd)

	// Register similarity command
	similarityCmd, err := NewSimilarityCommand()
	if err != nil {
		return errors.Wrap(err, "could not create similarity command")
	}
	*appCommands = append(*appCommands, similarityCmd)

	return nil
}
```

This function initializes all commands and adds them to the global command registry.

### Step 2: Implement the Embed Command

Now, let's create the embed command in `cmd/embeddings-cli/commands/embed.go`:

```go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/geppetto/pkg/embeddings"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/yourusername/embeddings-cli/cmd/embeddings-cli/layers" // Adjust import as needed
)

// EmbedCommand represents the command to generate embeddings
type EmbedCommand struct {
	*cmds.CommandDescription
}

// EmbedSettings contains settings for the embed command
type EmbedSettings struct {
	Text        string `glazed.parameter:"text"`
	OutputFile  string `glazed.parameter:"output-file"`
	FormatJSON  bool   `glazed.parameter:"format-json"`
	FormatArray bool   `glazed.parameter:"format-array"`
}

// NewEmbedCommand creates a new embed command
func NewEmbedCommand() (*EmbedCommand, error) {
	// Get all layers
	cmdLayers, err := layers.GetEmbeddingsLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings layers")
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
			cmds.WithLayersList(cmdLayers...),
		),
	}, nil
}

// RunIntoWriter implements the WriterCommand interface
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
```

Let's break down this command implementation:

1. **Command Description**: Defines the command's name, help text, and parameters.

2. **Parameter Definition**: Each parameter has:
   - A name
   - A type
   - Help text
   - Optionally, a default value and validation rules

3. **Settings Struct**: We define a struct (`EmbedSettings`) to hold parameter values.
   - Note the `glazed.parameter` tags that map struct fields to parameter names

4. **RunIntoWriter Method**: This is where the command's logic lives:
   - It initializes settings from parsed layers
   - Creates an embeddings provider based on configuration
   - Generates embeddings
   - Formats and outputs the result

**🧩 How It Fits Together: Command Structure**

Commands in Glazed have a consistent structure:
1. A struct embedding `CommandDescription`
2. A settings struct with tagged fields
3. A factory function that creates the command with parameters and layers
4. An implementation of `RunIntoWriter` or `Run` that executes the command

This structure makes commands:
- **Self-contained**: Each command has everything it needs
- **Declarative**: Parameters are defined with clear metadata
- **Consistent**: All commands follow the same pattern

**✅ Checkpoint:** You should now understand how commands are structured in Glazed. Make sure you can explain the relationship between command parameters, settings structs, and parsed layers.

### Step 3: Implement the Similarity Command

Now, let's implement the similarity command in `cmd/embeddings-cli/commands/similarity.go`:

```go
package commands

import (
	"context"
	"fmt"
	"io"
	"math"

	"github.com/go-go-golems/geppetto/pkg/embeddings"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/yourusername/embeddings-cli/cmd/embeddings-cli/layers" // Adjust import as needed
)

// SimilarityCommand represents the command to compute similarity
type SimilarityCommand struct {
	*cmds.CommandDescription
}

// SimilaritySettings contains settings for the similarity command
type SimilaritySettings struct {
	Text1 string `glazed.parameter:"text1"`
	Text2 string `glazed.parameter:"text2"`
}

// NewSimilarityCommand creates a new similarity command
func NewSimilarityCommand() (*SimilarityCommand, error) {
	// Get all layers
	cmdLayers, err := layers.GetEmbeddingsLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings layers")
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
			cmds.WithLayersList(cmdLayers...),
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

// RunIntoWriter implements the WriterCommand interface
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
```

This command follows the same structure as the embed command but implements a different functionality: comparing the similarity between two texts.

**🔬 Technical Note: Cosine Similarity**

The similarity computation uses cosine similarity, which measures the cosine of the angle between two vectors. Values range from -1 (exactly opposite) to 1 (exactly the same), with 0 indicating orthogonality (no similarity).

In natural language processing, cosine similarity is commonly used to measure document similarity because it's:
- Independent of vector magnitude (document length)
- Efficient to compute
- Interpretable (values close to 1 mean similar documents)

**✅ Checkpoint:** You should now have two working commands. Make sure you understand how each command processes parameters and produces output.

## Configuring Your Application

One of Glazed's strengths is its flexible configuration system. Let's explore how users can configure your application.

### Configuration File

Users can create a configuration file to avoid typing the same parameters repeatedly. The default location would be `$HOME/.config/embeddings-cli/config.yaml`:

```yaml
# Default configuration for embeddings-cli
embeddings-type: openai
embeddings-engine: text-embedding-3-small
embeddings-dimensions: 1536
embeddings-cache-type: memory
```

### Environment Variables

Users can also set parameters through environment variables. Glazed (via Viper) automatically converts parameter names to environment variables:

```bash
# Set OpenAI API key
export EMBEDDINGS_CLI_OPENAI_API_KEY=sk-your-api-key

# Set embeddings type
export EMBEDDINGS_CLI_EMBEDDINGS_TYPE=openai
```

### Profiles

Profiles allow users to switch between different configurations. They create a profiles file at `$HOME/.config/embeddings-cli/profiles.yaml`:

```yaml
default:
  embeddings-type: openai
  embeddings-engine: text-embedding-3-small
  openai-api-key: sk-your-default-api-key

ollama:
  embeddings-type: ollama
  embeddings-engine: all-minilm
  embeddings-dimensions: 384
  ollama-base-url: http://localhost:11434

production:
  # Production settings with higher quality model
  embeddings-type: openai
  embeddings-engine: text-embedding-3-large
  openai-api-key: sk-your-production-api-key
```

**🧩 How It Fits Together: Configuration Hierarchy**

The configuration system follows this precedence (highest to lowest):

1. Command-line flags
2. Configuration file specified with `--config`
3. Profile values (if a profile is specified)
4. Environment variables
5. Default values specified in parameter definitions

This gives users flexibility in how they configure the application while maintaining sensible defaults.

**💡 Real-World Example: Development vs. Production**

Imagine a team of developers working on a project that uses embeddings:

- Developers use the `ollama` profile during development to avoid API costs
- The CI/CD pipeline uses environment variables to inject production API keys
- The production deployment uses a configuration file with optimal settings

All of this works with the same application, no code changes needed!

## Running and Testing

Let's see how to run our application with different configuration options.

### Using Command-Line Arguments

```bash
# Generate embeddings for text
./embeddings-cli embed --text "Hello, world!" \
  --embeddings-type openai \
  --embeddings-engine text-embedding-3-small \
  --openai-api-key "sk-yourapikey"

# Compute similarity between texts
./embeddings-cli similarity \
  --text1 "Hello, world!" \
  --text2 "Hi there, globe!" \
  --embeddings-type openai \
  --openai-api-key "sk-yourapikey"

# Save embeddings to a file
./embeddings-cli embed \
  --text "Hello, world!" \
  --embeddings-type openai \
  --openai-api-key "sk-yourapikey" \
  --output-file embeddings.json
```

### Using Configuration File

```bash
# Using a specific config file
./embeddings-cli embed --text "Hello, world!" --config my-config.yaml

# Config file with API key overridden on command line
./embeddings-cli embed --text "Hello, world!" \
  --config my-config.yaml \
  --openai-api-key "sk-newkey"
```

### Using Profiles

```bash
# Use the default profile
./embeddings-cli embed --text "Hello, world!" --profile default

# Use the ollama profile
./embeddings-cli embed --text "Hello, world!" --profile ollama

# Use a specific profile file
./embeddings-cli embed --text "Hello, world!" \
  --profile-file custom-profiles.yaml \
  --profile myprofile
```

### Testing Your Application

To verify your application works correctly:

1. **Check Help Text**:
```bash
./embeddings-cli --help
./embeddings-cli embed --help
./embeddings-cli similarity --help
```

2. **Try Different Configuration Methods**:
   - Command-line arguments
   - Configuration file
   - Environment variables
   - Profiles

3. **Test Error Handling**:
   - Missing required parameters
   - Invalid parameter values
   - API errors

**✅ Checkpoint:** Your application should now be fully functional! Make sure you understand how the different configuration methods work together.

## Advanced Customization

Now that you have a working application, let's explore some advanced customization options.

### Custom Parameter Layers

You can create custom parameter layers for application-specific parameters:

```go
func NewCustomLayer() (*layers.ParameterLayerImpl, error) {
	return layers.NewParameterLayer("my-custom",
		"My Custom Settings",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"custom-option",
				parameters.ParameterTypeString,
				parameters.WithHelp("A custom option"),
				parameters.WithDefault("default-value"),
			),
			parameters.NewParameterDefinition(
				"advanced-flag",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable advanced features"),
				parameters.WithDefault(false),
			),
		),
	)
}
```

Then add this layer to your command's layers list and whitelist it in the middleware function.

### Custom Command Types

Glazed supports different command types for different use cases:

- **BareCommand**: Simple command with no structured output
- **WriterCommand**: Command that writes output to a provided writer
- **GlazeCommand**: Command that outputs structured data rows

Choose the command type that best fits your needs.

### Command Aliases

You can create aliases for frequently used commands:

```go
aliases := []*alias.CommandAlias{
	alias.NewCommandAlias(
		"sim",
		[]string{"similarity"},
		"Short alias for similarity",
	),
	alias.NewCommandAlias(
		"e",
		[]string{"embed"},
		"Short alias for embed",
	),
}
```

Then register these aliases with the root command.

**🚀 Beyond the Basics: What's Next?**

Once you're comfortable with Glazed's basic architecture, you can explore:

- **Batch Processing**: Add commands for processing multiple files
- **Vector Storage**: Integrate with vector databases
- **Visualization**: Add commands for visualizing embeddings
- **Custom Providers**: Create your own embeddings providers

## Conclusion

Congratulations! You've built a powerful, well-structured command-line application for generating and comparing text embeddings. Let's recap what you've learned:

1. **Glazed Architecture**: 
   - Commands with clear structure and purpose
   - Parameter layers for organization and reuse
   - Middleware chain for parameter resolution
   - Configuration from multiple sources

2. **Geppetto Integration**:
   - Easy access to embeddings functionality
   - Support for different embedding providers
   - Similarity computation between texts

3. **Best Practices**:
   - Clear separation of concerns
   - Explicit parameter definitions
   - Consistent error handling
   - Comprehensive help documentation

You now have a solid foundation for building more sophisticated CLI applications. The patterns you've learned can be applied to a wide range of applications, not just those involving text embeddings.

**🤔 What could you build with this knowledge?**

- A semantic search tool for local documents
- A content recommendation system
- A document clustering tool
- A duplicate detection system

The possibilities are endless! Happy coding!

**Resources for Further Learning**:

- [Glazed GitHub Repository](https://github.com/go-go-golems/glazed)
- [Geppetto GitHub Repository](https://github.com/go-go-golems/geppetto)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Viper Documentation](https://github.com/spf13/viper) 