package update

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantPrerel string
		wantErr    bool
	}{
		{
			name:      "simple version",
			input:     "1.2.3",
			wantMajor: 1, wantMinor: 2, wantPatch: 3,
		},
		{
			name:      "version with v prefix",
			input:     "v1.2.3",
			wantMajor: 1, wantMinor: 2, wantPatch: 3,
		},
		{
			name:      "version with prerelease",
			input:     "v1.2.3-beta.1",
			wantMajor: 1, wantMinor: 2, wantPatch: 3,
			wantPrerel: "beta.1",
		},
		{
			name:      "zero version",
			input:     "0.0.0",
			wantMajor: 0, wantMinor: 0, wantPatch: 0,
		},
		{
			name:      "large numbers",
			input:     "v100.200.300",
			wantMajor: 100, wantMinor: 200, wantPatch: 300,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - missing patch",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "invalid format - letters",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid format - extra parts",
			input:   "1.2.3.4",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVersion(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseVersion(%q) unexpected error: %v", tt.input, err)
				return
			}
			if v.Major != tt.wantMajor {
				t.Errorf("Major = %d, want %d", v.Major, tt.wantMajor)
			}
			if v.Minor != tt.wantMinor {
				t.Errorf("Minor = %d, want %d", v.Minor, tt.wantMinor)
			}
			if v.Patch != tt.wantPatch {
				t.Errorf("Patch = %d, want %d", v.Patch, tt.wantPatch)
			}
			if v.Prerelease != tt.wantPrerel {
				t.Errorf("Prerelease = %q, want %q", v.Prerelease, tt.wantPrerel)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name string
		v    Version
		want string
	}{
		{
			name: "simple",
			v:    Version{Major: 1, Minor: 2, Patch: 3},
			want: "v1.2.3",
		},
		{
			name: "with prerelease",
			v:    Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha"},
			want: "v1.0.0-alpha",
		},
		{
			name: "zeros",
			v:    Version{Major: 0, Minor: 0, Patch: 0},
			want: "v0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"major greater", "2.0.0", "1.0.0", 1},
		{"major less", "1.0.0", "2.0.0", -1},
		{"minor greater", "1.2.0", "1.1.0", 1},
		{"minor less", "1.1.0", "1.2.0", -1},
		{"patch greater", "1.0.2", "1.0.1", 1},
		{"patch less", "1.0.1", "1.0.2", -1},
		{"release > prerelease", "1.0.0", "1.0.0-beta", 1},
		{"prerelease < release", "1.0.0-beta", "1.0.0", -1},
		{"prerelease equal", "1.0.0-beta", "1.0.0-beta", 0},
		{"prerelease alpha < beta", "1.0.0-alpha", "1.0.0-beta", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := ParseVersion(tt.a)
			b, _ := ParseVersion(tt.b)
			if got := a.Compare(b); got != tt.want {
				t.Errorf("Compare(%s, %s) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestVersionComparisons(t *testing.T) {
	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("2.0.0")

	if !v1.LessThan(v2) {
		t.Error("v1.0.0 should be less than v2.0.0")
	}
	if !v2.GreaterThan(v1) {
		t.Error("v2.0.0 should be greater than v1.0.0")
	}
	if !v1.Equal(v1) {
		t.Error("v1.0.0 should equal v1.0.0")
	}
	if v1.Equal(v2) {
		t.Error("v1.0.0 should not equal v2.0.0")
	}
}
