# HAL Capability Gates — Gate 1

## Purpose

Refine the HAL so different substrates may provide different host capabilities
without weakening Aiki's architecture. Aiki remains Aiki 1.

The existing names remain distinct:

- Aiki name — programmer meaning
- HAL name — architectural contract
- substrate name — concrete realization/provenance

A substrate need not implement every HAL operation. A runtime profile declares
which capabilities it requires; everything else is optional.

## Operations, capabilities, and profiles

HAL operations are primitive architectural contracts. Capabilities are
Aiki-visible names for one or more HAL operations.

A capability exists only when substrate support is materially required. Do not
create capabilities for convenience aliases, minor OS quirks, pure Aiki library
features, or functionality implementable entirely over existing HAL operations.

Profiles have only two states: required and optional. Required capabilities are
checked before execution; every capability not required by the selected profile
is optional.

## Capability registry

One authoritative registry maps Aiki capability names to HAL operation sets. It
drives profile validation and source-level `system.has(:capability)` /
`system.require(:capability)`.

The registry is metadata, not dispatch. Calls remain:

    Aiki -> HAL -> substrate

## Availability and authority

Availability and authority are independent:

- available + authorized -> proceed
- available + unauthorized -> authority failure
- unavailable -> unsupported capability

Diagnostics must not confuse unsupported capability with authority denial.

## Architectural metadata refactor

HAL architectural knowledge belongs under `engine/`, with separate concerns
joined by stable identities and validated as one coherent graph.

The pattern is:

> Centralize identity, decentralize concern, validate composition.

Operations, authority, capabilities, profiles, and substrate bindings/provenance
may live in separate files. The important property is a single coherent identity
graph, not a giant registry file.

This is a broader engine pattern: authoritative language/runtime knowledge lives
under `engine/`; separate artifacts join through stable names; invariants prove
the joins.

## Invariants

Mechanical checks should ensure:

- HAL identities are unique;
- authority policy names real runtime operations;
- capabilities name real HAL operations;
- profiles name real capabilities;
- substrate provenance names real HAL operations;
- `system.has` reflects actual substrate bindings;
- required capabilities are present before execution;
- unsupported and unauthorized failures remain distinct;
- capability metadata never participates in dispatch;
- ordinary Aiki source never depends on raw `HAL.*` identities.

## Aiki 1 compatibility

Aiki 1 is one evolving language contract. Additive change is preferred, but a
breaking refinement is permitted when it materially improves or corrects the
language or architecture. Arbitrary semantic drift is not.

Changing `+` to subtraction would corrupt an established contract. Changing
`and`/`or` to lazy logical control, or revising a poorly chosen operation such
as `is`, may break code while still legitimately refining Aiki 1.

## Phase 1 — common I/O

Provide common sequential I/O over standard endpoint symbols and file handles.
Initial surface:

- `io.read(source, count)`
- `io.read_line(source)`
- `io.write(destination, data)`
- `:stdin`, `:stdout`, `:stderr`

Standard endpoints are symbols, not a new public Stream type. File handles and
standard endpoints share the I/O behavior.

## Phase 2 — directories and symlinks

Add recursive directory walking and symbolic-link operations. Directory walking
belongs to the normal filesystem capability. Symlink support is a separate
optional capability because host support legitimately differs.

## Phase 3 — permissions and metadata

Aiki is Unix-friendly without pretending all permission models are Unix.

Layering remains thin:

- Aiki defines programmer-facing permission operations and errors;
- HAL defines the architectural contract;
- substrate maps that contract to the host.

Let the substrate normalize implementation differences, but do not let it hide
semantic differences that matter to Aiki. The Go substrate may use Go's
`FileMode`/`Chmod` vocabulary. A substrate that cannot faithfully provide the
Aiki permission contract reports the capability unavailable or the operation
unsupported rather than manufacturing false portability.

Ownership, ACLs, umask, and richer host permission models remain out of scope
until a concrete Aiki need justifies them.
