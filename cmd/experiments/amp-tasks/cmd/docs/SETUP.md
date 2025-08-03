# Project Setup Guide

Complete initialization workflow for setting up a new project with proper structure, agents, and task hierarchy.

## 1. Project Creation

### Create Your Project
```bash
# Create a new project with guidelines
amp-tasks projects create "Web Dashboard" \
  --description "Customer-facing analytics dashboard" \
  --guidelines "Use TypeScript, follow React patterns, write tests for all components, prioritize accessibility"

# View the created project
amp-tasks projects list
```

### Set as Default Project
```bash
# Set as your default project (replace with actual project ID)
amp-tasks projects set-default <project-id>

# Verify it's set correctly
amp-tasks projects default
```

### Guidelines Best Practices
- **Technical Standards**: Language, frameworks, coding patterns
- **Quality Requirements**: Testing, code review, documentation
- **Business Context**: User needs, project goals, constraints
- **Collaboration Rules**: Communication channels, decision processes

## 2. Agent Type Setup

### Choose Agent Types for Your Project
Consider what types of work your project needs:

#### Development Agent Types
```bash
# Frontend specialists
amp-tasks agent-types create "Frontend Developer" \
  --description "React/TypeScript components, UI/UX implementation, responsive design" \
  --project-id <project-id>

# Backend specialists  
amp-tasks agent-types create "Backend Developer" \
  --description "API development, database design, server infrastructure" \
  --project-id <project-id>

# Full-stack generalists
amp-tasks agent-types create "Full Stack Developer" \
  --description "End-to-end feature development, system integration" \
  --project-id <project-id>
```

#### Quality & Review Agent Types
```bash
# Code quality experts
amp-tasks agent-types create "Code Reviewer" \
  --description "Code quality, security review, architectural guidance" \
  --project-id <project-id>

# Testing specialists
amp-tasks agent-types create "Test Engineer" \
  --description "Test automation, quality assurance, test strategy" \
  --project-id <project-id>

# Security experts
amp-tasks agent-types create "Security Specialist" \
  --description "Security analysis, vulnerability assessment, compliance" \
  --project-id <project-id>
```

#### Operations & Infrastructure Agent Types
```bash
# Infrastructure specialists
amp-tasks agent-types create "DevOps Engineer" \
  --description "CI/CD, deployment, monitoring, infrastructure as code" \
  --project-id <project-id>

# Data specialists
amp-tasks agent-types create "Data Engineer" \
  --description "Data pipelines, analytics, database optimization" \
  --project-id <project-id>
```

#### Management & Coordination Agent Types
```bash
# Project coordination
amp-tasks agent-types create "Tech Lead" \
  --description "Technical leadership, architecture decisions, team coordination" \
  --project-id <project-id>

# Documentation and communication
amp-tasks agent-types create "Documentation Specialist" \
  --description "Technical writing, API docs, user guides, knowledge management" \
  --project-id <project-id>
```

### View Your Agent Types
```bash
# See all agent types for your project
amp-tasks agent-types list --project-id <project-id>
```

## 3. Agent Creation

### Create Initial Agents
```bash
# Create agents for each type (examples)
amp-tasks agents create "Alice Frontend" --agent-type-id <frontend-type-id>
amp-tasks agents create "Bob Backend" --agent-type-id <backend-type-id>
amp-tasks agents create "Carol Review" --agent-type-id <code-reviewer-type-id>
amp-tasks agents create "David Test" --agent-type-id <test-engineer-type-id>
amp-tasks agents create "Eve DevOps" --agent-type-id <devops-type-id>

# View your agent workforce
amp-tasks agents list
```

### Agent Naming Best Practices
- **Descriptive Names**: Include role/specialty ("Alice Frontend", "Security-Bob")
- **Team Organization**: Use prefixes for teams ("Team1-Alice", "Core-Bob")
- **Skill Indication**: Include key skills ("React-Alice", "K8s-Bob")

## 4. Initial Task Structure

### Create Epic-Level Tasks (Top-Level Features)
```bash
# Major feature areas
amp-tasks tasks create "User Authentication System" \
  --description "Complete auth flow: login, registration, password reset, session management"

amp-tasks tasks create "Analytics Dashboard" \
  --description "Real-time charts, filtering, data export, customizable views"

amp-tasks tasks create "API Infrastructure" \
  --description "REST endpoints, rate limiting, validation, error handling"

amp-tasks tasks create "Data Pipeline" \
  --description "ETL processes, data validation, monitoring, alerting"
```

### Create Feature-Level Tasks (Under Epics)
```bash
# Get epic task IDs first
amp-tasks tasks list

# Create sub-features under authentication epic
amp-tasks tasks create "Login Component" \
  --description "React component with form validation, error handling" \
  --parent <auth-epic-id>

amp-tasks tasks create "JWT Token Service" \
  --description "Token generation, validation, refresh logic" \
  --parent <auth-epic-id>

amp-tasks tasks create "Password Reset Flow" \
  --description "Email sending, token validation, password update" \
  --parent <auth-epic-id>

# Create sub-features under dashboard epic  
amp-tasks tasks create "Chart Components" \
  --description "Reusable D3/Chart.js components with responsive design" \
  --parent <dashboard-epic-id>

amp-tasks tasks create "Data Fetching Layer" \
  --description "API client, caching, error handling, loading states" \
  --parent <dashboard-epic-id>
```

### Create Implementation Tasks (Specific Work Items)
```bash
# Get feature task IDs
amp-tasks tasks list

# Implementation tasks under login component
amp-tasks tasks create "Login Form UI" \
  --description "Form layout, styling, responsive design" \
  --parent <login-component-id>

amp-tasks tasks create "Form Validation" \
  --description "Client-side validation, error messages, accessibility" \
  --parent <login-component-id>

amp-tasks tasks create "API Integration" \
  --description "Connect form to auth API, handle responses" \
  --parent <login-component-id>

amp-tasks tasks create "Unit Tests" \
  --description "Test form validation, API mocking, edge cases" \
  --parent <login-component-id>
```

