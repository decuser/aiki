# Aiki Alpha 2

Alpha 2 is the second public milestone of Aiki. The language surface remains deliberately small, but the system around it has changed substantially since Alpha 1. Grammar authority, self-description, tooling, distribution, systems facilities, host-boundary enforcement, library realization, experiments, and validation are all materially stronger.

This note summarizes the Alpha 1 to Alpha 2 changes as a release milestone rather than as a commit-by-commit changelog.

## Language authority and syntax

The grammar is now the sole authority for Aiki syntax. Derived grammar analysis is centralized and reused by the engine rather than being reconstructed independently by multiple subsystems. Exceptional syntax paths were hardened against drift, and newline termination became an explicit language rule with structural and behavioral coverage.

Formatter, linter, parser, evaluator, module loading, engine-smoke coverage, and related tooling are now tied more directly to the same grammar authority. This reduces the chance that a locally reasonable implementation change quietly defines a different language.

## Self-description and language services

Aiki now contains substantial implementations of its own language machinery in Aiki itself, including lexer, parser, normalization, evaluator/bootstrap, token authority, and grammar authority work. Conformance checks exercise these paths against the reference implementation.

The practical result is that self-description is no longer only an architectural aspiration. The experiment and conformance suites exercise nested interpretation and self-hosted language behavior as working parts of the system.

Language tooling also grew considerably. Alpha 2 includes language-server support, formatting services, completion and hover support, tags, Xed integration, and a VS Code extension. These services use the same language structures as the interpreter rather than forming an unrelated editor-side model of Aiki.

## Distribution and relocation

Alpha 2 can be built and tested as a relocatable distribution. The executable resolves its shipped library relative to the installed release rather than depending on the source working directory. Distribution checks exercise the built artifact from temporary locations and verify that package discovery does not accidentally depend on the development tree.

The repository also has a baseline mechanism for portable development snapshots and stronger separation between source-tree development artifacts and user distributions.

## Systems programming and HAL

The host abstraction layer was redesigned and then tightened through explicit authority and capability metadata. The boundary now distinguishes programmer-facing Aiki meaning, HAL architectural contracts, and substrate implementation/provenance.

Trusted sources receive explicit authority for the runtime or provider primitives they use, and invariants check both directions: undeclared dependencies are rejected, and stale grants are rejected. Host-facing facilities therefore live behind an enforced boundary rather than a naming convention.

The systems surface expanded with modules and operations for files, paths, processes, signals, terminals, networking, system information, time, I/O, byte-oriented work, and related conveniences. These facilities are exposed as Aiki library capabilities while their irreducible host interactions remain governed by HAL.

## Native, FFI, capability, and interop

Alpha 2 makes library realization an explicit part of the architecture.

Portable semantic capabilities must have a genuine Aiki-native path. When a native realization exists, the bare package name resolves to it. Provider-backed `/ffi` modules are explicit choices and may be mixed freely with native modules.

The separation is enforced in both directions:

- native modules may not acquire FFI transitively;
- portable FFI realizations may not silently fall back to their native semantic authority;
- exported portable FFI affordances must actually reach provider primitives;
- FFI-only packages do not become accidental bare defaults.

The current portable realization pairs include bits, bytes, hash, and string. Mathematics deliberately distinguishes native exact/rational algorithms from provider-backed inexact interoperability rather than pretending that the two surfaces are identical. `regex/ffi` likewise remains explicit provider interoperability.

Capabilities such as canvas and store are not mislabeled as FFI merely because they ultimately require runtime support. Higher-level Aiki libraries may remain native while terminating in a legitimate capability boundary; turtle, for example, uses native mathematics over the canvas capability.

The command:

```sh
aiki check --ffi-use program.ai
```

reports direct and transitive FFI imports without changing normal import mechanics or imposing an additional declaration system.

## Library growth and cleanup

The supplied library expanded substantially since Alpha 1. New module families include I/O, networking, numbers, paths, processes, self-hosting, signals, and terminal facilities, while existing modules gained additional behavior and clearer contracts.

The same pass removed or relocated affordances whose placement did not reflect their semantics. In particular, Store is again the explicit mutable indexed-storage capability rather than a home for workload-shaped conversion or checksum helpers.

## Experiments as executable witnesses

Alpha 2 adds a structured experiment layer rather than treating large examples as informal demonstrations.

Experiment 001 calibrates semantic profiling and exercises one- and two-level self-hosted interpretation. Experiment 002 reconstructs Ken Thompson's 1968 regular-expression compiler/runtime path through an IBM 7094 emulator, including monitor behavior and the published object-program geometry. Experiment 003 exercises concurrent systems programming through Four-Way Life, including processes, channels, byte boundaries, storage, and explicit native/FFI choices.

These experiments are useful both as demonstrations and as regression witnesses for architectural changes.

## Validation and architectural invariants

Validation is substantially stricter in Alpha 2. In addition to ordinary Aiki tests, the tree now uses behavioral smoke tests, engine-structure coverage, conformance tests, self-host checks, architectural invariants, boundary tests, property tests, fuzzing, distribution checks, tree ownership checks, formatter/AST preservation checks, and gold-backed engine tests.

The `rigorous` gate composes these checks into a release-strength validation path. The native/FFI redesign was additionally exercised against pre-existing behavioral witnesses; those witnesses exposed implementation defects in the new native bit operations and turtle's cardinal movement, and the implementation was corrected without changing the expected behavior.

## What remains recognizably Aiki

Alpha 2 is a substantial implementation and architecture milestone, not a replacement language. The central programming model remains the same: exact rational arithmetic by default, left-to-right binary evaluation, explicit grouping, first-class functions, lists and shapes, matching, pipelines, recoverable errors as values, a small prelude, explicit modules, and isolated message-passing concurrency.

Performance work remains secondary to correctness, architectural sanity, and inspectability. Alpha 2 should be read as a more disciplined and more self-describing implementation of the same small-language project rather than as a move toward a larger language surface.

## Documentation

The repository documentation, help, module documentation, and *This Is Aiki* are being reconciled to the Alpha 2 architecture. The Report on the Programming Language Aiki remains the Alpha 1 language report; a second report can describe the Alpha 2 architecture and implementation more precisely without making that report necessary for ordinary use of the repository.
