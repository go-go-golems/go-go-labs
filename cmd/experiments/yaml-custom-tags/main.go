package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Resolver interface {
	Resolve(node *yaml.Node) (*yaml.Node, error)
}

var tagResolvers = make(map[string]Resolver)

type Fragment struct {
	content *yaml.Node
}

func (f *Fragment) UnmarshalYAML(value *yaml.Node) error {
	var err error
	// process includes in fragments
	f.content, err = resolveTags(value)
	return err
}

type CustomTagProcessor struct {
	target interface{}
}

func (i *CustomTagProcessor) UnmarshalYAML(value *yaml.Node) error {
	resolved, err := resolveTags(value)
	if err != nil {
		return err
	}
	return resolved.Decode(i.target)
}

func resolveTags(node *yaml.Node) (*yaml.Node, error) {
	fmt.Printf("resolving tags for node: %v\n", node)
	for tag, resolver := range tagResolvers {
		if node.Tag == tag {
			fmt.Printf("resolving tag %s: %v\n", tag, node)
			ret, err := resolver.Resolve(node)
			if err != nil {
				fmt.Printf("error resolving tag %s: %v\n", tag, err)
				return nil, err
			}
			fmt.Printf("resolved tag %s: %v\n", tag, ret)
			return ret, err
		}
	}
	if node.Kind == yaml.SequenceNode || node.Kind == yaml.MappingNode {
		var err error
		for i := range node.Content {
			node.Content[i], err = resolveTags(node.Content[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return node, nil
}

type IncludeResolver struct {
}

func (i *IncludeResolver) Resolve(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!include on a non-scalar node")
	}
	file, err := os.ReadFile(node.Value)
	if err != nil {
		return nil, err
	}
	var f Fragment
	err = yaml.Unmarshal(file, &f)
	return f.content, err
}

type EnvResolver struct{}

func (e *EnvResolver) Resolve(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!getValueFromEnv on a non-scalar node")
	}
	value := os.Getenv(node.Value)
	if value == "" {
		return nil, fmt.Errorf("environment variable %v not set", node.Value)
	}
	var f Fragment
	err := yaml.Unmarshal([]byte(value), &f)
	return f.content, err
}

func AddResolvers(tag string, resolver Resolver) {
	tagResolvers[tag] = resolver
}

func main() {

	// Register custom tag resolvers
	AddResolvers("!include", &IncludeResolver{})
	AddResolvers("!getValueFromEnv", &EnvResolver{})
	resolver := NewVarResolver()
	AddResolvers("!Var", resolver)
	AddResolvers("!Defaults", resolver)

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

	for {
		var s_ Document

		err = decoder.Decode(&CustomTagProcessor{&s_})
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
