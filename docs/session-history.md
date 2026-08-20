# Aiki Session History

Consolidated decisions, findings, and architectural outcomes from the AI
working record. Individual session provenance files have been removed from
the repository; this document preserves the durable findings.

Sessions covered: 2026-08-14 through 2026-08-16.

---

## Profiling and computational visibility (2026-08-14)

- Two-view profiling architecture: deterministic Aiki semantic counts (what
  computation occurred) plus sampled Go substrate costs (what the host did to
  realize it). The two are correlated at the HAL boundary via pprof labels but
  never collapsed into a synthetic cost.
- Attribution can distort the computation it observes. Source-line caching in
  environments was required to prevent the profiler from dominating its own
  allocation measurements.
- Profiling exposed a concurrency correctness bug: spawned execution shared
  mutable call-stack state with the parent through environment chains. Fixed by
  splitting `NewCallEnv` (lexical bindings from definition, dynamic state from
  caller) and `NewIsolatedEnclosedEnv` for spawn.
- Channel send accounting requires a happens-before rule: send is counted before
  the handoff so a receiver's profile view is consistent.
- Go allocation profiles do not retain Aiki pprof labels. Allocation support is
  deliberately split between measured interval totals and Go hotspots.
- `Store` is an explicit exception to spawn isolation (sync.RWMutex, shareable
  via closure capture). Documented, not a defect.

## Code review and designer isolation (2026-08-14)

- Designer isolation is constitutive, not a deficit. Not reading the Go
  implementation mirrors the language's own commitments (exact rationals
  eliminate hidden numeric state; isolated spawn eliminates hidden coupling).
  The designer's confusion is diagnostic signal, not a knowledge gap.
- The executable documentation conformance suite is what makes isolation
  rigorous rather than aspirational.
- `value` importing `engine` for profiling types creates a dependency direction
  the designer is tracking.

## Paper review: "Beyond Effective Use" MISQ rejection (2026-08-14)

- EIC desk rejection contained a categorical mismatch: assigned Theory
  Development, prescribed Theory-Generative Research Synthesis method.
- One substantive finding survives: Section 7 (Capability Concealment) is
  asserted rather than evidenced, and the strongest cited study (Brynjolfsson
  et al.) is a counterexample at a load-bearing point.
- The self-citation dependency (Senn 2026 for K ≠ K′) is a venue vulnerability.

## Distribution integrity: treecheck (2026-08-14)

- `aiki treecheck` detects files with no recognized structural disposition.
  Infers common relationships aggressively; keeps intentional standalone
  material explicit in a small exception file.
- First run found eight real stale artifacts without broad false positives.

## Provenance repair (2026-08-14)

- Git history discontinuity from repository restart was repaired by joining the
  historical and current repositories as one lineage. Abandoned refactor
  experiments preserved as side branches. Colliding pre-restart tags namespaced
  under `pre-restart/`.

## Alpha release prep (2026-08-14)

- Authorship boundary formalized: Will Senn retains design authority. AI
  produces implementation under that authority. AI output has no authority
  merely because it was generated.
- Publication hygiene completed: historical debugging artifacts removed from
  reachable history, no secrets/credentials/personal paths in tracked content.

## Grammar as sole syntax authority (2026-08-14)

- D2: Newline policy remains behavior-preserving while grammar analysis is made
  explicit. `grammar.ebnfx` now declares the newline token, preceding token
  classes that trigger termination, and delimiter pairs that suppress it.
- `}` can end a complete expression (function literal is a primary) but is not
  a newline-completion token. A function literal can therefore continue as a
  call across a newline. This is a deliberate design choice, not an accident.
- Evaluator validation closes bidirectionally over 32 productions with dead
  handler detection. Formatter coverage is explicit with six parent-handled
  cases; unknown leaves cannot be silently dropped. BINOP membership is owned
  by syntax with bidirectional checking.
- `AMBIGUOUS = ( [ -` and the derived overblocked set
  `* + . / < <= > >= and or |>` are now executable analysis results.

## Negative-fixture tooling (2026-08-14)

- `# @negative parse` declares intent; smoke gold is evidence. Smoke checks
  the two bidirectionally.
- Parse-negative specimens are scoped to `*_smoke.ai` fixtures only; the same
  marker in ordinary source is an error.
- Recursive fmt/lint traverse past individual malformed files and accumulate
  errors. CLI dispatch propagates integer subcommand statuses.
- Lint module resolution for public package names was replaced with
  `ModuleRegistry` resolution, fixing a hand-written filesystem heuristic.

## Centralized grammar analysis (2026-08-15)

- D3: "Parse once. Derive shared structural facts once. Consume everywhere."
  Grammar package owns one cached analysis; consumers use it instead of
  independently traversing grammar expressions.
- Grammars are analyzed once at load. Tests that mutate a grammar must call
  `Reanalyze()` explicitly.

## Post-grammar hardening (2026-08-15)

- Parser newline suppression tracks expected closing delimiters instead of one
  aggregate depth. Stray closers cannot drive depth negative.
- Cached newline analysis is an optional diagnostic refinement, not a parse
  prerequisite. Unavailability is observable rather than silently degraded.
- Audit findings ledger (`docs/audit-findings.md`) uses stable `AF-###`
  identifiers with explicit dispositions.

## Relocatable release distribution (2026-08-15)

- The executable identifies its own distribution. Grammar and prelude are
  embedded; shipped modules resolve relative to the executable.
- Named-package discovery does not scan the process working directory. Installed
  Aiki uses executable-relative `lib/` and `vendor/`; development uses explicit
  `./lib` and `./vendor`; user packages live under `~/.aiki/lib`.
- `make dist` produces an unpacked distribution and archive. `make distcheck`
  tests the archive from an unrelated directory.

## Self-description and language services (2026-08-15)

- Independent Aiki front end (lexer, normalizer, parser) written without Go
  front-end facilities. Grammar reader derives lexical and newline policy facts
  from `grammar.ebnfx`.
- `equal()` is atomic and non-structural for lists. The then-current
  non-short-circuit behavior of `and`/`or` was discovered by running code, not
  inferred; this later became D4 and was corrected on 2026-08-16.
- `to_symbol("foo")` and `shaped(:point, [1,2])` were added to close a
  value-model sufficiency gap found by self-hosting.
- Self-hosted module loading delegates host `import()` only for platform
  effects; Aiki-source modules are loaded end-to-end by the self-host
  interpreter.
- `[@error, ...]` is a recoverable value, not an evaluator halt. The self-host
  evaluator uses private `:self_fault` for its own halting propagation.
- Three editor clients (Xed, nvi, VS Code) confirmed the LSP adapter model.
  Desktop editor PATH is not evidence of the editor's PATH.
- Full self-interpretation succeeded: Go runs Aiki interpreter; that runs
  another Aiki interpreter; the inner program `1 + 2 * 3` returns `9`.
  Performance was unlocked by rune snapshotting and lexical dispatch cleanup,
  not by architectural change.

## Experiment framework (2026-08-15)

- Experiments separate procedure, observation, and interpretation in the
  filesystem. This makes it harder to rewrite expectations after seeing results.
- Sequence identity comes from the distribution; creation occurs out of tree.
  Promotion is manual.

## Thompson 7094 regex reconstruction (2026-08-15 through 2026-08-16)

- Machine → paper → compiler: the narrative is deliberately sequential.
  The emulator required no language extension.
- Historical discrepancy in Thompson 1968: the printed final listing sends
  location 4 to `CODE+13`, but the compiler source overwrites that to
  `CODE+16`. The reconstruction independently confirms `CODE+16` by executing
  the compiler's own logic. Both variants are preserved as evidence.
- The test runner initially ran `*_test.ai` as ordinary programs rather than
  through `aiki test`. Aiki assertions accumulate without failing unless the
  test command reports them. This was a genuine gate defect, not a test defect.
- Aiki's left-to-right evaluation and then-eager `and`/`or` behavior produced
  several compiler bugs during reconstruction (e.g., `i < length(xs) - 1`
  evaluates as `(i < length(xs)) - 1`). These were language-semantics findings,
  not machine-semantics findings. The `and`/`or` portion was later corrected by
  D4.
- Phase IV added the operator monitor with octal presentation, disassembly,
  Thompson-specific views, and replayable command files.

## HAL redesign (2026-08-16)

- 117 registered primitives classified into five architectural roles: host
  capabilities, language/evaluator intrinsics, native/value primitives,
  native/FFI library providers, runtime/tooling services.
- Three-name invariant at every host crossing: Aiki name (meaning), HAL name
  (contract), substrate name (realization/provenance).
- Capability domains within the HAL (file, process, env, time), not a new
  architectural layer.
- Authority separated from scope: `AuthorityForSource` grants per-source
  bindings translated to canonical HAL identities. Filesystem path bootstraps
  trust but does not define it.
- Spawn authority defect identified and targeted: `NewIsolatedEnclosedEnv`
  inherits `ScopePrelude`, potentially exposing raw `_` primitives to
  user-spawned code.
- Runtime ownership: args, env, cwd, I/O, RNG, file reader caches, module
  registry, test state, and REPL environment are all runtime-scoped. Working
  directory is runtime-owned (does not mutate OS process cwd).
- Canvas narrowed from 17 individual primitives to `HAL.canvas.open`,
  `HAL.canvas.command`, plus resource accessors.
- Host-backed channel endpoints (not a new Transport type) are the candidate
  for future Canvas event integration. Evidence: `time.after` already creates
  host-produced receive-only channels that work with ordinary `select`.
- Programmer-facing surface added: `file.stat`, `file.rename`, `file.mkdir`,
  `file.mkdir_all`, `file.remove_all`, `file.temp`, `file.temp_dir`,
  `file.copy`, `file.size`; `path` module (pure Aiki except `separator` and
  `cwd`); `system.cwd`, `system.chdir`, `system.exec`; `time.now`.
- M1–M7 implementation migrations completed serially with evidence gates.

## Distribution formatting and project style (2026-08-16)

- Two formatting layers: `aiki fmt` (canonical, width-unaware) and
  `aiki distfmt` (project presentation, fixed 100-column threshold).
- `aiki fmt` preserves explicitly expanded lists/calls/parameters without
  selecting expansion itself. `distfmt` chooses where to expand.
- Bounded fixed-point iteration ensures one `distfmt` invocation produces final
  layout even from noncanonical input.
- Go restyling is AST-position-driven; source text inside strings/comments
  cannot be mistaken for structure.
- Compact calls containing multiline function bodies are not falsely expanded
  (immediate delimiter/element layout, not descendant span).

## Profiler calibration baseline (2026-08-16)

- Semantic counts identical across alpha-25 and beta-1 for all native/sanity
  workloads. Two-level self-host counts shift proportionally due to expanded
  baseline bootstrap, confirming no regression from HAL redesign.

## Guide mode for the Thompson 7094 monitor (2026-08-16)

- Added `guide on/off` mode to the monitor console. Guided trace shows
  Thompson's published CACM comments as the primary layer, modern gloss
  underneath, symbolic named addresses, and region context for compiled code.
- `info machine` and `info thompson` provide reference cards. `demo` runs a
  self-contained guided walkthrough. All fixed output fits 80×24.
- Annotation corpus is a static Aiki list of address/comment/gloss triples,
  separate from display logic. Covers CNODE, NNODE, FAIL, XCHG, INIT,
  TSXCMD, TRACMD, GETCHA, and FOUND.
- Compiled CODE region receives region context ("compiled pattern code"),
  not fabricated Thompson commentary.
- Store is a dense indexed array, not a sparse map. Using `store.new(256)`
  for address-keyed lookup with addresses at 1000+ fails. Simple list scan
  is the correct Aiki pattern for small static lookups.
- `and`/`or` non-short-circuit trap recurred in range checks. This motivated
  D4: make `and`/`or` lazy logical control operators while preserving eager
  function calls and eager ordinary binary operators.

## Lazy logical control operators: profiler comparison (2026-08-16)

- Pre-lazy and post-lazy profiler runs show no semantic-count change for
  sanity and native loop workloads. Arithmetic, comparison, and iteration
  counts match exactly in those fixtures.
- Self-host workloads improve after lazy `and`/`or`. Two-level self-host runs
  show lower wall-clock time, fewer allocation bytes, and fewer mallocs.
- The semantic-count shift is the main explanation: comparison, call, and index
  counts generally drop because guarded right-hand expressions are no longer
  evaluated when the left operand determines the result.
- This is expected evidence of the lazy logical-control change: it preserves
  native behavior while reducing unnecessary self-host evaluator work.

