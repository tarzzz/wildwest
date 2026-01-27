# Persona Names System

## Overview

Instead of generic names like `engineer-1`, `intern-2`, personas are automatically assigned interesting names from famous historical figures across various fields.

## Name Categories (100 names)

### Scientists (20)
Einstein, Curie, Tesla, Newton, Darwin, Hawking, Feynman, Bohr, Galileo, Kepler, Faraday, Maxwell, Planck, Heisenberg, Schrödinger, Turing, Lovelace, Franklin, Nobel, Pasteur

### Artists (20)
Picasso, Monet, Da Vinci, Michelangelo, Rembrandt, Van Gogh, Dalí, Kahlo, Warhol, Banksy, Basquiat, Klimt, Matisse, Kandinsky, Pollock, Rothko, Hopper, O'Keeffe, Cassatt, Renoir

### Musicians (20)
Mozart, Beethoven, Bach, Chopin, Vivaldi, Tchaikovsky, Brahms, Debussy, Handel, Wagner, Hendrix, Lennon, Presley, Armstrong, Coltrane, Davis, Ellington, Parker, Mingus, Monk

### Writers (20)
Shakespeare, Hemingway, Tolkien, Austen, Dickens, Twain, Wilde, Orwell, Kafka, Poe, Woolf, Joyce, Nabokov, Dostoyevsky, Tolstoy, Márquez, Borges, Cervantes, Homer, Dante

### Philosophers (20)
Socrates, Plato, Aristotle, Kant, Nietzsche, Descartes, Locke, Hume, Spinoza, Wittgenstein, Russell, Sartre, Camus, Kierkegaard, Epicurus, Aquinas, Confucius, Laozi, Buddha, Seneca

### Inventors & Engineers (20)
Edison, Bell, Wright, Marconi, Diesel, Gutenberg, Watt, Ford, Nobel, Archimedes, Brunel, Hopper, Lovelace, Babbage, Jacquard, Stephenson, Musk, Jobs, Gates, Berners-Lee

### Explorers & Pioneers (10)
Magellan, Columbus, Armstrong, Shackleton, Hillary, Cousteau, Gagarin, Earhart, Polo, Livingstone

**Total: 130 names**

## Assignment Strategy

Names are assigned based on persona type:

### Engineering Manager → Philosophers
*Strategic thinkers and leaders*
- Socrates, Plato, Aristotle, Kant, Nietzsche, etc.

### Solutions Architect → Artists or Inventors
*Designers and builders*
- Picasso, Da Vinci, Edison, Tesla, etc.

### Software Engineers → Scientists or Inventors
*Problem solvers and creators*
- Einstein, Curie, Turing, Lovelace, etc.

### Interns → Writers or Explorers
*Learners and documenters*
- Hemingway, Tolkien, Magellan, Earhart, etc.

## Usage

### View All Names
```bash
wildwest names
```

Output:
```
Available Persona Names
═══════════════════════════════════════════════════

Total: 130 names across 7 categories

╔══════════════════════════════════════════════════╗
║  Scientists (20)
╚══════════════════════════════════════════════════╝
  einstein       curie          tesla          newton
  darwin         hawking        feynman        bohr
  ...
```

### Automatic Assignment

When creating a team:
```bash
$ wildwest team start "Build REST API"

Creating Engineering Manager directory (Level 1)...
  Name: socrates
  Directory: engineering-manager-1706012345678

Creating Solutions Architect directory (Level 2)...
  Name: davinci
  Directory: solutions-architect-1706012345679
```

### In Session Directories

Each persona directory shows their name:
```
.database/
├── engineering-manager-1706012345678/     # Socrates
│   ├── session.json                        # PersonaName: "socrates"
│   └── ...
├── solutions-architect-1706012345679/      # Da Vinci
└── software-engineer-1706012345680/        # Einstein
```

## Name Uniqueness

The system ensures no duplicate names within a workspace:
- Tracks used names in memory
- Loads existing session names on startup
- If all names exhausted, falls back to `persona-{timestamp}`

## Examples

### Manager Team Example
```
Engineering Manager: socrates
Solutions Architect: davinci
Software Engineers:
  - einstein
  - curie
  - turing
Interns:
  - hemingway
  - tolkien
```

### API Development Team
```
Manager: plato
Architect: tesla
Engineers:
  - newton (backend)
  - lovelace (frontend)
  - bohr (database)
Interns:
  - polo (testing)
  - earhart (docs)
```

