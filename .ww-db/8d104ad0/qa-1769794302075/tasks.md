# Tasks

## Task: Test Orchestrator System
Status: in progress

### Test Cases:

1. **Agent Spawning from Request Directories** - not started
   - Create a test request directory
   - Verify agent is spawned automatically
   - Verify directory is renamed with timestamp

2. **Instruction Monitoring Responsiveness** - not started
   - Write to another agent's instructions.md
   - Verify notification within 5 seconds
   - Test background monitoring task

3. **Inter-Agent Communication** - not started
   - Test communication between different agent types
   - Verify message delivery and reading
   - Test bi-directional communication

4. **Task Completion Detection** - not started
   - Create agent with completable tasks
   - Mark all tasks as completed
   - Verify automatic termination and archival

5. **Tmux Session Lifecycle** - not started
   - Verify tmux session creation
   - Test session persistence
   - Verify cleanup on completion
