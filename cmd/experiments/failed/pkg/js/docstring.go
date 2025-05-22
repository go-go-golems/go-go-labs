package js

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

// CommentInfo stores information about a JavaScript comment
type CommentInfo struct {
	Text     string
	StartPos int  // Line position
	EndPos   int  // Line position
	IsJSDoc  bool // Indicates if this is a JSDoc comment (starts with /**)
	IsTopDoc bool // Indicates if this is a top-level docstring (first comment in file)
}

// FunctionInfo stores information about a JavaScript function
type FunctionInfo struct {
	Name       string
	Parameters []string
	Position   int // Line position
}

// FunctionWithDocs combines function information with its docstring
type FunctionWithDocs struct {
	Name       string
	Parameters []string
	DocString  string
	IsTopLevel bool // Indicates if this is a top-level function
	Filename   string
}

// ExtractDocstrings extracts docstrings and function signatures from JavaScript code
func ExtractDocstrings(data []byte, filename string, jsdocOnly bool) ([]FunctionWithDocs, error) {
	log.Debug().
		Str("filename", filename).
		Bool("jsdocOnly", jsdocOnly).
		Int("dataSize", len(data)).
		Msg("Starting docstring extraction")

	// Collect all comments
	comments, err := collectComments(data, jsdocOnly)
	if err != nil {
		log.Error().Err(err).Msg("Failed to collect comments")
		return nil, errors.Wrap(err, "failed to collect comments")
	}
	log.Debug().Int("commentsFound", len(comments)).Msg("Comments collected")

	// Collect all functions by manually traversing tokens
	functions, err := collectFunctions(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to collect functions")
		return nil, errors.Wrap(err, "failed to collect functions")
	}
	log.Debug().Int("functionsFound", len(functions)).Msg("Functions collected")

	// Match comments to functions
	result := matchCommentsToFunctions(comments, functions, filename)

	// Add top-level docstring if it exists
	var topLevelDocs []FunctionWithDocs
	for _, comment := range comments {
		if comment.IsTopDoc {
			log.Debug().
				Int("startPos", comment.StartPos).
				Str("commentPreview", truncateString(comment.Text, 30)).
				Msg("Found top-level documentation")

			topLevelDocs = append(topLevelDocs, FunctionWithDocs{
				Name:       "Overview",
				Parameters: nil,
				DocString:  comment.Text,
				IsTopLevel: true,
				Filename:   filename,
			})
			break
		}
	}

	finalResult := append(topLevelDocs, result...)
	log.Debug().
		Int("topLevelDocs", len(topLevelDocs)).
		Int("functionDocs", len(result)).
		Int("totalResults", len(finalResult)).
		Msg("Completed docstring extraction")

	return finalResult, nil
}

// collectComments extracts all comments from the JavaScript code
func collectComments(data []byte, jsdocOnly bool) ([]CommentInfo, error) {
	log.Debug().Msg("Starting collectComments")
	var comments []CommentInfo

	// Compute line offsets for position tracking
	lineOffsets := make([]int, 0)
	lineOffsets = append(lineOffsets, 0) // First line starts at offset 0

	for i, b := range data {
		if b == '\n' {
			lineOffsets = append(lineOffsets, i+1)
		}
	}
	log.Debug().Int("lines", len(lineOffsets)).Msg("Computed line offsets")

	// Function to convert byte offset to line number
	offsetToLine := func(offset int) int {
		// Binary search to find the line
		lo, hi := 0, len(lineOffsets)-1
		for lo <= hi {
			mid := (lo + hi) / 2
			if offset < lineOffsets[mid] {
				hi = mid - 1
			} else {
				lo = mid + 1
			}
		}
		return hi + 1 // Line numbers are 1-indexed
	}

	input := parse.NewInput(bytes.NewReader(data))
	lexer := js.NewLexer(input)

	isFirst := true // Track if this is the first comment in the file
	currentPos := 0

	for {
		tt, text := lexer.Next()
		if tt == js.ErrorToken {
			if lexer.Err() != io.EOF {
				log.Error().Err(lexer.Err()).Msg("Lexer error")
				return nil, lexer.Err()
			}
			log.Debug().Msg("Reached EOF in collectComments")
			break
		}

		// Track position based on token length
		startPos := currentPos
		currentPos += len(text)

		log.Debug().
			Str("token", tt.String()).
			Str("text", string(text)).
			Int("startPos", startPos).
			Int("currentPos", currentPos).
			Msg("Processing token")

		if tt == js.CommentToken {
			commentText := string(text)
			isJSDoc := strings.HasPrefix(commentText, "/**")

			// Skip if we only want JSDoc comments and this isn't one
			if jsdocOnly && !isJSDoc {
				log.Debug().Str("comment", commentText).Bool("isJSDoc", isJSDoc).Msg("Skipping non-JSDoc comment")
				continue
			}

			// Clean up the comment text
			commentText = cleanCommentText(commentText)

			// Get position information
			startLine := offsetToLine(startPos)
			endLine := offsetToLine(currentPos)

			log.Debug().
				Str("comment", commentText).
				Int("startLine", startLine).
				Int("endLine", endLine).
				Bool("isJSDoc", isJSDoc).
				Bool("isTopDoc", isFirst && len(strings.TrimSpace(commentText)) > 0).
				Msg("Found comment")

			comments = append(comments, CommentInfo{
				Text:     commentText,
				StartPos: startLine,
				EndPos:   endLine,
				IsJSDoc:  isJSDoc,
				IsTopDoc: isFirst && len(strings.TrimSpace(commentText)) > 0,
			})

			isFirst = false
		} else if tt != js.WhitespaceToken {
			// Any non-whitespace token means we're no longer at the top of the file
			isFirst = false
		}
	}

	log.Debug().Int("count", len(comments)).Msg("Finished collecting comments")
	return comments, nil
}

