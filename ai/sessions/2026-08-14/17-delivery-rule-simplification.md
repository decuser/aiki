# Milestone 17 - Delivery rule simplification

Status: GATED

## Intent

Align the durable Aiki working method with the clarified delivery practice: when repository content changes, the updated `ai/` record travels inside the single full-repository archive rather than as a redundant second tarball.

## Decision

Delivery is now scope-sensitive:

- If code, tests, docs, proposals, library modules, or other repository content changed, deliver one full repository tarball containing `ai/`.
- If only `ai/` session records changed, deliver a small `ai-session.tgz` rooted at `ai/`.
- Do not produce both forms by default.

## Rationale

The full repository archive already contains the authoritative session record. A second `ai/` archive adds duplication and invites ambiguity about which artifact is authoritative. The drop-in `ai-session.tgz` remains useful for discussion/review sessions that change only provenance records.

## Validation

- Updated the `Delivering session records` section in `ai/README.md`.
- Confirmed the rule distinguishes repository-changing and session-record-only work.
- No implementation or language behavior changed.

## Next action

Package the current complete Aiki tree once, including this updated working method and session record, and deliver that archive as the sole final artifact for this session.
