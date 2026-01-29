# Claude Wrapper Architecture

## Overview

Claude Wrapper is a multi-agent system where different personas (roles) collaborate on tasks through a shared workspace with individual directories.

## Directory Structure

```
.database/                          # Workspace root
├── shared/                         # Shared files accessible to all personas
│   ├── architecture.md
│   ├── requirements.md
│   └── common-utils.go
├── solutions-architect-*/          # Architect's directory
│   ├── session.json               # Session metadata
│   ├── tasks.md                   # Architect's task list (only architect updates)
│   ├── instructions.md            # Instructions from other personas (read-only for architect)
│   ├── system-design.md           # Architect's outputs
│   └── architecture-diagram.png
├── engineering-manager-*/          # Manager's directory
│   ├── session.json
│   ├── tasks.md                   # Manager's task list (only manager updates)
│   ├── instructions.md
│   ├── task-breakdown.md
│   └── review-notes.md
├── software-engineer-1-*/          # Engineer 1's directory
│   ├── session.json
│   ├── tasks.md                   # Engineer's task list (only this engineer updates)
│   ├── instructions.md            # Instructions from manager/architect
│   ├── feature-implementation.go
│   └── unit-tests.go
└── intern-1-*/                     # Intern's directory
    ├── session.json
    ├── tasks.md                   # Intern's task list (only this intern updates)
    ├── instructions.md
    ├── documentation.md
    └── test-cases.md
```

## Communication Model

### 1. Task Assignment Flow (Hierarchy)

```
Engineering Manager (TOP)
         ↓ instructions.md
Solutions Architect
         ↓ instructions.md
Software Engineers
         ↓ instructions.md
Interns (BOTTOM)
```

**Communication Rules:**
- Engineering Manager → instructs Solutions Architect
- Solutions Architect → instructs Software Engineers
- Software Engineers → instruct Interns
- Interns → do NOT instruct anyone (bottom of hierarchy)

### 2. Task Management

Each persona maintains their own `tasks.md`:

```markdown
# Tasks

## Task: Design system architecture
- **Status**: completed
- **Assigned by**: system
- **Created**: 2024-01-20 10:00:00

## Task: Review engineer implementations
- **Status**: in progress
- **Assigned by**: engineering-manager-123
- **Created**: 2024-01-20 11:00:00

## Task: Create API documentation
- **Status**: not started
- **Assigned by**: engineering-manager-123
- **Created**: 2024-01-20 11:30:00
```

### 3. Instructions Format

When assigning work to another persona, write to their `instructions.md`:

```markdown
---
## Instructions from engineering-manager-123 (2024-01-20 11:00:00)

Please implement the user authentication feature based on the architecture in shared/architecture.md.

Requirements:
- Use JWT for tokens
- Implement password hashing with bcrypt
- Add rate limiting for login attempts
- Write comprehensive unit tests

Please update your tasks.md with progress and write your implementation to your directory.
```

## Persona Constraints and Hierarchy

### Engineering Manager (Level 1 - TOP)
- **Singleton**: Only one manager per workspace
- **Position**: Top of hierarchy
- **Can write to**: Solutions Architect's instructions.md, shared/
- **Cannot write to**: Engineers or Interns directly (goes through architect)
- **Responsibilities**:
  - Understand project requirements
  - Write detailed project summaries
  - Provide high-level direction to Solutions Architect
  - Review all major deliverables
  - Make final technical decisions
  - Coordinate overall project
- **Communication**: Gives instructions TO Solutions Architect only

### Solutions Architect (Level 2)
- **Singleton**: Only one architect per workspace
- **Position**: Second in hierarchy, reports to Manager
- **Can write to**: Software Engineers' instructions.md, shared/
- **Cannot write to**: Manager's instructions.md (only receives from manager)
- **Responsibilities**:
  - Read instructions from Engineering Manager
  - Design system architecture and data models
  - Create system diagrams and technical specs
  - Provide implementation guidance to Engineers
  - Ensure architectural consistency
- **Communication**:
  - RECEIVES instructions FROM Engineering Manager
  - GIVES instructions TO Software Engineers

### Software Engineers (Level 3)
- **Multiple allowed**: Can have many engineers
- **Position**: Third in hierarchy, report to Architect
- **Can write to**: Interns' instructions.md, shared/
- **Cannot write to**: Manager or Architect's instructions.md (only receive)
- **Responsibilities**:
  - Read instructions from Solutions Architect
  - Implement major features and functionality
  - Write production-quality code
  - Assign minor tasks to Interns
  - Review intern work
- **Communication**:
  - RECEIVES instructions FROM Solutions Architect
  - GIVES instructions TO Interns

### Interns (Level 4 - BOTTOM)
- **Multiple allowed**: Can have many interns
- **Position**: Bottom of hierarchy, report to Engineers
- **Can write to**: shared/ only
- **Cannot write to**: Anyone's instructions.md (receive only)
- **Responsibilities**:
  - Read instructions from Software Engineers
  - Handle minor tasks (tests, linting, documentation)
  - Write unit tests for engineer code
  - Fix code style and linting issues
  - Learn and grow skills
- **Communication**:
  - RECEIVES instructions FROM Software Engineers
  - Does NOT give instructions to anyone

## Workflow Example

### Scenario: Building a REST API

1. **Manager writes project summary**
   ```
   engineering-manager/
   ├── tasks.md (initial task from system)
   ├── project-requirements.md (creates detailed requirements)
   └── writes to: solutions-architect/instructions.md
       "Design a REST API for user management with auth, CRUD, and audit logging..."
   ```

