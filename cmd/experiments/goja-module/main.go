package main

import (
	"fmt"
	"os"

	"github.com/dop251/goja"
)

func main() {
	vm := goja.New()
	bytes, err := os.ReadFile("bundle.js")
	if err != nil {
		fmt.Println("Error reading script:", err)
		return
	}

	res, err := vm.RunScript("bundle.js", string(bytes))
	if err != nil {
		fmt.Println("Error running script:", err)
		return
	}

	// Call a function from the module
	greetFunc, ok := res.Export().(func(string) string)
	if !ok {
		fmt.Println("Exported function is not of expected type")
		fmt.Println(res.Export())
		return
	}

	name := "World"
	resultStr := greetFunc(name)
	fmt.Println(resultStr)
}
