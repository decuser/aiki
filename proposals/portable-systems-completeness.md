# Proposal: Portable Systems Completeness

## Purpose

Bring Aiki to the point where it has no material capability gaps for portable systems programming at its intended level of abstraction.

The goal is not to reproduce Unix, expose every host facility, or add a general escape hatch. The goal is that ordinary portable systems software should not require leaving Aiki because a fundamental host capability is absent.

## Governing Principles

> As much as possible belongs in Aiki. HAL exists only for operations that cannot be expressed without host participation.

> Completeness requires sufficient portable capability, not the union of host facilities.

> Every new byte-oriented host resource must first be considered as an implementation of the existing I/O abstraction. A parallel read/write API requires a semantic reason why the resource cannot participate in `io`.

A capability is complete when Aiki exposes the strongest useful contract that can be represented consistently across intended substrates. Portable completeness does not require exposing every facility of every host.

## Completeness Criterion

Aiki is materially complete for portable systems programming when an Aiki program can:

- manipulate files, directories, paths, links, permissions, and portable file locks sufficient for interprocess coordination;
- inspect and modify environment and working-directory state;
- work with text and arbitrary bytes;
- start and manage processes;
- communicate with running processes through standard streams;
- observe process completion and status;
- terminate processes;
- receive and send portable process signals;
- establish and accept network connections;
- use stream and datagram networking;
- perform interactive terminal I/O;
- coordinate concurrent work;
- use time and timers;
- determine at runtime whether required host capabilities exist.

Higher-level behavior should be expressible in Aiki from those facilities.

## Phase 1 — Process Lifecycle and Process I/O

Complete the current execute-and-collect process model with start, attached stdin/stdout/stderr endpoints, wait/status, and terminate.

### HAL mechanism

Provide only irreducible host contracts for process start, process endpoint acquisition/ownership, wait/status, and termination.

Process endpoints must first be treated as ordinary `io` endpoints. Do not create `process.read` or `process.write` if the existing I/O abstraction can express the behavior honestly.

### Aiki layer

Keep policy and composition in Aiki where practical: run-and-wait, captured execution, attached execution, pipelines, forwarding, supervision, retries, timeout policy, and concurrent process management.

Existing `system.exec` remains the simple convenience form and may later be implemented over the general process machinery.

## Phase 2 — Signals

Add the smallest portable signal capability needed for process/runtime coordination.

Prefer signal receipt as ordinary Aiki channel activity so it composes with `select`. Keep the portable vocabulary intentionally small and do not expose a host-specific signal namespace as the portable contract.

Higher-level shutdown, cleanup, forwarding, and supervision policy belongs in Aiki.

## Phase 3 — Networking

Add the smallest useful portable network substrate, initially TCP and UDP.

TCP connections should participate in the existing `io` abstraction where semantics overlap. UDP remains message-oriented and must not be disguised as a stream.

Keep line protocols, buffering, framing, retries, request/response patterns, concurrent servers, and protocol parsing in Aiki where practical.

HTTP, TLS, JSON, and other higher-level protocols are not required for this completeness claim.

## Phase 4 — Terminal / TTY

Add only the portable terminal mechanisms required for serious interactive programs: terminal detection, dimensions, raw/noncanonical mode, and reliable state restoration.

ANSI handling and higher-level terminal behavior should remain Aiki code where possible.

## Phase 5 — File Locking

Provide portable advisory locking sufficient for interprocess coordination.

Portable completeness requires a useful locking capability, not identical exposure of every host locking mode. Exclusive advisory locking alone satisfies this phase if that is the strongest contract supported substrates can represent consistently.

Lock-file policy, retries, timeouts, singleton guards, and transactional update conventions belong in Aiki.

## Cross-Cutting I/O Architecture

I/O remains the abstraction. Files, standard endpoints, process pipes, terminals, and TCP connections should share `io` operations where their semantics genuinely coincide.

Do not force artificial uniformity: UDP and other message-oriented resources retain their real boundaries.

## Capability and Authority Architecture

Every irreducible host facility participates in Gate 1 capability metadata and the existing authority model.

Availability answers whether the runtime can perform an operation. Authority answers whether the current code may perform it. These remain independent.

Aiki source queries Aiki capability names only, never HAL or substrate identities.

## Concurrency

Use existing channels and `select` as the preferred coordination model for external events where natural: signals, process completion, timers, and possibly accepted connections.

Do not add futures, promises, callback frameworks, host event loops, or polling APIs without concrete evidence that channels/select are insufficient.

## Representation Boundaries

HAL accepts legitimate Aiki representations and normalizes them at the boundary. Reads may return a canonical efficient representation; writes should accept legitimate Aiki representations where practical.

