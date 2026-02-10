# Aiki Concurrency Design (Go-Lite)

Minimal threading with channels. 

## Primitives

```aiki
spawn(f)       # green thread, returns immediately
channel()      # unbuffered channel
send(ch, val)  # blocks until received
recv(ch)       # blocks until sent
```

No new syntax. Just builtins.

## Example

```aiki
let ch = channel()

spawn(() {
    send(ch, 42)
})

let val = recv(ch)
print(val)
```

## Known Limitations

- No protection against data races
- Shared mutable state is dangerous
- No preemption - long-running computations block
- Channel operations are the only yield points

Best practice: Don't share mutable state. Communicate through channels.

## Why Go-Lite

Go's model (goroutines + channels) is well-understood and maps cleanly to composition philosophy. Message passing through channels is encouraged. Shared mutable state is possible but discouraged.

## Implementation Notes

Needs:
- `value.Channel` type (wraps Go chan)
- Builtins in eval/builtins.go:
  - `spawn`: wrap function call in goroutine
  - `channel`: return new Channel value
  - `send`: send on channel, block
  - `recv`: receive from channel, block
- Consider: buffered channels `channel(n)`
- Consider: select/alt for multiple channels

## Deferred

- Process-level parallelism (process spawning, IPC, sockets)
