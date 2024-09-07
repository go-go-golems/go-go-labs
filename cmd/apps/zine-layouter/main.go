package main

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/go-emrichen/pkg/emrichen"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/go-go-golems/go-go-labs/pkg/zinelayout"
)

var rootCmd = &cobra.Command{
	Use:   "zine-layouter [flags] [input_files...]",
	Short: "A tool to layout zine pages",
	Run:   run,
}

var (
	testFlag    bool
	specFile    string
	outputDir   string
	verboseFlag bool // Add this line
)

func init() {
	rootCmd.Flags().BoolVar(&testFlag, "test", false, "Generate test images instead of reading input images")
	rootCmd.Flags().StringVar(&specFile, "spec", "layout.yaml", "Path to the YAML specification file")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", ".", "Directory to save output images")
	rootCmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Enable verbose output") // Add this line
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// Read YAML file
	yamlFile, err := os.ReadFile(specFile)
	if err != nil {
		fmt.Printf("Error reading YAML file: %v\n", err)
		return
	}

	env := map[string]interface{}{}
	interpreter, err := emrichen.NewInterpreter(emrichen.WithVars(env),
		emrichen.WithFuncMap(sprig.TxtFuncMap()))
	if err != nil {
		fmt.Printf("Error creating Emrichen interpreter: %v\n", err)
		return
	}

	// Create an Emrichen interpreter
	if err != nil {
		fmt.Printf("Error creating Emrichen interpreter: %v\n", err)
		return
	}

	var doc interface{}
	err = yaml.Unmarshal(yamlFile, &doc)
	if err != nil {
		fmt.Printf("Error unmarshaling YAML file: %v\n", err)
		return
	}
	fmt.Printf("Unmarshaled YAML: %+v\n", doc)

	f, err := os.Open(specFile)
	if err != nil {
		fmt.Printf("Error opening spec file: %v\n", err)
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)

	// Process the YAML with Emrichen
	for {
		fmt.Println("Decoding")
		var document interface{}
		err = decoder.Decode(interpreter.CreateDecoder(&document))

		if err == io.EOF {
			fmt.Println("EOF")
			break
		}
		if err != nil {
			fmt.Printf("Error processing YAML with Emrichen: %v\n", err)
			return
		}

		if verboseFlag {
			fmt.Println("Input YAML:")
			fmt.Println(string(yamlFile))
			fmt.Println("Document:")
			fmt.Println(document)
		}

		if document == nil {
			fmt.Println("Nil")
			continue
		}

		// Marshal the processed YAML back to bytes
		processedYAMLBytes, err := yaml.Marshal(document)
		if err != nil {
			fmt.Printf("Error marshaling processed YAML: %v\n", err)
			return
		}

		if verboseFlag {
			fmt.Println("Processed YAML:")
			fmt.Println(string(processedYAMLBytes))
		}

		// Parse the processed YAML
		var zineLayout zinelayout.ZineLayout
		err = yaml.Unmarshal(processedYAMLBytes, &zineLayout)
		if err != nil {
			fmt.Printf("Error parsing processed YAML: %v\n", err)
			return
		}

		// Add this block to print verbose output
		if verboseFlag {
			fmt.Println("Parsed ZineLayout:")
			printZineLayout(zineLayout)
			fmt.Println()
		}

		// Read or generate input images
		var inputImages []image.Image
		if testFlag {
			inputImages, err = zinelayout.GenerateTestImages(len(zineLayout.OutputPages))
		} else {
			if len(args) == 0 {
				fmt.Println("Error: No input files provided")
				return
			}
			inputImages, err = readInputImages(args)
		}
		if err != nil {
			fmt.Printf("Error with input images: %v\n", err)
			return
		}

		// Check if all input images have the same size
		if !zinelayout.AllImagesSameSize(inputImages) {
			fmt.Println("Error: All input images must have the same size")
			return
		}

		// Create output images
		for _, outputPage := range zineLayout.OutputPages {
			fmt.Printf("Processing page %s\n", outputPage.ID)
			outputImage := zinelayout.CreateOutputImage(zineLayout.PageSetup, outputPage, inputImages)
			saveOutputImage(outputImage, filepath.Join(outputDir, outputPage.ID))
		}
	}

}

func readInputImages(inputFiles []string) ([]image.Image, error) {
	var images []image.Image

	for _, file := range inputFiles {
		if filepath.Ext(file) == ".png" {
			f, err := os.Open(file)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			img, err := png.Decode(f)
			if err != nil {
				return nil, err
			}
			images = append(images, img)
		}
	}

	return images, nil
}

func saveOutputImage(img image.Image, filename string) {
	fullPath := filename + ".png"
	f, err := os.Create(fullPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}

	// Get file info to retrieve size
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}

	fmt.Printf("Saved output image: %s (Size: %d bytes)\n", fullPath, fileInfo.Size())
}

// Add this new function to print the ZineLayout
func printZineLayout(zl zinelayout.ZineLayout) {
	fmt.Printf("Global:\n")
	fmt.Printf("  Margin: %+v\n", zl.Global.Margin)

	fmt.Printf("PageSetup:\n")
	fmt.Printf("  Orientation: %s\n", zl.PageSetup.Orientation)
	fmt.Printf("  GridSize: Rows: %d, Columns: %d\n", zl.PageSetup.GridSize.Rows, zl.PageSetup.GridSize.Columns)
	fmt.Printf("  Margins: %+v\n", zl.PageSetup.Margins)

	fmt.Printf("OutputPages:\n")
	for i, page := range zl.OutputPages {
		fmt.Printf("  Page %d:\n", i+1)
		fmt.Printf("    ID: %s\n", page.ID)
		fmt.Printf("    Margin: %+v\n", page.Margin)
		fmt.Printf("    Layout:\n")
		for j, layout := range page.Layout {
			fmt.Printf("      Layout %d:\n", j+1)
			fmt.Printf("        InputIndex: %d\n", layout.InputIndex)
			fmt.Printf("        Position: Row: %f, Column: %f\n", layout.Position.Row, layout.Position.Column)
			fmt.Printf("        Rotation: %d\n", layout.Rotation)
			fmt.Printf("        Margin: %+v\n", layout.Margin)
		}
	}
}
