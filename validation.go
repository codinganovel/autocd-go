package autocd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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

	// Character whitelist
	if runtime.GOOS == "windows" {
		if !isValidWindowsPath(path) {
			return "", ErrSecurityViolation
		}
	} else {
		if !isValidUnixPath(path) {
			return "", ErrSecurityViolation
		}
	}

	// Length limits
	if len(path) > 260 && runtime.GOOS == "windows" {
		return "", ErrSecurityViolation
	}
	if len(path) > 4096 {
		return "", ErrSecurityViolation
	}

	return path, nil
}

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

func validatePermissive(path string) (string, error) {
	// Minimal validation - just clean the path
	return filepath.Clean(path), nil
}

// isValidWindowsPath checks if path contains only valid Windows characters
func isValidWindowsPath(path string) bool {
	// Windows path character validation
	// Invalid: < > : " | ? * and control characters (0-31)
	invalidChars := `[<>:"|?*\x00-\x1f]`
	matched, _ := regexp.MatchString(invalidChars, path)
	return !matched
}

// isValidUnixPath checks if path contains only valid Unix characters
func isValidUnixPath(path string) bool {
	// Unix paths can contain most characters, but we'll be conservative
	// Invalid: null bytes and control characters that could cause issues
	invalidChars := `[\x00-\x1f\x7f]`
	matched, _ := regexp.MatchString(invalidChars, path)
	return !matched
}
