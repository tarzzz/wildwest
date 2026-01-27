# Claude Wrapper Demo

## Quick Demo: Named Persona Team

### Step 1: View Available Names
```bash
$ ./bin/wildwest names

Available Persona Names
═══════════════════════════════════════════════════

Total: 130 names across 7 categories

╔══════════════════════════════════════════════════╗
║  Scientists (20)
╚══════════════════════════════════════════════════╝
  einstein       curie          tesla          newton
  ...

╔══════════════════════════════════════════════════╗
║  Philosophers (20)
╚══════════════════════════════════════════════════╝
  socrates       plato          aristotle      kant
  ...
```

### Step 2: Start a Team

```bash
$ ./bin/wildwest team start "Build a blog platform with posts and comments" --engineers 2

Created workspace: ws-1706012345
Workspace path: .database

Creating Engineering Manager directory (Level 1)...
  Name: kant
  Directory: engineering-manager-1706012345678

Creating Solutions Architect directory (Level 2)...
  Name: davinci
  Directory: solutions-architect-1706012345679

Creating 2 Software Engineer director(ies) (Level 3)...
  Name: einstein
  Directory: software-engineer-1706012345680
  Name: tesla
  Directory: software-engineer-1706012345681

✅ Team structure created successfully!
📁 Workspace: .database

⚠️  IMPORTANT: Start the orchestrator to spawn Claude instances:
   wildwest orchestrate --workspace .database
```

### Step 3: Start Orchestrator

```bash
# Terminal 2
$ ./bin/wildwest orchestrate --workspace .database

🎯 Project Manager Orchestrator Started
   Workspace: .database
   Poll Interval: 5s

🚀 Spawning engineering-manager: kant
   ✅ Session: engineering-manager-1706012345678 (PID: 12345)

🚀 Spawning solutions-architect: davinci
   ✅ Session: solutions-architect-1706012345679 (PID: 12346)

🚀 Spawning software-engineer: einstein
   ✅ Session: software-engineer-1706012345680 (PID: 12347)

🚀 Spawning software-engineer: tesla
   ✅ Session: software-engineer-1706012345681 (PID: 12348)
```

### Step 4: Monitor Team

```bash
# Terminal 3
$ ./bin/wildwest attach --list

Available Sessions:
═══════════════════════════════════════════════════

╔══════════════════════════════════════════════════╗
║  ENGINEERING MANAGER
╚══════════════════════════════════════════════════╝

🔄 kant
   Session ID: engineering-manager-1706012345678
   Status: active
   Started: 2024-01-26 20:00:00
   PID: 12345

╔══════════════════════════════════════════════════╗
║  SOLUTIONS ARCHITECT
╚══════════════════════════════════════════════════╝

🔄 davinci
   Session ID: solutions-architect-1706012345679
   Status: active
   Started: 2024-01-26 20:00:05
   PID: 12346

╔══════════════════════════════════════════════════╗
║  SOFTWARE ENGINEERS
╚══════════════════════════════════════════════════╝

🔄 einstein
   Session ID: software-engineer-1706012345680
   Status: active
   Started: 2024-01-26 20:00:10
   PID: 12347

🔄 tesla
   Session ID: software-engineer-1706012345681
   Status: active
   Started: 2024-01-26 20:00:15
   PID: 12348
```

### Step 5: Attach to Kant (Manager)

```bash
$ ./bin/wildwest attach

🔗 Attaching to session: engineering-manager-1706012345678
   Directory: .database/engineering-manager-1706012345678

You are now in the session's directory.

# Inside Kant's workspace
$ ls
tasks.md  instructions.md  tracker.json  session.json  project-requirements.md

$ cat tasks.md
# Tasks

## Task: Build a blog platform with posts and comments
- **Status**: in progress
- **Assigned by**: system
- **Created**: 2024-01-26 20:00:00

$ cat session.json
{
  "id": "engineering-manager-1706012345678",
  "persona_type": "engineering-manager",
  "persona_name": "kant",
  "start_time": "2024-01-26T20:00:00Z",
  "status": "active",
  "pid": 12345
}

$ exit  # Back to main shell
```

### Step 6: Request More Team Members

Kant (Manager) decides to request another engineer:

```bash
$ ./bin/wildwest attach  # Attach to Kant

# Inside Kant's session
$ mkdir ../software-engineer-request-frontend-specialist

$ cat > ../software-engineer-request-frontend-specialist/instructions.md <<EOF
## Instructions from engineering-manager-1706012345678 (2024-01-26 20:30:00)

Implement the React frontend for the blog platform.

Requirements:
- Use React with TypeScript
- Implement components for blog post listing and viewing
- Add comment section with real-time updates
- Follow the API spec in shared/api-spec.md
- Write comprehensive tests

Deliverables:
- src/components/BlogPost.tsx
- src/components/CommentSection.tsx
- src/App.tsx
- tests/
EOF

$ exit
```

Orchestrator detects and spawns:
```
🚀 Spawning software-engineer: curie
   ✅ Session: software-engineer-1706012345682 (PID: 12349)
```

### Step 7: Track Progress

