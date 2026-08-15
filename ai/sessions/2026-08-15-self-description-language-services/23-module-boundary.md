# Milestone 23 — Module-value boundary

Status: SUPERSEDED by Milestone 24.

## Original finding

Delegated `import()` returned an opaque native `:module`, while ordinary Aiki had no runtime-name export accessor. Standard Aiki modules also depend on blessed HAL-facing source modules.

## Resolution

The project chose not to add module reflection. Cut III.4 now self-hosts all Aiki-source module loading. A privileged bootstrap privately supplies HAL function values to interpreted blessed-library environments without exposing raw HAL bindings to callers. See Milestone 24.
