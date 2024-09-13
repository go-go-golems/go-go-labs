package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/go-go-labs/pkg/zinelayout/parser"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	ppi        float64
	outputUnit string
)

var rootCmd = &cobra.Command{
	Use:   "distance-calculator [expression]",
	Short: "Calculate distances based on unit expressions",
	Long:  `A tool to calculate distances based on unit expressions, converting various units to pixels or other specified units.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		expression := args[0]
		p := &parser.ExpressionParser{
			PPI:   ppi,
			Debug: debug,
		}

		result, err := p.Parse(expression)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		uc := &parser.UnitConverter{PPI: ppi}
		outputValue, err := convertToOutputUnit(uc, result, outputUnit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Result: %.2f %s\n", outputValue, outputUnit)
	},
}

func init() {
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug output")
	rootCmd.Flags().Float64Var(&ppi, "ppi", 96, "Pixels per inch (default is 96)")
	rootCmd.Flags().StringVar(&outputUnit, "unit", "px", "Output unit (px, in, cm, mm, pt, pc, em)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func convertToOutputUnit(uc *parser.UnitConverter, value parser.Value, outputUnit string) (float64, error) {
	// First, convert the input value to pixels
	pixels, err := uc.ToPixels(value.Val, value.Unit)
	if err != nil {
		return 0, fmt.Errorf("error converting to pixels: %v", err)
	}

	// Then, convert from pixels to the desired output unit
	result, err := uc.FromPixels(pixels, outputUnit)
	if err != nil {
		return 0, fmt.Errorf("error converting to output unit: %v", err)
	}

	return result, nil
}
