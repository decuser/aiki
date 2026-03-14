Aiki Error Refactor Implementation Contract

This refactor must follow the specification in this document exactly.
Do not change the error model, representations, or semantics unless explicitly instructed.

The following rules are mandatory.

Do not change the two strata model

Aiki has exactly two failure strata:

Fault
internal evaluation failure
halts evaluation immediately
not an ordinary Aiki value

Shaped error
[@error, :kind, "message"]
recoverable language visible value
flows through the program
falsy
pipes may short circuit on it

Do not collapse these two strata into one type.

Do not rename Fault to Error.

Do not represent Fault as a shaped list.

Do not introduce additional error types.

Type and representation rules

The internal halting type must remain:

value.Fault

Recoverable errors must remain shaped values:

[@error, :kind, "message"]

Do not introduce alternate representations.

Do not modify the shape of the shaped error value.

Evaluator rules

The evaluator must:

halt on IsFault
never halt on shaped errors

Shaped errors must behave as ordinary values except where explicitly defined.

Do not introduce ambient propagation for shaped errors.

Pipe rules

Pipes must short circuit only on shaped errors.

Pipes must not check for Fault.

Faults halt evaluation before pipe continuation.

Truthiness rules

Shaped errors must be falsy.

Faults must never participate in truthiness because evaluation halts before that point.

Runner rules

Fault causes execution failure and nonzero exit.

A final shaped error value is treated as a normal result.

It prints normally and exits zero.

Boundary rules

Program lex and parse failures remain host boundary failures.

They are not shaped values.

Callable parsing utilities may return shaped errors.

HAL audit rule

Do not mechanically convert previous error calls.

For each site ask:

Did the expression fail to evaluate validly?
Then return Fault.

Did the operation validly complete but return failure as data?
Then return a shaped error.

Implementation order must remain

1 value package
2 evaluator core
3 runner behavior
4 pipe behavior
5 HAL audit
6 tests

Do not change this order.

Refactor scope

This refactor must not modify:

type system
value representation
AST structure
evaluation semantics unrelated to errors
existing operator contracts

Only the error model described here may change.

If uncertainty arises

Stop and ask for clarification.

Do not invent new semantics.

Aiki Error Refactor Guidance

Core model

Aiki has two failure strata:

Fault
internal evaluation failure
halts evaluation
not an ordinary Aiki value

Shaped error
[@error, :kind, "message"]
recoverable language visible value
falsy
inspectable, matchable, bindable
honored specially only by explicit constructs such as pipes

Classification rule

Use Fault when the current expression could not validly evaluate.

Use shaped [@error, :kind, "message"] when an operation validly completed and returned failure as data.

Representation

Introduce an internal halting type:

value.Fault
or, if preferred, value.EvalFailure

Constructors for evaluator and HAL internals:

NewFault(...)

Keep shaped recoverable value constructors only for language visible recoverable failure:

HALError(...)
PreludeError(...)
UserError(...)

Each returns:

[@error, :kind, "message"]

Do not collapse these two strata into one type.

Do not rename Fault back to Error.

Do not make :kind secretly control evaluator halting.

Detection

Add:

IsFault(v)
true only for the internal halting type

IsShapedError(v)
true only for visible recoverable values of the form
[@error, :kind, "message"]

Remove or excise old evaluator level IsError logic.
Do not use one predicate for both strata.

Evaluator rules

Halt on IsFault.

Never halt on IsShapedError.

Shaped errors are values, not ambient control flow.

Faults never flow through ordinary evaluation.

Truthiness

Shaped errors are falsy.

Faults should never participate in ordinary truthiness because evaluation halts before they can behave as values.

Pipe rule

Pipes short circuit on shaped recoverable errors only.

Faults never enter pipe flow because they halt evaluation earlier.

Runner rule

Fault is execution failure.
Nonzero exit.
Reported by the host through stderr or equivalent.

A final shaped error value is just a result.
It prints normally.
It exits zero.

Boundary rule

Program lex and parse failure stay boundary failures.
They are not shaped values.

Callable parsing utilities inside the language may return shaped recoverable errors.

Naming rule

Keep the visible language concept of error associated with shaped error values.

Keep the internal halting concept named distinctly as Fault or EvalFailure.

Do not overload one name across both strata if it can be avoided.

HAL audit rule

Do not mechanically convert old NewError call sites.

At each site ask:

Did the expression fail to evaluate validly?
Then return Fault.

Did the operation validly complete and return failure as data?
Then return shaped [@error, :kind, "message"].

Typical Fault cases

wrong arity
type mismatch
bad operand kinds
division by zero
sqrt of negative if numbers are strictly real
calling a non function
invalid application
index out of bounds
undefined symbol if evaluator failure
stack overflow
broken evaluator state
malformed internal module state

Typical shaped recoverable cases

missing key, if lookup miss is intended recoverable
file not found
cannot read resource
resource unavailable
closed resource, if treated as recoverable resource state
callable parse failure such as to_number("abc")
user signaled recoverable failure
requested export missing from an otherwise valid module interface, if intended recoverable

Module and loader guidance

Top level program lex and parse failure are boundary failures, not shaped values.

Callable module operations such as import and load may return shaped recoverable errors for:

not found
cannot read
callable parse failure in loaded module text, if these operations are defined as library style calls

But malformed or inconsistent module structure should usually be Fault, for example:

wrong package identity, if that violates module validity
exported name declared but not actually defined
internally inconsistent module artifact

Partial operator guidance

Keep partial operators as Fault when their preconditions are violated.

Examples:

first([]) -> Fault
first("") -> Fault

Do not force core operators into maybe style behavior unless that is intentional.

If later needed, add explicit total alternatives such as:

first_or(xs, default)
maybe_first(xs)

Prelude helper

Add a prelude helper:

is_error(x)

Meaning:

returns true for shaped recoverable errors
returns false for ordinary values

It should not expose or test Fault, because faults are not ordinary Aiki values.

Implementation order

value/
define Fault
add IsFault
add IsShapedError
keep shaped errors falsy

evaluator core
halt on IsFault
never halt on shaped errors
remove vestigial IsError logic

runner and CLI path
Fault gives nonzero exit and host reporting
final shaped error value prints normally and exits zero

pipe behavior
short circuit on IsShapedError, not on Fault

HAL and substrate audit
classify each site by contract, not convenience

tests
update last, after semantics are stable

Non goals

Do not reintroduce ambient privileged propagation for shaped errors.

Do not make :hal mean halting.

Do not let arithmetic operators widen into number or shaped error unless you intentionally want lifted arithmetic across the whole numeric system.

Do not backslide into a one stratum model by renaming Fault to Error.

Design sentence to preserve

Aiki has two failure strata.

Fault is internal evaluation failure and halts.

[@error, :kind, "message"] is recoverable language visible failure and flows as data.
