package autocd

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// checkShellDepth examines the current shell nesting level and displays
// helpful warnings when appropriate based on platform capabilities
func checkShellDepth(opts *Options) {
	// Skip if warnings are disabled
	if opts.DisableDepthWarnings {
		return
	}

	// Unix systems: Check SHLVL environment variable
	shlvlStr := os.Getenv("SHLVL")
	if shlvlStr == "" {
		// No SHLVL means likely not in a shell or shell doesn't support it
		return
	}

	shlvl, err := strconv.Atoi(shlvlStr)
	if err != nil {
		// Invalid SHLVL value, skip warning
		return
	}

	// Show warning if above threshold
	if shlvl >= opts.DepthWarningThreshold {
		fmt.Fprintf(os.Stderr, "💡 Tip: You have %d nested shells from navigation.\n", shlvl)
		fmt.Fprintf(os.Stderr, "For better performance, consider opening a fresh terminal.\n")
	}
}

// ExitWithDirectory inherits the target directory to the parent shell when the process exits.
// This is the main library function providing the core autocd functionality.
//
// On success, this function never returns (the current process is replaced with
// a shell transition script). On failure, it returns an error allowing for graceful
// fallback handling.
//
// Example:
//
//	if err := autocd.ExitWithDirectory("/some/target/directory"); err != nil {
//		// Handle failure - process continues normally
//		log.Printf("autocd failed: %v", err)
//		os.Exit(1)
//	}
//	// This line is never reached on success
func ExitWithDirectory(targetPath string) error {
	return ExitWithDirectoryAdvanced(targetPath, nil)
}

// ExitWithDirectoryAdvanced provides advanced configuration options for directory inheritance.
// This function offers the same core functionality as ExitWithDirectory but with additional
// control over security levels, shell detection, debug mode, and other options.
//
// On success, this function never returns (the current process is replaced).
// On failure, it returns a structured AutoCDError for detailed error handling.
//
// Example:
//
//	opts := &autocd.Options{
//		SecurityLevel:         autocd.SecurityStrict,
//		DebugMode:             true,
//		DepthWarningThreshold: 10,
//	}
//	if err := autocd.ExitWithDirectoryAdvanced("/target/dir", opts); err != nil {
//		if autocd.IsPathError(err) {
//			log.Printf("Path validation failed: %v", err)
//		}
//		os.Exit(1)
//	}
func ExitWithDirectoryAdvanced(targetPath string, opts *Options) error {
	// Set defaults if options not provided
	if opts == nil {
		opts = &Options{
			SecurityLevel:         SecurityNormal,
			DebugMode:             os.Getenv("AUTOCD_DEBUG") != "",
			DepthWarningThreshold: 15,
			DisableDepthWarnings:  false,
		}
	}

	// Set defaults for new fields if not specified
	if opts.DepthWarningThreshold == 0 {
		opts.DepthWarningThreshold = 15
	}

	// Check shell depth and show helpful warnings if appropriate
	checkShellDepth(opts)

	// 1. Clean up old temporary scripts from previous runs
	if err := cleanupOldScripts(1 * time.Hour); err != nil {
		// Non-fatal error - log if debug mode but continue
		if opts.DebugMode {
			fmt.Fprintf(os.Stderr, "autocd: cleanup warning: %v\n", err)
		}
	}

	// 2. Validate target directory
	validatedPath, err := validateTargetPath(targetPath, opts.SecurityLevel)
	if err != nil {
		return newPathValidationError(targetPath, err)
	}

	// 3. Detect shell
	shell := detectShell(opts.Shell)

	if !shell.IsValid {
		return newShellDetectionError("no valid shell found")
	}

	if opts.DebugMode {
		fmt.Fprintf(os.Stderr, "autocd: shell=%s (%d)\n",
			shell.Path, shell.Type)
	}

	// 4. Generate appropriate script
	scriptContent, err := generateScript(validatedPath, shell)
	if err != nil {
		return newScriptGenerationError(err)
	}

	// 5. Write script to temporary file
	scriptPath, err := createTemporaryScript(scriptContent, shell.ScriptExt, opts.TempDir)
	if err != nil {
		return newScriptCreationError(err)
	}

	// 6. Execute script (this should never return)
	err = ExecReplacement(scriptPath, shell, opts.DebugMode)

	// If we reach here, execution failed
	os.Remove(scriptPath) // Cleanup on failure
	return newScriptExecutionError(err)
}

// ExitWithDirectoryOrFallback guarantees process exit
// Never returns - either succeeds with directory inheritance or calls fallback
func ExitWithDirectoryOrFallback(targetPath string, fallback func()) {
	if err := ExitWithDirectory(targetPath); err != nil {
		if os.Getenv("AUTOCD_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "autocd failed: %v\n", err)
		}
		fallback()
	}

	// Should never reach here, but just in case
	os.Exit(1)
}

// IsSupported checks if the current environment supports autocd
func IsSupported() bool {
	shell := detectShell("")
	return shell.IsValid
}

// GetCurrentShellInfo returns information about the detected shell
func GetCurrentShellInfo() *ShellInfo {
	return detectShell("")
}

// ValidateDirectory checks if a directory is valid for autocd without executing
func ValidateDirectory(targetPath string, securityLevel SecurityLevel) error {
	_, err := validateTargetPath(targetPath, securityLevel)
	if err != nil {
		return newPathValidationError(targetPath, err)
	}
	return nil
}
