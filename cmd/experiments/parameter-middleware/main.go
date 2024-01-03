package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
)

type ParameterDefinition struct {
	Name    string
	Default string
}

type ParsedParameter struct {
	Definition *ParameterDefinition
	Value      string
	Source     string
}

type Definitions map[string]*ParameterDefinition
type Parsed map[string]*ParsedParameter

type HandlerFunc func(definition Definitions, parsed Parsed) error

type Middleware func(f HandlerFunc) HandlerFunc

func (p Parsed) String() string {
	strings_ := []string{}
	for name, parsedParameter := range p {
		strings_ = append(strings_, fmt.Sprintf("%s=%s (%s)", name, parsedParameter.Value, parsedParameter.Source))
	}
	return strings.Join(strings_, ", ")
}

func (d Definitions) String() string {
	strings_ := []string{}
	for name, def := range d {
		strings_ = append(strings_, fmt.Sprintf("%s=%s", name, def.Default))
	}
	return strings.Join(strings_, ", ")
}

func (d Definitions) Clone() Definitions {
	ret := Definitions{}
	for name, def := range d {
		ret[name] = def
	}
	return ret
}

func WithOnlyDefinedDefaults(defaults Definitions) Middleware {
	return func(f HandlerFunc) HandlerFunc {
		return func(definitions Definitions, parsed Parsed) error {
			defer func() {
				fmt.Println("WithOnlyDefinedDefaults: defaults:", defaults.String(), "parsed:", parsed.String())
			}()

			err := f(definitions, parsed)
			if err != nil {
				return err
			}

			fmt.Println("WithOnlyDefinedDefaults:", definitions.String())

			for name, def := range defaults {
				if _, ok := definitions[name]; !ok {
					continue
				}

				if _, ok := parsed[name]; !ok {
					parsed[name] = &ParsedParameter{
						Definition: def,
						Value:      def.Default,
						Source:     "default-middleware",
					}
				}
			}
			return nil
		}
	}
}

func WithAllDefaults(definitions Definitions) Middleware {
	return func(f HandlerFunc) HandlerFunc {
		return func(definitions_ Definitions, parsed Parsed) error {
			defer func() {
				fmt.Println("WithAllDefaults: parsed", parsed.String())
			}()

			// make a copy of the definitions to be sure to set them all
			oldDefinitions := definitions.Clone()

			err := f(definitions_, parsed)
			if err != nil {
				return err
			}
			fmt.Println("WithAllDefaults:", definitions.String())

			for name, def := range oldDefinitions {
				if _, ok := parsed[name]; !ok {
					parsed[name] = &ParsedParameter{
						Definition: def,
						Value:      def.Default,
						Source:     "all-default",
					}
				}
			}
			return nil
		}
	}
}

func WithEnv(env map[string]string) Middleware {
	return func(f HandlerFunc) HandlerFunc {
		return func(definitions Definitions, parsed Parsed) error {
			defer func() {
				fmt.Println("WithEnv: parsed", parsed.String())
			}()

			err := f(definitions, parsed)
			if err != nil {
				return err
			}

			fmt.Println("WithEnv:", definitions.String())
			for name, def := range env {
				if _, ok := definitions[name]; !ok {
					continue
				}

				fmt.Println("WithEnv: setting", name, def)
				parsed[name] = &ParsedParameter{
					Definition: definitions[name],
					Value:      def,
					Source:     "env",
				}
			}
			return nil
		}
	}
}

func WithHiddenDefinitions(toHide []string) Middleware {
	return func(f HandlerFunc) HandlerFunc {
		return func(definitions Definitions, parsed Parsed) error {
			fmt.Println("WithHiddenDefinitions:", definitions.String())
			defer func() {
				fmt.Println("WithHiddenDefinitions: parsed", parsed.String())
			}()

			for _, name := range toHide {
				_, ok := definitions[name]

				if !ok {
					return errors.New("WithHiddenDefinitions: unknown parameter: " + name)
				}
				delete(definitions, name)
			}

			// remove definitions before further parsing
			return f(definitions, parsed)
		}
	}
}

func WithRestrictedParameters(allowed []string, m Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(definitions Definitions, parsed Parsed) error {
			restoredNext := func(_ Definitions, parsed Parsed) error {
				return next(definitions, parsed)
			}

			restricted := Definitions{}
			for _, name := range allowed {
				if _, ok := definitions[name]; !ok {
					return errors.New("WithRestrictedParameters: unknown parameter: " + name)
				}
				restricted[name] = definitions[name]
			}

			return m(restoredNext)(restricted, parsed)
		}
	}
}

