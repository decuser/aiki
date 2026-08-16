package substrate

// registerHAL installs the runtime binding surface separated by architectural
// responsibility. Only entries registered through registerHost have canonical
// HAL identities; the remaining registries implement non-HAL runtime roles.
func (g *GoRuntime) registerHAL() {
	// True host-facing operations. Canonical contracts currently exist for I/O,
	// time, program context, file operations, and the narrowed Canvas resource/command
	// boundary. Randomness remains a host-role compatibility operation.
	g.registerHost(hostOperationDescriptors["_print"], g.halPrint)
	g.registerHost(hostOperationDescriptors["_read"], g.halRead)
	g.registerHost(hostOperationDescriptors["_sleep"], halSleep)
	g.registerHost(hostOperationDescriptors["_after"], halAfter)
	g.registerHost(hostOperationDescriptors["_system_args"], g.halSystemArgs)
	g.registerHost(hostOperationDescriptors["_system_env"], g.halSystemEnv)
	g.registerHost(hostOperationDescriptors["_file_open"], g.halFileOpenPath)
	g.registerHost(hostOperationDescriptors["_file_read_text"], halFileReadText)
	g.registerHost(hostOperationDescriptors["_file_read_bytes"], halFileReadBytes)
	g.registerHost(hostOperationDescriptors["_file_read_line"], g.halFileReadLine)
	g.registerHost(hostOperationDescriptors["_file_write_text"], halFileWriteText)
	g.registerHost(hostOperationDescriptors["_file_write_bytes"], halFileWriteBytes)
	g.registerHost(hostOperationDescriptors["_file_close"], g.halFileClose)
	g.registerHost(hostOperationDescriptors["_file_exists"], g.halFileExistsPath)
	g.registerHost(hostOperationDescriptors["_file_delete"], g.halFileDeletePath)
	g.registerHost(hostOperationDescriptors["_file_list"], g.halFileListPath)
	g.registerHost(hostOperationDescriptors["_file_read_at"], halFileReadAt)
	g.registerHost(hostOperationDescriptors["_file_write_at"], halFileWriteAt)
	g.registerHost(hostOperationDescriptors["_file_stat"], g.halFileStatPath)
	g.registerHost(hostOperationDescriptors["_file_rename"], g.halFileRenamePath)
	g.registerHost(hostOperationDescriptors["_file_mkdir"], g.halFileMkdirPath)
	g.registerHost(hostOperationDescriptors["_file_mkdir_all"], g.halFileMkdirAllPath)
	g.registerHost(hostOperationDescriptors["_file_remove_all"], g.halFileRemoveAllPath)
	g.registerHost(hostOperationDescriptors["_file_temp"], halFileTemp)
	g.registerHost(hostOperationDescriptors["_file_temp_dir"], halFileTempDir)
	g.registerHost(hostOperationDescriptors["_file_copy"], g.halFileCopyPath)
	g.registerHost(hostOperationDescriptors["_file_size"], g.halFileSizePath)
	g.registerHost(hostOperationDescriptors["_time_now"], halTimeNow)
	g.registerHost(hostOperationDescriptors["_system_cwd"], g.halSystemCwd)
	g.registerHost(hostOperationDescriptors["_system_chdir"], g.halSystemChdir)
	g.registerHost(hostOperationDescriptors["_system_exec"], g.halSystemExec)
	g.registerHost(hostOperationDescriptors["_path_separator"], halPathSeparator)
	g.registerRole(roleHost, "_seed", g.halSeed)
	g.registerRole(roleHost, "_random", g.halRandom)
	g.registerHost(hostOperationDescriptors["_canvas"], g.halCanvas)
	g.registerHost(hostOperationDescriptors["_canvas_command"], g.halCanvasCommand)
	g.registerHost(hostOperationDescriptors["_destroy"], g.halDestroy)
	g.registerHost(hostOperationDescriptors["_canvas_width"], g.halCanvasWidth)
	g.registerHost(hostOperationDescriptors["_canvas_height"], g.halCanvasHeight)
	g.registerHost(hostOperationDescriptors["_canvas_alive"], g.halCanvasAlive)

	// Evaluator/language intrinsics.
	g.registerRole(roleIntrinsic, "_apply", halApply)
	g.registerRole(roleIntrinsic, "_import", g.halImport)
	g.registerRole(roleIntrinsic, "_use", g.halUse)
	g.registerRole(roleIntrinsic, "_export", halExport)
	g.registerRole(roleIntrinsic, "_load", halLoad)
	g.registerRole(roleIntrinsic, "_spawn", halSpawn)
	g.registerRole(roleIntrinsic, "_channel", halChannel)
	g.registerRole(roleIntrinsic, "_send", halSend)
	g.registerRole(roleIntrinsic, "_recv", halRecv)

	// Language/value primitives implemented natively.
	for name, fn := range map[string]BuiltinFunc{
		"_first": halFirst, "_rest": halRest, "_length": halLength,
		"_prepend": halPrepend, "_append": halAppend, "_empty": halEmpty, "_range": halRange,
		"_type": halType, "_stack_limit": halStackLimit, "_inspect": halInspect,
		"_equal": halEqual, "_ord": halOrd, "_chr": halChr,
		"_floor": halFloor, "_ceil": halCeil, "_truncate": halTruncate, "_modulo": halModulo,
		"_bytes_new": halBytesNew, "_bytes_length": halBytesLength, "_bytes_get": halBytesGet,
		"_bytes_slice": halBytesSlice, "_str_to_bytes": halStrToBytes,
		"_bytes_to_str": halBytesToStr, "_bytes_to_str_pure": halBytesToStrPure,
		"_shape": halShape, "_make_shaped_list": halMakeShapedList, "_to_str": halToStr,
		"_to_decimal": halToDecimal, "_to_number": halToNumber, "_to_symbol": halToSymbol,
		"_store_new": halStoreNew, "_store_get": halStoreGet, "_store_set": halStoreSet,
		"_store_length": halStoreLength,
		"_bits_and":     halBitsAnd, "_bits_or": halBitsOr, "_bits_xor": halBitsXor,
		"_bits_not": halBitsNot, "_bits_shl": halBitsShl, "_bits_shr": halBitsShr,
	} {
		g.registerRole(roleNative, name, fn)
	}

	// Native/FFI library providers. Native realization does not imply host
	// authority; these are alternate implementations of library behavior.
	for name, fn := range map[string]BuiltinFunc{
		"_sqrt_inexact": halSqrt, "_cos_inexact": halCos, "_sin_inexact": halSin,
		"_upper": halUpper, "_lower": halLower, "_chars": halChars,
		"_upper_rune": halUpperRune, "_lower_rune": halLowerRune,
		"_regex_match": halRegexMatch, "_regex_find": halRegexFind,
		"_regex_find_all": halRegexFindAll, "_regex_replace": halRegexReplace,
		"_regex_replace_first": halRegexReplaceFirst, "_regex_split": halRegexSplit,
	} {
		g.registerRole(roleProvider, name, fn)
	}

	// Runtime/tooling/session services.
	for name, fn := range map[string]BuiltinFunc{
		"_module_roots":   g.halModuleRoots,
		"_profile_counts": halProfileCounts, "_profile_measure": halProfileMeasure,
		"_profile_experiment": halProfileExperiment,
		"_quit":               g.halQuit, "_reset": g.halReset, "_delete": g.halDelete, "_help": g.halHelp, "_doc": g.halDoc,
		"_test_equal": g.halTestEqual, "_test_not_equal": g.halTestNotEqual,
		"_test_true": g.halTestTrue, "_test_false": g.halTestFalse,
		"_test_error": g.halTestError, "_test_not_error": g.halTestNotError,
		"_test_run": g.halTestRun, "_test_faults": g.halTestFaults,
	} {
		g.registerRole(roleService, name, fn)
	}
}
