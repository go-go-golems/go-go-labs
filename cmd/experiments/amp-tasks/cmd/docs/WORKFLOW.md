# Agent Workflow

## Daily Work Cycle

1. **Check Context**
   amp-tasks projects default
   - Read project guidelines
   - Understand current objectives

2. **Understand the Agent System**
   amp-tasks agent-types list
   amp-tasks agents list
   - See what agent types exist
   - Understand the workforce structure

3. **Find Available Work**
   amp-tasks tasks available
   - See tasks ready for assignment
   - Check dependencies are met

4. **Read Previous Work (if dependencies exist)**
   amp-tasks deps show <task-id>
   - Review notes from agents who worked on dependency tasks
   - Learn from TIL entries related to this work

5. **Assign Work Strategically**
   # By agent type (flexible distribution)
   amp-tasks agent-types assign <task-id> <type-id>
   
   # By specific agent (targeted assignment)
   amp-tasks tasks assign <task-id> <your-agent-id>
   - Task status automatically becomes 'in_progress'

6. **Do the Work**
   - Follow project guidelines
   - Complete the task requirements  
   - Take progress notes: amp-tasks notes add <task-id> "progress update"
   - Check task details: amp-tasks tasks show <task-id>

7. **Share Learning**
   amp-tasks til create "insight title" --content "what you learned"
   - Create task-specific TIL: --task <task-id>
   - Share insights that help other agents

8. **Update Status**
   amp-tasks tasks status <task-id> completed
   - System shows newly available tasks
   - Dependencies are automatically resolved

9. **Create Additional Work (if needed)**
   amp-tasks tasks create "New task" --description "Details"
   amp-tasks deps add <new-task> <depends-on-task>

## Key Principles

- Always check project guidelines first
- Understand the agent type system before assigning work
- Choose assignment strategy based on work requirements
- Only work on available tasks (dependencies met)
- Read notes from dependency tasks to understand context
- Document your progress with notes for transparency
- Share valuable insights through TIL entries
- Update status promptly
- Create clear, actionable tasks
- Follow the dependency chain

## Getting Help

- amp-tasks docs agent-guide    # Essential commands
- amp-tasks docs readme         # Complete documentation  
- amp-tasks --help              # Command reference
- amp-tasks <command> --help    # Specific command help
