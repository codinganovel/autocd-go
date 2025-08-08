package autocd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Test validateShellOverride - currently 0% coverage
func TestValidateShellOverride(t *testing.T) {
	tests := []struct {
		name          string
		shellOverride string
		wantValid     bool
	}{
		{
			name:          "absolute_path_bash",
			shellOverride: "/bin/bash",
			wantValid:     true,
		},
		{
			name:          "shell_name_in_path",
			shellOverride: "sh",
			wantValid:     true,
		},
		{
			name:          "non_existent_shell",
			shellOverride: "/non/existent/shell",
			wantValid:     false,
		},
		{
			name:          "relative_shell_name",
			shellOverride: "bash",
			wantValid:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For shell names, check if they exist in PATH first
			if !filepath.IsAbs(tt.shellOverride) {
				if _, err := exec.LookPath(tt.shellOverride); err != nil && tt.wantValid {
					t.Skip("Shell not found in PATH")
				}
			} else if !fileExists(tt.shellOverride) && tt.wantValid {
				t.Skip("Shell path does not exist")
			}

			result := validateShellOverride(tt.shellOverride)

			if result.IsValid != tt.wantValid {
				t.Errorf("validateShellOverride(%s).IsValid = %v, want %v",
					tt.shellOverride, result.IsValid, tt.wantValid)
			}
		})
	}
}

// Test createTemporaryScript - currently 0% coverage
func TestCreateTemporaryScript(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		extension string
		tempDir   string
		wantErr   bool
	}{
		{
			name:      "basic_shell_script",
			content:   "#!/bin/bash\necho 'test'",
			extension: ".sh",
			tempDir:   "",
			wantErr:   false,
		},
		{
			name:      "batch_script",
			content:   "@echo off\necho test",
			extension: ".bat",
			tempDir:   "",
			wantErr:   false,
		},
		{
			name:      "custom_temp_dir",
			content:   "test content",
			extension: ".sh",
			tempDir:   os.TempDir(),
			wantErr:   false,
		},
		{
			name:      "invalid_temp_dir",
			content:   "test",
			extension: ".sh",
			tempDir:   "/non/existent/directory",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath, err := createTemporaryScript(tt.content, tt.extension, tt.tempDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("createTemporaryScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file exists (check with os.Stat, not fileExists which checks executable)
				if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
					t.Error("Created script file does not exist")
				}

				// Verify content
				content, err := os.ReadFile(scriptPath)
				if err != nil {
					t.Errorf("Failed to read created script: %v", err)
				} else if string(content) != tt.content {
					t.Errorf("Script content mismatch: got %q, want %q", content, tt.content)
				}

				// Verify permissions on Unix
				info, err := os.Stat(scriptPath)
				if err != nil {
					t.Errorf("Failed to stat script: %v", err)
				} else if info.Mode()&0700 != 0700 {
					t.Errorf("Script permissions incorrect: %v", info.Mode())
				}

				// Clean up
				os.Remove(scriptPath)
			}
		})
	}
}

// Test error constructors - currently 0% coverage
func TestNewScriptCreationError(t *testing.T) {
	cause := errors.New("permission denied")
	err := newScriptCreationError(cause)

	autoCDErr := err
	if autoCDErr == nil {
		t.Fatal("newScriptCreationError should return *AutoCDError")
	}

	if autoCDErr.Type != ErrorScriptGeneration {
		t.Errorf("Error type = %v, want %v", autoCDErr.Type, ErrorScriptGeneration)
	}

	if !strings.Contains(autoCDErr.Message, "script creation failed") {
		t.Error("Error message should mention script creation")
	}

	if autoCDErr.Cause != cause {
		t.Error("Error cause not properly wrapped")
	}
}

func TestNewScriptExecutionError(t *testing.T) {
	cause := errors.New("exec format error")
	err := newScriptExecutionError(cause)

	autoCDErr := err
	if autoCDErr == nil {
		t.Fatal("newScriptExecutionError should return *AutoCDError")
	}

	if autoCDErr.Type != ErrorScriptExecution {
		t.Errorf("Error type = %v, want %v", autoCDErr.Type, ErrorScriptExecution)
	}

	if !strings.Contains(autoCDErr.Message, "script execution failed") {
		t.Error("Error message should mention script execution")
	}
}

