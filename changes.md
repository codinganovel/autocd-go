# Session Changes

## Refactoring `errors.go`

### Summary

Replaced the custom, manually implemented `contains` and `indexAt` functions within the `errors.go` file with the equivalent, standard library functions from the `strings` package.

### Rationale

The library's `go.mod` file specifies a minimum Go version of `1.19`. The standard library functions (`strings.Contains` and `strings.Index`) have been available long before this version, making the custom implementations unnecessary.

This change provides several benefits:

1.  **Reduced Code:** Removed redundant code, making the codebase smaller and easier to maintain.
2.  **Improved Readability:** Developers are more familiar with the standard library, making the code easier to understand.
3.  **Better Performance:** The Go standard library's string functions are highly optimized and will perform better than the simple, custom implementations.

This modification makes the code more idiomatic and efficient without changing its functionality, as confirmed by running the project's test suite.

## Refactoring Error Handling to Use `errors.Is()`

### Summary

The error handling mechanism was refactored from relying on string-based matching of error messages to using specific error variables and the `errors.Is()` function. This significantly improves the robustness and maintainability of the error handling logic.

### Rationale

The original approach of checking `strings.Contains()` on error messages was brittle. Any change to the human-readable error message text would break the error identification logic. By using specific error variables and `errors.Is()`, the error's identity is decoupled from its message, aligning with idiomatic Go error handling and allowing for more precise and future-proof error management.

### Key Changes

1.  **Defined Exported Error Variables:** In `errors.go`, new `var` declarations were added for specific error types, such as `ErrPathNotFound`, `ErrPathNotDirectory`, `ErrPathNotAccessible`, and `ErrSecurityViolation`, using `errors.New()`.
2.  **Modified `validation.go` to Return Specific Errors:** Functions within `validation.go` (e.g., `validateTargetPath`, `validateStrict`, `validateNormal`) were updated to return these newly defined error variables directly, instead of `fmt.Errorf` with generic strings.
3.  **Updated `newPathValidationError` in `errors.go`:** The logic within `newPathValidationError` was changed to use `errors.Is(cause, Err...)` to determine the specific `ErrorType` based on the underlying error, replacing the `strings.Contains()` checks.
4.  **Added `Unwrap()` Method to `AutoCDError`:** In `types.go`, the `AutoCDError` struct was updated to implement the `Unwrap()` method. This method returns the `Cause` field of the `AutoCDError`, which is crucial for `errors.Is()` to correctly traverse and identify wrapped errors in the error chain.
5.  **Updated Test Suite (`autocd_test.go`):** The tests in `autocd_test.go` that assert on error types (e.g., `TestValidateDirectory_NonExistentDirectory`, `TestValidateDirectory_FileInsteadOfDirectory`, `TestErrorTypes`) were modified to use `errors.Is()` for their assertions, reflecting the new error handling approach.

### Issues Encountered and Fixed

1.  **Incorrect Backslash Escaping in `validation.go`:**
    *   **Problem:** A syntax error occurred in `validation.go` due to incorrect escaping of a backslash in a string literal (`"..\"`). This led to "newline in string" and "syntax error" compilation errors.
    *   **Fix:** The string literal was corrected to `"..\\"` to properly represent a literal backslash in Go.

2.  **"imported and not used" Compiler Errors:**
    *   **Problem:** The Go compiler reported "imported and not used" errors for the `strings` package in `errors.go` and the `errors` package in `validation.go`.
    *   **Fix:** The `import "strings"` statement was removed from `errors.go` as its functionality was replaced by `errors.Is()`.
    *   **Fix:** The `import "errors"` statement was removed from `validation.go` as its usage was indirect and not recognized by the compiler as a direct use, and the code functioned correctly without it.

3.  **`errors.Is()` Not Functioning Correctly:**
    *   **Problem:** Even after refactoring, `errors.Is()` was not correctly identifying the wrapped errors in tests. This was traced to the `AutoCDError` struct not implementing the `Unwrap()` method, which `errors.Is()` requires to traverse error chains.
    *   **Fix:** The `Unwrap()` method was added to the `AutoCDError` struct in `types.go`, returning its `Cause` field.

