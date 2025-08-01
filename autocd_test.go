package autocd

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Test basic directory validation
func TestValidateDirectory_ValidDirectory(t *testing.T) {
	tempDir := os.TempDir()
	err := ValidateDirectory(tempDir, SecurityNormal)
	if err != nil {
		t.Errorf("ValidateDirectory failed for valid directory %s: %v", tempDir, err)
	}
}

func TestValidateDirectory_NonExistentDirectory(t *testing.T) {
	nonExistentDir := "/nonexistent/path/that/should/not/exist"
	err := ValidateDirectory(nonExistentDir, SecurityNormal)
	if !errors.Is(err, ErrPathNotFound) {
		t.Errorf("Expected ErrPathNotFound, got: %v", err)
	}
}

func TestValidateDirectory_FileInsteadOfDirectory(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "autocd_test_file_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	err = ValidateDirectory(tempFile.Name(), SecurityNormal)
	if !errors.Is(err, ErrPathNotDirectory) {
		t.Errorf("Expected ErrPathNotDirectory, got: %v", err)
	}
}

// Test security levels
func TestPathValidation_SecurityLevels(t *testing.T) {
	tempDir := os.TempDir()

	tests := []struct {
		name  string
		level SecurityLevel
	}{
		{"normal", SecurityNormal},
		{"strict", SecurityStrict},
		{"permissive", SecurityPermissive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectory(tempDir, tt.level)
			if err != nil {
				t.Errorf("ValidateDirectory failed for %s security level: %v", tt.name, err)
			}
		})
	}
}

// Test dangerous path characters
func TestPathValidation_DangerousCharacters(t *testing.T) {
	dangerousPaths := []string{
		"/tmp/test; rm -rf /",
		"/tmp/test|cat /etc/passwd",
		"/tmp/test`whoami`",
		"/tmp/test$(whoami)",
	}

	for _, path := range dangerousPaths {
		t.Run("dangerous_"+path, func(t *testing.T) {
			err := ValidateDirectory(path, SecurityNormal)
			if err == nil {
				t.Errorf("Expected validation to fail for dangerous path: %s", path)
			}
		})
	}
}

// Test platform support
func TestIsSupported(t *testing.T) {
	supported := IsSupported()
	// Should be supported on most platforms, just log the result
	t.Logf("Platform supported: %v", supported)
}

// Test shell detection
func TestGetCurrentShellInfo(t *testing.T) {
	shellInfo := GetCurrentShellInfo()
	if shellInfo == nil {
		t.Error("GetCurrentShellInfo returned nil")
		return
	}

	t.Logf("Detected shell: %s (type: %d)", shellInfo.Path, shellInfo.Type)
	t.Logf("Script extension: %s", shellInfo.ScriptExt)
	t.Logf("Shell valid: %v", shellInfo.IsValid)

	// Basic validation
	if shellInfo.ScriptExt == "" {
		t.Error("Script extension should not be empty")
	}
}

// Test helper functions
func TestDirectoryExists(t *testing.T) {
	tempDir := os.TempDir()
	if !DirectoryExists(tempDir) {
		t.Errorf("DirectoryExists failed for valid directory: %s", tempDir)
	}

	if DirectoryExists("/absolutely/nonexistent/path") {
		t.Error("DirectoryExists should return false for non-existent directory")
	}
}

func TestIsDirectoryAccessible(t *testing.T) {
	tempDir := os.TempDir()
	if !IsDirectoryAccessible(tempDir) {
		t.Errorf("IsDirectoryAccessible failed for accessible directory: %s", tempDir)
	}

	if IsDirectoryAccessible("/absolutely/nonexistent/path") {
		t.Error("IsDirectoryAccessible should return false for non-existent directory")
	}
}

func TestGetTempDir(t *testing.T) {
	// Test with empty custom dir
	result := GetTempDir("")
	if result != os.TempDir() {
		t.Errorf("Expected system temp dir, got: %s", result)
	}

	// Test with valid custom dir
	validDir := os.TempDir()
	result = GetTempDir(validDir)
	if result != validDir {
		t.Errorf("Expected %s, got %s", validDir, result)
	}

	// Test with invalid custom dir
	result = GetTempDir("/nonexistent/directory")
	if result == "/nonexistent/directory" {
		t.Error("Should fallback to system temp dir for invalid custom dir")
	}
}

