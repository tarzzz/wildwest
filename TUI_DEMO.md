# Static Org Chart TUI Demo

## Overview
A basic, static TUI showing the WildWest team organization chart with interactive navigation.

## Features

### 1. **Static Org Chart Display**
   - Engineering Manager (Level 1)
   - Solutions Architect & QA Engineer (Level 2)
   - Software Engineer 1 & 2 (Level 3)
   - Intern 1 & 2 (Level 4)

### 2. **Navigation**
   - Arrow keys (↑↓←→) or Vim keys (hjkl) to move between components
   - Selected component is highlighted with a thick pink border

### 3. **Component Details**
   - Press **Enter** or **Space** to view detailed information about the selected component
   - Shows:
     - Name and emoji
     - Role
     - Description
     - Component ID
   - Press **Enter** or **Space** again to close details

### 4. **Visual Styling**
   - Each component shown as a rounded box
   - Visual connectors showing hierarchy
   - Color-coded borders:
     - Blue (63): Default components
     - Bright pink (205): Selected component
     - Cyan (86): Details panel

## Running the TUI

```bash
# Build the project
go build -o wildwest

# Run the static TUI
./wildwest tui
```

## Controls

- `↑` `↓` `←` `→` or `h` `j` `k` `l` - Navigate between components
- `Enter` or `Space` - Toggle component details
- `q` or `Ctrl+C` - Quit

## Architecture

### Files
- `pkg/orchestrator/tui.go` - Complete TUI implementation
- `cmd/tui.go` - Command handler

### Key Components
- `Component` struct - Represents each team member
- `OrgChartModel` - Bubble Tea model with static data
- `renderOrgChart()` - Renders the hierarchical layout
- `renderDetails()` - Shows detailed component info

## Next Steps

This is a foundation for:
1. Hooking up real orchestrator data
2. Making components clickable to spawn/attach sessions
3. Showing real-time status updates
4. Adding interactive actions (spawn, stop, attach)

For now, it demonstrates:
- Clean org chart layout
- Keyboard navigation
- Component selection
- Information display