### Verification

After all changes and fixes were applied, the entire test suite was run using `go test -v ./...`, and all tests passed successfully, confirming the correctness and robustness of the refactored error handling.

## Bug Investigation and AUTOCD_DEBUG Fix

### Summary

Investigated reported bugs in the autocd-go library and fixed the `AUTOCD_DEBUG` environment variable handling issue.

### Bug Reports Investigated

1. **ExitWithDirectory() "Silent Failure"** - **NOT A LIBRARY BUG**
   - **Finding**: Library works correctly, issue is in client integration code
   - **Evidence**: Created reproduction cases that work perfectly
   - **Conclusion**: The reported failure is an integration issue, not a library bug

2. **Debug Output Issues** - **PARTIALLY FIXED**
   - **Issue**: `AUTOCD_DEBUG` environment variable was ignored by `ExitWithDirectory()`
   - **Root Cause**: Default options didn't check environment variable
   - **Fix Applied**: Modified `ExitWithDirectoryAdvanced()` to check `AUTOCD_DEBUG` when options are nil

### Changes Made

**File**: `autocd.go` line 24

**Before**:
```go
opts = &Options{
    SecurityLevel: SecurityNormal,
    DebugMode:     false,
}
```

**After**:
```go
opts = &Options{
    SecurityLevel: SecurityNormal,
    DebugMode:     os.Getenv("AUTOCD_DEBUG") != "",
}
```

### Rationale

The `ExitWithDirectory()` function is the primary library interface and should respect the `AUTOCD_DEBUG` environment variable for consistency with `ExitWithDirectoryOrFallback()`. Previously, only the advanced function with explicit options could enable debug mode.

### Verification

- Tested debug output appears only when `AUTOCD_DEBUG=1` is set
- Confirmed library functions correctly in all scenarios
- All existing tests continue to pass

## Shell Depth Warning Feature Implementation

### Summary

Implemented a comprehensive shell depth warning system that detects when users have accumulated many nested shells from navigation and provides helpful performance guidance. This addresses the issue where users unknowingly build up deep shell nesting (15+ levels) through repeated AutoCD usage, which can impact performance.

### Feature Specification

Based on the requirements in `autocd_shell_depth_feature.md`, the feature provides:
- **Unix Systems**: Uses `SHLVL` environment variable to detect shell depth, shows warnings at configurable threshold (default: 15)
- **Windows**: Always shows warning about unreliable detection with recommendation for periodic terminal restarts
- **Configuration**: Fully configurable threshold and disable options
- **User Experience**: Helpful, non-intrusive messages that don't interrupt workflow

### Implementation Details

#### 1. Configuration Options (types.go)
Extended the `Options` struct with two new fields:
```go
type Options struct {
    // ... existing fields ...
    DepthWarningThreshold int  // Shell depth threshold for warnings (default: 15)
    DisableDepthWarnings  bool // Disable shell depth warning messages (default: false)
}
```

#### 2. Core Warning Function (autocd.go)
Implemented `checkShellDepth(opts *Options)` with platform-specific logic:
- **Unix**: Reads `SHLVL` environment variable, compares against threshold
- **Windows**: Always shows reliability warning
- **Error Handling**: Graceful degradation - missing/invalid SHLVL silently skips
- **Performance**: ~17 nanoseconds overhead per call

#### 3. Integration Points (autocd.go)
Added shell depth checking to `ExitWithDirectoryAdvanced()` at line 74:
```go
// Check shell depth and show helpful warnings if appropriate
checkShellDepth(opts)
```

This ensures all entry points (`ExitWithDirectory`, `ExitWithDirectoryAdvanced`, `ExitWithDirectoryOrFallback`) get depth checking through the central Advanced function.

#### 4. Default Value Handling (autocd.go)
Enhanced options setup to provide sensible defaults:
```go
if opts == nil {
    opts = &Options{
        SecurityLevel:         SecurityNormal,
        DebugMode:             os.Getenv("AUTOCD_DEBUG") != "",
        DepthWarningThreshold: 15,    // New default
        DisableDepthWarnings:  false, // New default
    }
}

// Handle partial options
if opts.DepthWarningThreshold == 0 {
    opts.DepthWarningThreshold = 15
}
```