```bash
$ ./bin/wildwest track

═══════════════════════════════════════════════════
           PROJECT STATUS DASHBOARD
═══════════════════════════════════════════════════

╔══════════════════════════════════════════════════╗
║  ENGINEERING MANAGER
╚══════════════════════════════════════════════════╝

📋 kant (engineering-manager-1706012345678)
   Status: active
   Started: 2024-01-26 20:00:00

   Current Tasks:
      🔄 Build a blog platform with posts and comments [in progress]

   Latest Instructions:
      ## Instructions from system

╔══════════════════════════════════════════════════╗
║  SOLUTIONS ARCHITECT
╚══════════════════════════════════════════════════╝

📋 davinci (solutions-architect-1706012345679)
   Status: active
   Started: 2024-01-26 20:00:05

   Current Tasks:
      ✅ Design system architecture [completed]
      🔄 Create API documentation [in progress]

   Output Files:
      • system-design.md
      • data-model.md
      • api-spec.md

╔══════════════════════════════════════════════════╗
║  SOFTWARE ENGINEERS
╚══════════════════════════════════════════════════╝

📋 einstein (software-engineer-1706012345680)
   Status: active

   Current Tasks:
      ✅ Implement backend API [completed]
      🔄 Add authentication [in progress]

   Output Files:
      • blog_handler.go
      • post_service.go
      • comment_service.go

📋 tesla (software-engineer-1706012345681)
   Status: active

   Current Tasks:
      🔄 Implement database layer [in progress]

   Output Files:
      • models.go
      • db.go

📋 curie (software-engineer-1706012345682)
   Status: active

   Current Tasks:
      ⏸️  Implement React frontend [not started]

═══════════════════════════════════════════════════
                OVERALL SUMMARY
═══════════════════════════════════════════════════

Total Team Members: 4
Total Tasks: 7
Completed: 2
In Progress: 4
Not Started: 1

Overall Completion: 28.6%
[███████████░░░░░░░░░░░░░░░░░░░░░░░░░░░] 28%
```

### Step 8: Engineer Requests Intern

Einstein needs help with tests:

```bash
$ ./bin/wildwest attach software-engineer-1706012345680

# Inside Einstein's session
$ mkdir ../intern-request-backend-tester

$ cat > ../intern-request-backend-tester/instructions.md <<EOF
## Instructions from software-engineer-1706012345680 (2024-01-26 21:00:00)

Write comprehensive unit tests for the blog backend.

Tasks:
- Write tests for blog_handler.go with >80% coverage
- Write tests for post_service.go
- Write tests for comment_service.go
- Fix any linting issues
- Run: go test -cover ./...

Deliverables:
- blog_handler_test.go
- post_service_test.go
- comment_service_test.go
EOF

$ exit
```

Orchestrator spawns:
```
🚀 Spawning intern: hemingway
   ✅ Session: intern-1706012345683 (PID: 12350)
```

### Step 9: Completion and Cleanup

When Einstein marks all tasks complete:

```bash
# Einstein's tasks.md
## Task: Implement backend API
- **Status**: completed

## Task: Add authentication
- **Status**: completed
```

Orchestrator detects:
```
🎉 All tasks completed for einstein (software-engineer-1706012345680)
   📦 Archived to: software-engineer-1706012345680-completed
```

## Complete Team Example

```
Manager: kant (philosopher)
Architect: davinci (artist)
Engineers:
  - einstein (scientist) - Backend
  - tesla (inventor) - Database
  - curie (scientist) - Frontend
Interns:
  - hemingway (writer) - Testing
  - tolkien (writer) - Documentation
```

## Real-World Scenario: E-commerce Platform

```bash
# Start team
wildwest team start "Build complete e-commerce platform" --engineers 4

Team Created:
- Manager: plato
- Architect: michelangelo
- Engineers: newton, bohr, feynman, lovelace

# Plato requests more engineers
mkdir software-engineer-request-payment-specialist
mkdir software-engineer-request-search-specialist

# Orchestrator spawns:
- curie (payment)
- hawking (search)

# Newton requests interns
mkdir intern-request-payment-tester
mkdir intern-request-api-docs

# Orchestrator spawns:
- shakespeare (documentation)
- dickens (testing)

Final Team:
- Manager: plato
- Architect: michelangelo
- Engineers: newton, bohr, feynman, lovelace, curie, hawking
- Interns: shakespeare, dickens
```

## Benefits

1. **Memorable Names**: "Ask Einstein about the API" vs "Check engineer-3"
2. **Personality**: Names add character to the team
3. **Easy Tracking**: "Hemingway finished the tests"
4. **Professional**: Historical figures inspire quality work
5. **Automatic**: No manual naming required
6. **Unique**: No duplicates within workspace
7. **Scalable**: 130 names available

## Commands Summary

```bash
# View names
wildwest names

# Start team (auto-named)
wildwest team start "task" --engineers N --interns N

# Orchestrate
wildwest orchestrate --workspace .database

# List sessions (with names)
wildwest attach --list

# Attach by default (manager)
wildwest attach

# Track progress (with names)
wildwest track
```

## Name Categories by Role

**Managers** (Philosophers): Strategic thinking, wisdom
- kant, plato, aristotle, socrates, nietzsche

**Architects** (Artists/Inventors): Design, creativity
- davinci, michelangelo, picasso, edison, tesla

**Engineers** (Scientists/Inventors): Problem solving, innovation
- einstein, curie, newton, turing, lovelace

**Interns** (Writers/Explorers): Documentation, learning
- hemingway, shakespeare, tolkien, magellan, earhart
