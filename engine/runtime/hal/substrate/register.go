package substrate

// registerHAL installs the runtime binding surface separated by architectural
// responsibility. Only entries registered through registerHost have canonical
// HAL identities; the remaining registries implement non-HAL runtime roles.
func (g *GoRuntime) registerHAL() {
	// True host-facing operations. Canonical contracts currently exist for I/O,
	// time, program context, file operations, and the narrowed Canvas resource/command
	// boundary. Randomness is runtime-owned host state with canonical HAL contracts.
	g.registerHost(goHostOperation("_print"), g.halPrint)
	g.registerHost(goHostOperation("_read"), g.halRead)
	g.registerHost(goHostOperation("_io_read"), g.halIORead)
	g.registerHost(goHostOperation("_io_read_line"), g.halIOReadLine)
	g.registerHost(goHostOperation("_io_write"), g.halIOWrite)
	g.registerHost(goHostOperation("_io_close"), g.halIOClose)
	g.registerHost(goHostOperation("_sleep"), halSleep)
	g.registerHost(goHostOperation("_after"), halAfter)
	g.registerHost(goHostOperation("_system_args"), g.halSystemArgs)
	g.registerHost(goHostOperation("_system_env"), g.halSystemEnv)
	g.registerHost(goHostOperation("_system_environ"), g.halSystemEnviron)
	g.registerHost(goHostOperation("_system_set_env"), g.halSystemSetEnv)
	g.registerHost(goHostOperation("_system_unset_env"), g.halSystemUnsetEnv)
	g.registerHost(goHostOperation("_file_open"), g.halFileOpenPath)
	g.registerHost(goHostOperation("_file_read_text"), halFileReadText)
	g.registerHost(goHostOperation("_file_read_bytes"), halFileReadBytes)
	g.registerHost(goHostOperation("_file_read_line"), g.halFileReadLine)
	g.registerHost(goHostOperation("_file_write_text"), halFileWriteText)
	g.registerHost(goHostOperation("_file_write_bytes"), halFileWriteBytes)
	g.registerHost(goHostOperation("_file_close"), g.halFileClose)
	g.registerHost(goHostOperation("_file_exists"), g.halFileExistsPath)
	g.registerHost(goHostOperation("_file_delete"), g.halFileDeletePath)
	g.registerHost(goHostOperation("_file_list"), g.halFileListPath)
	g.registerHost(goHostOperation("_file_read_at"), halFileReadAt)
	g.registerHost(goHostOperation("_file_write_at"), halFileWriteAt)
	g.registerHost(goHostOperation("_file_stat"), g.halFileStatPath)
	g.registerHost(goHostOperation("_file_rename"), g.halFileRenamePath)
	g.registerHost(goHostOperation("_file_mkdir"), g.halFileMkdirPath)
	g.registerHost(goHostOperation("_file_mkdir_all"), g.halFileMkdirAllPath)
	g.registerHost(goHostOperation("_file_remove_all"), g.halFileRemoveAllPath)
	g.registerHost(goHostOperation("_file_temp"), halFileTemp)
	g.registerHost(goHostOperation("_file_temp_dir"), halFileTempDir)
	g.registerHost(goHostOperation("_file_copy"), g.halFileCopyPath)
	g.registerHost(goHostOperation("_file_size"), g.halFileSizePath)
	g.registerHost(goHostOperation("_file_walk"), g.halFileWalkPath)
	g.registerHost(goHostOperation("_file_symlink"), g.halFileSymlinkPath)
	g.registerHost(goHostOperation("_file_read_link"), g.halFileReadLinkPath)
	g.registerHost(goHostOperation("_file_lock"), g.halFileLock)
	g.registerHost(goHostOperation("_file_try_lock"), g.halFileTryLock)
	g.registerHost(goHostOperation("_file_unlock"), g.halFileUnlock)
	g.registerHost(goHostOperation("_file_permissions"), g.halFilePermissionsPath)
	g.registerHost(goHostOperation("_file_chmod"), g.halFileChmodPath)
	g.registerHost(goHostOperation("_time_now"), halTimeNow)
	g.registerHost(goHostOperation("_system_cwd"), g.halSystemCwd)
	g.registerHost(goHostOperation("_system_chdir"), g.halSystemChdir)
	g.registerHost(goHostOperation("_term_is"), g.halTermIs)
	g.registerHost(goHostOperation("_term_size"), g.halTermSize)
	g.registerHost(goHostOperation("_term_raw"), g.halTermRaw)
	g.registerHost(goHostOperation("_term_restore"), g.halTermRestore)
	g.registerHost(goHostOperation("_system_exec"), g.halSystemExec)
	g.registerHost(goHostOperation("_process_start"), g.halProcessStart)
	g.registerHost(goHostOperation("_process_stdin"), g.halProcessStdin)
	g.registerHost(goHostOperation("_process_stdout"), g.halProcessStdout)
	g.registerHost(goHostOperation("_process_stderr"), g.halProcessStderr)
	g.registerHost(goHostOperation("_process_wait"), g.halProcessWait)
	g.registerHost(goHostOperation("_process_terminate"), g.halProcessTerminate)
	g.registerHost(goHostOperation("_net_connect"), g.halNetConnect)
	g.registerHost(goHostOperation("_net_listen"), g.halNetListen)
	g.registerHost(goHostOperation("_net_accept"), g.halNetAccept)
	g.registerHost(goHostOperation("_net_local"), g.halNetLocal)
	g.registerHost(goHostOperation("_net_remote"), g.halNetRemote)
	g.registerHost(goHostOperation("_net_close"), g.halNetClose)
	g.registerHost(goHostOperation("_net_udp_bind"), g.halNetUDPBind)
	g.registerHost(goHostOperation("_net_udp_send"), g.halNetUDPSend)
	g.registerHost(goHostOperation("_net_udp_recv"), g.halNetUDPRecv)
	g.registerHost(goHostOperation("_signal_watch"), g.halSignalWatch)
	g.registerHost(goHostOperation("_signal_stop"), g.halSignalStop)
	g.registerHost(goHostOperation("_signal_send"), g.halSignalSend)
	g.registerHost(goHostOperation("_path_separator"), halPathSeparator)
	g.registerHost(goHostOperation("_seed"), g.halSeed)
	g.registerHost(goHostOperation("_random"), g.halRandom)
	g.registerHost(goHostOperation("_canvas"), g.halCanvas)
	g.registerHost(goHostOperation("_canvas_command"), g.halCanvasCommand)
	g.registerHost(goHostOperation("_destroy"), g.halDestroy)
	g.registerHost(goHostOperation("_canvas_width"), g.halCanvasWidth)
	g.registerHost(goHostOperation("_canvas_height"), g.halCanvasHeight)
	g.registerHost(goHostOperation("_canvas_alive"), g.halCanvasAlive)

	// Evaluator/language intrinsics.
	g.registerPrimitive("_apply", halApply)
	g.registerPrimitive("_import", g.halImport)
	g.registerPrimitive("_use", g.halUse)
	g.registerPrimitive("_export", halExport)
	g.registerPrimitive("_load", halLoad)
	g.registerPrimitive("_spawn", halSpawn)
	g.registerPrimitive("_channel", halChannel)
	g.registerPrimitive("_send", halSend)
	g.registerPrimitive("_recv", halRecv)

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
		g.registerPrimitive(name, fn)
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
		g.registerPrimitive(name, fn)
	}

	// Runtime/tooling/session services.
	for name, fn := range map[string]BuiltinFunc{
		"_module_roots":   g.halModuleRoots,
		"_system_exit":    g.halSystemExit,
		"_system_has":     g.halSystemHas,
		"_system_require": g.halSystemRequire,
		"_profile_counts": halProfileCounts, "_profile_measure": halProfileMeasure,
		"_profile_experiment": halProfileExperiment,
		"_quit":               g.halQuit, "_reset": g.halReset, "_delete": g.halDelete, "_help": g.halHelp, "_doc": g.halDoc,
		"_test_equal": g.halTestEqual, "_test_not_equal": g.halTestNotEqual,
		"_test_true": g.halTestTrue, "_test_false": g.halTestFalse,
		"_test_error": g.halTestError, "_test_not_error": g.halTestNotError,
		"_test_run": g.halTestRun, "_test_faults": g.halTestFaults,
	} {
		g.registerPrimitive(name, fn)
	}
}
