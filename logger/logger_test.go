package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew_BothQuietAndDebug_ReturnsError(t *testing.T) {
	_, err := New(true, true)
	if err == nil {
		t.Error("Expected error when both quiet and debug are true, got nil")
	}
}

func TestNew_ValidConfigurations(t *testing.T) {
	tests := []struct {
		name  string
		quiet bool
		debug bool
	}{
		{"neither", false, false},
		{"quiet only", true, false},
		{"debug only", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := New(tt.quiet, tt.debug)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if l == nil {
				t.Error("Expected logger instance, got nil")
			}
		})
	}
}

func TestInit_BothQuietAndDebug_ReturnsError(t *testing.T) {
	err := Init(true, true)
	if err == nil {
		t.Error("Expected error when both quiet and debug are true, got nil")
	}
}

func TestLogger_QuietMode(t *testing.T) {
	l, _ := New(true, false)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	// Info should be suppressed
	l.Info("info message")
	if buf.Len() > 0 {
		t.Errorf("Info should be suppressed in quiet mode, got: %s", buf.String())
	}

	// Debug should be suppressed
	l.Debug("debug message")
	if buf.Len() > 0 {
		t.Errorf("Debug should be suppressed in quiet mode, got: %s", buf.String())
	}

	// Success should print without prefix
	buf.Reset()
	l.Success("success message")
	output := buf.String()
	if !strings.Contains(output, "success message") {
		t.Errorf("Success message not found in output: %s", output)
	}
	if strings.Contains(output, "[+]") {
		t.Errorf("Success should not have prefix in quiet mode: %s", output)
	}
	if strings.Contains(output, "[") {
		t.Errorf("Success should not have timestamp in quiet mode: %s", output)
	}
}

func TestLogger_DebugMode(t *testing.T) {
	l, _ := New(false, true)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	// Debug should print with [DEBUG] prefix
	l.Debug("debug message")
	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Debug should have [DEBUG] prefix: %s", output)
	}
	if !strings.Contains(output, "debug message") {
		t.Errorf("Debug message not found in output: %s", output)
	}

	// Info should print with [*] prefix
	buf.Reset()
	l.Info("info message")
	output = buf.String()
	if !strings.Contains(output, "[*]") {
		t.Errorf("Info should have [*] prefix: %s", output)
	}

	// Success should print with [+] prefix
	buf.Reset()
	l.Success("success message")
	output = buf.String()
	if !strings.Contains(output, "[+]") {
		t.Errorf("Success should have [+] prefix: %s", output)
	}
}

func TestLogger_NormalMode(t *testing.T) {
	l, _ := New(false, false)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	// Debug should be suppressed in normal mode
	l.Debug("debug message")
	if buf.Len() > 0 {
		t.Errorf("Debug should be suppressed in normal mode, got: %s", buf.String())
	}

	// Info should print with [*] prefix and timestamp
	buf.Reset()
	l.Info("info message")
	output := buf.String()
	if !strings.Contains(output, "[*]") {
		t.Errorf("Info should have [*] prefix: %s", output)
	}
	if !strings.Contains(output, "info message") {
		t.Errorf("Info message not found in output: %s", output)
	}

	// Success should print with [+] prefix and timestamp
	buf.Reset()
	l.Success("success message")
	output = buf.String()
	if !strings.Contains(output, "[+]") {
		t.Errorf("Success should have [+] prefix: %s", output)
	}
}

func TestLogger_FormattedMethods(t *testing.T) {
	l, _ := New(false, true)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.Infof("count: %d", 42)
	if !strings.Contains(buf.String(), "count: 42") {
		t.Errorf("Infof formatting failed: %s", buf.String())
	}

	buf.Reset()
	l.Debugf("name: %s", "test")
	if !strings.Contains(buf.String(), "name: test") {
		t.Errorf("Debugf formatting failed: %s", buf.String())
	}

	buf.Reset()
	l.Successf("processed %d items", 100)
	if !strings.Contains(buf.String(), "processed 100 items") {
		t.Errorf("Successf formatting failed: %s", buf.String())
	}
}

