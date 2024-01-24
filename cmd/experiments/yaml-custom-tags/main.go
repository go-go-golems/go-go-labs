package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Resolver interface {
	Process(node *yaml.Node) (*yaml.Node, error)
}

func main() {
	// Register custom tag resolvers
	type Person struct {
		FullName   string `yaml:"fullName"`
		CurrentAge int    `yaml:"currentAge"`
	}

	type Document struct {
		Person Person `yaml:"person"`
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	decoder := yaml.NewDecoder(f)
	interpreter, err := NewEmrichenInterpreter()
	if err != nil {
		panic(err)
	}

	for {
		var s_ Document

		err = decoder.Decode(interpreter.CreateDecoder(&s_))
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		fmt.Printf("document: %v\n", s_)

	}

	//type MyStructure struct {
	//	// this structure holds the values you want to load after processing
	//	// includes, e.g.
	//	Num int
	//}
	//var s MyStructure
	//err = os.Setenv("FOO", `{"num": 42}`)
	//if err != nil {
	//	panic("Error setting environment variable")
	//}
	//err = yaml.Unmarshal([]byte("!getValueFromEnv FOO"), &CustomTagProcessor{&s})
	//if err != nil {
	//	panic("Error encountered during unmarshalling")
	//}
	//
	//fmt.Printf("\nNum: %v", s.Num)
	//
	//err = yaml.Unmarshal([]byte("!include foo.yaml"), &CustomTagProcessor{&s})
	//if err != nil {
	//	panic("Error encountered during unmarshalling")
	//}
	//fmt.Printf("\nNum: %v", s.Num)
}
