package main

import "github.com/yuin/gopher-lua"

func main() {
	L := lua.NewState()
	defer L.Close()

	err := L.DoString(`print("Hello from lua!")`)
	if err != nil {
		panic(err)
	}
}