### Warning Messages

#### Unix Systems (when SHLVL >= threshold)
```
💡 Tip: You have 18 nested shells from navigation.
For better performance, consider opening a fresh terminal.
```

#### Windows Systems (always shown unless disabled)
```
💡 Shell nesting detection is not reliable on Windows.
Close and reopen your terminal from time to time to ensure optimal performance.
```

### Testing Implementation

Created comprehensive test suite in `shell_depth_test.go`:

#### Test Coverage
- **TestCheckShellDepth_Unix**: 8 test cases covering all Unix scenarios
  - Below/at/above threshold variations
  - Custom threshold configuration
  - Warnings disabled
  - Missing/invalid/negative SHLVL edge cases
- **TestCheckShellDepth_Windows**: 2 test cases for Windows behavior
  - Always-warn when enabled
  - Respects disable flag
- **TestOptions_ShellDepthDefaults**: 5 test cases for configuration defaults
- **TestShellDepthIntegration**: Integration test with main AutoCD functions
- **BenchmarkCheckShellDepth**: Performance benchmark (17.45 ns/op)

#### Test Results
- All tests pass on Unix systems
- Windows tests skip appropriately on non-Windows platforms
- Full test suite continues to pass (no regressions)
- Performance overhead is negligible

### Real-World Testing

Created test application `test_shell_depth/test_shell_depth_app` for manual verification:
- **Successful testing**: SHLVL tracking from 1 to 15, warning triggered exactly at threshold
- **Message format**: Matches specification exactly with 💡 emoji and helpful advice
- **Integration**: Works seamlessly with existing AutoCD functionality
- **Timing**: Warning appears before directory change, doesn't interfere with process replacement

### Documentation Updates

Updated all three documentation files:

#### CLAUDE.md
- Updated `Options` description and advanced usage examples
- Added comprehensive "Shell Depth Warnings" section with:
  - How it works (Unix vs Windows)
  - Example warning messages
  - Configuration options and disable instructions

#### README.md  
- Enhanced advanced usage example with new options
- Added concise "Shell Depth Warnings" section with basic explanation and disable option
- Kept brief as requested for standard user documentation

#### readme-developers.md
- Updated `Options` struct documentation
- Added comprehensive "Shell Depth Warning System" section with:
  - Feature overview and problem addressed
  - Platform-specific implementation details
  - Configuration options with types and defaults
  - Multiple usage examples and message examples
  - Implementation details, error handling, and performance metrics
  - Testing coverage summary

### Configuration Examples

#### Default Behavior
```go
// Uses threshold=15, warnings enabled
err := autocd.ExitWithDirectory("/target/path")
```

#### Custom Configuration
```go
opts := &autocd.Options{
    DepthWarningThreshold: 10,   // Custom threshold
    DisableDepthWarnings:  false, // Keep warnings enabled
}
err := autocd.ExitWithDirectoryAdvanced("/target/path", opts)
```

#### Disabled Warnings
```go
opts := &autocd.Options{DisableDepthWarnings: true}
err := autocd.ExitWithDirectoryAdvanced("/target/path", opts)
```

### Design Philosophy

The implementation follows the existing library patterns:
- **Non-intrusive**: Warnings go to stderr, never block core functionality
- **Graceful degradation**: Missing/invalid SHLVL doesn't cause errors
- **Platform-aware**: Different approaches for Unix vs Windows limitations
- **Configurable**: Full control for developers who want different behavior
- **Opt-out by default**: Helpful warnings enabled unless explicitly disabled
- **No dependencies**: Uses only Go standard library consistent with library design

### Backward Compatibility

- All existing code continues to work unchanged
- New fields get sensible defaults whether `opts` is nil or partially specified
- No breaking changes to public API
- Performance impact is negligible (sub-20 nanoseconds)

This feature successfully addresses the original problem of users unknowingly accumulating nested shells while maintaining the library's philosophy of being helpful without being intrusive.