package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	// Test default value (empty string returns "dev")
	originalVersion := version
	defer func() { version = originalVersion }()

	version = ""
	if got := String(); got != "dev" {
		t.Errorf("String() with empty version = %q, want %q", got, "dev")
	}

	// Test with a set version
	version = "1.2.3"
	if got := String(); got != "1.2.3" {
		t.Errorf("String() with version set = %q, want %q", got, "1.2.3")
	}

	// Test with commit hash
	version = "abc1234"
	if got := String(); got != "abc1234" {
		t.Errorf("String() with commit hash = %q, want %q", got, "abc1234")
	}

	// Test with dirty suffix
	version = "abc1234-dirty"
	if got := String(); got != "abc1234-dirty" {
		t.Errorf("String() with dirty suffix = %q, want %q", got, "abc1234-dirty")
	}
}

func TestUserAgent(t *testing.T) {
	ua := UserAgent()

	// Should contain app name
	if !strings.Contains(ua, AppName) {
		t.Errorf("UserAgent should contain app name %q, got %q", AppName, ua)
	}

	// Should contain Go version
	if !strings.Contains(ua, runtime.Version()) {
		t.Errorf("UserAgent should contain Go version %q, got %q", runtime.Version(), ua)
	}

	// Should contain OS
	if !strings.Contains(ua, runtime.GOOS) {
		t.Errorf("UserAgent should contain OS %q, got %q", runtime.GOOS, ua)
	}

	// Should contain architecture
	if !strings.Contains(ua, runtime.GOARCH) {
		t.Errorf("UserAgent should contain architecture %q, got %q", runtime.GOARCH, ua)
	}
}

func TestUserAgentWithVersion(t *testing.T) {
	originalVersion := version
	defer func() { version = originalVersion }()

	tests := []struct {
		name            string
		versionValue    string
		expectedVersion string
	}{
		{
			name:            "tagged release",
			versionValue:    "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			name:            "snapshot build",
			versionValue:    "abc1234",
			expectedVersion: "abc1234",
		},
		{
			name:            "dirty build",
			versionValue:    "abc1234-dirty",
			expectedVersion: "abc1234-dirty",
		},
		{
			name:            "dev build",
			versionValue:    "",
			expectedVersion: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version = tt.versionValue
			ua := UserAgent()

			expectedPrefix := AppName + "/" + tt.expectedVersion
			if !strings.HasPrefix(ua, expectedPrefix) {
				t.Errorf("UserAgent() should start with %q, got %q", expectedPrefix, ua)
			}
		})
	}
}
