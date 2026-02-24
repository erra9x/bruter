package utils

import "testing"

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		substrs []string
		want    bool
	}{
		{
			name:    "match first substring",
			s:       "hello world",
			substrs: []string{"hello", "foo", "bar"},
			want:    true,
		},
		{
			name:    "match last substring",
			s:       "hello world",
			substrs: []string{"foo", "bar", "world"},
			want:    true,
		},
		{
			name:    "match middle substring",
			s:       "hello world foo",
			substrs: []string{"bar", "world", "baz"},
			want:    true,
		},
		{
			name:    "no match",
			s:       "hello world",
			substrs: []string{"foo", "bar", "baz"},
			want:    false,
		},
		{
			name:    "empty substrs slice",
			s:       "hello world",
			substrs: []string{},
			want:    false,
		},
		{
			name:    "empty string matches empty substr",
			s:       "",
			substrs: []string{""},
			want:    true,
		},
		{
			name:    "empty string no match",
			s:       "",
			substrs: []string{"foo"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsAny(tt.s, tt.substrs...)
			if got != tt.want {
				t.Errorf("ContainsAny(%q, %v) = %v, want %v", tt.s, tt.substrs, got, tt.want)
			}
		})
	}
}
