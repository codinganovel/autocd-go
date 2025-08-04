package autocd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
				Path:      shellOverride,
				Type:      ShellUnknown,
				ScriptExt: ".sh",
				IsValid:   false,
			}
		}
	}

	shellType := classifyShell(shellPath)
	scriptExt := ".sh" // All Unix shells use .sh scripts

	return &ShellInfo{
		Path:      shellPath,
		Type:      shellType,
		ScriptExt: scriptExt,
		IsValid:   fileExists(shellPath),
	}
}

func detectUnixShell() *ShellInfo {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh" // POSIX fallback
	}

	shellType := classifyShell(shell)

	return &ShellInfo{
		Path:      shell,
		Type:      shellType,
		ScriptExt: ".sh",
		IsValid:   fileExists(shell),
	}
}

func classifyShell(shellPath string) ShellType {
	basename := filepath.Base(shellPath)
	lower := strings.ToLower(basename)

	switch {
	case strings.Contains(lower, "bash"):
		return ShellBash
	case strings.Contains(lower, "zsh"):
		return ShellZsh
	case strings.Contains(lower, "fish"):
		return ShellFish
	case strings.Contains(lower, "dash"):
		return ShellDash
	case strings.Contains(lower, "sh"):
		return ShellSh
	default:
		return ShellUnknown
	}
}

// findExecutable looks for executable in PATH
func findExecutable(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	return ""
}

// fileExists checks if a file exists and is accessible
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
