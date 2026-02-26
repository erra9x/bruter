package modules

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// startMockHTTPBasic spins up a test server that responds with the given status code.
func startMockHTTPBasic(t *testing.T, statusCode int) (*httptest.Server, *string, *string) {
	t.Helper()
	capturedHost := new(string)
	capturedAuth := new(string)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*capturedHost = r.Host
		*capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(statusCode)
	}))
	t.Cleanup(srv.Close)
	return srv, capturedHost, capturedAuth
}

func parseHTTPTestServer(t *testing.T, srv *httptest.Server) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Atoi port: %v", err)
	}
	return host, port
}

func TestHTTPBasicHandler_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantOk     bool
		wantErr    bool
	}{
		{"200 OK → success", http.StatusOK, true, false},
		{"302 Found → success", http.StatusFound, true, false},
		{"401 Unauthorized → fail", http.StatusUnauthorized, false, false},
		{"403 Forbidden → fail", http.StatusForbidden, false, false},
		{"500 Server Error → error", http.StatusInternalServerError, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _, _ := startMockHTTPBasic(t, tt.statusCode)
			host, port := parseHTTPTestServer(t, srv)

			target := &Target{
				IP:             net.ParseIP(host),
				Port:           port,
				OriginalTarget: host,
				Encryption:     false,
			}
			ok, err := HTTPBasicHandler(context.Background(), newTestDialer(t), 5*time.Second,
				target, &Credential{Username: "admin", Password: "admin"})

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestHTTPBasicHandler_DomainTarget_SetsHostHeader(t *testing.T) {
	srv, capturedHost, _ := startMockHTTPBasic(t, http.StatusOK)
	host, port := parseHTTPTestServer(t, srv)

	target := &Target{
		IP:             net.ParseIP(host),
		Port:           port,
		OriginalTarget: "example.com",
		Encryption:     false,
	}
	_, _ = HTTPBasicHandler(context.Background(), newTestDialer(t), 5*time.Second,
		target, &Credential{Username: "admin", Password: "admin"})

	if *capturedHost != "example.com" {
		t.Errorf("Host = %q, want %q", *capturedHost, "example.com")
	}
}

func TestHTTPBasicHandler_DomainWithPort_SetsHostWithoutPort(t *testing.T) {
	srv, capturedHost, _ := startMockHTTPBasic(t, http.StatusOK)
	host, port := parseHTTPTestServer(t, srv)

	target := &Target{
		IP:             net.ParseIP(host),
		Port:           port,
		OriginalTarget: "example.com:8080",
		Encryption:     false,
	}
	_, _ = HTTPBasicHandler(context.Background(), newTestDialer(t), 5*time.Second,
		target, &Credential{Username: "admin", Password: "admin"})

	if *capturedHost != "example.com" {
		t.Errorf("Host = %q, want %q (port should be stripped)", *capturedHost, "example.com")
	}
}

func TestHTTPBasicHandler_IPTarget_DoesNotOverrideHost(t *testing.T) {
	srv, capturedHost, _ := startMockHTTPBasic(t, http.StatusOK)
	host, port := parseHTTPTestServer(t, srv)

	target := &Target{
		IP:             net.ParseIP(host),
		Port:           port,
		OriginalTarget: host, // bare IP — no override
		Encryption:     false,
	}
	_, _ = HTTPBasicHandler(context.Background(), newTestDialer(t), 5*time.Second,
		target, &Credential{Username: "admin", Password: "admin"})

	if *capturedHost == "example.com" {
		t.Error("Host should not be overridden to a domain when OriginalTarget is an IP")
	}
}

func TestHTTPBasicHandler_DialFailure(t *testing.T) {
	target := &Target{IP: net.ParseIP("127.0.0.1"), Port: 19997, OriginalTarget: "127.0.0.1"}
	_, err := HTTPBasicHandler(context.Background(), newTestDialer(t), 500*time.Millisecond,
		target, &Credential{Username: "admin", Password: "admin"})
	if err == nil {
		t.Error("expected connection error, got nil")
	}
}
