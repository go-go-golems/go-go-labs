package main

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/json"
	"os"
)

func main() {
	//	jsonStr := `{
	//        "foo": "foo
	//        bla",
	//        "bla": "foo\nbla"
	//}`
	//
	//	fmt.Println("Original JSON:")
	//	fmt.Println(jsonStr)
	//	fmt.Println("\nSanitized JSON:")
	//	fmt.Println(SanitizeJSONString(jsonStr))
	//
	//	jsonStr = `
	//{
	//    "example": "This is a \"quote\" inside the string.
	//    Here is a new line.",
	//    "normal": "This is without any issues.",
	//    "escaped": "This has a \\\"quote\\\" and newline
	//    here."
	//}
	//`
	//	fmt.Println("Original JSON:")
	//	fmt.Println(jsonStr)
	//	fmt.Println("\nSanitized JSON:")
	//	fmt.Println(SanitizeJSONString(jsonStr))

	// open cmd/sanitize/test/test.md and read it
	v, err := os.ReadFile("cmd/sanitize/test/test2.md")
	if err != nil {
		panic(err)
	}

	fmt.Println("Original markdown:")
	fmt.Println(string(v))

	jsonBlocks := json.ExtractJSON(string(v))
	for _, block := range jsonBlocks {
		fmt.Println("JSON block:")
		fmt.Println(block)
		fmt.Println("Sanitized JSON block:")
		fmt.Println(json.SanitizeJSONString(block))
	}

}
