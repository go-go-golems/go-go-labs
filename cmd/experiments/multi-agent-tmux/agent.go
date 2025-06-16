package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Agent represents a mock LLM agent
type Agent interface {
	ID() string
	Name() string
	Role() string
	Execute(ctx context.Context, task string, output chan<- AgentMessage) error
}

// AgentMessage represents a message from an agent
type AgentMessage struct {
	AgentID   string    `json:"agent_id"`
	AgentName string    `json:"agent_name"`
	Type      string    `json:"type"` // "status", "progress", "result", "error"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	id   string
	name string
	role string
}

func (a *BaseAgent) ID() string   { return a.id }
func (a *BaseAgent) Name() string { return a.name }
func (a *BaseAgent) Role() string { return a.role }

// ResearchAgent simulates a research-focused LLM agent
type ResearchAgent struct {
	BaseAgent
}

func NewResearchAgent(id string) *ResearchAgent {
	return &ResearchAgent{
		BaseAgent: BaseAgent{
			id:   id,
			name: "Research Agent",
			role: "Conducts research and gathers information",
		},
	}
}

func (a *ResearchAgent) Execute(ctx context.Context, task string, output chan<- AgentMessage) error {
	defer close(output)

	a.sendMessage(output, "status", "Starting research task: "+task)

	researchSteps := []string{
		"Analyzing research requirements",
		"Searching academic databases",
		"Reviewing recent publications",
		"Identifying key sources",
		"Extracting relevant information",
		"Validating findings",
		"Compiling research summary",
	}

	for i, step := range researchSteps {
		select {
		case <-ctx.Done():
			a.sendMessage(output, "error", "Research task cancelled")
			return ctx.Err()
		default:
		}

		a.sendMessage(output, "progress", fmt.Sprintf("[%d/%d] %s", i+1, len(researchSteps), step))

		// Simulate variable processing time
		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

		// Occasionally send detailed findings
		if rand.Float32() < 0.3 {
			findings := []string{
				"Found relevant study on distributed systems",
				"Identified performance bottleneck in current approach",
				"Discovered new optimization technique",
				"Located comprehensive benchmark data",
			}
			a.sendMessage(output, "result", findings[rand.Intn(len(findings))])
		}
	}

	a.sendMessage(output, "result", "Research completed successfully. Found 12 relevant sources with key insights.")
	return nil
}

