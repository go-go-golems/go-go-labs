# Understanding Glazed Architecture

This document provides an in-depth look at the Glazed architecture, focusing on parameter layers, middleware chains, command registration, and the overall compositional approach.

## Table of Contents
1. [Core Concepts](#core-concepts)
2. [Parameter Layers](#parameter-layers)
3. [Middleware Chain](#middleware-chain)
4. [Command Registration](#command-registration)
5. [Parameter Resolution](#parameter-resolution)
6. [Integration with Cobra](#integration-with-cobra)
7. [Command Aliases](#command-aliases)
8. [Profiles](#profiles)
9. [Best Practices](#best-practices)

## Core Concepts

Glazed is a framework for building command-line applications with a focus on:

- **Separation of concerns**: Different aspects of a command are handled by different components.
- **Composition over inheritance**: Building complex behavior by combining simple components.
- **Configuration flexibility**: Parameters can come from multiple sources with clear precedence.
- **Extensibility**: Easy to add new commands, middleware, or parameter sources.

The key architectural components of Glazed are:

1. **Commands**: Encapsulate functionality and parameter definitions
2. **Parameter Layers**: Group related parameters
3. **Middleware Chain**: Process parameters from various sources
4. **Parsed Layers**: Hold parameter values after processing

## Parameter Layers

Parameter layers are a core concept in Glazed. They organize parameters into logical groups, making commands more maintainable and reusable.

### Structure

A parameter layer consists of:

- **Slug**: A unique identifier for the layer (e.g., `embeddings`, `output-format`)
- **Description**: Human-readable description of the layer
- **Parameters**: Set of parameter definitions with types, defaults, etc.

### Example

```go
func NewOutputFormatLayer() (*layers.ParameterLayerImpl, error) {
    return layers.NewParameterLayer(
        "output-format",
        "Output formatting options",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "output",
                parameters.ParameterTypeString,
                parameters.WithHelp("Output format (json, yaml, table, etc.)"),
                parameters.WithDefault("table"),
            ),
            parameters.NewParameterDefinition(
                "pretty",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Pretty print the output"),
                parameters.WithDefault(true),
            ),
        ),
    )
}
```

### Benefits

- **Reusability**: The same layer can be used across multiple commands
- **Organization**: Related parameters are grouped together
- **Documentation**: Each layer has a clear purpose and description
- **Versioning**: Layers can evolve independently

## Middleware Chain

The middleware chain is responsible for collecting parameter values from various sources and resolving them according to precedence rules.

### Common Middleware Components

1. **ParseFromCobraCommand**: Extracts parameters from command-line flags
2. **GatherArguments**: Collects positional arguments
3. **LoadParametersFromFile**: Loads parameters from a configuration file
4. **GatherFlagsFromProfiles**: Applies parameters from profile configurations
5. **GatherFlagsFromViper**: Collects parameters from environment variables
6. **SetFromDefaults**: Applies default values for parameters not set otherwise

### Middleware Order

The order of middleware in the chain determines parameter precedence. Typically:

1. Command-line arguments (highest priority)
2. Configuration file
3. Profile values
4. Environment variables
5. Default values (lowest priority)

### Example

```go
func GetMiddlewareChain(
    parsedLayers *layers.ParsedLayers,
    cmd *cobra.Command,
    args []string,
) ([]middlewares.Middleware, error) {
    middlewareStack := []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.GatherArguments(args),
        // Configuration file middleware
        middlewares.LoadParametersFromFile("/path/to/config.yaml"),
        // Profile middleware
        middlewares.GatherFlagsFromProfiles(
            defaultProfileFile,
            profileFile,
            profileName,
        ),
        // Environment variables
        middlewares.GatherFlagsFromViper(),
        // Defaults
        middlewares.SetFromDefaults(),
    }
    
    return middlewareStack, nil
}
```

## Command Registration

Commands in Glazed encapsulate specific functionality and are typically registered with a Cobra command.

### Command Structure

A Glazed command typically consists of:

1. **Command Description**: Metadata about the command
2. **Parameter Layers**: Grouped parameters used by the command
3. **RunIntoWriter**: Implementation of the command functionality

### Registration Process

Commands are registered with the root Cobra command using helper functions:

```go
func main() {
    rootCmd := &cobra.Command{
        Use:   "app",
        Short: "Application description",
    }
    
    // Register commands
    commands := []cmds.Command{
        mycommand.NewCommand(),
        anothercommand.NewCommand(),
    }
    
    // Add commands to root command
    err := cli.AddCommandsToRootCommand(
        rootCmd,
        commands,
        aliases,
        cli.WithCobraMiddlewaresFunc(GetMiddlewares),
    )
    if err != nil {
        // Handle error
    }
    
    // Execute
    rootCmd.Execute()
}
```

## Parameter Resolution

Parameter resolution follows these steps:

1. The command is executed, triggering the middleware chain
2. Each middleware processes parameters from its source
3. Parameters are stored in a `ParsedLayers` object
4. The command accesses parameter values from the `ParsedLayers`

### ParsedLayers

The `ParsedLayers` object:

- Maps layer slugs to parameter values
- Tracks which source provided each parameter
- Resolves parameter values based on precedence
- Can initialize structs with parameter values

### Example

```go
func (c *MyCommand) RunIntoWriter(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    w io.Writer,
) error {
    // Initialize settings from parsed layers
    settings := &MySettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return err
    }
    
    // Use settings
    fmt.Fprintf(w, "Value: %s\n", settings.Value)
    
    return nil
}
```

## Integration with Cobra

Glazed integrates with Cobra in several ways:

1. **Command Registration**: Glazed commands are wrapped with Cobra commands
2. **Flag Definition**: Parameters are exposed as Cobra flags
3. **Execution**: The Cobra PreRun and Run hooks trigger the Glazed middleware chain
4. **Help System**: Glazed enhances Cobra's help system with parameter layer documentation

### Example

```go
// Add Glazed commands to Cobra root command
err := cli.AddCommandsToRootCommand(
    rootCmd,
    commands,
    aliases,
    cli.WithCobraMiddlewaresFunc(middlewaresFunc),
    cli.WithProfileSettingsLayer(),
)
```

## Command Aliases

Command aliases allow creating shortcuts or alternative names for commands:

```go
aliases := []*alias.CommandAlias{
    alias.NewCommandAlias(
        "s",
        []string{"search"},
        "Short alias for search",
    ),
}
```

## Profiles

Profiles provide a way to save and reuse parameter configurations:

1. **Profile files**: Typically stored in `~/.config/app/profiles.yaml`
2. **Profile selection**: Users specify which profile to use
3. **Profile parameters**: Applied during middleware processing

### Example Profile File

```yaml
default:
  output: json
  pretty: true
  verbose: false

development:
  output: table
  pretty: true
  verbose: true
  debug: true
```

## Best Practices

### Command Design

1. **Single Responsibility**: Each command should do one thing well
2. **Parameter Grouping**: Use parameter layers to group related parameters
3. **Clear Documentation**: Provide helpful short and long descriptions
4. **Meaningful Defaults**: Choose sensible default values

### Layer Design

1. **Cohesion**: Group related parameters in the same layer
2. **Reusability**: Design layers to be reusable across commands
3. **Validation**: Add parameter validation where appropriate
4. **Naming**: Use clear, consistent naming conventions

### Middleware Chain

1. **Precedence**: Define a clear precedence order for parameter sources
2. **Error Handling**: Handle errors properly in each middleware
3. **Logging**: Log parameter resolution for debugging

### Application Structure

1. **Organization**: Group related commands in packages
2. **Common Layers**: Share parameter layers across commands
3. **Configuration**: Provide flexible configuration options
4. **Testing**: Test commands with different parameter sources

By following these principles, you can build maintainable, user-friendly command-line applications with Glazed. 