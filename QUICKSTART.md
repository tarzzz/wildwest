
# Quick Start Guide

## Installation

```bash
make build
make install  # Optional: installs to $GOPATH/bin
```

## Your First Team Project

### 1. Start a Team Structure

```bash
# Terminal 1: Create workspace and initial personas
./bin/wildwest team start "Build a REST API for a blog with posts and comments"
```

This creates directories for:
- 1 Engineering Manager
- 1 Solutions Architect
- Workspace at `.database/`

### 2. Start the Orchestrator

```bash
# Terminal 2: Start the orchestrator daemon
./bin/wildwest orchestrate --workspace .database
```

The orchestrator will:
- Spawn Claude instances for Manager and Architect
- Watch for new spawn requests
- Manage Claude process lifecycle
- Archive completed sessions

### 3. Monitor and Attach

```bash
# Terminal 3: Monitor progress
./bin/wildwest attach --list       # List all sessions
./bin/wildwest attach               # Attach to manager
./bin/wildwest track               # Status dashboard

# Or watch in real-time
watch -n 5 './bin/wildwest track'
```

You'll see:
- What each persona is working on
- Task completion status
- Recent instructions
- Output files created

### 3. Inspect the Workspace

```bash
# View directory structure
tree .database/

# Read manager's tasks
cat .database/engineering-manager-*/tasks.md

# Read architect's outputs
ls .database/solutions-architect-*/

# Check what instructions were given
cat .database/software-engineer-*/instructions.md
```

### 4. Check Individual Outputs

```bash
# See what the architect designed
cat .database/solutions-architect-*/system-design.md

# See what the engineer implemented
ls .database/software-engineer-*/*.go

# Check the tracker state
cat .database/engineering-manager-*/tracker.json
```

## Dynamic Team Growth Example

### How Personas Request More Team Members

#### From Manager/Architect Session

```bash
# Attach to manager
./bin/wildwest attach

# Inside manager's session, request an engineer
$ mkdir ../software-engineer-request-backend-dev
$ cat > ../software-engineer-request-backend-dev/instructions.md <<EOF
Implement backend API for product catalog
Follow the architecture in shared/api-spec.md
EOF

# Orchestrator automatically spawns the engineer
# New directory appears: software-engineer-1234567890/
```

#### From Engineer Session

```bash
# Attach to engineer
./bin/wildwest attach software-engineer-1234567890

# Request an intern
$ mkdir ../intern-request-test-writer
$ cat > ../intern-request-test-writer/instructions.md <<EOF
Write unit tests for product_handler.go
Ensure >80% coverage
EOF

# Orchestrator spawns the intern
# New directory: intern-1234567891/
```

### Workflow That Happens

1. **Manager** writes requirements in their directory
2. **Manager** requests engineers via request directories
3. **Orchestrator** spawns engineer Claude instances
4. **Architect** designs system in their directory
5. **Architect** writes to engineer instructions.md files
6. **Engineers** implement and request interns
7. **Orchestrator** spawns intern Claude instances
8. **Interns** write tests and fix linting
9. **Orchestrator** archives completed sessions

## Using Individual Personas

### Run as Engineering Manager

```bash
./bin/wildwest run \
  --persona engineering-manager \
  "Analyze requirements for user authentication system"
```

### Run as Solutions Architect

```bash
./bin/wildwest run \
  --persona solutions-architect \
  "Design microservices architecture for scaling to 1M users"
```

### Run as Software Engineer

```bash
./bin/wildwest run \
  --persona software-engineer \
  "Implement JWT authentication middleware in Go"
```

### Run as Intern

```bash
./bin/wildwest run \
  --persona intern \
  "Write unit tests for the authentication middleware"
```

## Understanding the Workspace

### Directory Layout After Team Start

```
.database/
├── shared/
├── engineering-manager-1706012345678/
│   ├── session.json
│   ├── tasks.md
│   ├── instructions.md
│   ├── tracker.json
│   └── project-requirements.md
├── solutions-architect-1706012345679/
│   ├── session.json
│   ├── tasks.md
│   ├── instructions.md
│   ├── tracker.json
│   ├── system-design.md
│   └── api-spec.md
├── software-engineer-1-1706012345680/
│   ├── session.json
│   ├── tasks.md
│   ├── instructions.md
│   ├── tracker.json
│   ├── main.go
│   └── handlers.go
└── intern-1-1706012345681/
    ├── session.json
    ├── tasks.md
    ├── instructions.md
    ├── tracker.json
    └── main_test.go
```

