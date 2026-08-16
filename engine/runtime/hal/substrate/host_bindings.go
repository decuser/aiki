package substrate

import (
	"sort"

	"aiki/engine/runtime/hal"
)

var hostOperationDescriptors = map[string]hal.HostOperation{
	"_print": hostOperation("HAL.io.print", "_print", "go:fmt.Fprint/io.Writer", "runtime.io", "may block", "fault", bindings(
		binding("print", "engine/runtime/prelude/prelude.ai"),
		binding("println", "engine/runtime/prelude/prelude.ai"),
	)),
	"_read": hostOperation("HAL.io.read", "_read", "go:bufio.Reader.ReadString/io.Reader", "runtime.io", "may block", "fault or shaped :io/:end result", bindings(
		binding("read", "engine/runtime/prelude/prelude.ai"),
		binding("input", "engine/runtime/prelude/prelude.ai"),
	)),
	"_sleep": hostOperation("HAL.time.sleep", "_sleep", "go:time.Sleep", "runtime.clock", "waits externally", "fault", bindings(
		binding("time.sleep", "lib/time/time.ai"),
	)),
	"_after": {
		Identity:            "HAL.time.after",
		Primitive:           "_after",
		Authority:           "HAL.time.after",
		Context:             []string{"runtime.clock"},
		Effect:              "async source",
		Blocking:            "nonblocking",
		Lifetime:            "returns resource",
		Optionality:         "constitutive",
		ErrorContract:       "fault",
		SemanticObservation: "time.after",
		SubstrateProvenance: "go:time.AfterFunc/value.NewEventChannel",
		AikiBindings:        bindings(binding("time.after", "lib/time/time.ai")),
	},
	"_system_args": hostOperation("HAL.system.args", "_system_args", "go:runtime-owned programArgs", "runtime.process", "nonblocking", "fault", bindings(
		binding("system.args", "lib/system/system.ai"),
	)),
	"_system_env": hostOperation("HAL.system.env", "_system_env", "go:os.LookupEnv", "runtime.process", "nonblocking", "fault or shaped :environment result", bindings(
		binding("system.env", "lib/system/system.ai"),
	)),
	"_file_open":        resourceAcquisition("HAL.file.open", "_file_open", "go:os.Open/os.Create/os.OpenFile", "file.open", binding("file.open", "lib/file/file.ai")),
	"_file_read_text":   resourceOperation("HAL.file.read_text", "_file_read_text", "go:io.ReadAll(*os.File)", "file.read_text", binding("file.read_text", "lib/file/file.ai")),
	"_file_read_bytes":  resourceOperation("HAL.file.read_bytes", "_file_read_bytes", "go:io.ReadAll(*os.File)", "file.read_bytes", binding("file.read_bytes", "lib/file/file.ai")),
	"_file_read_line":   resourceOperation("HAL.file.read_line", "_file_read_line", "go:bufio.Reader.ReadString(*os.File)", "file.read_line", binding("file.read_line", "lib/file/file.ai")),
	"_file_write_text":  resourceOperation("HAL.file.write_text", "_file_write_text", "go:os.File.WriteString", "file.write_text", binding("file.write_text", "lib/file/file.ai")),
	"_file_write_bytes": resourceOperation("HAL.file.write_bytes", "_file_write_bytes", "go:os.File.Write", "file.write_bytes", binding("file.write_bytes", "lib/file/file.ai")),
	"_file_close":       resourceOperation("HAL.file.close", "_file_close", "go:os.File.Close", "file.close", binding("file.close", "lib/file/file.ai")),
	"_file_exists":      pathOperation("HAL.file.exists", "_file_exists", "go:os.Stat", "file.exists", binding("file.exists", "lib/file/file.ai")),
	"_file_delete":      pathOperation("HAL.file.delete", "_file_delete", "go:os.Remove", "file.delete", binding("file.delete", "lib/file/file.ai")),
	"_file_list":        pathOperation("HAL.file.list", "_file_list", "go:os.ReadDir", "file.list", binding("file.list", "lib/file/file.ai")),
	"_file_read_at":     resourceOperation("HAL.file.read_at", "_file_read_at", "go:os.File.ReadAt", "file.read_at", binding("file.read_at", "lib/file/file.ai")),
	"_file_write_at":    resourceOperation("HAL.file.write_at", "_file_write_at", "go:os.File.WriteAt", "file.write_at", binding("file.write_at", "lib/file/file.ai")),
	"_file_stat":        pathOperation("HAL.file.stat", "_file_stat", "go:os.Stat", "file.stat", binding("file.stat", "lib/file/file.ai")),
	"_file_rename":      pathOperation("HAL.file.rename", "_file_rename", "go:os.Rename", "file.rename", binding("file.rename", "lib/file/file.ai")),
	"_file_mkdir":       pathOperation("HAL.file.mkdir", "_file_mkdir", "go:os.Mkdir", "file.mkdir", binding("file.mkdir", "lib/file/file.ai")),
	"_file_mkdir_all":   pathOperation("HAL.file.mkdir_all", "_file_mkdir_all", "go:os.MkdirAll", "file.mkdir_all", binding("file.mkdir_all", "lib/file/file.ai")),
	"_file_remove_all":  pathOperation("HAL.file.remove_all", "_file_remove_all", "go:os.RemoveAll", "file.remove_all", binding("file.remove_all", "lib/file/file.ai")),
	"_file_temp":        pathOperation("HAL.file.temp", "_file_temp", "go:os.CreateTemp", "file.temp", binding("file.temp", "lib/file/file.ai")),
	"_file_temp_dir":    pathOperation("HAL.file.temp_dir", "_file_temp_dir", "go:os.MkdirTemp", "file.temp_dir", binding("file.temp_dir", "lib/file/file.ai")),
	"_file_copy":        pathOperation("HAL.file.copy", "_file_copy", "go:os.Open/os.Create/io.Copy", "file.copy", binding("file.copy", "lib/file/file.ai")),
	"_file_size":        pathOperation("HAL.file.size", "_file_size", "go:os.Stat", "file.size", binding("file.size", "lib/file/file.ai")),
	"_time_now":         hostOperation("HAL.time.now", "_time_now", "go:time.Now", "runtime.clock", "nonblocking", "fault", bindings(binding("time.now", "lib/time/time.ai"))),
	"_system_cwd":       hostOperation("HAL.system.cwd", "_system_cwd", "go:runtime working-directory context", "runtime.process", "nonblocking", "fault", bindings(binding("system.cwd", "lib/system/system.ai"))),
	"_system_chdir":     hostOperation("HAL.system.chdir", "_system_chdir", "go:os.Stat + runtime working-directory context", "runtime.process", "may block", "fault or shaped :io result", bindings(binding("system.chdir", "lib/system/system.ai"))),
	"_path_separator":   hostOperation("HAL.path.separator", "_path_separator", "go:os.PathSeparator", "runtime.platform", "nonblocking", "fault", bindings(binding("path.separator", "lib/path/path.ai"))),
	"_canvas": {
		Identity:            "HAL.canvas.open",
		Primitive:           "_canvas",
		Authority:           "HAL.canvas.open",
		Context:             []string{"runtime.resources", "runtime.graphics"},
		Effect:              "resource acquisition",
		Blocking:            "waits externally",
		Lifetime:            "returns resource",
		Optionality:         "optional",
		ErrorContract:       "fault or shaped :canvas result",
		SemanticObservation: "canvas.canvas",
		SubstrateProvenance: "go:canvas session/IPC/Ebiten",
		AikiBindings:        bindings(binding("canvas.canvas", "lib/canvas/canvas.ai")),
	},
	"_canvas_command": {
		Identity:            "HAL.canvas.command",
		Primitive:           "_canvas_command",
		Authority:           "HAL.canvas.command",
		Context:             []string{"runtime.resources", "runtime.graphics"},
		Effect:              "bounded call",
		Blocking:            "may block",
		Lifetime:            "operates on resource",
		Optionality:         "optional",
		ErrorContract:       "fault or shaped :canvas result",
		SemanticObservation: "canvas.command",
		SubstrateProvenance: "go:CanvasCmd/Canvas IPC bridge",
		AikiBindings: bindings(
			binding("canvas.dot", "lib/canvas/canvas.ai"),
			binding("canvas.line", "lib/canvas/canvas.ai"),
			binding("canvas.rect", "lib/canvas/canvas.ai"),
			binding("canvas.fill_rect", "lib/canvas/canvas.ai"),
			binding("canvas.circle", "lib/canvas/canvas.ai"),
			binding("canvas.fill_circle", "lib/canvas/canvas.ai"),
			binding("canvas.arc", "lib/canvas/canvas.ai"),
			binding("canvas.clear", "lib/canvas/canvas.ai"),
			binding("canvas.set_bg", "lib/canvas/canvas.ai"),
			binding("canvas.set_fg", "lib/canvas/canvas.ai"),
			binding("canvas.pen_size", "lib/canvas/canvas.ai"),
		),
	},
	"_destroy":       canvasResourceOperation("HAL.canvas.close", "_destroy", "go:canvas bridge shutdown/process reap", "canvas.destroy", binding("canvas.destroy", "lib/canvas/canvas.ai")),
	"_canvas_width":  canvasResourceOperation("HAL.canvas.width", "_canvas_width", "go:runtime canvas resource metadata", "canvas.width", binding("canvas.width", "lib/canvas/canvas.ai")),
	"_canvas_height": canvasResourceOperation("HAL.canvas.height", "_canvas_height", "go:runtime canvas resource metadata", "canvas.height", binding("canvas.height", "lib/canvas/canvas.ai")),
	"_canvas_alive":  canvasResourceOperation("HAL.canvas.alive", "_canvas_alive", "go:runtime canvas resource lifecycle", "canvas.alive", binding("canvas.alive", "lib/canvas/canvas.ai")),
	"_system_exec": {
		Identity:            "HAL.process.exec",
		Primitive:           "_system_exec",
		Authority:           "HAL.process.exec",
		Context:             []string{"runtime.process"},
		Effect:              "bounded call",
		Blocking:            "waits externally",
		Lifetime:            "stateless",
		Optionality:         "constitutive",
		ErrorContract:       "fault or shaped :process result",
		SemanticObservation: "system.exec",
		SubstrateProvenance: "go:os/exec.Command + pipes/wait",
		AikiBindings:        bindings(binding("system.exec", "lib/system/system.ai")),
	},
}

