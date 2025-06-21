package graceful

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// handlerFunc is a mock slog.Handler (test only)
type handlerFunc struct {
	fn func(r slog.Record) error
}

func (h handlerFunc) Handle(_ context.Context, r slog.Record) error { return h.fn(r) }
func (h handlerFunc) Enabled(_ context.Context, _ slog.Level) bool  { return true }
func (h handlerFunc) WithAttrs(_ []slog.Attr) slog.Handler          { return h }
func (h handlerFunc) WithGroup(_ string) slog.Handler               { return h }

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
	if _, err := bufOut.ReadFrom(rOut); err != nil {
		t.Fatal(err)
	}
	if _, err := bufErr.ReadFrom(rErr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(bufOut.String(), "info-message: foo") {
		t.Errorf("Infof did not write expected message to stdout: %q", bufOut.String())
	}
	if !strings.Contains(bufErr.String(), "error-message: bar") {
		t.Errorf("Errorf did not write expected message to stderr: %q", bufErr.String())
	}
}

func TestNewSlogLogger_Text(t *testing.T) {
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	logger := NewSlogLogger()
	logger.Infof("info-text: %s", "foo")
	logger.Errorf("error-text: %s", "bar")

	wOut.Close()
	wErr.Close()
	var bufOut, bufErr bytes.Buffer
	if _, err := bufOut.ReadFrom(rOut); err != nil {
		t.Fatal(err)
	}
	if _, err := bufErr.ReadFrom(rErr); err != nil {
		t.Fatal(err)
	}

	// Text handler should print plaintext, not JSON/bracketed
	outStr := bufOut.String()
	if !strings.Contains(outStr, "info-text: foo") {
		t.Errorf("Text mode Infof missing: %q", outStr)
	}
}

func TestNewSlogLogger_Json(t *testing.T) {
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	logger := NewSlogLogger(WithJSON())
	logger.Infof("info-json: %s", "foo")
	logger.Errorf("error-json: %s", "bar")

	wOut.Close()
	wErr.Close()
	var bufOut, bufErr bytes.Buffer
	if _, err := bufOut.ReadFrom(rOut); err != nil {
		t.Fatal(err)
	}
	if _, err := bufErr.ReadFrom(rErr); err != nil {
		t.Fatal(err)
	}

	// JSON handler should output JSON encoded log
	outStr := bufOut.String()
	if !strings.Contains(outStr, "\"msg\":\"info-json: foo\"") {
		t.Errorf("JSON mode Infof missing/invalid: %q", outStr)
	}
}

func TestNewSlogLogger_WithSlog(t *testing.T) {
	var captured []string
	l := slog.New(handlerFunc{fn: func(r slog.Record) error {
		var buf bytes.Buffer
		buf.WriteString(r.Message)
		captured = append(captured, buf.String())
		return nil
	}})
	logger := NewSlogLogger(WithSlog(l))
	logger.Infof("injected: %s", "foo")
	if len(captured) != 1 || captured[0] != "injected: foo" {
		t.Errorf("Custom slog.Logger was not used/injected properly, got %+v", captured)
	}
}
