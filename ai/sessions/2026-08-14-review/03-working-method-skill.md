# 03 — Working-method skill

Status: **GATED**

## Intent

Capture the `ai/` working method as a portable Claude skill so it triggers
on any Aiki-related conversation and produces maintainable session records
regardless of whether the session involves code changes.

## Implemented

A Claude skill (`aiki/SKILL.md`) encoding:

- project context (what Aiki is, key architectural concepts, project
  structure);
- the ten working rules from `ai/README.md`;
- session layout, status vocabulary, and resume procedure;
- explicit handling of non-coding sessions (discussion, review, planning);
- drop-in tarball delivery protocol for merging session records into the
  repository without conflict.

The skill triggers on tarball handoffs, Aiki-specific terminology, references
to the `ai/` directory, or any substantial Aiki conversation.

## Design decisions

- **Aiki-specific, not general.** The method without project context would
  produce empty ceremony on smaller tasks. The skill encodes Aiki's
  architecture, validation workflow, and the designer's isolation principle.
- **Non-coding sessions are first-class.** The skill specifies milestone
  format for discussions, reviews, and planning sessions. The test is
  restartability: could another AI answer what's true, why, and what's next?
- **Tarball delivery for session records.** The skill produces `ai/` tarballs
  rooted correctly for extraction into the repository. It never overwrites
  `ai/README.md` unless intentionally changed. New dated directories and
  updated session pointers merge without conflict.
- **Same-date sessions.** When a second session occurs on the same calendar
  date as an existing one, a descriptive suffix distinguishes them
  (e.g., `2026-08-14-review`).

## Validation

- Skill triggers correctly for the current session type (non-coding review).
- Session record produced for this conversation follows the skill's own
  specification.
- Tarball structure verified to merge cleanly with existing `ai/` content.

## Next action

Install the skill in the user's Claude profile. Use it for all subsequent
Aiki sessions.
