// Package shversion contains version information being set via linker flags when building via the
// Makefile
package shversion

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// This gets set via ldflags when building via the Makefile.
var version string

// Version returns shuttermint's version string.
func Version() string {
	var raceinfo string
	if raceDetectorEnabled {
		raceinfo = ", race detector enabled"
	}
	return fmt.Sprintf("%s (%s, %s-%s%s)", VersionShort(), runtime.Version(), runtime.GOOS, runtime.GOARCH, raceinfo)
}

func VersionShort() string {
	if version == "" {
		info, ok := debug.ReadBuildInfo()
		if ok {
			versionShort := info.Main.Version
			if versionShort == "(devel)" {
				for _, s := range info.Settings {
					if s.Key == "vcs.revision" {
						return fmt.Sprintf("(devel-%s)", s.Value)
					}
				}
			}
			return versionShort
		}
	}
	return version
}
