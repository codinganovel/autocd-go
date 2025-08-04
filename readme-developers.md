# AutoCD Go Library - Developer Technical Guide

A comprehensive technical guide for developers working with the `autocd-go` library.

## Table of Contents

- [Overview](#overview)
- [Core Concept](#core-concept)
- [Architecture](#architecture)
- [File Structure](#file-structure)
- [API Reference](#api-reference)
- [Platform Support](#platform-support)
- [Security Model](#security-model)
- [Script Generation](#script-generation)
- [Error Handling](#error-handling)
- [Temporary File Management](#temporary-file-management)
- [Process Replacement](#process-replacement)
- [Testing](#testing)
- [Integration Patterns](#integration-patterns)
- [Development Commands](#development-commands)
- [Implementation Examples](#implementation-examples)
- [Limitations & Considerations](#limitations--considerations)

## Overview

The AutoCD library solves a fundamental Unix limitation: when CLI applications navigate directories, the shell doesn't inherit that final location when the app exits. This library enables applications to transfer their final directory to the parent shell through secure process replacement.

**Key Features:**
- Unix-like system support (Linux, macOS, BSD)
- Multiple shell support (bash, zsh, fish, dash, sh)
- Three security validation levels
- Zero external dependencies
- Self-cleaning temporary scripts with automatic cleanup of old scripts
- Comprehensive error handling
- Simplified architecture focused on core Unix use case

## Core Concept

Instead of complex shell monitoring or protocols, the library uses a simple but effective approach:

1. **Generate a transition script** when the application is ready to exit
2. **Replace the current process** with this script using `syscall.Exec`
3. **The script changes directory** and spawns a new shell
4. **User gets a shell** in the final directory

### Example Transition Script (Unix)
```bash
#!/bin/bash
# autocd transition script - auto-cleanup on exit
trap 'rm -f "$0" 2>/dev/null || true' EXIT INT TERM

TARGET_DIR="/final/directory"
SHELL_PATH="/bin/bash"

# Attempt to change directory with error handling
if cd "$TARGET_DIR" 2>/dev/null; then
    echo "Directory changed to: $TARGET_DIR"
else
    echo "Warning: Could not change to $TARGET_DIR" >&2
    echo "Continuing in current directory" >&2
fi

# Replace current process with shell
exec "$SHELL_PATH"
```

## Architecture

### Core Flow
The library follows a 6-step process orchestrated by `autocd.go`:

1. **Cleanup Old Scripts** � `cleanupOldScripts()` (tempfile.go) - Automatic maintenance
2. **Path Validation** � `validateTargetPath()` (validation.go)
3. **Shell Detection** � `detectShell()` (shell.go)
4. **Script Generation** � `generateScript()` (script.go)
5. **Temp File Creation** � `createTemporaryScript()` (tempfile.go)
6. **Process Replacement** � `ExecReplacement()` � `syscall.Exec` (exec.go)

### Data Flow Diagram
```
User App � ExitWithDirectory() � cleanupOldScripts() � validateTargetPath() � detectShell()
    �
generateScript() � createTemporaryScript() � executeScript() � syscall.Exec � New Shell
```

## File Structure

### Core Implementation Files

| File | Purpose | Key Functions |
|------|---------|---------------|
| `autocd.go` | Main API and orchestration | `ExitWithDirectory`, `ExitWithDirectoryAdvanced`, `ExitWithDirectoryOrFallback` |
| `types.go` | Type definitions and constants | All structs, enums, and constants |
| `validation.go` | Path validation with security levels | `validateTargetPath`, `validateStrict`, `validateNormal`, `validatePermissive` |
| `shell.go` | Shell detection and configuration | `detectShell`, `detectUnixShell`, `classifyShell` |
| `script.go` | Script generation for Unix shells | `generateScript`, `generateUnixScript`, `sanitizePathForShell` |
| `exec.go` | Process replacement via syscall.Exec | `executeScript`, `ExecReplacement` |
| `tempfile.go` | Temporary file management and utilities | `createTemporaryScript`, `CleanupOldScripts`, `DirectoryExists` |
| `errors.go` | Structured error handling | `AutoCDError`, error type classification functions |

### Test and Documentation Files

| File | Purpose |
|------|---------|
| `autocd_test.go` | Comprehensive unit and integration tests |
| `testsuit.md` | Test planning and coverage documentation |
| `go.mod` | Go module definition (no external dependencies) |

## API Reference

### Primary Functions

#### ExitWithDirectory
```go
func ExitWithDirectory(targetPath string) error
```
**Purpose:** Simple directory inheritance with default settings.
**Behavior:** Never returns on success (process is replaced). Returns error on failure.
**Security Level:** Normal (default)

**Example:**
```go
import "github.com/codinganovel/autocd-go"

func main() {
    finalDir := "/path/to/target/directory"
    if err := autocd.ExitWithDirectory(finalDir); err != nil {
        fmt.Fprintf(os.Stderr, "autocd failed: %v\n", err)
        os.Exit(1)
    }
    // This line never executes on success
}
```

#### ExitWithDirectoryAdvanced
```go
func ExitWithDirectoryAdvanced(targetPath string, opts *Options) error
```
**Purpose:** Full control over autocd behavior with custom options.

**Options Structure:**
```go
type Options struct {
    Shell                  string        // Override shell detection ("", "bash", "cmd", etc.)
    SecurityLevel         SecurityLevel // Strict, Normal, Permissive
    DebugMode             bool          // Enable verbose logging to stderr
    TempDir               string        // Override temp directory ("" = system default)
    DepthWarningThreshold int           // Shell depth threshold for warnings (default: 15)
    DisableDepthWarnings  bool          // Disable shell depth warning messages (default: false)
}
```

**Example:**
```go
opts := &autocd.Options{
    Shell:                 "zsh",                    // Force zsh usage
    SecurityLevel:         autocd.SecurityStrict,   // Paranoid validation
    DebugMode:             true,                    // Verbose logging
    TempDir:               "/custom/temp",          // Custom temp directory
    DepthWarningThreshold: 10,                     // Show warnings at 10+ shells
    DisableDepthWarnings:  false,                  // Keep warnings enabled
}

err := autocd.ExitWithDirectoryAdvanced("/target/path", opts)
if err != nil {
    log.Printf("Advanced autocd failed: %v", err)
    os.Exit(1)
}
```

#### ExitWithDirectoryOrFallback
```go
func ExitWithDirectoryOrFallback(targetPath string, fallback func())
```
**Purpose:** Guarantees process exit - either via autocd or fallback function.
**Behavior:** Never returns. Calls fallback on autocd failure.

**Example:**
```go
autocd.ExitWithDirectoryOrFallback("/target/path", func() {
    fmt.Println("AutoCD failed, exiting normally")
    fmt.Printf("Final directory was: %s\n", targetPath)
    os.Exit(0)
})
// This line never executes
```

### Utility Functions

#### ValidateDirectory
```go
func ValidateDirectory(path string, level SecurityLevel) error
```
**Purpose:** Validate directory path without executing autocd.

#### CleanupOldScripts
```go
func CleanupOldScripts() error
func CleanupOldScriptsWithAge(maxAge time.Duration) error
```
**Purpose:** Remove old temporary autocd scripts (maintenance function).

#### DirectoryExists / IsDirectoryAccessible
```go
func DirectoryExists(path string) bool
func IsDirectoryAccessible(path string) bool
```
**Purpose:** Check directory existence and accessibility.

## Shell Depth Warning System

The library includes an intelligent shell depth warning system that helps users understand when they've accumulated many nested shells from navigation, providing helpful performance guidance.

### Feature Overview

**Problem Addressed:**
- Each AutoCD call spawns a new shell in the target directory
- Users can `exit` multiple times to backtrack through navigation path  
- Deep nesting (15+ shells) impacts performance
- This behavior is not obvious to users

**Solution:**
- Automatic detection of shell nesting depth
- Platform-aware warnings with helpful guidance
- Configurable thresholds and disable options
- Non-intrusive stderr messages

### Platform-Specific Implementation

#### Unix Systems (Linux, macOS, BSD)
Uses the `SHLVL` environment variable for reliable detection:

```go
// Unix shell depth detection
shlvlStr := os.Getenv("SHLVL")
shlvl, err := strconv.Atoi(shlvlStr)
if err == nil && shlvl >= opts.DepthWarningThreshold {
    fmt.Fprintf(os.Stderr, "💡 Tip: You have %d nested shells from navigation.\n", shlvl)
    fmt.Fprintf(os.Stderr, "For better performance, consider opening a fresh terminal.\n")
}
```


### Configuration Options

#### DepthWarningThreshold
- **Type:** `int`
- **Default:** `15`
- **Purpose:** Shell depth threshold for showing warnings

#### DisableDepthWarnings  
- **Type:** `bool`
- **Default:** `false`
- **Purpose:** Completely disable shell depth warning system
- **Use Case:** Power users who prefer silent operation

### Usage Examples

#### Default Behavior
```go
// Uses default threshold of 15, warnings enabled
err := autocd.ExitWithDirectory("/target/path")
```

#### Custom Threshold
```go
opts := &autocd.Options{
    DepthWarningThreshold: 10, // Warning at 10+ shells instead of 15
}
err := autocd.ExitWithDirectoryAdvanced("/target/path", opts)
```

#### Disabled Warnings
```go
opts := &autocd.Options{
    DisableDepthWarnings: true, // Silent operation
}
err := autocd.ExitWithDirectoryAdvanced("/target/path", opts)
```

#### Environment-Based Configuration
```go
opts := &autocd.Options{
    DepthWarningThreshold: getDepthThreshold(), // From env var or config
    DisableDepthWarnings:  os.Getenv("AUTOCD_QUIET") != "",
}
```

### Warning Message Examples

#### Unix Warning Message
```
💡 Tip: You have 18 nested shells from navigation.
For better performance, consider opening a fresh terminal.
```


### Implementation Details

#### Integration Point
Shell depth checking occurs early in `ExitWithDirectoryAdvanced()`:

```go
func ExitWithDirectoryAdvanced(targetPath string, opts *Options) error {
    // Set defaults...
    
    // Check shell depth and show helpful warnings if appropriate
    checkShellDepth(opts)
    
    // Continue with normal autocd flow...
}
```

#### Error Handling
- **Missing SHLVL:** Silently skip (graceful degradation)
- **Invalid SHLVL:** Silently skip (robust against malformed values)
- **Disabled warnings:** Respect user preference
- **Non-blocking:** Warnings never interfere with core functionality

#### Performance
- **Overhead:** ~17 nanoseconds per call (measured)
- **Platform detection:** Cached result from existing detection logic
- **Environment access:** Single `os.Getenv("SHLVL")` call on Unix
- **String conversion:** Only when SHLVL exists and is numeric

### Testing

The shell depth system includes comprehensive test coverage in `shell_depth_test.go`:

- **Unix platform testing:** All SHLVL scenarios and thresholds
- **Configuration testing:** Default values and custom options
- **Integration testing:** Verification with main AutoCD functions
- **Performance testing:** Benchmark confirming minimal overhead

## Platform Support

### Supported Platforms

| Platform | Status | Shells Supported |
|----------|--------|------------------|
| **macOS** |  Full | bash, zsh, fish, dash, sh |
| **Linux** |  Full | bash, zsh, fish, dash, sh |
| **FreeBSD** |  Full | sh, bash, zsh |
| **OpenBSD** |  Full | sh, bash, zsh |
| **NetBSD** |  Full | sh, bash, zsh |
| **Generic Unix** |  Fallback | sh, bash |

**Note:** AutoCD Go is focused exclusively on Unix-like systems. Windows support has been removed to simplify the architecture and focus on the core Unix use case where directory inheritance provides the most value.

### Shell Detection Priority

The library uses a simplified approach for Unix systems:

1. **SHELL environment variable** - User's preferred shell
2. **Fallback** to `/bin/sh` (POSIX standard)

### Shell Classification
```go
func classifyShell(shellPath string) ShellType {
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
        return ShellSh  // Generic sh-compatible
    }
}
```

## Security Model

The library implements a three-tier security model for path validation:

### SecurityStrict
**Use Case:** Paranoid environments, security-critical applications.

**Restrictions:**
- L No path traversal (`..` sequences)
- L Character whitelist validation
- L Length limits (4096 chars for Unix filesystems)
- L Strict character validation

```go
func validateStrict(path string) (string, error) {
    // No path traversal
    if strings.Contains(path, "..") {
        return "", ErrSecurityViolation
    }
    
    // Character whitelist validation
    if !isValidUnixPath(path) {
        return "", ErrSecurityViolation
    }
    
    // Length limits (Unix filesystems typically support up to 4096 bytes)
    if len(path) > 4096 {
        return "", ErrSecurityViolation
    }
    
    return path, nil
}
```

### SecurityNormal (Default)
**Use Case:** Most applications, balanced security and usability.

**Restrictions:**
- L Path traversal prevention via `filepath.Clean`
- L Shell injection character filtering: `;|&`$()<>`
-  Standard directory paths allowed

```go
func validateNormal(path string) (string, error) {
    // Prevent obvious path traversal
    cleanPath := filepath.Clean(path)
    if strings.Contains(cleanPath, "../") || strings.Contains(cleanPath, "..\\\\") {
        return "", ErrSecurityViolation
    }
    
    // Basic shell injection prevention
    dangerous := []string{";", "|", "&", "`", "$", "(", ")", "<", ">"}
    for _, char := range dangerous {
        if strings.Contains(path, char) {
            return "", ErrSecurityViolation
        }
    }
    
    return cleanPath, nil
}
```

### SecurityPermissive
**Use Case:** Trusted environments, user handles validation.

**Restrictions:**
-  Minimal validation - just path cleaning
-  User responsible for security

```go
func validatePermissive(path string) (string, error) {
    return filepath.Clean(path), nil
}
```

### Shell Injection Prevention

The library sanitizes paths differently for each shell type:

```go
func sanitizePathForShell(path string, shellType ShellType) string {
    // Escape for Unix shells
    path = strings.ReplaceAll(path, `\`, `\\`)
    path = strings.ReplaceAll(path, `"`, `\"`)
    return path
}
```

## Script Generation

The library generates platform and shell-specific transition scripts:

### Unix Script Template
```bash
#!/bin/bash  # or appropriate shebang
# autocd transition script - auto-cleanup on exit
trap 'rm -f "$0" 2>/dev/null || true' EXIT INT TERM

TARGET_DIR="/path/to/directory"
SHELL_PATH="/bin/bash"

# Attempt to change directory with error handling
if cd "$TARGET_DIR" 2>/dev/null; then
    echo "Directory changed to: $TARGET_DIR"
else
    echo "Warning: Could not change to $TARGET_DIR" >&2
    echo "Continuing in current directory" >&2
fi

# Replace current process with shell
exec "$SHELL_PATH"
```



### Shebang Selection
```go
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
```

## Error Handling

The library uses a sophisticated error handling system with pre-defined error variables, structured error types, and classification functions.

### Pre-defined Error Variables
```go
// Exported error variables for specific validation failures
var (
    ErrPathNotFound      = errors.New("path does not exist")
    ErrPathNotDirectory  = errors.New("path is not a directory")
    ErrPathNotAccessible = errors.New("path is not accessible")
    ErrSecurityViolation = errors.New("security violation")
)
```

### Error Types
```go
type ErrorType int
const (
    ErrorPathNotFound ErrorType = iota    // Directory doesn't exist
    ErrorPathNotDirectory                 // Path is not a directory
    ErrorPathNotAccessible               // Permission denied
    ErrorShellNotFound                   // No valid shell detected
    ErrorScriptGeneration               // Script creation failed
    ErrorScriptExecution                // Process replacement failed
    ErrorPlatformUnsupported            // Unsupported OS
    ErrorSecurityViolation              // Path validation failed
)
```

### Structured Error Type
```go
type AutoCDError struct {
    Type    ErrorType
    Message string
    Path    string
    Cause   error
}

func (e *AutoCDError) Error() string {
    return e.Message
}

// Unwrap returns the underlying cause of the error
func (e *AutoCDError) Unwrap() error {
    return e.Cause
}

func (e *AutoCDError) IsRecoverable() bool {
    switch e.Type {
    case ErrorPathNotFound, ErrorPathNotAccessible:
        return true  // Can fallback to normal exit
    case ErrorShellNotFound, ErrorPlatformUnsupported:
        return false // Fundamental issue
    default:
        return true
    }
}
```

### Error Classification Functions
```go
// Check error categories
func IsPathError(err error) bool    // Path validation errors
func IsShellError(err error) bool   // Shell detection errors  
func IsScriptError(err error) bool  // Script generation/execution errors
```

### Error Construction Helper Functions
The library provides internal helper functions for consistent error creation:
```go
func newPathValidationError(path string, cause error) *AutoCDError
func newShellDetectionError(message string) *AutoCDError
func newScriptGenerationError(cause error) *AutoCDError
func newScriptExecutionError(cause error) *AutoCDError
func newPlatformUnsupportedError(platform string) *AutoCDError
```

### Modern Error Handling Examples

#### Using Pre-defined Error Variables
```go
if err := autocd.ExitWithDirectory("/path"); err != nil {
    // Test for specific error types
    if errors.Is(err, autocd.ErrPathNotFound) {
        fmt.Println("Directory does not exist")
    } else if errors.Is(err, autocd.ErrPathNotDirectory) {
        fmt.Println("Path is not a directory")
    } else if errors.Is(err, autocd.ErrSecurityViolation) {
        fmt.Println("Security validation failed")
    }
    
    os.Exit(1)
}
```

#### Using Error Classification Functions
```go
if err := autocd.ExitWithDirectory("/path"); err != nil {
    // Test error categories
    if autocd.IsPathError(err) {
        fmt.Printf("Path-related error: %v\n", err)
    } else if autocd.IsShellError(err) {
        fmt.Printf("Shell detection error: %v\n", err)
    } else if autocd.IsScriptError(err) {
        fmt.Printf("Script error: %v\n", err)
    }
    
    os.Exit(1)
}
```

#### Advanced Error Handling with AutoCDError
```go
if err := autocd.ExitWithDirectory("/path"); err != nil {
    if autoCDErr, ok := err.(*autocd.AutoCDError); ok {
        switch autoCDErr.Type {
        case autocd.ErrorPathNotFound:
            fmt.Printf("Directory not found: %s\n", autoCDErr.Path)
        case autocd.ErrorShellNotFound:
            fmt.Printf("No compatible shell found\n")
        case autocd.ErrorSecurityViolation:
            fmt.Printf("Path validation failed: %s\n", autoCDErr.Message)
        }
        
        // Check if error is recoverable
        if autoCDErr.IsRecoverable() {
            fmt.Printf("Falling back to normal exit\n")
            os.Exit(0)
        } else {
            fmt.Printf("Fundamental issue, cannot recover\n")
            os.Exit(1)
        }
        
        // Access underlying cause if needed
        if autoCDErr.Cause != nil {
            fmt.Printf("Underlying cause: %v\n", autoCDErr.Cause)
        }
    }
}
```

#### Error Wrapping and Unwrapping
```go
// The AutoCDError type supports error unwrapping
if err := autocd.ExitWithDirectory("/path"); err != nil {
    // Can use errors.Unwrap to get the underlying cause
    if cause := errors.Unwrap(err); cause != nil {
        fmt.Printf("Underlying error: %v\n", cause)
    }
    
    // errors.Is works with wrapped errors
    if errors.Is(err, os.ErrNotExist) {
        fmt.Println("File system error detected")
    }
}

## Temporary File Management

### Core Functionality
```go
func createTemporaryScript(content, extension string, tempDir string) (string, error) {
    // Use custom temp dir or system default
    if tempDir == "" {
        tempDir = os.TempDir()
    }
    
    // Create temporary file with proper prefix and extension
    pattern := "autocd_*" + extension
    tmpFile, err := os.CreateTemp(tempDir, pattern)
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    defer tmpFile.Close()
    
    // Write script content
    if _, err := tmpFile.WriteString(content); err != nil {
        os.Remove(tmpFile.Name())
        return "", fmt.Errorf("failed to write script: %w", err)
    }
    
    // Set appropriate permissions (owner read/write/execute only)
    if err := os.Chmod(tmpFile.Name(), 0700); err != nil {
        os.Remove(tmpFile.Name())
        return "", fmt.Errorf("failed to set permissions: %w", err)
    }
    
    return tmpFile.Name(), nil
}
```

### Cleanup Utilities
- **File Pattern:** `autocd_*.sh` (shell scripts)
- **Permissions:** Files get 0700 (owner only read/write/execute)
- **Self-Cleanup:** Scripts use `trap` for automatic cleanup
- **Automatic Cleanup:** Library automatically cleans up scripts older than 1 hour on each call
- **Manual Cleanup:** `CleanupOldScripts()` available for additional maintenance

### Utility Functions
```go
// Check directory existence and accessibility
func DirectoryExists(path string) bool
func IsDirectoryAccessible(path string) bool

// Get temp directory with optional override
func GetTempDir(customDir string) string

// Set executable permissions on Unix
func SetExecutablePermissions(filePath string) error

// Cleanup old scripts
func CleanupOldScripts() error
func CleanupOldScriptsWithAge(maxAge time.Duration) error
```

## Process Replacement

### Unix Implementation
```go
func executeUnixScript(scriptPath string, shell *ShellInfo) error {
    executable := shell.Path  // e.g., "/bin/bash"
    args := []string{shell.Path, scriptPath}
    
    // Replace current process with the script
    return syscall.Exec(executable, args, os.Environ())
}
```

### Process Replacement Flow
1. **Script Creation** - Generate and write transition script
2. **Permission Setting** - Make executable (Unix only)
3. **Process Replacement** - `syscall.Exec` replaces current process
4. **Script Execution** - OS runs the script
5. **Directory Change** - Script changes to target directory
6. **Shell Spawn** - Script starts new shell and exits
7. **Cleanup** - Script removes itself using trap

## Testing

### Test Structure
The library includes comprehensive testing in `autocd_test.go`:

**Test Categories:**
-  **Path Validation Tests** - All security levels, error conditions
-  **Platform Detection Tests** - Cross-platform compatibility
-  **Shell Detection Tests** - Multiple shells, overrides
-  **Script Generation Tests** - All platforms and shells
-  **Error Handling Tests** - Error types and recovery
-  **Security Tests** - Injection prevention, path traversal
-  **Utility Function Tests** - Temp file management, cleanup
-  **Integration Tests** - Full workflow testing

### Running Tests
```bash
# Run all tests
go test

# Run with verbose output
go test -v

# Run specific test
go test -run TestFunctionName

# Run with coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
```

### Test Planning
Detailed test planning is documented in `testsuit.md` covering:
- Test case specifications
- Coverage requirements
- Platform-specific testing
- Integration test scenarios

## Integration Patterns

### Basic CLI Application
```go
package main

import (
    "flag"
    "os"
    "github.com/codinganovel/autocd-go"
)

func main() {
    var enableAutoCD = flag.Bool("autocd", false, "Change to final directory on exit")
    flag.Parse()
    
    app := NewMyApp()
    app.Run() // Your application logic
    
    // On exit, inherit directory if requested and changed
    if *enableAutoCD && app.CurrentDirectory() != app.StartingDirectory() {
        if err := autocd.ExitWithDirectory(app.CurrentDirectory()); err != nil {
            fmt.Fprintf(os.Stderr, "autocd failed: %v\n", err)
        }
    }
    
    os.Exit(0)
}
```

### File Manager Integration
```go
type FileManager struct {
    startDir      string
    currentDir    string
    autoCDEnabled bool
    debugMode     bool
}

func (fm *FileManager) Exit() {
    if fm.autoCDEnabled && fm.currentDir != fm.startDir {
        opts := &autocd.Options{
            SecurityLevel: autocd.SecurityNormal,
            DebugMode:     fm.debugMode,
        }
        
        if err := autocd.ExitWithDirectoryAdvanced(fm.currentDir, opts); err != nil {
            // Log error but don't crash
            fmt.Fprintf(os.Stderr, "Directory inheritance failed: %v\n", err)
            fmt.Fprintf(os.Stderr, "Final directory: %s\n", fm.currentDir)
        }
    }
    
    os.Exit(0)
}
```

### Guaranteed Exit Pattern
```go
func (app *MyApp) exitWithAutoCD() {
    autocd.ExitWithDirectoryOrFallback(app.targetDir, func() {
        fmt.Printf("AutoCD failed, but final directory was: %s\n", app.targetDir)
        os.Exit(0)
    })
    // Never reaches here
}
```

### Integration with Selection Tools (e.g., fzf)
```go
func runWithAutoCD(options *Options) {
    if options.Autocd {
        options.Output = make(chan string, 1)
        
        // Run selection tool in goroutine
        resultChan := make(chan struct{ code int; err error }, 1)
        go func() {
            code, err := runSelectionTool(options)
            resultChan <- struct{ code int; err error }{code, err}
        }()
        
        // Race between selection and completion
        select {
        case selectedPath := <-options.Output:
            // User made a selection - inherit that directory
            autocd.ExitWithDirectory(selectedPath)
            return
        case result := <-resultChan:
            // Tool completed normally
            exit(result.code, result.err)
            return
        }
    }
    
    // Normal execution without autocd
    code, err := runSelectionTool(options)
    exit(code, err)
}
```

## Development Commands

### Building and Testing
```bash
# Build the library
go build

# Run all tests
go test

# Run tests with verbose output
go test -v

# Run specific test
go test -run TestFunctionName

# Run tests with coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
```

### Dependencies
- **Go Version:** 1.19+ (specified in go.mod)
- **External Dependencies:** None (uses only Go standard library)

### Documentation References
- `testsuit.md` - Test planning and coverage documentation
- `readme.md` - User-facing library documentation
- `readme-developers.md` - This comprehensive technical guide

## Implementation Examples

### Simple Navigation Tool
```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "github.com/codinganovel/autocd-go"
)

func main() {
    reader := bufio.NewReader(os.Stdin)
    currentDir, _ := os.Getwd()
    
    for {
        fmt.Printf("nav [%s]> ", currentDir)
        input, _ := reader.ReadString('\n')
        command := strings.TrimSpace(input)
        
        switch {
        case command == "exit":
            // Exit with directory inheritance
            autocd.ExitWithDirectory(currentDir)
            return
        case command == "quit":
            // Exit without directory inheritance
            os.Exit(0)
        case strings.HasPrefix(command, "cd "):
            newDir := strings.TrimPrefix(command, "cd ")
            if err := os.Chdir(newDir); err != nil {
                fmt.Printf("Error: %v\n", err)
            } else {
                currentDir, _ = os.Getwd()
                fmt.Printf("Changed to: %s\n", currentDir)
            }
        default:
            fmt.Println("Commands: cd <path>, exit (with autocd), quit (without autocd)")
        }
    }
}
```

### Advanced Configuration Example
```go
func createAdvancedAutoCD(targetPath string) error {
    // Determine security level based on environment
    securityLevel := autocd.SecurityNormal
    if os.Getenv("AUTOCD_STRICT") != "" {
        securityLevel = autocd.SecurityStrict
    } else if os.Getenv("AUTOCD_PERMISSIVE") != "" {
        securityLevel = autocd.SecurityPermissive
    }
    
    // Configure options
    opts := &autocd.Options{
        SecurityLevel: securityLevel,
        DebugMode:     os.Getenv("AUTOCD_DEBUG") != "",
        TempDir:       os.Getenv("AUTOCD_TEMPDIR"), // Custom temp dir
    }
    
    // Override shell if specified
    if shell := os.Getenv("AUTOCD_SHELL"); shell != "" {
        opts.Shell = shell
    }
    
    return autocd.ExitWithDirectoryAdvanced(targetPath, opts)
}
```

### Error Recovery Example
```go
func exitWithRecovery(targetPath string) {
    err := autocd.ExitWithDirectory(targetPath)
    if err != nil {
        // Use modern error handling with pre-defined errors and classification
        if errors.Is(err, autocd.ErrPathNotFound) {
            fmt.Fprintf(os.Stderr, "Directory not found: %s\n", targetPath)
            
            // Try parent directory as fallback
            parentPath := filepath.Dir(targetPath)
            if parentPath != targetPath { // Avoid infinite recursion
                fmt.Fprintf(os.Stderr, "Trying parent directory: %s\n", parentPath)
                if err2 := autocd.ExitWithDirectory(parentPath); err2 == nil {
                    return // Success with parent directory
                }
            }
        } else if errors.Is(err, autocd.ErrPathNotDirectory) {
            fmt.Fprintf(os.Stderr, "Path is not a directory: %s\n", targetPath)
        } else if autocd.IsShellError(err) {
            fmt.Fprintf(os.Stderr, "Shell detection failed: %v\n", err)
            fmt.Fprintf(os.Stderr, "Try setting SHELL environment variable\n")
        } else if autocd.IsScriptError(err) {
            fmt.Fprintf(os.Stderr, "Script generation/execution failed: %v\n", err)
        }
        
        // Check if error is recoverable
        if autoCDErr, ok := err.(*autocd.AutoCDError); ok && autoCDErr.IsRecoverable() {
            fmt.Fprintf(os.Stderr, "Recoverable error, falling back to normal exit\n")
            fmt.Fprintf(os.Stderr, "Final directory: %s\n", targetPath)
            os.Exit(0)
        }
        
        // Final fallback for unrecoverable errors
        fmt.Fprintf(os.Stderr, "AutoCD failed completely: %v\n", err)
        fmt.Fprintf(os.Stderr, "Final directory: %s\n", targetPath)
        os.Exit(1)
    }
}
```

## Limitations & Considerations

### Functional Limitations
1. **No Return on Success** - Functions never return when successful (process is replaced)
2. **Shell Dependency** - Requires compatible shell installation
3. **Temporary Files** - Creates temporary scripts that must be cleaned up
4. **Platform Specific** - Behavior varies slightly between operating systems

### Security Considerations
1. **Script Injection** - Path sanitization prevents injection but strict validation recommended for untrusted input
2. **Temporary File Security** - Scripts created with restrictive permissions (0700 on Unix)
3. **Path Traversal** - Validation prevents directory traversal attacks
4. **Shell Compatibility** - Different shells may have different security models

### Performance Considerations
1. **File System Access** - Multiple file system operations during execution
2. **Process Replacement** - `syscall.Exec` has OS overhead
3. **Shell Detection** - May check multiple paths during shell detection
4. **Script Generation** - String manipulation and file I/O overhead

### Integration Considerations
1. **Error Handling** - Always provide fallback handling for autocd failures
2. **User Experience** - Consider making autocd optional with command-line flags
3. **Testing** - Mock process replacement for unit tests
4. **Debugging** - Use debug mode during development and troubleshooting

### Future Improvements
1. **Additional Shell Support** - nushell, elvish, xonsh
2. **Better Cleanup** - More robust temporary file lifecycle management
3. **Enhanced Error Recovery** - More granular error handling options
4. **Performance Optimization** - Cache shell detection results
5. **Plugin Architecture** - Allow custom script generation for specialized shells

---

*This comprehensive technical guide covers all aspects of the AutoCD Go library for developers implementing directory inheritance functionality.*