package modules

import (
	"context"
	"net"
	"testing"
	"time"
)

// TestMySQLHandler_DialFailure verifies that a connection refused error is returned
// as an error (not misclassified as authentication success or failure).
func TestMySQLHandler_DialFailure(t *testing.T) {
	target := &Target{
		IP:             net.ParseIP("127.0.0.1"),
		Port:           19996,
		OriginalTarget: "127.0.0.1",
		Encryption:     false,
	}
	ok, err := MySQLHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "root", Password: "root"})

	if err == nil {
		t.Error("expected connection error, got nil")
	}
	if ok {
		t.Error("ok should be false on dial failure")
	}
}

// TestMySQLHandler_EncryptionFlag verifies that the handler doesn't panic or
// crash when Encryption is set to true on a non-TLS endpoint.
func TestMySQLHandler_EncryptionFlag(t *testing.T) {
	target := &Target{
		IP:             net.ParseIP("127.0.0.1"),
		Port:           19995,
		OriginalTarget: "127.0.0.1",
		Encryption:     true,
	}
	ok, err := MySQLHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "root", Password: "root"})

	// Should fail to connect (port closed), not panic.
	if err == nil {
		t.Error("expected connection error with Encryption=true on closed port, got nil")
	}
	if ok {
		t.Error("ok should be false when connection fails")
	}
}
