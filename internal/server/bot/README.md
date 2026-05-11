# Bot module boundary

`internal/server/bot` owns bot-only behavior. The surrounding `server.Room`
keeps responsibility for room state, actor serialization, engine locking, and
message delivery.

## Files

- `memory.go`: public reveal memory used for lightweight hand inference.
- `runtime.go`: prompt-actionability checks that do not require a `Room`.
- `decision.go`: pure-ish bot decision policy from state, prompt, skills, and
  memory to `model.PlayerAction`.

## Dependency direction

- `server` may call `bot`.
- `bot` must not import `internal/server`.
- `bot` may use `model` and `viewmodel` DTOs, but must not access `Room`,
  `Client`, websocket connections, timers, or actor queues.

This keeps bot strategy testable without room transport and lets later room
refactors focus on lifecycle, protocol, and frontend interaction boundaries.
