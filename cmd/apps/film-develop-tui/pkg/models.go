package pkg

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/charmbracelet/lipgloss"
)

// Film represents a film type with its properties
type Film struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	EIRatings   []int                     `json:"ei_ratings"`
	Times20C    map[string]map[int]string `json:"times_20c"`
	Description string                    `json:"description"`
	Icon        string                    `json:"icon"`
}

// FilmDatabase contains all available films
type FilmDatabase struct {
	Films map[string]Film `json:"films"`
}

// GetFilms returns all films as a slice
func (fd *FilmDatabase) GetFilms() []Film {
	films := make([]Film, 0, len(fd.Films))
	for _, film := range fd.Films {
		films = append(films, film)
	}
	return films
}

// GetFilmByID returns a film by its ID
func (fd *FilmDatabase) GetFilmByID(id string) (Film, bool) {
	film, ok := fd.Films[id]
	return film, ok
}

// TankSize represents tank size requirements
type TankSize struct {
	Format string `json:"format"`
	Rolls  int    `json:"rolls"`
	Volume int    `json:"volume"` // in ml
}

// TankDatabase contains tank size information
type TankDatabase struct {
	Sizes map[string]map[int]int `json:"sizes"`
}

// GetTankSize returns the tank size for given format and roll count
func (td *TankDatabase) GetTankSize(format string, rolls int) (int, bool) {
	formatSizes, ok := td.Sizes[format]
	if !ok {
		return 0, false
	}
	size, ok := formatSizes[rolls]
	return size, ok
}

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

// RollSetup represents the roll configuration
type RollSetup struct {
	Format35mm  int `json:"format_35mm"`
	Format120mm int `json:"format_120mm"`
	TotalVolume int `json:"total_volume"`
}

// String returns a human-readable description of the roll setup
func (rs *RollSetup) String() string {
	if rs.Format35mm > 0 && rs.Format120mm > 0 {
		return fmt.Sprintf("%dx 35mm + %dx 120mm", rs.Format35mm, rs.Format120mm)
	} else if rs.Format35mm > 0 {
		return fmt.Sprintf("%dx 35mm", rs.Format35mm)
	} else if rs.Format120mm > 0 {
		return fmt.Sprintf("%dx 120mm", rs.Format120mm)
	}
	return "No rolls"
}

// TotalRolls returns the total number of rolls
func (rs *RollSetup) TotalRolls() int {
	return rs.Format35mm + rs.Format120mm
}

// FixerState represents the current state of the fixer
type FixerState struct {
	CapacityPerLiter int `json:"capacity_per_liter"`
	UsedRolls        int `json:"used_rolls"`
	TotalCapacity    int `json:"total_capacity"`
}

// RemainingCapacity returns the remaining capacity of the fixer
func (fs *FixerState) RemainingCapacity() int {
	return fs.TotalCapacity - fs.UsedRolls
}

// CanProcess checks if the fixer can process the given number of rolls
func (fs *FixerState) CanProcess(rolls int) bool {
	return fs.RemainingCapacity() >= rolls
}

// UseFixer processes the given number of rolls
func (fs *FixerState) UseFixer(rolls int) {
	fs.UsedRolls += rolls
}

// ApplicationState represents the complete application state
type ApplicationState struct {
	SelectedFilm *Film                 `json:"selected_film"`
	SelectedEI   int                   `json:"selected_ei"`
	RollSetup    *RollSetup            `json:"roll_setup"`
	Dilution     string                `json:"dilution"`
	Calculations []DilutionCalculation `json:"calculations"`
	FixerState   *FixerState           `json:"fixer_state"`
	TimerState   *TimerState           `json:"timer_state"`
	FilmDB       *FilmDatabase         `json:"film_db"`
	TankDB       *TankDatabase         `json:"tank_db"`
	ChemicalDB   *ChemicalDatabase     `json:"chemical_db"`
}

// NewApplicationState creates a new application state with default values
func NewApplicationState() *ApplicationState {
	return &ApplicationState{
		Dilution: "1+9",
		FixerState: &FixerState{
			CapacityPerLiter: 24,
			UsedRolls:        0,
			TotalCapacity:    24,
		},
		FilmDB:     NewFilmDatabase(),
		TankDB:     NewTankDatabase(),
		ChemicalDB: NewChemicalDatabase(),
	}
}

// IsComplete checks if the application state is complete for calculation
func (as *ApplicationState) IsComplete() bool {
	return as.SelectedFilm != nil && as.SelectedEI > 0 && as.RollSetup != nil && as.RollSetup.TotalRolls() > 0
}

// GetDevelopmentTime returns the development time for the current selection
func (as *ApplicationState) GetDevelopmentTime() string {
	if as.SelectedFilm == nil {
		return "--:--"
	}

	dilutionTimes, ok := as.SelectedFilm.Times20C[as.Dilution]
	if !ok {
		return "--:--"
	}

	time, ok := dilutionTimes[as.SelectedEI]
	if !ok {
		return "--:--"
	}

	return time
}

// CalculateChemicals calculates the chemical dilutions for the current setup
func (as *ApplicationState) CalculateChemicals() {
	if !as.IsComplete() {
		return
	}

	totalVolume := as.RollSetup.TotalVolume
	as.Calculations = []DilutionCalculation{}

	// ILFOSOL 3
	as.Calculations = append(as.Calculations, calculateDilution("ILFOSOL 3", as.Dilution, totalVolume, as.GetDevelopmentTime()))

	// ILFOSTOP
	ilfostop, _ := as.ChemicalDB.GetChemical("ilfostop")
	as.Calculations = append(as.Calculations, calculateDilution("ILFOSTOP", "1+19", totalVolume, ilfostop.Time))

	// SPRINT FIXER
	fixer, _ := as.ChemicalDB.GetChemical("sprint_fixer")
	as.Calculations = append(as.Calculations, calculateDilution("SPRINT FIXER", "1+4", totalVolume, fixer.Time))

	// Create timer state from calculations
	as.TimerState = NewTimerState(as.Calculations)
}

