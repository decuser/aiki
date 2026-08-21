# Proposal: Engine Authority Centralization — Gate 2

## Purpose

Apply the architectural pattern established during HAL Gate 1 across the rest of
Aiki's engine-owned language and runtime authorities.

Gate 2 is behavior-preserving. It adds no language feature and intentionally
changes no Aiki semantics.

The governing rule is:

> Centralize identity, decentralize concern, validate composition.

A second rule follows from it:

> Authoritative knowledge about the language or runtime belongs under `engine/`
> unless a concrete architectural reason places the authority elsewhere.

Authored library source and user-facing documentation may remain beside the
libraries they describe. Gate 2 is about where architectural facts are derived,
owned, and joined, not about moving every artifact into one directory.

## Completion Criterion

Gate 2 is complete when no important language/runtime fact is independently
maintained in multiple places without a mechanically checked join, and consumers
no longer reconstruct facts for which an engine authority already exists.

## Constraints

Gate 2 must not:

- change grammar or language semantics;
- add prelude or library procedures;
- redesign the formatter;
- introduce another runtime dispatch layer;
- centralize concerns into one god registry;
- move authored library source merely for directory neatness;
- replace stable independent artifacts when a checked join is the better design;
- perform unrelated cleanup because nearby code is untidy.

## Phase A — Authority Inventory

Inventory language/runtime facts and record, for each:

- the authority;
- its current location;
- consumers;
- duplicated derivations or parallel inventories;
- existing invariant joins;
- whether movement, derivation, or no change is appropriate.

Initial areas:

- grammar and evaluator coverage;
- prelude surface and help/docs;
- shipped module/export/help/doc surface;
- language-service vocabulary;
- runtime module registry and resolution policy;
- other runtime registries or architectural inventories discovered during the
  inventory.

The inventory itself must distinguish authored artifacts from architectural
metadata. A `.ai`, `.help`, or `.doc` file can remain outside `engine/` while its
structural interpretation is owned by an engine package.

## Phase B — Runtime Registry Ownership

Move runtime registries and policies that currently live under a concrete
substrate package into an engine-owned runtime home when they are not actually
substrate-specific.

The first candidate is the named-module registry and module-root policy. It is
used by the runner, language services, runtime help, imports, and invariants, and
therefore must not be treated as Go-substrate authority merely because its first
implementation is in Go.

The substrate may still realize filesystem access. Runtime identity, resolution
policy, cache semantics, help/doc association, and canonical package naming are
engine concerns.

## Phase C — Shared Derived Views

Where consumers currently re-parse or reconstruct an authoritative fact, create
or reuse a derived engine view instead.

Examples include:

- module exported-procedure/signature analysis;
- prelude-visible vocabulary;
- runtime help/doc entries;
- language-service completion/hover vocabulary;
- grammar terminal/operator/keyword views.

A derived view must be computed from the owning authority. It must not become a
second independently maintained inventory.

## Phase D — Stable Joins

Make separate concerns meet through stable identities.

Important joins include:

- grammar node identity → evaluator handler;
- module/package identity → source → exports → help → docs;
- prelude procedure identity → implementation → help → docs;
- language-service name → actual engine/runtime authority;
- runtime registry package identity → import/help/language-service consumers.

Separate files are expected. Agreement is mechanical rather than positional.

## Phase E — Invariants

Add or relocate invariants so they validate the engine-owned joins instead of
reconstructing parallel authority in test code.

At minimum, preserve or strengthen checks for:

- grammar nodes without evaluator handlers and vice versa;
- prelude exports/help/docs disagreement;
- shipped module exports/help/docs disagreement;
- help templates disagreeing with actual procedure signatures;
- language services inventing runtime names;
- module registries producing identities inconsistent with module declarations;
- duplicated independently maintained authority lists.

Tests may live under `test/invariant`, but reusable extraction/analysis logic must
live with the engine authority it describes.

## Inventory Expectations

Some areas may already satisfy Gate 2 and should not be moved merely to create
work.

In particular, the grammar already owns reusable structural analysis and the
evaluator validates handler coverage against it. Gate 2 should record such areas
as already centralized rather than redesign them.

Likewise, language services are already engine-owned. The relevant question is
whether their catalog adapters derive vocabulary from runtime authorities or
reconstruct it independently.

## Validation

Each behavior-preserving cut should use focused Go tests where available.

The final Gate 2 requirement is:

`make validate`

Validation must demonstrate no observable Aiki behavior change.

## Expected Result

After Gate 2, Aiki's architecture should be easier to inspect because the
location of a fact predicts its authority:

- grammar facts derive from grammar;
- evaluator facts derive from evaluator;
- runtime/module facts derive from engine runtime authorities;
- language services consume those facts rather than recreate them;
- authored library and documentation artifacts remain independently readable but
  are mechanically joined to the engine's interpretation of them.

The repository remains Unix-like: many small artifacts, stable names, simple
interfaces, and strong composition checks rather than a monolithic registry.

## Implementation Status — 2026-08-17

Gate 2 implementation is complete pending the authoritative repository gate.

- Phase A inventory: complete. Grammar/evaluator authority was already
  centralized and was intentionally left unchanged.
- Phase B runtime registry ownership: complete. Named-module registry and root
  policy moved from `engine/runtime/hal/substrate` to `engine/runtime/modules`.
- Phase C shared derived views: complete. `modules.AnalyzeSource` now owns
  package/export/function-signature derivation; `prelude.LoadCatalog` joins
  prelude source/help/docs once; language analysis and invariant tests consume
  those views.
- Phase D stable joins: complete. Runtime primitive role classification now lives
  in `engine/runtime/primitives`; Go registration binds implementations by stable
  primitive identity and validates against the engine-owned role catalog.
- Phase E invariants: complete in code. Module source analysis, prelude catalog,
  primitive-role coverage, and substrate binding coverage have focused tests.

One pre-existing gap was discovered rather than folded into this refactor:
AF-017 records that `truncate` has prelude help but no full `.doc` entry. Gate 2
keeps the prior help-coverage contract and does not silently strengthen it.

Authoritative completion gate: `make validate` on the Go 1.24 development tree.
