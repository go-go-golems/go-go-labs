package main

import (
	_ "embed"
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"log"
)

var BasicLexer = lexer.MustSimple([]lexer.SimpleRule{
	{"Comment", `^#[^\n]*\n`},
	{"EmptyLine", `\n`},
	{"Whitespace", `[ \t]+`},
	{"Ident", `[A-Za-z_][A-Za-z0-9_-]*`},
	{"CommandSign", `!!?`},
	{"VarName", `>`},
	{"GlobalName", `>>`},
	{"InputCol", `<`},
	{"InputTable", `<<`},
	{"Equal", `=`},
	{"BareString", `[^ \t\n!"'><]\S*`},
	{"EscapedString", `"([^"\\]*(\\.[^"\\]*)*)"`},
	{"Prompt", `(<< | \n)`},
	{"StringLine", `[^!#\n][^\n]*(\n|$) | \n`},
})

type Program struct {
	Lines []Line `@@*`
}

type Line struct {
	EmptyLine string    `@EmptyLine*`
	Commands  []Command `@@+`
	Prompt    *Prompt   `@@?`
}

type Prompt struct {
	Starter string   `@(@InputTable | @EmptyLine)`
	Lines   []string `{ (@StringLine+ | EmptyLine) }`
}

type Command struct {
	CommandSign string `@CommandSign`
	Name        string `@Ident`
	ArgList     []*Arg `( @@* )`
}

type Arg struct {
	KeyPair    *KeyPair    `@@`
	Literal    *string     `| ( @BareString | @EscapedString )`
	VarName    *VarName    `| @@`
	GlobalName *GlobalName `| @@`
	InputCol   *InputCol   `| @@`
	InputTable *InputTable `| @@`
}

type KeyPair struct {
	Key   string `@Ident`
	Value string `"=" ( @BareString | @EscapedString )`
}

type VarName struct {
	Identifier string `'>' @Ident`
}

type GlobalName struct {
	Identifier string `'>>' @Ident`
}

type InputCol struct {
	Identifier string `'<' @Ident`
}

type InputTable struct {
	Identifier string `'<<' @Ident`
}

//go:embed examples/test3.aipl
var exampleProgram string

func main() {
	symbols := BasicLexer.Symbols()
	log.Printf("symbols: %v\n", symbols)
	l, err := BasicLexer.LexString("", exampleProgram)
	findSymbol := func(t lexer.TokenType) string {
		for k, v := range symbols {
			if v == t {
				return k
			}
		}
		return fmt.Sprintf("unknown symbol: %d", t)
	}
	for tok, err := l.Next(); err == nil && tok.Type != lexer.EOF; tok, err = l.Next() {
		if tok.Type == symbols["Whitespace"] || tok.Type == symbols["Comment"] || tok.Type == symbols["EmptyLine"] {
			continue
		}
		log.Printf("Token: %s, Value: %s\n", findSymbol(tok.Type), tok.Value)
	}

	log.Printf("lexer error: %v\n", err)

	parser, err := participle.Build[Program](
		participle.Lexer(BasicLexer),
		participle.Elide("Comment", "Whitespace"))
	if err != nil {
		log.Fatalf("failed to build parser: %v", err)
	}

	log.Printf("Parsing:\n%s\n", exampleProgram)

	program, err := parser.ParseString("", exampleProgram)
	if err != nil {
		log.Fatalf("failed to parse: %v", err)
	}

	log.Printf("Parsed: %v\n", program)

	for _, line := range program.Lines {
		for _, command := range line.Commands {
			log.Printf("Command: %s\n", command.Name)
			for _, arg := range command.ArgList {
				switch {
				case arg.KeyPair != nil:
					log.Printf("KeyPair: %s=%s\n", arg.KeyPair.Key, arg.KeyPair.Value)
				case arg.Literal != nil:
					log.Printf("Literal: %s\n", *arg.Literal)
				case arg.VarName != nil:
					log.Printf("VarName: %s\n", arg.VarName.Identifier)
				case arg.GlobalName != nil:
					log.Printf("GlobalName: %s\n", arg.GlobalName.Identifier)
				case arg.InputCol != nil:
					log.Printf("InputCol: %s\n", arg.InputCol.Identifier)
				case arg.InputTable != nil:
					log.Printf("InputTable: %s\n", arg.InputTable.Identifier)
				default:
					log.Fatalf("unknown argument: %v", arg)
				}
			}
		}

		if line.Prompt != nil && len(line.Prompt.Lines) > 0 {
			log.Printf("Prompt: %v\n", line.Prompt)
		}
	}
}
