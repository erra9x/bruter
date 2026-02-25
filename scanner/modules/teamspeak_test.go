package modules

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

// mockTS3Server starts a minimal TeamSpeak 3 ServerQuery server.
// errorLine is the response sent after the login command (e.g. "error id=0 msg=ok").
func mockTS3Server(t *testing.T, errorLine string) (string, int) {
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

		// Send TS3 banner (2 lines).
		_, _ = fmt.Fprintf(conn, "TS3\n")
		_, _ = fmt.Fprintf(conn, "Welcome to the TeamSpeak 3 ServerQuery interface\n")

		// Read login command.
		buf := make([]byte, 256)
		conn.Read(buf) //nolint:errcheck

		// Send error response.
		_, _ = fmt.Fprintf(conn, "%s\n", errorLine)
	}()

	addr := ln.Addr().(*net.TCPAddr)
	return "127.0.0.1", addr.Port
}

func TestTeamSpeakHandler_AuthSuccess(t *testing.T) {
	host, port := mockTS3Server(t, "error id=0 msg=ok")
	target := &Target{IP: net.ParseIP(host), Port: port, OriginalTarget: host}
	ok, err := TeamSpeakHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "serveradmin", Password: "admin"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected ok=true for TS3 error id=0 msg=ok")
	}
}

func TestTeamSpeakHandler_AuthFailure(t *testing.T) {
	host, port := mockTS3Server(t, "error id=520 msg=invalid\\spermissions")
	target := &Target{IP: net.ParseIP(host), Port: port, OriginalTarget: host}
	ok, err := TeamSpeakHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "serveradmin", Password: "wrong"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected ok=false for TS3 non-zero error id")
	}
}

func TestTeamSpeakHandler_NotTS3(t *testing.T) {
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
		_, _ = fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
	}()
	addr := ln.Addr().(*net.TCPAddr)
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: addr.Port, OriginalTarget: "127.0.0.1"}
	ok, err := TeamSpeakHandler(context.Background(), newTestDialer(t), 3*time.Second,
		target, &Credential{Username: "serveradmin", Password: "admin"})
	if err == nil {
		t.Error("expected error for non-TS3 banner, got nil")
	}
	if ok {
		t.Error("ok should be false for non-TS3 server")
	}
}

func TestTeamSpeakHandler_DialFailure(t *testing.T) {
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: 19970, OriginalTarget: "127.0.0.1"}
	_, err := TeamSpeakHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "serveradmin", Password: "admin"})
	if err == nil {
		t.Error("expected connection error, got nil")
	}
}
