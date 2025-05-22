Yes — but you have to use the **compiler-front-end part of Goja** (the `github.com/dop251/goja/parser` and `github.com/dop251/goja/ast` packages), not the VM interface you normally call with `vm.RunString`.
Here is the workflow that people usually follow:

---

### 1. Parse the JavaScript source and keep the comments

```go
import (
	"go/token"
	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

func parseJS(src []byte, filename string) (*ast.Program, *token.FileSet, error) {
	fset := token.NewFileSet()
	// parser.WithComments makes the parser attach a *ast.Comment list
	// to the resulting *ast.Program -— this is where we’ll pick up “doc-strings”.
	prog, err := parser.ParseFile(
		fset,
		filename,
		src,
		0,
		parser.WithComments,      // keep // and /** … */ blocks
	)
	return prog, fset, err
}
```

`WithComments` was added in March 2025 and is now the canonical way to keep JSDoc blocks and ordinary comments around the AST nodes.([Go Packages][1])

---

### 2. Walk the AST and collect every function-like thing

A function can appear in three main shapes:

| JS construct                       | AST node(s) to look for                             | How to get its name                                         |
| ---------------------------------- | --------------------------------------------------- | ----------------------------------------------------------- |
| `function foo(a,b){}`              | `*ast.FunctionDeclaration`                          | `fn.Id.Name`                                                |
| `const foo = function(a,b){}`      | `*ast.VariableDeclaration` → `*ast.FunctionLiteral` | the identifier on the left side of the `VariableDeclarator` |
| `const bar = (a,b)=>{}` (arrow fn) | same as above, `*ast.FunctionLiteral`               | idem                                                        |

```go
func collect(prog *ast.Program) []FunctionInfo {
	var out []FunctionInfo
	ast.Walk(func(n ast.Node) bool {
		switch v := n.(type) {

		case *ast.FunctionDeclaration:
			out = append(out, makeInfo(v, v.Id.Name))

		case *ast.VariableDeclaration:
			for _, d := range v.List {
				// let/const name = function(...) {…}  OR  arrow function
				if init, ok := d.Initializer.(*ast.FunctionLiteral); ok {
					name := d.Id.String()          // identifier on the LHS
					out = append(out, makeInfo(init, name))
				}
			}
		}
		return true
	}, prog)
	return out
}
```

`makeInfo` is a tiny helper that builds a `FunctionInfo` (your own struct) from an `*ast.FunctionLiteral`: grab `Parameters.List` for the signature and stash the node’s `Pos()/End()` so we can look up surrounding comments.

---

### 3. Attach the “doc-string” comment block

The AST does **not** attach comments directly to each node, so we usually take the start position of the function and walk **backwards** in the original file until we see the nearest comment whose ending line is immediately above the declaration.
Because we kept the comments in step 1 we can do this without extra parsing:

```go
func docStringFor(fn ast.Node, file *token.File, prog *ast.Program) string {
	fnLine := file.Position(fn.Pos()).Line
	for _, c := range prog.Comments {
		endLine := file.Position(c.End()).Line
		if endLine == fnLine-1 {        // comment is right above fn
			return string(c.Text)
		}
	}
	return ""
}
```

You can use a stricter heuristic (require `/**` so you only pick JSDoc, etc.).

---

### 4. Putting it together

```go
type FunctionInfo struct {
	Name       string
	Signature  []string   // param names
	Doc        string
	StartLine  int
	EndLine    int
}

func ParseFunctions(src []byte, filename string) ([]FunctionInfo, error) {
	prog, fset, err := parseJS(src, filename)
	if err != nil {
		return nil, err
	}
	f := fset.File(prog.Base)
	funcs := collect(prog)

	for i := range funcs {
		fnNode := funcs[i].astNode   // store node earlier
		funcs[i].Doc = docStringFor(fnNode, f, prog)
		pos := f.Position(fnNode.Pos())
		end := f.Position(fnNode.End())
		funcs[i].StartLine = pos.Line
		funcs[i].EndLine = end.Line
	}
	return funcs, nil
}
```

---

## Caveats & tips

| Limitation                               | What it means                                                                                                                                                                                                                                                                |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ES 2015+ syntax coverage**             | Goja’s parser is ES 5.1-first; most ES 6 patterns work, but stage-3/TypeScript features will not. For modern code-bases you may want to pre-run Babel/Swc/esbuild to down-transpile, or use a fully-featured parser such as `github.com/tdewolff/parse/v2/js`. ([GitHub][2]) |
| **No automatic comment-to-node binding** | You have to implement the “nearest-above” rule (shown above) yourself.                                                                                                                                                                                                       |
| **Signatures are *syntactic* only**      | Default values, rest parameters, overloads (TS) and JSDoc `@param` types need extra parsing if you want them.                                                                                                                                                                |
| **Interpreter route rarely helps**       | Executing the script in Goja and asking “what functions exist?” (`Object.getOwnPropertyNames(...)`) won’t recover comments, and it only sees what actually runs at runtime. AST parsing is the right tool here.                                                              |

---

### If Goja’s parser is not enough

* **tdewolff/parse/v2/js** – fastest pure-Go tokenizer, keeps comments, handles ES2024 grammar (no type-info).
* **flowchartsman/go-jsc** – Go wrapper around tree-sitter JavaScript with full comment range support.
* **esbuild’s parser API** – you can call it through its Go bindings and walk the json AST it returns.

But for straight ES5/ES6 code Goja’s built-in parser keeps your dependency tree small and avoids cgo.

---

### TL;DR

*Yes*, you can absolutely do it with Goja alone:

1. `parser.ParseFile(..., parser.WithComments)`
2. Walk the AST to gather `FunctionDeclaration` and `FunctionLiteral` nodes.
3. Use the preserved `prog.Comments` plus token positions to recover the doc-block.
4. Extract parameter names from `fn.ParameterList.List`.

The snippet above is enough to get you a slice of structs each containing the function name, parameter list, doc-string, and source span – which you can then marshal to JSON, feed into your RAG pipeline, or generate Markdown documentation from.

[1]: https://pkg.go.dev/github.com/dop251/goja/parser?utm_source=chatgpt.com "parser package - github.com/dop251/goja/parser - Go Packages"
[2]: https://github.com/dop251/goja/issues/478?utm_source=chatgpt.com "Modify nodes and export new JS code · Issue #478 · dop251/goja ..."