### Communication Flow Visualization

```
┌─────────────────────────────────────────┐
│  Engineering Manager                    │
│  - Writes requirements                  │
│  - Reviews final deliverables           │
└───────────────┬─────────────────────────┘
                │ instructions.md
                ↓
┌─────────────────────────────────────────┐
│  Solutions Architect                    │
│  - Designs architecture                 │
│  - Creates technical specs              │
└───────────────┬─────────────────────────┘
                │ instructions.md
                ↓
┌─────────────────────────────────────────┐
│  Software Engineers (multiple)          │
│  - Implement features                   │
│  - Write production code                │
└───────────────┬─────────────────────────┘
                │ instructions.md
                ↓
┌─────────────────────────────────────────┐
│  Interns (multiple)                     │
│  - Write tests                          │
│  - Fix linting                          │
└─────────────────────────────────────────┘
```

## Monitoring with Project Manager

The `track` command acts as a Project Manager persona:

```bash
./bin/wildwest track

# Output shows:
# - Active team members
# - Current tasks and status
# - Recent instructions
# - Output files created
# - Overall progress percentage
```

## Task Status Updates

Each persona updates their `tasks.md`:

```markdown
# Tasks

## Task: Design system architecture
- **Status**: completed
- **Assigned by**: engineering-manager-123
- **Created**: 2024-01-20 10:00:00

## Task: Create API documentation
- **Status**: in progress
- **Assigned by**: engineering-manager-123
- **Created**: 2024-01-20 11:00:00
```

## Instructions Format

When personas communicate via `instructions.md`:

```markdown
---
## Instructions from engineering-manager-123 (2024-01-20 10:00:00)

Please design a REST API for the blog system with the following requirements:

- Posts resource with CRUD operations
- Comments nested under posts
- User authentication with JWT
- Rate limiting on all endpoints
- PostgreSQL as the database

Deliverables:
1. System architecture diagram
2. API specification (OpenAPI)
3. Database schema
4. Technology recommendations

Please write your design to your directory and then provide implementation instructions to the engineers.
```

## State Recovery with Tracker

The `tracker.json` file helps personas resume:

```json
{
  "session_id": "software-engineer-1-1706012345680",
  "instructions_last_read": "2024-01-20T10:30:00Z",
  "instructions_last_position": 1024,
  "tasks_last_read": "2024-01-20T10:25:00Z",
  "tasks_last_position": 512,
  "last_check_time": "2024-01-20T10:30:00Z"
}
```

If a persona disconnects and reconnects, it reads from `instructions_last_position` to catch up on new instructions.

## Tips

### 1. Use Descriptive Task Descriptions
```bash
# Good
wildwest team start "Build REST API with JWT auth, CRUD for posts/comments, PostgreSQL"

# Less effective
wildwest team start "make an API"
```

### 2. Monitor Regularly
```bash
# Set up a watch window
watch -n 10 './bin/wildwest track'
```

### 3. Review Intermediate Outputs
```bash
# Check architect's design before engineers start
cat .database/solutions-architect-*/system-design.md
```

### 4. Adjust Team Size Based on Complexity
- Simple task: Default (1 manager, 1 architect, 1 engineer)
- Medium task: 2-3 engineers
- Complex task: 3+ engineers + 1-2 interns

## Troubleshooting

### No Output in Persona Directories
- Check if Claude is running: `ps aux | grep claude`
- Verify workspace path: `ls -la .database/`
- Check session status: `./bin/wildwest team status`

### Personas Not Communicating
- Verify instructions.md exists: `ls .database/*/instructions.md`
- Check if tracker is updating: `cat .database/*/tracker.json`
- Ensure timestamps are present in instructions

### Tasks Not Updating
- Each persona only updates their own tasks.md
- Check if persona is active: `./bin/wildwest team status`
- Review tracker to see if persona is reading

## Next Steps

1. **Customize Personas**:
   ```bash
   ./bin/wildwest persona init
   # Edit ~/.claude-personas.yaml
   ```

2. **Configure Environments**:
   ```bash
   # Create ~/.wildwest.yaml
   # Add custom environments
   ```

3. **Integrate with Your Workflow**:
   - Use in CI/CD pipelines
   - Automate team starts for new features
   - Create monitoring dashboards

4. **Extend the System**:
   - Add custom personas
   - Create notification hooks
   - Build web UI for tracking