## HAL capability gates and host affordances (2026-08-17)

- Gate 1 selected as the next HAL effort: centralize HAL architectural metadata,
  add invariants, then add capability/profile metadata without adding a dispatch
  layer.
- Architectural pattern: centralize identity, decentralize concern, validate
  composition. `engine/runtime/hal` owns operation, authority, capability, and
  profile metadata; Go substrate files retain realization/provenance.
- Capability availability and authority are independent. Source-level queries
  use Aiki names via `system.has` and `system.require`, never raw HAL identities.
- Phase 1 introduces common `io.read`, `io.read_line`, and `io.write` over
  `:stdin`, `:stdout`, `:stderr`, and file handles. Stdin buffering is runtime
  owned so repeated line reads cannot lose buffered input.
- Phase 2 adds `file.walk`, `file.symlink`, and `file.read_link`; symlink is an
  optional capability. Relative symlink targets remain literal while link paths
  follow runtime-owned cwd semantics.
- Phase 3 adds `file.permissions` and `file.chmod` using the substrate's portable
  permission vocabulary. Unix-shaped mode bits are not asserted to describe all
  host permission models.
- Validation in this environment is limited: repository requires Go 1.24, while
  the available toolchain is Go 1.23.2 and network access prevents toolchain
  download. `make validate` is therefore required on the user's local tree.

## Engine authority centralization — Gate 2 (2026-08-17)

- Gate 1 completed locally by the user with `make validate` passing after the
  executable `io.read_line` documentation example was corrected to retain its
  write/close side effects through explicit expected results.
- Gate 2 scoped as behavior-preserving engine authority centralization. Governing
  rule: centralize identity, decentralize concern, validate composition.
- Gate 2 does not move authored artifacts merely for locality. It moves or
  centralizes architectural ownership and reusable derivation under `engine/`.
- Initial inventory already shows grammar/evaluator coverage is substantially in
  the desired state: grammar owns cached structural analysis and evaluator
  validates its handlers against that authority.
- First concrete Gate 2 candidate: named-module registry and module-root policy
  currently live under `engine/runtime/hal/substrate` despite being consumed by
  runner, language services, imports, help, and invariants. Treating this as
  substrate authority leaks a concrete implementation into engine-level
  consumers.
- Next action: move the module registry/policy to an engine-owned runtime package
  without changing resolution semantics, then retarget consumers and invariants.

- Gate 2 probe found an existing documentation asymmetry: prelude help coverage is
  complete, but `truncate` has no prelude `.doc` entry. Recorded as AF-017; Gate 2
  does not strengthen that contract silently because this project is
  behavior-preserving.

- Gate 2 implementation completed pending local `make validate`: module registry
  and root policy are engine-owned under `engine/runtime/modules`; module source
  metadata is derived once from the real AST; prelude source/help/docs join
  through `prelude.LoadCatalog`; runtime primitive roles are authoritative under
  `engine/runtime/primitives`; language services no longer reconstruct prelude
  exports or depend on the Go substrate for primitive vocabulary.
- Disposable Go 1.23 probe (go.mod lowered only in `/tmp`) passed focused tests
  for `engine/runtime/modules`, `engine/runtime/prelude`,
  `engine/runtime/primitives`, and `engine/language`. Full validation remains
  unavailable in this environment because Go 1.24 and uncached ebiten dependencies
  cannot be downloaded.
- Exact next action: apply the Gate 2 overlay to the user's `hal-capability`
  branch, delete the two retired substrate registry files, rebuild, and run
  `make validate`.

## Invariant system overhaul — Gate 3 (2026-08-17)

- New authoritative baseline: commit `9cc220c` on branch `hal-capability`.
- Gate 2 and the prelude help/doc coverage correction are committed and validated locally by the user with `make validate` passing.
- Gate 3 begins as an invariant-system overhaul, not a feature effort.
- Primary build contract: `make test` checks behavior, `make invariant` checks architecture, and `make validate` requires both.
- Critical rule: an architectural invariant is not considered protected until a negative test proves that violating it is detected through the same invariant path used by the real gate.
- HAL is the first full specimen; engine authorities and representation guardrails follow.
- Exact next action: inventory the current invariant test package and Makefile wiring, classify checks, then introduce the dedicated `make invariant` target without weakening existing validation.

### Gate 3 implementation state (2026-08-17)

- Added dedicated `make invariant`; `make test` excludes invariant/boundary packages and `make check`/`make validate` require `make invariant`.
- Reclassified execution-heavy self-host tests to `test/conformance` and exact-number execution behavior to `test/contract`; executable documentation examples moved to conformance with shared doc-example parsing support.
- HAL metadata now has a reusable validator with negative mutations for duplicate identity, unknown capability operation, and unknown profile capability.
- HAL capability/profile rules are pure architectural functions used by GoRuntime and directly mutation-tested.
- Runtime host binding coverage now compares canonical identities rather than fixed counts and has a missing-binding negative test.
- Primitive-role validation no longer depends on fixed totals; runtime primitive registrations are compared by exact name and role with missing/wrong-role negative tests.
- Prelude source/help/doc negative assurance now runs under the invariant gate through the same catalog validation path.
- Library export/help/doc coverage uses a shared coverage validator with missing/phantom negative tests.
- Added engine layer invariant: only the HAL substrate itself and `engine/runner` composition root may import the concrete Go substrate; injected leak is rejected.
- Added runtime-ownership invariant over Aiki-facing I/O/system implementations to prevent regression to ambient `os.Stdin`, `os.Stdout`, `os.Stderr`, `os.Args`, environment, or cwd APIs; injected regression is rejected.
- Added exact-number architecture invariant: production Go under `engine/semantics` and `engine/syntax` contains no float types or conversion paths; injected `float64` is rejected. Host-side graphics/time/inexact math remain outside this boundary.
- Disposable Go 1.23 probe compiled `engine/runtime/hal`, `engine/runtime/prelude`, `engine/runtime/primitives`, and shared doc-example support. Full invariant packages remain environment-limited by unavailable Ebiten download; user-side Go 1.24 validation is required.
- Exact next action: apply Gate 3 delta on `hal-capability`, rebuild, run `make invariant`, then run `make validate`; any invariant failure is treated as a Gate 3 defect before commit.

## Alpha-29 contract-surface audit cleanup (2026-08-17)

- Audited major contract surfaces after Gates 1–3; no broad Gate 4 justified.
- AF-018 resolved: randomness now has canonical HAL contracts `HAL.random.seed` and `HAL.random.below`, a `random` capability, desktop profile requirement, Go provenance, and host registration. Removed `_seed`/`_random` from duplicated non-HAL role metadata.
- Primitive-role validation now rejects any host-role primitive lacking a canonical HAL operation; negative assurance added.
- AF-019 resolved: canonical HAL descriptor vocabularies are validated, including contexts, effects, blocking classes, lifetimes, optionality, and error contracts; negative assurance mutates each family.
- AF-020 resolved: `spawn` no longer falls back to ambient stderr. Runtimes must supply asynchronous fault reporting before spawn launches; GoRuntime already does. Runtime-ownership invariant now covers the intrinsic implementation.
- Focused disposable Go 1.23 probe compiles `engine/runtime/hal` and `engine/runtime/primitives`. Full substrate/invariant validation remains user-side because uncached Ebiten and Go 1.24 cannot be downloaded here.
- Next action: apply cleanup overlay, rebuild, run `make invariant`, then `make validate`.

## Systems programmer convenience refinements (2026-08-17)

- New authoritative baseline: commit `20bc799`, release line `v0.4.0-alpha-29`, clean `master`.
- Project branch: `systems-convenience`.
- Proposal: `proposals/systems-programmer-convenience-refinements.md`.
- Scope is deliberately small: Phase 1 natural scalar ordering plus list sorting, Phase 2 pure `lib/number` base conversion, Phase 3 whole-file text/byte convenience. Phase 1 includes a semantic comparison refinement; no phase adds HAL capability.
- Phase 1 refinement: Aiki natural scalar ordering is centralized for numbers, strings (rune-lexicographic), runes (code point), and symbols (name). `<`, `>`, `<=`, `>=`, and default `list.sort(xs)` share that contract. Mixed types and composites remain unordered; `list.sort(xs, fn)` supplies explicit ordering. Sorting remains stable and non-mutating, with the pure-Aiki merge sort isolated behind an implementation seam for later native/FFI replacement if profiling justifies it.
- Phase 1 semantic decision recorded as D5. Exact next action: gate the combined systems-convenience implementation locally with `make invariant` and `make validate`; treat any failure as project work before commit.

### Systems convenience implementation state (2026-08-17)

- Phase 1 implemented in pure Aiki: stable non-mutating merge sort exposed as `list.sort(xs)` and `list.sort(xs, fn)`. Natural scalar ordering is centralized for numbers, strings, runes, and symbols and shared by comparison operators and default sort; custom comparator must be a function returning boolean. Merge sort is isolated behind a private implementation seam for later native/FFI replacement if profiling justifies it.
- Phase 2 implemented as new pure `lib/number`: `to_base` / `from_base`, bases 2..36, signed exact integers, lowercase formatting, upper/lowercase parsing, shaped domain errors, no HAL or float path.
- Phase 3 implemented by composition in `lib/file`: `read_all`, `write_all`, `read_all_bytes`, `write_all_bytes`. Helpers own open/read-or-write/close and add no HAL operation.
- Help, full docs, executable examples/tests, list tests, number tests, and file behavior smoke coverage were added with each surface.
- Sandbox limitation: authoritative Aiki execution is unavailable because the source-only baseline requires Go 1.24 and toolchain download is blocked.
- Exact next action: apply the project delta locally, rebuild, run `make invariant`, then `make validate`; treat any gate failure as project work before commit.

## Portable systems completeness (2026-08-17)

- Authoritative working baseline: `v0.4.0-alpha-30`, commit `c9769d6`, with the byte-boundary normalization refinement carried forward in this working tree.
- Project branch: `portable-systems`.
- Proposal: `proposals/portable-systems-completeness.md`.
- Goal: eliminate the remaining material gaps for portable Go-like systems programming while keeping policy/composition in Aiki and HAL limited to irreducible host mechanisms.
- Governing additions: sufficient portability rather than host-union completeness; every new byte-oriented host resource must first be tested against the existing `io` endpoint abstraction.
- Phase order: process lifecycle/pipes -> signals -> TCP/UDP networking -> terminal/TTY -> portable file locking -> final completeness audit.
- Exact next action: inventory current process HAL operations, `system.exec`, runtime resource ownership, and `io` endpoint dispatch; define the smallest Phase 1 contract before changing behavior.

### Portable systems Phase 1 implementation state (2026-08-17)

- Process lifecycle surface added as `lib/process`: `start`, `stdin`, `stdout`, `stderr`, `wait`, and `terminate`.
- Added opaque runtime-owned `process` and `endpoint` values. Process pipes are generic endpoints consumed by existing `io.read`, `io.read_line`, and `io.write`; no process-specific read/write API was introduced.
- Added `io.close` because closing child stdin is required to deliver EOF and complete pipeline/interactive process semantics. It also accepts file handles, preserving I/O convergence.
- Added canonical HAL operations, Go provenance, authority grants, capability membership, host registrations, and runtime-owned process/endpoint resource maps.
- Runtime teardown is generalized through `CloseAllResources`, which terminates/reaps child processes before closing canvases; existing canvas-only cleanup remains available for focused tests.
- `process.wait` is idempotent and returns the exit code, including nonzero codes; start failures remain shaped `:process` errors. `process.terminate` is the portable forceful termination primitive; graceful signaling remains Phase 2.
- Focused Go tests cover attached stdin/stdout/stderr, EOF by closing stdin, nonzero exit status, repeated wait, and cross-runtime handle rejection.
- Runtime-ownership invariant now includes process implementation. HAL metadata/role/binding invariants derive the new operations automatically.
- Sandbox probe confirms `engine/semantics/value` and `engine/runtime/hal` compile after the change; full substrate tests remain blocked by uncached Ebiten/Go 1.24 dependencies.
- Exact next action: apply Phase 1 on local `portable-systems`, rebuild, run `make invariant`, then `make validate`; any gate failure is Phase 1 work before Signals begins.

### 2026-08-17 — Portable Systems Completeness, Phase 2

Phase 1 process lifecycle/process I/O passed `make invariant` and `make validate` on the authoritative development tree.

