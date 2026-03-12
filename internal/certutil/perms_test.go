package certutil

import "testing"

func TestAllowsPath(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		destPath string
		want     bool
	}{
		{
			name:     "empty paths allows everything",
			paths:    nil,
			destPath: "libera/#general",
			want:     true,
		},
		{
			name:     "exact match",
			paths:    []string{"libera/#general"},
			destPath: "libera/#general",
			want:     true,
		},
		{
			name:     "exact match rejected",
			paths:    []string{"libera/#general"},
			destPath: "libera/#random",
			want:     false,
		},
		{
			name:     "server prefix matches any channel",
			paths:    []string{"libera"},
			destPath: "libera/#general",
			want:     true,
		},
		{
			name:     "server prefix does not match different server",
			paths:    []string{"libera"},
			destPath: "freenode/#general",
			want:     false,
		},
		{
			name:     "server prefix does not match itself without slash",
			paths:    []string{"libera"},
			destPath: "libera",
			want:     true,
		},
		{
			name:     "channel-only entry matches any server",
			paths:    []string{"#general"},
			destPath: "#general",
			want:     true,
		},
		{
			name:     "channel-only entry does not match different channel",
			paths:    []string{"#general"},
			destPath: "#random",
			want:     false,
		},
		{
			name:     "multiple paths, one matches",
			paths:    []string{"libera/#general", "freenode/#dev"},
			destPath: "freenode/#dev",
			want:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Permissions{Paths: tc.paths}
			if got := p.AllowsPath(tc.destPath); got != tc.want {
				t.Errorf("AllowsPath(%q) = %v, want %v", tc.destPath, got, tc.want)
			}
		})
	}
}
