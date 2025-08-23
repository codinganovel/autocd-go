package autocd

import (
	"os"
	"os/exec"
	"path/filepath"
)

// detectShell implements priority-based shell detection
func detectShell(shellOverride string) *ShellInfo {
	// 1. Explicit override (if provided)
	if shellOverride != "" {
		return validateShellOverride(shellOverride)
	}

	// 2. Environment variables (Unix)
	return detectUnixShell()
}

func validateShellOverride(shellOverride string) *ShellInfo {
	// Check if it's a shell name or full path
	var shellPath string
	if filepath.IsAbs(shellOverride) {
		shellPath = shellOverride
	} else {
		// Look for executable in PATH
		if path, err := exec.LookPath(shellOverride); err == nil {
			shellPath = path
		} else {
			return &ShellInfo{
				Path:    shellOverride,
				IsValid: false,
			}
		}
	}

	return &ShellInfo{
		Path:    shellPath,
		IsValid: fileExists(shellPath),
	}
}

func detectUnixShell() *ShellInfo {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh" // POSIX fallback
	}

	// If SHELL is set but invalid, fallback to /bin/sh
	if !fileExists(shell) {
		shell = "/bin/sh"
	}

	return &ShellInfo{
		Path:    shell,
		IsValid: fileExists(shell),
	}
}

// fileExists checks if a file exists and is executable
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	// On any stat error (including not-exist or permission issues), treat as not existing/usable
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	// Check if file is executable (any execute bit set)
	return info.Mode()&0111 != 0
}
