package substrate

import "aiki/engine/semantics/value"

// registerHAL registers all HAL primitives (_prefixed, prelude-only).
func (g *GoRuntime) registerHAL() {
	// IO
	g.registerHALBuiltin("_print", halPrint)
	g.registerHALBuiltin("_read", halRead)

	// List
	g.registerHALBuiltin("_first", halFirst)
	g.registerHALBuiltin("_rest", halRest)
	g.registerHALBuiltin("_length", halLength)
	g.registerHALBuiltin("_prepend", halPrepend)
	g.registerHALBuiltin("_append", halAppend)
	g.registerHALBuiltin("_empty", halEmpty)
	g.registerHALBuiltin("_range", halRange)

	// Type
	g.registerHALBuiltin("_type", halType)
	g.registerHALBuiltin("_inspect", halInspect)
	g.registerHALBuiltin("_equal", halEqual)
	g.registerHALBuiltin("_ord", halOrd)

	// Math
	g.registerHALBuiltin("_floor", halFloor)
	g.registerHALBuiltin("_ceil", halCeil)
	g.registerHALBuiltin("_modulo", halModulo)
	g.registerHALBuiltin("_cos", halCos)
	g.registerHALBuiltin("_sin", halSin)

	// Time
	g.registerHALBuiltin("_sleep", halSleep)

	// Canvas
	g.registerHALBuiltin("_canvas", halCanvas)
	g.registerHALBuiltin("_dot", halDot)
	g.registerHALBuiltin("_line", halLine)
	g.registerHALBuiltin("_rect", halRect)
	g.registerHALBuiltin("_fill_rect", halFillRect)
	g.registerHALBuiltin("_circle", halCircle)
	g.registerHALBuiltin("_fill_circle", halFillCircle)
	g.registerHALBuiltin("_clear", halClear)
	g.registerHALBuiltin("_destroy", halDestroy)
	g.registerHALBuiltin("_set_bg", halSetBG)
	g.registerHALBuiltin("_set_fg", halSetFG)
	g.registerHALBuiltin("_pen_size", halPenSize)
}

// registerPrelude registers user-visible builtins.
// These are direct mappings until prelude.ai can shadow them.
func (g *GoRuntime) registerPrelude() {
	// IO
	g.registerPreludeBuiltin("print", halPrint)
	g.registerPreludeBuiltin("read", halRead)
	g.registerPreludeBuiltin("input", halRead)

	// List
	g.registerPreludeBuiltin("first", halFirst)
	g.registerPreludeBuiltin("rest", halRest)
	g.registerPreludeBuiltin("length", halLength)
	g.registerPreludeBuiltin("prepend", halPrepend)
	g.registerPreludeBuiltin("append", halAppend)
	g.registerPreludeBuiltin("empty", halEmpty)
	g.registerPreludeBuiltin("range", halRange)

	// Type
	g.registerPreludeBuiltin("type", halType)
	g.registerPreludeBuiltin("inspect", halInspect)
	g.registerPreludeBuiltin("equal", halEqual)
	g.registerPreludeBuiltin("ord", halOrd)

	// Math
	g.registerPreludeBuiltin("floor", halFloor)
	g.registerPreludeBuiltin("ceil", halCeil)
	g.registerPreludeBuiltin("modulo", halModulo)
	g.registerPreludeBuiltin("cos", halCos)
	g.registerPreludeBuiltin("sin", halSin)

	// Time
	g.registerPreludeBuiltin("sleep", halSleep)

	// Canvas
	g.registerPreludeBuiltin("canvas", halCanvas)
	g.registerPreludeBuiltin("dot", halDot)
	g.registerPreludeBuiltin("line", halLine)
	g.registerPreludeBuiltin("rect", halRect)
	g.registerPreludeBuiltin("fill_rect", halFillRect)
	g.registerPreludeBuiltin("circle", halCircle)
	g.registerPreludeBuiltin("fill_circle", halFillCircle)
	g.registerPreludeBuiltin("clear", halClear)
	g.registerPreludeBuiltin("destroy", halDestroy)
	g.registerPreludeBuiltin("set_bg", halSetBG)
	g.registerPreludeBuiltin("set_fg", halSetFG)
	g.registerPreludeBuiltin("pen_size", halPenSize)
}

// registerHALBuiltin adds a HAL primitive.
func (g *GoRuntime) registerHALBuiltin(name string, fn func(args []value.Value) value.Value) {
	g.mu.Lock()
	g.halRegistry[name] = &Builtin{name: name, fn: fn}
	g.mu.Unlock()
}
