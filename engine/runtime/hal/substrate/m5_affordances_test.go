package substrate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"aiki/engine/semantics/value"
)

func TestM5FileMetadataDirectoryAndCopyAffordances(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("aiki"), 0644); err != nil {
		t.Fatal(err)
	}

	stat := halFileStat([]value.Value{&value.String{Val: src}}, nil)
	st, ok := stat.(*value.List)
	if !ok || st.Shape != "stat" || len(st.Elements) != 3 {
		t.Fatalf("stat = %#v, want [@stat, size, modified, is_dir]", stat)
	}
	if got := st.Elements[0].Inspect(); got != "4" {
		t.Fatalf("stat size = %s, want 4", got)
	}
	if st.Elements[2] != value.FALSE {
		t.Fatalf("stat is_dir = %s, want false", st.Elements[2].Inspect())
	}

	size := halFileSize([]value.Value{&value.String{Val: src}}, nil)
	if got := size.Inspect(); got != "4" {
		t.Fatalf("size = %s, want 4", got)
	}

	dst := filepath.Join(dir, "copy.txt")
	if got := halFileCopy([]value.Value{&value.String{Val: src}, &value.String{Val: dst}}, nil); got != value.TRUE {
		t.Fatalf("copy = %s", got.Inspect())
	}
	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "aiki" {
		t.Fatalf("copied data = %q, %v", data, err)
	}

	renamed := filepath.Join(dir, "renamed.txt")
	if got := halFileRename([]value.Value{&value.String{Val: dst}, &value.String{Val: renamed}}, nil); got != value.TRUE {
		t.Fatalf("rename = %s", got.Inspect())
	}

	nested := filepath.Join(dir, "one", "two")
	if got := halFileMkdirAll([]value.Value{&value.String{Val: nested}}, nil); got != value.TRUE {
		t.Fatalf("mkdir_all = %s", got.Inspect())
	}
	single := filepath.Join(dir, "single")
	if got := halFileMkdir([]value.Value{&value.String{Val: single}}, nil); got != value.TRUE {
		t.Fatalf("mkdir = %s", got.Inspect())
	}
	if got := halFileRemoveAll([]value.Value{&value.String{Val: filepath.Join(dir, "one")}}, nil); got != value.TRUE {
		t.Fatalf("remove_all = %s", got.Inspect())
	}
}

func TestM5TempAffordancesCreateRealPaths(t *testing.T) {
	fileV := halFileTemp(nil, nil)
	filePath, ok := fileV.(*value.String)
	if !ok {
		t.Fatalf("temp = %T, want string", fileV)
	}
	defer os.Remove(filePath.Val)
	if _, err := os.Stat(filePath.Val); err != nil {
		t.Fatalf("temp path does not exist: %v", err)
	}

	dirV := halFileTempDir(nil, nil)
	dirPath, ok := dirV.(*value.String)
	if !ok {
		t.Fatalf("temp_dir = %T, want string", dirV)
	}
	defer os.RemoveAll(dirPath.Val)
	info, err := os.Stat(dirPath.Val)
	if err != nil || !info.IsDir() {
		t.Fatalf("temp_dir is not a directory: %v", err)
	}
}

func TestM5TimeNowIsMillisecondsSinceEpoch(t *testing.T) {
	before := time.Now().UnixMilli()
	got := halTimeNow(nil, nil)
	after := time.Now().UnixMilli()
	n, ok := got.(*value.Number)
	if !ok || !n.IsInt() || !n.IsInt64() {
		t.Fatalf("time.now = %T %v, want integer number", got, got)
	}
	ms := n.Int64Value()
	if ms < before || ms > after {
		t.Fatalf("time.now = %d, outside [%d,%d]", ms, before, after)
	}
}

func TestM5WorkingDirectoryIsRuntimeOwnedAndDrivesFileBindings(t *testing.T) {
	processCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir1, "only-here"), []byte("one"), 0644); err != nil {
		t.Fatal(err)
	}

	rt1 := NewGoRuntime()
	rt2 := NewGoRuntime()
	if got := rt1.halSystemChdir([]value.Value{&value.String{Val: dir1}}, nil); got != value.TRUE {
		t.Fatalf("rt1 chdir = %s", got.Inspect())
	}
	if got := rt2.halSystemChdir([]value.Value{&value.String{Val: dir2}}, nil); got != value.TRUE {
		t.Fatalf("rt2 chdir = %s", got.Inspect())
	}
	if got := rt1.halFileExistsPath([]value.Value{&value.String{Val: "only-here"}}, nil); got != value.TRUE {
		t.Fatal("rt1 relative file lookup did not use rt1 cwd")
	}
	if got := rt2.halFileExistsPath([]value.Value{&value.String{Val: "only-here"}}, nil); got != value.FALSE {
		t.Fatal("rt2 relative file lookup leaked rt1 cwd")
	}
	if cwd, _ := os.Getwd(); cwd != processCwd {
		t.Fatalf("runtime chdir changed process cwd from %q to %q", processCwd, cwd)
	}
}

func TestM5SystemExecCapturesOutputExitCodeAndRuntimeCwd(t *testing.T) {
	rt := NewGoRuntime()
	dir := t.TempDir()
	if got := rt.halSystemChdir([]value.Value{&value.String{Val: dir}}, nil); got != value.TRUE {
		t.Fatalf("chdir = %s", got.Inspect())
	}

	got := rt.halSystemExec([]value.Value{
		&value.String{Val: os.Args[0]},
		&value.List{Elements: []value.Value{
			&value.String{Val: "-test.run=TestM5ExecHelperProcess"},
			&value.String{Val: "--"},
			&value.String{Val: "7"},
		}},
	}, nil)
	result, ok := got.(*value.List)
	if !ok || result.Shape != "ok" || len(result.Elements) != 3 {
		t.Fatalf("exec = %#v, want [@ok, stdout, stderr, exit_code]", got)
	}
	stdout := result.Elements[0].(*value.String).Val
	stderr := result.Elements[1].(*value.String).Val
	if stdout != dir+"\n" {
		t.Fatalf("stdout = %q, want runtime cwd %q", stdout, dir+"\n")
	}
	if stderr != "helper stderr\n" {
		t.Fatalf("stderr = %q", stderr)
	}
	if result.Elements[2].Inspect() != "7" {
		t.Fatalf("exit code = %s, want 7", result.Elements[2].Inspect())
	}
}

func TestM5ExecHelperProcess(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-2] != "--" {
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		os.Exit(98)
	}
	_, _ = os.Stdout.WriteString(cwd + "\n")
	_, _ = os.Stderr.WriteString("helper stderr\n")
	if os.Args[len(os.Args)-1] == "7" {
		os.Exit(7)
	}
}
