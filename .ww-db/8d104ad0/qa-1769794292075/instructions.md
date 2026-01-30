## Instructions from engineering-manager (2026-01-30 12:32:11)
BUG FOUND: Orchestrator spawned duplicate QA agents from single request.
Investigate the spawning logic in pkg/orchestrator/orchestrator.go lines 170-220.
Check why processSpawnRequests() creates multiple agents for one request directory.
