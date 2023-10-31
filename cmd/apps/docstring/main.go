package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:   "docscanner",
	Short: "DocScanner scans files for docstrings based on provided regular expressions",
}

var (
	files    []string
	output   string
	language string
)

func init() {
	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the provided files for docstrings",
		Run:   runScan,
	}
	scanCmd.Flags().StringSliceVarP(&files, "files", "f", []string{}, "Files to scan")
	scanCmd.Flags().StringVarP(&output, "output", "o", "json", "Output format: json or yaml")
	scanCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language: php, python, java, etc.")
	_ = viper.BindPFlag("files", scanCmd.Flags().Lookup("files"))
	_ = viper.BindPFlag("output", scanCmd.Flags().Lookup("output"))
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) {
	if len(files) == 0 {
		fmt.Println("No files provided. Use -f to provide file paths.")
		return
	}

	for _, file := range files {
		var scanner *Scanner

		if language == "" { // Infer language from file extension
			ext := filepath.Ext(file) // This will return ".php" for example
			scanner = getLanguageByExtension(ext)
			if scanner == nil {
				fmt.Printf("Unknown file extension: %s. Skipping file: %s\n", ext, file)
				continue
			}
		} else {
			scanner = languageSpecs[language].Scanner
			if scanner == nil {
				fmt.Printf("Unknown language: %s. Skipping file: %s\n", language, file)
				continue
			}
		}
		docs, err := scanner.ScanFile(file) // Adjust this line to scan based on the file's extension or other logic
		cobra.CheckErr(err)
		if output == "yaml" {
			y, err := yaml.Marshal(docs)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(y))
		} else {
			j, err := json.MarshalIndent(docs, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(j))
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
