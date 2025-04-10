package cmds

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/go-go-golems/go-emrichen/pkg/emrichen"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// --- Custom Tag Handlers ---

// Helper to parse mapping node arguments
func parseStringArgument(node *yaml.Node, key string) (string, bool) {
	if node.Kind != yaml.MappingNode {
		return "", false
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			if node.Content[i+1].Kind == yaml.ScalarNode {
				return node.Content[i+1].Value, true
			}
			return "", false // Key found but value is not a scalar
		}
	}
	return "", false // Key not found
}

func parseIntArgument(node *yaml.Node, key string) (int, bool, error) {
	s, ok := parseStringArgument(node, key)
	if !ok {
		return 0, false, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, true, errors.Wrapf(err, "failed to parse integer argument '%s'", key)
	}
	return v, true, nil
}

func parseFloatArgument(node *yaml.Node, key string) (float64, bool, error) {
	s, ok := parseStringArgument(node, key)
	if !ok {
		return 0, false, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, true, errors.Wrapf(err, "failed to parse float argument '%s'", key)
	}
	return v, true, nil
}

func parseStringListArgument(node *yaml.Node, key string) ([]string, bool) {
	if node.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			if node.Content[i+1].Kind == yaml.SequenceNode {
				var result []string
				for _, itemNode := range node.Content[i+1].Content {
					if itemNode.Kind == yaml.ScalarNode {
						result = append(result, itemNode.Value)
					} else {
						// Skip non-scalar items in the list
						fmt.Printf("Warning: Non-scalar item found in '%s' list, skipping.\n", key)
					}
				}
				return result, true
			}
			return nil, false // Key found but value is not a sequence
		}
	}
	return nil, false // Key not found
}

func handleFakerName(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	// TODO: Add locale handling if specified
	name := faker.Name()
	return emrichen.ValueToNode(name)
}

func handleFakerEmail(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	email := faker.Email()
	return emrichen.ValueToNode(email)
}

func handleFakerPassword(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	minLength := 8  // Default min length (changed from 6)
	maxLength := 16 // Default max length

	if node.Kind == yaml.MappingNode {
		minVal, minOk, minErr := parseIntArgument(node, "minLength")
		if minErr != nil {
			return nil, minErr
		}
		if minOk {
			minLength = minVal
		}

		maxVal, maxOk, maxErr := parseIntArgument(node, "maxLength")
		if maxErr != nil {
			return nil, maxErr
		}
		if maxOk {
			maxLength = maxVal
		}

		if minLength > maxLength {
			return nil, errors.New("minLength cannot be greater than maxLength for !FakerPassword")
		}
		if minLength <= 0 {
			return nil, errors.New("minLength must be positive for !FakerPassword")
		}
	} else if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!FakerPassword must be a scalar or a mapping node")
	}

	// Determine the exact length for the password within the min/max range
	length := minLength
	if maxLength > minLength {
		length += rand.Intn(maxLength - minLength + 1)
	}

	// Use faker.GetRandomStringLength to generate an alphanumeric string of the desired length.
	// This is simpler than faker.Password() and fits the min/max length requirement better.
	password := faker.Password(options.WithRandomStringLength(uint(length)))

	// NOTE: faker.GetRandomStringLength produces only letters and numbers.
	// If special characters are needed, a more complex generator would be required.

	return emrichen.ValueToNode(password)
}

func handleFakerInt(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	minVal := 0   // Default min
	maxVal := 100 // Default max

	if node.Kind == yaml.MappingNode {
		parsedMin, minOk, minErr := parseIntArgument(node, "min")
		if minErr != nil {
			return nil, minErr
		}
		if minOk {
			minVal = parsedMin
		}

		parsedMax, maxOk, maxErr := parseIntArgument(node, "max")
		if maxErr != nil {
			return nil, maxErr
		}
		if maxOk {
			maxVal = parsedMax
		}

		if minVal > maxVal {
			return nil, errors.New("min cannot be greater than max for !FakerInt")
		}
	} else if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!FakerInt must be a scalar or a mapping node")
	}

	// Use standard Go math/rand for integer range
	if minVal == maxVal {
		return emrichen.ValueToNode(minVal)
	}
	// rand.Intn panics if argument is <= 0
	randomRange := maxVal - minVal + 1
	if randomRange <= 0 {
		// This case should technically be caught by the minVal > maxVal check above,
		// but adding safety.
		return nil, errors.New("invalid range for !FakerInt: max must be >= min")
	}
	intValue := minVal + rand.Intn(randomRange)
	return emrichen.ValueToNode(intValue)
}

