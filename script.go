package autocd

import (
	"fmt"
	"strings"
)

// generateScript creates appropriate script for detected platform/shell
func generateScript(targetDir string, shell *ShellInfo) (string, error) {
	// Sanitize path for script injection prevention
	safePath := sanitizePathForShell(targetDir, shell.Type)

	switch shell.Type {
	case ShellCmd:
		return generateBatchScript(safePath, shell.Path), nil
	case ShellPowerShell, ShellPowerShellCore:
		return generatePowerShellScript(safePath, shell.Path), nil
	default:
		return generateUnixScript(safePath, shell.Path, shell.Type), nil
	}
}

func generateBatchScript(targetDir, shellPath string) string {
	return fmt.Sprintf(`@echo off
REM autocd transition script - auto-cleanup on exit
cd /d "%s" 2>nul || (
    echo Warning: Could not change to %s >&2
    echo Continuing in current directory >&2
)
"%s"
`, targetDir, targetDir, shellPath)
}

func generatePowerShellScript(targetDir, shellPath string) string {
	return fmt.Sprintf(`# autocd transition script - auto-cleanup on exit
try {
    Set-Location -Path "%s" -ErrorAction Stop
    Write-Host "Directory changed to: %s"
} catch {
    Write-Warning "Could not change to %s : $_"
    Write-Host "Continuing in current directory"
}

& "%s"
`, targetDir, targetDir, targetDir, shellPath)
}

func generateUnixScript(targetDir, shellPath string, shellType ShellType) string {
	shebang := getShebang(shellType)

	return fmt.Sprintf(`%s
# autocd transition script - auto-cleanup on exit
trap 'rm -f "$0" 2>/dev/null || true' EXIT INT TERM

TARGET_DIR="%s"
SHELL_PATH="%s"

# Attempt to change directory with error handling
if cd "$TARGET_DIR" 2>/dev/null; then
    echo "Directory changed to: $TARGET_DIR"
else
    echo "Warning: Could not change to $TARGET_DIR" >&2
    echo "Continuing in current directory" >&2
fi

# Replace current process with shell
exec "$SHELL_PATH"
`, shebang, targetDir, shellPath)
}

func getShebang(shellType ShellType) string {
	switch shellType {
	case ShellBash:
		return "#!/bin/bash"
	case ShellZsh:
		return "#!/bin/zsh"
	case ShellFish:
		return "#!/usr/bin/fish"
	case ShellDash:
		return "#!/bin/dash"
	default:
		return "#!/bin/sh"
	}
}

// sanitizePathForShell prevents shell injection
func sanitizePathForShell(path string, shellType ShellType) string {
	switch shellType {
	case ShellCmd:
		// Escape dangerous characters for batch files
		path = strings.ReplaceAll(path, `"`, `""`) // Escape quotes
		path = strings.ReplaceAll(path, `;`, `^;`) // Escape command separator
		path = strings.ReplaceAll(path, `&`, `^&`) // Escape command separator
		path = strings.ReplaceAll(path, `|`, `^|`) // Escape pipe
		path = strings.ReplaceAll(path, `<`, `^<`) // Escape input redirection
		path = strings.ReplaceAll(path, `>`, `^>`) // Escape output redirection
		path = strings.ReplaceAll(path, `%`, `%%`) // Escape variable expansion
		return path
	case ShellPowerShell, ShellPowerShellCore:
		// Escape dangerous characters for PowerShell
		path = strings.ReplaceAll(path, `"`, `""`) // Escape quotes
		path = strings.ReplaceAll(path, `;`, "`;") // Escape command separator
		path = strings.ReplaceAll(path, `&`, "`&") // Escape command separator
		path = strings.ReplaceAll(path, `|`, "`|") // Escape pipe
		path = strings.ReplaceAll(path, `<`, "`<") // Escape input redirection
		path = strings.ReplaceAll(path, `>`, "`>") // Escape output redirection
		path = strings.ReplaceAll(path, `$`, "`$") // Escape variable expansion
		return path
	default:
		// Escape for Unix shells
		path = strings.ReplaceAll(path, `\`, `\\`)
		path = strings.ReplaceAll(path, `"`, `\"`)
		return path
	}
}
