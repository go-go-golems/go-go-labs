package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"log"
	"strings"
)

type RowData map[string]interface{}

func readExcelToJson(file string, sheet string) []RowData {
	f, err := excelize.OpenFile(file)
	if err != nil {
		log.Fatalf("unable to open file: %v", err)
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		log.Fatalf("unable to get rows: %v", err)
	}

	headers := rows[0]
	var data []RowData

	// Process headers to replace \n with --
	for i, header := range headers {
		headers[i] = strings.ReplaceAll(header, "\n", "--")
	}

	// Identify columns that should be arrays
	arrayColumns := make(map[int]bool)
	for _, row := range rows[1:] {
		for i, cell := range row {
			if strings.Contains(cell, "\n") {
				arrayColumns[i] = true
			}
		}
	}

	// Process cells
	for _, row := range rows[1:] {
		datum := RowData{}
		for i, cell := range row {
			key := headers[i]

			// If this column should be a list, make it so
			// Else, just assign the string as is
			if arrayColumns[i] {
				datum[key] = strings.Split(cell, "\n")
			} else {
				datum[key] = cell
			}
		}
		data = append(data, datum)
	}

	return data
}

func listSheets(file string) {
	f, err := excelize.OpenFile(file)
	if err != nil {
		log.Fatalf("unable to open file: %v", err)
	}

	sheets := f.GetSheetMap()
	for index, name := range sheets {
		fmt.Printf("Sheet %d: %s\n", index, name)
	}
}

func main() {
	var sheetName string

	var rootCmd = &cobra.Command{
		Use:   "excel2json",
		Short: "A converter tool for Excel to JSON",
		Long:  "Excel2json converts your excel sheet to a JSON file",
	}

	var convertCmd = &cobra.Command{
		Use:   "convert [filename]",
		Short: "Convert excel file to json",
		Long:  "Convert excel file to a series of JSON objects, based on sheet name provided",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			data := readExcelToJson(filename, sheetName)
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Fatalf("unable to marshal data: %v", err)
			}

			fmt.Println(string(jsonData))
		},
	}

	var listSheetsCmd = &cobra.Command{
		Use:   "list-sheets [filename]",
		Short: "List the names of all sheets in an excel file",
		Long:  "List the names of all sheets in the given excel file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			listSheets(filename)
		},
	}

	convertCmd.Flags().StringVarP(&sheetName, "sheet", "s", "", "Sheet name inside Excel")
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(listSheetsCmd)
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