This applies especially to process, network, and terminal byte I/O.

## Explicitly Not Required

Portable systems completeness does not require mmap, arbitrary file descriptors, ioctl, raw devices, proc/sys interfaces, epoll/kqueue/IOCP exposure, shared memory, raw sockets, packet capture, Unix-domain sockets as a separate initial capability, host-specific signal namespaces, host-specific process attributes, kernel interfaces, HTTP, TLS, JSON, or database clients.

No general host/Go escape hatch is introduced.

## Implementation Discipline

For each phase:

1. identify the irreducible host operation;
2. define canonical HAL identities and metadata;
3. implement the substrate mechanism;
4. build useful policy/composition in Aiki;
5. document Aiki concepts rather than substrate details;
6. add invariant coverage and negative assurance for architectural joins;
7. add behavioral/conformance coverage;
8. gate with `make invariant` and `make validate`.

## Phase Order

1. Process lifecycle and pipes
2. Signals
3. Networking
4. Terminal / TTY
5. File locking
6. Completeness audit

## Representative Completion Programs

The final audit should include primarily-Aiki programs demonstrating a pipeline runner, process supervisor, concurrent TCP service, UDP utility, interactive terminal program, and locked filesystem updater.

These are acceptance tests for the architectural claim, not feature demos.

## Completion Claim

> Aiki provides the irreducible capabilities required for general portable systems programming, while keeping systems policy and composition in Aiki itself.

Further host features should thereafter be added only under concrete program pressure, not in pursuit of API completeness for its own sake.


## Phase 6 — Completeness Audit Result

The source-level completion audit was performed after Phases 1–5 gated.

The audit found one material miss against this proposal's own criterion: the
runtime environment was readable but not mutable, and child processes inherited
the embedding process environment rather than an Aiki-runtime-owned snapshot.
That gap was remediated with `system.environ`, `system.set_env`, and
`system.unset_env`; both `system.exec` and `process.start` now inherit a snapshot
of the runtime-owned environment at process creation.

After that remediation, the audited portable systems surface is:

- filesystem, directories, links, permissions, whole-file and streaming I/O;
- runtime-owned arguments, environment, and working directory;
- synchronous execution and interactive process lifecycle with pipe endpoints;
- portable signal receipt/delivery through channels and process handles;
- TCP stream endpoints and UDP datagrams;
- terminal detection, dimensions, raw mode, and restoration;
- timers, channels, spawn, and select for coordination and timeout policy;
- exclusive advisory interprocess file locking;
- capability discovery and authority separation for host-dependent facilities.

The endpoint test held across the work: files, standard streams, process pipes,
and TCP connections use the common `io` abstraction. UDP remains datagram
oriented because its message boundary is semantically real.

No further irreducible portable capability gap was identified. Features such as
TLS, HTTP, JSON, mmap, ioctl, raw descriptors, raw sockets, host-specific signal
namespaces, and kernel-specific event facilities remain outside the completeness
criterion and should be added only under concrete program pressure.

The bounded completion claim therefore stands, subject to the normal
`make invariant` and `make validate` gates:

> Aiki provides the irreducible capabilities required for general portable
> systems programming, while keeping systems policy and composition in Aiki
> itself.
## Final Reconciliation

A line-by-line source reconciliation was performed after the Phase 6 remediation
passed `make validate`. The implementation matches the bounded completeness
criterion and the five capability phases.

Confirmed joins include:

- process lifecycle and attached standard streams through opaque runtime-owned
  process/endpoint values;
- portable signal receipt through Aiki channels and signal delivery through
  process handles;
- TCP connections through the common `io` abstraction and UDP through explicit
  datagram operations;
- terminal detection, dimensions, raw mode, and restoration over existing I/O
  resources;
- exclusive advisory file locking sufficient for portable interprocess
  coordination;
- runtime-owned environment and cwd inherited by both synchronous and streaming
  child-process paths;
- capability/profile, authority, HAL identity, substrate registration,
  provenance, and runtime-ownership invariants for the new host surfaces.

The representative-program pressure check found no additional irreducible host
mechanism. Pipeline forwarding can be written in Aiki by passing process
endpoints explicitly to spawned functions; supervisor waiting and notification
can likewise be composed from process handles, channels, signals, timers, and
`select`. No direct process-to-process splice primitive is required for the
completeness claim.

One reconciliation defect was found in documentation: `io.read`, `io.read_line`,
`io.write`, and `io.close` still described only standard streams/files even
though process pipes and TCP connections already use the same endpoint
abstraction. The help/docs were corrected; no runtime change was required.

No material implementation/proposal contradiction remains. The portable systems
completeness project is reconciled and ready for its separate showcase work.

