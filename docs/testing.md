# Testing, Blessing, and Validation

Aiki separates three different acts: checking correctness, blessing a known-good reference state, and validating later work against that state.

## `make check`

`make check` runs correctness checks that do not depend on existing behavior or engine gold snapshots. It builds and formats the tree, runs the linter, verifies the distribution tree with `aiki treecheck`, runs Go tests and Aiki tests, and verifies that the structural engine specimens exercise every production declared by `grammar.ebnfx`.

`make check` never writes gold files.

Use it while developing an intentional behavior or structural change, before updating reference snapshots.

## `make bless`

`make bless` first runs `make check`. Only if those independent checks succeed does it replace the blessed reference snapshots:

```text
behavior smoke transcripts
lexer structural golds
parser structural golds
evaluator structural golds
EBNF grammar gold
```

Blessing is not validation. It records the current implementation as the new reference state after the intended behavior has been established independently.

The underlying commands are:

```bash
./aiki smoke --gold test/behavior/
./aiki enginesmoke --stage all --gold test/structure/engine
```

`smoke --gold` preserves authored `IN:` and `DISPLAY:` directives from an existing transcript, then regenerates observed `OUT:`, `ERR:`, `EXIT:`, and `CANVAS:` records using the smoke framework's canonical transcript encoding. A new no-input smoke can be blessed without an existing gold; input-bearing smokes should establish their `IN:` directives before blessing.

### Negative behavior fixtures

A smoke specimen whose intended behavior is failure declares that intent in a
leading source comment. The declaration is textual so it remains readable even
when the specimen intentionally does not parse:

```text
# @negative parse
```

The marker is a general negative-fixture declaration; `parse` is currently the
only supported kind. Unknown kinds are errors. Negative declarations are valid
only on `*_smoke.ai` specimens; placing one in ordinary `.ai` source is an
error, so application or library source cannot exempt itself from formatter or
linter coverage. A parse-negative fixture is excluded from formatter and linter
source traversal because successful parsing is not part of that specimen's
contract.

Intent and evidence are checked separately. `# @negative parse` is the intent;
the smoke transcript is the observation. Smoke validation requires both
directions: every declared parse-negative must have a nonzero parser-failure
transcript, and every parser-failure transcript must belong to a specimen that
declares `# @negative parse`. Blessing applies the same check to the observed
run before writing a gold, so a newly broken ordinary specimen cannot authorize
itself by being reblessed.

`enginesmoke --gold` refuses to bless an incomplete structural suite: every EBNF production must first be exercised by at least one `*_engine.ai` specimen.

## `make validate`

`make validate` runs `make check`, then compares current behavior and engine structure with the already-blessed golds.

It is read-only with respect to gold files.

A successful validation means:

- the independent correctness checks pass;
- every declared grammar production is exercised structurally;
- behavior smoke output matches its blessed transcript;
- lexer, parser, evaluator, and EBNF structures match their blessed snapshots.

## Normal workflow

For a change that should not alter blessed behavior or structure:

```bash
make validate
```

For an intentional behavior or structural change:

```bash
make check
# inspect and establish that the new behavior/structure is correct
make bless
make validate
```

Then commit the implementation and updated golds together.

## Why golds come after checks

A gold is an oracle for future regression detection; generating one does not prove that the captured behavior is correct. Unit, semantic, invariant, and human review establish correctness first. Blessing then freezes that known-good state so later drift is visible.

## Distribution tree invariant

`aiki treecheck` checks that every file in the current source tree has a recognized distribution relationship or an explicit standalone disposition. It is part of `make check` and therefore `make validate`.

The checker infers ordinary relationships such as standard-library `.ai`/`.help`/`.doc` sets, Aiki-native tests, smoke specimens and golds, engine specimens and stage golds, grammar/prelude artifacts, samples, profiling drivers, Go implementation files, and direct references from already-justified text files. It also detects structural contradictions such as a gold without its source specimen or a module companion without an owning `.ai` file.

Intentional standalone artifacts are listed in the small root file `treecheck.allow`. That file is an exception list, not a manifest of the repository. Directory entries end in `/`; other entries use `filepath.Match`-style patterns. Add an exception only when a file is intentionally standalone and no stronger structural relationship describes it.

When Git metadata is available, `treecheck` examines tracked files plus untracked non-ignored files that actually exist in the working tree. This makes overlay additions visible before staging while also allowing a removed tracked path to disappear normally. Without Git metadata, it falls back to walking the source tree.
