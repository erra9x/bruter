package modules

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"testing"
	"time"
)

// generateTestPEM returns a PEM-encoded EC private key for use in tests.
func generateTestPEM(t *testing.T) []byte {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
}

func TestSSHKeyHandler_InvalidPEM(t *testing.T) {
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: 19985, OriginalTarget: "127.0.0.1"}
	ok, err := SSHKeyHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "root", Password: "-----BEGIN INVALID-----\nbadsdata\n-----END INVALID-----\n"})

	if err == nil {
		t.Error("expected error for invalid PEM, got nil")
	}
	if ok {
		t.Error("ok should be false for invalid PEM")
	}
}

func TestSSHKeyHandler_MissingKeyFile(t *testing.T) {
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: 19984, OriginalTarget: "127.0.0.1"}
	ok, err := SSHKeyHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "root", Password: "/nonexistent/path/to/key.pem"})

	if err == nil {
		t.Error("expected error for missing key file, got nil")
	}
	if ok {
		t.Error("ok should be false for missing key file")
	}
}

func TestSSHKeyHandler_ValidKeyDialFailure(t *testing.T) {
	pemBytes := generateTestPEM(t)
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: 19983, OriginalTarget: "127.0.0.1"}
	ok, err := SSHKeyHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "root", Password: string(pemBytes)})

	if err == nil {
		t.Error("expected connection error on closed port, got nil")
	}
	if ok {
		t.Error("ok should be false on dial failure")
	}
}

func TestSSHKeyHandler_KeyFromFile(t *testing.T) {
	pemBytes := generateTestPEM(t)

	f, err := os.CreateTemp(t.TempDir(), "testkey*.pem")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err = f.Write(pemBytes); err != nil {
		t.Fatalf("write key file: %v", err)
	}
	_ = f.Close()

	// Using file path (no "-----" prefix) — should load from file, then fail on connect.
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: 19982, OriginalTarget: "127.0.0.1"}
	ok, err := SSHKeyHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "root", Password: f.Name()})

	if err == nil {
		t.Error("expected connection error on closed port, got nil")
	}
	if ok {
		t.Error("ok should be false on dial failure")
	}
}
