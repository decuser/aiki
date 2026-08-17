# Proposal: Systems Programmer Convenience Refinements

## Purpose

Close the small remaining gaps that repeatedly force systems-oriented Aiki programs to reimplement common algorithms or resource-management patterns.

This proposal adds no new HAL capability class. It refines three systems-programmer surfaces: scalar ordering and list sorting, numeric base conversion, and whole-file convenience I/O.

The governing rule is:

> Add an operation when real programs repeatedly reimplement a nontrivial algorithm or resource-management pattern. Do not add operations merely because other languages have them.

## Phase 1 — List Ordering

Add one operation with two forms:

`list.sort(xs)`

`list.sort(xs, fn)`

Both forms land together. Sorting returns a new list and leaves the input unchanged.

The default form uses Aiki's natural scalar ordering. The language defines one shared ordering contract for numbers, strings, runes, and symbols: numbers compare numerically, strings compare lexicographically by rune sequence, runes compare by code point, and symbols compare by name. `<`, `>`, `<=`, `>=`, and `list.sort(xs)` all use that same relation. Mixed types and composite values remain unordered.

The function form supplies a strict ordering relation `fn(a, b)` returning a boolean and is the general form for shapes, composites, descending order, and other programmer-defined orderings.

Sorting remains pure Aiki, deterministic, stable where equal elements are concerned, and requires no HAL support. The initial implementation is a stable non-mutating merge sort isolated behind a private implementation seam so a later native/FFI implementation can replace it without changing the public contract if profiling justifies that work.

## Phase 2 — Number Base Conversion

Create pure `lib/number` rather than expanding the prelude.

Surface:

`number.to_base(n, base)`

`number.from_base(text, base)`

Rules:

- bases 2 through 36;
- optional leading sign when parsing;
- formatting accepts exact integer-valued Aiki numbers only;
- parsing returns exact Aiki numbers;
- lowercase alphabetic digits for formatting;
- invalid bases, digits, or values return ordinary shaped Aiki errors;
- no floating-point path;
- no HAL primitive.

This keeps the prelude restricted to its existing evaluator-privileged and irreducible-wrapper role while making numeric representation available through an explicit library.

## Phase 3 — Whole-File I/O

Add whole-file convenience on top of the existing handle-oriented `file` API.

Text:

`file.read_all(path)`

`file.write_all(path, text)`

Bytes:

`file.read_all_bytes(path)`

`file.write_all_bytes(path, bytes)`

The existing rule remains: strings are text; bytes are bytes. Text behavior follows `file.read_text`; no new encoding machinery is introduced.

Write operations create or replace/truncate the destination. The helpers own open/read-or-write/close choreography and return normal Aiki errors. They compose existing file operations and add no HAL contract.

## Explicitly Not Included

This proposal does not add list positional search/slicing, scalar min/max, networking, JSON, signals, encoding selection, strict UTF-8 validation, or new HAL capabilities.

## Validation

Every phase lands with implementation, help, full docs, executable examples, behavioral tests, and relevant invariant coverage. `make invariant` remains green and fast. `make validate` gates the completed project.

## Completion

Complete when Aiki can sort naturally or under a supplied ordering relation, convert exact integers to and from bases 2–36 without hand-rolled code, and perform whole-file text/byte I/O without manual handle choreography, while preserving current architectural invariants.
