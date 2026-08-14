# Milestone 16 - working method refinement and final delivery

Status: **GATED**

## Intent

Adopt the refined Aiki working method jointly developed by Will and Claude as the durable repository method, then prepare final delivery artifacts from the current authoritative tree.

## Changes

- Replaced `ai/README.md` with the refined working method supplied by Will.
- Added explicit rules for substantial non-coding Aiki sessions.
- Added the required drop-in `ai/` session delivery convention.
- Preserved the single-tree, serial-cut, evidence-gated, restartable method.

## Decision

The uploaded working method is authoritative for future Aiki sessions. In particular:

- substantial design/review/planning work is recorded even when no code changes;
- the dated session index remains the mutable restart point;
- numbered milestones remain the durable engineering record;
- `summary.md` remains the curated conceptual narrative;
- when a session produces `ai/` content, a drop-in tarball rooted at `ai/` is delivered in addition to any requested source artifact.

## Validation

- Confirmed the updated `ai/README.md` is present in the authoritative tree.
- Confirmed the current session remains `2026-08-14` and already contains milestones 01-15.
- Packaging inspection is performed as part of final delivery.

## Caveat

The container remains offline and cannot perform the real Ebiten-dependent executable-doc validation. The current source validation state is recorded in milestone 12 and later milestones; this governance update changes no runtime code.

## Next action

Package the current authoritative source tree and the drop-in `ai/` session archive, inspect both archives, and deliver them with checksums.
