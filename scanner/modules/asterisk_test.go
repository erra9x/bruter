package modules

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

// mockAMIServer starts a minimal Asterisk AMI server.
// responseBlock is sent after the login action (e.g. "Response: Success\r\nMessage: Authentication accepted\r\n\r\n").
func mockAMIServer(t *testing.T, responseBlock string) (string, int) {
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

		// Send AMI banner.
		_, _ = fmt.Fprintf(conn, "Asterisk Call Manager/5.0.0\r\n")

		// Read login action (drain until double CRLF).
		buf := make([]byte, 512)
		for {
			n, err := conn.Read(buf)
			if err != nil || n == 0 {
				return
			}
			if containsStr(string(buf[:n]), "\r\n\r\n") {
				break
			}
		}
		_, _ = fmt.Fprint(conn, responseBlock)
	}()

	addr := ln.Addr().(*net.TCPAddr)
	return "127.0.0.1", addr.Port
}

func TestAsteriskHandler_AuthSuccess(t *testing.T) {
	host, port := mockAMIServer(t, "Response: Success\r\nMessage: Authentication accepted\r\n\r\n")
	target := &Target{IP: net.ParseIP(host), Port: port, OriginalTarget: host}
	ok, err := AsteriskHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "admin", Password: "admin"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected ok=true for AMI success response")
	}
}

func TestAsteriskHandler_AuthFailure(t *testing.T) {
	host, port := mockAMIServer(t, "Response: Error\r\nMessage: Authentication failed\r\n\r\n")
	target := &Target{IP: net.ParseIP(host), Port: port, OriginalTarget: host}
	ok, err := AsteriskHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "admin", Password: "wrong"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected ok=false for AMI error response")
	}
}

func TestAsteriskHandler_NotAMI(t *testing.T) {
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
		_, _ = fmt.Fprintf(conn, "SSH-2.0-OpenSSH_8.9\r\n")
	}()
	addr := ln.Addr().(*net.TCPAddr)
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: addr.Port, OriginalTarget: "127.0.0.1"}
	ok, err := AsteriskHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "admin", Password: "admin"})
	if err == nil {
		t.Error("expected error for non-AMI banner, got nil")
	}
	if ok {
		t.Error("ok should be false for non-AMI server")
	}
}

func TestAsteriskHandler_DialFailure(t *testing.T) {
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: 19972, OriginalTarget: "127.0.0.1"}
	_, err := AsteriskHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "admin", Password: "admin"})
	if err == nil {
		t.Error("expected connection error, got nil")
	}
}