2. **Architect designs system**
   ```
   solutions-architect/
   ├── reads: instructions.md (from manager)
   ├── tasks.md (updates: "Design architecture" → in progress → completed)
   ├── system-design.md (creates architecture doc)
   ├── data-model.md (creates ER diagram and schema)
   ├── api-spec.md (creates API contracts)
   └── writes to:
       - software-engineer-1/instructions.md ("Implement auth module per attached spec...")
       - software-engineer-2/instructions.md ("Implement CRUD operations per data model...")
       - shared/architecture.md (shares with team)
   ```

3. **Engineers implement features**
   ```
   software-engineer-1/
   ├── reads: instructions.md (from architect)
   ├── tasks.md (updates: "not started" → "in progress")
   ├── auth.go (implements authentication)
   ├── tasks.md (updates: "in progress" → "completed")
   └── writes to: intern-1/instructions.md
       "Write unit tests for auth.go, ensure >80% coverage, fix any linting issues"

   software-engineer-2/
   ├── reads: instructions.md (from architect)
   ├── tasks.md (updates progress)
   ├── crud.go (implements CRUD operations)
   └── writes to: intern-2/instructions.md
       "Add integration tests for CRUD endpoints, check for error handling"
   ```

4. **Interns handle minor tasks**
   ```
   intern-1/
   ├── reads: instructions.md (from engineer-1)
   ├── tasks.md (updates: "not started" → "in progress")
   ├── auth_test.go (writes unit tests)
   ├── fixes linting issues in auth.go
   ├── tasks.md (updates: "in progress" → "completed")

   intern-2/
   ├── reads: instructions.md (from engineer-2)
   ├── tasks.md (updates progress)
   ├── crud_test.go (writes integration tests)
   ├── tasks.md (marks completed)
   ```

5. **Manager reviews final deliverables**
   ```
   engineering-manager/
   ├── reads: solutions-architect/system-design.md
   ├── reads: software-engineer-1/auth.go
   ├── reads: software-engineer-2/crud.go
   ├── reads: intern-1/auth_test.go
   ├── final-review.md (creates review document)
   └── If changes needed, writes to: solutions-architect/instructions.md
       "Adjust the auth flow to include MFA support..."
   ```

## Reading Other Personas' Work

Any persona can read any other persona's directory:

```go
// Read engineer's output
content := readFile(".database/software-engineer-1-*/auth.go")

// Check engineer's progress
tasks := readFile(".database/software-engineer-1-*/tasks.md")

// Read shared architecture
arch := readFile(".database/shared/architecture.md")
```

## File Naming Conventions

### tasks.md
- **One per persona**
- **Format**: Markdown with structured task entries
- **Updates**: Only by the owning persona
- **Status values**: "not started", "in progress", "completed"

### instructions.md
- **One per persona**
- **Format**: Markdown with timestamped sections
- **Appended to**: By other personas assigning work
- **Read by**: The owning persona

### Output Files
- **Naming**: Descriptive names (feature-name.go, api-docs.md, etc.)
- **Format**: Any format appropriate to the content
- **Location**: Persona's own directory
- **Readable by**: All personas

### Shared Files
- **Location**: .database/shared/
- **Purpose**: Resources needed by multiple personas
- **Examples**: architecture.md, requirements.md, common-code.go
- **Writable by**: Any persona

## Session Management

### Session Metadata (session.json)
```json
{
  "id": "software-engineer-1-1706012345678",
  "persona_type": "software-engineer",
  "persona_name": "engineer-1",
  "start_time": "2024-01-20T10:00:00Z",
  "status": "active",
  "workspace_id": "ws-1706012345",
  "pid": 12345
}
```

### Session Lifecycle
1. **Created**: Session directory and files initialized
2. **Active**: Claude running with persona instructions
3. **Completed**: Task finished successfully
4. **Failed**: Error occurred during execution
5. **Stopped**: Manually stopped by user

## Best Practices

### For Engineering Managers
- Provide clear, comprehensive project requirements
- Write detailed summaries that architect can design from
- Review all major deliverables before project completion
- Give instructions ONLY to Solutions Architect (not directly to engineers)
- Make final decisions on approach and priorities

### For Solutions Architects
- Carefully read Manager's instructions before designing
- Create clear architectural diagrams and data models
- Write detailed technical specifications for Engineers
- Share architecture documents in shared/ for team reference
- Give instructions ONLY to Software Engineers (not to interns)
- Ensure your design aligns with Manager's requirements

### For Software Engineers
- Follow Architect's technical specifications precisely
- Implement major features with production quality
- Keep tasks.md updated with current status
- Assign minor tasks (tests, linting) to Interns clearly
- Review intern work before marking tasks complete
- Share reusable code in shared/

### For Interns
- Carefully read Engineer's instructions before starting
- Focus on assigned minor tasks (tests, linting, docs)
- Ask questions if instructions are unclear (via your own output files)
- Write detailed comments explaining your work
- Update tasks.md frequently with progress
- Mark tasks completed only after thorough review

## Error Handling

### Session Failures
- Session status set to "failed"
- Other personas can continue
- Manager should reassign work

### Communication Issues
- Personas should check instructions.md regularly
- If no response, write follow-up in instructions.md
- Manager can intervene and reassign

### File Conflicts
- Each persona has own directory → no conflicts
- Shared/ should be used carefully
- Manager coordinates shared/ updates
