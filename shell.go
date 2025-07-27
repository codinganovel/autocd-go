package autocd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// detectShell implements priority-based shell detection
func detectShell(platform PlatformType, shellOverride string) *ShellInfo {
	// 1. Explicit override (if provided)
	if shellOverride != "" {
		return validateShellOverride(shellOverride)
	}

	// 2. Environment variables
	switch platform {
	case PlatformWindows:
		return detectWindowsShell()
	default:
		return detectUnixShell()
	}
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

	shellType := classifyShellByPath(shellPath)
	scriptExt := getScriptExtensionForShell(shellType)

	return &ShellInfo{
		Path:      shellPath,
		Type:      shellType,
		ScriptExt: scriptExt,
		IsValid:   fileExists(shellPath),
	}
}

func detectWindowsShell() *ShellInfo {
	// Priority: PowerShell Core > PowerShell > Command Prompt

	// Check for pwsh (PowerShell Core)
	if path := findExecutable("pwsh.exe"); path != "" {
		return &ShellInfo{
			Path:      path,
			Type:      ShellPowerShellCore,
			ScriptExt: ".ps1",
			IsValid:   true,
		}
	}

	// Check for powershell.exe
	if path := findExecutable("powershell.exe"); path != "" {
		return &ShellInfo{
			Path:      path,
			Type:      ShellPowerShell,
			ScriptExt: ".ps1",
			IsValid:   true,
		}
	}

	// Check COMSPEC environment variable
	if comspec := os.Getenv("COMSPEC"); comspec != "" && fileExists(comspec) {
		return &ShellInfo{
			Path:      comspec,
			Type:      ShellCmd,
			ScriptExt: ".bat",
			IsValid:   true,
		}
	}

	// Fallback to cmd.exe
	return &ShellInfo{
		Path:      "cmd.exe",
		Type:      ShellCmd,
		ScriptExt: ".bat",
		IsValid:   true,
	}
}

func detectUnixShell() *ShellInfo {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh" // POSIX fallback
	}

	shellType := classifyUnixShell(shell)

	return &ShellInfo{
		Path:      shell,
		Type:      shellType,
		ScriptExt: ".sh",
		IsValid:   fileExists(shell),
	}
}

func classifyUnixShell(shellPath string) ShellType {
	basename := filepath.Base(shellPath)
	switch {
	case strings.Contains(basename, "bash"):
		return ShellBash
	case strings.Contains(basename, "zsh"):
		return ShellZsh
	case strings.Contains(basename, "fish"):
		return ShellFish
	case strings.Contains(basename, "dash"):
		return ShellDash
	default:
		return ShellSh // Generic sh-compatible
	}
}

func classifyShellByPath(shellPath string) ShellType {
	basename := filepath.Base(shellPath)
	lower := strings.ToLower(basename)

	switch {
	case strings.Contains(lower, "pwsh"):
		return ShellPowerShellCore
	case strings.Contains(lower, "powershell"):
		return ShellPowerShell
	case strings.Contains(lower, "cmd"):
		return ShellCmd
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

func getScriptExtensionForShell(shellType ShellType) string {
	switch shellType {
	case ShellCmd:
		return ".bat"
	case ShellPowerShell, ShellPowerShellCore:
		return ".ps1"
	default:
		return ".sh"
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
