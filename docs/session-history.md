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
