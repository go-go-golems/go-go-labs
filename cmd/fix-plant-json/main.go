package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var dryRun bool

var rootCmd = &cobra.Command{
	Use:   "jsonfixer",
	Short: "Fixes the JSON files based on given rules.",
	Long:  `A CLI tool that corrects given JSON files based on specific rules.`,
	Run:   run,
}

func init() {
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", true, "If true, display the corrected JSON without saving changes to the files.")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func fixField(data map[string]interface{}, field string) {
	if value, ok := data[field].(string); ok {
		dataArray := strings.Split(strings.TrimSpace(value), ";")
		for i, v := range dataArray {
			dataArray[i] = strings.TrimSpace(v)
		}
		data[field] = dataArray
	}
}

// New function to fix growingZones
func fixGrowingZones(data map[string]interface{}) {
	// if it is a single string of the form X-Y, convert to an array of ints
	if zone, ok := data["growingZones"].(string); ok {
		// split on -
		zoneArray := strings.Split(zone, "-")
		newZoneArray := make([]int, 0)
		// convert to ints
		for _, z := range zoneArray {
			newZone, err := strconv.Atoi(z)
			if err != nil {
				fmt.Printf("Error converting growingZones to int: %s\n", err)
				return
			}
			newZoneArray = append(newZoneArray, newZone)
		}
		data["growingZones"] = newZoneArray
		return
	}

	if zones, ok := data["growingZones"].([]interface{}); ok {
		for i, zone := range zones {
			if num, ok := zone.(string); ok { // Check if it's a string, then convert to int
				zones[i], _ = strconv.Atoi(num)
			}
		}
		data["growingZones"] = zones
	}
}

func run(cmd *cobra.Command, args []string) {
	for _, filepath := range args {
		fileBytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			fmt.Printf("Error reading the file %s: %s\n", filepath, err)
			continue
		}

		var data map[string]interface{}
		err = json.Unmarshal(fileBytes, &data)
		if err != nil {
			fmt.Printf("Error unmarshalling the JSON from file %s: %s\n", filepath, err)
			continue
		}

		fixField(data, "sunlight")
		fixField(data, "growthRate")
		fixField(data, "waterNeeds")
		fixGrowingZones(data)

		fixedJSON, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Printf("Error marshalling the corrected JSON for file %s: %s\n", filepath, err)
			continue
		}

		if dryRun {
			fmt.Printf("Corrected JSON for file %s:\n%s\n", filepath, string(fixedJSON))
		} else {
			err := ioutil.WriteFile(filepath, fixedJSON, 0644)
			if err != nil {
				fmt.Printf("Error writing corrected JSON to file %s: %s\n", filepath, err)
				continue
			}
			fmt.Printf("File %s has been updated.\n", filepath)
		}
	}
}

func main() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