### Microservices Migration Team
```
Manager: aristotle
Architect: edison
Engineers:
  - hawking (service-1)
  - feynman (service-2)
  - maxwell (service-3)
  - faraday (service-4)
Interns:
  - shakespeare (docs)
  - dickens (testing)
```

## Benefits

1. **Memorable**: Easy to remember "einstein" vs "engineer-3"
2. **Inspiring**: Famous names add character
3. **Organized**: Category-based assignment
4. **Unique**: No duplicates within workspace
5. **Scalable**: 130 names available
6. **Automatic**: No manual naming needed

## Attach to Named Sessions

```bash
# List with names
$ wildwest attach --list

╔══════════════════════════════════════════════════╗
║  SOFTWARE ENGINEERS
╚══════════════════════════════════════════════════╝

🔄 einstein
   Session ID: software-engineer-1706012345680
   Status: active

🔄 curie
   Session ID: software-engineer-1706012345681
   Status: active

# Attach by name
$ wildwest attach software-engineer-1706012345680

# Or remember: "Einstein is working on backend"
```

## Session Metadata

Each `session.json` contains:
```json
{
  "id": "software-engineer-1706012345680",
  "persona_type": "software-engineer",
  "persona_name": "einstein",
  "start_time": "2024-01-26T20:00:00Z",
  "status": "active"
}
```

## Dynamic Spawning with Names

When requesting new team members:
```bash
# Request in directory name (optional)
mkdir .database/software-engineer-request-backend-specialist

# Orchestrator spawns with auto-generated name
🚀 Spawning software-engineer: newton
   ✅ Session: software-engineer-1706012345682 (PID: 12345)
```

## Output Examples

### Team Start Output
```bash
$ wildwest team start "Build blog platform" --engineers 2

Created workspace: ws-1706012345
Workspace path: .database

Creating Engineering Manager directory (Level 1)...
  Name: kant
  Directory: engineering-manager-1706012345678

Creating Solutions Architect directory (Level 2)...
  Name: michelangelo
  Directory: solutions-architect-1706012345679

Creating 2 Software Engineer director(ies) (Level 3)...
  Name: tesla
  Directory: software-engineer-1706012345680
  Name: edison
  Directory: software-engineer-1706012345681
```

### Orchestrator Output
```bash
$ wildwest orchestrate

🎯 Project Manager Orchestrator Started
   Workspace: .database
   Poll Interval: 5s

🚀 Spawning engineering-manager: kant
   ✅ Session: engineering-manager-1706012345678 (PID: 12345)

🚀 Spawning solutions-architect: michelangelo
   ✅ Session: solutions-architect-1706012345679 (PID: 12346)

🚀 Spawning software-engineer: tesla
   ✅ Session: software-engineer-1706012345680 (PID: 12347)

🚀 Spawning software-engineer: edison
   ✅ Session: software-engineer-1706012345681 (PID: 12348)
```

### Track Output
```bash
$ wildwest track

═══════════════════════════════════════════════════
           PROJECT STATUS DASHBOARD
═══════════════════════════════════════════════════

╔══════════════════════════════════════════════════╗
║  ENGINEERING MANAGER
╚══════════════════════════════════════════════════╝

📋 kant (engineering-manager-1706012345678)
   Status: active
   Started: 2024-01-26 20:00:00

╔══════════════════════════════════════════════════╗
║  SOFTWARE ENGINEERS
╚══════════════════════════════════════════════════╝

🔄 tesla
   Session ID: software-engineer-1706012345680
   Status: active

🔄 edison
   Session ID: software-engineer-1706012345681
   Status: active
```

## Implementation

The name generator:
- Randomly selects from appropriate category
- Tracks used names
- Thread-safe with mutex
- Falls back gracefully if exhausted
- Loads existing names on startup

```go
// pkg/names/names.go
nameGen.GetNameForPersona("engineering-manager")  // Returns: "socrates"
nameGen.GetNameForPersona("software-engineer")    // Returns: "einstein"
nameGen.GetRandomName()                            // Returns: any available
```

## Name Pool Management

If you run out of names (>130 team members):
- Falls back to `persona-{timestamp}`
- Still maintains uniqueness
- Suggests expanding name pool

## Customization

To add more names, edit `pkg/names/names.go`:
```go
Scientists = []string{
    "einstein", "curie", "tesla",
    // Add your names here
    "your-scientist-name",
}
```

Rebuild:
```bash
make build
```