Phase 2 adds portable signals without introducing a parallel event model. `signal.watch(...signals)` returns a receive-only Aiki channel carrying semantic signal symbols, `signal.stop(source)` releases the runtime-owned subscription, and `signal.send(process, signal)` targets the opaque process handle introduced in Phase 1. The portable vocabulary begins with `:interrupt` and `:terminate`; individual host mappings may return shaped `:unsupported` when a semantic signal cannot be represented faithfully. Signal receipt is edge-like and may coalesce rather than acting as a lossless queue.

The Go substrate owns signal subscriptions and tears them down with other runtime resources. The `:signals` capability is required by the desktop profile. Higher-level shutdown, cleanup, forwarding, and supervision remain Aiki policy.

## 2026-08-17 — Portable systems Phase 3: networking

- Added TCP connect/listen/accept using the existing generic I/O endpoint abstraction.
- Added network address inspection for local/remote TCP endpoints and bound resources.
- Added message-oriented UDP bind/send/receive with datagram boundaries preserved.
- Added runtime-owned listener/datagram resources and cleanup.
- Added `:network` HAL capability and desktop-profile requirement.
- No parallel TCP read/write API was introduced; TCP uses `io.*`.

## 2026-08-17 — Portable systems Phase 4: terminal/TTY

- Added minimal terminal HAL: detect, size, raw, restore.
- Terminal operations reuse existing I/O endpoints; no parallel terminal stream type.
- Raw-mode restoration is represented by an opaque runtime-owned token and is restored at teardown.
- ANSI rendering remains Aiki/library policy, not HAL.
- Go substrate uses the already-shipped `github.com/chzyer/readline` terminal primitives, avoiding a new dependency.

## 2026-08-17 — Portable systems Phase 5: file locking

- Added exclusive interprocess locking as the sufficient portable contract.
- `file.lock(path)` blocks until acquired; `file.try_lock(path)` is nonblocking; `file.unlock(lock)` releases.
- Locks are opaque runtime-owned resources and are released during runtime teardown.
- Shared locking is intentionally excluded from the portable contract.
- Go substrate uses `github.com/gofrs/flock` v0.12.1; Aiki remains responsible for locking policy and composition.


## 2026-08-17 — Portable systems Phase 6 audit remediation: runtime environment

- Completeness audit found the system environment was read-only despite the proposal requiring inspect/modify semantics.
- Added runtime-owned environ/set_env/unset_env; neither mutates the embedding process environment.
- system.exec and process.start now inherit a snapshot of the Aiki runtime environment.
- This closes the only material irreducible capability gap found by the Phase 6 source audit.
## 2026-08-17 — Portable systems final reconciliation

- Reconciled the validated Phase 6 implementation line-by-line against `proposals/portable-systems-completeness.md`.
- Confirmed public surfaces, canonical HAL operations, capability/profile membership, authority grants, Go provenance/registration, runtime ownership, and module docs/tests for process, signals, networking, terminal control, file locking, and runtime environment.
- Representative-program pressure check found no additional irreducible host capability: pipeline forwarding and supervisor coordination compose from existing process endpoints, spawn arguments, channels, signals, timers, and select.
- AF-022 found/resolved: common `io` help/docs lagged endpoint convergence and now explicitly describe process pipes and TCP connections as runtime endpoints.
- No material proposal/implementation contradiction remains. Portable-systems completeness is reconciled; showcase work is intentionally separate.


## 2026-08-17 — Experiment 003 Four-Way Life begins

- Baseline: `e09cf60` (`v0.4.0-alpha-30-2-ge09cf60`), branch/workstream `four-way-life`.
- Four-Way Life is experiment `003`, following the repository experiment layout.
- Governing proposal saved as `proposals/four-way-life.md`.
- Gate 1 implementation started as a true five-process design: coordinator + four `aiki worker.ai` children.
- Store is coordinator-only. Workers receive immutable newline-delimited generation frames and return color-specific next-generation proposals.
- Headless acceptance runs the same seed twice and requires identical committed-generation transcripts; canvas mode is optional and renders only committed generations.

## 2026-08-17 — Four-Way Life Gate 2 worker domains

- Gate 1 validated as five real OS processes with deterministic headless transcripts.
- Gate 2 keeps the protocol/barrier unchanged and makes worker domains load-bearing by protecting existing owner cells only, preventing new cross-owner proposal conflicts.
- Engine A: recurring file/path/string/bytes pattern input.
- Engine B: runtime env + deterministic random/hash; wall time is fallback seed only when explicit seed is absent.
- Engine C: project-controlled Aiki helper subprocess + process pipe + regex + base conversion.
- Engine D: pure list/sort/map/filter/reduce + pipeline + shape/match + exact math path.


## 2026-08-17 — Four-Way Life Gate 4

- Gate 3 systems acceptance passed: lock held/released, signal shutdown,
  generation logging, and terminal raw/restore.
- Added Gate 4 hardening acceptance with no new public capability.
- `helper.ai` supports deliberate `FWL_HELPER_FAIL=1` status-7 failure so
  Worker C's local subprocess-failure contract is exercised end to end.
- `terminal_probe.ai` now exits nonzero on actual raw/restore failure; non-TTY
  execution remains an explicit skip.
- `gate4.sh` composes deterministic five-process acceptance, failure recovery,
  and Gate 3 systems acceptance.
- Final completion requires `gate4.sh` plus repository `make validate`.


## 2026-08-17 — Four-Way Life final showcase reconciliation

- `showcase.sh` now launches the intended interactive presentation: coordinator
  in the current terminal, canvas window, and four worker-status terminal
  windows.
- Worker-status terminals display log views from the actual coordinator-owned
  A/B/C/D processes; they do not replace or duplicate the worker protocol pipes.
- Debian `x-terminal-emulator` compatibility was verified against
  `gnome-terminal.wrapper`; the launcher uses the working `-e` form.
- Canvas closure now terminates the coordinator run, stops/reaps workers, and
  causes all four worker-view terminals to stop and close.
- Final showcase architecture and lifecycle reconciled into experiment docs and
  proposal.

## 2026-08-17 — string/ffi from Four-Way Life profiling

- Four-Way Life hot-path attribution identified pure-Aiki `string.split`/`substring`
  and string indexing as a substantial generic allocation/iteration cost during
  line-protocol decoding.
- Added explicit `string/ffi` provider-accelerated sibling while preserving bare
  `string` as the pure/reference implementation.
- Accelerated coarse string operations through RoleProvider primitives; no HAL
  capability or authority surface was added.
- Provider implementations preserve Aiki rune-index semantics and ASCII-only
  whitespace behavior.
- Added native-vs-FFI parity tests with multibyte text.
- Four-Way Life explicitly opts into `string/ffi` in worker/protocol code so the
  optimization can be measured before considering any default-module change.


## 2026-08-18 — Native/FFI separation clean pass

- Clean branch `native-ffi-separation-clean` created directly from supplied baseline `ffe3622` (`v0.4.0-alpha-31-dirty`).
- Prototype branch/tree `native-ffi-separation` is reference-only and is not validation evidence. No prototype commit will be cherry-picked wholesale.
- Preserved supplied baseline dirt exactly: modified Four-Way `results/four-way-life.log` and untracked `results/showcase-20260818-182659/`.
- Froze 395 tracked pre-existing behavioral witness files by Git blob identity in `ai/evidence/native-ffi-baseline-witnesses.tsv`; manifest SHA-256 `f2e46ee66eea1c6fcc963b82974b456894f3c9c07ed802ab62817e99f023b154`.
- Clean-pass rule: implementation and new architecture/conformance tests may change; frozen witnesses do not change until they have completed baseline-preservation validation. Intentional API removals such as misplaced Store helpers occur only afterward as separately identified semantic changes.
- Exact next action: reapply semantic-role/module-policy authority and native/FFI truthfulness cuts without modifying any frozen witness.

## 2026-08-18 — Native/FFI clean pass, preservation-phase evidence

- Clean implementation Cut 1 committed as `f252cc9`: frozen witness manifest + revised semantic-role/module-policy authority.
- Clean implementation Cut 2 committed as `355ca5a`: truthful native paths reconstructed for bits/bytes/math/string; hash/string FFI provider work reconstructed; turtle routed through `math/native`; source import graph support added. Existing behavioral witnesses were not edited.
- Clean-pass reconstruction error found by new module-policy invariant: renamed `lib/bits/ffi.ai` initially still declared package `bits`; corrected to the actual `bits/ffi` provider source before validation claims.
- Disposable Go 1.23 compatibility copy passes `go test ./engine/runtime/modules ./engine/runtime/primitives`. This is focused structural evidence only; authoritative Go 1.24/substrate/full validation remains local-user work because external modules are unavailable in the sandbox.
- Substrate focused compile remains environment-blocked by uncached `github.com/chzyer/readline`, `github.com/gofrs/flock`, and `github.com/hajimehoshi/ebiten/v2`.
- Frozen witness verifier continues to report all 395 baseline witness blobs unchanged.

## 2026-08-18 — Native/FFI clean-pass mutation evidence

- Added source-derived local call graph metadata and invariant requiring every exported function of a portable `/ffi` realization to reach a `RoleProvider` primitive directly or through local Aiki helpers.
- Disposable mutation checks all fail as intended: native→FFI transitive leak, FFI→native fallback, bare-default hijack, provider primitive mislabeled non-provider, and turtle→`math/ffi`.
- Mutation details retained at `ai/evidence/native-ffi-mutation-results.md`; mutations never touched the authoritative clean tree.
- Focused disposable Go 1.23 compatibility check passes modules, primitives, and `cmd/subcommands/tools/check`.
- Store legacy helpers remain intentionally present during the preservation phase; their pre-existing tests have not been edited or retired before a rebuilt baseline-witness run.
- Anti-taint verifier mutation-tested independently: a disposable edit to frozen `lib/bits/bits_test.ai` is detected by blob mismatch.
- Neutral `test/nativeffi/*_contract.ai` corpus parses cleanly under the authoritative grammar in the disposable Go 1.23 compatibility harness.
- Proposal reconciled to the clean branch: Store helper removal is explicitly deferred until the untouched Store tests have completed the rebuilt preservation run.

## 2026-08-18 — Native/FFI post-preservation Store cleanup

- Preservation checkpoint fixed at `1f4e906` and tagged locally as `native-ffi-preservation-candidate`; that commit retains all 395 baseline witness blobs unchanged.
- After that checkpoint, removed misplaced public `store.digits_to_text` and `store.checksum` semantics and their raw primitives/authority/registration.
- Four-Way coordinator now composes `store.snapshot` with explicit `bytes/ffi` for digit-text conversion and uses its existing Aiki `life.store_checksum` logic.
- Exactly three frozen witnesses change intentionally in this API cut: `engine/runtime/hal/substrate/builtins_store_test.go`, `lib/store/store_test.ai`, and `experiments/003-four-way-life/experiment/coordinator.ai`. Post-preservation witness verification allowlists only those files.

## 2026-08-18 — Native/FFI post-preservation primitive-role terminology

- After the preservation checkpoint, renamed the misleading primitive architecture role `native` to `runtime`; this role is reserved for constitutive language/value/runtime atoms, while library provider implementations remain `provider`.
- Exactly three additional frozen invariant tests change intentionally to follow the architecture term: substrate registry-role test, HAL registration invariant test, and primitive-role invariant test.
- Post-preservation witness verification now permits exactly six known witness changes total: three Store API cleanup witnesses and three role-terminology invariant witnesses.

## 2026-08-18 — Native/FFI clean-pass final static reconciliation

- Clean implementation branch is `native-ffi-separation-clean`; prototype `native-ffi-separation` remains reference-only and was never cherry-picked wholesale.
- Regression-preservation checkpoint: `1f4e906`, local tag `native-ffi-preservation-candidate`; all 395 baseline witness blobs unchanged there.
- Final post-preservation API/terminology cuts: `e5e4fe3` removes Store workload helpers; `f725af3` renames primitive role `native` to `runtime`.
- Final witness delta is exactly six allowlisted baseline files: three directly obsolete/updated Store/Four-Way witnesses and three invariant tests updated for the explicit `runtime` role name. No unapproved frozen witness changed.
- Disposable Go 1.23 compatibility harness passes `go test ./engine/runtime/modules ./engine/runtime/primitives ./cmd/subcommands/tools/check ./test/nativeffi`; authoritative Go 1.24 substrate/full gates remain unrun in the sandbox because required external modules are not cached and network access is unavailable.
- Later runtime validation has two distinct checkpoints: run `make rigorous` at `native-ffi-preservation-candidate` to establish behavioral preservation against untouched witnesses, then run `make rigorous` at final branch head to validate the intentional Store/API and role-terminology cleanup.

