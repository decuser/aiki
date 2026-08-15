package invariant

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type vscodeManifest struct {
	Main        string            `json:"main"`
	DependsOn   map[string]string `json:"dependencies"`
	Contributes struct {
		Languages []struct {
			ID         string   `json:"id"`
			Extensions []string `json:"extensions"`
		} `json:"languages"`
		Configuration struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"configuration"`
	} `json:"contributes"`
}

func TestVSCodeClientIsThinLSPAdapter(t *testing.T) {
	data, err := os.ReadFile("../../extra/editors/vscode/package.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest vscodeManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Main != "./extension.js" {
		t.Fatalf("VS Code main=%q, want ./extension.js", manifest.Main)
	}
	if _, ok := manifest.DependsOn["vscode-languageclient"]; !ok {
		t.Fatal("VS Code client must depend on vscode-languageclient")
	}
	if len(manifest.Contributes.Languages) != 1 || manifest.Contributes.Languages[0].ID != "aiki" {
		t.Fatalf("VS Code language registration=%#v", manifest.Contributes.Languages)
	}
	if !containsString(manifest.Contributes.Languages[0].Extensions, ".ai") {
		t.Fatalf("VS Code Aiki extension registration=%v, want .ai", manifest.Contributes.Languages[0].Extensions)
	}
	if _, ok := manifest.Contributes.Configuration.Properties["aiki.server.path"]; !ok {
		t.Fatal("VS Code client must expose aiki.server.path for desktop PATH-independent launch")
	}

	source, err := os.ReadFile("../../extra/editors/vscode/extension.js")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, needle := range []string{
		"vscode-languageclient/node",
		"new LanguageClient(",
		"get('server.path', 'aiki')",
		"args: ['lsp']",
		"documentSelector: [{ scheme: 'file', language: 'aiki' }]",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("VS Code client missing thin-adapter element %q", needle)
		}
	}
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