func (a *ResearchAgent) sendMessage(output chan<- AgentMessage, msgType, content string) {
	output <- AgentMessage{
		AgentID:   a.id,
		AgentName: a.name,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// AnalysisAgent simulates an analysis-focused LLM agent
type AnalysisAgent struct {
	BaseAgent
}

func NewAnalysisAgent(id string) *AnalysisAgent {
	return &AnalysisAgent{
		BaseAgent: BaseAgent{
			id:   id,
			name: "Analysis Agent",
			role: "Analyzes data and provides insights",
		},
	}
}

func (a *AnalysisAgent) Execute(ctx context.Context, task string, output chan<- AgentMessage) error {
	defer close(output)

	a.sendMessage(output, "status", "Starting analysis task: "+task)

	analysisSteps := []string{
		"Loading data for analysis",
		"Preprocessing raw data",
		"Applying statistical methods",
		"Running correlation analysis",
		"Identifying patterns and trends",
		"Generating insights",
		"Validating results",
		"Preparing recommendations",
	}

	for i, step := range analysisSteps {
		select {
		case <-ctx.Done():
			a.sendMessage(output, "error", "Analysis task cancelled")
			return ctx.Err()
		default:
		}

		a.sendMessage(output, "progress", fmt.Sprintf("[%d/%d] %s", i+1, len(analysisSteps), step))

		time.Sleep(time.Duration(rand.Intn(4)+1) * time.Second)

		// Send intermediate results
		if rand.Float32() < 0.4 {
			results := []string{
				"Correlation coefficient: 0.87 (strong positive)",
				"Outlier detected in dataset sector 3",
				"Trend analysis shows 15% improvement",
				"Statistical significance: p < 0.001",
			}
			a.sendMessage(output, "result", results[rand.Intn(len(results))])
		}
	}

	a.sendMessage(output, "result", "Analysis complete. Generated 5 key insights with high confidence.")
	return nil
}

func (a *AnalysisAgent) sendMessage(output chan<- AgentMessage, msgType, content string) {
	output <- AgentMessage{
		AgentID:   a.id,
		AgentName: a.name,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// WritingAgent simulates a writing-focused LLM agent
type WritingAgent struct {
	BaseAgent
}

func NewWritingAgent(id string) *WritingAgent {
	return &WritingAgent{
		BaseAgent: BaseAgent{
			id:   id,
			name: "Writing Agent",
			role: "Creates written content and documentation",
		},
	}
}

func (a *WritingAgent) Execute(ctx context.Context, task string, output chan<- AgentMessage) error {
	defer close(output)

	a.sendMessage(output, "status", "Starting writing task: "+task)

	writingSteps := []string{
		"Planning document structure",
		"Creating outline",
		"Writing introduction",
		"Developing main sections",
		"Adding supporting details",
		"Incorporating examples",
		"Writing conclusion",
		"Reviewing and editing",
		"Final formatting",
	}

	for i, step := range writingSteps {
		select {
		case <-ctx.Done():
			a.sendMessage(output, "error", "Writing task cancelled")
			return ctx.Err()
		default:
		}

		a.sendMessage(output, "progress", fmt.Sprintf("[%d/%d] %s", i+1, len(writingSteps), step))

		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)

		// Send writing updates
		if rand.Float32() < 0.3 {
			updates := []string{
				"Completed section 1: Introduction (487 words)",
				"Added technical diagram to section 2",
				"Incorporated research findings",
				"Updated bibliography with 8 new sources",
			}
			a.sendMessage(output, "result", updates[rand.Intn(len(updates))])
		}
	}

	a.sendMessage(output, "result", "Writing completed. Generated 2,847 words with proper formatting.")
	return nil
}

func (a *WritingAgent) sendMessage(output chan<- AgentMessage, msgType, content string) {
	output <- AgentMessage{
		AgentID:   a.id,
		AgentName: a.name,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// ReviewAgent simulates a review-focused LLM agent
type ReviewAgent struct {
	BaseAgent
}

func NewReviewAgent(id string) *ReviewAgent {
	return &ReviewAgent{
		BaseAgent: BaseAgent{
			id:   id,
			name: "Review Agent",
			role: "Reviews and provides feedback on work",
		},
	}
}

func (a *ReviewAgent) Execute(ctx context.Context, task string, output chan<- AgentMessage) error {
	defer close(output)

	a.sendMessage(output, "status", "Starting review task: "+task)

	reviewSteps := []string{
		"Loading content for review",
		"Checking structure and organization",
		"Verifying factual accuracy",
		"Reviewing grammar and style",
		"Assessing clarity and coherence",
		"Checking citations and references",
		"Evaluating completeness",
		"Generating feedback report",
	}

	for i, step := range reviewSteps {
		select {
		case <-ctx.Done():
			a.sendMessage(output, "error", "Review task cancelled")
			return ctx.Err()
		default:
		}

		a.sendMessage(output, "progress", fmt.Sprintf("[%d/%d] %s", i+1, len(reviewSteps), step))

		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		// Send review findings
		if rand.Float32() < 0.4 {
			findings := []string{
				"Minor grammatical error found in paragraph 3",
				"Excellent use of supporting evidence",
				"Suggestion: Add transition between sections",
				"Citation format needs standardization",
			}
			a.sendMessage(output, "result", findings[rand.Intn(len(findings))])
		}
	}

	a.sendMessage(output, "result", "Review completed. Overall quality: Excellent. 3 minor suggestions provided.")
	return nil
}

func (a *ReviewAgent) sendMessage(output chan<- AgentMessage, msgType, content string) {
	output <- AgentMessage{
		AgentID:   a.id,
		AgentName: a.name,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}
}
