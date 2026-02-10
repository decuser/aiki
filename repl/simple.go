package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type simpleReader struct {
	scanner *bufio.Scanner
}

// NewSimpleReader creates a basic LineReader with no editing.
func NewSimpleReader() LineReader {
	return &simpleReader{scanner: bufio.NewScanner(os.Stdin)}
}

func (r *simpleReader) Prompt(prompt string) (string, error) {
	fmt.Print(prompt)
	if r.scanner.Scan() {
		return r.scanner.Text(), nil
	}
	if err := r.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func (r *simpleReader) Close() error {
	return nil
}