func handleFakerFloat(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	minVal := 0.0 // Default min
	maxVal := 1.0 // Default max

	if node.Kind == yaml.MappingNode {
		parsedMin, minOk, minErr := parseFloatArgument(node, "min")
		if minErr != nil {
			return nil, minErr
		}
		if minOk {
			minVal = parsedMin
		}

		parsedMax, maxOk, maxErr := parseFloatArgument(node, "max")
		if maxErr != nil {
			return nil, maxErr
		}
		if maxOk {
			maxVal = parsedMax
		}

		if minVal > maxVal {
			return nil, errors.New("min cannot be greater than max for !FakerFloat")
		}
	} else if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!FakerFloat must be a scalar or a mapping node")
	}

	// Use standard Go math/rand for float range
	floatValue := minVal + rand.Float64()*(maxVal-minVal)
	return emrichen.ValueToNode(floatValue)
}

func handleFakerLat(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	minVal := -90.0
	maxVal := 90.0

	if node.Kind == yaml.MappingNode {
		parsedMin, minOk, minErr := parseFloatArgument(node, "min")
		if minErr != nil {
			return nil, minErr
		}
		if minOk {
			minVal = parsedMin
		}

		parsedMax, maxOk, maxErr := parseFloatArgument(node, "max")
		if maxErr != nil {
			return nil, maxErr
		}
		if maxOk {
			maxVal = parsedMax
		}

		if minVal < -90.0 || maxVal > 90.0 || minVal > maxVal {
			return nil, errors.New("invalid range for !FakerLat: must be within [-90, 90] and min <= max")
		}
	} else if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!FakerLat must be a scalar or a mapping node")
	}

	// Generate within the custom range if specified, otherwise use faker default range
	lat := minVal + rand.Float64()*(maxVal-minVal)
	return emrichen.ValueToNode(lat)
}

func handleFakerLong(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	minVal := -180.0
	maxVal := 180.0

	if node.Kind == yaml.MappingNode {
		parsedMin, minOk, minErr := parseFloatArgument(node, "min")
		if minErr != nil {
			return nil, minErr
		}
		if minOk {
			minVal = parsedMin
		}

		parsedMax, maxOk, maxErr := parseFloatArgument(node, "max")
		if maxErr != nil {
			return nil, maxErr
		}
		if maxOk {
			maxVal = parsedMax
		}

		if minVal < -180.0 || maxVal > 180.0 || minVal > maxVal {
			return nil, errors.New("invalid range for !FakerLong: must be within [-180, 180] and min <= max")
		}
	} else if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!FakerLong must be a scalar or a mapping node")
	}

	// Generate within the custom range if specified, otherwise use faker default range
	long := minVal + rand.Float64()*(maxVal-minVal)
	return emrichen.ValueToNode(long)
}

func handleFakerChoice(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	var choices []string
	var ok bool

	if node.Kind == yaml.MappingNode {
		choices, ok = parseStringListArgument(node, "choices")
		if !ok {
			// Allow lookup of variable if 'choices' key is missing/not a list
			choicesNode, keyFound := findMappingKey(node, "choices")
			if keyFound && choicesNode.Kind == yaml.ScalarNode && choicesNode.Tag == "!Var" {
				resolvedChoicesNode, err := ei.Process(choicesNode)
				if err != nil {
					return nil, errors.Wrap(err, "failed to resolve choices variable for !FakerChoice")
				}
				if resolvedChoicesNode.Kind == yaml.SequenceNode {
					choices = make([]string, 0, len(resolvedChoicesNode.Content))
					for _, itemNode := range resolvedChoicesNode.Content {
						if itemNode.Kind == yaml.ScalarNode {
							choices = append(choices, itemNode.Value)
						} else {
							return nil, errors.New("!FakerChoice resolved variable sequence must contain only scalar string values")
						}
					}
					ok = true
				} else {
					return nil, errors.New("!FakerChoice variable 'choices' must resolve to a sequence of strings")
				}
			} else {
				return nil, errors.New("!FakerChoice mapping node requires a 'choices' key with a list of strings or a !Var resolving to one")
			}
		}
	} else if node.Kind == yaml.SequenceNode {
		// Allow !FakerChoice [a, b, c]
		choices = make([]string, 0, len(node.Content))
		for _, itemNode := range node.Content {
			if itemNode.Kind == yaml.ScalarNode {
				choices = append(choices, itemNode.Value)
			} else {
				// Before failing, check if it's a !Var resolving to a sequence
				if itemNode.Tag == "!Var" {
					resolvedVarNode, err := ei.Process(itemNode)
					if err != nil {
						return nil, errors.Wrap(err, "failed to resolve variable in !FakerChoice sequence")
					}
					if resolvedVarNode.Kind == yaml.SequenceNode {
						// Append items from the resolved sequence
						for _, resolvedItem := range resolvedVarNode.Content {
							if resolvedItem.Kind == yaml.ScalarNode {
								choices = append(choices, resolvedItem.Value)
							} else {
								return nil, errors.New("!FakerChoice resolved variable sequence must contain only scalar string values")
							}
						}
						continue // Continue to the next item in the original sequence
					} else {
						return nil, errors.New("!FakerChoice sequence item !Var must resolve to a sequence")
					}
				}
				// If not scalar and not a resolvable !Var, then it's an error
				return nil, errors.New("!FakerChoice sequence node must contain only scalar string values or !Var tags resolving to sequences")
			}
		}
		ok = true

	} else {
		return nil, errors.New("!FakerChoice requires a sequence node or a mapping node with a 'choices' key")
	}

	if !ok || len(choices) == 0 {
		return nil, errors.New("!FakerChoice requires a non-empty list of 'choices'")
	}

	// Use standard Go math/rand to pick a random index
	chosenIndex := rand.Intn(len(choices))
	chosenValue := choices[chosenIndex]
	return emrichen.ValueToNode(chosenValue)
}

