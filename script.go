package autocd

import (
	"fmt"
	"strings"
)

// generateScript creates Unix shell script for directory transition
func generateScript(targetDir string, shell *ShellInfo) (string, error) {
	// Sanitize path for script injection prevention
	safePath := sanitizePathForShell(targetDir, shell.Type)

	// Generate Unix shell script
	return generateUnixScript(safePath, shell.Path, shell.Type), nil
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

// sanitizePathForShell prevents shell injection in Unix shells
func sanitizePathForShell(path string, shellType ShellType) string {
	// Use strings.Replacer for better performance with multiple replacements
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return replacer.Replace(path)
}