func binding(name, source string) hal.AikiBinding {
	return hal.AikiBinding{Name: name, Source: source}
}

func bindings(bs ...hal.AikiBinding) []hal.AikiBinding { return bs }

func hostOperation(identity, primitive, provenance, context, blocking, errors string, aiki []hal.AikiBinding) hal.HostOperation {
	observation := ""
	if len(aiki) > 0 {
		observation = aiki[0].Name
	}
	return hal.HostOperation{
		Identity:            identity,
		Primitive:           primitive,
		Authority:           identity,
		Context:             []string{context},
		Effect:              "bounded call",
		Blocking:            blocking,
		Lifetime:            "stateless",
		Optionality:         "constitutive",
		ErrorContract:       errors,
		SemanticObservation: observation,
		SubstrateProvenance: provenance,
		AikiBindings:        aiki,
	}
}

func fileOperation(identity, primitive, provenance, observation string, b hal.AikiBinding) hal.HostOperation {
	op := hostOperation(identity, primitive, provenance, "runtime.filesystem", "may block", "fault or shaped :io result", bindings(b))
	op.SemanticObservation = observation
	return op
}

func pathOperation(identity, primitive, provenance, observation string, b hal.AikiBinding) hal.HostOperation {
	return fileOperation(identity, primitive, provenance, observation, b)
}

