package autocd

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

// executeScript replaces current process with script
func executeScript(scriptPath string, shell *ShellInfo, debugMode bool) error {
	if debugMode {
		fmt.Fprintf(os.Stderr, "autocd: executing script %s with shell %s\n", scriptPath, shell.Path)
	}

	switch runtime.GOOS {
	case "windows":
		return executeWindowsScript(scriptPath, shell)
	default:
		return executeUnixScript(scriptPath, shell)
	}
}

func executeWindowsScript(scriptPath string, shell *ShellInfo) error {
	// Use the detected shell path, not hardcoded executable names
	executable := shell.Path
	var args []string

	switch shell.Type {
	case ShellPowerShell:
		args = []string{shell.Path, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
	case ShellPowerShellCore:
		args = []string{shell.Path, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
	default: // ShellCmd
		args = []string{shell.Path, "/c", scriptPath}
	}

	// Windows doesn't support syscall.Exec, use StartProcess instead
	attr := &syscall.ProcAttr{
		Dir:   "",
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	_, _, err := syscall.StartProcess(executable, args, attr)
	if err != nil {
		return err
	}

	// Exit current process after starting the new shell
	os.Exit(0)
	return nil
}

func executeUnixScript(scriptPath string, shell *ShellInfo) error {
	// Use the detected shell to execute the script
	executable := shell.Path
	args := []string{shell.Path, scriptPath}

	// Replace current process
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