// Test script generation
func TestGenerateScript_AllShellTypes(t *testing.T) {
	testPath := "/tmp/test"

	shells := []*ShellInfo{
		{Path: "/bin/bash", Type: ShellBash, ScriptExt: ".sh", IsValid: true},
		{Path: "/bin/zsh", Type: ShellZsh, ScriptExt: ".sh", IsValid: true},
		{Path: "/usr/bin/fish", Type: ShellFish, ScriptExt: ".sh", IsValid: true},
		{Path: "/bin/dash", Type: ShellDash, ScriptExt: ".sh", IsValid: true},
		{Path: "/bin/sh", Type: ShellSh, ScriptExt: ".sh", IsValid: true},
	}

	if runtime.GOOS == "windows" {
		shells = append(shells, []*ShellInfo{
			{Path: "cmd.exe", Type: ShellCmd, ScriptExt: ".bat", IsValid: true},
			{Path: "powershell.exe", Type: ShellPowerShell, ScriptExt: ".ps1", IsValid: true},
			{Path: "pwsh.exe", Type: ShellPowerShellCore, ScriptExt: ".ps1", IsValid: true},
		}...)
	}

	for _, shell := range shells {
		t.Run(shell.Type.String(), func(t *testing.T) {
			script, err := generateScript(testPath, shell)
			if err != nil {
				t.Errorf("Script generation failed for %v: %v", shell.Type, err)
				return
			}
			if !strings.Contains(script, testPath) {
				t.Error("Script doesn't contain target path")
			}
			if script == "" {
				t.Error("Generated script is empty")
			}
		})
	}
}

// Test script path sanitization - verify quotes are properly escaped
func TestScriptPathSanitization_QuoteEscaping(t *testing.T) {
	pathWithQuotes := `/tmp/test"quoted"path`

	tests := []struct {
		shell            *ShellInfo
		shouldContain    string
		shouldNotContain string
	}{
		{
			shell:            &ShellInfo{Path: "/bin/bash", Type: ShellBash, ScriptExt: ".sh", IsValid: true},
			shouldContain:    `/tmp/test\"quoted\"path`, // Escaped quotes for Unix
			shouldNotContain: `/tmp/test"quoted"path`,   // Original unescaped quotes
		},
		{
			shell:            &ShellInfo{Path: "cmd.exe", Type: ShellCmd, ScriptExt: ".bat", IsValid: true},
			shouldContain:    `/tmp/test""quoted""path`, // Doubled quotes for batch
			shouldNotContain: `/tmp/test"quoted"path`,   // Original unescaped quotes
		},
		{
			shell:            &ShellInfo{Path: "powershell.exe", Type: ShellPowerShell, ScriptExt: ".ps1", IsValid: true},
			shouldContain:    `/tmp/test""quoted""path`, // Doubled quotes for PowerShell
			shouldNotContain: `/tmp/test"quoted"path`,   // Original unescaped quotes
		},
	}

	for _, test := range tests {
		t.Run(test.shell.Type.String(), func(t *testing.T) {
			script, err := generateScript(pathWithQuotes, test.shell)
			if err != nil {
				t.Errorf("Script generation failed: %v", err)
				return
			}

			// Verify quotes are properly escaped
			if !strings.Contains(script, test.shouldContain) {
				t.Errorf("Script should contain escaped quotes: %s", test.shouldContain)
			}
			if strings.Contains(script, test.shouldNotContain) {
				t.Errorf("Script should not contain unescaped quotes: %s", test.shouldNotContain)
			}
		})
	}
}

