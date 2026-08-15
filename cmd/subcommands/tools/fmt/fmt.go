package fmt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"aiki/cmd/internal/testfixture"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Config controls formatting behavior.
type Config struct {
	// Write rewrites files in place.
	Write bool
	// ListOnly lists files that would change; no writes.
	ListOnly bool
	// PrintToStdout prints formatted output; no writes.
	PrintToStdout bool
	// Backup creates a .bak file before overwriting when changes occur.
	Backup bool
}

// FormatPath formats a file, directory pattern (./...), or a single path.
// Returns true if any file changed (or would change in ListOnly mode).
func FormatPath(path string, cfg Config) (bool, error) {
	if !strings.HasSuffix(path, "/...") && !strings.HasSuffix(path, string(filepath.Separator)+"...") {
		skip, err := testfixture.IsParseNegative(path)
		if err != nil {
			return false, err
		}
		if skip {
			return false, nil
		}
	}
	if strings.HasSuffix(path, string(filepath.Separator)+"...") || strings.HasSuffix(path, "/...") {
		dir := strings.TrimSuffix(path, "/...")
		dir = strings.TrimSuffix(dir, string(filepath.Separator)+"...")
		if dir == "" {
			dir = "."
		}
		return formatDir(dir, cfg)
	}
	return formatFile(path, cfg, os.Stdout)
}

func loadGrammar() (*grammar.Grammar, error) {
	return grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
}

func formatDir(dir string, cfg Config) (bool, error) {
	changedAny := false
	var formatErrs []error
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".ai") {
			return nil
		}
		skip, ferr := testfixture.IsParseNegative(path)
		if ferr != nil {
			return ferr
		}
		if skip {
			return nil
		}
		changed, ferr := formatFile(path, cfg, io.Discard)
		if ferr != nil {
			formatErrs = append(formatErrs, fmt.Errorf("%s: %w", path, ferr))
			return nil
		}
		if changed {
			changedAny = true
		}
		return nil
	})
	if walkErr != nil {
		return changedAny, walkErr
	}
	return changedAny, errors.Join(formatErrs...)
}

func formatFile(path string, cfg Config, stdout io.Writer) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	g, err := loadGrammar()
	if err != nil {
		return false, err
	}

	formatted, err := FormatSource(g, path, string(data))
	if err != nil {
		return false, err
	}

	changed := formatted != string(data)

	if cfg.PrintToStdout {
		if !changed {
			// Still print the current file when asked.
			_, _ = stdout.Write(data)
			return false, nil
		}
		_, _ = io.Copy(stdout, bytes.NewBufferString(formatted))
		return true, nil
	}

	if cfg.ListOnly {
		if changed {
			fmt.Fprintln(os.Stdout, path)
			return true, nil
		}
		return false, nil
	}

	if cfg.Write {
		if !changed {
			return false, nil
		}

		// Print changed filename.
		fmt.Fprintln(os.Stdout, path)

		if cfg.Backup {
			bak := path + ".bak"
			if err := os.WriteFile(bak, data, 0644); err != nil {
				return false, err
			}
		}

		if err := atomicWriteFile(path, []byte(formatted), 0644); err != nil {
			return false, err
		}
	}
	return changed, nil
}
