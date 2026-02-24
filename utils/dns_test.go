package utils

import (
	"net"
	"testing"
)

func TestLookupAddr(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
		// wantIP: if non-empty, check exact IP. If empty, just verify non-nil.
		wantIP string
	}{
		{
			name:   "valid IPv4",
			addr:   "1.2.3.4",
			wantIP: "1.2.3.4",
		},
		{
			name:   "valid IPv6 loopback",
			addr:   "::1",
			wantIP: "::1",
		},
		{
			name:   "valid IPv6 full",
			addr:   "2001:db8::1",
			wantIP: "2001:db8::1",
		},
		{
			name: "valid hostname localhost",
			addr: "localhost",
			// resolves via /etc/hosts — no real DNS call; can be 127.0.0.1 or ::1
			wantIP: "", // skip exact check, just verify non-nil
		},
		{
			name:    "invalid hostname",
			addr:    "this-host-does-not-exist.invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LookupAddr(tt.addr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LookupAddr(%q) expected error, got nil (ip=%v)", tt.addr, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("LookupAddr(%q) unexpected error: %v", tt.addr, err)
			}
			if got == nil {
				t.Fatalf("LookupAddr(%q) returned nil IP", tt.addr)
			}

			if tt.wantIP != "" {
				want := net.ParseIP(tt.wantIP)
				if !got.Equal(want) {
					t.Errorf("LookupAddr(%q) = %v, want %v", tt.addr, got, want)
				}
			}
		})
	}
}
