package autocd

import (
	"errors"
	"fmt"
)

// Exported error variables for specific validation failures
var (
	ErrPathNotFound      = errors.New("path does not exist")
	ErrPathNotDirectory  = errors.New("path is not a directory")
	ErrPathNotAccessible = errors.New("path is not accessible")
	ErrSecurityViolation = errors.New("security violation")
)

// Helper functions for common error cases
func newPathError(errType ErrorType, path string, cause error) *AutoCDError {
	return &AutoCDError{
		Type:    errType,
		Message: fmt.Sprintf("path error: %v", cause),
		Path:    path,
		Cause:   cause,
	}
}

func newPathValidationError(path string, cause error) *AutoCDError {
	errType := ErrorPathNotFound // Default error type
	switch {
	case errors.Is(cause, ErrPathNotDirectory):
		errType = ErrorPathNotDirectory
	case errors.Is(cause, ErrPathNotFound):
		errType = ErrorPathNotFound
	case errors.Is(cause, ErrPathNotAccessible):
		errType = ErrorPathNotAccessible
	case errors.Is(cause, ErrSecurityViolation):
		errType = ErrorSecurityViolation
	}

	return &AutoCDError{
		Type:    errType,
		Message: fmt.Sprintf("autocd: path validation failed: %v", cause),
		Path:    path,
		Cause:   cause,
	}
}

func newShellDetectionError(message string) *AutoCDError {
	return &AutoCDError{
		Type:    ErrorShellNotFound,
		Message: fmt.Sprintf("autocd: shell detection failed: %s", message),
		Path:    "",
		Cause:   nil,
	}
}

func newScriptGenerationError(cause error) *AutoCDError {
	return &AutoCDError{
		Type:    ErrorScriptGeneration,
		Message: fmt.Sprintf("autocd: script generation failed: %v", cause),
		Path:    "",
		Cause:   cause,
	}
}

func newScriptCreationError(cause error) *AutoCDError {
	return &AutoCDError{
		Type:    ErrorScriptGeneration,
		Message: fmt.Sprintf("autocd: script creation failed: %v", cause),
		Path:    "",
		Cause:   cause,
	}
}

func newScriptExecutionError(cause error) *AutoCDError {
	return &AutoCDError{
		Type:    ErrorScriptExecution,
		Message: fmt.Sprintf("autocd: script execution failed: %v", cause),
		Path:    "",
		Cause:   cause,
	}
}

// IsPathError checks if the error is related to path validation
func IsPathError(err error) bool {
	var autoCDErr *AutoCDError
	if errors.As(err, &autoCDErr) {
		return autoCDErr.Type == ErrorPathNotFound ||
			autoCDErr.Type == ErrorPathNotDirectory ||
			autoCDErr.Type == ErrorPathNotAccessible ||
			autoCDErr.Type == ErrorSecurityViolation
	}
	return false
}

// IsShellError checks if the error is related to shell detection
func IsShellError(err error) bool {
	var autoCDErr *AutoCDError
	if errors.As(err, &autoCDErr) {
		return autoCDErr.Type == ErrorShellNotFound
	}
	return false
}

// IsScriptError checks if the error is related to script generation/execution
func IsScriptError(err error) bool {
	var autoCDErr *AutoCDError
	if errors.As(err, &autoCDErr) {
		return autoCDErr.Type == ErrorScriptGeneration ||
			autoCDErr.Type == ErrorScriptExecution
	}
	return false
}
