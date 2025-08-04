package autocd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// createTemporaryScript writes script content to temp file
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

	// Set appropriate permissions (Unix systems)
	// Make executable (owner read/write/execute only)
	if err := os.Chmod(tmpFile.Name(), 0700); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to set permissions: %w", err)
	}

	return tmpFile.Name(), nil
}

// cleanupOldScripts removes old autocd scripts (optional cleanup)
func cleanupOldScripts(maxAge time.Duration) error {
	tempDir := os.TempDir()

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return err // Non-fatal - just return error
	}

	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "autocd_") {
			info, err := entry.Info()
			if err != nil {
				continue // Skip files we can't stat
			}

			// Use pre-calculated cutoff time for better performance
			if info.ModTime().Before(cutoff) {
				os.Remove(filepath.Join(tempDir, entry.Name()))
			}
		}
	}

	return nil
}

// CleanupOldScripts is a public function to clean up old autocd scripts
// Applications can call this periodically to prevent temp directory buildup
func CleanupOldScripts() error {
	// Clean up scripts older than 24 hours
	return cleanupOldScripts(24 * time.Hour)
}

// CleanupOldScriptsWithAge allows custom age threshold
func CleanupOldScriptsWithAge(maxAge time.Duration) error {
	return cleanupOldScripts(maxAge)
}

// DirectoryExists checks if a directory exists and is accessible
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsDirectoryAccessible checks if directory can be read and accessed
func IsDirectoryAccessible(path string) bool {
	if !DirectoryExists(path) {
		return false
	}

	// Try to read the directory
	_, err := os.ReadDir(path)
	return err == nil
}

// GetTempDir returns the system temp directory or a custom one if specified
func GetTempDir(customTempDir string) string {
	if customTempDir != "" && DirectoryExists(customTempDir) {
		return customTempDir
	}
	return os.TempDir()
}

// SetExecutablePermissions sets executable permissions on Unix systems
func SetExecutablePermissions(path string) error {
	return os.Chmod(path, fs.FileMode(0755))
}
