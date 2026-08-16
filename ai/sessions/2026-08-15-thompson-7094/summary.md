# Summary — Thompson 7094 regex reconstruction

The reconstruction now has a deliberately simple narrative: **machine -> paper
-> compiler**.

Phase I builds only enough IBM 7094 to make the historical program real. Phase
II asks whether Thompson's published runtime and worked object program actually
run on that machine. Phase III reconstructs the compiler only after the
machine-level historical result is independently established.

Aiki's current systems substrate materially simplifies the work compared with
the original sketch. Exact numbers, fixed-width bit operations, mutable stores,
isolated concurrency, channels/select, and the experiment framework already
exist. The emulator therefore requires no language extension before work can
begin.

The important architectural constraint is ownership rather than capability.
Aiki stores can cross spawn boundaries explicitly, but the eventual concurrent
machine process must remain semantic owner of emulated memory and registers.
Host control should be expressed through commands, not arbitrary shared-store
mutation.


## Cut 1 runtime gate

The user executed the Phase I / Cut 1 experiment on 2026-08-15 with the
repository's `aiki v0.4.0-alpha-26` executable. The experiment completed and
retained its result transcript, promoting the cut from BLOCKED to GATED. A
non-fatal warning that the experiment shadows the prelude name `read` should be
removed as Cut 2 begins; it is cleanup, not a semantic defect in the machine
representation.


## Primary-source correction during Phase I

The addition of IBM's 7094 *Principles of Operation* materially strengthened the
machine reconstruction. Before those encodings were available, Cut 2 briefly
used internal canonical opcode identifiers. That choice never became a gate. The
IBM manual showed both the actual operation codes and the correct numeric
orientation of IBM's left-numbered word fields, so the interim representation
was removed. The emulator now builds genuine 36-bit IBM instruction words for
the supported subset.

The same review sharpened the accumulator claim. A complete physical AC is not
needed: Thompson's runtime requires the logical P,1-35 slice. Modeling that slice
explicitly preserves the meaningful distinction between CLA and CAL without
pretending to emulate unused S/Q overflow machinery.

## Phase II — runtime reconstruction and a historical discrepancy

Thompson's runtime maps unusually cleanly onto the targeted emulator. CNODE and
NNODE use self-modifying address fields as list counts; XCHG copies executable
TSX continuations between NLIST and CLIST; TSX return registers and complemented
index arithmetic provide the control structure described in the paper without
adding any emulator facilities beyond Phase I.

The first object-level execution analysis exposed an important internal
inconsistency in the 1968 paper. The printed final listing sends location 4 to
`CODE+13`, entering the b|c alternation directly. Thompson's compiler source,
however, overwrites that operand entry during closure compilation with
`TRA CODE+pc`; pc is 16 for the worked example. Figure 5 agrees with the
compiler: the path after `a` enters the closure CNODE first, permitting the
lambda branch of `*`.

The reconstruction therefore treats historical fidelity as preservation rather
than normalization. The exact printed program remains executable as a verbatim
artifact, while the compiler-derived `CODE+16` variant is separately identified
and used to test the behavior Thompson describes. Phase III will turn this
source reconciliation into an executable invariant by generating the object
program from the compiler itself.

## Phase II gate and Phase III reconstruction

Phase II passed under the authoritative Aiki v0.4.0-alpha-26 executable. The
compiler reconstruction then preserved a strict provenance distinction:
Thompson specifies but does not publish the algorithms for syntax sieving and
RPN conversion, so those are conventional reconstruction choices; his ALGOL-60
third stage is reconstructed directly.

Executing the reconstructed third-stage logic independently explains the
printed object-code discrepancy found in Phase II. At closure compilation the
operand entry is location 4 and the current object pc is 16, so Thompson's own
rewrite necessarily emits `TRA CODE+16`. This establishes the correction by a
second route rather than by carrying forward a hand-edited word.
