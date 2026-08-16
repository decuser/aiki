# Milestone 01 — Thompson 7094 proposal baseline

Status: GATED

## Intent

Convert the earlier external 7094 cheat sheet and subsequent design discussion
into the authoritative implementation proposal for the reconstruction.

## Decisions

- Preserve the minimal 7094 criterion: implement only mechanisms required by
  Thompson's published object code and runtime.
- Preserve the 14-instruction initial inventory; do not add instructions
  speculatively.
- Distinguish IBM 7094 I/O channels, which remain out of scope, from Aiki
  channels, which may orchestrate host-side machine control later.
- Treat `store`, `bits`, `select`, spawn/channels, and the experiment framework
  as available substrate rather than unresolved prerequisites.
- The machine is the semantic owner of memory and register state. Explicitly
  shareable Aiki stores are not permission for a later monitor to mutate the
  machine behind its command boundary.
- Organize the reconstruction as three phases of three cuts: machine, published
  runtime/object result, compiler/end-to-end reproduction.
- Validate the published `a(b|c)*d` object program before implementing the
  compiler so machine correctness is separated from compiler correctness.

## Result

`proposals/thompson-7094-regex.md` is the authoritative implementation baseline.
