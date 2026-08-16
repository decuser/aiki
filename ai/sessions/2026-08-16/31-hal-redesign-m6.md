# HAL redesign — M6 Canvas pressure migration

Status: **GATED**

Baseline entering M6: M5 passed `go test ./...` on the user's authoritative tree.

## M6.a — narrow Canvas host boundary

Introduced one Aiki-defined Canvas command protocol crossing:

```text
Aiki canvas/turtle operation
    -> _canvas_command(canvas, :operation, args)
    -> HAL.canvas.command
    -> Go Canvas command / IPC realization
```

Canonical Canvas host contracts are now:

- `HAL.canvas.open`
- `HAL.canvas.command`
- `HAL.canvas.close`
- `HAL.canvas.width`
- `HAL.canvas.height`
- `HAL.canvas.alive`

The old per-command `_dot`, `_line`, `_rect`, `_fill_rect`, `_circle`,
`_fill_circle`, `_arc`, `_clear`, `_set_bg`, `_set_fg`, `_pen_size`, and
`_set_turtle` primitives remain registered only as migration compatibility
aliases. M7, not M6, removes compatibility names after the replacement path is
validated.

`lib/canvas/canvas.ai` now constructs the domain protocol in Aiki. Public
programmer-facing Canvas operations remain unchanged.

Registry accounting at this point:

```text
compatibility primitives: 132
canonical host contracts: 38
```

## M6.b — turtle migration

Both turtle implementations now use `_canvas_command` for drawing/state
commands. No production Aiki Canvas/turtle source uses the old per-command raw
Canvas primitives. Resource acquisition/query/lifetime crossings remain narrow
and explicit.

Trusted-source authority policy was reduced accordingly: Canvas/turtle modules
receive the generic command grant rather than the per-drawing-command grant
set. Selfhost bootstrap retains the old aliases temporarily and captures the
new command primitive as part of compatibility migration.

## Evidence available in this environment

- `gofmt` clean on changed Go files.
- `git diff --check` clean.
- source scan finds no old per-command raw Canvas primitive use in production
  Canvas/turtle Aiki sources.
- new `_canvas_command` use is present in Canvas and both turtle libraries.

The local tool environment cannot execute the repository's Go 1.24 test suite.
User should run `go test ./...` before M6.c because the next slice changes the
semantic Canvas value representation and this is the useful localization gate.

## Next

M6.c: remove concrete Go Canvas/session representation from the semantic value
layer by replacing it with an opaque runtime-owned resource reference while
preserving public Canvas behavior and the generic protocol path.

## M6.c — opaque Canvas resource representation

- Replaced the semantic `value.Canvas` implementation with an opaque runtime resource handle (`ID` only).
- Moved dimensions, color/pen state, drawing command queues, lifecycle channels, and turtle overlay state into substrate-owned `CanvasResource`.
- Moved `CanvasCmd` out of the semantic value package into the substrate.
- `GoRuntime` now owns the handle -> `CanvasResource` mapping; a handle from another runtime is not an active resource there.
- Removed the package-global Canvas session/bridge registries. The runtime-owned `CanvasResource` owns its child session and bridge lifecycle.
- The dedicated Canvas child/Ebiten renderer operates directly on substrate `CanvasResource`, not on an Aiki semantic value.
- Added executable invariant coverage that the semantic Canvas value cannot regain color/channel/mutex/protocol state.
- Public Canvas/turtle behavior and the M6.a generic `_canvas_command` protocol are unchanged by this representation cut.

Validation in this environment: `gofmt` and `git diff --check` clean; source/reference sweeps clean. Full Go execution remains unavailable locally because the provided environment has Go 1.23.2 while the repository requires Go 1.24. The user reported the M6.c `go test ./...` checkpoint passed on the authoritative tree. M6 is GATED.