## 2026-08-18 — Alpha 2 documentation reconciliation begins

- Authoritative baseline: `v0.4.0-alpha-34` / `8a33879`.
- Branch: `alpha2-doc-reconciliation`.
- Governing proposal: `proposals/alpha2-doc-reconciliation.md`.
- Scope: repo-wide README, help/doc, architecture/developer docs, examples, proposals, and durable internal records; remove leftover current-facing cruft while preserving explicitly historical evidence.
- Validation principle: implementation and current executable behavior remain authority; documentation is reconciled to code rather than code changed to match stale prose.

## 2026-08-18 — Alpha 2 documentation reconciliation, pre-validation closeout

- Audited README, all current Markdown docs, shipped library `.help`/`.doc` companions, project proposal statuses, repository READMEs, and both ODT language documents against `v0.4.0-alpha-34` / `8a33879`.
- Current-facing architecture now consistently separates stdlib semantic realization (`native`, `ffi`, capability, interop) from HAL authority/runtime boundaries; deleted paths and stale Store-helper/future-signal wording were removed from current instructions while historical evidence was preserved.
- `This Is Aiki` is the living Alpha 2 guide; RA1 remains the Aiki 1 report with its reference-implementation observations explicitly scoped to Alpha 1 (`v0.4.0-alpha` / `678aeea`).
- Reconciled CLI/tooling documentation including `debug -stage fmt`; replaced brittle release file counts with durable validation categories; corrected direct host-facing Go dependency count to readline, flock, and Ebiten.
- Static evidence: all current Markdown local links resolve; all shipped non-test library sources have `.doc` and `.help` companions; export/help/doc surface audit reports zero gaps; stale-term sweep has no unexplained current-facing hits; `git diff --check` passes.
- ODT evidence: both documents render successfully through LibreOffice as letter-size PDFs (`This Is Aiki` 54 pages; RA1 57 pages), with Alpha 2 title/provenance and Alpha 1 scope visually/textually verified.
- No disposable scratch/build cruft was found in the authoritative baseline. Historical experiment analyses, session history, and `ai/evidence` were deliberately retained.
- Remaining critical gate: apply the reconciliation drop to current Alpha 34 master and run `make validate` (or stronger) without blessing golds.

## 2026-08-19 — Experiment 004 PDP-11/40 V6 emulator begins

- Authoritative baseline: `v0.4.0-alpha-35` / `caade2f`, clean `master`.
- Project branch: `v6-emulator`.
- Governing proposal: `proposals/pdp11-40-v6-lions.md`.
- Experiment: `experiments/004-v6-emulator/`; Experiment 002 Thompson 7094 is the structural/process precedent, not an implementation template to copy mechanically.
- Target is the smallest architecturally faithful PDP-11/40 + UNIBUS machine required to construct and run the Lions V6 laboratory, not a general PDP-11 or all-V6 emulator.
- First campaign gates: (1) deterministic PDP-11/40 core diagnostics; (2) tape bootstrap/bus interaction; (3) standalone installer `=` from original distribution tape; (4) one system disk constructed by V6 `tmrk`; (5) boot that constructed disk through `@rkunix` to V6 `#`.
- Architectural decisions: one authoritative deterministic machine state; UNIBUS is the CPU/memory/device/DMA/interrupt coordination boundary; semantic folders enforce separation of concerns; observation/log/debug do not mutate or pace the machine; operator vocabulary is UNIX/V6 (`tape`, `disk`, `console`, etc.) while DEC controller names remain internal/deep-debug terminology; monitor identity/prompt is `aiki-pdp>`.
- Host-control contract: CTRL-T status without state change; CTRL-E suspends guest execution and returns to `aiki-pdp>`; CTRL-C and CTRL-D retain normal foreground interrupt/EOF effects. Guest hangs must not make the emulator monitor unrecoverable.
- Normative authorities supplied: 1972 PDP-11/40 Processor Handbook, 1979 UNIBUS Design Description, 1976 KT11-D manual, targeted peripheral manuals, Lions source/commentary, and the Lions laboratory setup guide.
- Working method: follow `ai/README.md`; make small serial cuts, retain evidence, update this durable record, report progress during work, and stop for user validation/input at critical design points and between project gates.
- Exact next action: implement Cut 1 in Gate 1 — machine state, byte/word memory, and declarative before/program/after diagnostic harness — without introducing peripheral behavior.

### 2026-08-19 — V6 emulator Gate 1 Cut 1 implementation written, validation blocked

- Added semantic experiment boundaries for `machine/`, `memory/`, `unibus/`, and `diagnostics/`; empty `cpu/`, `monitor/`, and `observe/` directories are not retained because Git does not track them and implementation has not begun there.
- Physical backing memory is explicitly 18-bit byte-addressed (262,144 bytes). Word access is little-endian and odd-address word access returns a shaped guest architectural fault rather than an emulator/host failure.
- Machine state contains eight 16-bit registers (R6/SP, R7/PC), 16-bit PSW, run state, deterministic step count, and logical machine time. Width reduction is centralized in `machine/width.ai`, including negative exact integers.
- Gate-1 UNIBUS currently routes byte/word accesses only to physical memory and exposes the single future device-advancement seam; no device behavior exists yet.
- Diagnostic harness establishes declarative before state, program words, bounded-step count, and expected after registers/PSW/memory. Cut 1 self-tests cover width wrapping, byte/word layout, odd-address fault classification, reset, machine/UNIBUS memory authority, and fixture comparison.
- Static evidence: `git diff --check` passes; manual source review completed against current Aiki syntax/examples.
- Validation limitation: no `aiki` executable is present in the supplied baseline or PATH. Building is blocked because `go.mod` requires Go 1.24 while the container has Go 1.23.2 and network access blocks toolchain download. Therefore Cut 1 remains ACTIVE, not GATED.
- Exact next action: run `aiki test experiments/004-v6-emulator/experiment/diagnostics/cut1_test.ai` with an Alpha-35-compatible Aiki executable. Any failure is Cut 1 work. Do not begin Cut 2 until this focused gate passes.

## 2026-08-19 — PDP-11/40 V6 emulator, Gate 1 Cut 1 gated / Cut 2 active

- Authoritative branch is `v6-emulator`; experiment is `004-v6-emulator`.
- User validation gated Cut 1: `aiki test experiments/004-v6-emulator/experiment/diagnostics/cut1_test.ai` -> 42 tests, 42 passed, 0 failed.
- Cut 2 removes local `reset` shadowing, adds fetch/decode and a single operand-resolution authority for PDP-11/40 addressing modes 0-7, including PC special forms and byte-vs-word register side effects.
- Added handbook-derived PDP-11 operator help for addressing, registers, PSW, instruction octal base codes, branches, traps, and selected mnemonics. Standalone `experiment/help.ai` is the current reader; the future `aiki-pdp>` monitor will use the same help authority.
- Process correction: active Cut-2 edits were detected in a scratch extraction before commit and migrated into the authoritative `v6-emulator` tree at `7e2596d`; the scratch tree is not evidence.
- Sandbox executable validation remains blocked by Go 1.23 vs repository Go 1.24 and no built `aiki`. Disposable Go-1.23 compatibility evidence passes `go test ./engine/syntax/...`, and every Cut-2 Aiki source lexes/parses under the authoritative grammar. Next evidence is user-run focused Cut-2 tests.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 2 addressing gate repair

- User validation kept Cut 1 green at 42/42 and gated PDP help at 16/16, but addressing returned 13/27 with 14 identical `unknown shape: operand` faults plus one `read` prelude-shadow warning.
- Root cause is Aiki lexical shape vocabulary, not PDP-11 addressing semantics: the shaped `@operand` value crossed the module boundary, while field names declared in `cpu/addressing.ai` did not become known in the importing diagnostic environment.
- Repair keeps `@operand` private to the addressing module and exports semantic accessors for kind/register/address; operand reading is renamed `read_operand` to remove the prelude shadow. No PDP-11 mode, side-effect, PC-relative, byte/word, or fetch semantics changed.
- Cut 2 remains ACTIVE until the repaired focused addressing test passes under the user's Alpha-35 Aiki executable.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 2 gated / Cut 3 execution core

- User validated the repaired Cut 2 addressing suite successfully; Cut 2 is GATED. Cut 1 remains GATED at 42/42 and PDP help at 16/16.
- Durable scope correction: KT11-D memory management is required before `rkunix`/V6 proper, but remains deliberately after tape bootstrap and standalone tape-to-disk construction. Proposal gates are now Gate 5 MMU, Gate 6 boot constructed disk into V6.
- Cut 3 is ACTIVE: added PSW/NZVC authority, architectural operand writes including MOVB register sign extension, representative double/single operand execution, full conditional branch family, JMP, JSR/RTS, HALT/NOP/condition-code operators, deterministic per-instruction machine/UNIBUS advancement, and bounded run.
- Execution distinguishes effective-address resolution from destination reading: MOV and CLR resolve destinations without performing an unnecessary read, preserving the correct future boundary for write-only or side-effecting device registers.
- The declarative Gate-1 fixture now owns bounded execution as well as before/after comparison, preserving one diagnostic authority. Focused Cut-3 tests include exact register/PSW/control-flow/stack behavior and a bounded accumulator program ending in HALT.
- Sandbox cannot run the Alpha-35 Aiki executable. Disposable Go-1.23 syntax evidence passes `go test ./engine/syntax/...`; all changed Cut-3 Aiki files lex and parse under the authoritative grammar. Next evidence: user runs `aiki test experiments/004-v6-emulator/experiment/diagnostics/cut3_execution_test.ai`.


### 2026-08-19 — PDP-11/40 V6 emulator, Cut 3 expanded to m40.s contract

- User directed Cut 3 to continue until it fulfills the actual Lions `m40.s` instruction workload rather than stop at a representative subset, and requested runtime audit of instruction and addressing-mode use.
- Static cross-check of Lions' commentary against the supplied `m40.s` source identifies 45 instruction forms in the workload; the cross-check caught omissions in the earlier hand inventory (`mul`, `ror`, `wait`), so the source-backed manifest now governs Cut 3.
- Cut 3 now includes execution support for the missing `m40.s` forms including ADC/SBC, SWAB, ROR/ASR/ASL, MUL/DIV/ASH/ASHC, SOB, MFPI/MTPI, RTT, WAIT, and RESET. Handbook review corrected the draft SBC carry rule before gating.
- Runtime audit is attached to the machine and counts decoded instruction kinds plus source/destination addressing modes through the execution path. A readable audit renderer uses octal-first numeric presentation; machine values and audit counts are not rendered in decimal.
- `MFPI`/`MTPI` are not claimed as full KT11-D behavior here: their instruction/stack-transfer boundary is exercised while previous-mode translation remains explicitly owned by Gate 5. Likewise WAIT/RESET/RTT establish the CPU/bus/control seams that later interrupt, device, and MMU gates complete.
- Added `cut3_m40_test.ai`: requires a 45-form source-backed contract, decoder recognition of every form, semantic witnesses for the newly admitted instruction families, and audit/addressing-mode evidence.
- Sandbox evidence: all Experiment 004 Aiki sources lex and parse under the authoritative grammar; `git diff --check` passes. Cut 3 remains ACTIVE pending user execution of both Cut-3 diagnostic files.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 3 packaging repair

- User validation reconfirmed Cut 2 addressing at 54/54. The supplied continued-Cut-3 drop then exposed a packaging defect rather than a new semantic defect: `cut3_execution_test.ai` was absent, and the user's tree still lacked the Cut-3 `write_operand` export from `cpu/addressing.ai`.
- Root cause: the continued-Cut-3 tarball was built from only the later `56f75c2` delta, even though it was intended to be applied directly over the gated Cut-2 baseline. It therefore omitted files introduced by `b5afeaf`, including `cpu/addressing.ai`, `cpu/psw.ai`, `diagnostics/cut3_execution_test.ai`, and the updated fixture.
- The observed `unknown shape: error` failures occurred downstream of this incomplete module surface and are not yet evidence against the PDP instruction semantics.
- Repair policy: no CPU semantic changes. Reissue Cut 3 cumulatively from gated Cut 2 (`d802169`) through the current m40 completion, including every file required by both Cut-3 diagnostic suites. Cut 3 remains ACTIVE pending user execution of both suites against the cumulative drop.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 3 semantic/grouping repair

