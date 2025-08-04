package autocd

// ShellType represents different shell types
type ShellType int

const (
	ShellUnknown ShellType = iota
	ShellBash
	ShellZsh
	ShellFish
	ShellDash
	ShellSh
)

// String returns the string representation of the ShellType
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
	default:
		return "unknown"
	}
}

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
	ScriptExt string    // Script file extension (.sh)
	IsValid   bool      // Whether shell exists and is executable
}

// Options provides configuration for ExitWithDirectoryAdvanced
type Options struct {
	Shell                 string        // Override shell detection ("", "bash", "zsh", etc.)
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
	case ErrorShellNotFound:
		return false // Fundamental issue
	default:
		return true
	}
}