// calculateDilution calculates the dilution for a given chemical
func calculateDilution(chemical, dilution string, totalVolume int, time string) DilutionCalculation {
	var concentrate, water int

	switch dilution {
	case "1+9":
		concentrate = totalVolume / 10
		water = totalVolume - concentrate
	case "1+14":
		concentrate = totalVolume / 15
		water = totalVolume - concentrate
	case "1+19":
		concentrate = totalVolume / 20
		water = totalVolume - concentrate
	case "1+4":
		concentrate = totalVolume / 5
		water = totalVolume - concentrate
	}

	return DilutionCalculation{
		Chemical:    chemical,
		Dilution:    dilution,
		TotalVolume: totalVolume,
		Concentrate: concentrate,
		Water:       water,
		Time:        time,
	}
}

// ParseDuration parses a time string like "6:30" into a time.Duration
func ParseDuration(timeStr string) (time.Duration, error) {
	var minutes, seconds int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &minutes, &seconds)
	if err != nil {
		return 0, err
	}
	return time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

// FormatDuration formats a time.Duration into a string like "6:30"
func FormatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// TimerStep represents a development step that can be timed
type TimerStep struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Started  bool          `json:"started"`
	Finished bool          `json:"finished"`
}

// TimerState represents the timer functionality
type TimerState struct {
	CurrentStep int           `json:"current_step"`
	Steps       []TimerStep   `json:"steps"`
	StartTime   time.Time     `json:"start_time"`
	ElapsedTime time.Duration `json:"elapsed_time"`
	IsRunning   bool          `json:"is_running"`
	IsPaused    bool          `json:"is_paused"`
	IsComplete  bool          `json:"is_complete"`
}

// NewTimerState creates a new timer state from chemical calculations
func NewTimerState(calculations []DilutionCalculation) *TimerState {
	var steps []TimerStep

	for _, calc := range calculations {
		duration, err := ParseDuration(calc.Time)
		if err != nil {
			continue
		}

		steps = append(steps, TimerStep{
			Name:     calc.Chemical,
			Duration: duration,
			Started:  false,
			Finished: false,
		})
	}

	return &TimerState{
		CurrentStep: 0,
		Steps:       steps,
		IsRunning:   false,
		IsPaused:    false,
		IsComplete:  false,
	}
}

// StartTimer starts the current step timer
func (ts *TimerState) StartTimer() {
	if ts.CurrentStep >= len(ts.Steps) {
		return
	}

	ts.IsRunning = true
	ts.IsPaused = false
	ts.StartTime = time.Now()
	ts.Steps[ts.CurrentStep].Started = true
}

// PauseTimer pauses the current timer
func (ts *TimerState) PauseTimer() {
	if ts.IsRunning {
		ts.IsPaused = true
		ts.ElapsedTime += time.Since(ts.StartTime)
	}
}

// ResumeTimer resumes the paused timer
func (ts *TimerState) ResumeTimer() {
	if ts.IsPaused {
		ts.IsPaused = false
		ts.StartTime = time.Now()
	}
}

// StopTimer stops the current timer
func (ts *TimerState) StopTimer() {
	ts.IsRunning = false
	ts.IsPaused = false
	ts.ElapsedTime = 0
}

// CompleteCurrentStep marks the current step as complete and moves to next
func (ts *TimerState) CompleteCurrentStep() {
	if ts.CurrentStep < len(ts.Steps) {
		ts.Steps[ts.CurrentStep].Finished = true
		ts.CurrentStep++
		ts.IsRunning = false
		ts.IsPaused = false
		ts.ElapsedTime = 0

		if ts.CurrentStep >= len(ts.Steps) {
			ts.IsComplete = true
		}
	}
}

// GetCurrentElapsed returns the current elapsed time
func (ts *TimerState) GetCurrentElapsed() time.Duration {
	if ts.IsRunning && !ts.IsPaused {
		return ts.ElapsedTime + time.Since(ts.StartTime)
	}
	return ts.ElapsedTime
}

// GetRemainingTime returns the remaining time for current step
func (ts *TimerState) GetRemainingTime() time.Duration {
	if ts.CurrentStep >= len(ts.Steps) {
		return 0
	}

	elapsed := ts.GetCurrentElapsed()
	remaining := ts.Steps[ts.CurrentStep].Duration - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsCurrentStepOvertime checks if current step has exceeded its duration
func (ts *TimerState) IsCurrentStepOvertime() bool {
	if ts.CurrentStep >= len(ts.Steps) {
		return false
	}

	return ts.GetCurrentElapsed() > ts.Steps[ts.CurrentStep].Duration
}

// GetCurrentStep returns the current step info
func (ts *TimerState) GetCurrentStep() *TimerStep {
	if ts.CurrentStep >= len(ts.Steps) {
		return nil
	}
	return &ts.Steps[ts.CurrentStep]
}

// Reset resets the timer to the beginning
func (ts *TimerState) Reset() {
	ts.CurrentStep = 0
	ts.IsRunning = false
	ts.IsPaused = false
	ts.IsComplete = false
	ts.ElapsedTime = 0

	for i := range ts.Steps {
		ts.Steps[i].Started = false
		ts.Steps[i].Finished = false
	}
}
