# WildWest

A Go wrapper for Claude Code that enables running Claude in multi-agent environments. 
Quick guide:

```
wildwest team start "Build a REST API for user management" --engineers 2
wildwest orchestrate --workspace .database
# Grab the tmux session ID and run:
tmux attach -t claude-orchestrator-*
```

**How it works:** The orchestrator spawns multiple agents corresponding to the individual persona (Manager, Architect, Software Engineer, Intern etc.) and they work with each other to accomplish the task at hand.
Some of the personas have capabilities to assign tasks and some of the personas has capabilities to add additional resources (claude sessions) and provide instructions to them.

**Communication between agents:** gRPC :x: TCP :x: Markdown files :white_check_mark:

**Disclaimer:** Not ready for production usage.

## Features

- **Multi-Agent Team System**: Hierarchical team of Claude personas working together
- **Tmux Integration**: Each persona runs in persistent background tmux sessions
- **Dynamic Team Scaling**: Personas can request additional engineers/interns as needed
- **Real-time Status Tracking**: View running/stopped/completed sessions at any time
- **Cost Tracking**: Monitor token usage and costs across all personas in real-time
- **Session Management**: Attach/detach from any session, automatic archival on completion
- **4 Persona Types**: Engineering Manager, Solutions Architect, Software Engineer, Intern

## Installation

```bash
# Build the project
go build -o wildwest

# Install to your PATH
go install
```

## Configuration

### Environment Variables

- **CLAUDE_BIN**: Path to the claude binary if installed in a custom location
  ```bash
  export CLAUDE_BIN=/path/to/custom/claude
  ```

## Quick Start

```bash
# 1. Create a team
wildwest team start "Build a REST API for user management" --engineers 2

# 2. Start the orchestrator (runs in tmux in the background)
wildwest orchestrate --workspace .database
# Returns immediately with orchestrator session name

# 3. View running sessions (including orchestrator)
tmux ls | grep claude
wildwest attach --list

# 4. Attach to orchestrator to monitor progress
tmux attach -t claude-orchestrator-*

# 5. Attach to any persona session (Ctrl+B then D to detach)
wildwest attach                     # Attach to manager (default)
wildwest attach <session-id>       # Attach to specific session

# 6. Add instructions to any persona
cat >> .database/engineering-manager-*/instructions.md <<EOF
## Instructions from User ($(date))
Please create API endpoints for user CRUD operations
EOF
# Claude detects new instructions within 5 seconds

# 7. Clean up stopped sessions
wildwest cleanup --workspace .database
```

## Usage

### Multi-Persona Team Collaboration (Orchestrator-Based)

The system uses a **Project Manager orchestrator** that runs in its own tmux session and dynamically spawns and manages Claude instances in **tmux sessions**:

```bash
# 1. Create team structure
wildwest team start "Build a REST API for user management"

# 2. Start orchestrator (returns immediately, runs in tmux background)
wildwest orchestrate --workspace .database
# Output: Session Name: claude-orchestrator-1234567890

# 3. View all sessions (including orchestrator)
tmux ls | grep claude
wildwest attach --list              # List all persona instances

# 4. Attach to orchestrator to monitor
tmux attach -t claude-orchestrator-1234567890

# 5. Attach to persona sessions
wildwest attach                     # Attach to manager (default)
wildwest attach <session-id>       # Attach to specific session

# Kill orchestrator (stops all management)
tmux kill-session -t claude-orchestrator-*

# Kill all Claude sessions (including orchestrator)
tmux kill-server
```

#### Dynamic Team Growth

Personas can request additional team members by creating directories:

```bash
# Manager/Architect requests an engineer
mkdir .database/software-engineer-request-api-developer
echo "Implement API endpoints" > .database/software-engineer-request-api-developer/instructions.md
# Orchestrator automatically spawns the engineer

# Engineer requests an intern
mkdir .database/intern-request-test-writer
echo "Write unit tests" > .database/intern-request-test-writer/instructions.md
# Orchestrator automatically spawns the intern
```

#### How Team Collaboration Works

1. **Workspace Structure**: Each persona gets their own directory:
   - `engineering-manager-{timestamp}/` - Manager's workspace
   - `solutions-architect-{timestamp}/` - Architect's workspace
   - `software-engineer-{timestamp}/` - Engineer workspaces
   - `intern-{timestamp}/` - Intern workspaces
   - `shared/` - Common resources

