package autocd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Pre-compiled regex for performance
var (
	invalidCharsRegex = regexp.MustCompile(`[\x00-\x1f\x7f]`)
)

// validateTargetPath performs security validation based on level
func validateTargetPath(path string, level SecurityLevel) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists and is directory
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrPathNotFound
		}
		return "", fmt.Errorf("failed to stat path: %w", err)
	}
	if !info.IsDir() {
		return "", ErrPathNotDirectory
	}
	
	// Check if directory is accessible (can be read/executed)
	if _, err := os.ReadDir(absPath); err != nil {
		return "", ErrPathNotAccessible
	}

	// Security level specific validation
	switch level {
	case SecurityStrict:
		return validateStrict(absPath)
	case SecurityNormal:
		return validateNormal(absPath)
	case SecurityPermissive:
		return validatePermissive(absPath)
	default:
		return validateNormal(absPath)
	}
}

func validateStrict(path string) (string, error) {

	// Character whitelist for Unix paths
	if !isValidUnixPath(path) {
		return "", ErrSecurityViolation
	}

	// Length limits (Unix filesystems typically support up to 4096 bytes)
	if len(path) > 4096 {
		return "", ErrSecurityViolation
	}

	return path, nil
}

func validateNormal(path string) (string, error) {
	// Clean the path first
	cleanPath := filepath.Clean(path)
	
	// With proper single-quote escaping in scripts, we don't need to block
	// most shell metacharacters in directory names. Only block the most
	// dangerous combinations that could break out of quotes.
	
	// Check for null bytes which can't be in valid paths
	if strings.Contains(path, "\x00") {
		return "", ErrSecurityViolation
	}

	return cleanPath, nil
}

func validatePermissive(path string) (string, error) {
	// Minimal validation - just clean the path
	return filepath.Clean(path), nil
}

// isValidUnixPath checks if path contains only valid Unix characters
func isValidUnixPath(path string) bool {
	// Unix paths can contain most characters, but we'll be conservative
	// Invalid: null bytes and control characters that could cause issues
	return !invalidCharsRegex.MatchString(path)
}
