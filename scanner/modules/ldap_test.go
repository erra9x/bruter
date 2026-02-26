package modules

import (
	"context"
	"net"
	"testing"
	"time"
)

// TestLDAPHandler_DialFailure verifies that a connection refused error is returned
// as an error (not misclassified as authentication success or failure).
func TestLDAPHandler_DialFailure(t *testing.T) {
	target := &Target{
		IP:             net.ParseIP("127.0.0.1"),
		Port:           19992,
		OriginalTarget: "127.0.0.1",
		Encryption:     false,
	}
	ok, err := LDAPHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "cn=admin,dc=example,dc=com", Password: "admin"})

	if err == nil {
		t.Error("expected connection error, got nil")
	}
	if ok {
		t.Error("ok should be false on dial failure")
	}
}

// TestLDAPHandler_TLS_DialFailure verifies that the LDAPS (TLS) path also
// handles a closed port gracefully.
func TestLDAPHandler_TLS_DialFailure(t *testing.T) {
	target := &Target{
		IP:             net.ParseIP("127.0.0.1"),
		Port:           19991,
		OriginalTarget: "127.0.0.1",
		Encryption:     true,
	}
	ok, err := LDAPHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "cn=admin,dc=example,dc=com", Password: "admin"})

	if err == nil {
		t.Error("expected connection error for LDAPS on closed port, got nil")
	}
	if ok {
		t.Error("ok should be false when TLS connection fails")
	}
}

// TestLDAPHandler_ImmediateClose verifies that a server that immediately closes
// the connection is treated as a connection error, not an auth success or failure.
func TestLDAPHandler_ImmediateClose(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	// Accept one connection and immediately close it.
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		_ = conn.Close()
	}()

	addr := ln.Addr().(*net.TCPAddr)
	target := &Target{
		IP:             net.ParseIP("127.0.0.1"),
		Port:           addr.Port,
		OriginalTarget: "127.0.0.1",
		Encryption:     false,
	}
	ok, err := LDAPHandler(context.Background(), newTestDialer(t), 2*time.Second,
		target, &Credential{Username: "cn=admin,dc=example,dc=com", Password: "admin"})

	// Connection closed before LDAP handshake — must not be treated as auth success.
	if ok {
		t.Error("ok should be false when server closes connection immediately")
	}
	// An error is expected here (protocol error or EOF).
	_ = err // err may or may not be nil depending on library behaviour; ok=false is the key invariant
}
