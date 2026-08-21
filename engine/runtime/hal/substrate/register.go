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
		"_bytes_digits_from_text": halBytesDigitsFromText, "_bytes_digits_to_text": halBytesDigitsToText,
		"_shape": halShape, "_make_shaped_list": halMakeShapedList, "_to_str": halToStr,
		"_to_decimal": halToDecimal, "_to_number": halToNumber, "_to_symbol": halToSymbol,
		"_store_new": halStoreNew, "_store_get": halStoreGet, "_store_set": halStoreSet,
		"_store_length": halStoreLength, "_store_snapshot": halStoreSnapshot,
		"_fixed_store_new_byte": halFixedStoreNewByte, "_fixed_store_new_word": halFixedStoreNewWord,
		"_fixed_store_new_addr18": halFixedStoreNewAddr18, "_fixed_store_new_counter": halFixedStoreNewCounter, "_fixed_store_get": halFixedStoreGet,
		"_fixed_store_set": halFixedStoreSet, "_fixed_store_length": halFixedStoreLength,
		"_fixed_store_snapshot": halFixedStoreSnapshot, "_fixed_store_word_read_addr": halFixedStoreWordReadAddr,
		"_fixed_store_word_write_addr": halFixedStoreWordWriteAddr, "_fixed_store_byte_read_addr": halFixedStoreByteReadAddr,
		"_fixed_store_byte_write_addr": halFixedStoreByteWriteAddr, "_fixed_store_counter_get": halFixedStoreCounterGet,
		"_fixed_store_counter_set": halFixedStoreCounterSet, "_fixed_store_counter_add": halFixedStoreCounterAdd,
		"_machine_word": halMachineWord, "_machine_byte": halMachineByte, "_machine_addr18": halMachineAddr18,
		"_machine_number": halMachineToNumber, "_machine_same": halMachineSame,
		"_machine_word_add": halMachineWordAdd, "_machine_word_sub": halMachineWordSub,
		"_machine_word_and": halMachineWordAnd, "_machine_word_or": halMachineWordOr,
		"_machine_word_xor": halMachineWordXor, "_machine_word_not": halMachineWordNot,
		"_machine_word_shl": halMachineWordShl, "_machine_word_shr": halMachineWordShr,
		"_machine_word_extract": halMachineWordExtract, "_machine_word_lt": halMachineWordLT,
		"_machine_word_gt": halMachineWordGT, "_machine_word_zero": halMachineWordZero,
		"_machine_word_sign": halMachineWordSign, "_machine_word_low_byte": halMachineWordLowByte,
		"_machine_word_high_byte": halMachineWordHighByte, "_machine_word_with_low_byte": halMachineWordWithLowByte,
		"_machine_word_with_high_byte": halMachineWordWithHighByte, "_machine_byte_to_word": halMachineByteToWord,
		"_machine_byte_sign_word": halMachineByteSignWord, "_machine_byte_zero": halMachineByteZero,
		"_machine_addr_add": halMachineAddrAdd, "_machine_addr_even": halMachineAddrEven,
		"_machine_word_field":         halMachineWordField,
		"_machine_word_mask":          halMachineWordMask,
		"_machine_word_eq_number":     halMachineWordEqNumber,
		"_machine_word_lt_number":     halMachineWordLTNumber,
		"_machine_word_ge_number":     halMachineWordGENumber,
		"_machine_word_add_small":     halMachineWordAddSmall,
		"_machine_word_sub_small":     halMachineWordSubSmall,
		"_machine_word_add_carry":     halMachineWordAddCarry,
		"_machine_word_sub_borrow":    halMachineWordSubBorrow,
		"_machine_word_bit":           halMachineWordBit,
		"_machine_word_any_mask":      halMachineWordAnyMask,
		"_machine_word_set_bit":       halMachineWordSetBit,
		"_machine_zero":               halMachineZero,
		"_machine_negative":           halMachineNegative,
		"_machine_addr_from_word":     halMachineAddrFromWord,
		"_machine_addr_lt_number":     halMachineAddrLTNumber,
		"_machine_addr_ge_number":     halMachineAddrGENumber,
		"_machine_add":                halMachineAdd,
		"_machine_sub":                halMachineSub,
		"_machine_and":                halMachineAnd,
		"_machine_or":                 halMachineOr,
		"_machine_xor":                halMachineXor,
		"_machine_not":                halMachineNot,
		"_machine_add_small":          halMachineAddSmall,
		"_machine_sub_small":          halMachineSubSmall,
		"_machine_lt":                 halMachineLT,
		"_machine_gt":                 halMachineGT,
		"_machine_add_carry_generic":  halMachineAddCarryGeneric,
		"_machine_sub_borrow_generic": halMachineSubBorrowGeneric,
		"_machine_shr1":               halMachineShiftRightOne,
		"_machine_shl1":               halMachineShiftLeftOne,
		"_machine_low_bit":            halMachineLowBit,
		"_machine_set_high_bit":       halMachineSetHighBit,
		"_bits_and":                   halBitsAnd, "_bits_or": halBitsOr, "_bits_xor": halBitsXor,
		"_bits_not": halBitsNot, "_bits_shl": halBitsShl, "_bits_shr": halBitsShr,
	} {
		switch name {
		case "_stack_limit", "_store_get", "_store_set":
			// These operations observe evaluator context for stack state or
			// semantic profiling. Keep the conservative context path.
			g.registerPrimitive(name, fn)
		default:
			g.registerContextFreePrimitive(name, fn)
		}
	}

	// Native/FFI library providers. Native realization does not imply host
	// authority; these are alternate implementations of library behavior.
	for name, fn := range map[string]BuiltinFunc{
		"_sqrt_inexact": halSqrt, "_cos_inexact": halCos, "_sin_inexact": halSin,
		"_upper": halUpper, "_lower": halLower, "_chars": halChars,
		"_string_substring": halStringSubstring, "_string_split": halStringSplit,
		"_string_index_of": halStringIndexOf, "_string_last_index_of": halStringLastIndexOf,
		"_string_contains": halStringContains, "_string_starts_with": halStringStartsWith,
		"_string_ends_with": halStringEndsWith, "_string_join": halStringJoin,
		"_string_replace": halStringReplace, "_string_replace_first": halStringReplaceFirst,
		"_string_repeat": halStringRepeat, "_string_reverse": halStringReverse,
		"_string_trim": halStringTrim, "_string_trim_start": halStringTrimStart,
		"_string_trim_end": halStringTrimEnd, "_string_compare": halStringCompare,
		"_string_is_whitespace": halStringIsWhitespace, "_string_is_digit": halStringIsDigit,
		"_string_is_alpha": halStringIsAlpha, "_string_is_upper": halStringIsUpper,
		"_string_is_lower": halStringIsLower, "_string_is_alnum": halStringIsAlnum,
		"_string_is_numeric": halStringIsNumeric, "_string_is_alphabetic": halStringIsAlphabetic,
		"_hash_code": halHashCode, "_hash_new": halHashNew, "_hash_get": halHashGet,
		"_hash_put": halHashPut, "_hash_has": halHashHas, "_hash_del": halHashDel,
		"_hash_keys": halHashKeys, "_hash_values": halHashValues,
		"_upper_rune": halUpperRune, "_lower_rune": halLowerRune,
		"_regex_match": halRegexMatch, "_regex_find": halRegexFind,
		"_regex_find_all": halRegexFindAll, "_regex_replace": halRegexReplace,
		"_regex_replace_first": halRegexReplaceFirst, "_regex_split": halRegexSplit,
	} {
		g.registerContextFreePrimitive(name, fn)
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