// Test platform-specific dangerous command injection
func TestScriptPathSanitization_PlatformSpecific(t *testing.T) {
	tests := []struct {
		name             string
		shell            *ShellInfo
		dangerousPath    string
		dangerousCommand string
	}{
		{
			name:             "unix_rm_command",
			shell:            &ShellInfo{Path: "/bin/bash", Type: ShellBash, ScriptExt: ".sh", IsValid: true},
			dangerousPath:    `/tmp/test"; rm -rf /; echo "`,
			dangerousCommand: `"; rm -rf /; echo "`,
		},
		{
			name:             "windows_rd_command",
			shell:            &ShellInfo{Path: "cmd.exe", Type: ShellCmd, ScriptExt: ".bat", IsValid: true},
			dangerousPath:    `C:\temp\test"; rd /s /q C:\; echo "`,
			dangerousCommand: `"; rd /s /q C:\; echo "`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script, err := generateScript(test.dangerousPath, test.shell)
			if err != nil {
				t.Errorf("Script generation failed: %v", err)
				return
			}

			// Verify the dangerous command sequence is not present in its original form
			if strings.Contains(script, test.dangerousCommand) {
				t.Errorf("Script contains unescaped dangerous command: %s", test.dangerousCommand)
			}

			t.Logf("Generated script snippet: %s", script[0:min(200, len(script))])
		})
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Test shell classification
func TestClassifyShellByPath(t *testing.T) {
	tests := []struct {
		path     string
		expected ShellType
	}{
		{"/bin/bash", ShellBash},
		{"/usr/bin/zsh", ShellZsh},
		{"/usr/bin/fish", ShellFish},
		{"/bin/dash", ShellDash},
		{"/bin/sh", ShellSh},
		{"cmd.exe", ShellCmd},
		{"powershell.exe", ShellPowerShell},
		{"pwsh.exe", ShellPowerShellCore},
		{"/bin/unknown", ShellUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := classifyShellByPath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

// Test platform detection
func TestPlatformDetection(t *testing.T) {
	platform := detectPlatform()
	// Should return a valid platform type
	validPlatforms := []PlatformType{PlatformWindows, PlatformMacOS, PlatformLinux, PlatformBSD, PlatformUnix}
	found := false
	for _, valid := range validPlatforms {
		if platform == valid {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Invalid platform type: %d", platform)
	}
}

// Test error handling and error types
func TestErrorTypes(t *testing.T) {
	// Test path error
	pathErr := newPathValidationError("/nonexistent", ErrPathNotFound)
	if !IsPathError(pathErr) {
		t.Error("newPathValidationError should create a path error")
	}
	if !errors.Is(pathErr, ErrPathNotFound) {
		t.Error("Expected pathErr to be ErrPathNotFound")
	}

	// Test shell error
	shellErr := newShellDetectionError("test error")
	if !IsShellError(shellErr) {
		t.Error("newShellDetectionError should create a shell error")
	}

	// Test script error
	scriptErr := newScriptGenerationError(os.ErrNotExist)
	if !IsScriptError(scriptErr) {
		t.Error("newScriptGenerationError should create a script error")
	}
}

// Test integrated unused functions
func TestUnusedFunctionIntegration(t *testing.T) {
	// Test newPathError
	pathErr := newPathError(ErrorPathNotFound, "/test", os.ErrNotExist)
	if pathErr.Type != ErrorPathNotFound {
		t.Error("newPathError should set correct error type")
	}
	if pathErr.Path != "/test" {
		t.Error("newPathError should set correct path")
	}

	// Test newPlatformUnsupportedError
	platformErr := newPlatformUnsupportedError("unknown-os")
	if platformErr.Type != ErrorPlatformUnsupported {
		t.Error("newPlatformUnsupportedError should set correct error type")
	}
	if !strings.Contains(platformErr.Message, "unknown-os") {
		t.Error("newPlatformUnsupportedError should include platform in message")
	}
}

// Test error recoverability
func TestAutoCDError_IsRecoverable(t *testing.T) {
	recoverableErr := &AutoCDError{Type: ErrorPathNotFound}
	if !recoverableErr.IsRecoverable() {
		t.Error("Path not found error should be recoverable")
	}

	unrecoverableErr := newPlatformUnsupportedError("test platform")
	if unrecoverableErr.IsRecoverable() {
		t.Error("Platform unsupported error should not be recoverable")
	}
}

// Test cleanup functions
func TestCleanupOldScripts(t *testing.T) {
	err := CleanupOldScripts()
	if err != nil {
		t.Logf("CleanupOldScripts returned error (may be expected): %v", err)
	}
}

func TestCleanupOldScriptsWithAge(t *testing.T) {
	err := CleanupOldScriptsWithAge(1 * time.Hour)
	if err != nil {
		t.Logf("CleanupOldScriptsWithAge returned error (may be expected): %v", err)
	}
}

// Test path validation edge cases
func TestPathValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		level      SecurityLevel
		shouldFail bool
	}{
		{"empty_path", "", SecurityNormal, false}, // Empty path resolves to current directory
		{"dot_path", ".", SecurityNormal, false},
		{"path_traversal", "../../../etc", SecurityStrict, true},
		{"very_long_path", strings.Repeat("a", 5000), SecurityStrict, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateTargetPath(tt.path, tt.level)
			if tt.shouldFail && err == nil {
				t.Error("Expected validation to fail")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

// Test temp file permissions setting
func TestSetExecutablePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping executable permissions test on Windows")
	}

	tempFile, err := os.CreateTemp("", "autocd_perm_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	err = SetExecutablePermissions(tempFile.Name())
	if err != nil {
		t.Errorf("SetExecutablePermissions failed: %v", err)
	}

	// Check permissions
	info, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	mode := info.Mode()
	if mode&0111 == 0 {
		t.Error("File is not executable after SetExecutablePermissions")
	}
}

// Helper function to create ShellType string representation
func (s ShellType) String() string {
	switch s {
	case ShellBash:
		return "bash"
	case ShellZsh:
		return "zsh"
	case ShellFish:
		return "fish"
	case ShellDash:
		return "dash"
	case ShellSh:
		return "sh"
	case ShellCmd:
		return "cmd"
	case ShellPowerShell:
		return "powershell"
	case ShellPowerShellCore:
		return "powershell_core"
	default:
		return "unknown"
	}
}
