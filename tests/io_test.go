package tests

import (
	"os"
	"testing"

	"aiki/lang/eval"
	"aiki/lang/value"
)

func setupIOEnv() *value.Env {
	env := value.NewEnv(nil)
	preludeSource := loadPreludeSourceForTest()
	result := eval.Run(preludeSource, env)
	if _, ok := result.(*value.Error); ok {
		panic("failed to load prelude: " + result.Inspect())
	}
	return env
}

func TestPipeUnwrapOk(t *testing.T) {
	env := setupIOEnv()
	result := eval.Run(`[@ok, 42] |> type()`, env)

	sym, ok := result.(*value.Symbol)
	if !ok {
		t.Fatalf("expected Symbol, got %T: %v", result, result)
	}
	if sym.Value != "number" {
		t.Errorf("got :%s, want :number", sym.Value)
	}
}

func TestPipeUnwrapOkChained(t *testing.T) {
	env := setupIOEnv()
	input := `
let double = (n) { return n * 2 }
let add1 = (n) { return n + 1 }
[@ok, 5] |> double() |> add1()
`
	result := eval.Run(input, env)
	testNumberValue(t, result, "11")
}

func TestPipeErrorShortCircuit(t *testing.T) {
	env := setupIOEnv()
	input := `
let double = (n) { return n * 2 }
[@error, "failed"] |> double() |> double()
`
	result := eval.Run(input, env)

	list, ok := result.(*value.List)
	if !ok {
		t.Fatalf("expected List, got %T: %v", result, result)
	}
	if list.Shape != "error" {
		t.Errorf("expected error shape, got %s", list.Shape)
	}
}

func TestPipeRawValuePassthrough(t *testing.T) {
	env := setupIOEnv()
	input := `
let double = (n) { return n * 2 }
5 |> double() |> double()
`
	result := eval.Run(input, env)
	testNumberValue(t, result, "20")
}

func TestFileCreateWriteClose(t *testing.T) {
	env := setupIOEnv()

	// Clean up before and after
	os.Remove("test_io.txt")
	defer os.Remove("test_io.txt")

	input := `
let h = create("test_io.txt")
fwrite(h, "hello")
fclose(h)
`
	result := eval.Run(input, env)
	if _, ok := result.(*value.Error); ok {
		t.Fatalf("unexpected error: %v", result)
	}

	// Verify file exists and has content
	data, err := os.ReadFile("test_io.txt")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("got %q, want %q", string(data), "hello")
	}
}

func TestFileOpenReadClose(t *testing.T) {
	env := setupIOEnv()

	// Create test file
	os.WriteFile("test_io_read.txt", []byte("aiki test"), 0644)
	defer os.Remove("test_io_read.txt")

	input := `
let h = open("test_io_read.txt")
let content = fread(h)
fclose(h)
content
`
	result := eval.Run(input, env)

	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T: %v", result, result)
	}
	if str.Value != "aiki test" {
		t.Errorf("got %q, want %q", str.Value, "aiki test")
	}
}

func TestFileOpenError(t *testing.T) {
	env := setupIOEnv()
	result := eval.Run(`open("nonexistent_file_12345.txt")`, env)

	list, ok := result.(*value.List)
	if !ok {
		t.Fatalf("expected List (error), got %T: %v", result, result)
	}
	if list.Shape != "error" {
		t.Errorf("expected error shape, got %s", list.Shape)
	}
}

func TestFileReadEOF(t *testing.T) {
	env := setupIOEnv()

	// Create empty file
	os.WriteFile("test_io_empty.txt", []byte{}, 0644)
	defer os.Remove("test_io_empty.txt")

	input := `
let h = open("test_io_empty.txt")
let result = fread(h)
fclose(h)
shape(result)
`
	result := eval.Run(input, env)

	sym, ok := result.(*value.Symbol)
	if !ok {
		t.Fatalf("expected Symbol, got %T: %v", result, result)
	}
	if sym.Value != "end" {
		t.Errorf("got :%s, want :end", sym.Value)
	}
}

func TestFileRoundTrip(t *testing.T) {
	env := setupIOEnv()

	os.Remove("test_io_round.txt")
	defer os.Remove("test_io_round.txt")

	input := `
let h = create("test_io_round.txt")
fwrite(h, "line one\n")
fwrite(h, "line two\n")
fclose(h)

let r = open("test_io_round.txt")
let data = fread(r)
fclose(r)
data
`
	result := eval.Run(input, env)

	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T: %v", result, result)
	}
	expected := "line one\nline two\n"
	if str.Value != expected {
		t.Errorf("got %q, want %q", str.Value, expected)
	}
}
