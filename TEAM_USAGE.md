# Team Mode Usage Guide

## Overview

Claude Wrapper's team mode allows you to spawn multiple Claude sessions with different personas that collaborate on complex tasks. Each persona has specific capabilities and constraints, and they communicate through a shared workspace.

## Personas

### Solutions Architect (Singleton)
- **Role**: System design and architecture
- **Responsibilities**:
  - Design scalable architectures
  - Evaluate technology choices
  - Create technical specifications
  - Identify risks and mitigation strategies

### Engineering Manager (Singleton)
- **Role**: Technical leadership and coordination
- **Responsibilities**:
  - Break down tasks
  - Coordinate team members
  - Review code quality
  - Balance technical debt with delivery

### Software Engineer (Multiple Allowed)
- **Role**: Feature implementation
- **Responsibilities**:
  - Write production code
  - Implement features
  - Debug and fix issues
  - Write comprehensive tests

### Intern (Multiple Allowed)
- **Role**: Learning-focused contribution
- **Responsibilities**:
  - Handle simpler tasks
  - Write detailed comments
  - Ask questions when unsure
  - Learn from senior team members

## Workspace Structure

When you start a team session, a `.database` directory is created:

```
.database/
├── workspace.json          # Workspace metadata
├── sessions/               # Active session records
│   ├── solutions-architect-*.json
│   ├── engineering-manager-*.json
│   ├── software-engineer-*.json
│   └── intern-*.json
├── messages/               # Inter-persona communication
│   └── msg-*.json
├── outputs/                # Persona deliverables
│   ├── architect/
│   ├── manager/
│   ├── engineer-1/
│   └── intern-1/
└── shared/                 # Shared resources
    └── *.md, *.json, etc.
```

## Communication Protocol

Personas communicate via JSON messages in the `messages/` directory:

```json
{
  "id": "msg-1234567890",
  "from": "engineering-manager-1234",
  "from_persona": "engineering-manager",
  "to": "software-engineer-1",
  "to_persona": "software-engineer",
  "timestamp": "2024-01-20T10:30:00Z",
  "type": "task",
  "subject": "Implement user authentication",
  "content": "Please implement JWT-based authentication...",
  "parent_id": ""
}
```

### Message Types
- **task**: Assignment of work
- **question**: Request for clarification
- **response**: Answer to a question
- **notification**: Status update or announcement

## Example Workflows

### Simple Feature Development

```bash
# Start a basic team
claude-wrapper team start "Build a user authentication system"
```

This creates:
1. Solutions Architect: Designs the auth architecture
2. Engineering Manager: Breaks down into tasks
3. Software Engineer: Implements the features

### Large Project

```bash
# Start a full team
claude-wrapper team start \
  --engineers 3 \
  --interns 2 \
  "Migrate monolithic app to microservices"
```

Workflow:
1. **Architect** designs microservices architecture
2. **Manager** creates task breakdown
3. **Engineers** implement core services
4. **Interns** handle documentation and tests
5. **Manager** reviews and coordinates

### Team Collaboration Flow

```
1. Solutions Architect
   ↓ (posts architecture.md to shared/)

2. Engineering Manager
   ↓ (breaks down into tasks)
   ↓ (posts messages to engineers)

3. Software Engineers
   ↓ (implement features)
   ↓ (write outputs to their directories)

4. Interns
   ↓ (add tests and documentation)

5. Engineering Manager
   ↓ (reviews outputs)
   ↓ (provides feedback via messages)
```

## Best Practices

### 1. Clear Task Definition
```bash
# Good: Specific and actionable
claude-wrapper team start "Build REST API with CRUD operations for User, Post, and Comment models using Express and PostgreSQL"

# Bad: Too vague
claude-wrapper team start "make an API"
```

### 2. Right Team Size
- **Small task**: Default team (1 architect, 1 manager, 1 engineer)
- **Medium task**: 2-3 engineers
- **Large task**: 3+ engineers, 1-2 interns

### 3. Monitor Progress
```bash
# Check status regularly
claude-wrapper team status

# Watch the workspace
watch -n 5 'ls -la .database/outputs/*/'
```

### 4. Review Outputs
```bash
# Check what each persona has produced
ls -la .database/outputs/architect/
ls -la .database/outputs/engineer-1/

# Read messages
jq . .database/messages/*.json
```

## Customizing Personas

You can customize persona behavior:

```bash
# Initialize personas file
claude-wrapper persona init

# Edit ~/.claude-personas.yaml
# Modify instructions, capabilities, or constraints

# Your custom personas will be used automatically
claude-wrapper team start "your task"
```

## Advanced Usage

### Custom Workspace Location
```bash
claude-wrapper team start \
  --workspace ./project-workspace \
  "Build feature X"
```

### Sequential Team Operations
```bash
# Phase 1: Design
claude-wrapper run --persona solutions-architect \
  "Design architecture for feature X" > .database/shared/architecture.md

# Phase 2: Implementation
claude-wrapper team start --engineers 2 \
  "Implement feature X based on .database/shared/architecture.md"
```

## Troubleshooting

### Sessions Not Starting
- Check that `claude` is in your PATH
- Verify workspace directory permissions
- Check for singleton violations (multiple managers/architects)

### Personas Not Communicating
- Verify `.database/messages/` directory exists
- Check message JSON format
- Ensure sessions are active (`team status`)

### Output Not Generated
- Check `.database/outputs/<session-id>/` directories
- Verify persona has write permissions
- Check session logs for errors

## Examples

### Example 1: API Development
```bash
claude-wrapper team start "Create a REST API for a todo app with user authentication, CRUD operations, and SQLite database"
```

Expected outputs:
- `architect/`: System design doc
- `manager/`: Task breakdown
- `engineer-1/`: API implementation
- Test results and documentation

### Example 2: Code Migration
```bash
claude-wrapper team start --engineers 2 --interns 1 \
  "Migrate Python 2.7 codebase to Python 3.11"
```

Expected flow:
- Architect: Migration strategy
- Manager: Priority and task assignment
- Engineers: Code migration
- Intern: Update documentation

### Example 3: Bug Investigation
```bash
claude-wrapper team start --engineers 2 \
  "Debug and fix memory leak in production application"
```

Expected outputs:
- Architect: Root cause analysis
- Manager: Fix strategy
- Engineers: Implementation and tests
