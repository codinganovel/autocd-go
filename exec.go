package autocd

import (
	"fmt"
	"os"
	"syscall"
)

// executeScript replaces current process with script using Unix syscall.Exec
func executeScript(scriptPath string, shell *ShellInfo, debugMode bool) error {
	if debugMode {
		fmt.Fprintf(os.Stderr, "autocd: executing script %s with shell %s\n", scriptPath, shell.Path)
	}

	// Use the detected shell to execute the script
	executable := shell.Path
	args := []string{shell.Path, scriptPath}

	// Replace current process with Unix syscall.Exec
	return syscall.Exec(executable, args, os.Environ())
}

// ExecReplacement handles the actual process replacement
// This is the core function that never returns on success
func ExecReplacement(scriptPath string, shell *ShellInfo, debugMode bool) error {
	// Validate inputs
	if scriptPath == "" {
		return newPathError(ErrorPathNotFound, "", fmt.Errorf("script path is empty"))
	}
	if shell == nil {
		return fmt.Errorf("shell info is nil")
	}
	if !shell.IsValid {
		return fmt.Errorf("shell is not valid: %s", shell.Path)
	}

	// Check that script file exists
	if !fileExists(scriptPath) {
		return newPathError(ErrorPathNotFound, scriptPath, fmt.Errorf("script file does not exist"))
	}

	// Execute the script - this should never return
	return executeScript(scriptPath, shell, debugMode)
}
