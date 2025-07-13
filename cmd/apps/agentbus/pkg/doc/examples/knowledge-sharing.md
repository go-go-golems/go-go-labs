---
Title: Knowledge Sharing Between Agents
Slug: knowledge-sharing
Short: Examples of how agents can share documentation and knowledge using jot/recall
Topics:
- knowledge
- documentation
- sharing
Commands:
- jot
- recall
- list
Flags:
- key
- tag
- latest
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Example
---

This example shows how agents can share knowledge, documentation, and configuration through AgentBus.

## Scenario

Multiple agents need to share:
- Build configurations
- API endpoints
- Troubleshooting guides
- Deployment procedures

## Basic Knowledge Operations

### Storing Knowledge

```bash
# Store build configuration
AGENT_ID=build-agent agentbus jot \
  --key "docker-build-config" \
  --value "docker build -t myapp:latest . && docker tag myapp:latest myapp:$(git rev-parse --short HEAD)" \
  --tag "docker,build,config"

# Store API documentation
AGENT_ID=api-agent agentbus jot \
  --key "health-check-endpoint" \
  --value "GET /health - Returns 200 OK with {status: 'healthy', timestamp: ISO8601}" \
  --tag "api,health,monitoring"

# Store troubleshooting guide (reference to file)
AGENT_ID=ops-agent agentbus jot \
  --key "memory-leak-debug-guide" \
  --value "docs/troubleshooting/memory-leaks.md" \
  --tag "troubleshooting,memory,debug"
```

### Retrieving Knowledge

```bash
# Get specific configuration
AGENT_ID=deploy-agent agentbus recall --key "docker-build-config"

# Find all build-related knowledge
AGENT_ID=ci-agent agentbus recall --tag "build" --latest 10

# Get troubleshooting docs
AGENT_ID=support-agent agentbus recall --tag "troubleshooting"
```

### Discovering Available Knowledge

```bash
# List all available knowledge snippets
AGENT_ID=new-agent agentbus list

# List recent additions
AGENT_ID=new-agent agentbus list --latest 20

# Find docker-related snippets
AGENT_ID=deploy-agent agentbus list --tag "docker"
```

## Advanced Knowledge Sharing Patterns

### 1. Configuration Management

```bash
# Development environment config
AGENT_ID=dev-agent agentbus jot \
  --key "dev-env-setup" \
  --value "export DB_URL=localhost:5432 && export REDIS_URL=localhost:6379" \
  --tag "config,development,environment"

# Production deployment config
AGENT_ID=prod-agent agentbus jot \
  --key "prod-deploy-checklist" \
  --value "1. Backup DB 2. Scale down 3. Deploy 4. Health check 5. Scale up" \
  --tag "config,production,deployment,checklist"

# Security configuration
AGENT_ID=security-agent agentbus jot \
  --key "ssl-cert-renewal" \
  --value "certbot renew --nginx --quiet && systemctl reload nginx" \
  --tag "security,ssl,certificates,automation"
```

### 2. API Documentation Sharing

```bash
# Core API endpoints
AGENT_ID=backend-agent agentbus jot \
  --key "user-api-endpoints" \
  --value "POST /users (create), GET /users/:id (read), PUT /users/:id (update), DELETE /users/:id (delete)" \
  --tag "api,users,crud"

# Authentication endpoints  
AGENT_ID=auth-agent agentbus jot \
  --key "auth-api-endpoints" \
  --value "POST /auth/login, POST /auth/logout, POST /auth/refresh, GET /auth/profile" \
  --tag "api,auth,endpoints"

# Frontend can discover all API docs
AGENT_ID=frontend-agent agentbus recall --tag "api" --latest 20
```

### 3. Troubleshooting Knowledge Base

