package scenarios

import (
	"math/rand"
	"time"
)

// Scenario represents a predefined work scenario for agents
type Scenario struct {
	Name                string
	Description         string
	Steps               []string
	EstimatedDuration   time.Duration
	ErrorProbability    float64 // 0.0 - 1.0
	QuestionProbability float64 // 0.0 - 1.0
	PossibleErrors      []string
	PossibleQuestions   []string
}

// predefinedScenarios contains all available scenarios
var predefinedScenarios = []Scenario{
	{
		Name:                "Bug Fix - Authentication Issue",
		Description:         "Fix a critical authentication bug reported by users",
		EstimatedDuration:   45 * time.Minute,
		ErrorProbability:    0.2,
		QuestionProbability: 0.15,
		Steps: []string{
			"Reproduce the authentication issue",
			"Analyze authentication flow and identify root cause",
			"Review related authentication modules",
			"Design fix strategy",
			"Implement authentication fix",
			"Add regression tests for authentication",
			"Test fix with various user scenarios",
			"Update authentication documentation",
			"Deploy fix to staging environment",
			"Validate fix in production-like environment",
		},
		PossibleErrors: []string{
			"Cannot reproduce the authentication issue locally",
			"Authentication service dependencies are unavailable",
			"JWT token validation logic is more complex than expected",
			"Third-party authentication provider API changes detected",
			"Database schema migration required for fix",
		},
		PossibleQuestions: []string{
			"Should we invalidate all existing user sessions as part of this fix?",
			"The fix affects the login flow - should we notify users about temporary disruption?",
			"Found additional security vulnerabilities during analysis - address them now?",
			"Multiple authentication methods are affected - fix all or prioritize main method?",
		},
	},
	{
		Name:                "Feature Development - User Dashboard",
		Description:         "Implement a new user dashboard with metrics and preferences",
		EstimatedDuration:   2 * time.Hour,
		ErrorProbability:    0.1,
		QuestionProbability: 0.05,
		Steps: []string{
			"Analyze dashboard requirements and create mockups",
			"Design database schema for user preferences",
			"Implement backend API endpoints for dashboard data",
			"Create frontend dashboard components",
			"Implement user preference management",
			"Add data visualization components",
			"Integrate real-time metrics updates",
			"Add responsive design for mobile devices",
			"Implement dashboard customization features",
			"Add comprehensive testing for dashboard functionality",
			"Create user documentation and help guides",
		},
		PossibleErrors: []string{
			"Frontend framework compatibility issues with new components",
			"Performance issues with real-time metrics updates",
			"Database query optimization needed for large datasets",
			"CSS conflicts with existing dashboard styles",
			"Mobile responsiveness breaking on certain devices",
		},
		PossibleQuestions: []string{
			"Should the dashboard be fully customizable or use predefined layouts?",
			"What's the preferred approach for real-time updates - WebSockets or polling?",
			"How should we handle dashboard loading for users with limited data plans?",
			"Should we add export functionality for dashboard data?",
			"What level of analytics tracking should we include in the dashboard?",
		},
	},
	{
		Name:                "Performance Optimization - Database Queries",
		Description:         "Optimize slow database queries affecting application performance",
		EstimatedDuration:   90 * time.Minute,
		ErrorProbability:    0.15,
		QuestionProbability: 0.04,
		Steps: []string{
			"Identify slow queries using performance monitoring",
			"Analyze query execution plans",
			"Review database indexes and schema design",
			"Implement query optimizations",
			"Add database connection pooling improvements",
			"Create performance benchmarks",
			"Test optimizations under load",
			"Monitor memory usage and connection limits",
			"Update ORM configurations for better performance",
			"Document optimization changes and best practices",
		},
		PossibleErrors: []string{
			"Query optimization breaks existing functionality",
			"Database connection pool configuration conflicts",
			"Index creation causing temporary database locks",
			"Memory usage increases unexpectedly after optimization",
			"Performance regression in previously fast queries",
		},
		PossibleQuestions: []string{
			"Should we prioritize query speed or maintain current data consistency guarantees?",
			"The optimization requires schema changes - schedule maintenance window?",
			"Found queries that could benefit from caching - implement Redis layer?",
			"Some optimizations trade memory for speed - what's our memory budget?",
		},
	},
	{
		Name:                "Security Audit - API Endpoint Review",
		Description:         "Conduct security audit of API endpoints and implement fixes",
		EstimatedDuration:   3 * time.Hour,
		ErrorProbability:    0.3,
		QuestionProbability: 0.03,
		Steps: []string{
			"Catalog all API endpoints and their authentication requirements",
			"Review input validation and sanitization",
			"Analyze authorization and access control logic",
			"Check for SQL injection and XSS vulnerabilities",
			"Review API rate limiting and abuse prevention",
			"Analyze data exposure and privacy compliance",
			"Test authentication bypass scenarios",
			"Review API documentation for security information leaks",
			"Implement security fixes and improvements",
			"Add automated security tests to CI pipeline",
			"Create security incident response procedures",
			"Update security documentation and guidelines",
		},
		PossibleErrors: []string{
			"Security scanning tools producing false positives",
			"Fixing vulnerabilities breaks backward compatibility",
			"Access control changes affecting legitimate user workflows",
			"Rate limiting implementation causing service disruptions",
			"Encryption changes requiring client-side updates",
		},
		PossibleQuestions: []string{
			"Found potential data leak in API responses - break compatibility to fix?",
			"Should we implement stricter rate limiting even if it affects power users?",
			"Discovered endpoints without proper authentication - immediate fix or planned deprecation?",
			"Security audit reveals need for API versioning - implement now?",
			"Should we add security headers that might break older browser support?",
		},
	},
	{
		Name:                "Refactoring - Legacy Code Modernization",
		Description:         "Refactor legacy code module to use modern patterns and practices",
		EstimatedDuration:   4 * time.Hour,
		ErrorProbability:    0.25,
		QuestionProbability: 0.02,
		Steps: []string{
			"Analyze legacy code structure and dependencies",
			"Identify code smells and technical debt",
			"Create comprehensive test coverage for existing functionality",
			"Design modern architecture for the module",
			"Implement new module structure with clean separation",
			"Migrate data handling to use modern patterns",
			"Update error handling and logging",
			"Refactor configuration management",
			"Update documentation and code comments",
			"Perform thorough testing of refactored functionality",
			"Plan gradual rollout strategy",
		},
		PossibleErrors: []string{
			"Legacy dependencies not compatible with modern frameworks",
			"Refactoring introduces subtle behavioral changes",
			"Test coverage insufficient to catch all edge cases",
			"Performance regression in refactored code",
			"Configuration changes breaking deployment scripts",
		},
		PossibleQuestions: []string{
			"Should we maintain backward compatibility or clean break from legacy API?",
			"Refactoring reveals design issues in related modules - address them too?",
			"Modern patterns require newer language features - upgrade minimum version?",
			"Legacy module has undocumented behaviors - preserve them or document as deprecated?",
			"Should we refactor related modules in the same effort or separate projects?",
		},
	},
	{
		Name:                "Infrastructure - CI/CD Pipeline Enhancement",
		Description:         "Improve continuous integration and deployment pipeline",
		EstimatedDuration:   150 * time.Minute,
		ErrorProbability:    0.2,
		QuestionProbability: 0.01,
		Steps: []string{
			"Audit current CI/CD pipeline performance and reliability",
			"Identify bottlenecks in build and deployment process",
			"Implement parallel testing strategies",
			"Add automated security scanning to pipeline",
			"Improve deployment rollback mechanisms",
			"Add comprehensive monitoring and alerting",
			"Implement blue-green deployment strategy",
			"Add automated performance testing",
			"Improve artifact management and caching",
			"Update pipeline documentation and runbooks",
		},
		PossibleErrors: []string{
			"Pipeline changes breaking existing deployment workflows",
			"New testing stages significantly increasing build times",
			"Security scanning tools causing false positive failures",
			"Deployment automation conflicts with manual override procedures",
			"Resource limits reached with enhanced pipeline features",
		},
		PossibleQuestions: []string{
			"Should we prioritize build speed or comprehensive testing coverage?",
			"New deployment strategy requires infrastructure changes - coordinate with ops team?",
			"Enhanced monitoring will increase costs - what's the budget approval process?",
			"Should we implement gradual feature flag rollouts as part of deployment?",
			"Pipeline improvements might require developer workflow changes - training needed?",
		},
	},
	{
		Name:                "Data Migration - Database Schema Update",
		Description:         "Perform major database schema migration with zero downtime",
		EstimatedDuration:   5 * time.Hour,
		ErrorProbability:    0.4,
		QuestionProbability: 0.05,
		Steps: []string{
			"Analyze current schema and migration requirements",
			"Design backward-compatible migration strategy",
			"Create comprehensive data backup procedures",
			"Implement dual-write pattern for transitional period",
			"Develop schema migration scripts with rollback capability",
			"Test migration on production-like dataset",
			"Implement data validation and consistency checks",
			"Plan communication strategy for stakeholders",
			"Execute phased migration with monitoring",
			"Validate data integrity post-migration",
			"Clean up transitional code and old schema elements",
		},
		PossibleErrors: []string{
			"Migration scripts timeout on large production datasets",
			"Data validation reveals consistency issues requiring manual intervention",
			"Application compatibility issues with new schema during transition",
			"Rollback procedures fail due to data changes during migration",
			"Performance degradation during dual-write period",
		},
		PossibleQuestions: []string{
			"Migration will require brief read-only mode - acceptable maintenance window?",
			"Data validation found inconsistencies - fix during migration or separate cleanup?",
			"Should we implement additional monitoring during the transition period?",
			"Migration timeline extends beyond planned window - continue or rollback?",
			"New schema enables additional features - implement them now or later?",
		},
	},
	{
		Name:                "Integration - Third-Party Service Setup",
		Description:         "Integrate new third-party service for enhanced functionality",
		EstimatedDuration:   210 * time.Minute,
		ErrorProbability:    0.3,
		QuestionProbability: 0.05,
		Steps: []string{
			"Research third-party service API and capabilities",
			"Set up developer accounts and obtain API credentials",
			"Design integration architecture and data flow",
			"Implement service client with proper error handling",
			"Add configuration management for service settings",
			"Implement data synchronization mechanisms",
			"Add monitoring and alerting for service health",
			"Create fallback procedures for service outages",
			"Implement usage tracking and billing monitoring",
			"Add comprehensive testing including service mocks",
			"Update documentation and operational procedures",
		},
		PossibleErrors: []string{
			"Third-party service API documentation is outdated or incorrect",
			"Service rate limits lower than expected for production usage",
			"Data format incompatibilities requiring additional transformation",
			"Service authentication flow incompatible with current security model",
			"Network connectivity issues in production environment",
		},
		PossibleQuestions: []string{
			"Service has usage limits that might affect our scale - upgrade plan needed?",
			"Should we implement caching to reduce third-party service calls?",
			"Service provides additional features we don't need - minimal integration or full features?",
			"Integration requires storing third-party data - privacy compliance review needed?",
			"Should we implement graceful degradation when service is unavailable?",
		},
	},
}

// GetRandomScenario returns a random scenario
func GetRandomScenario() Scenario {
	return predefinedScenarios[rand.Intn(len(predefinedScenarios))]
}

// GetScenarioByName returns a scenario by name
func GetScenarioByName(name string) *Scenario {
	for _, scenario := range predefinedScenarios {
		if scenario.Name == name {
			return &scenario
		}
	}
	return nil
}

// GetAllScenarios returns all available scenarios
func GetAllScenarios() []Scenario {
	return predefinedScenarios
}

// GetScenariosByType returns scenarios filtered by type (based on name prefix)
func GetScenariosByType(scenarioType string) []Scenario {
	var filtered []Scenario
	for _, scenario := range predefinedScenarios {
		if len(scenario.Name) > len(scenarioType) && scenario.Name[:len(scenarioType)] == scenarioType {
			filtered = append(filtered, scenario)
		}
	}
	return filtered
}
