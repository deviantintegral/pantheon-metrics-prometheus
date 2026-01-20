// Package version provides build-time version information for terminus-golang.
//
// Version information is injected at build time via ldflags by GoReleaser:
//
//	-X github.com/deviantintegral/terminus-golang/pkg/version.version={{ .Version }}
//
// GoReleaser automatically handles versioning:
//   - For tagged releases: version is the git tag (e.g., "0.6.0")
//   - For snapshot builds: version is the short commit hash (e.g., "abc1234")
//   - For dirty builds: version includes "-dirty" suffix (e.g., "abc1234-dirty")
//
// For local development without GoReleaser, use the Makefile which runs:
//
//	go build -ldflags "-X .../version.version=$(git describe --tags --always --dirty)"
package version

// version is set at build time via ldflags.
// GoReleaser sets this to the tag for releases or commit hash for snapshots.
var version string

// String returns the version string for use in user agents and display.
//
// Returns:
//   - The semantic version for tagged releases (e.g., "0.6.0")
//   - The short commit hash for snapshot builds (e.g., "abc1234")
//   - The commit hash with "-dirty" suffix for dirty builds (e.g., "abc1234-dirty")
//   - "dev" if no version information is available (local build without ldflags)
func String() string {
	if version == "" {
		return "dev"
	}
	return version
}