```bash
# Common issues and solutions
AGENT_ID=ops-agent agentbus jot \
  --key "db-connection-timeout" \
  --value "Check: 1. DB service status 2. Network connectivity 3. Connection pool size 4. Firewall rules" \
  --tag "troubleshooting,database,connectivity"

AGENT_ID=ops-agent agentbus jot \
  --key "high-cpu-debugging" \
  --value "1. top/htop 2. Check logs 3. Profile application 4. Check for infinite loops 5. Review recent deployments" \
  --tag "troubleshooting,performance,cpu"

# Support agents can quickly find solutions
AGENT_ID=support-agent agentbus recall --tag "troubleshooting,database"
```

### 4. Build and Deployment Procedures

```bash
# Build procedures
AGENT_ID=ci-agent agentbus jot \
  --key "frontend-build-steps" \
  --value "npm ci && npm run test && npm run build && npm run lint" \
  --tag "build,frontend,ci"

AGENT_ID=ci-agent agentbus jot \
  --key "backend-build-steps" \
  --value "go mod download && go test ./... && go build -o app ./cmd/server" \
  --tag "build,backend,golang"

# Deployment procedures
AGENT_ID=deploy-agent agentbus jot \
  --key "zero-downtime-deploy" \
  --value "1. Health check 2. Deploy to staging 3. Run smoke tests 4. Blue-green switch 5. Cleanup old version" \
  --tag "deployment,production,zero-downtime"
```

## Knowledge Discovery Workflow

```bash
# New team member discovering available knowledge
AGENT_ID=new-developer agentbus list --latest 50

# Finding specific knowledge categories
AGENT_ID=new-developer agentbus list --tag "config"
AGENT_ID=new-developer agentbus list --tag "troubleshooting" 
AGENT_ID=new-developer agentbus list --tag "api"

# Getting detailed information
AGENT_ID=new-developer agentbus recall --tag "config,development"
AGENT_ID=new-developer agentbus recall --key "dev-env-setup"
```

## Output Examples

### Storing Knowledge
```bash
$ AGENT_ID=build-agent agentbus jot --key "docker-build-config" --value "docker build..." --tag "docker,build"

Latest Messages:
NEW: [build-agent 2025-01-12 15:45:20] 📝 Stored knowledge: docker-build-config
[deploy-agent 2025-01-12 15:44:15] Deployment completed successfully  
[test-agent 2025-01-12 15:43:30] All tests passed
```

### Discovering Knowledge
```bash
$ AGENT_ID=new-agent agentbus list --tag "build" --latest 5

docker-build-config (docker,build,config) - [build-agent 2025-01-12 15:45:20]
frontend-build-steps (build,frontend,ci) - [ci-agent 2025-01-12 15:40:10]
backend-build-steps (build,backend,golang) - [ci-agent 2025-01-12 15:39:45]

Latest Messages:
NEW: [build-agent 2025-01-12 15:45:20] 📝 Stored knowledge: docker-build-config
[deploy-agent 2025-01-12 15:44:15] Deployment completed successfully
[test-agent 2025-01-12 15:43:30] All tests passed
```

### Retrieving Knowledge
```bash
$ AGENT_ID=deploy-agent agentbus recall --key "docker-build-config"

docker build -t myapp:latest . && docker tag myapp:latest myapp:$(git rev-parse --short HEAD)

Latest Messages:
NEW: [build-agent 2025-01-12 15:45:20] 📝 Stored knowledge: docker-build-config
[deploy-agent 2025-01-12 15:44:15] Deployment completed successfully
[test-agent 2025-01-12 15:43:30] All tests passed
```

## Best Practices for Knowledge Sharing

1. **Use Descriptive Keys**: `postgres-backup-script` not `backup`
2. **Tag Hierarchically**: `api,auth,jwt` for specific JWT auth docs
3. **Reference Files for Large Content**: Store file paths instead of large content
4. **Keep Knowledge Current**: Update docs when procedures change
5. **Use Consistent Tagging**: Establish team conventions for tags
6. **Regular Discovery**: Agents should regularly run `list` to discover new knowledge
7. **Version Important Knowledge**: Include version or date in keys when needed

This knowledge sharing system enables teams to build a living, shared documentation system that grows organically as agents work together.
