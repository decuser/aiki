package substrate

import "aiki/engine/semantics/value"

// registerHAL registers all HAL primitives (_prefixed, prelude-only).
func (g *GoRuntime) registerHAL() {
	// IO primitives
	g.registerHALBuiltin("_print", halPrint)
	g.registerHALBuiltin("_read", halRead)

	// List primitives
	g.registerHALBuiltin("_first", halFirst)
	g.registerHALBuiltin("_rest", halRest)
	g.registerHALBuiltin("_length", halLength)
	g.registerHALBuiltin("_prepend", halPrepend)
	g.registerHALBuiltin("_append", halAppend)
	g.registerHALBuiltin("_empty", halEmpty)

	// Type primitives
	g.registerHALBuiltin("_type", halType)
	g.registerHALBuiltin("_inspect", halInspect)
	g.registerHALBuiltin("_equal", halEqual)
	g.registerHALBuiltin("_ord", halOrd)

	// Math primitives
	g.registerHALBuiltin("_floor", halFloor)
	g.registerHALBuiltin("_ceil", halCeil)

	// Range
	g.registerHALBuiltin("_range", halRange)
}

// registerPrelude registers user-visible builtins.
// These are direct wrappers to HAL primitives until prelude.ai is loaded.
func (g *GoRuntime) registerPrelude() {
	// For now, expose HAL primitives directly with user-visible names.
	// Once prelude.ai is loaded, it can shadow these with its own wrappers.
	g.registerPreludeBuiltin("print", halPrint)
	g.registerPreludeBuiltin("read", halRead)
	g.registerPreludeBuiltin("input", halRead) // alias

	g.registerPreludeBuiltin("first", halFirst)
	g.registerPreludeBuiltin("rest", halRest)
	g.registerPreludeBuiltin("length", halLength)
	g.registerPreludeBuiltin("prepend", halPrepend)
	g.registerPreludeBuiltin("append", halAppend)
	g.registerPreludeBuiltin("empty", halEmpty)

	g.registerPreludeBuiltin("type", halType)
	g.registerPreludeBuiltin("inspect", halInspect)
	g.registerPreludeBuiltin("equal", halEqual)
	g.registerPreludeBuiltin("ord", halOrd)

	g.registerPreludeBuiltin("floor", halFloor)
	g.registerPreludeBuiltin("ceil", halCeil)

	g.registerPreludeBuiltin("range", halRange)
}

// registerHALBuiltin adds a HAL primitive.
func (g *GoRuntime) registerHALBuiltin(name string, fn func(args []value.Value) value.Value) {
	g.mu.Lock()
	g.halRegistry[name] = &Builtin{name: name, fn: fn}
	g.mu.Unlock()
}