- User validation against the cumulative Cut 3 drop: `cut3_execution_test.ai` -> 62/64, with two `cannot compare boolean and number` faults in bounded execution/fixture execution; `cut3_m40_test.ai` -> 103/105, with one ROR PSW expectation mismatch (`expected 11, got 9`) and one `cannot subtract boolean and number` fault in audit rendering.
- Handbook review confirms the ROR implementation was correct and the diagnostic expectation was wrong: ROR sets V to N xor C. For source `1` with incoming C=1, the result is `0100000`, N=1, C=1, V=0, hence NZVC value 011 octal (9 decimal).
- The two runtime type faults are Aiki left-to-right grouping defects, not PDP semantics: `run_bounded` now groups `(count < limit)`, and the audit renderer groups `(length(groups) - 1)` before comparison.
- No CPU instruction semantics changed in this repair. Cut 3 remains ACTIVE pending rerun of both focused suites.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 3 audit/PSW repair

- User validation: `cut3_execution_test.ai` is green at 73/73. `cut3_m40_test.ai` is 108/109; the only remaining failure is the audit-report assertion expecting an empty string while the renderer correctly returns the populated octal-first report.
- Audit contract is extended without changing PDP instruction semantics: every completed instruction now records full 16-bit PSW before/after and a bitwise change mask. The audit retains the latest transition plus per-bit change counts, avoiding an unbounded per-instruction trace during a future V6 boot.
- The report renders the latest PSW transition as octal `before`, `after`, and `delta`; condition-code changes are named `N Z V C` when present. Full masks are retained so later mode/priority changes from RTT/traps/interrupts are not discarded.
- Added `reference/octal.parse` (using the proven Experiment 002 monitor pattern) so PDP-facing tests/configuration can state machine values in octal strings even though Aiki numeric literals themselves are decimal.
- No CPU instruction semantics changed. Cut 3 remains ACTIVE pending the focused `cut3_m40_test.ai` rerun.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 3 gated / Cut 4 raw-tape bootstrap

- User validation gated Cut 3: `cut3_execution_test.ai` passed 73/73 and `cut3_m40_test.ai` passed after the PSW-transition audit repair. Cuts 1–3 are now GATED.
- User chose the bit-exact TUHS raw V6 tape artifact as the emulator media authority rather than OpenSIMH `enblock` framing. Uploaded `v6.tape.gz` expands to exactly 6,195,200 bytes = 12,100 fixed 512-byte records. Raw SHA-256 observed in the development environment: `18e3cff96933f7a2ced81050ca101507eed55a2d137d1fcfd7745ebaf1d4c2a5`.
- Cut 4 is ACTIVE. Raw media is isolated under `media/v6_tape.ai`; it treats the V6 artifact as fixed records plus a synthetic terminal tape mark. The master media is read-only.
- Gate-2 UNIBUS now owns the PDP-11/40 no-MMU I/O-page projection: CPU addresses `160000`–`177777` map to physical UNIBUS `760000`–`777777`. This is the first gate where that mapping is observable.
- TM11 behavior is isolated under `devices/tape/tm11.ai`; UNIBUS supplies physical DMA writes during deterministic device advancement. Initial scope is only MTS/MTC/MTBRC/MTCMA and the READ operation used by the six-word bootstrap. Controller INIT does not rewind the attached tape.
- Historical bootstrap execution exposed a CPU ordering defect not exercised by the earlier core fixtures: in `MOV R0,-(R0)`, the source value must be captured before destination autodecrement changes R0. Cut 4 corrects double-operand execution ordering rather than compensating in tape code.
- Static evidence: `git diff --check` passes; disposable Go-1.23 `engine/syntax/...` tests pass; all Experiment 004 Aiki files lex/parse against the authoritative grammar. Full Aiki execution remains user-side because the development container cannot build the Go-1.24 executable without network access.
- Exact next gate: run `cut4_tape_test.ai`, then run `tape_bootstrap.ai` against the uncompressed raw `v6.tape`. Gate 2 requires PC `100012`, tape position 1, MTCMA `001000`, and the known raw-record signature in memory. Do not begin console/standalone execution until user validates this gate.
- Observation decision recorded during Cut 4: future UI uses a separate UNIBUS observer plus one read-only semantic window per device family (tape, disk, paper tape, printer, console, clock), with CPU/debug separate. UI work is deferred; device/UNIBUS semantics must expose observation without accepting mutation from observers.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 4 observability repair

- User validation gated the synthetic raw-tape/controller suite at 31/31 after correcting shaped-record payload indexing (`@record` payload is element 0).
- Before real-media validation, the bootstrap runner now emits a concise octal-first execution narrative: each executed bootstrap instruction, the TM11 READ event, observed UNIBUS DMA range, TM11 byte-count transition/completion, final PC/device state, and known first-record memory signature.
- The trace reads authoritative machine/device state before and after architectural steps; it does not create a second mutation path. This is the same read-only state seam intended for future UNIBUS and per-device observer windows.
- No tape, CPU, or UNIBUS semantics changed for observability. Exact next gate remains real `v6.tape` bootstrap validation.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 4 gated / Cut 5 monitor and observers

- User real-media validation gates Cut 4: `tape_bootstrap.ai` against the TUHS `v6.tape` executed the historical six-word bootstrap, reported TM11 READ record 0 and UNIBUS DMA `000000..000777`, ended at PC `100012`, advanced tape to record `000001`, left MTCMA `001000` / byte count `173526`, loaded `000407 000654` at memory 0, and printed `TAPE BOOTSTRAP PASS`.
- The tiny `cut4_trace_test.ai` harness repair (`use("test")` / direct `run` and `equal`) is reconciled into the authoritative tree; it had been delivered separately after the earlier invalid `test.new()` use.
- Cut 5 is ACTIVE before Gate 3. Experiment 003 is the host-window/lifecycle precedent; Experiment 002 is the octal monitor-command precedent.
- `aiki-pdp.ai` is the normal control surface. Terminal input is raw and read by an isolated input process, but the main monitor event loop remains the single writer/stepper of machine state. CTRL-T is status-only; CTRL-E suspends; CTRL-C interrupts the foreground run; CTRL-D at an empty halted prompt is EOF/quit.
- Observer topology is one CPU/debug view, one UNIBUS view, and one view per implemented device family; Cut 5 currently instantiates CPU, UNIBUS, and tape. Future disk, paper-tape, printer, console, and clock views use the same contract.
- Review caught and corrected an observation-path design violation before gating: synchronous TCP writes would allow a slow observer to pace execution. Each observer now has a host-backed latest-snapshot mailbox and independent sender process. The machine loop only replaces snapshots; slow views coalesce updates and cannot block guest stepping.
- `showcase.sh`, following Experiment 003, owns terminal-emulator discovery and opens observer windows. Closing an observer drops only its read-only endpoint and does not alter machine state.
- Exact next evidence: user runs `aiki test experiments/004-v6-emulator/experiment/diagnostics/cut5_monitor_test.ai`, then `experiments/004-v6-emulator/experiment/showcase.sh`; manually attach `v6.tape`, deposit the six bootstrap words, `run 100000`, exercise CTRL-T/CTRL-E, and confirm `examine 0 2` plus CPU/UNIBUS/tape windows. Do not begin KL11/standalone Gate 3 until this cut passes.

### 2026-08-19 — PDP-11/40 V6 emulator, Cut 5 spawn-isolation repair

- User manual showcase validation reached the `aiki-pdp>` prompt but the observer broker failed immediately in its spawned accept loop with `undefined: net`; startup also emitted prelude-shadow warnings for local names `input` and `help`.
- Root cause is Aiki spawn isolation, not networking semantics: spawned computations do not capture module-local imports. Experiment 003 already demonstrates the correct pattern by importing `net` inside its spawned accept loop.
- Repair applies that rule consistently to every Cut-5 spawned computation, not just the first symptom: broker `accept_loop` imports `io`/`net` locally; broker `sender` imports `io`/`store`/`time` locally; monitor input `reader` imports `bytes/ffi`/`io` locally.
- Warning cleanup renames the main monitor module alias `input` to `monitor_input` and exported monitor function `help` to `help_text`; operator command remains `help`.
- No PDP, tape, UNIBUS, monitor-command, or observer-topology semantics changed. Cut 5 remains ACTIVE pending rerun of `showcase.sh` and the manual attach/deposit/run/CTRL-T/CTRL-E/examine gate.

### 2026-08-19 — V6 emulator Cut 5 observer spawn repair 2
- Local showcase startup exposed `undefined: VIEW_ALIVE` in `observe/broker.ai` after the import-isolation repair.
- Cause: spawned computations also do not capture module-level constants. `sender` still referenced `VIEW_VERSION`, `VIEW_TEXT`, and `VIEW_ALIVE` from its parent module.
- Repair: localize the observer-state indices inside `sender`; all current `spawn(...)` sites were rescanned for outer lexical dependencies.
- No PDP, UNIBUS, tape, monitor-command, or observer-topology semantics changed.


### 2026-08-19 — V6 emulator Cut 5 observer usability/state repair

- Manual bootstrap through `aiki-pdp>` is operational: real tape attach, six octal deposits, `run 100000`, repeated CTRL-T status while the terminal branch spins, CTRL-E return to the monitor, and `examine 000000 2` yielding `000407` / `000654`. CPU and tape observer windows display the corresponding machine/device state.
- User observation exposed two presentation/state issues before Cut 5 gating: after CTRL-E the CPU observer reported `HALTED`, and continuous observer redraw made terminal mouse selection impractical while the guest was running.
- Repair adds explicit machine `suspended` state. CTRL-E and foreground interruption set it; `run`/resume clears it. Observer and monitor status render `SUSPENDED` distinctly from guest architectural `HALTED`.
- Observer publication remains outside guest semantics and is now bounded to 500 ms while running. The broker does not increment a view version for identical text. CPU step/time counters are omitted while RUNNING and appear once stopped/suspended, allowing stable loops such as `BR 100012` to stop repainting after their architectural snapshot settles. Tape/UNIBUS views likewise do not repaint unchanged snapshots.
- Reconciled the prior ESC printable-key grouping repair: `(key >= 32) and (key <= 126)`. ESC at the main monitor is ignored and cannot terminate the monitor or observers.
- Static evidence: `git diff --check` clean; all 35 Experiment 004 `.ai` files lex/parse under the authoritative grammar in the disposable Go-1.23 syntax harness; `go test ./engine/syntax/...` passes there. Cut 5 remains ACTIVE pending user rerun of `cut5_monitor_test.ai` and manual showcase confirmation of SUSPENDED labeling plus mouse-selectable stable observer windows.


### 2026-08-19 — V6 emulator Cut 5 live observers, counters, restartable suspension, and tape boot

- User preferred the original visibly-live observer feel, but optimized rather than per-instruction redraw. Cut 5 now publishes at a bounded 100 ms cadence while preserving latest-snapshot coalescing; observer I/O remains unable to pace machine execution.
- CPU/debug again displays live step/time counters. Tape view adds selected unit plus cumulative reads, writes, bytes-in, and bytes-out; UNIBUS view adds cumulative DATI, DATO, DATOB, and NPR-write counters. Numeric presentation remains octal-first.
- CTRL-E suspension is explicitly restartable with `continue`, which resumes at the current PC and clears SUSPENDED without resetting machine/device state. Bare `run` remains capable of running from the current PC as before.
- Tape attachment is now unit-aware for TM11 units 0-7: `attach tape [UNIT] FILE`, with legacy `attach tape FILE` retaining unit-0 shorthand. TM11 media state is per unit under one controller.
- Added `boot tape UNIT`. It deposits the six-word V6 TM bootstrap at `100000` and uses the controller's real bits 10-8 Unit Select field; unit 0 encodes command `060003`, unit 1 `060403`, etc., then starts execution at `100000`.
- Focused diagnostics were extended to verify unit-select encoding, selected-unit raw-media transfer, tape/UNIBUS counters, and monitor help. Static evidence: `git diff --check` clean; all Experiment 004 Aiki sources lex/parse under the authoritative grammar; disposable `engine/syntax` tests pass. Runtime gate remains user-side.
- Exact next evidence: run `cut5_monitor_test.ai`; start `showcase.sh`; `attach tape 0 /path/to/v6.tape`; `boot tape 0`; observe live CPU/tape/UNIBUS counters; CTRL-E -> SUSPENDED; `continue` -> RUNNING from the same PC; CTRL-E again; `examine 000000 2` must yield `000407` / `000654`.
### 2026-08-19 — Cut 5 unit-aware tape test repair
- Local validation of the live-observer/unit-aware tape update produced 39/43 passes; all four failures came from the same assertion in `cut5_monitor_test.ai`.
- Cause: the diagnostic used `equal(shape(result), :pdp_fault)` while `equal` was the test assertion helper, and Aiki shaped lists retain base type `:list`; the main monitor also incorrectly used `shape(...)` to recognize `@pdp_fault`.
- Repair: add `monitor.is_pdp_fault(...)` using shaped-pattern matching and use that predicate in both the interactive monitor and the focused diagnostic.
- No PDP-11, TM11, UNIBUS, boot, or counter semantics changed. Cut 5 remains ACTIVE pending local rerun.

