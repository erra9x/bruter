package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
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

func liveServices(t *testing.T, servicesOutput string) []string {
	t.Helper()
	servicesCountRe := regexp.MustCompile(`(?m)^(\d+) services available$`)
	match := servicesCountRe.FindStringSubmatch(servicesOutput)
	if match == nil {
		t.Fatal("failed to find services count in --list-services output")
	}
	count, err := strconv.Atoi(match[1])
	if err != nil {
		t.Fatalf("parse service count: %v", err)
	}

	lineRe := regexp.MustCompile(`(?m)^([a-z0-9-]+)\s+\d+$`)
	matches := lineRe.FindAllStringSubmatch(servicesOutput, -1)
	services := make([]string, 0, len(matches))
	for _, m := range matches {
		services = append(services, m[1])
	}
	if len(services) != count {
		t.Fatalf("parsed %d services from --list-services, want %d", len(services), count)
	}
	return services
}

func readmeModuleTableServices(t *testing.T, readme string) []string {
	t.Helper()
	sectionRe := regexp.MustCompile(`(?ms)^### Available Modules \(\d+\)\n\n(?P<table>(?:^\|.*\|\n)+)`)
	match := sectionRe.FindStringSubmatch(readme)
	if match == nil {
		t.Fatal("failed to find Available Modules table in README")
	}
	mods := regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(match[1], -1)
	services := make([]string, 0, len(mods))
	for _, m := range mods {
		services = append(services, m[1])
	}
	return services
}

func sortedStrings(items []string) []string {
	out := append([]string(nil), items...)
	sort.Strings(out)
	return out
}

func TestReadmeMatchesLiveCLIHelp(t *testing.T) {
	readme := readmeText(t)
	help := runBruter(t, "--help")
	servicesOutput := runBruter(t, "--list-services")

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

	live := liveServices(t, servicesOutput)
	heading := "### Available Modules (" + strconv.Itoa(len(live)) + ")"
	if !strings.Contains(readme, heading) {
		t.Fatalf("README module count drifted from live CLI\nwant heading: %s", heading)
	}

	readmeMods := readmeModuleTableServices(t, readme)
	if len(readmeMods) != len(live) {
		t.Fatalf("README module table count drifted from live CLI\nreadme=%d live=%d", len(readmeMods), len(live))
	}

	liveSorted := sortedStrings(live)
	readmeSorted := sortedStrings(readmeMods)
	for i := range liveSorted {
		if liveSorted[i] != readmeSorted[i] {
			t.Fatalf("README module table drifted from live CLI\nreadme=%s\nlive=%s", fmt.Sprint(readmeSorted), fmt.Sprint(liveSorted))
		}
	}
}