func resourceAcquisition(identity, primitive, provenance, observation string, b hal.AikiBinding) hal.HostOperation {
	op := fileOperation(identity, primitive, provenance, observation, b)
	op.Effect = "resource acquisition"
	op.Lifetime = "returns resource"
	return op
}

func resourceOperation(identity, primitive, provenance, observation string, b hal.AikiBinding) hal.HostOperation {
	op := fileOperation(identity, primitive, provenance, observation, b)
	op.Lifetime = "operates on resource"
	return op
}

func canvasResourceOperation(identity, primitive, provenance, observation string, b hal.AikiBinding) hal.HostOperation {
	op := hostOperation(identity, primitive, provenance, "runtime.resources", "nonblocking", "fault or shaped :canvas result", bindings(b))
	op.Lifetime = "operates on resource"
	op.Optionality = "optional"
	op.SemanticObservation = observation
	return op
}

// HostOperations returns the canonical host-operation descriptors bound by this
// runtime, sorted by HAL identity. The returned slice is independent of runtime
// execution state and is intended for tooling, validation, and observation.
func (g *GoRuntime) HostOperations() []hal.HostOperation {
	g.mu.RLock()
	out := make([]hal.HostOperation, 0, len(g.hostBindings))
	for _, op := range g.hostBindings {
		copyOp := op
		copyOp.Context = append([]string(nil), op.Context...)
		copyOp.AikiBindings = append([]hal.AikiBinding(nil), op.AikiBindings...)
		out = append(out, copyOp)
	}
	g.mu.RUnlock()

	sort.Slice(out, func(i, j int) bool { return out[i].Identity < out[j].Identity })
	return out
}
