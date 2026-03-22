package utils

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

// --- CustomTransport ---

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestCustomTransport_SetsHeaders(t *testing.T) {
	var capturedReq *http.Request
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedReq = req
		return &http.Response{StatusCode: 200}, nil
	})

	ct := &CustomTransport{
		Transport: inner,
		UserAgent: "TestAgent/1.0",
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := ct.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.Header.Get("User-Agent") != "TestAgent/1.0" {
		t.Errorf("User-Agent = %q, want %q", capturedReq.Header.Get("User-Agent"), "TestAgent/1.0")
	}
	if capturedReq.Header.Get("Accept-Language") != "en-US,en;q=0.9" {
		t.Errorf("Accept-Language = %q, want %q", capturedReq.Header.Get("Accept-Language"), "en-US,en;q=0.9")
	}
}

func TestCustomTransport_PropagatesError(t *testing.T) {
	testErr := errors.New("transport error")
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, testErr
	})

	ct := &CustomTransport{
		Transport: inner,
		UserAgent: "TestAgent",
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := ct.RoundTrip(req)
	if !errors.Is(err, testErr) {
		t.Errorf("expected %v, got %v", testErr, err)
	}
}

// --- NewProxyAwareDialer ---

func TestNewProxyAwareDialer_NoProxy(t *testing.T) {
	d, err := NewProxyAwareDialer("", "", 5*time.Second, "TestUA", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil dialer")
	}
	if d.HTTPClient == nil {
		t.Error("expected non-nil HTTPClient")
	}
	if d.Timeout() != 5*time.Second {
		t.Errorf("Timeout() = %v, want 5s", d.Timeout())
	}
}

func TestNewProxyAwareDialer_InvalidProxyAuth(t *testing.T) {
	_, err := NewProxyAwareDialer("127.0.0.1:1080", "invalidformat", 5*time.Second, "TestUA", nil)
	if err == nil {
		t.Fatal("expected error for invalid proxy auth format")
	}
}

func TestNewProxyAwareDialer_ValidProxyAuth(t *testing.T) {
	// This won't actually connect, but should parse OK
	d, err := NewProxyAwareDialer("127.0.0.1:1080", "user:pass", 5*time.Second, "TestUA", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil dialer")
	}
}

func TestNewProxyAwareDialer_WithLocalAddr(t *testing.T) {
	d, err := NewProxyAwareDialer("", "", 3*time.Second, "TestUA", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.HTTPClient == nil {
		t.Error("expected non-nil HTTPClient")
	}
}

// --- TLSDialerWrapper ---

func TestTLSDialerWrapper_DialFails(t *testing.T) {
	d, err := NewProxyAwareDialer("", "", 1*time.Second, "TestUA", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := &TLSDialerWrapper{Dialer: d}
	// Dial to a non-existent host should fail
	_, err = w.Dial("tcp", "192.0.2.1:12345")
	if err == nil {
		t.Error("expected error dialing non-existent host")
	}
}

func TestProxyAwareDialer_DialAutoPlaintext(t *testing.T) {
	d, err := NewProxyAwareDialer("", "", 1*time.Second, "TestUA", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DialAuto with encryption=false should use plaintext Dial
	_, err = d.DialAuto("tcp", "192.0.2.1:12345", false)
	if err == nil {
		t.Error("expected error dialing non-existent host")
	}
}

func TestProxyAwareDialer_DialAutoEncrypted(t *testing.T) {
	d, err := NewProxyAwareDialer("", "", 1*time.Second, "TestUA", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DialAuto with encryption=true should use TLS
	_, err = d.DialAuto("tcp", "192.0.2.1:12345", true)
	if err == nil {
		t.Error("expected error dialing non-existent host")
	}
}

func TestProxyAwareDialer_DialTimeout(t *testing.T) {
	d, err := NewProxyAwareDialer("", "", 1*time.Second, "TestUA", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DialTimeout with zero timeout should use default
	_, err = d.DialTimeout("tcp", "192.0.2.1:12345", 0)
	if err == nil {
		t.Error("expected error dialing non-existent host")
	}
}
