package substrate

import (
	"fmt"
	"strings"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// runtimeTestState accumulates Aiki test results for one runtime.
type runtimeTestState struct {
	failures []testFailure
	passed   int
	failed   int
	current  string // name of the current test.run block, if any
	file     string // file of the current test, set by the test runner
}

type testFailure struct {
	test    string // test.run name, or "" for top-level
	file    string
	line    int
	message string
}

// ResetTestState clears accumulated test results for this runtime.
func (g *GoRuntime) ResetTestState() {
	g.mu.Lock()
	g.testState = runtimeTestState{}
	g.mu.Unlock()
}

// SetTestFile sets the file path for the current runtime test.
func (g *GoRuntime) SetTestFile(path string) {
	g.mu.Lock()
	g.testState.file = path
	g.mu.Unlock()
}

// TestResults returns the accumulated results for this runtime.
func (g *GoRuntime) TestResults() (passed, failed int, failures []string) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var msgs []string
	for _, f := range g.testState.failures {
		loc := fmt.Sprintf("%s:%d", f.file, f.line)
		if f.test != "" {
			msgs = append(msgs, fmt.Sprintf("  FAIL %s (%s): %s", f.test, loc, f.message))
		} else {
			msgs = append(msgs, fmt.Sprintf("  FAIL (%s): %s", loc, f.message))
		}
	}
	return g.testState.passed, g.testState.failed, msgs
}

func (g *GoRuntime) recordTestPass() {
	g.mu.Lock()
	g.testState.passed++
	g.mu.Unlock()
}

func (g *GoRuntime) recordTestFailure(ctx *hal.EvalContext, msg string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.testState.failed++
	file := g.testState.file
	line := 0
	// The stack frames carry the call site line but the callee's file.
	// For assertions inside a run() block, the "run" frame has the
	// correct line (from the user's test file) but the wrong file
	// (test.ai, because run's body is there). We use the runtime test file for
	// the file and take the line from the deepest non-<main> frame.
	if ctx != nil && ctx.Env != nil {
		stack := ctx.Env.CopyStack()
		for i := len(stack) - 1; i >= 0; i-- {
			f := stack[i]
			if f.Name != "<main>" && f.Line > 0 {
				line = f.Line
				break
			}
		}
		if line == 0 && len(stack) > 0 && stack[0].Line > 0 {
			line = stack[0].Line
		}
	}
	if file == "" && ctx != nil && ctx.Env != nil {
		file = ctx.Env.GetFile()
	}
	if line == 0 && ctx != nil && ctx.Node != nil {
		line = ctx.Node.Pos.Line
	}
	g.testState.failures = append(g.testState.failures, testFailure{
		test:    g.testState.current,
		file:    file,
		line:    line,
		message: msg,
	})
}

// halTestEqual asserts that actual and expected are deeply equal.
func (g *GoRuntime) halTestEqual(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("test.equal: want 2 arguments, got %d", len(args))
	}
	actual := args[0]
	expected := args[1]
	if value.DeepEqual(actual, expected) {
		g.recordTestPass()
		return value.TRUE
	}
	g.recordTestFailure(ctx, fmt.Sprintf("expected %s, got %s", expected.Inspect(), actual.Inspect()))
	return value.FALSE
}

// halTestNotEqual asserts that actual and expected are not equal.
func (g *GoRuntime) halTestNotEqual(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("test.not_equal: want 2 arguments, got %d", len(args))
	}
	actual := args[0]
	expected := args[1]
	if !value.DeepEqual(actual, expected) {
		g.recordTestPass()
		return value.TRUE
	}
	g.recordTestFailure(ctx, fmt.Sprintf("expected value different from %s", expected.Inspect()))
	return value.FALSE
}

// halTestTrue asserts that value is truthy.
func (g *GoRuntime) halTestTrue(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("test.true: want 1 argument, got %d", len(args))
	}
	if value.IsTruthy(args[0]) {
		g.recordTestPass()
		return value.TRUE
	}
	g.recordTestFailure(ctx, fmt.Sprintf("expected truthy, got %s", args[0].Inspect()))
	return value.FALSE
}

