package substrate

import "aiki/engine/semantics/value"

// registerHAL registers all HAL primitives (_prefixed, prelude-only).
func (g *GoRuntime) registerHAL() {
	// IO
	g.register("_print", halPrint)
	g.register("_read", halRead)

	// List
	g.register("_first", halFirst)
	g.register("_rest", halRest)
	g.register("_length", halLength)
	g.register("_prepend", halPrepend)
	g.register("_append", halAppend)
	g.register("_empty", halEmpty)
	g.register("_range", halRange)

	// Type
	g.register("_type", halType)
	g.register("_inspect", halInspect)
	g.register("_equal", halEqual)
	g.register("_ord", halOrd)

	// Math
	g.register("_floor", halFloor)
	g.register("_ceil", halCeil)
	g.register("_modulo", halModulo)
	g.register("_cos", halCos)
	g.register("_sin", halSin)

	// Time
	g.register("_sleep", halSleep)

	// Canvas
	g.register("_canvas", halCanvas)
	g.register("_dot", halDot)
	g.register("_line", halLine)
	g.register("_rect", halRect)
	g.register("_fill_rect", halFillRect)
	g.register("_circle", halCircle)
	g.register("_fill_circle", halFillCircle)
	g.register("_clear", halClear)
	g.register("_destroy", halDestroy)
	g.register("_set_bg", halSetBG)
	g.register("_set_fg", halSetFG)
	g.register("_pen_size", halPenSize)
	g.register("_shape", halShape)
	g.register("_to_str", halToStr)
}

// register adds a HAL primitive to the registry.
func (g *GoRuntime) register(name string, fn func(args []value.Value) value.Value) {
	g.mu.Lock()
	g.registry[name] = &Builtin{name: name, fn: fn}
	g.mu.Unlock()
}
