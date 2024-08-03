package shared

type Description struct {
	Name  string `yaml:"name"`
	Short string `yaml:"short"`
	Long  string `yaml:"long,omitempty"`
}

type ParameterType string

const (
	ParameterTypeString ParameterType = "string"

	// TODO(2023-02-13, manuel) Should the "default" of a stringFromFile be the filename, or the string?
	//
	// See https://github.com/go-go-golems/glazed/issues/137

	// ParameterTypeFile and ParameterTypeFileList are a more elaborate version that loads and parses
	// the file content and returns a list of FileData objects (or a single object in the case
	// of ParameterTypeFile).
	ParameterTypeFile     ParameterType = "file"
	ParameterTypeFileList ParameterType = "fileList"

	// TODO(manuel, 2023-09-19) Add some more types and maybe revisit the entire concept of loading things from files
	// - string (potentially from file if starting with @)
	// - string/int/float list from file is another useful type

	ParameterTypeObjectListFromFile  ParameterType = "objectListFromFile"
	ParameterTypeObjectListFromFiles ParameterType = "objectListFromFiles"
	ParameterTypeObjectFromFile      ParameterType = "objectFromFile"
	ParameterTypeStringListFromFile  ParameterType = "stringListFromFile"
	ParameterTypeStringListFromFiles ParameterType = "stringListFromFiles"

	// ParameterTypeKeyValue signals either a string with comma separate key-value options,
	// or when beginning with @, a file with key-value options
	ParameterTypeKeyValue ParameterType = "keyValue"

	ParameterTypeInteger     ParameterType = "int"
	ParameterTypeFloat       ParameterType = "float"
	ParameterTypeBool        ParameterType = "bool"
	ParameterTypeDate        ParameterType = "date"
	ParameterTypeStringList  ParameterType = "stringList"
	ParameterTypeIntegerList ParameterType = "intList"
	ParameterTypeFloatList   ParameterType = "floatList"
	ParameterTypeChoice      ParameterType = "choice"
	ParameterTypeChoiceList  ParameterType = "choiceList"
)

type Parameter struct {
	Name       string       `yaml:"name"`
	ShortFlag  string       `yaml:"shortFlag,omitempty"`
	Type       string       `yaml:"type"`
	Help       string       `yaml:"help,omitempty"`
	Default    *interface{} `yaml:"default,omitempty"`
	Choices    []string     `yaml:"choices,omitempty"`
	Required   bool         `yaml:"required,omitempty"`
	IsArgument bool         `yaml:"-"`
}

type Command interface {
	GetDescription() Description
	GetParameters() []Parameter
}
