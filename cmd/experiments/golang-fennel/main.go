package main

// https://www.phind.com/search?cache=kfde9siql5dyqiy7tvc0j4y9

import (
	"fmt"
	"github.com/yuin/gopher-lua"
	"os"
)

// Go function to be called from Fennel
func greet(l *lua.LState) int {
	name := l.ToString(1)
	fmt.Printf("Hello from go, %s!\n", name)
	return 0
}

func main() {
	L := lua.NewState()
	defer L.Close()

	L.OpenLibs()
	L.Register("greet", greet)

	// Load Fennel compiler
	fennelCode, err := os.ReadFile("fennel.lua")
	if err != nil {
		panic(err)
	}
	if err := L.DoString(string(fennelCode)); err != nil {
		panic(err)
	}

	// Define a Fennel script
	fennelScript := `
		(fn greet2 [name]
          (greet "Foobar")
		  (print (.. "Hello, " name "!")))
		(greet2 "World")
	`

	// Compile Fennel to Lua
	_ = runFennel(fennelScript, L)

	fennelScript = `
(macro when [condition ...]
` + "`" + `(if ,condition
	(do ,...)))

(when (> 5 3)
  (print "5 is greater than 3"))
`
	_ = runFennel(fennelScript, L)
}

func runFennel(fennelScript string, L *lua.LState) string {
	compileFennel := `
		local fennel = require("fennel")
		local luaCode = fennel.compileString([[
			` + fennelScript + `
		]])
		return luaCode
	`
	if err := L.DoString(compileFennel); err != nil {
		panic(err)
	}

	// Get the compiled Lua code
	luaCode := L.ToString(-1)
	L.Pop(1)

	// Execute the compiled Lua code
	if err := L.DoString(luaCode); err != nil {
		panic(err)
	}

	return compileFennel
}
