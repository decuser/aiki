# Session summary — 2026-08-14-review

## Purpose

A non-coding session that reviewed the profiling implementation, assessed
an academic paper rejection, and formalized the AI working method as a
portable skill. No Aiki source was modified.

## Designer isolation as method

The session opened with a challenge: the designer has never read the Aiki
implementation code. The initial AI assessment treated this as a deficit
requiring remediation. The designer rejected that framing.

The corrected understanding: not reading the implementation is constitutive
of the project's method. Aiki's commitments — exact rationals (no hidden
numeric state) and isolated spawn (no hidden coupling) — share a common
structure: no channel that isn't visible at the interface. The designer
applies that discipline reflexively. Reading the implementation would
acquire a model of Aiki that no user can have, contaminating design
judgments with implementation knowledge the language doesn't expose. The
designer's confusion is therefore diagnostic — if a doc entry is unclear
to the designer, it is a spec defect, not a knowledge gap.

The collaboration protocol has the same shape. Each session opens with a
tarball — an explicit message through a channel. No continuity is assumed
across sessions. The doc corpus is the only persistent state. This is the
language's own concurrency model (spawn, message-pass, isolated state)
realized as a development process.

Executable documentation is what makes this rigorous rather than merely
ignorant. If docs are the spec and the implementation is fungible, the spec
must be mechanically checkable against the implementation. The
`doc_examples_test.go` conformance suite is that boundary. The designer's
method is the existence proof for the language's pedagogical thesis about
local reasoning.

## Paper assessment

The MISQ rejection of "Beyond Effective Use: AI and the Failure of
Productive Stability" was assessed after the designer presented the EIC
letter and then the manuscript.

The letter contains a category contradiction: the EIC assigns the paper to
Theory Development (challenging existing foundations) and then prescribes the
method of Theory-Generative Research Synthesis (systematic review). The
"strawman" charge is misapplied — the paper engages TAM, UTAUT, DeLone &
McLean, TTF, and effective use, which are the IS evaluation canon.

One substantive finding survives: Section 7 (Capability Concealment) is the
paper's most consequential organizational claim and it is entirely asserted.
No case data, no evidence of an organization in that state. The strongest
workplace-scale study in the paper (Brynjolfsson et al.) found learning that
persisted when AI was unavailable — a counterexample to the erosion
mechanism. The self-citation dependency (Senn 2026 for the K ≠ K′ criterion)
is also a vulnerability at MISQ.

Assessment: desk reject with one real finding accessed by the wrong route.
The productive-stability reading of the five frameworks needs to be shown
from their texts rather than asserted. Capability concealment should be
demoted to proposition or evidenced. The venue may be a mismatch for the
paper's form.

The exchange also demonstrated the designer's method for using AI as a
diagnostic instrument on reviews — presenting the letter first (to check
whether the AI produces a deferential or independent reading), then the
paper (to check whether the AI's assessment changes). The prior conversation
about designer isolation measurably shifted the AI's reading of the rejection
toward independence.

## Profiling review

Full-tree validation confirmed: build passes (Go 1.24), all tests pass,
full-tree race detection passes. The profiling architecture is sound — the
two-view design (deterministic semantic counts vs. sampled substrate cost),
the probe interface, and the pprof label correlation across the HAL boundary
are well executed.

The concurrency fixes discovered during profiling work (killing
SetContext/currentCtx, introducing ContextCallable, splitting NewCallEnv
lexical-from-dynamic, NewIsolatedEnclosedEnv for spawn) were confirmed as
genuine correctness improvements, not merely profiling support.

One dependency concern: `value` now imports `engine` for the profiling
types. The designer is addressing this.

One documentation item: `Store` is an explicit exception to spawn isolation
(it has a sync.RWMutex and can be shared across spawns via closure capture).
This should be noted in store's documentation.

## Working-method skill

The `ai/` working record was assessed as the most valuable artifact in the
project for AI collaboration continuity. It was formalized as a portable
Claude skill that:

- triggers on any Aiki-related conversation, including non-coding sessions;
- encodes the serial-cut, evidence-gated working rules;
- specifies session record creation for discussions, reviews, and planning;
- produces drop-in tarballs of the `ai/` directory that merge without
  conflict into the repository.
