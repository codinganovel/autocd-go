package autocd

import (
	"fmt"
	"strings"
)

// generateScript creates Unix shell script for directory transition
func generateScript(targetDir string, shell *ShellInfo) (string, error) {
	// Sanitize path for script injection prevention
	safePath := sanitizePathForShell(targetDir)

	// Generate Unix shell script
	return generateUnixScript(safePath, shell.Path), nil
}

func generateUnixScript(targetDir, shellPath string) string {
	// Always use /bin/sh shebang since we execute with /bin/sh
	shebang := "#!/bin/sh"

	return fmt.Sprintf(`%s
# autocd transition script - auto-cleanup on exit
trap 'rm -f "$0" 2>/dev/null || true' EXIT INT TERM

TARGET_DIR='%s'
SHELL_PATH='%s'

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

// sanitizePathForShell prevents shell injection in Unix shells using single quotes
func sanitizePathForShell(path string) string {
	// Use single quotes for robust escaping
	// Only need to escape embedded single quotes as '"'"'
	return strings.ReplaceAll(path, `'`, `'"'"'`)
}
