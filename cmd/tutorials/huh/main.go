package main

import (
	"github.com/go-go-golems/clay/pkg/cmds"
	cmds2 "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
)

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
)

//nolint:unused
var (
	burger       string
	toppings     []string
	sauceLevel   int
	name         string
	instructions string
	discount     bool
)

// oh so we'll need custom widgets for stringlist, for textarea vs normal field, etc...

func main() {
	//tutorialForm()

	r := &cmds.RawCommandLoader{}
	filePath := "cmd/tutorials/huh/ps.yaml"
	fs_, filePath, err := loaders.FileNameToFsFilePath(filePath)
	if err != nil {
		panic(err)
	}
	cmds, err := r.LoadCommands(fs_, filePath, []cmds2.CommandDescriptionOption{}, []alias.Option{})
	if err != nil {
		panic(err)
	}

	if len(cmds) != 1 {
		panic(errors.Errorf("expected exactly one command, got %d", len(cmds)))
	}

	runFormForCommand(cmds[0].Description())
}

func runFormForCommand(description *cmds2.CommandDescription) {
	defaultLayer, ok := description.GetDefaultLayer()
	if !ok {
		panic(errors.Errorf("expected default layer"))
	}
	pds := defaultLayer.GetParameterDefinitions()

	// leave layers out of it for now
	//for _, layer := range description.Layers {
	//	for k, v := range layer.GetParameterDefinitions() {
	//		pds[k] = v
	//	}
	//}
	c := NewCommandForm()

	groups := []*huh.Group{}
	if len(description.Layout) > 0 {
		for _, section := range description.Layout {
			fields := []huh.Field{}

			for _, row := range section.Rows {
				for _, input := range row.Inputs {
					pd, ok := pds.Get(input.Name)
					if !ok {
						continue
					}

					field := c.makeFieldFromParameterDefinition(input, pd)
					if field != nil {
						fields = append(fields, field)
					}
				}
			}

			group := huh.NewGroup(fields...).Title(section.Title).Description(section.Description)

			groups = append(groups, group)
		}
	} else {
		fields := []huh.Field{}
		pds.ForEach(func(pd *parameters.ParameterDefinition) {
			options := []layout.Option{}
			for _, choice := range pd.Choices {
				options = append(options, layout.Option{
					Label: choice,
					Value: choice,
				})
			}
			field := c.makeFieldFromParameterDefinition(&layout.Input{
				Name:         pd.Name,
				Label:        "",
				Options:      options,
				DefaultValue: pd.Default,
				Help:         pd.Help,
				Validation:   nil,
				Condition:    nil,
			}, pd)
			if field != nil {
				fields = append(fields, field)
			}
		})
		group := huh.NewGroup(fields...)
		groups = append(groups, group)
	}

	form := huh.NewForm(groups...)

	err := form.Run()
	if err != nil {
		panic(err)
	}
}

type CommandForm struct {
	values map[string]interface{}
}

func NewCommandForm() *CommandForm {
	return &CommandForm{
		values: map[string]interface{}{},
	}
}

func (c *CommandForm) makeFieldFromParameterDefinition(
	input *layout.Input,
	pd *parameters.ParameterDefinition,
) huh.Field {
	switch pd.Type {
	case parameters.ParameterTypeInteger:
		ret := huh.NewInput().Title(pd.Name).Description(pd.Help)
		if pd.Default != nil {
			val := fmt.Sprintf("%d", pd.Default)
			c.values[name] = &val
			return ret.Value(&val)
		}
		c.values[name] = nil
		return ret
	case parameters.ParameterTypeFloat:
		val := fmt.Sprintf("%v", pd.Default)
		c.values[input.Name] = &val
		return huh.NewInput().Title(pd.Name).Description(pd.Help).
			Value(&val)
	case parameters.ParameterTypeString,
		parameters.ParameterTypeDate:
		val := ""
		if pd.Default != nil {
			val = (*pd.Default).(string)
		}
		c.values[input.Name] = &val
		return huh.NewInput().Title(pd.Name).Description(pd.Help).
			Value(&val)

	case parameters.ParameterTypeChoice:
		options := []huh.Option[string]{}
		for _, choice := range pd.Choices {
			options = append(options, huh.NewOption[string](choice, choice))
		}
		val := ""
		if pd.Default != nil {
			val = (*pd.Default).(string)
		}
		c.values[input.Name] = &val
		return huh.NewSelect[string]().Title(pd.Name).Description(pd.Help).
			Options(options...).
			Value(&val)

	case parameters.ParameterTypeChoiceList:
		options := []huh.Option[string]{}
		for _, choice := range pd.Choices {
			options = append(options, huh.NewOption[string](choice, choice))
		}
		vals := []string{}
		if pd.Default != nil {
			// clone the default values
			vals = append(vals, (*pd.Default).([]string)...)
		}
		c.values[input.Name] = &vals

		return huh.NewMultiSelect[string]().Title(pd.Name).Description(pd.Help).
			Options(options...).
			Value(&vals)

	case parameters.ParameterTypeBool,
		parameters.ParameterTypeFile,
		parameters.ParameterTypeFileList,
		parameters.ParameterTypeFloatList,
		parameters.ParameterTypeIntegerList,
		parameters.ParameterTypeStringList,
		parameters.ParameterTypeKeyValue,
		parameters.ParameterTypeObjectFromFile,
		parameters.ParameterTypeObjectListFromFiles,
		parameters.ParameterTypeObjectListFromFile,
		parameters.ParameterTypeStringFromFile,
		parameters.ParameterTypeStringFromFiles,
		parameters.ParameterTypeStringListFromFile,
		parameters.ParameterTypeStringListFromFiles:
		val := ""
		if pd.Default != nil {
			if _, ok := (*pd.Default).(string); ok {
				val = (*pd.Default).(string)
			} else {
				val = fmt.Sprintf("%v", pd.Default)
			}
		}
		c.values[input.Name] = &val
		return huh.NewInput().
			Title(fmt.Sprintf("%s (type: %s)", pd.Name, pd.Type)).
			Description(pd.Help).
			Value(&val)
	}
	return nil
}

//nolint:unused
func tutorialForm() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose your burger").
				Options(
					huh.NewOption("Charmburger Classic", "classic"),
					huh.NewOption("Charmburger Deluxe", "deluxe"),
					huh.NewOption("Chickwich", "chickwich"),
					huh.NewOption("Impossible Burger", "impossible"),
				).Value(&burger),
			huh.NewMultiSelect[string]().
				Title("Choose your toppings").
				Options(
					huh.NewOption("Lettuce", "lettuce"),
					huh.NewOption("Tomato", "tomato"),
					huh.NewOption("Onion", "onion"),
					huh.NewOption("Pickles", "pickles"),
					huh.NewOption("Cheese", "cheese"),
					huh.NewOption("Bacon", "bacon"),
				).
				Limit(4).
				Value(&toppings),

			huh.NewSelect[int]().
				Title("How much sauce?").
				Options(
					huh.NewOption("None", 0),
					huh.NewOption("A little", 1),
					huh.NewOption("A lot", 2),
				).
				Value(&sauceLevel),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("What's your name?").
				Value(&name).
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("name is required")
					}
					return nil
				}),

			huh.NewText().
				Title("Special instructions").
				CharLimit(300).
				Value(&instructions),

			huh.NewConfirm().
				Title("Would you like a discount?").
				Value(&discount),
		),
	)

	err := form.Run()
	if err != nil {
		panic(err)
	}

	if !discount {
		fmt.Println("No discount for you!")
	}
}
