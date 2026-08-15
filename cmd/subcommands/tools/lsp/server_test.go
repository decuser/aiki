package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"aiki/engine/language"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func framed(v any) []byte {
	b, _ := json.Marshal(v)
	return []byte(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(b), b))
}

func testLSPService(t *testing.T) *language.Service {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	return language.NewService(g)
}

func readFrames(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	tr := newTransport(bytes.NewReader(data), io.Discard)
	var out []map[string]any
	for {
		msg, err := tr.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, _ := json.Marshal(msg)
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		out = append(out, m)
	}
	return out
}

func TestInitializeAndPublishDiagnostics(t *testing.T) {
	var in bytes.Buffer
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai", "languageId": "aiki", "version": 1, "text": "let x =\n"}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "shutdown"}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "exit"}))
	var out bytes.Buffer
	if err := Serve(&in, &out, testLSPService(t)); err != nil {
		t.Fatal(err)
	}
	frames := readFrames(t, out.Bytes())
	if len(frames) != 3 {
		t.Fatalf("frames=%d %#v", len(frames), frames)
	}
	body := string(out.Bytes())
	for _, want := range []string{"positionEncoding", "utf-16", "textDocument/publishDiagnostics", "aiki-parse"} {
		if !strings.Contains(body, want) {
			t.Fatalf("output missing %q: %s", want, body)
		}
	}
}

func TestByteColumnConvertsToUTF16(t *testing.T) {
	p := lspPosition("αx", 1, 3) // α occupies two UTF-8 bytes, one UTF-16 code unit.
	if p["character"] != 1 {
		t.Fatalf("position=%v", p)
	}
	p = lspPosition("😀x", 1, 5) // emoji occupies four UTF-8 bytes, two UTF-16 code units.
	if p["character"] != 2 {
		t.Fatalf("position=%v", p)
	}
}

func TestDefinitionAndDocumentSymbols(t *testing.T) {
	var in bytes.Buffer
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}}))
	src := "let outer = 1\nlet f = () { return outer }\n"
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai", "languageId": "aiki", "version": 1, "text": src}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/definition", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai"}, "position": map[string]any{"line": 1, "character": 20}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 3, "method": "textDocument/documentSymbol", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai"}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 4, "method": "shutdown"}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "exit"}))
	var out bytes.Buffer
	if err := Serve(&in, &out, testLSPService(t)); err != nil {
		t.Fatal(err)
	}
	body := string(out.Bytes())
	for _, want := range []string{"definitionProvider", "documentSymbolProvider", "\"line\":0", "outer", "f"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %s", want, body)
		}
	}
}

func TestFormattingUsesLanguageService(t *testing.T) {
	var in bytes.Buffer
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai", "languageId": "aiki", "version": 1, "text": "let   x=1\n"}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/formatting", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai"}, "options": map[string]any{"tabSize": 4, "insertSpaces": false}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 3, "method": "shutdown"}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "exit"}))
	var out bytes.Buffer
	if err := Serve(&in, &out, testLSPService(t)); err != nil {
		t.Fatal(err)
	}
	body := string(out.Bytes())
	for _, want := range []string{"documentFormattingProvider", `"newText":"let x = 1\n"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %s", want, body)
		}
	}
}

func TestFormattingRejectsInvalidSource(t *testing.T) {
	var in bytes.Buffer
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai", "languageId": "aiki", "version": 1, "text": "let x =\n"}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/formatting", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai"}, "options": map[string]any{}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "exit"}))
	var out bytes.Buffer
	if err := Serve(&in, &out, testLSPService(t)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out.Bytes()), `"code":-32602`) {
		t.Fatalf("expected formatting error response, got %s", out.String())
	}
}

func TestCompletionAndHoverUseLanguageService(t *testing.T) {
	var in bytes.Buffer
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}}))
	src := "let outer = 1\nlet f = (x) { return outer + x }\n"
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai", "languageId": "aiki", "version": 1, "text": src}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/completion", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai"}, "position": map[string]any{"line": 1, "character": 23}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 3, "method": "textDocument/hover", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai"}, "position": map[string]any{"line": 1, "character": 24}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 4, "method": "shutdown"}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "exit"}))
	var out bytes.Buffer
	if err := Serve(&in, &out, testLSPService(t)); err != nil {
		t.Fatal(err)
	}
	body := string(out.Bytes())
	for _, want := range []string{"completionProvider", "hoverProvider", `"label":"outer"`, "defined at 1:5"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %s", want, body)
		}
	}
}

func TestNullableResponsesIncludeExplicitNullResult(t *testing.T) {
	var in bytes.Buffer
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}}))
	src := "let x = 1\n"
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai", "languageId": "aiki", "version": 1, "text": src}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/hover", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai"}, "position": map[string]any{"line": 0, "character": 0}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 3, "method": "textDocument/definition", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/test.ai"}, "position": map[string]any{"line": 0, "character": 0}}}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "id": 4, "method": "shutdown"}))
	in.Write(framed(map[string]any{"jsonrpc": "2.0", "method": "exit"}))
	var out bytes.Buffer
	if err := Serve(&in, &out, testLSPService(t)); err != nil {
		t.Fatal(err)
	}
	body := string(out.Bytes())
	for _, want := range []string{
		`{"jsonrpc":"2.0","id":2,"result":null}`,
		`{"jsonrpc":"2.0","id":3,"result":null}`,
		`{"jsonrpc":"2.0","id":4,"result":null}`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing explicit null response %s in %s", want, body)
		}
	}
}
