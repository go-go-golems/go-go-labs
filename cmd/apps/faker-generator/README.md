# Faker Generator

`faker-generator` is a command-line tool that processes YAML templates embedded with special `!Faker*` tags to generate structured data with random values powered by [go-faker/faker](https://github.com/go-faker/faker). It uses the [go-go-golems/go-emrichen](https://github.com/go-go-golems/go-emrichen) library to parse and interpret the YAML templates.

This allows you to define complex data structures in YAML and populate them with realistic fake data for testing, seeding databases, creating mock APIs, and more.

## Features

- Processes YAML files containing Emrichen templating logic.
- Provides custom `!Faker*` tags to generate various types of random data (names, emails, IPs, numbers, text, etc.).
- Supports parameters for some tags (e.g., min/max for numbers, length for passwords).
- Integrates with standard Emrichen features like `!Loop`, `!Var`, `!If`, etc.
- Offers two output modes:
  - `generate`: Outputs the processed raw YAML.
  - `process`: Outputs structured data compatible with [Glazed](https://github.com/go-go-golems/glazed) (tables, JSON, etc.).

## Installation

You can build the `faker-generator` from source using Go:

```bash
# Clone the repository if you haven't already
# git clone <repo-url>
# cd <repo-path>/go-go-labs/cmd/apps/faker-generator

go build .
```

This will create an executable named `faker-generator` (or `faker-generator.exe` on Windows) in the current directory.

## Usage

The tool has two main commands: `generate` and `process`.

### `generate` Command

Outputs the processed raw YAML to standard output or a file.

```bash
./faker-generator generate --input-file <template.yaml> [flags]
```

**Common Flags:**

- `--input-file <file>`: **Required.** Path to the input YAML template.
- `--output <file>`: Write the raw YAML output to a file instead of stdout.
- `--vars <file>`: Load variables from a YAML file.
- `--set <key>=<value>`: Set individual variables.

**`generate` Examples:**

```bash
# Generate raw YAML from a simple template
./faker-generator generate --input-file examples/01-simple.yaml

# Generate YAML using variables and save to a file
echo "num_users: 2" > my_vars.yaml
./faker-generator generate --input-file examples/05-loop.yaml --vars my_vars.yaml --output generated_users.yaml

# Generate YAML setting a variable directly
./faker-generator generate --input-file examples/05-loop.yaml --set num_users=1
```

### `process` Command

Processes the template and outputs structured data using the Glazed library. This allows formatting the output as tables, JSON, CSV, etc.

```bash
./faker-generator process --input-file <template.yaml> [glazed flags]
```

**Common Flags:**

- `--input-file <file>`: **Required.** Path to the input YAML template.
- Supports standard Glazed flags for output formatting (`--output`, `--table-format`, `--fields`, etc.), data manipulation (`--select`, `--sort`), and templating. See Glazed documentation for details.
- Also supports `--vars` and `--set` like the `generate` command.

**`process` Examples:**

```bash
# Process template and output as default table
./faker-generator process --input-file examples/05-loop.yaml

# Process template and output as JSON
./faker-generator process --input-file examples/05-loop.yaml --output json

# Process template, select specific fields, and output as CSV
./faker-generator process --input-file examples/05-loop.yaml --fields id,email --output csv > users.csv

# Process template using variables
echo "num_users: 3" > my_vars.yaml
./faker-generator process --input-file examples/05-loop.yaml --vars my_vars.yaml
```

## YAML Templates and `!Faker*` Tags

Create YAML files (`.yaml` or `.yml`) and use `!Faker*` tags where you want random data inserted. The tool processes these tags using Emrichen.

### Input Examples

Here are a few examples of how to structure your input YAML:

**1. Basic Tags:**

```yaml
# examples/01-simple.yaml
user_profile:
  name: !FakerName
  email: !FakerEmail
  active: true
---
# A list of names
- !FakerName
- !FakerName
```

**2. Tags with Arguments:**

```yaml
# examples/02-with-args.yaml
security_details:
  weak_password: !FakerPassword # Default length


  strong_password: !FakerPassword
    minLength: 12
    maxLength: 20

random_numbers:
  default_int: !FakerInt # Default range 0-100


  custom_int: !FakerInt
    min: 1000
    max: 2000
```

**3. Using Loops and Variables:**

```yaml
# examples/05-loop.yaml (partial)
num_users: 3
user_roles: ["admin", "editor", "viewer"]

users: !Loop
  over: !Range $num_users
  as: i
  do:
    id: !Expr "$i + 1000"
    name: !FakerName
    email: !FakerEmail
    role: !FakerChoice $user_roles # Use variable for choices
```

### Available Tags

_(Note: This list reflects the currently implemented tags. Refer to `cmds/shared.go` for the most up-to-date list and implementation details.)_

- **`!FakerName`**: Generates a full name (e.g., "John Doe").
- **`!FakerEmail`**: Generates an email address.
- **`!FakerPassword`**: Generates an alphanumeric password.
  - `minLength` (int, default: 8)
  - `maxLength` (int, default: 16)
- **`!FakerInt`**: Generates a random integer within a range.
  - `min` (int, default: 0)
  - `max` (int, default: 100)
- **`!FakerFloat`**: Generates a random float64 within a range.
  - `min` (float64, default: 0.0)
  - `max` (float64, default: 1.0)
- **`!FakerLat`**: Generates a latitude.
  - `min` (float64, default: -90.0)
  - `max` (float64, default: 90.0)
- **`!FakerLong`**: Generates a longitude.
  - `min` (float64, default: -180.0)
  - `max` (float64, default: 180.0)
- **`!FakerChoice`**: Picks a random string from a list.
  - Can be a sequence: `!FakerChoice [a, b, c]`
  - Or a mapping: `!FakerChoice { choices: [a, b, c] }`
  - The `choices` list can be provided via `!Var`.
- **`!FakerUUID`**: Generates a hyphenated UUID v4.
- **`!FakerPhoneNumber`**: Generates a phone number.
- **`!FakerUsername`**: Generates a username.
- **`!FakerIPv4`**: Generates an IPv4 address.
- **`!FakerIPv6`**: Generates an IPv6 address.
- **`!FakerWord`**: Generates a random word.
- **`!FakerSentence`**: Generates a random sentence.
- **`!FakerParagraph`**: Generates random paragraphs.
  - `count` (int, default: 3): Number of paragraphs to generate.

_(Address-related tags like `!FakerCountry`, `!FakerCity`, `!FakerStreetAddress`, `!FakerZip` and the structured `!FakerAddress` are planned but might be commented out currently)._

### Examples Files

See the `examples/` directory for various tag usage patterns:

- `01-simple.yaml`: Basic tag usage.
- `02-with-args.yaml`: Tags with arguments (password length, number ranges).
- `03-location.yaml`: Latitude and longitude generation.
- `04-choice.yaml`: Using `!FakerChoice`.
- `05-loop.yaml`: Combining faker tags with Emrichen `!Loop` and `!Var`.
- `06-ids-and-network.yaml`: UUIDs, IPs, phone numbers, usernames.
- `07-text-generation.yaml`: Word, sentence, paragraph generation.

## Extending

To add new `!Faker*` tags:

1.  Open `cmds/shared.go`.
2.  Create a new handler function following the pattern `handleFakerXYZ(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error)`.
3.  Inside the handler, call the desired `faker.` function.
4.  Parse any required arguments from the `node` if it's a `yaml.MappingNode` (use helper functions like `parseIntArgument`).
5.  Convert the result to a `*yaml.Node` using `emrichen.ValueToNode`.
6.  Register the new handler function in the `fakerTags` map within `NewFakerInterpreter()`.
7.  Rebuild the application (`go build .`).

## Dependencies

- [go-faker/faker](https://github.com/go-faker/faker)
- [go-go-golems/go-emrichen](https://github.com/go-go-golems/go-emrichen)
- [go-go-golems/glazed](https://github.com/go-go-golems/glazed)
- [spf13/cobra](https://github.com/spf13/cobra)
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
