package graceful

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestNewLogger_Infof_Errorf(t *testing.T) {
	// Save original stdout and stderr
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()
	// Create pipes for capturing
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	logger := NewLogger()
	logger.Infof("info-message: %s", "foo")
	logger.Errorf("error-message: %s", "bar")

	wOut.Close()
	wErr.Close()
	var bufOut, bufErr bytes.Buffer
	bufOut.ReadFrom(rOut)
	bufErr.ReadFrom(rErr)

	if !strings.Contains(bufOut.String(), "info-message: foo") {
		t.Errorf("Infof did not write expected message to stdout: %q", bufOut.String())
	}
	if !strings.Contains(bufErr.String(), "error-message: bar") {
		t.Errorf("Errorf did not write expected message to stderr: %q", bufErr.String())
	}
}
