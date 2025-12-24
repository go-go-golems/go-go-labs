package animals

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// ParseCSV parses a CSV reader and extracts animal names from the first column.
// Empty lines and whitespace-only names are skipped.
// Names are trimmed but case is preserved.
func ParseCSV(r io.Reader) ([]string, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	var names []string
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse CSV at line %d", lineNum+1)
		}
		lineNum++

		if len(record) == 0 {
			continue
		}

		name := strings.TrimSpace(record[0])
		if name == "" {
			continue
		}

		names = append(names, name)
	}

	return names, nil
}

