package autocd

import (
	"fmt"
	"os"
	"syscall"
)

// executeScript replaces current process with script using Unix syscall.Exec
func executeScript(scriptPath string, shell *ShellInfo, debugMode bool) error {
	if debugMode {
		fmt.Fprintf(os.Stderr, "autocd: executing script %s (target shell: %s)\n", scriptPath, shell.Path)
	}

	// Always use /bin/sh to execute our POSIX script, regardless of user's shell
	// This fixes fish compatibility and other exotic shells
	// The script will exec into the user's shell at the end
	executable := "/bin/sh"
	args := []string{executable, scriptPath}

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
		return newShellDetectionError("shell info is nil")
	}
	if !shell.IsValid {
		return newShellDetectionError(fmt.Sprintf("shell is not valid: %s", shell.Path))
	}

	// Check that script file exists and is executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return newPathError(ErrorPathNotFound, scriptPath, fmt.Errorf("script file does not exist"))
		}
		return newPathError(ErrorPathNotAccessible, scriptPath, fmt.Errorf("unable to access script: %w", err))
	}
	if info.IsDir() {
		return newPathError(ErrorPathNotDirectory, scriptPath, fmt.Errorf("script path is a directory"))
	}
	if info.Mode()&0111 == 0 {
		return newPathError(ErrorPathNotAccessible, scriptPath, fmt.Errorf("script file is not executable"))
	}

	// Execute the script - this should never return
	return executeScript(scriptPath, shell, debugMode)
}
