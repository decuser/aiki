package substrate

import (
	"os"
	"path/filepath"
	"testing"

	"aiki/engine/semantics/value"
)

func TestFileOpenReadClose(t *testing.T) {
	// Create temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Open
	result := halFileOpen([]value.Value{
		&value.String{Val: path},
		&value.Symbol{Val: "read"},
	}, nil)

	file, ok := result.(*value.File)
	if !ok {
		t.Fatalf("expected *value.File, got %T: %v", result, result)
	}
	if file.Mode != "read" {
		t.Errorf("expected mode 'read', got %s", file.Mode)
	}

	// Read text
	content := halFileReadText([]value.Value{file}, nil)
	s, ok := content.(*value.String)
	if !ok {
		t.Fatalf("expected *value.String, got %T: %v", content, content)
	}
	if s.Val != "hello world" {
		t.Errorf("expected 'hello world', got %q", s.Val)
	}

	// Close
	closeResult := halFileClose([]value.Value{file}, nil)
	if closeResult != value.TRUE {
		t.Errorf("expected TRUE, got %v", closeResult)
	}

	// Double close should error
	closeResult2 := halFileClose([]value.Value{file}, nil)
	if _, ok := closeResult2.(*value.List); !ok {
		t.Errorf("expected shaped error on double close, got %T", closeResult2)
	}
}

func TestFileWriteText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	// Open for write
	result := halFileOpen([]value.Value{
		&value.String{Val: path},
		&value.Symbol{Val: "write"},
	}, nil)

	file, ok := result.(*value.File)
	if !ok {
		t.Fatalf("expected *value.File, got %T: %v", result, result)
	}

	// Write
	writeResult := halFileWriteText([]value.Value{
		file,
		&value.String{Val: "test content"},
	}, nil)
	if writeResult != value.TRUE {
		t.Errorf("expected TRUE, got %v", writeResult)
	}

	// Close
	halFileClose([]value.Value{file}, nil)

	// Verify
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test content" {
		t.Errorf("expected 'test content', got %q", string(data))
	}
}

func TestFileAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "append.txt")

	// Create initial file
	if err := os.WriteFile(path, []byte("line1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Open for append
	result := halFileOpen([]value.Value{
		&value.String{Val: path},
		&value.Symbol{Val: "append"},
	}, nil)

	file, ok := result.(*value.File)
	if !ok {
		t.Fatalf("expected *value.File, got %T: %v", result, result)
	}

	// Append
	halFileWriteText([]value.Value{file, &value.String{Val: "line2\n"}}, nil)
	halFileClose([]value.Value{file}, nil)

	// Verify
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "line1\nline2\n" {
		t.Errorf("expected 'line1\\nline2\\n', got %q", string(data))
	}
}

func TestFileReadLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	if err := os.WriteFile(path, []byte("one\ntwo\nthree"), 0644); err != nil {
		t.Fatal(err)
	}

	result := halFileOpen([]value.Value{
		&value.String{Val: path},
		&value.Symbol{Val: "read"},
	}, nil)

	file := result.(*value.File)

	// Read line 1
	line1 := halFileReadLine([]value.Value{file}, nil)
	if s, ok := line1.(*value.String); !ok || s.Val != "one" {
		t.Errorf("line1: expected 'one', got %v", line1)
	}

	// Read line 2
	line2 := halFileReadLine([]value.Value{file}, nil)
	if s, ok := line2.(*value.String); !ok || s.Val != "two" {
		t.Errorf("line2: expected 'two', got %v", line2)
	}

	// Read line 3 (no trailing newline)
	line3 := halFileReadLine([]value.Value{file}, nil)
	if s, ok := line3.(*value.String); !ok || s.Val != "three" {
		t.Errorf("line3: expected 'three', got %v", line3)
	}

	// Read EOF
	eof := halFileReadLine([]value.Value{file}, nil)
	if list, ok := eof.(*value.List); !ok || list.Shape != "end" {
		t.Errorf("expected [@end], got %v", eof)
	}

	halFileClose([]value.Value{file}, nil)
}

func TestFileReadBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.dat")
	data := []byte{0x00, 0x01, 0x02, 0xFF}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	result := halFileOpen([]value.Value{
		&value.String{Val: path},
		&value.Symbol{Val: "read"},
	}, nil)

	file := result.(*value.File)
	content := halFileReadBytes([]value.Value{file}, nil)

	b, ok := content.(*value.Bytes)
	if !ok {
		t.Fatalf("expected *value.Bytes, got %T", content)
	}
	if len(b.Val) != 4 || b.Val[3] != 0xFF {
		t.Errorf("unexpected bytes: %v", b.Val)
	}

	halFileClose([]value.Value{file}, nil)
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")

	// Should not exist yet
	result := halFileExists([]value.Value{&value.String{Val: path}}, nil)
	if result != value.FALSE {
		t.Errorf("expected FALSE for non-existent, got %v", result)
	}

	// Create it
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should exist now
	result = halFileExists([]value.Value{&value.String{Val: path}}, nil)
	if result != value.TRUE {
		t.Errorf("expected TRUE for existing, got %v", result)
	}
}

func TestFileDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "delete.txt")

	// Create file
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	// Delete
	result := halFileDelete([]value.Value{&value.String{Val: path}}, nil)
	if result != value.TRUE {
		t.Errorf("expected TRUE, got %v", result)
	}

	// Should not exist
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not exist after delete")
	}

	// Delete non-existent should error
	result = halFileDelete([]value.Value{&value.String{Val: path}}, nil)
	if _, ok := result.(*value.List); !ok {
		t.Errorf("expected shaped error, got %T", result)
	}
}

func TestFileOpenInvalidMode(t *testing.T) {
	result := halFileOpen([]value.Value{
		&value.String{Val: "test.txt"},
		&value.Symbol{Val: "invalid"},
	}, nil)

	list, ok := result.(*value.List)
	if !ok || list.Shape != "error" {
		t.Errorf("expected [@error, ...], got %v", result)
	}
}

func TestFileOpenNonExistent(t *testing.T) {
	result := halFileOpen([]value.Value{
		&value.String{Val: "/nonexistent/path/file.txt"},
		&value.Symbol{Val: "read"},
	}, nil)

	list, ok := result.(*value.List)
	if !ok || list.Shape != "error" {
		t.Errorf("expected [@error, ...], got %v", result)
	}
}