### 2026-08-19 — Cut 5 gated
- User reran `aiki test experiments/004-v6-emulator/experiment/diagnostics/cut5_monitor_test.ai` after the shaped-fault recognition repair and reported the gate passed.
- Cut 5 is GATED. The accepted monitor contract includes `aiki-pdp>`, unit-aware tape attach and `boot tape N`, CTRL-T status, restartable CTRL-E suspension via `continue`, distinct SUSPENDED/HALTED states, live/coalesced CPU/UNIBUS/tape observers, and cumulative tape/UNIBUS activity counters.
- Next cut is KL11 console support and execution of the real standalone image from address `0` to the first `=` prompt. Observer architecture expands with a separate console window; disk remains out of scope until the following gate.


### 2026-08-19 — Cut 6 KL11 console begins
- Cut 5 is GATED from user validation. Cut 6 is ACTIVE and targets the first real standalone `=` prompt from the V6 tape-loaded image before any RK11/RK05 work.
- Added a KL11 console device at physical UNIBUS addresses `777560`/`777562`/`777564`/`777566`, corresponding to CPU-visible `177560`..`177566` while memory management is disabled. Transmitter READY is set after INIT; TPB writes clear READY, advance restores it and makes the character available to the host console; receiver injection sets DONE and reading TKB clears DONE.
- Added a dedicated read-only KL11 observer window with register state, DONE/READY and interrupt-enable state, cumulative RX/TX counters, and last characters. Observer isolation is unchanged.
- While the guest runs, CTRL-T and CTRL-E remain emulator-private. All other input bytes, including CTRL-C and CTRL-D, now enter the KL11 receiver; at `aiki-pdp>` CTRL-C/CTRL-D retain monitor-terminal behavior.
- Added TM11 REWIND function `111`, required by the standalone startup path described by the Lions laboratory.
- Static evidence: the complete Experiment 004 Aiki tree parses under the authoritative grammar. Runtime evidence remains user-side. Exact gate: `cut6_console_test.ai`, then real `boot tape 0`, suspend at `100012`, `run 0`, and observe the first standalone `=` through KL11.

### Cut 6 repair — observer port allocation

- Runtime validation found a stale prior broker could retain fixed port 41140 and prevent a new showcase session from starting.
- `showcase.sh` no longer defaults to 41140. Unless `AIKI_PDP_PORT` is explicitly supplied, it selects a free localhost TCP port before launching the monitor and observer windows, then exports that single selected port to all participants.
- Explicit `AIKI_PDP_PORT` remains supported for deterministic/manual runs.
- No PDP-11, KL11, TM11, UNIBUS, or monitor-command semantics changed.


### 2026-08-19 — Cut 6 gated

- User ran the real TUHS V6 tape through `aiki-pdp`: `boot tape 0`, CTRL-E at the six-word bootstrap loop, then `run 0`. The standalone image printed its real `=` prompt through the emulated KL11.
- At suspension in the standalone command loop: PC `137300`, PSW `000004`, 41036 steps; tape was rewound to record `000000`, with exactly one read and `1000` octal bytes transferred by NPR. KL11 showed transmitter READY and console polling; UNIBUS NPR writes remained fixed at `1000`, confirming no continuing tape DMA.
- Cut 6 is GATED. The accepted contract is: real V6 record-zero bootstrap -> TM11 rewind -> KL11 interactive standalone prompt, with separate CPU/UNIBUS/Tape/KL11 observer windows and dynamically allocated observer port.
- Next cut: RK11/RK05 with one controller and 8 attachable drive slots, topology visible in observers, sufficient real register/NPR behavior for standalone `tmrk` to construct disk 0. Stop before booting the constructed disk.


### 2026-08-19 — Cut 7 RK11/RK05 and `tmrk` begins

- Cut 6 is GATED. Cut 7 is ACTIVE and targets V6 standalone `tmrk` constructing disk 0 from the real TUHS distribution tape; booting the resulting disk is explicitly the next gate, not this cut.
- Added one RK11 controller with 8 attachable RK05 slots. RK05 media is a separate host-file abstraction: 4800 sectors, 512 bytes/sector, 2 surfaces, 200 cylinders, 12 sectors/track. Existing packs must have exact RK05 size; absent paths are created as blank writable packs.
- RK11 exposes RKDS/RKER/RKCS/RKWC/RKBA/RKDA/RKDB at the historical UNIBUS addresses, unit selection through RKDA bits 15-13, two's-complement word count, bus-address increment, sector-boundary RKDA increment, and bounded READ/WRITE functions through UNIBUS NPR.
- TM11 now implements SPACE FORWARD using MTBRC as a two's-complement record count, required for `tmrk` tape offsets 100/101.
- Observer topology is explicit: Tape shows all 8 tape-unit attachments; RK shows all 8 disk slots plus controller registers/counters; UNIBUS lists connected tape/disk media and adds NPR-read counting. `show rk` and a dedicated RK observer window are wired into `showcase.sh`.
- Focused diagnostic `cut7_rk_test.ai` covers 8-drive attachment, exact pack size, one-sector WRITE/READ round trip through NPR, TM11 space-forward, and topology views.
- Static evidence: all 40 Experiment 004 Aiki files lex/parse under the authoritative grammar; `go test ./engine/syntax/...` passes in the disposable Go-1.23 compatibility tree; `git diff --check` and `bash -n showcase.sh` are clean. Runtime gate remains user-side.
- Exact next evidence: run `cut7_rk_test.ai`; then `showcase.sh`, attach real tape unit 0 and writable disk 0, boot tape, run standalone to `=`, run `tmrk` with disk/tape/count `0/100/1` and `1/101/3999`. Gate only when V6's own standalone program completes both transfers and the RK observer/counters agree.

### 2026-08-19 — Cut 7 attachment policy tightened

Before running the historical `tmrk` gate, attachment semantics were made explicit and testable. V6 raw tape is immutable input media: attach opens an existing exact-size image read-only, and a missing tape path must fail without creating a file. RK05 media is writable removable storage: attaching a missing disk path creates a blank 2,457,600-byte RK05 pack, while an existing pack of the wrong size is rejected without resizing or truncating it. These are emulator attachment-policy invariants, not guest-visible device semantics.

### 2026-08-19 — Cut 7 standalone workload extends CPU contract with NEG(B)

- Real standalone `tmrk` execution reached `=` and then faulted at PC `137236` on instruction `005400`; the monitor reported the instruction payload as decimal `2816`. The PDP-11/40 handbook identifies `005400` as `NEG R0`.
- This is executable evidence that the earlier `m40.s` instruction contract was complete for that kernel source workload but not for the V6 standalone installation program. Historical execution remains the authority for extending the emulator surface.
- Added PDP-11/40 `NEG`/`NEGB` decode and execution. NZVC follows the handbook: N/Z from result; V set for the most-negative result (`100000` word / `200` byte); C cleared only for zero result and set otherwise.
- Added `cut7_cpu_extension_test.ai` with word, most-negative, zero, byte, and reserved-instruction formatting evidence.
- Reserved-instruction monitor output now renders the faulting PC and IR in octal (`PC ...`, `IR ...`) rather than exposing the Aiki decimal payload.
- No RK11, RK05, TM11, UNIBUS, media, or transfer semantics changed. Cut 7 remains ACTIVE pending local execution of the focused diagnostic and rerun of real `tmrk`.

### 2026-08-19 — Cut 7 standalone fault probe and demo convenience

- User validation of `cut7_cpu_extension_test.ai` passed 12/13; the only failure was the reserved-instruction formatting probe itself, which reached an older unparenthesized decoder range predicate and faulted with Aiki's left-to-right `cannot compare boolean and number` behavior before a PDP reserved-instruction value could be returned.
- Corrected the remaining comparison-range expressions of the same class in Experiment 004 by grouping each comparison explicitly. This is an Aiki evaluation-order repair, not a PDP semantic change.
- Added `demo V6_TAPE RK0` as a convenience command over the normal monitor operations. It attaches tape unit 0 read-only, attaches or creates writable disk unit 0, deposits the historical V6 TM bootstrap, and starts at `100000`. It does not bypass the machine or standalone program. The operator still performs CTRL-E, `run 0`, and `tmrk` manually.
- Demo setup is ordered tape-first so an invalid/missing tape cannot create a disk as a side effect. Focused diagnostic evidence checks this failure policy; successful tape boot remains covered by the existing `boot tape` and historical raw-media gates.
- Cut 7 remains ACTIVE. Exact next runtime evidence: rerun `cut7_cpu_extension_test.ai` and `cut7_rk_test.ai`; then use `demo <v6.tape> <rk0>`, CTRL-E, `run 0`, and retry real standalone `tmrk`.

### 2026-08-19 — Cut 7 performance profiling gate added

- Real standalone `tmrk` progressed correctly but exposed objectionable emulator throughput while executing the finite `CLR (R0)+ / CMP R0,R6 / BLO` memory-clear loop. CPU state showed genuine forward progress and eventually reached the `disk offset` prompt; this is a performance finding, not a correctness failure.
- Optimization is paused pending measurement. Added a CPU-only deterministic benchmark using the exact three guest instructions and no tape, disk, KL11, or observer broker.
- The performance sweep runs independent profiler processes at 1x, 2x, 10x, 50x, and 100x. One x is 256 loop iterations / 768 guest instructions; 100x is 25,600 iterations / 76,800 guest instructions, deliberately near the observed standalone workload scale.
- Each stage uses `aiki profile --counts`, preserving independent semantic counts plus elapsed/allocation/malloc/GC realization measurements. Results are written beneath `experiments/004-v6-emulator/results/cut7-perf/` and are not to be interpreted until all five stages are available.
- Code inspection before measurement identifies possible hot-path multipliers (ordinary RAM currently traverses the machine/UNIBUS path; mutable machine and RAM state use `store`), but these remain hypotheses. Do not optimize from inspection alone.
- Development-container limitation remains: no current `aiki` executable and Go 1.24 cannot be downloaded, so the profiling sweep must run user-side. Exact next action: run `experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_sweep.sh` and review the five profiles before any performance edit.

### 2026-08-19 — Cut 7 accelerated realization profile

- User-side staged profiling of the exact standalone `CLR (R0)+ / CMP R0,R6 / BLO` loop established a nearly linear but extreme fixed cost. At 10x (7,680 guest instructions), elapsed time was 35.88s with ~18.5M Aiki calls, ~14.25M arithmetic operations, ~7.58M comparisons, ~5.14M iterations, ~17.9GB allocated, and ~448M Go mallocs. The evidence points to per-instruction semantic realization cost rather than scale-dependent accumulation.
- Inspection found the dominant candidate: Experiment 004 used pure-Aiki `bits/native` throughout decode, execution, PSW, width, memory, audit, and devices. This is architecturally valid but unsuitable as the selected realization for a bit-oriented CPU emulator hot path. `store` is already provider-backed through `_store_*`; there is no `store/ffi` package and none is invented here.
- Experiment 004 now explicitly selects `bits/ffi` and `bytes/ffi` throughout. The pure native libraries and native/FFI parity architecture are unchanged. This is an explicit workload realization choice, not a semantic shortcut or boundary leak.
- The accelerated staged sweep writes to `experiments/004-v6-emulator/results/cut7-perf-ffi/`, preserving the prior `cut7-perf/` native baseline for direct 1x/2x/10x/50x/100x comparison.
- Static evidence in the disposable Go-1.23 harness: all 42 Experiment 004 Aiki files parse; `go test ./engine/syntax/...` passes; performance runner shell syntax and `git diff --check` are clean. Authoritative Go 1.24/runtime gates remain user-side because the sandbox cannot download the required toolchain.
- Exact next evidence: rerun `experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_sweep.sh` locally and compare semantic counts, elapsed, allocation, and malloc rates against the preserved native baseline before considering any further hot-path changes.

### 2026-08-19 — Cut 7 performance repair: remove linear audit tax