// Helper to find a key in a mapping node (used by !FakerChoice)
func findMappingKey(node *yaml.Node, key string) (*yaml.Node, bool) {
	if node.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1], true
		}
	}
	return nil, false
}

// Add new handlers here

func handleFakerUUID(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	uuid := faker.UUIDHyphenated()
	return emrichen.ValueToNode(uuid)
}

func handleFakerPhoneNumber(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	// TODO: Add locale handling/formatting options
	phone := faker.Phonenumber()
	return emrichen.ValueToNode(phone)
}

func handleFakerUsername(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	username := faker.Username()
	return emrichen.ValueToNode(username)
}

// func handleFakerCountry(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
// 	// TODO: Add locale handling
// 	country := faker.Country()
// 	return emrichen.ValueToNode(country)
// }

// func handleFakerCity(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
// 	// TODO: Add locale handling
// 	city := faker.City()
// 	return emrichen.ValueToNode(city)
// }

// func handleFakerStreetAddress(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
// 	// TODO: Add locale handling
// 	street := faker.StreetAddress()
// 	return emrichen.ValueToNode(street)
// }

// func handleFakerZip(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
// 	// TODO: Add locale handling
// 	zip := faker.Postcode()
// 	return emrichen.ValueToNode(zip)
// }

func handleFakerIPv4(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	ip := faker.IPv4()
	return emrichen.ValueToNode(ip)
}

func handleFakerIPv6(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	ip := faker.IPv6()
	return emrichen.ValueToNode(ip)
}

func handleFakerWord(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	word := faker.Word()
	return emrichen.ValueToNode(word)
}

func handleFakerSentence(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	// faker.Sentence() doesn't exist directly, let's use faker.Paragraph(1)
	sentence := faker.Paragraph() // Generates a single sentence paragraph
	return emrichen.ValueToNode(sentence)
}

func handleFakerParagraph(ei *emrichen.Interpreter, node *yaml.Node) (*yaml.Node, error) {
	count := 3 // Default number of paragraphs
	if node.Kind == yaml.MappingNode {
		parsedCount, ok, err := parseIntArgument(node, "count")
		if err != nil {
			return nil, err
		}
		if ok {
			count = parsedCount
		}
		if count <= 0 {
			return nil, errors.New("count must be positive for !FakerParagraph")
		}
	} else if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!FakerParagraph must be a scalar or a mapping node")
	}

	ret := ""
	for i := 0; i < count; i++ {
		ret += faker.Paragraph() + "\n"
	}
	return emrichen.ValueToNode(ret)
}

// --- Interpreter Factory ---

// NewFakerInterpreter creates a new Emrichen interpreter with the standard Faker tags registered.
func NewFakerInterpreter() (*emrichen.Interpreter, error) {
	// Ensure math/rand is seeded. A better place might be in the main application entry point.
	// rand.Seed(time.Now().UnixNano()) // Deprecated since Go 1.20
	// No explicit seeding needed for math/rand starting Go 1.20

	fakerTags := emrichen.TagFuncMap{
		"!FakerName":        handleFakerName,
		"!FakerEmail":       handleFakerEmail,
		"!FakerPassword":    handleFakerPassword,
		"!FakerInt":         handleFakerInt,
		"!FakerFloat":       handleFakerFloat,
		"!FakerLat":         handleFakerLat,
		"!FakerLong":        handleFakerLong,
		"!FakerChoice":      handleFakerChoice,
		"!FakerUUID":        handleFakerUUID,
		"!FakerPhoneNumber": handleFakerPhoneNumber,
		"!FakerUsername":    handleFakerUsername,
		// "!FakerCountry":       handleFakerCountry,
		// "!FakerCity":          handleFakerCity,
		// "!FakerStreetAddress": handleFakerStreetAddress,
		// "!FakerZip":           handleFakerZip,
		"!FakerIPv4":      handleFakerIPv4,
		"!FakerIPv6":      handleFakerIPv6,
		"!FakerWord":      handleFakerWord,
		"!FakerSentence":  handleFakerSentence,
		"!FakerParagraph": handleFakerParagraph,

		// Add more tags here by copying the pattern
	}

	ei, err := emrichen.NewInterpreter(
		emrichen.WithAdditionalTags(fakerTags),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create emrichen interpreter with faker tags")
	}
	return ei, nil
}
