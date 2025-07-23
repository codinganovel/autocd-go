package autocd

import (
	"runtime"
)

// detectPlatform determines the current operating system platform
func detectPlatform() PlatformType {
	switch runtime.GOOS {
	case "windows":
		return PlatformWindows
	case "darwin":
		return PlatformMacOS
	case "linux":
		return PlatformLinux
	case "freebsd", "openbsd", "netbsd":
		return PlatformBSD
	default:
		return PlatformUnix // Generic Unix fallback
	}
}
