package logx

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerWritesToExpectedStreams(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer

	logger := New(&out, &errBuf)
	logger.Info("hello")
	logger.Success("done")
	logger.Warn("watch")
	logger.Error("boom")

	outText := out.String()
	errText := errBuf.String()

	if !strings.Contains(outText, "[INFO] hello\n") {
		t.Fatalf("expected info in stdout, got %q", outText)
	}
	if !strings.Contains(outText, "[SUCCESS] done\n") {
		t.Fatalf("expected success in stdout, got %q", outText)
	}
	if strings.Contains(outText, "[WARN]") || strings.Contains(outText, "[ERROR]") {
		t.Fatalf("expected warn/error not to be written to stdout, got %q", outText)
	}

	if !strings.Contains(errText, "[WARN] watch\n") {
		t.Fatalf("expected warn in stderr, got %q", errText)
	}
	if !strings.Contains(errText, "[ERROR] boom\n") {
		t.Fatalf("expected error in stderr, got %q", errText)
	}
}

func TestLoggerDoesNotColorizeNonTTYByDefault(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer

	logger := New(&out, &errBuf)
	logger.Info("plain")

	if strings.Contains(out.String(), "\033[") {
		t.Fatalf("expected no ANSI color in non-tty output, got %q", out.String())
	}
}

func TestLoggerForceColor(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer

	logger := New(&out, &errBuf)
	logger.ForceColor = true
	logger.Success("ok")

	text := out.String()
	if !strings.Contains(text, "\033[32m") {
		t.Fatalf("expected success color prefix, got %q", text)
	}
	if !strings.Contains(text, "\033[0m") {
		t.Fatalf("expected color reset suffix, got %q", text)
	}
}
