# Claude Wrapper - Project Summary

## What Was Built

A comprehensive Go-based wrapper for Claude Code that enables multi-agent team collaboration with hierarchical communication and task management.

## Key Features

### 1. Multi-Persona System (5 Personas)

#### Project Manager (Read-Only Tracker)
- **Role**: Monitoring and status tracking
- **Capabilities**: Read all personas' tasks.md and instructions.md
- **Constraints**: Cannot write to any persona's files
- **Command**: `claude-wrapper track`

#### Engineering Manager (Level 1 - TOP)
- **Role**: Project leadership and high-level planning
- **Capabilities**: Write project requirements, give instructions to Solutions Architect
- **Hierarchy**: Top of the chain

#### Solutions Architect (Level 2)
- **Role**: System design, architecture, data modeling
- **Capabilities**: Create designs, diagrams, technical specs
- **Receives from**: Engineering Manager
- **Gives to**: Software Engineers

#### Software Engineers (Level 3) - Multiple Allowed
- **Role**: Major feature implementation
- **Capabilities**: Production code, complex features
- **Receives from**: Solutions Architect
- **Gives to**: Interns

#### Interns (Level 4 - BOTTOM) - Multiple Allowed
- **Role**: Minor tasks (tests, linting, documentation)
- **Receives from**: Software Engineers
- **Gives to**: No one (bottom of hierarchy)

### 2. Workspace Structure

```
.database/
├── shared/                          # Shared resources
├── engineering-manager-*/           # Manager's directory
│   ├── session.json                # Session metadata
│   ├── tasks.md                    # Manager's tasks
│   ├── instructions.md             # Instructions received
│   ├── tracker.json                # Read state tracking
│   └── *.md, *.go                  # Outputs
├── solutions-architect-*/           # Architect's directory
├── software-engineer-1-*/           # Engineer directories
└── intern-1-*/                      # Intern directories
```

### 3. Communication System

#### Hierarchy Flow
```
Engineering Manager
    ↓ instructions.md
Solutions Architect
    ↓ instructions.md
Software Engineers
    ↓ instructions.md
Interns
```

#### File Purposes
- **tasks.md**: Each persona maintains their own task list with statuses
  - Statuses: "not started", "in progress", "completed"
  - Only owner can modify

- **instructions.md**: Timestamped instructions from other personas
  - Appended with timestamp headers
  - Read-only for recipient

- **tracker.json**: Tracks reading state
  - Last read position (character/byte)
  - Last read timestamp
  - Helps resume after disconnect

### 4. State Management

Each persona has a tracker that maintains:
- Last character position read in instructions.md
- Last character position read in tasks.md
- Timestamps of last reads
- Enables resuming after disconnection

### 5. CLI Commands

#### Team Management
```bash
# Start a team
claude-wrapper team start "Build REST API" --engineers 2 --interns 1

# Check team status
claude-wrapper team status

# Stop team
claude-wrapper team stop
```

#### Monitoring
```bash
# Track team progress (Project Manager view)
claude-wrapper track --workspace .database
```

#### Persona Management
```bash
# List personas
claude-wrapper persona list

# Show persona details
claude-wrapper persona show software-engineer

# Initialize personas file
claude-wrapper persona init
```

#### Individual Runs
```bash
# Run with a specific persona
claude-wrapper run --persona software-engineer "Implement auth"

# Run with environment and persona
claude-wrapper run --env production --persona solutions-architect "Design caching"
```

## Project Structure

```
claude-wrapper/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── run.go             # Run single persona
│   ├── team.go            # Team management
│   ├── track.go           # Project manager tracking
│   ├── persona.go         # Persona management
│   ├── expand.go          # Prompt expansion
│   └── list.go            # List configs
├── pkg/
│   ├── config/            # Configuration
│   │   └── config.go
│   ├── persona/           # Persona definitions
│   │   └── persona.go     # 5 personas defined
│   ├── claude/            # Claude execution
│   │   └── executor.go
│   └── session/           # Session management
│       └── session.go     # Tracker, tasks, instructions
├── main.go
├── go.mod
├── Makefile
├── README.md
├── ARCHITECTURE.md         # Detailed architecture
├── TEAM_USAGE.md          # Team usage guide
└── SUMMARY.md             # This file
```

