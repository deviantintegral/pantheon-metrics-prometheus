// Package version provides build-time version information for pantheon-metrics-prometheus.
//
// Version information is injected at build time via ldflags by GoReleaser:
//
//	-X github.com/deviantintegral/pantheon-metrics-prometheus/internal/version.version={{ .Version }}
//
// GoReleaser automatically handles versioning:
//   - For tagged releases: version is the git tag (e.g., "1.2.3")
//   - For snapshot builds: version is the short commit hash (e.g., "abc1234")
//   - For dirty builds: version includes "-dirty" suffix (e.g., "abc1234-dirty")
//
// For local development without GoReleaser, use:
//
//	go build -ldflags "-X .../version.version=$(git describe --tags --always --dirty)"
package version

import (
	"fmt"
	"runtime"
)

// version is set at build time via ldflags.
// GoReleaser sets this to the tag for releases or commit hash for snapshots.
var version string

// AppName is the name of the application used in the user agent.
const AppName = "pantheon-metrics-prometheus"

// String returns the version string for use in user agents and display.
//
// Returns:
//   - The semantic version for tagged releases (e.g., "1.2.3")
//   - The short commit hash for snapshot builds (e.g., "abc1234")
//   - The commit hash with "-dirty" suffix for dirty builds (e.g., "abc1234-dirty")
//   - "dev" if no version information is available (local build without ldflags)
func String() string {
	if version == "" {
		return "dev"
	}
	return version
}

// UserAgent returns the user agent string for API requests.
// Format: pantheon-metrics-prometheus/<version> (go_version=<go>; os=<os>; arch=<arch>)
func UserAgent() string {
	return fmt.Sprintf("%s/%s (go_version=%s; os=%s; arch=%s)",
		AppName, String(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
