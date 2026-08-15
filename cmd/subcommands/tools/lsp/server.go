package lsp

import (
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"aiki/engine/language"
	"aiki/engine/semantics/value"
)

type documentState struct {
	URI     string
	Path    string
	Text    string
	Version int
}

type server struct {
	service   *language.Service
	transport *transport
	documents map[string]documentState
	shutdown  bool
}

func Serve(r io.Reader, w io.Writer, service *language.Service) error {
	s := &server{service: service, transport: newTransport(r, w), documents: make(map[string]documentState)}
	for {
		msg, err := s.transport.read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		exit, err := s.handle(msg)
		if err != nil {
			return err
		}
		if exit {
			return nil
		}
	}
}

func (s *server) handle(msg message) (bool, error) {
	switch msg.Method {
	case "initialize":
		return false, s.reply(msg.ID, map[string]any{
			"capabilities": map[string]any{
				"positionEncoding": "utf-16",
				"textDocumentSync": map[string]any{"openClose": true, "change": 1},
			},
			"serverInfo": map[string]any{"name": "aiki"},
		})
	case "initialized":
		return false, nil
	case "shutdown":
		s.shutdown = true
		return false, s.reply(msg.ID, nil)
	case "exit":
		return true, nil
	case "textDocument/didOpen":
		var p struct {
			TextDocument struct {
				URI, LanguageID, Text string
				Version               int
			} `json:"textDocument"`
		}
		if err := json.Unmarshal(msg.Params, &p); err != nil {
			return false, err
		}
		d := documentState{URI: p.TextDocument.URI, Path: pathFromURI(p.TextDocument.URI), Text: p.TextDocument.Text, Version: p.TextDocument.Version}
		s.documents[d.URI] = d
		return false, s.publishDiagnostics(d)
	case "textDocument/didChange":
		var p struct {
			TextDocument struct {
				URI     string
				Version int
			} `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		if err := json.Unmarshal(msg.Params, &p); err != nil {
			return false, err
		}
		d, ok := s.documents[p.TextDocument.URI]
		if !ok || len(p.ContentChanges) == 0 {
			return false, nil
		}
		d.Text = p.ContentChanges[len(p.ContentChanges)-1].Text
		d.Version = p.TextDocument.Version
		s.documents[d.URI] = d
		return false, s.publishDiagnostics(d)
	case "textDocument/didClose":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		if err := json.Unmarshal(msg.Params, &p); err != nil {
			return false, err
		}
		delete(s.documents, p.TextDocument.URI)
		return false, s.transport.write(notification{JSONRPC: "2.0", Method: "textDocument/publishDiagnostics", Params: map[string]any{"uri": p.TextDocument.URI, "diagnostics": []any{}}})
	default:
		if len(msg.ID) != 0 {
			return false, s.transport.write(response{JSONRPC: "2.0", ID: msg.ID, Error: &responseError{Code: -32601, Message: "method not found"}})
		}
		return false, nil
	}
}

func (s *server) reply(id json.RawMessage, result any) error {
	return s.transport.write(response{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *server) publishDiagnostics(d documentState) error {
	doc := language.Document{ID: d.URI, Path: d.Path, Source: d.Text, Version: d.Version}
	diagnostics := s.service.Diagnostics(doc, scopeForDocument(d))
	out := make([]map[string]any, 0, len(diagnostics))
	for _, diag := range diagnostics {
		start := lspPosition(d.Text, diag.Pos.Line, diag.Pos.Col)
		end := nextLSPPosition(d.Text, diag.Pos.Line, diag.Pos.Col)
		severity := 1
		if diag.Severity == "warning" {
			severity = 2
		}
		out = append(out, map[string]any{
			"range":    map[string]any{"start": start, "end": end},
			"severity": severity,
			"source":   "aiki-" + diag.Source,
			"message":  diag.Message,
		})
	}
	params := map[string]any{"uri": d.URI, "version": d.Version, "diagnostics": out}
	return s.transport.write(notification{JSONRPC: "2.0", Method: "textDocument/publishDiagnostics", Params: params})
}

func scopeForDocument(d documentState) value.Scope {
	if language.HasPackageDeclaration(d.Text) {
		return value.ScopePrelude
	}
	return value.ScopeUser
}

func pathFromURI(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "file" {
		return raw
	}
	path, err := url.PathUnescape(u.Path)
	if err != nil {
		return u.Path
	}
	return path
}

func lspPosition(source string, line, byteCol int) map[string]int {
	if line < 1 {
		line = 1
	}
	if byteCol < 1 {
		byteCol = 1
	}
	lines := strings.Split(source, "\n")
	if line > len(lines) {
		line = len(lines)
		if line < 1 {
			line = 1
		}
	}
	text := ""
	if len(lines) > 0 {
		text = lines[line-1]
	}
	byteIndex := byteCol - 1
	if byteIndex > len(text) {
		byteIndex = len(text)
	}
	for byteIndex > 0 && byteIndex < len(text) && !utf8.RuneStart(text[byteIndex]) {
		byteIndex--
	}
	units := 0
	for _, r := range text[:byteIndex] {
		units += len(utf16.Encode([]rune{r}))
	}
	return map[string]int{"line": line - 1, "character": units}
}

func nextLSPPosition(source string, line, byteCol int) map[string]int {
	lines := strings.Split(source, "\n")
	if line < 1 || line > len(lines) {
		return lspPosition(source, line, byteCol)
	}
	text := lines[line-1]
	idx := byteCol - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(text) {
		return lspPosition(source, line, byteCol)
	}
	_, size := utf8.DecodeRuneInString(text[idx:])
	if size < 1 {
		size = 1
	}
	return lspPosition(source, line, byteCol+size)
}
