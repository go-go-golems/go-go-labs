package view

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles holds all the styling for the application
var Styles = struct {
	// Base styles
	Title       lipgloss.Style
	Header      lipgloss.Style
	Pane        lipgloss.Style
	Selected    lipgloss.Style
	Highlight   lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	
	// Layout styles
	MainContainer lipgloss.Style
	Section       lipgloss.Style
	ActionBar     lipgloss.Style
	
	// Text styles
	Label       lipgloss.Style
	Value       lipgloss.Style
	Placeholder lipgloss.Style
	KeyBinding  lipgloss.Style
	
	// Special styles
	FilmIcon    lipgloss.Style
	ChemicalCol lipgloss.Style
}{
	// Base styles
	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Align(lipgloss.Center).
		Padding(0, 1),
	
	Header: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Align(lipgloss.Center),
	
	Pane: lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1).
		Margin(0, 1),
	
	Selected: lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 1),
	
	Highlight: lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true),
	
	Error: lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true),
	
	Success: lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true),
	
	// Layout styles
	MainContainer: lipgloss.NewStyle().
		Margin(1, 2).
		Padding(0),
	
	Section: lipgloss.NewStyle().
		Padding(1, 2).
		Margin(0, 0, 1, 0),
	
	ActionBar: lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1).
		Margin(0, 0, 0, 0),
	
	// Text styles
	Label: lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Bold(true),
	
	Value: lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true),
	
	Placeholder: lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true),
	
	KeyBinding: lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Background(lipgloss.Color("236")).
		Padding(0, 1),
	
	// Special styles
	FilmIcon: lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Bold(true),
	
	ChemicalCol: lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Width(24),
}

// RenderTitle renders the main application title
func RenderTitle() string {
	return Styles.Title.Render("🎞️  Film Development Calculator")
}

// RenderKeyBinding renders a key binding for help text
func RenderKeyBinding(key, description string) string {
	return Styles.KeyBinding.Render("["+key+"]") + " " + description
}

// RenderValue renders a value with proper styling
func RenderValue(value string) string {
	if value == "" || value == "--" {
		return Styles.Placeholder.Render("[ -- ]")
	}
	return Styles.Value.Render("[ " + value + " ]")
}

// RenderLabel renders a label with proper styling
func RenderLabel(label string) string {
	return Styles.Label.Render(label + ":")
}

// RenderFilmOption renders a film option with icon and description
func RenderFilmOption(key, name, ratings, description, icon string) string {
	keyPart := Styles.KeyBinding.Render("[" + key + "]")
	namePart := Styles.Highlight.Render(name)
	ratingsPart := Styles.Placeholder.Render("(" + ratings + ")")
	iconPart := Styles.FilmIcon.Render(icon)
	descPart := Styles.Placeholder.Render(description)
	
	return keyPart + " " + namePart + " " + ratingsPart + " " + iconPart + " " + descPart
}

// RenderSection renders a section with title and content
func RenderSection(title, content string) string {
	sectionTitle := Styles.Header.Render("─── " + title + " " + lipgloss.NewStyle().Render(lipgloss.PlaceHorizontal(80-len(title)-6, lipgloss.Right, "─")))
	return Styles.Section.Render(sectionTitle + "\n\n" + content)
}

// RenderChemicalColumn renders a chemical column for the results display
func RenderChemicalColumn(name, dilution, concentrate, water, time string) string {
	content := Styles.Highlight.Render(name) + "\n" +
		Styles.Label.Render(dilution+" dilution") + "\n" +
		Styles.Value.Render(concentrate+" conc") + "\n" +
		Styles.Value.Render(water+" water") + "\n" +
		Styles.Label.Render("Time: "+time)
	
	return Styles.ChemicalCol.Render(content)
}
