package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/go-go-golems/go-go-labs/cmd/apps/print-layouter/helpers"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Page struct {
	Width    string  `yaml:"width"`
	Height   string  `yaml:"height"`
	Margins  Margins `yaml:"margins"`
	Guides   []Guide `yaml:"guides"`
}

type Margins struct {
	Top    string `yaml:"top"`
	Right  string `yaml:"right"`
	Bottom string `yaml:"bottom"`
	Left   string `yaml:"left"`
}

type Guide struct {
	Type      string  `yaml:"type"`
	Position  string  `yaml:"position"`
	From      string  `yaml:"from"`
	Reference string  `yaml:"reference"`
	Gutter    string  `yaml:"gutter"`
	X         string  `yaml:"x"`
	Y         string  `yaml:"y"`
	Width     string  `yaml:"width"`
	Height    string  `yaml:"height"`
}

func (p *Page) ParseYAML(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, p)
}

func (p *Page) Print() {
	fmt.Printf("Page: %s x %s\n", p.Width, p.Height)
	fmt.Printf("Margins: Top: %s, Right: %s, Bottom: %s, Left: %s\n",
		p.Margins.Top, p.Margins.Right, p.Margins.Bottom, p.Margins.Left)
	fmt.Println("Guides:")
	for _, guide := range p.Guides {
		fmt.Printf("  - Type: %s, Position: %s, From: %s, Reference: %s, Gutter: %s\n",
			guide.Type, guide.Position, guide.From, guide.Reference, guide.Gutter)
		if guide.Type == "rect" {
			fmt.Printf("    X: %s, Y: %s, Width: %s, Height: %s\n",
				guide.X, guide.Y, guide.Width, guide.Height)
		}
	}
}

func (p *Page) DrawLayout(outputFile string) error {
	ppi := 300.0
	converter := helpers.UnitConverter{PPI: ppi}
	width, _ := converter.ToPixels(p.Width)
	height, _ := converter.ToPixels(p.Height)

	dc := gg.NewContext(int(width), int(height))
	dc.SetRGB(1, 1, 1) // White background
	dc.Clear()

	// Draw margins
	dc.SetColor(color.RGBA{255, 192, 203, 255}) // Pink color for margins
	dc.SetLineWidth(1)
	left, _ := converter.ToPixels(p.Margins.Left)
	top, _ := converter.ToPixels(p.Margins.Top)
	right, _ := converter.ToPixels(p.Margins.Right)
	bottom, _ := converter.ToPixels(p.Margins.Bottom)
	dc.DrawRectangle(left, top, width-left-right, height-top-bottom)
	dc.Stroke()

	// Draw guides
	dc.SetColor(color.RGBA{0, 0, 255, 255}) // Blue color for guides
	dc.SetLineWidth(0.5)
	for _, guide := range p.Guides {
		switch guide.Type {
		case "horizontal":
			y, _ := converter.ToPixels(guide.Position)
			dc.DrawLine(0, y, width, y)
			dc.Stroke()
		case "vertical":
			x, _ := converter.ToPixels(guide.Position)
			dc.DrawLine(x, 0, x, height)
			dc.Stroke()
		case "rect":
			x, _ := converter.ToPixels(guide.X)
			y, _ := converter.ToPixels(guide.Y)
			w, _ := converter.ToPixels(guide.Width)
			h, _ := converter.ToPixels(guide.Height)
			dc.DrawRectangle(x, y, w, h)
			dc.Stroke()
		}
	}

	return dc.SavePNG(outputFile)
}

func mmToPixels(mm float64, ppi float64) float64 {
	return mm * ppi / 25.4
}

func parseDimension(dim string) float64 {
	var value float64
	fmt.Sscanf(dim, "%fmm", &value)
	return value
}

func main() {
	var printFlag bool
	var rootCmd = &cobra.Command{
		Use:   "print-layouter [yaml file]",
		Short: "Print or draw layout from YAML file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			yamlFile := args[0]
			page := &Page{}
			err := page.ParseYAML(yamlFile)
			if err != nil {
				log.Fatalf("Error parsing YAML: %v", err)
			}

			if printFlag {
				page.Print()
			} else {
				outputFile := yamlFile[:len(yamlFile)-len(".yaml")] + ".png"
				err = page.DrawLayout(outputFile)
				if err != nil {
					log.Fatalf("Error drawing layout: %v", err)
				}
				fmt.Printf("Layout saved to %s\n", outputFile)
			}
		},
	}

	rootCmd.Flags().BoolVarP(&printFlag, "print", "p", false, "Print the parsed layout")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}