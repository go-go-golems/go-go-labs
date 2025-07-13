package types

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Chemical represents a chemical with its properties
type Chemical struct {
	Name      string   `json:"name"`
	Dilutions []string `json:"dilutions"`
	Default   string   `json:"default"`
	Time      string   `json:"time"`
	Type      string   `json:"type"`
	Capacity  string   `json:"capacity"`
}

// ChemicalDatabase contains all chemical information
type ChemicalDatabase struct {
	Chemicals map[string]Chemical `json:"chemicals"`
}

// GetChemical returns a chemical by name
func (cd *ChemicalDatabase) GetChemical(name string) (Chemical, bool) {
	chemical, ok := cd.Chemicals[name]
	return chemical, ok
}

// DilutionCalculation represents a chemical dilution calculation
type DilutionCalculation struct {
	Chemical    string `json:"chemical"`
	Dilution    string `json:"dilution"`
	TotalVolume int    `json:"total_volume"`
	Concentrate int    `json:"concentrate"`
	Water       int    `json:"water"`
	Time        string `json:"time"`
}

// ChemicalModel represents a chemical with its calculated values for display
type ChemicalModel struct {
	Name         string
	Dilution     string
	Concentrate  int
	Water        int
	Time         string
	IsCalculated bool
}

// NewChemicalModel creates a new ChemicalModel
func NewChemicalModel(name, dilution string, concentrate, water int, time string, isCalculated bool) ChemicalModel {
	return ChemicalModel{
		Name:         name,
		Dilution:     dilution,
		Concentrate:  concentrate,
		Water:        water,
		Time:         time,
		IsCalculated: isCalculated,
	}
}

// GetDefaultChemicals returns the default chemical models for display
func GetDefaultChemicals() []ChemicalModel {
	return []ChemicalModel{
		NewChemicalModel("ILFOSOL 3", "1+9", 0, 0, "--:--", false),
		NewChemicalModel("ILFOSTOP", "1+19", 0, 0, "0:10", false),
		NewChemicalModel("SPRINT FIXER", "1+4", 0, 0, "2:30", false),
	}
}

// GetCalculatedChemicals returns the calculated chemical models from DilutionCalculation
func GetCalculatedChemicals(calculations []DilutionCalculation) []ChemicalModel {
	if len(calculations) == 0 {
		return GetDefaultChemicals()
	}

	var chemicals []ChemicalModel
	for _, calc := range calculations {
		chemicals = append(chemicals, NewChemicalModel(
			calc.Chemical,
			calc.Dilution,
			calc.Concentrate,
			calc.Water,
			calc.Time,
			true,
		))
	}
	return chemicals
}

// ChemicalComponent represents a chemical with its own rendering logic
type ChemicalComponent struct {
	Name         string
	Dilution     string
	Concentrate  int
	Water        int
	Time         string
	IsCalculated bool
}

// Render renders the chemical component as a styled string
func (c *ChemicalComponent) Render() string {
	var b strings.Builder

	// Name
	b.WriteString(fmt.Sprintf("%-14s", c.Name))
	b.WriteString("\n")

	// Dilution
	b.WriteString(fmt.Sprintf("%-14s", c.Dilution+" dilution"))
	b.WriteString("\n")

	// Concentrate
	concStr := "--ml conc"
	if c.IsCalculated {
		concStr = fmt.Sprintf("%dml conc", c.Concentrate)
	}
	b.WriteString(fmt.Sprintf("%-14s", concStr))
	b.WriteString("\n")

	// Water
	waterStr := "--ml water"
	if c.IsCalculated {
		waterStr = fmt.Sprintf("%dml water", c.Water)
	}
	b.WriteString(fmt.Sprintf("%-14s", waterStr))
	b.WriteString("\n")

	// Time
	b.WriteString(fmt.Sprintf("%-14s", fmt.Sprintf("Time: %s", c.Time)))

	return b.String()
}

// RenderWithHighlight renders the chemical component with highlighting for calculated values
func (c *ChemicalComponent) RenderWithHighlight(highlightStyle lipgloss.Style) string {
	var b strings.Builder

	// Name
	b.WriteString(fmt.Sprintf("%-14s", c.Name))
	b.WriteString("\n")

	// Dilution
	b.WriteString(fmt.Sprintf("%-14s", c.Dilution+" dilution"))
	b.WriteString("\n")

	// Concentrate
	concStr := "--ml conc"
	if c.IsCalculated {
		concStr = fmt.Sprintf("%dml conc", c.Concentrate)
		concStr = highlightStyle.Render(concStr)
	}
	b.WriteString(fmt.Sprintf("%-14s", concStr))
	b.WriteString("\n")

	// Water
	waterStr := "--ml water"
	if c.IsCalculated {
		waterStr = fmt.Sprintf("%dml water", c.Water)
	}
	b.WriteString(fmt.Sprintf("%-14s", waterStr))
	b.WriteString("\n")

	// Time
	timeStr := fmt.Sprintf("Time: %s", c.Time)
	if c.IsCalculated {
		timeStr = highlightStyle.Render(timeStr)
	}
	b.WriteString(fmt.Sprintf("%-14s", timeStr))

	return b.String()
}

// NewChemicalComponent creates a new ChemicalComponent
func NewChemicalComponent(name, dilution string, concentrate, water int, time string, isCalculated bool) ChemicalComponent {
	return ChemicalComponent{
		Name:         name,
		Dilution:     dilution,
		Concentrate:  concentrate,
		Water:        water,
		Time:         time,
		IsCalculated: isCalculated,
	}
}

// ChemicalModelsToComponents converts ChemicalModel slice to ChemicalComponent slice
func ChemicalModelsToComponents(models []ChemicalModel) []ChemicalComponent {
	components := make([]ChemicalComponent, len(models))
	for i, model := range models {
		components[i] = NewChemicalComponent(
			model.Name,
			model.Dilution,
			model.Concentrate,
			model.Water,
			model.Time,
			model.IsCalculated,
		)
	}
	return components
}
