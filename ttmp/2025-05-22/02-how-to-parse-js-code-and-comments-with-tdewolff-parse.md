# Parsing JavaScript Code and Comments with tdewolff/parse

This document explains how to parse JavaScript code to extract docstrings and function signatures using the `github.com/tdewolff/parse/v2/js` package.

## Overview of tdewolff/parse

The `tdewolff/parse` library provides a JavaScript lexer and parser that supports ECMAScript 2020. Unlike Goja, which is primarily a JavaScript runtime with parser capabilities, tdewolff/parse is designed specifically for parsing and tokenizing JavaScript code.

Key features:

- Pure Go implementation (no CGO dependencies)
- ECMAScript 2020 support
- Preserves comments
- Fast and efficient tokenization
- AST (Abstract Syntax Tree) generation

## Parsing JavaScript Source Code

### 1. Setting up the Parser

First, we need to create a parser instance:

```go
import (
    "github.com/tdewolff/parse/v2"
    "github.com/tdewolff/parse/v2/js"
)

func parseJavaScript(sourceCode string) (*js.AST, error) {
    p := js.NewParser(parse.NewInputString(sourceCode))
    ast, err := p.Parse()
    return ast, err
}
```

### 2. Understanding the AST Structure

The tdewolff/parse library provides an AST representation that contains:

- `js.BlockStmt` - For code blocks
- `js.FuncDecl` - For function declarations
- `js.VarDecl` - For variable declarations
- `js.ExprStmt` - For expression statements
- Various other node types for different JavaScript constructs

Each node in the AST has position information that can be used to correlate with comments.

### 3. Working with Comments

The tdewolff/parse library preserves comments in the tokenization process. When parsing JavaScript code, you can access comments through the lexer:

```go
lexer := js.NewLexer(parse.NewInputString(sourceCode))
for {
    tt, text := lexer.Next()
    if tt == js.ErrorToken {
        break
    }

    if tt == js.CommentToken {
        // Process comment
        commentText := string(text)
        // ...
    }
}
```

## Extracting Docstrings and Function Signatures

### 1. Identifying JSDoc Comments

JSDoc comments typically follow this format:

```javascript
/**
 * This is a JSDoc comment
 * @param {string} name - Description of parameter
 * @returns {number} Description of return value
 */
function myFunction(name) {
  // function body
}
```

To identify these comments, we need to:

1. Look for comments that start with `/**`
2. Associate them with the following function or variable declaration

### 2. Collecting Function Signatures

To extract function signatures, we need to walk the AST and look for function declarations:

```go
func collectFunctions(ast *js.AST) []FunctionInfo {
    var functions []FunctionInfo

    // Walk the AST
    js.Walk(ast, func(node js.INode) bool {
        switch n := node.(type) {
        case *js.FuncDecl:
            // Extract function name and parameters
            name := n.Name.String()
            params := extractParameters(n.Params)
            functions = append(functions, FunctionInfo{
                Name:       name,
                Parameters: params,
                Position:   n.Pos(),
            })
        case *js.VarDecl:
            // Handle function expressions assigned to variables
            // ...
        }
        return true
    })

    return functions
}
```

### 3. Matching Comments to Functions

Since tdewolff/parse doesn't directly associate comments with AST nodes, we need to:

1. Collect all comments with their positions
2. Collect all functions with their positions
3. Match comments to functions based on proximity (a comment right before a function is likely its docstring)

```go
func matchCommentsToFunctions(comments []CommentInfo, functions []FunctionInfo) []FunctionWithDocs {
    var result []FunctionWithDocs

    for _, function := range functions {
        // Find the closest comment before this function
        var closestComment *CommentInfo
        for i := range comments {
            if comments[i].EndPos < function.Position &&
               (closestComment == nil || comments[i].EndPos > closestComment.EndPos) {
                closestComment = &comments[i]
            }
        }

        docString := ""
        if closestComment != nil {
            docString = closestComment.Text
        }

        result = append(result, FunctionWithDocs{
            Name:       function.Name,
            Parameters: function.Parameters,
            DocString:  docString,
        })
    }

    return result
}
```

## Handling Different JavaScript Patterns

When extracting function signatures, we need to handle various JavaScript patterns:

### 1. Named Function Declarations

```javascript
/**
 * Function description
 */
function myFunction(param1, param2) {
  // function body
}
```

### 2. Function Expressions Assigned to Variables

```javascript
/**
 * Function description
 */
const myFunction = function (param1, param2) {
  // function body
};

// Or arrow functions
const arrowFunction = (param1, param2) => {
  // function body
};
```

### 3. Method Definitions in Objects/Classes

```javascript
const obj = {
  /**
   * Method description
   */
  myMethod(param1, param2) {
    // method body
  },
};

class MyClass {
  /**
   * Method description
   */
  myMethod(param1, param2) {
    // method body
  }
}
```

## Generating Markdown Documentation

Once we have extracted function signatures and their associated docstrings, we can generate Markdown documentation:

````go
func generateMarkdown(functions []FunctionWithDocs) string {
    var md strings.Builder

    md.WriteString("# JavaScript Documentation\n\n")

    for _, fn := range functions {
        // Create function heading
        md.WriteString(fmt.Sprintf("## %s\n\n", fn.Name))

        // Add docstring
        if fn.DocString != "" {
            md.WriteString(fmt.Sprintf("%s\n\n", formatDocString(fn.DocString)))
        }

        // Add function signature
        md.WriteString("```javascript\n")
        md.WriteString(fmt.Sprintf("function %s(%s)\n", fn.Name, strings.Join(fn.Parameters, ", ")))
        md.WriteString("```\n\n")
    }

    return md.String()
}
````

## Conclusion

Using tdewolff/parse for extracting JavaScript docstrings and function signatures involves:

1. Parsing the JavaScript source code into an AST
2. Collecting all comments from the source code
3. Walking the AST to find function declarations and their signatures
4. Matching comments to functions based on their positions
5. Generating documentation from the collected information

The approach is more direct than using Goja since tdewolff/parse is specifically designed for parsing JavaScript code. It provides better support for modern JavaScript syntax and is more efficient for static analysis tasks like docstring extraction.
