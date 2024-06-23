package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// Using interface{} to handle any YAML structure
type yamlFile map[string]interface{}

func main() {
	var sourcePath, targetPath, mergePath, key string
	flag.StringVar(&sourcePath, "source", "", "Source YAML file")
	flag.StringVar(&targetPath, "target", "", "Target YAML file")
	flag.StringVar(&mergePath, "path", "", "Target path within YAML for merging")
	flag.StringVar(&key, "key", "", "Key to match items in sequences")
	flag.Parse()

	// Read and unmarshal source file
	sourceData, err := readYAMLInterface(sourcePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading source file: %v\n", err)
		os.Exit(1)
	}

	// Read and unmarshal target file
	targetData, err := readYAML(targetPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading target file: %v\n", err)
		os.Exit(1)
	}

	// Merge the data
	mergedData, err := mergeData(targetData, sourceData, mergePath, key)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error merging data: %v\n", err)
		os.Exit(1)
	}

	// Convert the merged data back to YAML and write it out
	mergedYAML, err := yaml.Marshal(&mergedData)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error marshaling merged data: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(mergedYAML))
}

// Read a YAML file and unmarshal the contents
func readYAML(filepath string) (yamlFile, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var fileContents yamlFile
	err = yaml.Unmarshal(data, &fileContents)
	if err != nil {
		return nil, err
	}
	return fileContents, nil
}

// Read a YAML file and unmarshal the contents
func readYAMLInterface(filepath string) (interface{}, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var fileContents interface{}
	err = yaml.Unmarshal(data, &fileContents)
	if err != nil {
		return nil, err
	}
	return fileContents, nil
}

// Merge source data into target at the specified path based on the key
func mergeData(target yamlFile, source interface{}, mergePath, key string) (yamlFile, error) {
	mergePoint, ok := target[mergePath]
	if !ok {
		return nil, errors.Errorf("merge path %s not found in target", mergePath)
	}

	mergeList, ok := mergePoint.([]interface{})
	if !ok {
		return nil, errors.Errorf("merge point is not a sequence")
	}

	sourceList, ok := source.([]interface{})
	if !ok {
		// If source is not a list, it's assumed to be a single item
		sourceList = []interface{}{source}
	}

	for _, sourceItem := range sourceList {
		sourceMap, ok := sourceItem.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("source item is not a map")
		}

		sourceKeyValue, ok := sourceMap[key]
		if !ok {
			return nil, errors.Errorf("key '%s' not found in source item", key)
		}

		// Find and replace or append the item in the mergeList
		found := false
		for i, targetItem := range mergeList {
			targetMap, ok := targetItem.(map[string]interface{})
			if !ok {
				continue // Non-map items are ignored
			}

			if targetMap[key] == sourceKeyValue {
				// Merge sourceMap into targetMap
				for k, v := range sourceMap {
					targetMap[k] = v
				}
				mergeList[i] = targetMap
				found = true
				break
			}
		}

		if !found {
			// If no matching item, append the source item
			mergeList = append(mergeList, sourceItem)
		}
	}

	target[mergePath] = mergeList
	return target, nil
}
