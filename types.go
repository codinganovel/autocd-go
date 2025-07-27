package autocd

// PlatformType represents different operating system platforms
type PlatformType int

const (
	PlatformWindows PlatformType = iota
	PlatformMacOS
	PlatformLinux
	PlatformBSD
	PlatformUnix // Generic Unix fallback
)

// ShellType represents different shell types
type ShellType int

const (
	ShellUnknown ShellType = iota
	ShellBash
	ShellZsh
	ShellFish
	ShellDash
	ShellSh
	ShellCmd
	ShellPowerShell
	ShellPowerShellCore
)

// SecurityLevel defines path validation strictness
type SecurityLevel int

const (
	SecurityNormal     SecurityLevel = iota // Default: basic validation
	SecurityStrict                          // Paranoid: strict path validation
	SecurityPermissive                      // Minimal: user handles validation
)

// ShellInfo contains detected shell information
type ShellInfo struct {
	Path      string    // Full path to shell executable
	Type      ShellType // Detected shell type
	ScriptExt string    // Script file extension (.sh, .bat, .ps1)
	IsValid   bool      // Whether shell exists and is executable
}

// Options provides configuration for ExitWithDirectoryAdvanced
type Options struct {
	Shell                  string        // Override shell detection ("", "bash", "cmd", etc.)
	SecurityLevel         SecurityLevel // Strict, Normal, Permissive
	DebugMode             bool          // Enable verbose logging to stderr
	TempDir               string        // Override temp directory ("" = system default)
	DepthWarningThreshold int           // Shell depth threshold for warnings (default: 15)
	DisableDepthWarnings  bool          // Disable shell depth warning messages (default: false)
}

// ErrorType categorizes different types of autocd errors
type ErrorType int

const (
	ErrorPathNotFound ErrorType = iota
	ErrorPathNotDirectory
	ErrorPathNotAccessible
	ErrorShellNotFound
	ErrorScriptGeneration
	ErrorScriptExecution
	ErrorPlatformUnsupported
	ErrorSecurityViolation
)

// AutoCDError provides structured error information
type AutoCDError struct {
	Type    ErrorType
	Message string
	Path    string
	Cause   error
}

func (e *AutoCDError) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause of the error.
func (e *AutoCDError) Unwrap() error {
	return e.Cause
}

// IsRecoverable determines if error allows fallback
func (e *AutoCDError) IsRecoverable() bool {
	switch e.Type {
	case ErrorPathNotFound, ErrorPathNotAccessible:
		return true // Can fallback to normal exit
	case ErrorShellNotFound, ErrorPlatformUnsupported:
		return false // Fundamental issue
	default:
		return true
	}
}
