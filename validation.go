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
	// No path traversal
	if strings.Contains(path, "..") {
		return "", ErrSecurityViolation
	}

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

// Pre-compiled dangerous characters for performance
var dangerousChars = [...]string{";", "|", "&", "`", "$", "(", ")", "<", ">"}

func validateNormal(path string) (string, error) {
	// Prevent obvious path traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "../") || strings.Contains(cleanPath, "..\\\\") {
		return "", ErrSecurityViolation
	}

	// Basic shell injection prevention - use array for better performance
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return "", ErrSecurityViolation
		}
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