// cleanCommentText removes comment markers and normalizes whitespace
func cleanCommentText(text string) string {
	// Remove comment markers
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimPrefix(text, "*")
	text = strings.TrimSuffix(text, "*/")

	// Split into lines and process each line
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Remove leading asterisks and whitespace
		line = regexp.MustCompile(`^\s*\*\s?`).ReplaceAllString(line, "")
		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

// collectFunctions extracts all functions from the JavaScript code
func collectFunctions(data []byte) ([]FunctionInfo, error) {
	log.Debug().Msg("Starting collectFunctions")
	var functions []FunctionInfo

	lineOffsets := make([]int, 0)
	lineOffsets = append(lineOffsets, 0)
	for i, b := range data {
		if b == '\n' {
			lineOffsets = append(lineOffsets, i+1)
		}
	}
	log.Debug().Int("lines", len(lineOffsets)).Msg("Computed line offsets")

	offsetToLine := func(offset int) int {
		lo, hi := 0, len(lineOffsets)-1
		for lo <= hi {
			mid := (lo + hi) / 2
			if offset < lineOffsets[mid] {
				hi = mid - 1
			} else {
				lo = mid + 1
			}
		}
		return hi + 1
	}

	input := parse.NewInput(bytes.NewReader(data))
	lexer := js.NewLexer(input)

	type parsingState int
	const (
		stateNormal                   parsingState = iota
		stateSawFunctionKeyword                    // After 'function' keyword, expects Ident (name) or (
		stateSawFunctionName                       // After 'function foo', expects (
		stateSawVarLetConst                        // After 'const/var/let' keyword
		stateSawVarName                            // After 'const foo'
		stateSawVarEquals                          // After 'const foo ='
		stateSawVarEqualsAsync                     // After 'const foo = async'
		stateSawVarEqualsFunction                  // After 'const foo = function'
		stateSawAsyncKeyword                       // After 'async'
		stateSawAsyncFunctionName                  // After 'async foo' (method-like)
		stateSawAsyncFunctionKeyword               // After 'async function'
		stateInParams                              // Collecting params for regular/anon functions or methods
		stateInArrowParams                         // Collecting params for arrow func
		stateSawArrowParamsCloseParen              // After ) in arrow params. Expects =>
		stateMethodName                            // Identifier in obj/class context, expecting (
	)

	stateNames := map[parsingState]string{
		stateNormal:                   "Normal",
		stateSawFunctionKeyword:       "SawFunctionKeyword",
		stateSawFunctionName:          "SawFunctionName",
		stateSawVarLetConst:           "SawVarLetConst",
		stateSawVarName:               "SawVarName",
		stateSawVarEquals:             "SawVarEquals",
		stateSawVarEqualsAsync:        "SawVarEqualsAsync",
		stateSawVarEqualsFunction:     "SawVarEqualsFunction",
		stateSawAsyncKeyword:          "SawAsyncKeyword",
		stateSawAsyncFunctionName:     "SawAsyncFunctionName",
		stateSawAsyncFunctionKeyword:  "SawAsyncFunctionKeyword",
		stateInParams:                 "InParams",
		stateInArrowParams:            "InArrowParams",
		stateSawArrowParamsCloseParen: "SawArrowParamsCloseParen",
		stateMethodName:               "MethodName",
	}

	state := stateNormal
	var currentFunction FunctionInfo
	var paramText []byte
	var isAsync bool
	currentPos := 0

	for {
		tt, text := lexer.Next()
		tokenString := string(text)

		if tt == js.ErrorToken {
			if lexer.Err() != io.EOF {
				log.Error().Err(lexer.Err()).Msg("Lexer error")
				return nil, lexer.Err()
			}
			log.Debug().Msg("Reached EOF in collectFunctions")
			break
		}

		tokenStartPos := currentPos
		currentPos += len(text)

		// Skip whitespace and comments for state logic, but log them.
		if tt == js.WhitespaceToken || tt == js.CommentToken || tt == js.CommentLineTerminatorToken {
			log.Debug().
				Str("token", tt.String()).
				Str("text", "\""+tokenString+"\"").
				Str("state", stateNames[state]).
				Msg("Skipping token")
			continue
		}

		log.Debug().
			Str("token", tt.String()).
			Str("text", "\""+tokenString+"\"").
			Str("state", stateNames[state]).
			Int("startPos", tokenStartPos).
			Int("currentPos", currentPos).
			Msg("Processing token")

		switch state {
		case stateNormal:
			currentFunction = FunctionInfo{} // Reset for new potential function
			isAsync = false
			if tt == js.FunctionToken { // function foo() {} or function() {}
				currentFunction.Position = offsetToLine(tokenStartPos)
				state = stateSawFunctionKeyword
			} else if tt == js.ConstToken || tt == js.VarToken || tt == js.LetToken { // const foo = ...
				currentFunction = FunctionInfo{} // ensure fresh
				state = stateSawVarLetConst
			} else if tt == js.IdentifierToken && tokenString == "async" { // async function or async foo()
				currentFunction.Position = offsetToLine(tokenStartPos)
				isAsync = true
				state = stateSawAsyncKeyword
			} else if tt == js.IdentifierToken { // Potential method: foo() {}
				// This is the risky "method detection"
				// Keep it for now as it caught sample.js methods.
				// It will be refined or replaced if it causes too many issues.
				currentFunction.Name = tokenString
				currentFunction.Position = offsetToLine(tokenStartPos)
				state = stateMethodName
			}

		case stateSawFunctionKeyword: // After 'function' keyword
			if tt == js.IdentifierToken { // function foo
				currentFunction.Name = tokenString
				state = stateSawFunctionName
			} else if tt == js.PunctuatorToken && tokenString == "(" { // function (
				paramText = []byte{}
				state = stateInParams
			} else {
				log.Warn().Msgf("Unexpected token after 'function' keyword: %s", tokenString)
				state = stateNormal
			}

		case stateSawFunctionName: // After 'function foo'
			if tt == js.PunctuatorToken && tokenString == "(" {
				paramText = []byte{}
				state = stateInParams
			} else {
				log.Warn().Msgf("Unexpected token after function name: %s", tokenString)
				state = stateNormal
			}

		case stateSawVarLetConst: // After 'const', 'var', or 'let'
			if tt == js.IdentifierToken {
				currentFunction.Name = tokenString
				currentFunction.Position = offsetToLine(tokenStartPos)
				state = stateSawVarName
			} else {
				log.Warn().Msgf("Expected identifier after const/var/let, got: %s", tokenString)
				state = stateNormal
			}

		case stateSawVarName: // After 'const foo'
			if tt == js.PunctuatorToken && tokenString == "=" {
				state = stateSawVarEquals
			} else {
				log.Warn().Msgf("Expected '=' after variable name, got: %s", tokenString)
				state = stateNormal
			}

		case stateSawVarEquals: // After 'const foo ='
			if tt == js.FunctionToken { // const foo = function
				state = stateSawVarEqualsFunction
			} else if tt == js.IdentifierToken && tokenString == "async" { // const foo = async
				isAsync = true
				state = stateSawVarEqualsAsync
			} else if tt == js.PunctuatorToken && tokenString == "(" { // const foo = (
				paramText = []byte{}
				state = stateInArrowParams
			} else {
				log.Warn().Msgf("Unexpected token after '=', got: %s", tokenString)
				state = stateNormal
			}

		case stateSawVarEqualsAsync: // After 'const foo = async'
			if tt == js.FunctionToken { // const foo = async function
				state = stateSawVarEqualsFunction // Name, Pos, IsAsync already set
			} else if tt == js.PunctuatorToken && tokenString == "(" { // const foo = async (
				paramText = []byte{}
				state = stateInArrowParams // Name, Pos, IsAsync already set
			} else {
				log.Warn().Msgf("Unexpected token after 'async' in assignment, got: %s", tokenString)
				state = stateNormal
			}

		case stateSawVarEqualsFunction: // After 'const foo = function'
			if tt == js.PunctuatorToken && tokenString == "(" {
				paramText = []byte{}
				state = stateInParams // Name, Pos already set
			} else {
				log.Warn().Msgf("Expected '(' after 'function' in assignment, got: %s", tokenString)
				state = stateNormal
			}

		case stateSawAsyncKeyword: // After 'async' keyword (not in assignment)
			if tt == js.FunctionToken { // async function
				// currentFunction.Position was for 'async'. Update to 'function' or decide.
				// For now, let's say 'async function' pos is 'async'.
				state = stateSawAsyncFunctionKeyword
			} else if tt == js.IdentifierToken { // async foo() (method-like)
				currentFunction.Name = tokenString
				// Position is for 'async'. If we want 'foo', it's tokenStartPos.
				// Let's keep 'async' as the main position for now.
				state = stateSawAsyncFunctionName
			} else {
				log.Warn().Msgf("Unexpected token after 'async', got: %s", tokenString)
				state = stateNormal
			}

		case stateSawAsyncFunctionName: // After 'async foo'
			if tt == js.PunctuatorToken && tokenString == "(" {
				paramText = []byte{}
				state = stateInParams // Name, Pos, IsAsync set
			} else {
				log.Warn().Msgf("Expected '(' after async method name, got: %s", tokenString)
				state = stateNormal
			}

		case stateSawAsyncFunctionKeyword: // After 'async function'
			if tt == js.IdentifierToken { // async function foo
				currentFunction.Name = tokenString
				state = stateSawFunctionName // Reuses stateSawFunctionName, IsAsync is implicitly true
			} else if tt == js.PunctuatorToken && tokenString == "(" { // async function (
				paramText = []byte{}
				state = stateInParams // IsAsync is true
			} else {
				log.Warn().Msgf("Unexpected token after 'async function', got: %s", tokenString)
				state = stateNormal
			}

		case stateInParams:
			if tt == js.PunctuatorToken && tokenString == ")" {
				currentFunction.Parameters = extractParameters(string(paramText))
				if isAsync && !strings.HasPrefix(currentFunction.Name, "async ") {
					// Optionally qualify name, or add a field to FunctionInfo
					// currentFunction.Name = "async " + currentFunction.Name // Example
				}
				functions = append(functions, currentFunction)
				log.Debug().
					Str("name", currentFunction.Name).
					Strs("params", currentFunction.Parameters).
					Int("position", currentFunction.Position).
					Bool("isAsync", isAsync).
					Msg("Added function (std/method)")
				state = stateNormal
			} else {
				paramText = append(paramText, text...)
			}

		case stateInArrowParams:
			if tt == js.PunctuatorToken && tokenString == ")" {
				state = stateSawArrowParamsCloseParen
			} else {
				paramText = append(paramText, text...)
			}

		case stateSawArrowParamsCloseParen: // After 'const foo = (a, b)'
			if tt == js.ArrowToken { // =>
				currentFunction.Parameters = extractParameters(string(paramText))
				if isAsync && !strings.HasPrefix(currentFunction.Name, "async ") {
					// currentFunction.Name = "async " + currentFunction.Name // Example
				}
				functions = append(functions, currentFunction)
				log.Debug().
					Str("name", currentFunction.Name).
					Strs("params", currentFunction.Parameters).
					Int("position", currentFunction.Position).
					Bool("isAsync", isAsync).
					Msg("Added arrow function")
				state = stateNormal
			} else {
				log.Warn().Msgf("Expected '=>' after arrow function params, got %s", tokenString)
				state = stateNormal
			}

		case stateMethodName: // After an Identifier in stateNormal, e.g. 'constructor' or 'add' in a class/obj
			if tt == js.PunctuatorToken && tokenString == "(" {
				paramText = []byte{}
				state = stateInParams // Name, Pos already set from stateNormal's Identifier branch
			} else {
				// Not a method call, reset. This means the identifier was not a method name.
				log.Debug().Str("token", tokenString).Msgf("Identifier '%s' was not a method name, resetting state", currentFunction.Name)
				state = stateNormal
				// We need to re-process the current token 'tt' under stateNormal if it wasn't a (
				// This is tricky. For now, we lose this token if it wasn't '('.
				// A better way would be a pushback mechanism for the token or for the state machine to not consume.
				// Let's assume for now if it's not '(', it was a false positive.
			}

		default:
			log.Error().Msgf("Unhandled state: %s", stateNames[state])
			state = stateNormal
		}
	}

	log.Debug().Int("count", len(functions)).Msg("Finished collecting functions")
	return functions, nil
}

// extractParameters parses a parameter string like "(a, b, c)" and returns a slice of parameter names
func extractParameters(paramsText string) []string {
	// Remove outer parentheses
	paramsText = strings.TrimPrefix(paramsText, "(")
	paramsText = strings.TrimSuffix(paramsText, ")")

	if strings.TrimSpace(paramsText) == "" {
		return nil
	}

	// Split by commas
	parts := strings.Split(paramsText, ",")
	var params []string

	for _, part := range parts {
		// Extract parameter name (handle patterns like destructuring, default values)
		part = strings.TrimSpace(part)

		// Handle default value: "a = 5"
		if idx := strings.Index(part, "="); idx != -1 {
			part = strings.TrimSpace(part[:idx])
		}

		// Handle destructuring: "{a, b}"
		if strings.HasPrefix(part, "{") || strings.HasPrefix(part, "[") {
			// Just use the whole pattern for now
			params = append(params, part)
		} else {
			params = append(params, part)
		}
	}

	return params
}

// matchCommentsToFunctions associates comments with the nearest following function
func matchCommentsToFunctions(comments []CommentInfo, functions []FunctionInfo, filename string) []FunctionWithDocs {
	log.Debug().
		Int("comments", len(comments)).
		Int("functions", len(functions)).
		Str("filename", filename).
		Msg("Starting matchCommentsToFunctions")

	var result []FunctionWithDocs

	for i, function := range functions {
		// Find the closest comment before this function
		var closestComment *CommentInfo
		minDistance := int(^uint(0) >> 1) // Max int

		log.Debug().
			Int("functionIndex", i).
			Str("functionName", function.Name).
			Int("functionPosition", function.Position).
			Msg("Matching comments to function")

		for j := range comments {
			// Comment must come before the function
			if comments[j].EndPos < function.Position {
				distance := function.Position - comments[j].EndPos
				// Find the comment with the smallest distance to the function
				if distance < minDistance {
					minDistance = distance
					closestComment = &comments[j]

					log.Debug().
						Int("commentIndex", j).
						Int("commentStartPos", comments[j].StartPos).
						Int("commentEndPos", comments[j].EndPos).
						Int("distance", distance).
						Str("commentPreview", truncateString(comments[j].Text, 30)).
						Msg("Found closer comment")
				}
			}
		}

		docString := ""
		if closestComment != nil {
			docString = closestComment.Text
			log.Debug().
				Str("functionName", function.Name).
				Int("commentStartPos", closestComment.StartPos).
				Int("commentEndPos", closestComment.EndPos).
				Int("distance", minDistance).
				Str("docStringPreview", truncateString(docString, 30)).
				Msg("Matched comment to function")
		} else {
			log.Debug().
				Str("functionName", function.Name).
				Msg("No matching comment found for function")
		}

		result = append(result, FunctionWithDocs{
			Name:       function.Name,
			Parameters: function.Parameters,
			DocString:  docString,
			IsTopLevel: false,
			Filename:   filename,
		})
	}

	log.Debug().Int("matchedFunctions", len(result)).Msg("Finished matching comments to functions")
	return result
}

// truncateString truncates a string to the specified maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GenerateMarkdown creates a Markdown document from the extracted function information
func GenerateMarkdown(functions []FunctionWithDocs) string {
	log.Debug().Int("functions", len(functions)).Msg("Generating markdown for single file")

	var md strings.Builder

	// Add title
	md.WriteString("# JavaScript Documentation\n\n")

	// First add any top-level documentation
	topLevelCount := 0
	for _, fn := range functions {
		if fn.IsTopLevel {
			topLevelCount++
			log.Debug().Str("name", fn.Name).Msg("Adding top-level documentation")
			md.WriteString(fn.DocString)
			md.WriteString("\n\n")
		}
	}
	log.Debug().Int("topLevelCount", topLevelCount).Msg("Added top-level documentation")

	// Then add functions
	functionCount := 0
	for _, fn := range functions {
		if !fn.IsTopLevel {
			functionCount++
			log.Debug().
				Str("name", fn.Name).
				Strs("params", fn.Parameters).
				Bool("hasDocString", fn.DocString != "").
				Msg("Adding function documentation")

			// Create function heading
			md.WriteString(fmt.Sprintf("## %s\n\n", fn.Name))

			// Add docstring if it exists
			if fn.DocString != "" {
				md.WriteString(fmt.Sprintf("%s\n\n", fn.DocString))
			}

			// Add function signature
			md.WriteString("```javascript\n")
			md.WriteString(fmt.Sprintf("function %s(%s)\n", fn.Name, strings.Join(fn.Parameters, ", ")))
			md.WriteString("```\n\n")
		}
	}
	log.Debug().Int("functionCount", functionCount).Msg("Added function documentation")

	result := md.String()
	log.Debug().Int("markdownLength", len(result)).Msg("Generated markdown document")
	return result
}

// GenerateMarkdownForMultipleFiles creates a Markdown document for multiple JavaScript files
func GenerateMarkdownForMultipleFiles(fileResults map[string][]FunctionWithDocs) string {
	log.Debug().Int("files", len(fileResults)).Msg("Generating markdown for multiple files")

	var md strings.Builder

	md.WriteString("# JavaScript Documentation\n\n")

	// Table of contents
	md.WriteString("## Table of Contents\n\n")

	for file := range fileResults {
		// Clean the filename for display
		displayName := file
		if lastSlash := strings.LastIndex(displayName, "/"); lastSlash != -1 {
			displayName = displayName[lastSlash+1:]
		}

		md.WriteString(fmt.Sprintf("- [%s](#%s)\n", displayName, strings.ReplaceAll(displayName, ".", "")))
	}

	md.WriteString("\n")
	log.Debug().Msg("Generated table of contents")

	// File documentation
	for file, functions := range fileResults {
		// Clean the filename for display
		displayName := file
		if lastSlash := strings.LastIndex(displayName, "/"); lastSlash != -1 {
			displayName = displayName[lastSlash+1:]
		}

		log.Debug().
			Str("file", file).
			Str("displayName", displayName).
			Int("functions", len(functions)).
			Msg("Processing file")

		md.WriteString(fmt.Sprintf("## %s\n\n", displayName))

		// First add any top-level documentation
		topLevelCount := 0
		for _, fn := range functions {
			if fn.IsTopLevel {
				topLevelCount++
				log.Debug().
					Str("file", displayName).
					Str("name", fn.Name).
					Msg("Adding top-level documentation")

				md.WriteString(fn.DocString)
				md.WriteString("\n\n")
			}
		}
		log.Debug().
			Str("file", displayName).
			Int("topLevelCount", topLevelCount).
			Msg("Added top-level documentation")

		// Then add functions
		functionCount := 0
		for _, fn := range functions {
			if !fn.IsTopLevel {
				functionCount++
				log.Debug().
					Str("file", displayName).
					Str("name", fn.Name).
					Strs("params", fn.Parameters).
					Bool("hasDocString", fn.DocString != "").
					Msg("Adding function documentation")

				// Create function heading
				md.WriteString(fmt.Sprintf("### %s\n\n", fn.Name))

				// Add docstring if it exists
				if fn.DocString != "" {
					md.WriteString(fmt.Sprintf("%s\n\n", fn.DocString))
				}

				// Add function signature
				md.WriteString("```javascript\n")
				md.WriteString(fmt.Sprintf("function %s(%s)\n", fn.Name, strings.Join(fn.Parameters, ", ")))
				md.WriteString("```\n\n")
			}
		}
		log.Debug().
			Str("file", displayName).
			Int("functionCount", functionCount).
			Msg("Added function documentation")
	}

	result := md.String()
	log.Debug().Int("markdownLength", len(result)).Msg("Generated markdown document")
	return result
}
