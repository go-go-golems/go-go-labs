package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/go-emrichen/pkg/emrichen"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/go-go-golems/go-go-labs/pkg/zinelayout"
)

const (
	cornerBorderLength = 20 // Length of corner dashes in pixels
)

var rootCmd = &cobra.Command{
	Use:   "zine-layouter [flags] [input_files...]",
	Short: "A tool to layout zine pages",
	Run:   run,
}

var (
	testFlag          bool
	specFile          string
	outputDir         string
	verboseFlag       bool
	globalBorderFlag  *bool
	pageBorderFlag    *bool
	layoutBorderFlag  *bool
	innerBorderFlag   *bool
	borderColorString *string
	borderTypeString  *string
)

func init() {
	rootCmd.Flags().BoolVar(&testFlag, "test", false, "Generate test images instead of reading input images")
	rootCmd.Flags().StringVar(&specFile, "spec", "layout.yaml", "Path to the YAML specification file")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", ".", "Directory to save output images")
	rootCmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Enable verbose output")
	globalBorderFlag = rootCmd.Flags().Bool("global-border", false, "Draw a global border")
	pageBorderFlag = rootCmd.Flags().Bool("page-border", false, "Draw a page border")
	layoutBorderFlag = rootCmd.Flags().Bool("layout-border", false, "Draw layout borders")
	innerBorderFlag = rootCmd.Flags().Bool("inner-border", false, "Draw inner layout borders")
	borderColorString = rootCmd.Flags().String("border-color", "", "Border color in R,G,B,A format (0-255 for each)")
	borderTypeString = rootCmd.Flags().String("border-type", "", "Border type: plain, dotted, dashed, or corner")
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
	interpreter, err := emrichen.NewInterpreter(
		emrichen.WithVars(env),
		emrichen.WithFuncMap(sprig.TxtFuncMap()))
	if err != nil {
		fmt.Printf("Error creating Emrichen interpreter: %v\n", err)
		return
	}

	f, err := os.Open(specFile)
	if err != nil {
		fmt.Printf("Error opening spec file: %v\n", err)
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	decoder := yaml.NewDecoder(f)

	// Process the YAML with Emrichen
	for {
		var document interface{}
		err = decoder.Decode(interpreter.CreateDecoder(&document))

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error processing YAML with Emrichen: %v\n", err)
			return
		}

		if verboseFlag {
			fmt.Println("Input YAML:")
			fmt.Println(string(yamlFile))
		}

		if document == nil {
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

		// Override border settings from command-line flags only if they are set
		if cmd.Flags().Changed("global-border") {
			zineLayout.GlobalBorder = *globalBorderFlag
		}
		if cmd.Flags().Changed("page-border") {
			zineLayout.PageBorder = *pageBorderFlag
		}
		if cmd.Flags().Changed("layout-border") {
			zineLayout.LayoutBorder = *layoutBorderFlag
		}
		if cmd.Flags().Changed("inner-border") {
			zineLayout.InnerLayoutBorder = *innerBorderFlag
		}
		if cmd.Flags().Changed("border-color") {
			borderColor, err := parseBorderColor(*borderColorString)
			if err != nil {
				fmt.Printf("Error parsing border color: %v\n", err)
				return
			}
			zineLayout.BorderColor = zinelayout.CustomColor{RGBA: borderColor}
		}
		if cmd.Flags().Changed("border-type") {
			borderType, err := zinelayout.ParseBorderType(*borderTypeString)
			if err != nil {
				fmt.Printf("Error parsing border type: %v\n", err)
				return
			}
			zineLayout.BorderType = borderType
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
			totalInputImages := len(zineLayout.OutputPages) * zineLayout.PageSetup.GridSize.Columns * zineLayout.PageSetup.GridSize.Rows
			inputImages, err = zinelayout.GenerateTestImages(totalInputImages)
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
			outputImage, err := zineLayout.CreateOutputImage(outputPage, inputImages)
			if err != nil {
				fmt.Printf("Error creating output image for page %s: %v\n", outputPage.ID, err)
				continue
			}
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
			defer func(f *os.File) {
				_ = f.Close()
			}(f)

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
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

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

	fmt.Printf("Saved output image: %s (Size: %d bytes, Dimensions: %dx%d)\n", fullPath, fileInfo.Size(), img.Bounds().Dx(), img.Bounds().Dy())
}

func parseBorderColor(colorString string) (color.RGBA, error) {
	parts := strings.Split(colorString, ",")
	if len(parts) != 4 {
		return color.RGBA{}, fmt.Errorf("invalid color format, expected R,G,B,A")
	}

	var rgba [4]uint8
	for i, part := range parts {
		val, err := strconv.ParseUint(strings.TrimSpace(part), 10, 8)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("invalid color component: %s", part)
		}
		rgba[i] = uint8(val)
	}

	return color.RGBA{R: rgba[0], G: rgba[1], B: rgba[2], A: rgba[3]}, nil
}

func printZineLayout(zl zinelayout.ZineLayout) {
	fmt.Printf("PageSetup:\n")
	fmt.Printf("  GridSize: Rows: %d, Columns: %d\n", zl.PageSetup.GridSize.Rows, zl.PageSetup.GridSize.Columns)
	fmt.Printf("  Margin: %+v\n", zl.PageSetup.Margin)

	fmt.Printf("OutputPages:\n")
	for i, page := range zl.OutputPages {
		fmt.Printf("  Page %d:\n", i+1)
		fmt.Printf("    ID: %s\n", page.ID)
		fmt.Printf("    Margin: %+v\n", page.Margin)
		fmt.Printf("    Layout:\n")
		for j, layout := range page.Layout {
			fmt.Printf("      Layout %d:\n", j+1)
			fmt.Printf("        InputIndex: %d\n", layout.InputIndex)
			fmt.Printf("        Position: Row: %d, Column: %d\n", layout.Position.Row, layout.Position.Column)
			fmt.Printf("        Rotation: %d\n", layout.Rotation)
			fmt.Printf("        Margin: %+v\n", layout.Margin)
		}
	}

	fmt.Printf("Border Settings:\n")
	fmt.Printf("  GlobalBorder: %v\n", zl.GlobalBorder)
	fmt.Printf("  PageBorder: %v\n", zl.PageBorder)
	fmt.Printf("  LayoutBorder: %v\n", zl.LayoutBorder)
	fmt.Printf("  InnerLayoutBorder: %v\n", zl.InnerLayoutBorder)
	fmt.Printf("  BorderColor: R:%d G:%d B:%d A:%d\n", zl.BorderColor.R, zl.BorderColor.G, zl.BorderColor.B, zl.BorderColor.A)
	fmt.Printf("  BorderType: %s\n", zl.BorderType)
}
