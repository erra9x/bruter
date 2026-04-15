package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func runBruter(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func readmeText(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	return string(data)
}

func TestReadmeMatchesLiveCLIHelp(t *testing.T) {
	readme := readmeText(t)
	help := runBruter(t, "--help")
	services := runBruter(t, "--list-services")

	commandsLineRe := regexp.MustCompile(`(?m)^Commands: .+$`)
	commandsLine := commandsLineRe.FindString(help)
	if commandsLine == "" {
		t.Fatal("failed to find Commands line in --help output")
	}
	if !strings.Contains(readme, commandsLine) {
		t.Fatalf("README command list drifted from live CLI\nwant line: %s", commandsLine)
	}

	threadsLineRe := regexp.MustCompile(`(?m)^\s*-c, --concurrent-threads=\d+\s+Number of parallel threads per service$`)
	threadsLine := threadsLineRe.FindString(help)
	if threadsLine == "" {
		t.Fatal("failed to find concurrent-threads line in --help output")
	}
	if !strings.Contains(readme, threadsLine) {
		t.Fatalf("README flag defaults drifted from live CLI\nwant line: %s", threadsLine)
	}

	servicesCountRe := regexp.MustCompile(`(?m)^(\d+) services available$`)
	match := servicesCountRe.FindStringSubmatch(services)
	if match == nil {
		t.Fatal("failed to find services count in --list-services output")
	}
	count, err := strconv.Atoi(match[1])
	if err != nil {
		t.Fatalf("parse service count: %v", err)
	}
	heading := "### Available Modules (" + strconv.Itoa(count) + ")"
	if !strings.Contains(readme, heading) {
		t.Fatalf("README module count drifted from live CLI\nwant heading: %s", heading)
	}
}