func TestLogger_TimestampFormat(t *testing.T) {
	l, _ := New(false, false)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.Info("test")
	output := buf.String()

	// Check timestamp format (YYYY-MM-DD HH:MM:SS)
	// Example: 2024-01-15 10:30:45
	if len(output) < 19 {
		t.Errorf("Output too short to contain timestamp: %s", output)
	}

	// Verify timestamp structure
	timestampPart := output[:19]
	if timestampPart[4] != '-' || timestampPart[7] != '-' ||
		timestampPart[10] != ' ' || timestampPart[13] != ':' || timestampPart[16] != ':' {
		t.Errorf("Invalid timestamp format: %s", timestampPart)
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Initialize global logger
	err := Init(false, true)
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	buf := &bytes.Buffer{}
	SetOutput(buf)

	Info("global info")
	if !strings.Contains(buf.String(), "global info") {
		t.Errorf("Global Info failed: %s", buf.String())
	}

	buf.Reset()
	Debug("global debug")
	if !strings.Contains(buf.String(), "global debug") {
		t.Errorf("Global Debug failed: %s", buf.String())
	}

	buf.Reset()
	Success("global success")
	if !strings.Contains(buf.String(), "global success") {
		t.Errorf("Global Success failed: %s", buf.String())
	}
}

// --- Verbose mode tests ---

func TestVerbosef_SuppressedWhenNotEnabled(t *testing.T) {
	l, _ := New(false, false)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	// verbose NOT enabled — nothing should be printed
	l.Verbosef("should not appear: %s", "hidden")

	if buf.Len() > 0 {
		t.Errorf("Verbosef should not print when verbose=false, got: %s", buf.String())
	}
}

func TestVerbosef_PrintsWhenEnabled(t *testing.T) {
	l, _ := New(false, false)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)
	l.SetVerbose(true)

	l.Verbosef("attempt: %s -> %s", "user", "pass")
	output := buf.String()

	if !strings.Contains(output, "attempt: user -> pass") {
		t.Errorf("Verbosef message not found in output: %s", output)
	}
	if !strings.Contains(output, "[VERBOSE]") {
		t.Errorf("Verbosef should have [VERBOSE] prefix: %s", output)
	}
}

func TestVerbosef_TimestampFormat(t *testing.T) {
	l, _ := New(false, false)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)
	l.SetVerbose(true)

	l.Verbosef("ts test")
	output := buf.String()

	// Timestamp must be at the start: "2006-01-02 15:04:05"
	if len(output) < 19 {
		t.Fatalf("Output too short for timestamp: %q", output)
	}
	ts := output[:19]
	if ts[4] != '-' || ts[7] != '-' || ts[10] != ' ' || ts[13] != ':' || ts[16] != ':' {
		t.Errorf("Timestamp format wrong, got: %q", ts)
	}
}

func TestVerbosef_IndependentOfQuietMode(t *testing.T) {
	// verbose should still work even in quiet mode
	l, _ := New(true, false) // quiet=true
	buf := &bytes.Buffer{}
	l.SetOutput(buf)
	l.SetVerbose(true)

	l.Verbosef("quiet+verbose: %d", 42)
	output := buf.String()

	if !strings.Contains(output, "quiet+verbose: 42") {
		t.Errorf("Verbosef should print even in quiet mode when verbose=true, got: %s", output)
	}
}

func TestSetVerbose_Toggle(t *testing.T) {
	l, _ := New(false, false)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	// disabled by default
	l.Verbosef("off")
	if buf.Len() > 0 {
		t.Errorf("should be silent initially, got: %s", buf.String())
	}

	// enable
	l.SetVerbose(true)
	l.Verbosef("on")
	if !strings.Contains(buf.String(), "on") {
		t.Errorf("should print after SetVerbose(true), got: %s", buf.String())
	}

	// disable again
	buf.Reset()
	l.SetVerbose(false)
	l.Verbosef("off-again")
	if buf.Len() > 0 {
		t.Errorf("should be silent after SetVerbose(false), got: %s", buf.String())
	}
}
