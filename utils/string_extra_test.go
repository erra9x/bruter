package utils

import (
	"testing"
)

func TestMD5Hex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:  "hello",
			input: "hello",
			want:  "5d41402abc4b2a76b9719d911017c592",
		},
		{
			name:  "password",
			input: "password",
			want:  "5f4dcc3b5aa765d61d8327deb882cf99",
		},
		{
			name:  "unicode",
			input: "日本語",
			want:  "00110af8b4393ef3f72c50be5b332bec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MD5Hex(tt.input)
			if got != tt.want {
				t.Errorf("MD5Hex(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMD5Hex_Deterministic(t *testing.T) {
	a := MD5Hex("test")
	b := MD5Hex("test")
	if a != b {
		t.Errorf("MD5Hex should be deterministic: got %q and %q", a, b)
	}
}

func TestMD5Hex_DifferentInputs(t *testing.T) {
	a := MD5Hex("foo")
	b := MD5Hex("bar")
	if a == b {
		t.Errorf("MD5Hex(foo) == MD5Hex(bar), hashes should differ")
	}
}