2. **Persona Roles**:
   - **Project Manager**: Orchestrator that spawns/manages Claude instances
   - **Engineering Manager** (1 max): Writes requirements, requests engineers
   - **Solutions Architect** (1 max): Designs systems, requests engineers
   - **Software Engineers** (multiple): Implement features, request interns
   - **Interns** (multiple): Handle tests, linting, docs

3. **Tmux Sessions**:
   - Each persona runs in its own tmux session: `claude-{session-id}`
   - Claude runs interactively in the background
   - Attach/detach at any time without disrupting work
   - Sessions persist until tasks complete or manually killed

4. **Automatic Instruction Monitoring**:
   - Each Claude instance creates a background Bash task on startup
   - Monitors `instructions.md` every 5 seconds for new content
   - Detects file size changes and reads new instructions immediately
   - No manual polling required - fully autonomous

5. **Status Tracking**:
   - `wildwest attach --list` shows real-time status of all sessions
   - Detects when tmux sessions stop or are killed
   - `wildwest cleanup` archives stopped sessions
   - Sessions marked as: running 🟢, stopped ⏸️, completed ✅, failed ❌

6. **Dynamic Spawning**:
   - Manager/Architect create `software-engineer-request-{name}/` directories
   - Engineers create `intern-request-{name}/` directories
   - Orchestrator detects requests and spawns Claude instances
   - Completed sessions automatically archived

### Monitor Token Usage and Costs

The orchestrator automatically tracks token usage and calculates costs for all active personas:

```bash
# Show current cost snapshot
wildwest team cost

# Watch costs update in real-time (updates every minute)
wildwest team cost --watch
```

**Output example:**
```
💰 Team Cost Summary
====================

📊 Ada Lovelace (engineering-manager)
   Session: engineering-manager-1234567890
   Model: sonnet
   Input Tokens: 45,230
   Output Tokens: 12,450
   Total Tokens: 57,680
   Cost: $0.3230
   Last Updated: 2026-01-26 15:30:45

📊 Alan Turing (software-engineer)
   Session: software-engineer-1234567891
   Model: sonnet
   Input Tokens: 32,100
   Output Tokens: 8,900
   Total Tokens: 41,000
   Cost: $0.2300
   Last Updated: 2026-01-26 15:30:45

====================
💵 Total Team Cost: $0.5530
```

**How it works:**
- Token usage is automatically polled every minute from tmux sessions
- Costs are calculated based on Claude model pricing:
  - Sonnet: $3/MTok input, $15/MTok output
  - Opus: $15/MTok input, $75/MTok output
  - Haiku: $0.25/MTok input, $1.25/MTok output
- Usage data is stored in each session's `tokens.json` file
- No manual tracking needed - fully automated

### Run Claude with custom environment

```bash
# Run in development environment
wildwest run --env development "Add user authentication"

# Run with custom specs
wildwest run --spec "Use TypeScript" --spec "Add tests" "Create API endpoint"

# Run with custom instructions file
wildwest run --instructions ./custom-instructions.md "Build feature"
```

### Expand minimal prompts

```bash
# Expand a minimal prompt into detailed instructions
wildwest expand "Add login feature"

# Expand with environment context
wildwest expand --env production "Optimize database queries"
```

### Persona Management

```bash
# List available personas
wildwest persona list

# Show details of a specific persona
wildwest persona show software-engineer

# Initialize default personas file for customization
wildwest persona init
```

### Run with a specific persona

```bash
# Run as a Software Engineer
wildwest run --persona software-engineer "Add authentication middleware"

# Run as an Intern (will include detailed comments and ask questions)
wildwest run --persona intern "Create unit tests for auth module"

# Run as Solutions Architect
wildwest run --persona solutions-architect "Design caching strategy"
```

### Session Management

```bash
# List all sessions with status
wildwest attach --list

# Attach to manager (default)
wildwest attach

# Attach to specific session
wildwest attach engineering-manager-1234567890

# Filter sessions by type
wildwest attach --list --filter engineer

# Clean up stopped sessions (archives them)
wildwest cleanup --workspace .database

# View all tmux sessions
tmux ls

# Attach to tmux session directly
tmux attach -t claude-engineering-manager-1234567890

# Detach from tmux
# Press: Ctrl+B, then D
```

### List available environments

```bash
# Show all configured environments
wildwest list

# List available persona names
wildwest names
```

### Examples

#### One-off execution with custom specs
```bash
wildwest run \
  --spec "Use React hooks" \
  --spec "Include TypeScript types" \
  --spec "Add error boundaries" \
  "Create user dashboard component"
```