- FFI profiling remained nearly linear at 50x: 38,400 guest instructions in 58.25s, with ~36.84M Aiki calls, ~1.78M iterations, ~561M Go mallocs, and ~27.1GB allocated. Per-instruction semantic ratios were stable, confirming fixed hot-path cost rather than accumulating state.
- One-step source tracing found that the measured ~46 Aiki iterations per guest instruction were almost exactly explained by audit implementation: `record_instruction` linearly searched the 62-entry instruction manifest, and `record_psw_transition` scanned all 16 PSW bits after every instruction. The benchmark instruction mix (`CLR`, `CMP`, `BLO`) makes the expected instruction-search average plus the 16-bit scan line up with the profiler count.
- Audit instruction lookup is now constant-time by instruction symbol. `INSTRUCTIONS` remains the factual vocabulary authority; `validate_instruction_index()` checks the derived fast map against manifest order and is covered by the Cut 7 CPU-extension diagnostic.
- PSW auditing still stores full before/after/mask data and all 16 per-bit counters. The normal NZVC path checks bits 0-3 directly; upper bits 4-15 are scanned only when the transition mask actually changes an upper PSW bit. No audit information is discarded.
- UNIBUS/device advancement, operand shapes, machine/state representation, and memory routing are intentionally unchanged in this cut. Exact next evidence is the same staged FFI performance sweep, written separately beneath `results/cut7-perf-ffi-audit/`; only after measuring this audit repair should the next fixed-cost layer be changed.

### 2026-08-19 — Cut 7 machine-shaped execution refactor ACTIVE

- Post-FFI/post-audit profiling remained linear but expensive: at 50x, 38,400 guest instructions took 36.87s with ~26.5M Aiki calls, ~387M Go mallocs, and ~19.6GB allocated. The remaining cost is fixed representation/dispatch churn rather than scale-dependent accumulation.
- Architectural refinement: the emulator execution representation now follows the machine more directly. CPU state is one compact store; physical memory is word-backed while preserving byte-addressed PDP semantics; addressing operands are compact internal tuples; instruction decode is classified once per instruction.
- UNIBUS now has an ordinary-RAM fast path and a configuration-time I/O ownership table built from controller-declared backplane address ranges. Runtime RAM accesses do not poll TM11/KL11/RK11 `handles` functions. I/O accesses select the configured controller.
- Controller/device ownership is explicit: TM11 owns controller registers/commands and references TU10 devices; TU10 owns tape attachment/media/position. RK11 owns controller registers/commands/NPR and references RK05 drive devices; each drive owns its pack attachment/media state.
- Idle devices no longer receive unconditional per-instruction `advance` calls. UNIBUS maintains one pending-work flag; only active controllers are serviced. Machine state remains single-authority and deterministic; concurrency remains at host input/observer/media-control boundaries rather than allowing concurrent mutation of architectural state.
- Observer lifecycle repair: `show cpu|unibus|tape|rk|kl11` remains an in-monitor snapshot. `observe TYPE` (alias `open TYPE`) invokes the same reusable terminal launcher used by `showcase.sh`, allowing a closed observer window to be spawned again without affecting the machine.
- New focused diagnostic `cut7_architecture_test.ai` checks word-backed byte semantics, CPU I/O-page projection, and configured KL11 routing. The staged performance sweep now writes to `results/cut7-perf-machine-architecture/` so native, FFI, audit, and architecture measurements remain separate.
- Static evidence in a disposable Go-1.23 syntax harness: all 39 Experiment 004 Aiki files parse; `engine/syntax` parser tests pass; `bash -n` passes for `showcase.sh`, `observe/open.sh`, and the performance runner. Full authoritative runtime/Go-1.24 validation remains user-side because this environment cannot obtain the required Go 1.24 toolchain/dependencies.
- Exact next evidence: run `cut7_architecture_test.ai`, `cut7_cpu_extension_test.ai`, `cut7_rk_test.ai`, then the same staged `cut7_perf_sweep.sh`. If semantics remain green, compare 1x/10x/50x against the FFI+audit baseline before returning to the real `demo` -> CTRL-E -> `run 0` -> `tmrk` gate.

### 2026-08-19 — Cut 7 explicit fixed-width machine domain ACTIVE

- Profiling after the machine-shaped refactor still showed ~607 Aiki calls, ~18 store reads, ~14 store writes, and ~8,815 Go mallocs per guest instruction at 50x. Inspection showed that the remaining hot path repeatedly represented PDP words/registers/RAM cells as general Aiki exact rationals / `value.Value` store cells.
- A proposed transparent "small/fast Number" realization was considered and rejected before integration. It is SUPERSEDED: ordinary Aiki `number` remains exact rational with no hidden fixed-width/int/float representation contract. Machine-width behavior is an explicit imported capability, not a language numeric semantic.
- Accepted boundary: `machine/ffi` provides opaque fixed-width machine values and operations. Private provider kinds distinguish byte, 16-bit word, and 18-bit physical address, but `inspect` exposes only `<opaque>` and no `:word`, `:byte`, or `:addr18` language types are introduced. Conversion to/from exact Aiki `number` is explicit.
- `store/ffi` is a typed mutable machine-storage capability, not an accelerated arbitrary Aiki store. Word backing is `[]uint16`, byte backing is `[]uint8`, and 18-bit address backing is bounded host storage; cells carry only matching opaque machine values. Ordinary `store` remains unchanged for heterogeneous Aiki values and emulator metadata.
- Current serial cut migrates the CPU/RAM hot path: R0-R7, PSW, fetched instructions, effective addresses, operand values, and physical RAM remain opaque fixed-width values through ordinary execution. PDP RAM is a 16-bit word store addressed by opaque 18-bit byte addresses; byte access selects/updates a lane without rationalizing the word.
- Existing diagnostic/control APIs remain exact-number-facing (`deposit`, `examine`, register/PC/PSW access, standalone addressing/decode diagnostics). Execution-facing APIs use explicit `*_machine` / `*_word` forms so boundary crossings are visible in source and do not occur in the hot path.
- PSW auditing was also corrected to remain fixed-width during execution. The audit stores opaque before/after/mask values and counts NZVC changes with machine-domain predicates; exact-number conversion occurs only when the audit report is requested.
- Controllers are an explicitly retained compatibility island in this cut: TM11/KL11/RK11 still expose their legacy numeric register surfaces at the I/O-page boundary. Ordinary RAM and CPU execution do not cross that boundary. Controller-register migration is the next bounded representation cut after profiling this CPU/RAM change.
- EIS `MUL/DIV/ASH/ASHC` remains a documented compatibility island using exact-number arithmetic internally for its wider signed calculations. The common V6 `CLR/CMP/BLO` performance path and ordinary PDP word/byte instructions are fixed-domain end-to-end.
- Compatibility/runtime evidence in the disposable Go-1.23 harness: focused `machine/ffi` / `store/ffi` substrate tests pass; `engine/semantics/value`, runtime module policy, and primitive catalog pass. Full substrate has only the known fake-`flock` contention failure caused by the local stub, unrelated to this cut.
- A disposable locally built Aiki runtime passes every Experiment 004 diagnostic: Cut 1 42/42, Cut 2 54/54, Cut 3 73/73 + m40 114/114, Cut 4 31/31 + trace 4/4, Cut 5 39/39, Cut 6 24/24, Cut 7 architecture 10/10, CPU extension 16/16, fixed-domain 12/12, RK 47/47, and help 16/16. All 46 Experiment 004 `.ai` files parse under the authoritative grammar.
- Non-authoritative local 1x semantic profile after the fixed-domain + opaque-PSW-audit change: arithmetic 9,627; comparison 13,173; calls 272,400; iteration 999; index 34,789; store reads 8,713; store writes 8,935. The prior machine-architecture 1x baseline was arithmetic 34,465; comparison 39,306; calls 467,537; index 22,991; store reads 14,091; store writes 10,996. Wall-clock from the Go-1.23/stub executable is not comparable to the user's Go-1.24 runtime and is not acceptance evidence.
- Exact next evidence: user runs the unchanged scaling workload, now writing beneath `results/cut7-perf-fixed-domain/`, and compares 1x/2x/10x/50x against the prior machine-architecture baseline before controller migration or returning to real `tmrk`.

### 2026-08-19 — Cut 7 fixed-domain realization: intern finite machine values

- User-side fixed-domain profiling at 50x (38,400 guest instructions) measured 19.31s, ~13.54M Aiki calls, ~174.3M Go mallocs, and ~9.97GB allocated. Semantic work was substantially lower than the prior machine-architecture baseline, but the substrate still allocated ~4,540 objects per guest instruction.
- Inspection found a representation-level cause below the Aiki semantics: every `machine/ffi` operation and `store/ffi.get` constructed a fresh `*value.Opaque`, even though the private machine domains are finite (256 byte values, 65,536 word values, 262,144 18-bit physical addresses).
- The provider now interns all three finite machine domains at process initialization. Machine operations and fixed-store reads return canonical opaque values; equal fixed-width results therefore require no per-result heap allocation. The Aiki boundary is unchanged: values remain opaque, `number` remains exact rational, and no `:word`/`:byte`/`:addr18` language types are introduced.
- Focused substrate invariant verifies canonical identity for equal byte/word/address results and for fixed-store round trips. Disposable Go-1.23 compatibility evidence with local readline/flock stubs: focused `machine/ffi` / `store/ffi` tests pass. Authoritative Go-1.24 runtime evidence remains user-side.
- This cut intentionally does not change the remaining Aiki list/index structure. Exact next evidence is the unchanged Cut-7 staged profile; compare malloc/allocation reduction first, then continue with the separately identified operand/machine-state indexing tax.

### 2026-08-19 — Cut 7 structural hot-path binding ACTIVE

- User-side interning measurements showed fixed-width opaque-value interning was not the dominant remaining allocation cost: the 10x benchmark improved only from 4.430s to 4.136s, with mallocs and allocated bytes changing by roughly 0.3%; semantic counts were identical. Heap boxing of machine values is therefore ruled out as the principal remaining cost.
- The next measured target is Aiki structural churn: the fixed-domain 50x profile still executes about 353 Aiki calls and 45 index operations per guest instruction. Inspection found repeated rediscovery of `machine -> processor state / UNIBUS / audit` and construction/indexing of operand descriptor lists in the common CPU path.
- `run_bounded` now binds processor state, UNIBUS, and audit once for the run. Fetch works directly against bound state + bus. `CLR/CLRB`, `CMP/CMPB`, and all branch instructions use the bound path; other instruction families fall back to the proven general executor after the already-fetched instruction, preserving serial migration rather than duplicating the whole CPU.
- Addressing now provides direct bound read/write-by-specification operations. The common CLR/CMP path performs PDP addressing-mode side effects and the actual register/memory access without allocating `[kind, register, address]` operand descriptors. Existing descriptor-based APIs remain for diagnostics and instruction families not yet migrated.
- Added `cut7_structural_hotpath_test.ai`, which compares twelve instructions of the exact `CLR (R0)+ / CMP R0,R6 / BLO` loop under legacy `step` and bound `run_bounded`, including registers, PC, PSW, memory, instruction/mode audit counts, step count, and machine time.
- Performance evidence for this realization is isolated beneath `results/cut7-perf-structural-hotpath/`. Exact next evidence: run the focused structural test and existing Cut 7 semantic diagnostics, then profile 1x/10x. The acceptance signal is a material reduction in `index` and `call`; store/counter policy is intentionally unchanged in this serial cut.
- Disposable-runtime semantic evidence after the structural hot-path change: 1x calls fell from 272,400 to 228,369 and indexes from 34,789 to 19,687; 10x calls were 2,269,713 and indexes 194,791. Store reads/writes remained exactly at the fixed-domain baseline (1x 8,713/8,935; 10x 87,049/87,271), confirming that this serial cut removed list/machine traversal without changing scalar-store policy. Wall-clock/allocation figures from the Go-1.23 stub runtime are not comparable to the user's authoritative build.
- Disposable semantic gate: `cut7_structural_hotpath_test.ai` 27/27; fixed-domain 12/12; architecture 10/10; CPU extension 16/16; RK 47/47. Full Experiment 004 parse count is 47 Aiki files. User-side 1x/10x profiling remains the acceptance evidence for actual elapsed/allocation behavior.

### 2026-08-19 — Cut 7 execution-bookkeeping store ledger and counter blocks

