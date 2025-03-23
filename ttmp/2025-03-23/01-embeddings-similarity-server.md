Base package: 	"github.com/go-go-golems/go-go-labs/cmd/apps/embeddings/"
Base module: /home/manuel/code/wesen/corporate-headquarters/go-go-labs
Base path: /home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/apps/embeddings

- [x] http server with mock API for compute-embeddings + compute-similarity
- [x] htmx + templ.guide template + bootstrap css webui to compare 3 pieces of text similarity wise (A vs B + A vs C + B vs C)
- [x] hook it up with geppetto's embeddings provider + glazed parameter layers
- [x] structure the application using standard Glazed patterns (layers, middlewares, help system)

## Implementation Notes

The server now integrates with the Geppetto embeddings package to provide real embeddings and similarity computation:

1. Added a tutorial in `ttmp/2025-03-23/02-embeddings-cli-tutorial.md` explaining how to use Geppetto embeddings with Glazed in a CLI application.
2. Refactored the server to use an `EmbeddingsServer` struct that contains the embeddings provider factory.
3. Implemented the `ServerCommand` as a Glazed `BareCommand` for more elegant parameter handling:
   - Created appropriate Glazed parameter layers for embeddings configuration
   - Used `ServerSettings` struct with tagged fields for command parameters
   - The parameter layers are the same ones used in the CLI tutorial, showing consistent API design
4. Implemented the actual computation of embeddings and similarity scores using the Geppetto framework.
5. Structured the entire application using Glazed conventions:
   - Added proper middleware chain in `GetEmbeddingsMiddlewares` for parameter processing
   - Set up standardized help system support
   - Added configuration via files, environment variables, and profiles
   - Implemented proper logging with zerolog integration
   - Used Clay for Viper initialization and logger setup

## Architecture

The application follows a well-structured architecture:

1. **Parameter Layers**: Define all configuration options with proper defaults, help texts, and validation.
2. **Command Registration**: Commands are defined as Glazed `Command` implementations and collected in `embeddings_commands`.
3. **Middleware Chain**: Parameter processing follows a standardized flow (command line → config file → profile → environment → defaults).
4. **Parser Integration**: Commands are registered with the Cobra root command using a Glazed parser.

## Usage

You can start the server with various configuration options using Glazed parameter formats:

```bash
# Using OpenAI
go run ./cmd/apps/embeddings serve --embeddings-type openai --embeddings-engine text-embedding-3-small --openai-api-key YOUR_API_KEY --port 8080

# Using Ollama (local)
go run ./cmd/apps/embeddings serve --embeddings-type ollama --embeddings-engine all-minilm --embeddings-dimensions 384 --ollama-base-url http://localhost:11434 --port 8080

# Using configuration file (recommended for sensitive keys)
go run ./cmd/apps/embeddings serve --config embeddings-config.yaml
```

Example configuration file (embeddings-config.yaml):
```yaml
embeddings-type: openai
embeddings-engine: text-embedding-3-small
openai-api-key: sk-your-api-key
embeddings-cache-type: memory
port: 8080
```

The server exposes the following endpoints:
- `GET /` - Web UI for comparing text similarity
- `POST /compare` - HTMX endpoint for comparing text similarity
- `POST /compute-embeddings` - API endpoint for computing embeddings
- `POST /compute-similarity` - API endpoint for computing similarity

With the addition of the Geppetto embeddings integration, the server now provides accurate embeddings and similarity scores based on the chosen provider, rather than random mock data.

## Future Enhancements

Potential improvements for the future:
1. Add more commands for batch processing of embeddings
2. Implement vector storage backends
3. Create dashboard for visualizing embeddings in 2D/3D space
4. Add support for more embedding providers