package main

type LanguageSpec struct {
	Scanner  *Scanner
	FileExts []string
}

var languageSpecs = map[string]LanguageSpec{
	"php":        {NewScanner(`^\s*/\*\*`, `^\s*\*+/`, `\*\s+@(\w+)(.*)`, `^\s*\*`), []string{".php"}},
	"python":     {NewScanner(`^"""`, `^"""`, `^:param \w+: (.*)`, ``), []string{".py"}},
	"java":       {NewScanner(`^\s*/\*\*`, `^\s*\*+/`, `^\s*\* @(\w+) (.*)`, `^\s*\*`), []string{".java"}},
	"javascript": {NewScanner(`^\s*/\*\*`, `^\s*\*+/`, `^\s*\* @(\w+) (.*)`, `^\s*\*`), []string{".js", ".jsx", ".ts", ".tsx"}},
	"go":         {NewScanner(`^\s*/\*\*`, `^\s*\*+/`, `^\s*\* @(\w+) (.*)`, `^\s*\*`), []string{".go"}},
	"csharp":     {NewScanner(`^\s*/\*\*`, `^\s*\*+/`, `^\s*\* @(\w+) (.*)`, `^\s*\*`), []string{".cs"}},
	"ruby":       {NewScanner(`^=begin`, `^=end`, `^\* @(\w+) (.*)`, `^\*`), []string{".rb"}},
}

func getLanguageByExtension(ext string) *Scanner {
	for _, spec := range languageSpecs {
		for _, fileExt := range spec.FileExts {
			if ext == fileExt {
				return spec.Scanner
			}
		}
	}
	return nil
}
