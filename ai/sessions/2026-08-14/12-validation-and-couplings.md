# Milestone 12 - validation and executable couplings

## Intent

Re-run validation after the profiling dependency cleanup and documentation drift repair.

## Result

The post-cleanup tree passed the executable checks that can be exercised in the offline container harness:

- Aiki tests: 406/406
- grammar coverage: 32/32 productions across 10 inputs
- behavior smokes: 34/34
- property tests: pass
- non-doc-example invariant tests: pass

The executable documentation invariant cannot be completed faithfully in the offline harness because its canvas examples require real Ebiten child-process behavior. The harness uses no-op graphics stubs solely to compile and exercise non-visual validation; this is an environment limitation, not a source failure.

## Additional clarification

Store sharing across spawn is now documented explicitly. A store crosses the isolation boundary only when passed as an argument. The Go implementation protects shared store access with an `RWMutex`; the mutex is an implementation mechanism, not an Aiki language feature.

## Restart

Validation is sufficiently green to continue the ordered queue. Next: spawned abnormal termination.