## How It Works

### Starting a Team

1. User runs: `claude-wrapper team start "Build feature X" --engineers 2`

2. System creates:
   - Workspace directory (.database/)
   - Session for each persona in hierarchy order
   - tasks.md, tracker.json for each session
   - Initial instructions

3. Each persona:
   - Gets spawned as separate Claude process
   - Reads their instructions.md regularly
   - Updates their tasks.md with progress
   - Writes to other personas' instructions.md

### Communication Flow Example

1. Manager writes to `solutions-architect-*/instructions.md`:
   ```markdown
   ## Instructions from engineering-manager-123 (2024-01-20 10:00:00)

   Design REST API for user management...
   ```

2. Architect reads instructions.md (tracker updates)
3. Architect creates design docs in their directory
4. Architect writes to `software-engineer-1-*/instructions.md`
5. Engineer reads, implements, updates tasks.md
6. Engineer writes to `intern-1-*/instructions.md`
7. Intern reads, writes tests, updates tasks.md

### Tracking Progress

```bash
$ claude-wrapper track

═══════════════════════════════════════════════════
           PROJECT STATUS DASHBOARD
═══════════════════════════════════════════════════

╔══════════════════════════════════════════════════╗
║  ENGINEERING MANAGER
╚══════════════════════════════════════════════════╝

📋 manager (engineering-manager-1234567890)
   Status: active
   Started: 2024-01-20 10:00:00

   Current Tasks:
      ✅ Write project requirements [completed]
      🔄 Review deliverables [in progress]

   Latest Instructions:
      ## Instructions from system
      Initial project setup

╔══════════════════════════════════════════════════╗
║  SOFTWARE ENGINEERS
╚══════════════════════════════════════════════════╝

📋 engineer-1 (software-engineer-1234567891)
   Status: active

   Current Tasks:
      🔄 Implement authentication [in progress]
      ⏸️ Add error handling [not started]
```

## Technical Highlights

### 1. Hierarchical Constraints
- Singletons: Manager, Architect, Project Manager
- Multiple: Engineers, Interns
- Enforced at session creation

### 2. File-Based Communication
- No network required
- Simple, debuggable
- Works with Claude Code directly

### 3. State Recovery
- tracker.json enables resume
- Character-position tracking
- Timestamp-based checks

### 4. Extensible Design
- Easy to add new personas
- Configurable via YAML
- Custom environments supported

## Usage Examples

### Small Project
```bash
claude-wrapper team start "Create todo app API"
# Uses default: 1 manager, 1 architect, 1 engineer
```

### Medium Project
```bash
claude-wrapper team start --engineers 3 "Migrate to microservices"
```

### Large Project
```bash
claude-wrapper team start --engineers 4 --interns 2 "Build e-commerce platform"
```

### Monitor Progress
```bash
# Watch progress in real-time
watch -n 10 'claude-wrapper track'
```

## Configuration Files

### ~/.claude-wrapper.yaml
```yaml
claude_path: "claude"
environments:
  dev:
    description: "Development environment"
    env_vars:
      DEBUG: "true"
```

### ~/.claude-personas.yaml
```yaml
personas:
  custom-persona:
    name: "Custom Role"
    instructions: "..."
```

## Compilation

```bash
# Build
make build

# Install
make install

# Run tests
make test

# Clean
make clean
```

## Benefits

1. **Structured Collaboration**: Clear hierarchy and communication
2. **Scalable**: Add more engineers/interns as needed
3. **Traceable**: All communication is timestamped and logged
4. **Resumable**: State tracking enables recovery
5. **Monitoring**: Project Manager view shows real-time status
6. **Flexible**: Custom personas and environments
7. **Simple**: File-based, no complex infrastructure

## Future Enhancements (Possible)

- Web UI for monitoring
- Real-time notifications
- Automatic task assignment
- Performance metrics
- Integration with issue trackers
- Slack/Discord notifications
- Git integration for commits
- Code review workflows
