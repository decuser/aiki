package substrate

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
	g.register("_stack_limit", halStackLimit)
	g.register("_inspect", halInspect)
	g.register("_equal", halEqual)
	g.register("_ord", halOrd)
	g.register("_chr", halChr)

	// Math
	g.register("_floor", halFloor)
	g.register("_ceil", halCeil)
	g.register("_modulo", halModulo)
	g.register("_sqrt_inexact", halSqrt)
	g.register("_cos_inexact", halCos)
	g.register("_sin_inexact", halSin)
	g.register("_seed", halSeed)
	g.register("_random", halRandom)

	// Bytes
	g.register("_bytes_new", halBytesNew)
	g.register("_bytes_length", halBytesLength)
	g.register("_bytes_get", halBytesGet)
	g.register("_bytes_slice", halBytesSlice)
	g.register("_str_to_bytes", halStrToBytes)
	g.register("_bytes_to_str", halBytesToStr)
	g.register("_bytes_to_str_pure", halBytesToStrPure)

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
	g.register("_arc", halArc)
	g.register("_clear", halClear)
	g.register("_destroy", halDestroy)
	g.register("_set_bg", halSetBG)
	g.register("_set_fg", halSetFG)
	g.register("_pen_size", halPenSize)
	g.register("_canvas_width", halCanvasWidth)
	g.register("_canvas_height", halCanvasHeight)
	g.register("_canvas_alive", halCanvasAlive)
	g.register("_set_turtle", halSetTurtle)

	// Convert
	g.register("_shape", halShape)
	g.register("_make_shaped_list", halMakeShapedList)
	g.register("_to_str", halToStr)
	g.register("_to_decimal", halToDecimal)
	g.register("_to_number", halToNumber)

	// String (Unicode case conversion)
	g.register("_upper", halUpper)
	g.register("_lower", halLower)
	g.register("_upper_rune", halUpperRune)
	g.register("_lower_rune", halLowerRune)

	// Regex
	g.register("_regex_match", halRegexMatch)
	g.register("_regex_find", halRegexFind)
	g.register("_regex_find_all", halRegexFindAll)
	g.register("_regex_replace", halRegexReplace)
	g.register("_regex_replace_first", halRegexReplaceFirst)
	g.register("_regex_split", halRegexSplit)

	// File I/O
	g.register("_file_open", halFileOpen)
	g.register("_file_read_text", halFileReadText)
	g.register("_file_read_bytes", halFileReadBytes)
	g.register("_file_read_line", halFileReadLine)
	g.register("_file_write_text", halFileWriteText)
	g.register("_file_write_bytes", halFileWriteBytes)
	g.register("_file_close", halFileClose)
	g.register("_file_exists", halFileExists)
	g.register("_file_delete", halFileDelete)

	// Intrinsics - these use evaluation context
	g.register("_apply", halApply)
	g.register("_import", halImport)
	g.register("_use", halUse)
	g.register("_export", halExport)
	g.register("_load", halLoad)
	g.register("_spawn", halSpawn)
	g.register("_channel", halChannel)
	g.register("_send", halSend)
	g.register("_recv", halRecv)

	// REPL
	g.register("_quit", halQuit)
	g.register("_reset", halReset)
	g.register("_delete", halDelete)
	g.register("_help", halHelp)
	g.register("_doc", halDoc)
}