func WithOverrides(overrides map[string]string) Middleware {
	return func(f HandlerFunc) HandlerFunc {
		return func(definitions Definitions, parsed Parsed) error {
			fmt.Println("WithOverrides:", definitions.String())
			defer func() {
				fmt.Println("WithOverrides: parsed", parsed.String())
			}()

			for name := range overrides {
				_, ok := definitions[name]

				if !ok {
					// not sure if we need to actually check here, although it would be weird to provide
					// overrides for parameters that were not parsed, usually?
					// except maybe for commands that don't have glazed but we have wrappers to override with glazed?
					// if at this point the parameter has been hidden, we should fail?
					// otherwise we'll have to delete it from the parsed array to avoid overriding it later on
					if _, ok := definitions[name]; !ok {
						return errors.New("WithOverrides: unknown parameter: " + name)
					}
					return errors.New("WithOverrides: unknown parameter: " + name)
				}
				delete(definitions, name)
			}

			err := f(definitions, parsed)
			if err != nil {
				return err
			}

			for name, override := range overrides {
				// no need to check for definition here, since we checked above
				parsed[name] = &ParsedParameter{
					Definition: definitions[name],
					Value:      override,
					Source:     "override",
				}
			}

			return nil
		}
	}
}

func runWithMiddlewares(middlewares []Middleware, definitions Definitions) error {
	parsed := Parsed{}

	handler := func(definitions Definitions, parsed Parsed) error {
		return nil
	}

	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	err := handler(definitions, parsed)
	if err != nil {
		return err
	}

	for name, parsedParameter := range parsed {
		println(name, parsedParameter.Value, parsedParameter.Source)
	}
	return nil
}

func main() {
	fixedDefinitions := Definitions{
		"foo": &ParameterDefinition{
			Name:    "foo",
			Default: "foo-default",
		},
		"bar": &ParameterDefinition{
			Name:    "bar",
			Default: "bar-default",
		},
		"baz": &ParameterDefinition{
			Name:    "baz",
			Default: "baz-default",
		},
		"qux": &ParameterDefinition{
			Name:    "qux",
			Default: "qux-default",
		},
		"quux": &ParameterDefinition{
			Name:    "quux",
			Default: "quux-default",
		},
	}

	env := map[string]string{
		"foo":  "foo-env",
		"quux": "quux-env",
		"qux":  "qux-env",
		//"bar": "bar-env",
	}

	overrides := map[string]string{
		"bar": "bar-override",
	}

	defaults := Definitions{
		"baz": &ParameterDefinition{
			Name:    "baz",
			Default: "baz2",
		},
		"quux": &ParameterDefinition{
			Name:    "quux",
			Default: "quux2",
		},
	}

	defaultsToBeOverridden := Definitions{
		"bar": &ParameterDefinition{
			Name:    "bar",
			Default: "bar2",
		},
	}

	definitions := fixedDefinitions.Clone()

	middlewares := []Middleware{
		WithEnv(env),
		WithHiddenDefinitions([]string{"qux"}),
		WithOnlyDefinedDefaults(defaultsToBeOverridden),
		// should override remove it from hidden definitions?
		WithOverrides(overrides),
		WithOnlyDefinedDefaults(defaults),
		WithAllDefaults(definitions),
	}

	err := runWithMiddlewares(middlewares, definitions)
	if err != nil {
		log.Warn().Err(err).Msg("failed to run with middlewares")
	}
	fmt.Println("")

	definitions = fixedDefinitions.Clone()
	middlewares = []Middleware{
		WithOnlyDefinedDefaults(defaultsToBeOverridden),
		// should override remove it from hidden definitions?
		WithOverrides(overrides),
		WithOnlyDefinedDefaults(defaults),
		WithAllDefaults(definitions),
		WithEnv(env),
		//WithHiddenDefinitions([]string{"qux"}),
	}
	err = runWithMiddlewares(middlewares, definitions)
	if err != nil {
		log.Warn().Err(err).Msg("failed to run with middlewares")
	}
	fmt.Println("")

	definitions = fixedDefinitions.Clone()
	middlewares = []Middleware{
		WithOnlyDefinedDefaults(defaultsToBeOverridden),
		// should override remove it from hidden definitions?
		WithOnlyDefinedDefaults(defaults),
		WithAllDefaults(definitions),
		WithEnv(env),
		// this should fail
		WithOverrides(overrides),
		WithOverrides(overrides),
		//WithHiddenDefinitions([]string{"qux"}),
	}
	err = runWithMiddlewares(middlewares, definitions)
	if err != nil {
		log.Warn().Err(err).Msg("failed to run with middlewares")
	}
	fmt.Println("")

	fmt.Println("restricting env to foo")
	fmt.Println("---")
	definitions = fixedDefinitions.Clone()
	middlewares = []Middleware{
		WithRestrictedParameters([]string{"bar"}, WithOverrides(overrides)),
		WithRestrictedParameters([]string{"foo", "quux"}, WithEnv(env)),
		WithRestrictedParameters([]string{"bar"}, WithEnv(env)),
		WithAllDefaults(definitions),
	}
	err = runWithMiddlewares(middlewares, definitions)
	if err != nil {
		log.Warn().Err(err).Msg("failed to run with middlewares")
	}
	fmt.Println("")
}
