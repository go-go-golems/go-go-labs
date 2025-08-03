# Quick Start Guide

## ⚠️ BEFORE YOU START

**ALL AGENTS MUST:**
1. **Verify your agent identity exists** in the system
2. **Ensure your agent-type is defined** for the current project
3. **Follow proper persona management** if switching roles

```bash
# Quick identity check
amp-tasks agents list | grep "Your Name"
amp-tasks agent-types list | grep "Your Type"
amp-tasks projects default
```

## Coordinator vs Worker Agents

### For Coordinator Agents
If you're responsible for **setting up and organizing** work:
1. **Ensure coordinator persona exists** - Create "Your Name Coordinator" agent
2. **Check current project first** - `amp-tasks projects default`
3. **Create project-relevant agent types** - `amp-tasks agent-types create`
4. **Set up task structure** and dependencies
5. **Assign work** to appropriate agent types

### For Worker Agents  
If you're here to **get and complete** assigned work:
1. **Verify worker persona exists** - Create "Your Name Worker" agent with appropriate type
2. **Find your agent entry** - see [Agent Guide](AGENT_GUIDE.md) for self-identification
3. **Check available tasks** - `amp-tasks tasks available`
4. **Take on work** within your agent type capabilities
5. **Update progress** and mark completion

### For LLM Agents (Persona Switching)
If you need to **switch between coordination and work**:
1. **Create multiple personas** - "Alice Coordinator", "Alice Developer", "Alice Reviewer"
2. **Assign appropriate agent-types** - Tech Lead, Frontend Developer, Code Reviewer
3. **Switch context as needed** - Use different agent IDs for different work types
4. **Document persona changes** - Note when and why you switch roles

---

## 1. Create your first project
```bash
# Initialize a new project with guidelines
amp-tasks projects create "My Project" \
  --description "Project description" \
  --guidelines "Follow coding standards, write tests, document APIs"

# Set as default project
amp-tasks projects set-default <project-id>

# Verify project is active
amp-tasks projects default
```

## 2. Check existing agent types first
```bash
# ALWAYS check existing agent types before creating new ones
amp-tasks agent-types list

# Search for similar types that might already exist
amp-tasks agent-types list | grep -i "developer\|frontend\|backend\|reviewer\|tester"

# Check global agent types (available across all projects)
amp-tasks agent-types list --output json | jq '.[] | select(.global == true)'
```

## 2.1 Create agent types ONLY if needed
```bash
# Only create if no suitable agent type exists
# Example: Check for "Developer" types first
amp-tasks agent-types list | grep -i "develop"

# If none exist, then create project-specific agent types
amp-tasks agent-types create "Developer" \
  --description "Full-stack development, feature implementation"

amp-tasks agent-types create "Code Reviewer" \
  --description "Code quality, security review, architectural guidance"

# Create global agent types only if truly needed across all projects
amp-tasks agent-types create "Documentation Writer" \
  --description "Technical writing, documentation maintenance" \
  --global

# Verify your agent types were created
amp-tasks agent-types list
```

## 3. Create and verify your agent identity

### CRITICAL: Check First, Then Create Your Agent Entry
```bash
# FIRST: Check if you already exist to avoid duplicates
amp-tasks agents list | grep -i "your.*name"
amp-tasks agents list | grep -i "coordinator\|developer\|reviewer"

# SECOND: Check available agent types to choose from existing ones
amp-tasks agent-types list

# THIRD: Only create if you don't already exist
amp-tasks agents create "Your Actual Name" --agent-type-id <existing-type-id>

# Verify your agent was created successfully
amp-tasks agents list | grep "Your Actual Name"

# For LLM agents who need multiple personas - CHECK FIRST:
amp-tasks agents list | grep "Your Name"
# Only create personas that don't exist:
amp-tasks agents create "Your Name Coordinator" --agent-type-id <existing-tech-lead-type>
amp-tasks agents create "Your Name Developer" --agent-type-id <existing-developer-type>
amp-tasks agents create "Your Name Reviewer" --agent-type-id <existing-reviewer-type>
```

### Create Other Team Agents (Optional)
```bash
# ALWAYS check first to avoid duplicates
amp-tasks agents list | grep -i "alice\|bob"

# Only create agents that don't already exist
amp-tasks agents create "Alice Dev" --agent-type-id <existing-developer-type>
amp-tasks agents create "Bob Reviewer" --agent-type-id <existing-reviewer-type>

# See your complete agent workforce
amp-tasks agents list
```

## 4. Create initial tasks
```bash
# Create your first tasks
amp-tasks tasks create "Setup project structure" \
  --description "Initialize modules and directory structure"

amp-tasks tasks create "Implement authentication" \
  --description "User login and session management"

# View your tasks
amp-tasks tasks list
```

## 5. See available work
```bash
# Check what's ready to work on
amp-tasks tasks available
```

## 6. Assign and complete work
```bash
# Assign work strategically:
# By agent type (flexible - any qualified agent can pick it up)
amp-tasks agent-types assign <task-id> <agent-type-id>

# By specific agent (targeted - specific expertise needed)
amp-tasks tasks assign <task-id> <agent-id>
```

## 7. Track progress with notes
```bash
amp-tasks notes add <task-id> "Working on authentication middleware"
```

## 8. Share insights
```bash
amp-tasks til create "Auth Pattern" --content "Use JWT for stateless auth"
```

## 9. Update status
```bash
amp-tasks tasks status <task-id> completed
```

## 10. Add dependencies and organize work
```bash
# Create dependent tasks
amp-tasks tasks create "Frontend auth" --description "Login UI components"
amp-tasks deps add <frontend-auth-id> <backend-auth-id>

# Visualize your work structure
amp-tasks deps graph
```

## 11. Working across projects
```bash
# Switch to different project
amp-tasks projects list
amp-tasks projects set-default <project-id>

# Agent types automatically scope to current project
amp-tasks agent-types list  # Shows global + current project types

# Create cross-project agent type
amp-tasks agent-types create "DevOps Engineer" \
  --description "Infrastructure and deployment" \
  --global

# Verify project scoping works
amp-tasks agent-types assign <task-id> <agent-type-slug>  # Validates project access
```

## Key Concepts Learned

### Project Scoping
- **Agent types** are project-specific by default, global with `--global` flag
- **Tasks** belong to specific projects
- **Assignment verification** ensures agent types are available in task's project
- **Global agent types** appear in all projects (useful for shared roles)

### Coordinator vs Worker Workflows
- **Coordinators**: Set up projects, create agent types, assign work strategically
- **Workers**: Find available work, self-identify agent type, update progress

For detailed help: amp-tasks docs agent-guide
For complete docs: amp-tasks docs readme
