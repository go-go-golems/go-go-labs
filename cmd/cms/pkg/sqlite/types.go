package sqlite

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
)

var sqliteDataTypeMap = map[parameters.ParameterType]string{
	parameters.ParameterTypeString:             "TEXT",
	parameters.ParameterTypeStringFromFile:     "TEXT",
	parameters.ParameterTypeObjectListFromFile: "TEXT",
	parameters.ParameterTypeObjectFromFile:     "TEXT",
	parameters.ParameterTypeKeyValue:           "TEXT",
	parameters.ParameterTypeInteger:            "INTEGER",
	parameters.ParameterTypeFloat:              "REAL",
	parameters.ParameterTypeBool:               "INTEGER",
	parameters.ParameterTypeDate:               "TEXT",
	parameters.ParameterTypeStringList:         "TEXT",
	parameters.ParameterTypeIntegerList:        "TEXT",
	parameters.ParameterTypeFloatList:          "TEXT",
	parameters.ParameterTypeChoice:             "TEXT",
	parameters.ParameterTypeChoiceList:         "TEXT",
}

func sqliteDataType(t parameters.ParameterType) string {
	return sqliteDataTypeMap[t]
}

type TableData struct {
	TableName string
	Fields    []pkg.Field
}