### Task Hierarchy Best Practices
- **Epic Level** (3-6 months): Major features or systems
- **Feature Level** (2-4 weeks): Cohesive functionality within epics  
- **Implementation Level** (1-5 days): Specific, actionable work items
- **Task Level** (2-8 hours): Individual commits or small changes

## 5. Dependency Setup

### Add Cross-Feature Dependencies
```bash
# API must exist before frontend can integrate
amp-tasks deps add <login-component-id> <jwt-service-id>
amp-tasks deps add <chart-components-id> <data-fetching-layer-id>

# Infrastructure before application features
amp-tasks deps add <auth-epic-id> <api-infrastructure-id>
amp-tasks deps add <dashboard-epic-id> <data-pipeline-id>

# Foundation before specialization
amp-tasks deps add <form-validation-id> <login-form-ui-id>
amp-tasks deps add <api-integration-id> <form-validation-id>
amp-tasks deps add <unit-tests-id> <api-integration-id>
```

### Add Quality Gates
```bash
# All implementation tasks need testing
amp-tasks tasks create "Code Review" \
  --description "Security review, code quality, architectural compliance" \
  --parent <login-component-id>

amp-tasks deps add <code-review-id> <unit-tests-id>

# Integration tests depend on unit tests
amp-tasks tasks create "Integration Tests" \
  --description "End-to-end auth flow testing" \
  --parent <auth-epic-id>

amp-tasks deps add <integration-tests-id> <code-review-id>
```

### Dependency Best Practices
- **Technical Dependencies**: API before UI, database before API
- **Quality Gates**: Tests before review, review before merge
- **Risk Management**: Core features before nice-to-have features
- **Team Coordination**: Shared components before features using them

## 6. Verification Steps

### Verify Project Structure
```bash
# Check project is set as default
amp-tasks projects default

# View agent types and coverage
amp-tasks agent-types list

# View agent workforce
amp-tasks agents list

# Check task hierarchy
amp-tasks tasks list
```

### Verify Dependencies
```bash
# Visualize dependency graph
amp-tasks deps graph

# Check for circular dependencies (should be none)
amp-tasks deps validate

# See what's ready to start
amp-tasks tasks available
```

### Verify Assignment Strategy
```bash
# Test flexible assignment
amp-tasks agent-types assign <available-task-id> <frontend-developer-type-id>

# Verify assignment worked
amp-tasks tasks show <task-id>

# Test specific assignment  
amp-tasks tasks assign <another-task-id> <alice-agent-id>

# Check workload distribution
amp-tasks agents workload
```

## 7. Start Development Workflow

### Begin First Sprint
```bash
# Check what's ready to work on
amp-tasks tasks available

# Assign foundational tasks first
amp-tasks agent-types assign <api-infrastructure-task> <backend-developer-type>
amp-tasks agent-types assign <ui-foundation-task> <frontend-developer-type>

# Set up monitoring for progress
amp-tasks notes add <task-id> "Sprint 1 started - focusing on foundation"
```

### Establish Knowledge Sharing
```bash
# Document architectural decisions
amp-tasks til create "Project Architecture" \
  --content "Using React + TypeScript frontend, Node.js + Express backend, PostgreSQL database"

# Share setup insights
amp-tasks til create "Development Environment" \
  --content "Use Docker for local DB, pnpm for package management, runs on port 3000"
```

## Quick Setup Template

For a rapid setup, copy and adapt this sequence:

```bash
#!/bin/bash
# Quick project setup script

# 1. Create project
PROJECT_ID=$(amp-tasks projects create "Your Project" --description "Description" --guidelines "Your guidelines" --output json | jq -r '.id')
amp-tasks projects set-default $PROJECT_ID

# 2. Create agent types
FRONTEND_TYPE=$(amp-tasks agent-types create "Frontend Dev" --description "UI/UX" --project-id $PROJECT_ID --output json | jq -r '.id')
BACKEND_TYPE=$(amp-tasks agent-types create "Backend Dev" --description "API/DB" --project-id $PROJECT_ID --output json | jq -r '.id')
REVIEWER_TYPE=$(amp-tasks agent-types create "Code Reviewer" --description "Quality" --project-id $PROJECT_ID --output json | jq -r '.id')

# 3. Create agents
amp-tasks agents create "Alice" --agent-type-id $FRONTEND_TYPE
amp-tasks agents create "Bob" --agent-type-id $BACKEND_TYPE  
amp-tasks agents create "Carol" --agent-type-id $REVIEWER_TYPE

# 4. Create initial tasks
EPIC1=$(amp-tasks tasks create "Core Features" --description "Main functionality" --output json | jq -r '.id')
EPIC2=$(amp-tasks tasks create "Infrastructure" --description "Foundation" --output json | jq -r '.id')

# 5. Add dependencies
amp-tasks deps add $EPIC1 $EPIC2

# 6. Verify
amp-tasks projects default
amp-tasks tasks available
amp-tasks deps graph
```

## Next Steps

After setup is complete:

1. **Start Small**: Begin with foundational tasks
2. **Document Progress**: Use notes and TIL entries actively  
3. **Review Dependencies**: Adjust as you learn more about the work
4. **Iterate Structure**: Add more tasks and agent types as needed
5. **Monitor Workload**: Use `agents workload` to balance assignments
6. **Share Knowledge**: Create TIL entries for discoveries and best practices

For ongoing work, see: `amp-tasks docs agent-guide` and `amp-tasks docs workflow`