// halTestFalse asserts that value is falsy.
func (g *GoRuntime) halTestFalse(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("test.false: want 1 argument, got %d", len(args))
	}
	if !value.IsTruthy(args[0]) {
		g.recordTestPass()
		return value.TRUE
	}
	g.recordTestFailure(ctx, fmt.Sprintf("expected falsy, got %s", args[0].Inspect()))
	return value.FALSE
}

// halTestError asserts that value is a shaped error.
func (g *GoRuntime) halTestError(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("test.error: want 1 argument, got %d", len(args))
	}
	if value.IsShapedError(args[0]) {
		g.recordTestPass()
		return value.TRUE
	}
	g.recordTestFailure(ctx, fmt.Sprintf("expected error, got %s", args[0].Inspect()))
	return value.FALSE
}

// halTestNotError asserts that value is not a shaped error.
func (g *GoRuntime) halTestNotError(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("test.not_error: want 1 argument, got %d", len(args))
	}
	if !value.IsShapedError(args[0]) {
		g.recordTestPass()
		return value.TRUE
	}
	g.recordTestFailure(ctx, fmt.Sprintf("expected non-error, got %s", args[0].Inspect()))
	return value.FALSE
}

// halTestRun runs a named test function, isolating faults.
func (g *GoRuntime) halTestRun(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 2 {
		return value.NewFault("test.run: want (name, fn), got %d arguments", len(args))
	}
	nameVal, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("test.run: name must be a string")
	}
	name := nameVal.Val

	g.mu.Lock()
	prev := g.testState.current
	g.testState.current = name
	g.mu.Unlock()

	// The callback was defined in the user's test file. Its closure
	// env carries that file path, which we use for failure locations.
	fn := args[1]
	if fnVal, ok := fn.(*value.Function); ok {
		if fnEnv, ok := fnVal.Env.(*value.Env); ok {
			g.mu.Lock()
			g.testState.file = fnEnv.GetFile()
			g.mu.Unlock()
		}
	}

	// Call the test function. If it faults, record the fault as a failure.
	var result value.Value
	if ctx != nil && ctx.Eval != nil {
		fnVal, ok := fn.(*value.Function)
		if ok {
			fnEnv, envOk := fnVal.Env.(*value.Env)
			if envOk {
				callEnv := value.NewEnclosedEnv(fnEnv)
				if body, bodyOk := fnVal.Body.(*syntax.Node); bodyOk {
					result = ctx.Eval(body, callEnv)
				}
			}
		} else if callable, ok := fn.(value.Callable); ok {
			result = callable.Call(nil)
		}
	}

	// If the result is a fault, record it as a test failure
	if fault, ok := result.(*value.Fault); ok {
		var msg strings.Builder
		msg.WriteString("fault: ")
		msg.WriteString(fault.Message)
		g.recordTestFailure(ctx, msg.String())
	}

	g.mu.Lock()
	g.testState.current = prev
	g.mu.Unlock()

	return value.EMPTY
}

// halTestFaults asserts that fn faults when called.
func (g *GoRuntime) halTestFaults(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("test.faults: want 1 argument (fn), got %d", len(args))
	}
	fn := args[0]

	var result value.Value
	if ctx != nil && ctx.Eval != nil {
		if fnVal, ok := fn.(*value.Function); ok {
			if fnEnv, ok := fnVal.Env.(*value.Env); ok {
				callEnv := value.NewEnclosedEnv(fnEnv)
				if body, ok := fnVal.Body.(*syntax.Node); ok {
					result = ctx.Eval(body, callEnv)
				}
			}
		} else if callable, ok := fn.(value.Callable); ok {
			result = callable.Call(nil)
		}
	}

	if _, ok := result.(*value.Fault); ok {
		g.recordTestPass()
		return value.TRUE
	}
	g.recordTestFailure(ctx, "expected a fault, but none occurred")
	return value.FALSE
}
