package utils

import (
	"crypto/tls"
	"testing"
)

func TestGetTLSConfig_ReturnsClone(t *testing.T) {
	cfg1 := GetTLSConfig()
	cfg2 := GetTLSConfig()

	if cfg1 == cfg2 {
		t.Error("GetTLSConfig should return distinct clones, got same pointer")
	}
}

func TestGetTLSConfig_InsecureSkipVerify(t *testing.T) {
	cfg := GetTLSConfig()
	if !cfg.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify = true")
	}
}

func TestGetTLSConfig_MinVersion(t *testing.T) {
	cfg := GetTLSConfig()
	if cfg.MinVersion != tls.VersionTLS10 {
		t.Errorf("expected MinVersion = TLS 1.0 (%d), got %d", tls.VersionTLS10, cfg.MinVersion)
	}
}

func TestGetTLSConfig_MutationSafety(t *testing.T) {
	cfg1 := GetTLSConfig()
	cfg1.ServerName = "example.com"

	cfg2 := GetTLSConfig()
	if cfg2.ServerName == "example.com" {
		t.Error("mutating a clone should not affect subsequent clones")
	}
}
