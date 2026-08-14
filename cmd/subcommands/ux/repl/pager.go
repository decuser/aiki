package repl

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/chzyer/readline"
)

// newPageOutput returns a presentation hook for long help/documentation text.
// It pages only when out is an actual terminal and the rendered text is taller
// than the current screen. Returning false asks the caller to print normally.
func newPageOutput(out io.Writer) func(string) bool {
	f, ok := out.(*os.File)
	if !ok || !readline.IsTerminal(int(f.Fd())) {
		return nil
	}

	return func(text string) bool {
		width, height, err := readline.GetSize(int(f.Fd()))
		if err != nil || width <= 0 || height <= 0 {
			return false
		}
		if displayLines(text, width) < height {
			return false
		}
		return runPager(text, f)
	}
}

// displayLines estimates the number of terminal rows occupied by text,
// including simple wrapping at the current terminal width. Help/doc text is
// predominantly ASCII; rune counts are sufficient for deciding whether to
// page and avoid entangling paging with terminal rendering semantics.
func displayLines(text string, width int) int {
	if width <= 0 {
		return 0
	}
	lines := strings.Split(text, "\n")
	total := 0
	for _, line := range lines {
		n := utf8.RuneCountInString(line)
		rows := 1
		if n > 0 {
			rows = (n + width - 1) / width
		}
		total += rows
	}
	return total
}

// runPager sends text to $PAGER, defaulting to less -R. The pager receives the
// documentation on stdin and remains attached to the real terminal on stdout.
// If the pager cannot be started, false lets the caller fall back to ordinary
// output. Once it starts, the pager owns presentation even if it later exits
// with a non-zero status.
func runPager(text string, terminal *os.File) bool {
	name, args := pagerCommand(os.Getenv("PAGER"))
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = terminal
	cmd.Stderr = terminal
	if err := cmd.Start(); err != nil {
		return false
	}
	_ = cmd.Wait()
	return true
}

func pagerCommand(env string) (string, []string) {
	fields := strings.Fields(env)
	if len(fields) == 0 {
		return "less", []string{"-R"}
	}
	return fields[0], fields[1:]
}
