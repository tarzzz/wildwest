# Changelog

## v2.1 (2024-01-26) - Named Personas

### Added
- **Interesting Persona Names**: 130 famous names from history
  - 20 Scientists (Einstein, Curie, Tesla, etc.)
  - 20 Artists (Picasso, Da Vinci, Monet, etc.)
  - 20 Musicians (Mozart, Beethoven, Bach, etc.)
  - 20 Writers (Shakespeare, Hemingway, Tolkien, etc.)
  - 20 Philosophers (Socrates, Plato, Aristotle, etc.)
  - 20 Inventors (Edison, Bell, Wright, etc.)
  - 10 Explorers (Magellan, Armstrong, Hillary, etc.)

- **Smart Name Assignment**: Names assigned by persona type
  - Managers → Philosophers
  - Architects → Artists or Inventors
  - Engineers → Scientists or Inventors
  - Interns → Writers or Explorers

- **New Command**: `claude-wrapper names`
  - Lists all 130 available names by category
  - Shows assignment strategy

### Changed
- Session creation now auto-generates interesting names
- Orchestrator displays persona names in spawn messages
- Team start shows assigned names
- Track dashboard shows names instead of IDs
- Attach command lists sessions with names

### Technical
- New package: `pkg/names/names.go`
- Name generator with uniqueness tracking
- Thread-safe name assignment
- Automatic loading of existing names on startup

### Example Output
```
Creating Engineering Manager directory (Level 1)...
  Name: kant
  Directory: engineering-manager-1706012345678

Creating Solutions Architect directory (Level 2)...
  Name: davinci
  Directory: solutions-architect-1706012345679
```

---

## v2.0 (2024-01-26) - Orchestrator Edition

### Major Changes
- **Dynamic Team Management**: From static to orchestrator-based spawning
- **Active Orchestration**: Project Manager now manages Claude instances
- **Request-Based Spawning**: Create directories to request team members
- **Automatic Cleanup**: Completed sessions auto-archived

### Added
- **New Command**: `claude-wrapper orchestrate`
  - Daemon mode for managing Claude instances
  - Watches for spawn requests
  - Monitors running processes
  - Auto-archives completed sessions

- **New Command**: `claude-wrapper attach [session-id]`
  - List all running sessions
  - Attach to interactive shell
  - Filter by persona type
  - Default attach to manager

- **Dynamic Spawning**: Personas request team members via directories
  - Manager/Architect: Create `software-engineer-request-{name}/`
  - Engineers: Create `intern-request-{name}/`
  - Orchestrator spawns automatically

- **Process Management**:
  - PID tracking for all Claude instances
  - Process health monitoring
  - Graceful shutdown on completion
  - Exit status tracking

### Changed
- `team start` now creates directory structure only (no spawning)
- Orchestrator must be started separately
- Session directories named with timestamps
- Completed sessions archived to `{session-id}-completed`

### Documentation
- New: `ORCHESTRATOR.md` - Complete orchestrator architecture
- New: `FINAL_STATUS.md` - Project status
- Updated: `README.md` - New workflow
- Updated: `QUICKSTART.md` - Updated examples

---

## v1.0 (2024-01-26) - Initial Release

### Core Features
- Multi-persona system (5 personas)
- Hierarchical communication
- File-based collaboration
- State tracking with tracker.json

### Personas
- Project Manager (Read-only tracker)
- Engineering Manager (Level 1)
- Solutions Architect (Level 2)
- Software Engineers (Level 3, multiple)
- Interns (Level 4, multiple)

### Commands
- `claude-wrapper team start` - Start team sessions
- `claude-wrapper team status` - Check status
- `claude-wrapper team stop` - Stop all
- `claude-wrapper track` - Status dashboard
- `claude-wrapper persona list` - List personas
- `claude-wrapper run` - Run single persona

### Workspace
- Directory per persona
- tasks.md - Task management
- instructions.md - Communication
- tracker.json - State tracking
- shared/ - Common resources

### Documentation
- README.md
- ARCHITECTURE.md
- TEAM_USAGE.md
- QUICKSTART.md
- SUMMARY.md