- User-side structural profiling left the exact `CLR (R0)+ / CMP R0,R6 / BLO` workload at 10x with 87,049 general-store reads and 87,271 writes. Static accounting reconciled those operations almost entirely to emulator bookkeeping rather than PDP machine storage: UNIBUS transaction counters, running/waiting flags, steps/time, audit instruction/mode/PSW counters, last-PSW capture, and the pending-device flag.
- This establishes a representation boundary: PDP RAM/register words remain on typed fixed-width `store/ffi`; heterogeneous configuration/controller metadata may remain on ordinary `store`; hot scalar execution bookkeeping must not perform `store.get -> Aiki arithmetic -> store.set` on every guest instruction.
- `store/ffi` now provides explicit provider-backed non-negative counter blocks with `new_counter`, `counter_get`, `counter_set`, and `counter_add`. `counter_add` performs the increment within the provider so the hot path does not materialize a get/add/set cycle in Aiki. This does not add an Aiki integer type and does not alter exact-rational `number` semantics.
- CPU running/waiting/suspended flags and steps/time, UNIBUS DATI/DATO/DATOB/NPR counters and pending-work flag, and audit instruction/mode/PSW counters now use counter blocks. Audit last-PSW before/after/mask remains fixed-width word storage; the last instruction kind is stored as a counter-coded manifest index and translated back to the symbol only at observation.
- Ordinary `store` remains in the UNIBUS configured I/O-route table because that structure is cold configuration/dispatch state; normal RAM traffic does not consult it. No controller/device semantics changed.
- Disposable runtime semantic evidence: structural hot-path 27/27, fixed-domain 16/16, architecture 10/10, CPU extension 16/16, RK 47/47. In the local 10x semantic profile, general-store traffic collapsed from 87,049/87,271 reads/writes to 9/117, identical at 1x and 10x and therefore initialization-only. Calls fell from 2,269,713 to 2,108,410 and arithmetic from 92,571 to 43,928. Wall-clock from the Go-1.23/stub runtime is not comparable to the user's authoritative Go-1.24 build.
- Exact next evidence: rebuild locally (Go/runtime changed), rerun the focused fixed-domain/structural/RK diagnostics, then profile 1x/10x. Acceptance signal is near-flat general-store counts across scale plus preserved visible audit/UNIBUS/steps results.

### 2026-08-19 — Cut 7 instruction-path flattening

- User-side bookkeeping profiling left the 10x `CLR (R0)+ / CMP R0,R6 / BLO` workload at 2,108,410 Aiki calls, 176,869 indexes, and initialization-only general-store traffic (9 reads / 117 writes). The next serial target was therefore the common CPU execution call graph rather than storage representation or bookkeeping.
- The bound hot path now decodes the instruction fields actually consumed by the modeled hardware once. `CMP/CMPB` extract source/destination mode/register fields directly from the fetched instruction and pass those fields to addressing; `CLR/CLRB` do the same for their destination. Addressing exposes `read_decoded_bound` / `write_decoded_bound` so the hot path no longer constructs an operand specification only to split it back into mode/register immediately.
- Branches now classify directly from the branch opcode byte and evaluate the condition from PSW bits without routing through the general instruction decoder and then a second symbol-to-condition dispatch. CLR/CMP NZVC updates on the bound path write the four fixed-width PSW bits directly rather than traversing the generic PSW helper chain.
- The general decoder/executor remains the compatibility path for instruction families not yet migrated; diagnostic/control APIs are unchanged. No PDP semantics, fixed-width representation, store/counter policy, controller behavior, or observer behavior changed.
- Disposable semantic evidence: `cut7_structural_hotpath_test.ai` 27/27; fixed-domain 16/16; architecture 10/10; CPU extension 16/16; RK 47/47; all 47 Experiment 004 Aiki files parse under the authoritative grammar.
- Disposable semantic profile at 10x: calls 1,706,490 (from 2,108,410, ~19% lower), arithmetic 20,888 (from 43,928, ~52% lower), comparisons 105,327 (from 128,367, ~18% lower); indexes remained 176,869 and general-store traffic remained 9/117. Wall-clock/allocation figures from the Go-1.23/stub runtime are not comparable to the user's authoritative build.
- Exact next evidence: user reruns 1x/10x on the normal Go-1.24 build. If semantic counts match, profile/trace the remaining 176,869 index operations separately; this cut deliberately does not combine that next structural target.

### 2026-08-19 — Cut 7 source-attributed component binding

- User-side post-flattening profiling left the 10x `CLR (R0)+ / CMP R0,R6 / BLO` workload at 1,706,490 Aiki calls and 176,869 index operations, with ordinary-store traffic already initialization-only (9 reads / 117 writes). The next serial target was therefore source-attributed structural indexing rather than further representation or store changes.
- Added `cut7_perf_sites.ai`, which runs the exact benchmark body under `profile.measure` and reports the hottest `:call` and `:index` source sites. Setup/validation occur outside the measurement interval. This uses the existing profiling source-attribution contract rather than adding runtime instrumentation.
- The 1x attribution identified repeated component rediscovery as the dominant index source: `state.words` 6,398 indexes, `state.flags` 2,308, `state.counts` 1,536, `unibus.physical_memory` 1,024, `unibus.counts` 1,024, and `unibus.pending` 768. These six sites accounted for 13,058 of the prior 19,687 local 1x indexes (~66%). Audit store selection was the next remaining index class.
- `run_bounded` now resolves the CPU word/flag/count stores and UNIBUS memory/count/pending stores once per run. Bound fetch/addressing/CLR/CMP/branch paths receive those owning stores directly. Ordinary RAM accesses use a bound UNIBUS fast path; I/O-page accesses still fall back to the configured controller-routing authority. No controller, device, audit, number, machine-value, or store representation changed.
- Added bound state operations for register/PC/PSW/flags/completion and bound UNIBUS read/write/advance operations. These are execution-only APIs; existing diagnostic/control facades remain unchanged.
- Disposable semantic gate: structural hot path 27/27, fixed-domain 16/16, architecture 10/10, CPU extension 16/16, RK 47/47.
- Disposable semantic profile after the component-binding cut: 1x calls 136,979 (from 172,026), comparisons 3,699 (from 10,863), indexes 4,843 (from 17,893), store 9/117 unchanged. At 10x: calls 1,355,795, comparisons 33,651, indexes 46,315, store 9/117. Wall-clock/allocation from the Go-1.23 stub runtime remains non-authoritative.
- Exact next evidence: user runs 1x/10x on the normal Go-1.24 build. The next source-attributed target is now audit store selection plus the remaining call-heavy fault/equality/PSW machinery; do not fold that next cut into this one before user-side evidence.

### 2026-08-19 — Store intent preserved; audit components bound

- User restated the intended Aiki `store` abstraction: `store` is isolated mutable mapped memory, meant to be extremely cheap and addressed through the existing functional access surface. It is not intended to become a generic object container or to pay collection-style overhead on every access. If isolation guarantees single authority, per-access mutex synchronization is redundant; synchronization belongs at the isolation/ownership boundary. This is a durable Aiki runtime follow-on and is intentionally not mixed into the current PDP serial performance cuts unless profiling directly implicates `store` again.
- After source-attributed CPU/UNIBUS component binding, the next remaining index class was audit aggregate selection. `run_bounded` now resolves the six audit stores once (instruction counts, source/destination modes, PSW-change counts, last-kind, last-PSW words) and passes the owning stores directly to bound audit operations.
- Audit semantics are unchanged: instruction/mode counts, all 16 PSW-change counters, and full last before/after/mask data remain available. The general diagnostic/control audit API remains intact; only the execution path uses the bound forms.
- Disposable semantic gate remains green: structural hot path 27/27, fixed-domain 16/16, architecture 10/10, CPU extension 16/16, RK 47/47.
- Disposable semantic profile after audit binding: index collapsed from 4,843 at 1x / 46,315 at 10x to 246 total at both scales, proving essentially all per-instruction list indexing has been removed. Calls and general-store counts are unchanged; the next source-attributed target is call-heavy fault recognition and PSW/audit machinery.

### 2026-08-19 — Cut 7 direct shaped-fault recognition

- After audit components were bound, source attribution showed repeated PDP fault recognition as the next avoidable call chain: `is_fault` called `shape` and `equal` several times per guest instruction even though PDP faults have one exact shaped-list form.
- `execute.ai` now recognizes `[@pdp_fault, kind, detail]` directly with Aiki pattern matching. No PDP-specific substrate primitive was added; fault representation and propagation remain entirely in Aiki.
- Disposable semantic gate remains green: structural hot path 27/27, fixed-domain 16/16, architecture 10/10, CPU extension 16/16, RK 47/47.
- Disposable 10x semantic profile: calls fell from 1,355,801 to 1,232,921 (~9%); index remained 246 total and general-store traffic remained 9/117. Next attribution target is the remaining PSW/audit/machine-FFI call surface.

### 2026-08-19 — Cut 7 call-site control cleanup

- After audit binding and direct shaped-fault matching, source attribution showed a remaining class of call overhead caused by control tests expressed through `equal(...)` rather than by machine work: CPU running/waiting flags, audit source/destination role selection, fetch fault recognition, idle-device pending checks, and hot opcode classification.
- Hot audit mode recording now has direct source/destination entry points, avoiding role-symbol comparisons in the execution path. Fetch uses shaped-pattern recognition for PDP faults. Counter flags use direct numeric comparisons. Bound opcode classification uses `match` dispatch rather than repeated `equal` calls.
- No machine representation, store policy, controller/device semantics, audit data, or fault representation changed. This is an Aiki control-flow simplification only.
- Disposable semantic gate remains green: structural hot path 27/27, fixed-domain 16/16, architecture 10/10, CPU extension 16/16, RK 47/47.
- Disposable 10x semantic profile after this cleanup: calls 1,094,677 (from 1,232,921 after direct fault matching, and 1,355,801 before audit/fault cleanup); index remains 246 total; general-store traffic remains 9/117. Comparison semantic counts rise because former `equal(...)` function calls are now direct comparisons/match dispatch; this is expected and intentionally trades function overhead for primitive comparison work.


### 2026-08-20 — Cut 7 PSW-audit no-change fast path ACTIVE

- Source attribution after call cleanup left PSW/audit machine-FFI calls as the next measured hot-path class. Branch instructions do not alter the PDP-11 PSW, yet the audit path still performed XOR + per-bit delta inspection and rewrote the full transition record on every branch.
- Added an Aiki-level `record_psw_unchanged_bound` path. Known PSW-preserving instructions can now preserve the exact last-transition report (`before == after`, mask zero, last kind) without invoking PSW delta analysis. No PDP semantics or audit observability changed.
- The exact `CLR/CMP/BLO` benchmark uses this path for BLO only; CLR/CMP continue through full PSW-delta auditing. This is intentionally a serial, source-attributed call-reduction cut.
- Separate durable runtime follow-on remains unchanged: general `store` is intended as isolated mutable mapped memory; per-access mutex synchronization should be removed if isolation already guarantees single authority.

### 2026-08-20 — Cut 7 live bounded execution and zero-wrapper FFI

- Source attribution showed that simple FFI modules were paying an interpreted Aiki lambda frame merely to forward to substrate primitives. `machine/ffi`, `store/ffi`, the primitive entries in `bits/ffi`, and `bytes/ffi` now bind one-to-one substrate functions directly; policy-bearing helpers such as `store/ffi.snapshot` remain Aiki functions.
- The trusted bounded PDP path addresses already-bound CPU/audit/UNIBUS typed stores directly and gives ordinary RAM the first UNIBUS branch before I/O-page dispatch. No PDP instruction semantics moved into the substrate.
- The interactive `aiki-pdp` monitor now runs bounded slices (128 guest instructions per host turn) rather than returning through the monitor loop after every guest instruction. This is the path profiled by the Cut 7 performance diagnostic.
- KL11 host output is queued after characters leave TPB, so bounded slices cannot overwrite console output before the host drains it. The PDP-visible READY contract is preserved.
- New `cut7_live_slice_test.ai` proves queued A/B output ordering and architectural HALT within a 128-instruction slice (16/16 assertions).
- Disposable semantic profile at 10x: calls 1,040,917 -> 541,586; index 246 -> 258; general store 9 reads / 119 writes (setup-level). Wall-clock from the disposable Go 1.23/stub build is not acceptance evidence.
- Preserve as a separate runtime-design follow-on: `store` was intended as isolated mutable mapped memory. The current general store permits crossing `spawn` and therefore uses mutexes. Reconcile the sharing/isolation contract before removing synchronization; do not conflate that language/runtime correction with PDP emulator tuning.
