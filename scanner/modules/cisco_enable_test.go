package modules

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

// mockCiscoEnableServer simulates the full enable flow: login → user mode → enable → privileged mode.
// loginResp: response after login password (e.g. "Router>" for success, "% Login invalid" for failure)
// enableResp: response after enable password (e.g. "Router#" or "% Access denied")
func mockCiscoEnableServer(t *testing.T, loginResp, enableResp string) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

		buf := make([]byte, 256)

		_, _ = fmt.Fprintf(conn, "\r\nUsername: ")
		conn.Read(buf) //nolint:errcheck

		_, _ = fmt.Fprintf(conn, "Password: ")
		conn.Read(buf) //nolint:errcheck

		_, _ = fmt.Fprintf(conn, "%s", loginResp)

		if loginResp == "Router>" {
			// Wait for "enable\r\n"
			conn.Read(buf) //nolint:errcheck

			_, _ = fmt.Fprintf(conn, "Password: ")
			conn.Read(buf) //nolint:errcheck

			_, _ = fmt.Fprintf(conn, "%s", enableResp)
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	return "127.0.0.1", addr.Port
}

func TestCiscoEnableHandler_Success(t *testing.T) {
	host, port := mockCiscoEnableServer(t, "Router>", "Router#")
	target := &Target{
		IP:             net.ParseIP(host),
		Port:           port,
		OriginalTarget: host,
	}
	ok, err := CiscoEnableHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "cisco", Password: "enablesecret"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected ok=true when privileged-mode prompt received")
	}
}

func TestCiscoEnableHandler_WrongEnablePassword(t *testing.T) {
	host, port := mockCiscoEnableServer(t, "Router>", "% Access denied\r\nPassword: ")
	target := &Target{
		IP:             net.ParseIP(host),
		Port:           port,
		OriginalTarget: host,
	}
	ok, err := CiscoEnableHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "cisco", Password: "wrong"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected ok=false for wrong enable password")
	}
}

func TestCiscoEnableHandler_LoginFailed(t *testing.T) {
	host, port := mockCiscoEnableServer(t, "% Login invalid\r\n\r\nUsername: ", "")
	target := &Target{
		IP:             net.ParseIP(host),
		Port:           port,
		OriginalTarget: host,
	}
	ok, err := CiscoEnableHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "cisco", Password: "cisco"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected ok=false when login fails")
	}
}

func TestCiscoEnableHandler_DialFailure(t *testing.T) {
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: 19987, OriginalTarget: "127.0.0.1"}
	_, err := CiscoEnableHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "cisco", Password: "cisco"})
	if err == nil {
		t.Error("expected connection error, got nil")
	}
}