#### Run with environment and instructions
```bash
wildwest run \
  --env development \
  --instructions ./project-guidelines.md \
  "Implement payment processing"
```

#### Expand minimal prompt for review
```bash
# First expand to see what Claude will do
wildwest expand "Refactor auth module"

# Then execute if it looks good
wildwest run --env development "Refactor auth module"
```

## Architecture

```
wildwest/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and config initialization
│   ├── run.go             # Run command
│   ├── expand.go          # Expand command
│   └── list.go            # List command
├── pkg/                    # Internal packages
│   ├── config/            # Configuration management
│   │   └── config.go
│   └── claude/            # Claude execution logic
│       └── executor.go
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── .gitignore
└── README.md
```

## Troubleshooting

### Tmux Session Conflicts

If new sessions fail to spawn due to conflicts:

```bash
# List all Claude tmux sessions (including orchestrator)
tmux ls | grep claude

# Kill orchestrator (stops spawning new sessions)
tmux kill-session -t claude-orchestrator-*

# Kill all Claude persona sessions
tmux ls 2>/dev/null | grep "claude-engineering\|claude-software\|claude-solutions" | cut -d: -f1 | xargs -I {} tmux kill-session -t {}

# Kill everything (orchestrator + all personas)
tmux kill-server

# Clean up database
wildwest cleanup --workspace .database
```

### Session Not Running

If `wildwest attach` says session not running:

```bash
# Check if orchestrator is running
ps aux | grep "wildwest orchestrate"

# Restart orchestrator
wildwest orchestrate --workspace .database
```

### Files Appear Corrupt

All files in `.database/` should be valid JSON/markdown. If you see corruption:

```bash
# Check file types
find .database -type f -name "*.md" -o -name "*.json" | xargs file

# Verify JSON files
find .database -name "*.json" -exec sh -c 'echo "{}:" && jq . "{}" >/dev/null 2>&1 && echo "✓ Valid" || echo "✗ Invalid"' \;
```

### Background Task Not Working

If Claude isn't detecting new instructions:

```bash
# Attach to session and check for background task
tmux attach -t claude-engineering-manager-*

# Look for "1 bash" in status bar (bottom right)
# This indicates the background monitoring task is running
```

## Development

```bash
# Run tests
go test ./...

# Build
make build

# Run locally
./bin/wildwest --help

# Format code
go fmt ./...

# Lint
golangci-lint run
```

## Use Cases

1. **Team-based Development**: Spawn multiple personas to work on complex projects
   - Architect designs the system
   - Manager breaks down tasks
   - Engineers implement features
   - Interns handle simpler tasks

2. **Multi-project Management**: Define different environments for different projects

3. **Team Standards**: Share environment configs to enforce team coding standards

4. **CI/CD Integration**: Use specific environments in automated pipelines

5. **Quick Prototyping**: Use minimal prompts with expand to quickly iterate on ideas

6. **Context-aware Execution**: Automatically set up the right context for each task

7. **Code Review Simulation**: Have manager persona review engineer outputs

8. **Learning and Mentorship**: Intern personas learn from engineer personas



### Configuration File

Create a configuration file at `~/.wildwest.yaml`:

```yaml
# Default path to claude binary (overridden by CLAUDE_BIN env var)
claude_path: "claude"

# Define custom environments
environments:
  # Development environment
  development:
    description: "Development environment with debugging enabled"
    working_dir: "/path/to/project"
    env_vars:
      DEBUG: "true"
      LOG_LEVEL: "debug"
    default_specs:
      - "Use detailed logging"
      - "Prefer readable code over performance"
    pre_commands:
      - "echo 'Starting development session'"
    post_commands:
      - "echo 'Development session complete'"

  # Production environment
  production:
    description: "Production environment with optimizations"
    working_dir: "/path/to/production"
    env_vars:
      DEBUG: "false"
      LOG_LEVEL: "error"
    default_specs:
      - "Optimize for performance"
      - "Include error handling"
      - "Add comprehensive tests"

  # Testing environment
  testing:
    description: "Testing environment"
    working_dir: "/path/to/tests"
    default_specs:
      - "Generate unit tests"
      - "Include edge cases"

# Reusable prompt templates
templates:
  refactor: "Refactor the code to improve readability and maintainability"
  optimize: "Optimize the code for better performance"
  debug: "Debug and fix issues in the code"
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