// Test AutoCDError.Error() method - currently 0% coverage
func TestAutoCDError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *AutoCDError
		want string
	}{
		{
			name: "basic_error",
			err: &AutoCDError{
				Type:    ErrorPathNotFound,
				Message: "test error message",
				Path:    "/test/path",
				Cause:   nil,
			},
			want: "test error message",
		},
		{
			name: "error_with_cause",
			err: &AutoCDError{
				Type:    ErrorScriptGeneration,
				Message: "script failed",
				Path:    "",
				Cause:   errors.New("underlying cause"),
			},
			want: "script failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("AutoCDError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test edge cases for error classification
func TestErrorClassification_FalseCases(t *testing.T) {
	// Create non-path error
	nonPathErr := &AutoCDError{Type: ErrorShellNotFound}
	if IsPathError(nonPathErr) {
		t.Error("IsPathError should return false for non-path errors")
	}

	// Create non-shell error
	nonShellErr := &AutoCDError{Type: ErrorScriptGeneration}
	if IsShellError(nonShellErr) {
		t.Error("IsShellError should return false for non-shell errors")
	}

	// Create non-script error
	nonScriptErr := &AutoCDError{Type: ErrorPathNotFound}
	if IsScriptError(nonScriptErr) {
		t.Error("IsScriptError should return false for non-script errors")
	}

	// Test with non-AutoCDError
	regularErr := errors.New("regular error")
	if IsPathError(regularErr) || IsShellError(regularErr) || IsScriptError(regularErr) {
		t.Error("Error classification should return false for non-AutoCDError")
	}
}

// Test cleanupOldScripts with actual files
func TestCleanupOldScripts_WithRealFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create old and new autocd scripts
	oldFile := filepath.Join(tempDir, "autocd_old.sh")
	newFile := filepath.Join(tempDir, "autocd_new.sh")
	notAutoCDFile := filepath.Join(tempDir, "other_file.sh")

	// Create files
	for _, file := range []string{oldFile, newFile, notAutoCDFile} {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Make old file actually old
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old file time: %v", err)
	}

	// Save original temp dir and temporarily override
	originalTempDir := os.TempDir()
	os.Setenv("TMPDIR", tempDir)
	defer os.Setenv("TMPDIR", originalTempDir)

	// Run cleanup
	err := cleanupOldScripts(1 * time.Hour)
	if err != nil {
		t.Errorf("cleanupOldScripts failed: %v", err)
	}

	// Check results (use os.Stat since fileExists checks executable bit)
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old autocd file should have been deleted")
	}
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New autocd file should still exist")
	}
	if _, err := os.Stat(notAutoCDFile); os.IsNotExist(err) {
		t.Error("Non-autocd file should still exist")
	}
}

// Test additional validation edge cases
func TestValidateStrict_AllCases(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errType error
	}{
		{
			name:    "valid_path",
			path:    "/tmp/valid",
			wantErr: false,
		},
		// Removed path_with_traversal test - no longer relevant after removing dead check
		{
			name:    "unix_invalid_chars",
			path:    "/tmp/test\x00file",
			wantErr: true,
			errType: ErrSecurityViolation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for valid path test
			if tt.name == "valid_path" {
				os.MkdirAll(tt.path, 0755)
				defer os.RemoveAll(tt.path)
			}

			_, err := validateStrict(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateStrict() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errType != nil {
				if tt.errType == ErrSecurityViolation && !errors.Is(err, ErrSecurityViolation) {
					t.Errorf("validateStrict() error type = %v, want %v", err, tt.errType)
				}
			}
		})
	}
}

// Test fileExists edge cases
func TestFileExists_DirectoryCase(t *testing.T) {
	tempDir := t.TempDir()

	// Test with directory (should return false)
	if fileExists(tempDir) {
		t.Error("fileExists should return false for directories")
	}

	// Test with actual file (non-executable)
	tempFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if fileExists(tempFile) {
		t.Error("fileExists should return false for non-executable file")
	}

	// Make it executable
	if err := os.Chmod(tempFile, 0755); err != nil {
		t.Fatalf("Failed to make file executable: %v", err)
	}

	if !fileExists(tempFile) {
		t.Error("fileExists should return true for executable file")
	}
}

