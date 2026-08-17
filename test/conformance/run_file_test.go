package conformance

import (
	"bytes"
	"os/exec"
)

func runAikiFile(exe, path, workDir string) (stdout, stderr string, exitCode int, err error) {
	cmd := exec.Command(exe, path)
	cmd.Dir = workDir

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout, stderr, ee.ExitCode(), nil
		}
		return stdout, stderr, 1, runErr
	}

	return stdout, stderr, 0, nil
}
