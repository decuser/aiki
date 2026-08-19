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