// Test pre-compiled regex performance
func TestRegexPrecompilation(t *testing.T) {
	// This tests that regex patterns compile successfully
	unixPattern := `[\x00-\x1f\x7f]`

	_, err := regexp.Compile(unixPattern)
	if err != nil {
		t.Errorf("Unix regex pattern failed to compile: %v", err)
	}
}

// Test validateNormal with new single-quote protection
func TestValidateNormal_AdditionalCases(t *testing.T) {
	// These characters are now safe with single-quote escaping
	safeChars := []string{";", "|", "&", "`", "$", "(", ")", "<", ">"}

	for _, char := range safeChars {
		t.Run("char_"+char, func(t *testing.T) {
			path := "/tmp/test" + char + "file"
			cleanedPath, err := validateNormal(path)

			if err != nil {
				t.Errorf("validateNormal should not reject path with %q character when using single quotes: %v", char, err)
			}

			if cleanedPath == "" {
				t.Error("validateNormal should return cleaned path")
			}
		})
	}

	// Test null byte (should still be rejected)
	t.Run("null_byte", func(t *testing.T) {
		path := "/tmp/test\x00file"
		_, err := validateNormal(path)

		if err == nil {
			t.Error("validateNormal should reject path with null byte")
		}

		if !errors.Is(err, ErrSecurityViolation) {
			t.Error("Expected ErrSecurityViolation for null byte")
		}
	})
}

// Test SetExecutablePermissions error case
func TestSetExecutablePermissions_ErrorCase(t *testing.T) {
	// Test with non-existent file
	err := SetExecutablePermissions("/non/existent/file")
	if err == nil {
		t.Error("SetExecutablePermissions should fail for non-existent file")
	}
}

// Test classifyShell removed - function no longer exists

// Test detectShell with shell override
func TestDetectShell_WithOverride(t *testing.T) {
	// Test with valid override
	shell := detectShell("echo") // echo should exist on all systems
	if _, err := exec.LookPath("echo"); err == nil {
		if !shell.IsValid {
			t.Error("detectShell with valid override should return valid shell")
		}
	}

	// Test with invalid override
	invalidShell := detectShell("/definitely/not/a/shell")
	if invalidShell.IsValid {
		t.Error("detectShell with invalid override should return invalid shell")
	}
}

// Test detectUnixShell edge cases
func TestDetectUnixShell_EdgeCases(t *testing.T) {

	// Save original SHELL
	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)

	// Test with empty SHELL
	os.Unsetenv("SHELL")
	shell := detectUnixShell()
	if shell.Path != "/bin/sh" {
		t.Errorf("detectUnixShell with no SHELL env should default to /bin/sh, got %s", shell.Path)
	}

	// Test with custom SHELL; fallback to /bin/sh if invalid
	os.Setenv("SHELL", "/usr/bin/fish")
	shell = detectUnixShell()
	if fileExists("/usr/bin/fish") {
		if shell.Path != "/usr/bin/fish" {
			t.Error("detectUnixShell should detect fish shell from SHELL env when present")
		}
	} else {
		if shell.Path != "/bin/sh" {
			t.Error("detectUnixShell should fallback to /bin/sh when SHELL is invalid")
		}
	}
}

// Test IsRecoverable for all error types
func TestIsRecoverable_AllErrorTypes(t *testing.T) {
	tests := []struct {
		errorType   ErrorType
		recoverable bool
	}{
		{ErrorPathNotFound, true},
		{ErrorPathNotDirectory, true},
		{ErrorPathNotAccessible, true},
		{ErrorShellNotFound, false},
		{ErrorScriptGeneration, true},
		{ErrorScriptExecution, true},
		{ErrorSecurityViolation, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("error_type_%d", tt.errorType), func(t *testing.T) {
			err := &AutoCDError{Type: tt.errorType}
			got := err.IsRecoverable()
			if got != tt.recoverable {
				t.Errorf("IsRecoverable() for type %v = %v, want %v",
					tt.errorType, got, tt.recoverable)
			}
		})
	}
}
