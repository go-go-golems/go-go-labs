package main

import (
	"sync"
	"time"
)

// StreamStore manages stream data
type StreamStore struct {
	mutex  sync.RWMutex
	stream Stream
}

// NewStreamStore creates a new store with default data
func NewStreamStore() *StreamStore {
	return &StreamStore{
		stream: Stream{
			Info: StreamInfo{
				Title:       "Building a React Component Library",
				Description: "Creating reusable UI components with TailwindCSS",
				StartTime:   time.Now(),
				Language:    "JavaScript/React",
				GithubRepo:  "https://github.com/yourusername/component-library",
				ViewerCount: 42,
			},
			Steps: StepInfo{
				Completed: []string{
					"Project setup and initialization",
					"Design system planning",
				},
				Active: "Setting up component architecture",
				Upcoming: []string{
					"Implement Button component",
					"Create Card component",
					"Build Form elements",
					"Add dark mode toggle",
				},
			},
		},
	}
}

// GetStreamInfo returns the current stream info
func (s *StreamStore) GetStreamInfo() StreamInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stream.Info
}

// UpdateStreamInfo updates stream info
func (s *StreamStore) UpdateStreamInfo(info StreamInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stream.Info = info
}

// GetSteps returns all steps
func (s *StreamStore) GetSteps() StepInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stream.Steps
}

// SetActiveStep sets a new active step
func (s *StreamStore) SetActiveStep(step string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Add current active to completed if it exists
	if s.stream.Steps.Active != "" {
		s.stream.Steps.Completed = append(s.stream.Steps.Completed, s.stream.Steps.Active)
	}
	
	s.stream.Steps.Active = step
}

// AddUpcomingStep adds a new upcoming step
func (s *StreamStore) AddUpcomingStep(step string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stream.Steps.Upcoming = append(s.stream.Steps.Upcoming, step)
}

// CompleteActiveStep completes the current active step
func (s *StreamStore) CompleteActiveStep() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.stream.Steps.Active != "" {
		// Add to completed
		s.stream.Steps.Completed = append(s.stream.Steps.Completed, s.stream.Steps.Active)
		
		// Set next step as active if available
		if len(s.stream.Steps.Upcoming) > 0 {
			s.stream.Steps.Active = s.stream.Steps.Upcoming[0]
			s.stream.Steps.Upcoming = s.stream.Steps.Upcoming[1:]
		} else {
			s.stream.Steps.Active = ""
		}
	}
}

// ReactivateStep moves a step from completed/upcoming to active
func (s *StreamStore) ReactivateStep(step string, source string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Add current active to completed
	if s.stream.Steps.Active != "" {
		s.stream.Steps.Completed = append(s.stream.Steps.Completed, s.stream.Steps.Active)
	}
	
	// Set step as active
	s.stream.Steps.Active = step
	
	// Remove from source list
	if source == "upcoming" {
		for i, s := range s.stream.Steps.Upcoming {
			if s == step {
				s.stream.Steps.Upcoming = append(s.stream.Steps.Upcoming[:i], s.stream.Steps.Upcoming[i+1:]...)
				break
			}
		}
	} else if source == "completed" {
		for i, s := range s.stream.Steps.Completed {
			if s == step {
				s.stream.Steps.Completed = append(s.stream.Steps.Completed[:i], s.stream.Steps.Completed[i+1:]...)
				break
			}
		}
	}
}