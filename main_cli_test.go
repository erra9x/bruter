package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("BRUTER_CLI_HELPER") == "1" {
		args := []string{"bruter"}
		for i, arg := range os.Args {
			if arg == "--" {
				args = append(args, os.Args[i+1:]...)
				break
			}
		}
		os.Args = args
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func runBruterCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmdArgs := append([]string{"-test.run=TestCLIHelperProcess", "--"}, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = append(os.Environ(), "BRUTER_CLI_HELPER=1")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestCLIHelperProcess(t *testing.T) {
	if os.Getenv("BRUTER_CLI_HELPER") != "1" {
		return
	}
}

func TestRootCLIHelpSmoke(t *testing.T) {
	output, err := runBruterCLI(t, "--help")
	if err != nil {
		t.Fatalf("bruter --help failed: %v\n%s", err, output)
	}
	for _, want := range []string{
		"usage: bruter",
		"bruter is a network services bruteforce tool.",
		"Commands:",
		"ssh",
		"http-basic",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("bruter --help output missing %q\n%s", want, output)
		}
	}
}

func TestRootCLIVersionSmoke(t *testing.T) {
	output, err := runBruterCLI(t, "--version")
	if err != nil {
		t.Fatalf("bruter --version failed: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(output); got == "" {
		t.Fatalf("bruter --version returned empty output")
	}
}

func TestRootCLIListServicesSmoke(t *testing.T) {
	output, err := runBruterCLI(t, "--list-services")
	if err != nil {
		t.Fatalf("bruter --list-services failed: %v\n%s", err, output)
	}
	for _, want := range []string{
		"SERVICE",
		"DEFAULT PORT",
		"ssh",
		"http-basic",
		"services available",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("bruter --list-services output missing %q\n%s", want, output)
		}
	}
}
