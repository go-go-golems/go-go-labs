package types

import "time"

// TimerStep represents a single step in the development process
type TimerStep struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Started  bool          `json:"started"`
	Finished bool          `json:"finished"`
}

// TimerState represents the state of the development timer
type TimerState struct {
	CurrentStep int           `json:"current_step"`
	Steps       []TimerStep   `json:"steps"`
	StartTime   time.Time     `json:"start_time"`
	ElapsedTime time.Duration `json:"elapsed_time"`
	IsRunning   bool          `json:"is_running"`
	IsPaused    bool          `json:"is_paused"`
	IsComplete  bool          `json:"is_complete"`
}

// NewTimerState creates a new timer state from dilution calculations
func NewTimerState(calculations []DilutionCalculation) *TimerState {
	var steps []TimerStep
	
	for _, calc := range calculations {
		duration, err := ParseDuration(calc.Time)
		if err != nil {
			// Default to 30 seconds if parsing fails
			duration = 30 * time.Second
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
		StartTime:   time.Time{},
		ElapsedTime: 0,
		IsRunning:   false,
		IsPaused:    false,
		IsComplete:  false,
	}
}

// StartTimer starts the timer
func (ts *TimerState) StartTimer() {
	if !ts.IsRunning {
		ts.StartTime = time.Now()
		ts.IsRunning = true
		ts.IsPaused = false
		
		if ts.CurrentStep < len(ts.Steps) {
			ts.Steps[ts.CurrentStep].Started = true
		}
	}
}

// PauseTimer pauses the timer
func (ts *TimerState) PauseTimer() {
	if ts.IsRunning && !ts.IsPaused {
		ts.ElapsedTime += time.Since(ts.StartTime)
		ts.IsPaused = true
	}
}

// ResumeTimer resumes the timer
func (ts *TimerState) ResumeTimer() {
	if ts.IsRunning && ts.IsPaused {
		ts.StartTime = time.Now()
		ts.IsPaused = false
	}
}

// StopTimer stops the timer
func (ts *TimerState) StopTimer() {
	ts.IsRunning = false
	ts.IsPaused = false
	ts.ElapsedTime = 0
}

// CompleteCurrentStep completes the current step and moves to the next
func (ts *TimerState) CompleteCurrentStep() {
	if ts.CurrentStep < len(ts.Steps) {
		ts.Steps[ts.CurrentStep].Finished = true
		ts.CurrentStep++
		ts.ElapsedTime = 0
		ts.StartTime = time.Now()
		
		if ts.CurrentStep >= len(ts.Steps) {
			ts.IsComplete = true
			ts.IsRunning = false
		} else if ts.IsRunning {
			ts.Steps[ts.CurrentStep].Started = true
		}
	}
}

// GetCurrentElapsed returns the elapsed time for the current step
func (ts *TimerState) GetCurrentElapsed() time.Duration {
	if !ts.IsRunning {
		return ts.ElapsedTime
	}
	if ts.IsPaused {
		return ts.ElapsedTime
	}
	return ts.ElapsedTime + time.Since(ts.StartTime)
}

// GetRemainingTime returns the remaining time for the current step
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

// IsCurrentStepOvertime returns true if the current step is over time
func (ts *TimerState) IsCurrentStepOvertime() bool {
	if ts.CurrentStep >= len(ts.Steps) {
		return false
	}
	
	elapsed := ts.GetCurrentElapsed()
	return elapsed > ts.Steps[ts.CurrentStep].Duration
}

// GetCurrentStep returns the current step
func (ts *TimerState) GetCurrentStep() *TimerStep {
	if ts.CurrentStep >= len(ts.Steps) {
		return nil
	}
	return &ts.Steps[ts.CurrentStep]
}

// Reset resets the timer state
func (ts *TimerState) Reset() {
	ts.CurrentStep = 0
	ts.StartTime = time.Time{}
	ts.ElapsedTime = 0
	ts.IsRunning = false
	ts.IsPaused = false
	ts.IsComplete = false
	
	for i := range ts.Steps {
		ts.Steps[i].Started = false
		ts.Steps[i].Finished = false
	}
} 